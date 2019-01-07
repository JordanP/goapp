package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/jordanp/goapp/pkg/log"
)

func WithRecover(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if log := log.G(r.Context()); log != nil {
					log.Error(err)
					log.Error(string(debug.Stack()))
				}
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h(w, r)
	}
}
