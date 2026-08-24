package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Surname   string
	Email     string
	CreatedAt time.Time
	Role      string
}
