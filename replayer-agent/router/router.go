package router

import (
	"context"
	"net/http"
	"runtime/debug"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didichuxing/sharingan/replayer-agent/controller"

	"github.com/julienschmidt/httprouter"
)

var Handler *httprouter.Router

func Init() {
	Handler = httprouter.New()
	globalRouter := Handler

	// httprouter Handle signature
	globalRouter.HandlerFunc(http.MethodGet, "/", RecoveryHF(new(controller.ShaRinGan).Index))
	globalRouter.HandlerFunc(http.MethodGet, "/autoreplay/", RecoveryHF(new(controller.ShaRinGan).AutoReplay))
	globalRouter.HandlerFunc(http.MethodPost, "/search/", RecoveryHF(new(controller.ShaRinGan).Search))
	globalRouter.HandlerFunc(http.MethodPost, "/noise/", RecoveryHF(new(controller.ShaRinGan).Noise))
	globalRouter.HandlerFunc(http.MethodPost, "/noise/del", RecoveryHF(new(controller.ShaRinGan).DelNoise))
	globalRouter.HandlerFunc(http.MethodPost, "/xxd", RecoveryHF(new(controller.ShaRinGan).Xxd))

	// platform apis
	globalRouter.HandlerFunc(http.MethodGet, "/platform/get/dsl", RecoveryHF(new(controller.ShaRinGan).PlatformGetDsl))
	globalRouter.HandlerFunc(http.MethodPost, "/platform/post/dsl", RecoveryHF(new(controller.ShaRinGan).PlatformPostDsl))
	globalRouter.HandlerFunc(http.MethodGet, "/platform/module/names", RecoveryHF(new(controller.ShaRinGan).PlatformModules))

	// code coverage
	globalRouter.HandlerFunc(http.MethodGet, "/coverage", RecoveryHF(new(controller.ShaRinGan).CodeCoverage))
	globalRouter.Handle("GET", "/coverage/report/:covFile", RecoveryHL(new(controller.ShaRinGan).CodeCoverageReport))

	// idl
	globalRouter.HandlerFunc(http.MethodPost, "/diffbinary", RecoveryHF(new(controller.ShaRinGan).DiffBinary))

	// restful
	globalRouter.Handle("GET", "/replay/:sessionId", RecoveryHL(new(controller.ShaRinGan).Replay))
	globalRouter.Handle("GET", "/replayed/:sessionId", RecoveryHL(new(controller.ShaRinGan).Replayed))
	globalRouter.Handle("GET", "/session/:sessionId", RecoveryHL(new(controller.ShaRinGan).Session))

	//FileServer
	globalRouter.Handler("GET", "/public/*filepath", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
}

// RecoveryHF ...
func RecoveryHF(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "PANIC:%s\n%s", err, debug.Stack())
				http.Error(w, "500 Server internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, req)
	}
}

// RecoveryHL ...
func RecoveryHL(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		defer func() {
			if err := recover(); err != nil {
				tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "PANIC:%s\n%s", err, debug.Stack())
				http.Error(w, "500 Server internal error", http.StatusInternalServerError)
			}
		}()
		next(w, req, ps)
	}
}
