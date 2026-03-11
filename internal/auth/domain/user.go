package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	PrepUserID int64
	Email      string
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewUser(prepUserID int64, email, name string) *User {
	return &User{
		ID:         uuid.Must(uuid.NewV7()),
		PrepUserID: prepUserID,
		Email:      email,
		Name:       name,
	}
}
