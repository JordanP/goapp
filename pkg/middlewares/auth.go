package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/jordanp/goapp/pkg/auth"
	"github.com/jordanp/goapp/pkg/log"
)

var (
	bearerRegex = regexp.MustCompile(`^\s*Bearer\s+([^\s]+)\s*$`)
)

type ctxUser struct{}

func MakeAuthenticator(t auth.TokenManager, kind string) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			matches := bearerRegex.FindStringSubmatch(r.Header.Get("Authorization"))

			if len(matches) != 2 {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Authorization header is empty or doesn't start with 'Bearer '\n"))
				return
			}

			var user auth.Who
			var err error
			if kind == "access" {
				user, err = t.ParseAccessToken(matches[1])
			} else if kind == "admin" {
				user, err = t.ParseAdminToken(matches[1])
			} else {
				panic(fmt.Sprintf("unknown kind %s", kind))
			}

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error() + "\n"))
				return
			}

			ctx := context.WithValue(r.Context(), ctxUser{}, user)
			ctx = log.WithLogger(ctx, log.G(ctx).F("who", user.Who()))
			h(w, r.WithContext(ctx))
		}
	}
}

func AdminUserFromCtx(ctx context.Context) auth.AdminUser {
	return ctx.Value(ctxUser{}).(auth.AdminUser)
}

func UserFromCtx(ctx context.Context) auth.User {
	return ctx.Value(ctxUser{}).(auth.User)
}
