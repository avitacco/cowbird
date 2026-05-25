package encryption

import (
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
)

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func DeriveEncKey(password, salt []byte) []byte {
	master := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	key := make([]byte, 32)
	r := hkdf.New(sha256.New, master, salt, []byte("cowbird-enc-v1"))
	if _, err := io.ReadFull(r, key); err != nil {
		panic(err)
	}
	return key
}
