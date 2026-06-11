package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32

	SaltLen = 32
)

// GenerateSalt returns a 32-byte cryptographically random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// DeriveUnlockKey derives a 32-byte AES key from the user's unlock password and salt.
// The HKDF step adds explicit domain separation so the same Argon2id output
// can safely produce distinct keys for different purposes in the future.
func DeriveUnlockKey(password, salt []byte) []byte {
	master := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	r := hkdf.New(sha256.New, master, salt, []byte("cowbird-unlock-v1"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		panic(err) // unreachable: HKDF does not fail with valid inputs
	}
	return key
}