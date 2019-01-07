package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jordanp/goapp/pkg/handlers"
	"github.com/jordanp/goapp/pkg/log"
	"github.com/jordanp/goapp/pkg/middlewares"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (a *Application) Routes() http.Handler {
	r := mux.NewRouter()
	r.Use(func(h http.Handler) http.Handler { return middlewares.With(middlewares.WithRecover)(h.ServeHTTP) })
	r.HandleFunc("/status", handlers.Status(VERSION))
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/token/access", a.GetAccessToken()).Methods(http.MethodPost)
	r.HandleFunc("/token/admin", a.GetAdminToken()).Methods(http.MethodPost)

	admin := r.PathPrefix("/admin").Subrouter()
	adminOnly := middlewares.MakeAuthenticator(a.TokenManager, "admin")
	admin.Use(func(h http.Handler) http.Handler { return middlewares.With(adminOnly)(h.ServeHTTP) })
	admin.HandleFunc("/users/new", a.CreateUser).Methods(http.MethodPost)
	admin.HandleFunc("/users/all", a.GetAllUsers).Methods(http.MethodGet)
	admin.HandleFunc("/users/{id}", a.DeleteUser).Methods(http.MethodDelete)
	admin.HandleFunc("/companies/new", a.CreateCompany).Methods(http.MethodPost)
	admin.HandleFunc("/companies/{id}", a.GetCompany).Methods(http.MethodGet)
	admin.HandleFunc("/companies/{id}", a.DeleteCompany).Methods(http.MethodDelete)

	user := r.PathPrefix("/users").Subrouter()
	userOnly := middlewares.MakeAuthenticator(a.TokenManager, "access")
	user.Use(func(h http.Handler) http.Handler { return middlewares.With(userOnly)(h.ServeHTTP) })
	user.HandleFunc("/me", a.Me).Methods(http.MethodGet)

	logger := middlewares.MakeLogger(a.log, log.RequestAll)
	cors := middlewares.MakeCORS()
	metrics := middlewares.MakeMetrics(nil, "", nil)
	// We don't use `Use()` here because middlewares only execute on route matches
	// and we also want to log HTTP 404 and add the OPTIONS method on all routes
	// See https://github.com/gorilla/mux/issues/416#issuecomment-434662071
	return logger(cors(metrics(r.ServeHTTP)))
}
