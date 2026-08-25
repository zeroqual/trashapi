package helpers

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashToken(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}
