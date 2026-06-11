package crypto

import (
	"bytes"
	"testing"
)

func TestSealOpen(t *testing.T) {
	key, err := NewItemKey()
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("hello, cowbird")

	nonce, ciphertext, err := Seal(key, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if len(nonce) == 0 || len(ciphertext) == 0 {
		t.Fatal("expected non-empty nonce and ciphertext")
	}

	got, err := Open(key, nonce, ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("decrypted %q, want %q", got, plaintext)
	}
}

func TestSealProducesDistinctCiphertexts(t *testing.T) {
	key, _ := NewItemKey()
	plaintext := []byte("same plaintext")

	_, c1, _ := Seal(key, plaintext)
	_, c2, _ := Seal(key, plaintext)
	if bytes.Equal(c1, c2) {
		t.Fatal("two Seal calls must produce distinct ciphertexts (random nonce)")
	}
}

func TestOpenWrongKey(t *testing.T) {
	key, _ := NewItemKey()
	nonce, ciphertext, _ := Seal(key, []byte("secret"))

	wrongKey, _ := NewItemKey()
	_, err := Open(wrongKey, nonce, ciphertext)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestOpenTamperedCiphertext(t *testing.T) {
	key, _ := NewItemKey()
	nonce, ciphertext, _ := Seal(key, []byte("secret"))

	ciphertext[0] ^= 0xff
	_, err := Open(key, nonce, ciphertext)
	if err == nil {
		t.Fatal("expected error decrypting tampered ciphertext")
	}
}