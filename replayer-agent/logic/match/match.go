package match

import (
	"bytes"
	"context"
	"math"
	"regexp"
	"sync"

	"github.com/didi/sharingan/replayer-agent/common/global"
	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/logic/replayed"
	"github.com/didi/sharingan/replayer-agent/model/protocol"
	"github.com/didi/sharingan/replayer-agent/model/recording"
	"github.com/didi/sharingan/replayer-agent/model/replaying"
	"github.com/didi/sharingan/replayer-agent/utils/helper"
)

var expect100 = []byte("Expect: 100-continue")
var httpRegex = regexp.MustCompile(`(?U)(^(?:GET|POST|PUT|DELETE|HEAD|OPTIONS|TRACE|CONNECT) .*?) HTTP/`)

type Matcher struct {
	sync.Mutex

	EnableCursor    bool         // 是否开启 MaxMatchedIndex 游标
	MaxMatchedIndex int          // 游标，用于标记最大当前已匹配下标，在颠倒相同请求顺序时有效
	Visited         map[int]bool // 标记已匹配 Action
}

func New() *Matcher {
	return &Matcher{
		EnableCursor:    global.FlagHandler.EnableCursor,
		MaxMatchedIndex: -1,
		Visited:         make(map[int]bool),
	}
}

func (m *Matcher) MatchOutboundTalk(
	ctx context.Context, session *replaying.Session, lastMatchedIndex int, request []byte) (int, float64, *recording.CallOutbound) {
	unit := 16
	chunks := cutToChunks(request, unit)
	reqCandidates := getRequests(session)
	scores := make([]int, len(session.CallOutbounds))
	reqExpect100 := bytes.Contains(request, expect100)
	for i, callOutbound := range session.CallOutbounds {
		if reqExpect100 != bytes.Contains(callOutbound.Request, expect100) {
			scores[i] = math.MinInt64
		}
	}
	maxScore := 0
	maxScoreIndex := 0
	for chunkIndex, chunk := range chunks {
		for j, reqCandidate := range reqCandidates {
			if j <= lastMatchedIndex {
				continue
			}
			if len(reqCandidate) < len(chunk) {
				continue
			}
			pos := bytes.Index(reqCandidate, chunk)
			if pos >= 0 {
				// avoid long jump
				// if pos < 96 {
				// 	reqCandidates[j] = reqCandidate[pos:]
				// }
				if chunkIndex == 0 && pos == 0 && lastMatchedIndex == -1 {
					scores[j] += len(chunks) // first chunk has more weight
				} else {
					scores[j]++
				}
				hasBetterScore := m.HasBetterScore(scores[j], j, maxScore)
				if hasBetterScore {
					maxScore = scores[j]
					maxScoreIndex = j
				}
			}
		}
	}

	// fix mis-match for concurrent disconnection of mysql
	m.Lock()
	defer m.Unlock()

	if m.EnableCursor {
		for j, score := range scores {
			if score == maxScore && m.MaxMatchedIndex < j {
				maxScoreIndex = j
				break
			}
		}
	}
	mark := float64(maxScore) / float64(len(chunks))
	tlog.Handler.Debugf(ctx, tlog.DebugTag,
		"%s||maxScoreIndex=%v||lastMatchedIndex=%v||maxMatchedIndex=%v||maxScore=%v||totalScore=%v||mark=%v||scores=%v||visited=%v",
		helper.CInfo("|||match detail|||"), maxScoreIndex, lastMatchedIndex, m.MaxMatchedIndex, maxScore, len(chunks), mark, scores, m.Visited)
	if maxScore == 0 {
		return -1, 0, nil
	}
	if lastMatchedIndex != -1 {
		// not starting from beginning, should have minimal score
		if mark < 0.85 {
			return -1, 0, nil
		}
	} else {
		if mark < 0.1 {
			return -1, 0, nil
		}
	}

	// 密码请求，需要满足minimal score，线上通常不会有这个请求，线下有
	if bytes.HasSuffix(request, []byte("mysql_native_password")) && mark < 0.85 {
		return -1, 0, nil
	}

	if m.EnableCursor && m.MaxMatchedIndex < maxScoreIndex {
		m.MaxMatchedIndex = maxScoreIndex
	}
	m.Visited[maxScoreIndex] = true
	return maxScoreIndex, mark, session.CallOutbounds[maxScoreIndex]
}

