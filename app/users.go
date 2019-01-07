package app

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jordanp/goapp/entity"
	pkglog "github.com/jordanp/goapp/pkg/log"
	"github.com/jordanp/goapp/pkg/middlewares"
	"github.com/jordanp/goapp/store"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func (a *Application) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := a.UserStore.GetAll(ctx)
	if err != nil {
		WriteInternalServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(entity.Users{Users: users})
}

func (a *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := a.UserStore.DeleteByID(ctx, mux.Vars(r)["id"]) // Gorilla Mux will match route iff 'id' is not empty
	if err != nil {
		switch errors.Cause(err).(type) {
		case *store.NotFoundError:
			WriteNotFoundError(w, err)
		default:
			WriteInternalServerError(w, err)
		}
		return
	}
}

func (a *Application) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := pkglog.G(ctx)

	var user entity.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteBadRequestError(w, "unable to decode json: %s", err)
		return
	}
	if err := user.Validate(); err != nil {
		WriteBadRequestError(w, "input validation error: %s", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.WithError(err).Error("failed to hash password with Bcrypt")
		WriteInternalServerError(w, "failed to hash password with Bcrypt")
		return
	}

	log = log.F("login", user.Login, "email", user.Email, "role", user.Role)
	insertedUser, err := a.UserStore.Add(pkglog.WithLogger(ctx, log), user.Login, string(hashedPassword), user.Email, user.Role)
	if err != nil {
		switch errors.Cause(err).(type) {
		case *store.AlreadyExistsError:
			WriteUnprocessableEntity(w, err)
		default:
			WriteInternalServerError(w, err)
		}
		return
	}

	log.Info("user inserted")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(insertedUser)
}

func (a *Application) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middlewares.UserFromCtx(ctx)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(user)
}
