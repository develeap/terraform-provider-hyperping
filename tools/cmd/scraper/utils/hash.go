package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashContent creates a SHA256 hash of text content
func HashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
