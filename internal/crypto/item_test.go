package crypto

import (
	"bytes"
	"testing"
)

func TestNewItemKey(t *testing.T) {
	k1, err := NewItemKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(k1) != ItemKeyLen {
		t.Fatalf("expected %d bytes, got %d", ItemKeyLen, len(k1))
	}
	k2, err := NewItemKey()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(k1, k2) {
		t.Fatal("two item keys must not be equal")
	}
}