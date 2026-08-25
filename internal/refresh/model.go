package refresh

import (
	"time"

	"github.com/google/uuid"
)

type Refresh struct {
	UserID    uuid.UUID
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	Revoked   bool
}
