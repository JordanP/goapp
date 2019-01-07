package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password,omitempty"`
	Email     string    `json:"email"`
	Role      string    `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (u User) Validate() error {
	if u.Login == "" {
		return errors.New("missing or empty 'login'")
	}
	if u.Password == "" {
		return errors.New("missing or empty 'password'")
	}
	if u.Email == "" {
		return errors.New("missing or empty 'email'")
	}
	if u.Role == "" {
		return errors.New("missing or empty 'role'")
	}
	return nil
}

type Users struct {
	Users []User `json:"users"`
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u UserCredentials) Validate() error {
	if u.Login == "" {
		return errors.New("missing or empty 'login'")
	}
	if u.Password == "" {
		return errors.New("missing or empty 'password'")
	}
	return nil
}
