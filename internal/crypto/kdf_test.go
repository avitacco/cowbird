package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	s1, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	if len(s1) != SaltLen {
		t.Fatalf("expected %d bytes, got %d", SaltLen, len(s1))
	}
	s2, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(s1, s2) {
		t.Fatal("two salts must not be equal")
	}
}

func TestDeriveUnlockKey(t *testing.T) {
	password := []byte("hunter2")
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}

	k1 := DeriveUnlockKey(password, salt)
	if len(k1) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(k1))
	}

	k2 := DeriveUnlockKey(password, salt)
	if !bytes.Equal(k1, k2) {
		t.Fatal("same inputs must produce the same key")
	}

	other := DeriveUnlockKey([]byte("different"), salt)
	if bytes.Equal(k1, other) {
		t.Fatal("different passwords must produce different keys")
	}
}