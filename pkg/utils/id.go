package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// NewRandomID returns a random hex string suitable for file names.
func NewRandomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
