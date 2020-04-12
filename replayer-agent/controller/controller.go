package controller

import (
	"encoding/json"
	"net/http"
)

//BaseController 基础Controller
type BaseController struct {
}

// EchoJSON json格式输出
func (bs *BaseController) EchoJSON(w http.ResponseWriter, r *http.Request, body interface{}) {
	if cType := w.Header().Get("Content-Type"); cType == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	b, err := json.Marshal(body)
	if err != nil {
		bs.Echo(w, r, []byte(`{"errno":1, "errmsg":"`+err.Error()+`"}`))
	} else {
		bs.Echo(w, r, b)
	}
}

// Echo 原始输出,包含tracelog
func (bs *BaseController) Echo(w http.ResponseWriter, req *http.Request, body []byte) {
	if cType := w.Header().Get("Content-Type"); cType == "" {
		w.Header().Set("Content-Type", "text/plain")
	}
	w.Write(body)
}
