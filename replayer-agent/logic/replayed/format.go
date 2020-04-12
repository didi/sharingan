package replayed

import (
	"fmt"
	"strconv"

	"github.com/didichuxing/sharingan/replayer-agent/model/protocol"
	"github.com/didichuxing/sharingan/replayer-agent/utils/helper"
)

type CommonItem struct {
	Project string `json:"project"`
	Request string `json:"request"`
}

type PairItem struct {
	Key    string `json:"key"`
	Online string `json:"online"`
	Test   string `json:"test"`
}

type DiffItem struct {
	SubRequest string     `json:"sub_request"`
	Pairs      []PairItem `json:"pairs"`
}

type FailedItem struct {
	CommonItem
	SessionId string     `json:"session_id"`
	Url       string     `json:"url"`
	Diffs     []DiffItem `json:"diff"`
}

type SuccessItem struct {
	CommonItem
	SubRequests []string `json:"sub_request"`
}

func QDEFormat(sid string, isSuccess bool, diffs []*DiffRecord) interface{} {
	if len(diffs) == 0 {
		return nil
	}

	if isSuccess {
		return formatSuccess(diffs)
	}
	return formatFailed(sid, diffs)
}

func formatFailed(sid string, diffs []*DiffRecord) FailedItem {
	f := FailedItem{}
	f.Project = diffs[0].Project
	f.Request = diffs[0].RequestMark
	f.SessionId = sid
	f.Url = fmt.Sprintf("http://%s:"+helper.PortVal+"/replay/%s", helper.LocalIp, sid)
	for _, d := range diffs {
		di := DiffItem{SubRequest: strconv.Quote(d.Noise)}
		if d.IsDiff != 2 {
			continue
		}
		for _, f := range d.FormatDiff {
			if f.Is != 2 {
				continue
			}
			if len(f.Children) != 2 {
				continue
			}
			ol, ok := f.Children[0]["label"]
			if !ok {
				continue
			}
			tl, ok := f.Children[1]["label"]
			if !ok {
				continue
			}
			di.Pairs = append(di.Pairs, PairItem{Key: f.Key, Online: ol.(string), Test: tl.(string)})
		}
		f.Diffs = append(f.Diffs, di)
	}
	return f
}

func formatSuccess(diffs []*DiffRecord) SuccessItem {
	s := SuccessItem{}
	s.Project = diffs[0].Project
	s.Request = diffs[0].RequestMark
	for _, d := range diffs {
		if len(d.Noise) == 0 {
			continue
		}
		s.SubRequests = append(s.SubRequests, strconv.Quote(d.Noise))
	}
	return s
}

// optimizeDisReq 优化前端展示，如去除乱码字符等
func optimizeDisReq(protocolReq string, matchReq []byte) string {
	if protocolReq == protocol.MYSQL_PRO {
		if len(matchReq) >= 5 {
			start := 4
			for index, bValue := range matchReq {
				if index >= 5 && bValue > 0x1f && bValue < 0x7f {
					start = index
					break
				}
			}
			// 无内容case
			if start == 4 {
				switch matchReq[4] {
				case 0x0E:
					return "COM_PING"
				case 0x16:
					return "COM_STMT_PREPARE"
				case 0x17:
					return "COM_STMT_EXECUTE"
				case 0x19:
					return "COM_STMT_CLOSE"
				}
			}

			return helper.BytesToString(matchReq[start:])
		}
	}

	return string(matchReq)
}
