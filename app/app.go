package app

import (
	"context"
	"database/sql"

	"github.com/jordanp/goapp/cache"
	"github.com/jordanp/goapp/pkg/auth"
	"github.com/jordanp/goapp/pkg/log"
	"github.com/jordanp/goapp/store"
	"github.com/pkg/errors"
)

// VERSION is the app-global version string, which should be substituted with a
// real value during build. https://github.com/thockin/go-build-template
var (
	VERSION string
)

type Application struct {
	log log.Logger
	db  *sql.DB

	TokenManager auth.TokenManager
	UserStore    *store.User
	CompanyStore *store.Company
	UserCache    *cache.User
}

func NewApplication(log log.Logger, config *Config) (*Application, error) {
	tokenManager, err := auth.NewTokenManager(config.secretKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize token manager")
	}

	db, err := NewDBConnection(config.dataSourceName)
	if err != nil {
		return nil, err
	}

	userStore, err := store.NewUserStore(log.F("component", "userstore"), db)
	if err != nil {
		return nil, err
	}

	companyStore, err := store.NewCompanyStore(log.F("component", "companystore"), db)
	if err != nil {
		return nil, err
	}

	userCache, err := cache.NewUserCache(log.F("component", "usercache"), userStore)
	if err != nil {
		return nil, err
	}

	// admin/admin backdoor/init
	userStore.Add(context.Background(), "admin", "$2y$10$CpVqJK/usJ8K8musmkaM1u3K7agJ0m/YOGQPLuwiBZ1M15cDHbkcu", "admin@goapp", "admin")

	return &Application{
		log: log, db: db,
		TokenManager: tokenManager,
		UserStore:    userStore, CompanyStore: companyStore,
		UserCache: userCache,
	}, nil
}

func NewDBConnection(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open DB connection")
	}

	if err = db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping DB")
	}

	return db, nil
}

func (a *Application) Stop() {
	a.UserCache.Stop()

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.log.WithError(err).Error("failed to close DB connection")
		}
	}
}
