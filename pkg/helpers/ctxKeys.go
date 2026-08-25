package helpers

import "github.com/google/uuid"

type contextKey struct{}

var UserKey contextKey

type UserContext struct {
	UserID   uuid.UUID
	UserRole string
}
