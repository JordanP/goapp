package main

import (
	"flag"
	"os"
	"time"

	"github.com/jordanp/goapp/app"
	"github.com/jordanp/goapp/pkg/graceful"
	pkglog "github.com/jordanp/goapp/pkg/log"
)

// export VERSION=$(git describe --tags --always --dirty)
func main() {
	secretKey := flag.String("secretKey", os.Getenv("SECRET_KEY"), "JWT secret key")
	sqlDSN := flag.String("sqlDSN", os.Getenv("SQL_DSN"), "SQL connection string")
	flag.Parse()

	log := pkglog.New("mygoapp", app.VERSION, pkglog.DebugLevel)
	config := app.NewConfig(*secretKey, *sqlDSN)
	log.Infof("starting application with: %s", config)
	app, err := app.NewApplication(log, config)
	if err != nil {
		log.Fatal(err)
	}

	listenAndServe := graceful.MakeListenAndServe(log, time.Second)
	if err := listenAndServe(":2000", app.Routes()); err != nil {
		// Don't call `Fatal()` here since we still want to stop the app.
		log.WithError(err).Error("listenAndServe")
	}

	app.Stop()
	log.Info("application stopped")
}
