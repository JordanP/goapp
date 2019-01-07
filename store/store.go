package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jordanp/goapp/pkg/log"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	ErrFKViolation               = "23503"
	ErrUniqViolation             = "23505"
	ErrInvalidTextRepresentation = "22P02"
)

var (
	ErrGenericDBFailure = errors.New("DB error")
	ErrNoRows           = sql.ErrNoRows
)

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func buildWhere(args map[string]interface{}) (querySuffix string, parsedArgs []interface{}) {
	if len(args) == 0 {
		return
	}

	querySuffix += " WHERE "
	parsedArgs = make([]interface{}, 0, len(args))
	columnNo := 1
	for columnName, columnValue := range args {
		querySuffix += fmt.Sprintf("%s = $%d", columnName, columnNo)
		parsedArgs = append(parsedArgs, columnValue)
		if columnNo < len(args) {
			querySuffix += " AND "
		}
		columnNo++
	}

	return
}

func deleteOne(ctx context.Context, db *sql.DB, query string, args []interface{}) error {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		if err2, ok := err.(*pq.Error); ok && err2.Code == ErrInvalidTextRepresentation {
			return ErrNoRows
		}
		log.G(ctx).F(args).WithError(err).Error("failed to delete record")
		return ErrGenericDBFailure
	}

	if n, err := result.RowsAffected(); err != nil {
		log.G(ctx).F(args).WithError(err).Error("failed to delete record")
		return ErrGenericDBFailure
	} else if n == 0 {
		return ErrNoRows
	} else {
		log.G(ctx).F(args).Warnf("deleted %d records", n)
	}

	return nil
}

func getOne(ctx context.Context, db *sql.DB, query string, args []interface{}, dest ...interface{}) error {
	if err := db.QueryRowContext(ctx, query, args...).Scan(dest...); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoRows
		}
		if err, ok := err.(*pq.Error); ok && err.Code == ErrInvalidTextRepresentation {
			return ErrNoRows
		}
		log.G(ctx).F(args).WithError(err).Error("failed to get record")
		return ErrGenericDBFailure
	}

	return nil
}

type NotFoundError struct {
	Kind string
	ID   string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s '%s' not found", e.Kind, e.ID)
}

func NewNotFoundError(kind string, ID string) *NotFoundError {
	return &NotFoundError{Kind: kind, ID: ID}
}

type AlreadyExistsError struct {
	Kind string
	ID   string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s '%s' already exists", e.Kind, e.ID)
}

func NewAlreadyExistsError(kind string, ID string) *AlreadyExistsError {
	return &AlreadyExistsError{Kind: kind, ID: ID}
}
