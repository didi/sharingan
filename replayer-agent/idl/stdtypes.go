package idl

type ReplayedReq struct {
	Trace   string `json:"trace"`
	RetType string `json:"retType"`
	Project string `json:"project"`
}

type ReplayedResp struct {
	Success bool        `json:"success"`
	Errmsg  string      `json:"errmsg"`
	Diffs   interface{} `json:"diffs,omitempty"`
}

type SearchReq struct {
	// request params
	Project          string   `json:"project"`
	InboundRequest   string   `json:"inbound_request"`
	InboundResponse  string   `json:"inbound_response"`
	OutboundRequest  string   `json:"outbound_request"`
	OutboundResponse string   `json:"outbound_response"`
	Apollo           string   `json:"apollo"`
	SessionId        string   `json:"session_id"`
	Page             int      `json:"page"`
	Size             int      `json:"size"`
	Field            []string `json:"field"`
	Date             []string `json:"date"`
	// for heuristic search
	Heuristic []string `json:"heuristic"`
	// inferred fields
	Start string `json:"-"`
	End   string `json:"-"`
	Extend interface{} `json:"extend"`
}

type SearchResp struct {
	Results  interface{} `json:"results"`
	Parallel int         `json:"parallel"`
	Errmsg   string      `json:"errmsg"`
}
