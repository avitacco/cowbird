package crypto

import (
	"bytes"
	"testing"
)

func TestWrapUnwrapKey(t *testing.T) {
	recipient, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	itemKey, err := NewItemKey()
	if err != nil {
		t.Fatal(err)
	}

	ephPub, nonce, wrapped, err := WrapKey(recipient.EncryptionPub, itemKey)
	if err != nil {
		t.Fatal(err)
	}

	unwrapped, err := UnwrapKey(recipient.EncryptionPriv, ephPub, nonce, wrapped)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(itemKey, unwrapped) {
		t.Fatal("unwrapped key does not match original")
	}
}

func TestWrapProducesDistinctCiphertexts(t *testing.T) {
	recipient, _ := NewIdentity()
	itemKey, _ := NewItemKey()

	_, _, w1, err := WrapKey(recipient.EncryptionPub, itemKey)
	if err != nil {
		t.Fatal(err)
	}
	_, _, w2, err := WrapKey(recipient.EncryptionPub, itemKey)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(w1, w2) {
		t.Fatal("two wraps of the same key must produce different ciphertexts (ephemeral keys)")
	}
}

func TestUnwrapWrongKey(t *testing.T) {
	recipient, _ := NewIdentity()
	attacker, _ := NewIdentity()
	itemKey, _ := NewItemKey()

	ephPub, nonce, wrapped, _ := WrapKey(recipient.EncryptionPub, itemKey)

	_, err := UnwrapKey(attacker.EncryptionPriv, ephPub, nonce, wrapped)
	if err == nil {
		t.Fatal("expected error unwrapping with wrong private key")
	}
}