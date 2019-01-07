package middlewares

import "net/http"

type Middleware = func(http.HandlerFunc) http.HandlerFunc

// With chains middleware
// h := With(midd1, midd2, middl3) is equivalent to h := midd1(midd2(midd3))
// Note: When h will be called with h(w, *r), midd1 will be executed first !
func With(ms ...Middleware) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		for i := len(ms) - 1; i >= 0; i-- {
			h = ms[i](h)
		}
		return h
	}
}
