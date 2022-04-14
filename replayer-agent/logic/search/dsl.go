package search

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/didi/sharingan/replayer-agent/idl"
	"github.com/didi/sharingan/replayer-agent/model/nuwaplt"
)

type DSL struct {
	Size   int                    `json:"size"`
	From   int                    `json:"from"`
	Query  map[string]interface{} `json:"query"`
	Source []string               `json:"_source,omitempty"`
	Sort   map[string]interface{} `json:"sort"`
}

func toDSL(req *idl.SearchReq) ([]byte, error) {
	dsl := new(DSL)

	// page limit: from, size
	if req.Page <= 0 {
		req.Page = 1
	}
	dsl.Size = req.Size
	if req.Size <= 0 {
		dsl.Size = 20
	}
	dsl.From = (req.Page - 1) * dsl.Size

	if len(req.Field) != 0 {
		dsl.Source = req.Field
	}

	dsl.Sort = newCond("CallFromInbound.OccurredAt", "order", "desc")

	must := make([]interface{}, 0)
	mustNot := make([]interface{}, 0)

	// filter out sessions without ReturnInbound
	must = append(must, newCond("exists", "field", "ReturnInbound"))

	if p, ok := nuwaplt.Module2Host[req.Project]; ok {
		var conds []interface{}
		for _, s := range strings.Split(p, ",") {
			conds = append(conds, newCond("match_phrase", "Context", s))
		}
		must = append(must, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               conds,
				"minimum_should_match": 1,
			},
		})
	}

	if req.InboundRequest != "" {
		must, mustNot = unionMulti(must, mustNot, "CallFromInbound.Request", req.InboundRequest, parseCond)
	}

	if req.InboundResponse != "" {
		must, mustNot = unionMulti(must, mustNot, "ReturnInbound.Response", req.InboundResponse, parseCond)
	}

	if req.OutboundRequest != "" {
		must, mustNot = unionMulti(must, mustNot, "Actions.Request", req.OutboundRequest, parseCond)
	}

	if req.OutboundResponse != "" {
		must, mustNot = unionMulti(must, mustNot, "Actions.Response", req.OutboundResponse, parseCond)
	}

	if req.Apollo != "" {
		must, mustNot = unionMulti(must, mustNot, "Actions.Content", req.Apollo, parseApollo)
	}

	if req.SessionId != "" {
		must = append(must, newCond("term", "SessionId", req.SessionId))
	}

	must = unionTime(must, req.Start, req.End)

	dsl.Query = map[string]interface{}{
		"bool": map[string]interface{}{
			"must":     must,
			"must_not": mustNot,
		},
	}

	return json.Marshal(dsl)
}

func unionMulti(must []interface{}, mustNot []interface{}, field string, words string,
	parseFunc func(field string, word string) map[string]interface{}) ([]interface{}, []interface{}) {

	for _, word := range strings.Fields(words) {
		word = strings.ToLower(word)
		if word[0] == '!' {
			if len(word) > 1 {
				mustNot = append(mustNot, parseFunc(field, word[1:]))
			}
			continue
		}
		must = append(must, parseFunc(field, word))
	}
	return must, mustNot
}

func unionTime(must []interface{}, start, end string) []interface{} {
	if start == "" && end == "" {
		return must
	}

	timeRange := map[string]int64{}
	if start != "" {
		st, _ := parseTime(start)
		timeRange["gte"] = st.UnixNano()
	}
	if end != "" {
		ed, hasTime := parseTime(end)
		if !hasTime {
			ed = ed.Add(24 * time.Hour)
		}
		timeRange["lt"] = ed.UnixNano()
	}

	must = append(must, map[string]interface{}{
		"range": map[string]interface{}{
			"CallFromInbound.OccurredAt": timeRange,
		},
	})
	return must
}

func parseTime(str string) (time.Time, bool) {
	hasTime := true
	if strings.Index(str, "T") < 0 {
		str += "T00:00:00"
		hasTime = false
	}

	layout := "2006-01-02T15:04:05"
	t, err := time.ParseInLocation(layout, str, time.Local)
	if err != nil {
		return time.Now(), hasTime
	}
	return t, hasTime
}

func parseCond(field string, word string) map[string]interface{} {
	if strings.ContainsAny(word, "?*") {
		return newCond("wildcard", field, word)
	}
	return newCond("match_phrase", field, word)
}

func parseApollo(field string, word string) map[string]interface{} {
	sp := strings.Split(word, "=")
	if len(sp) == 2 {
		return newCond("match_phrase", field, fmt.Sprintf("%s %s %s", sp[0], sp[1], sp[0]))
	}
	return newCond("match_phrase", field, word)
}

func newCond(keyword, field, word string) map[string]interface{} {
	return map[string]interface{}{
		keyword: map[string]interface{}{
			field: word,
		},
	}
}
