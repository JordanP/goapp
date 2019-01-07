package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jordanp/goapp/entity"
	"github.com/jordanp/goapp/pkg/log"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type Company struct {
	log log.Logger
	db  *sql.DB
}

// https://medium.com/@beld_pro/postgres-with-golang-3b788d86f2ef
// The store must return human friendly, safe error messages. Failure details must be logged but never
// returned to the caller.

func NewCompanyStore(log log.Logger, db *sql.DB) (*Company, error) {
	// Exec executes a query without returning any rows.
	if _, err := db.Exec(createTableCompanies); err != nil {
		return nil, errors.Wrap(err, "failed to create companies table")
	}
	return &Company{log: log, db: db}, nil
}

func (s *Company) Add(ctx context.Context, name string, userIDs []uuid.UUID) (entity.Company, error) {
	var company entity.Company

	tx, err := s.db.Begin()
	if err != nil {
		log.G(ctx).WithError(err).Error("failed to begin transaction")
		return company, ErrGenericDBFailure
	}

	if err := tx.QueryRowContext(ctx, insertCompany, name).Scan(&company.ID, &company.Name, &company.CreatedAt); err != nil {
		tx.Rollback()
		if err, ok := err.(*pq.Error); ok && err.Code == ErrUniqViolation && err.Constraint == "unq_name" {
			return company, NewAlreadyExistsError("company", name)
		}
		log.G(ctx).WithError(err).Error("failed to insert company in DB")
		return company, ErrGenericDBFailure
	}

	for k := range userIDs {
		if err := addUserInCompany(ctx, tx, userIDs[k], company.ID); err != nil {
			tx.Rollback()
			return company, err
		}
	}
	if err := tx.Commit(); err != nil {
		log.G(ctx).WithError(err).Error("failed to commit transaction")
		return company, ErrGenericDBFailure
	}

	return company, nil
}

func (s *Company) DeleteByID(ctx context.Context, ID string) error {
	filter := map[string]interface{}{"ID": ID}
	querySuffix, parsedArgs := buildWhere(filter)
	err := deleteOne(ctx, s.db, deleteCompany+querySuffix, parsedArgs)
	if err == ErrNoRows {
		return NewNotFoundError("company", ID)
	}
	return err // Either nil or ErrGenericDBFailure
}

func (s *Company) GetByID(ctx context.Context, ID string) (entity.Company, error) {
	var company entity.Company
	filter := map[string]interface{}{"id": ID}
	querySuffix, parsedArgs := buildWhere(filter)
	err := getOne(ctx, s.db, selectCompany+querySuffix, parsedArgs, &company.ID, &company.Name, &company.CreatedAt)
	if err != nil {
		if err == ErrNoRows {
			return company, NewNotFoundError("company", ID)
		}
		return company, err
	}

	rows, err := s.db.QueryContext(ctx, selectCompanyUsers, ID)
	if err != nil {
		log.G(ctx).WithError(err).Error("failed to get company in DB")
		return company, ErrGenericDBFailure
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		if err = rows.Scan(&user.ID, &user.Login, &user.Email, &user.Role, &user.CreatedAt); err != nil {
			log.G(ctx).WithError(err).Error("failed to scan company users in DB")
			return company, ErrGenericDBFailure
		}
		company.Users = append(company.Users, user)
	}

	if err = rows.Err(); err != nil {
		log.G(ctx).WithError(err).Error("failed to loop through company users")
		return company, ErrGenericDBFailure
	}

	return company, nil
}

func (s *Company) DeleteAll() error {
	if _, err := s.db.Exec(deleteAllCompanies); err != nil {
		return errors.Wrap(err, "failed to truncate companies table")
	}
	return nil
}

func addUserInCompany(ctx context.Context, querier Querier, userID, companyID uuid.UUID) error {
	_, err := querier.ExecContext(ctx, insertUserInCompany, companyID, userID)
	if err == nil {
		return nil
	}

	if err, ok := err.(*pq.Error); ok {
		if err.Code == ErrFKViolation && err.Constraint == "users_companies_user_id_fkey" {
			return NewNotFoundError("user", userID.String())
		} else if err.Code == ErrFKViolation && err.Constraint == "users_companies_company_id_fkey" {
			return NewNotFoundError("company", companyID.String())
		} else if err.Code == ErrUniqViolation && err.Constraint == "unq_set" {
			return NewDupUserInCompanyError(userID.String())
		}
	}

	log.G(ctx).WithError(err).Error("failed to insert user in company")
	return ErrGenericDBFailure
}

type DupUserInCompanyError struct {
	User string
}

func (e *DupUserInCompanyError) Error() string {
	return fmt.Sprintf("duplicate user '%s' in company", e.User)
}

func NewDupUserInCompanyError(user string) *DupUserInCompanyError {
	return &DupUserInCompanyError{User: user}
}
