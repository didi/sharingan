package template

import (
	"html/template"
	"net/http"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/conf"
)

var Handler *template.Template

func Init() {
	Handler = template.Must(template.ParseGlob(conf.Root + "/template/*.html"))
}

func Render(w http.ResponseWriter, name string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return Handler.ExecuteTemplate(w, name, data)
}
