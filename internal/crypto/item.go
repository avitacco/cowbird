package crypto

import (
	"crypto/rand"
	"fmt"
)

const ItemKeyLen = 32

// NewItemKey generates a random 32-byte AES-256-GCM item key.
func NewItemKey() ([]byte, error) {
	key := make([]byte, ItemKeyLen)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generating item key: %w", err)
	}
	return key, nil
}