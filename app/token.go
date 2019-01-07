package app

import (
	"encoding/json"
	"net/http"

	"github.com/jordanp/goapp/entity"
	"github.com/jordanp/goapp/pkg/auth"
	pkglog "github.com/jordanp/goapp/pkg/log"
	"github.com/jordanp/goapp/store"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidRole = errors.New("you don't have the admin role")
)

func (a *Application) GetAccessToken() http.HandlerFunc {
	return a.getToken(func(login, email, role string) (string, error) {
		return a.TokenManager.GenerateAccessToken(auth.NewUser(login, email, role))
	})
}

func (a *Application) GetAdminToken() http.HandlerFunc {
	return a.getToken(func(login, email, role string) (string, error) {
		if role != "admin" {
			return "", ErrInvalidRole
		}
		return a.TokenManager.GenerateAdminToken(auth.NewAdminUser(login))
	})
}
func (a *Application) getToken(tokenGen func(login, email, role string) (string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := pkglog.G(ctx)

		var creds entity.UserCredentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			WriteBadRequestError(w, "unable to decode json: %s", err)
			return
		}
		if err := creds.Validate(); err != nil {
			WriteBadRequestError(w, "input validation error: %s", err)
			return
		}

		log = log.F("login", creds.Login)
		user, err := a.UserStore.GetByLogin(pkglog.WithLogger(ctx, log), creds.Login)
		if err != nil {
			fake := []byte("$2y$10$LoxBCn5Q1tNROmao8acYE..b3m4Yvw83HjnE4m6oum.At0FX2ICUW")
			bcrypt.CompareHashAndPassword(fake, []byte(creds.Password)) // Avoid timing attack
			switch errors.Cause(err).(type) {
			case *store.NotFoundError:
				WriteUnauthorizedError(w, "invalid credentials")
			default:
				WriteInternalServerError(w, err)
			}
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
		if err != nil {
			switch err {
			case bcrypt.ErrMismatchedHashAndPassword:
				WriteUnauthorizedError(w, "invalid credentials")
			default:
				log.WithError(err).Error("failed to hash and compare password")
				WriteInternalServerError(w, "failed to hash and compare password")
			}
			return
		}

		token, err := tokenGen(user.Login, user.Email, user.Role)
		if err != nil {
			switch err {
			case ErrInvalidRole:
				log.Debug("invalid role")
				WriteForbiddenError(w, ErrInvalidRole)
			default:
				log.WithError(err).Error("failed to generate token")
				WriteInternalServerError(w, "failed to generate token")
			}
			return
		}

		log.Info("token generated")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(entity.Token{Token: token})
	}
}