func getRequests(session *replaying.Session) [][]byte {
	keys := make([][]byte, len(session.CallOutbounds))
	for i, entry := range session.CallOutbounds {
		//keys[i] = entry.Request
		keys[i] = sortRequestParams(entry.Request)
	}
	return keys
}

func cutToChunks(key []byte, unit int) [][]byte {
	// 对Request内的参数进行排序
	key = sortRequestParams(key)
	if unit <= 0 {
		unit = 1
	}
	for len(key) < unit && unit > 1 {
		unit = unit >> 1
	}
	chunks := [][]byte{}
	key, chunks = cutHttpRequestToChunks(key, chunks)

	if len(key) > 256 {
		offset := 0
		for {
			strikeStart, strikeLen := findReadableChunk(key[offset:])
			if strikeStart == -1 {
				break
			}
			if strikeLen > (unit >> 1) {
				firstChunkLen := strikeLen
				if firstChunkLen > unit {
					firstChunkLen = unit
				}
				chunks = append(chunks, key[offset+strikeStart:offset+strikeStart+firstChunkLen])
				key = key[offset+strikeStart+firstChunkLen:]
				break
			}
			offset += strikeStart + strikeLen
		}
	}
	chunkCount := len(key) / unit
	for i := 0; i < len(key)/unit; i++ {
		chunks = append(chunks, key[i*unit:(i+1)*unit])
	}
	lastChunk := key[chunkCount*unit:]
	if len(lastChunk) > 0 {
		chunks = append(chunks, lastChunk)
	}
	return chunks
}

// sortRequestParams 对HTTP参数排序，解决因两次请求参数顺序不一致而导致匹配失败的情况
func sortRequestParams(key []byte) []byte {
	requestMarkDiff := &replayed.Diff{A: helper.BytesToString(key)}
	pairs, _, _, pro := requestMarkDiff.ParseProtocol(requestMarkDiff.A)
	switch pro {
	case protocol.HTTP_PRO:
		if len(pairs) > 0 {
			key, _ = protocol.SortHTTP(key)
		}
	}
	return key
}

// findReadableChunk returns: the starting index of the trunk, length of the trunk
func findReadableChunk(key []byte) (int, int) {
	start := bytes.IndexFunc(key, func(r rune) bool {
		return r > 31 && r < 127
	})
	if start == -1 {
		return -1, -1
	}
	end := bytes.IndexFunc(key[start:], func(r rune) bool {
		return r <= 31 || r >= 127
	})
	if end == -1 {
		return start, len(key) - start
	}
	return start, end
}

func (m *Matcher) HasBetterScore(score int, index int, maxScore int) bool {
	m.Lock()
	defer m.Unlock()

	hasBetterScore := score > maxScore
	// add a new coefficient to reduce the weight of reqs already matched
	if hasBetterScore && m.Visited[index] {
		hasBetterScore = int(float64(score)*0.9) > maxScore
	}

	return hasBetterScore
}

// cutHttpRequestToChunks 切分http请求
func cutHttpRequestToChunks(request []byte, chunks [][]byte) ([]byte, [][]byte) {
	matchs := httpRegex.FindSubmatch(request)
	if len(matchs) <= 1 {
		return request, chunks
	}

	// case GET xxx, GET xxx?a=1&b=2
	request = bytes.TrimLeft(request, string(matchs[1]))
	s := bytes.Split(matchs[1], []byte("?"))

	if len(s) >= 1 {
		chunks = append(chunks, s[0])
	}

	if len(s) >= 2 {
		params := bytes.Split(s[1], []byte("&"))
		for _, param := range params {
			chunks = append(chunks, param)
		}
	}

	return request, chunks
}
