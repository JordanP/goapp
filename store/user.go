package store

import (
	"context"
	"database/sql"

	"github.com/jordanp/goapp/entity"
	"github.com/jordanp/goapp/pkg/log"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type User struct {
	log log.Logger
	db  *sql.DB
}

// https://medium.com/@beld_pro/postgres-with-golang-3b788d86f2ef
// The store must return human friendly, safe error messages. Failure details must be logged but never
// returned to the caller.

func NewUserStore(log log.Logger, db *sql.DB) (*User, error) {
	// Exec executes a query without returning any rows.
	if _, err := db.Exec(createTableUsers); err != nil {
		return nil, errors.Wrap(err, "failed to create users table")
	}
	return &User{log: log, db: db}, nil
}

func (s *User) DeleteAll() error {
	if _, err := s.db.Exec(deleteAllUsers); err != nil {
		return errors.Wrap(err, "failed to truncate users table")
	}
	return nil
}

func (s *User) GetByLogin(ctx context.Context, login string) (entity.User, error) {
	var user entity.User
	filter := map[string]interface{}{"login": login}
	querySuffix, parsedArgs := buildWhere(filter)
	err := getOne(ctx, s.db, selectUser+querySuffix, parsedArgs, &user.ID, &user.Login, &user.Password, &user.Email, &user.Role, &user.CreatedAt)
	if err == ErrNoRows {
		return user, NewNotFoundError("user", login)
	}
	return user, err // err is either nil or ErrGenericDBFailure
}

func (s *User) DeleteByID(ctx context.Context, id string) error {
	filter := map[string]interface{}{"id": id}
	querySuffix, parsedArgs := buildWhere(filter)
	err := deleteOne(ctx, s.db, deleteUser+querySuffix, parsedArgs)
	if err == ErrNoRows {
		return NewNotFoundError("user", id)
	}
	return err // Either nil or ErrGenericDBFailure
}

func (s *User) GetAll(ctx context.Context) ([]entity.User, error) {
	rows, err := s.db.QueryContext(ctx, selectAllUsers)
	if err != nil {
		log.G(ctx).WithError(err).Error("failed to list users in DB")
		return nil, ErrGenericDBFailure
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		err = rows.Scan(&user.ID, &user.Login, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			log.G(ctx).WithError(err).Error("failed to scan user in DB")
			return nil, ErrGenericDBFailure
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.G(ctx).WithError(err).Error("failed to loop through users list")
		return nil, ErrGenericDBFailure
	}

	return users, nil
}

func (s *User) Add(ctx context.Context, login, password, email, role string) (entity.User, error) {
	var user entity.User
	err := s.db.QueryRowContext(ctx, insertUser, login, password, email, role).Scan(&user.ID, &user.Login, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		if err2, ok := err.(*pq.Error); ok && err2.Code == ErrUniqViolation {
			if err2.Constraint == "unq_login" {
				return user, NewAlreadyExistsError("login", login)
			} else if err2.Constraint == "unq_email" {
				return user, NewAlreadyExistsError("email", email)
			}
		}
		log.G(ctx).WithError(err).Error("failed to insert user in DB")
		return user, ErrGenericDBFailure

	}
	return user, nil
}
