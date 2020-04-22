package bind

import (
	"net/http"
	"strings"

	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"

	"github.com/ppltools/binding"
)

func Bind(r *http.Request, req interface{}) error {
	b := binding.Default(r.Method, getContentType(r.Header.Get("Content-Type")))
	err := b.Bind(r, req)
	if err != nil {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "bind request failed||err=%s", err)
		return err
	}

	return nil
}

func getContentType(ctype string) string {
	typeSplits := strings.Split(ctype, ";")
	if len(typeSplits) > 1 {
		return typeSplits[0]
	}
	return ctype
}
