package replayer

import (
	"context"
	"net/http"
)

func AddHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const traceIDKey = "Sharingan-Replayer-TraceID"
		if s := r.Header.Get(traceIDKey); s != "" {
			r = r.WithContext(context.WithValue(r.Context(), traceIDKey, s))
		}
		next.ServeHTTP(w, r)
	})
}
