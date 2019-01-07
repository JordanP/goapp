package middlewares

import (
	"net/http"

	"github.com/rs/cors"
)

func MakeCORS() Middleware {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "GET", "DELETE"},
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "Authorization"},
		MaxAge:           3600,
		AllowCredentials: true,
	})
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cors.Handler(h).ServeHTTP(w, r)
		}
	}
}
