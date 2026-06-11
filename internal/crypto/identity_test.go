package crypto

import (
	"bytes"
	"testing"
)

func TestNewIdentity(t *testing.T) {
	id, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}

	var zero [32]byte
	if id.EncryptionPub == zero {
		t.Fatal("public key must not be zero")
	}
	if id.EncryptionPriv == zero {
		t.Fatal("private key must not be zero")
	}
	if id.Fingerprint == "" {
		t.Fatal("fingerprint must not be empty")
	}

	id2, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if id.EncryptionPub == id2.EncryptionPub {
		t.Fatal("two identities must have different public keys")
	}
}

func TestLockUnlockIdentity(t *testing.T) {
	id, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	password := []byte("test-password")

	locked, err := LockIdentity(id, password)
	if err != nil {
		t.Fatal(err)
	}
	if len(locked.Salt) == 0 || len(locked.Nonce) == 0 || len(locked.Ciphertext) == 0 {
		t.Fatal("locked identity must have non-empty salt, nonce, and ciphertext")
	}

	unlocked, err := UnlockIdentity(locked, password)
	if err != nil {
		t.Fatal(err)
	}
	if unlocked.EncryptionPriv != id.EncryptionPriv {
		t.Fatal("private key mismatch after unlock")
	}
	if unlocked.EncryptionPub != id.EncryptionPub {
		t.Fatal("public key mismatch after unlock")
	}
	if unlocked.Fingerprint != id.Fingerprint {
		t.Fatal("fingerprint mismatch after unlock")
	}
}

func TestUnlockIdentityWrongPassword(t *testing.T) {
	id, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	locked, err := LockIdentity(id, []byte("correct"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = UnlockIdentity(locked, []byte("wrong"))
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
}

func TestLockProducesDistinctCiphertexts(t *testing.T) {
	id, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	password := []byte("same-password")

	l1, err := LockIdentity(id, password)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := LockIdentity(id, password)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(l1.Salt, l2.Salt) {
		t.Fatal("two locks must have different salts")
	}
	if bytes.Equal(l1.Ciphertext, l2.Ciphertext) {
		t.Fatal("two locks must produce different ciphertexts")
	}
}