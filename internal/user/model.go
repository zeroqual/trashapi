package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	PasswordHash string
	Name         string
	Surname      string
	Email        string
	CreatedAt    time.Time
	Role         string
}

// тестовое это будем менять
type UserUpdate struct {
	Name    string
	Surname string
}
