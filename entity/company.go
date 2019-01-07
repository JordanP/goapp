package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Users     []User    `json:"users,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (c Company) Validate() error {
	if c.Name == "" {
		return errors.New("missing or empty 'name'")
	}
	return nil
}
