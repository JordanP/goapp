package app

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jordanp/goapp/entity"
	pkglog "github.com/jordanp/goapp/pkg/log"
	"github.com/jordanp/goapp/store"
	"github.com/pkg/errors"
)

func (a *Application) CreateCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := pkglog.G(ctx)

	var company entity.Company
	if err := json.NewDecoder(r.Body).Decode(&company); err != nil {
		WriteBadRequestError(w, "unable to decode json: %s", err)
		return
	}
	if err := company.Validate(); err != nil {
		WriteBadRequestError(w, "input validation error: %s", err)
		return
	}

	log = log.F("company", company.Name)
	userIDs := make([]uuid.UUID, len(company.Users))
	for k := range company.Users {
		userIDs[k] = company.Users[k].ID
	}
	insertedCompany, err := a.CompanyStore.Add(pkglog.WithLogger(ctx, log), company.Name, userIDs)
	if err != nil {
		switch errors.Cause(err).(type) {
		case *store.AlreadyExistsError, *store.NotFoundError, *store.DupUserInCompanyError:
			WriteUnprocessableEntity(w, err)
		default:
			WriteInternalServerError(w, err)
		}
		return
	}

	insertedCompany.Users = company.Users
	log.Info("company inserted")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(insertedCompany)
}

func (a *Application) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := a.CompanyStore.DeleteByID(ctx, mux.Vars(r)["id"]) // Gorilla Mux will match route iff 'id' is not empty
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

func (a *Application) GetCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	company, err := a.CompanyStore.GetByID(ctx, mux.Vars(r)["id"]) // Gorilla Mux will match route iff 'name' is not empty
	if err != nil {
		switch errors.Cause(err).(type) {
		case *store.NotFoundError:
			WriteNotFoundError(w, err)
		default:
			WriteInternalServerError(w, err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(company)
}
