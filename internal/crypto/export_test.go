package crypto

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestExportImportKey(t *testing.T) {
	id, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	passphrase := []byte("my-recovery-passphrase")

	data, err := ExportKey(id, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	var exported ExportedKey
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatal("exported data is not valid JSON:", err)
	}
	if exported.Version != exportVersion {
		t.Fatalf("expected version %d, got %d", exportVersion, exported.Version)
	}

	imported, err := ImportKey(data, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	if imported.EncryptionPriv != id.EncryptionPriv {
		t.Fatal("private key mismatch after import")
	}
	if imported.EncryptionPub != id.EncryptionPub {
		t.Fatal("public key mismatch after import")
	}
	if imported.Fingerprint != id.Fingerprint {
		t.Fatal("fingerprint mismatch after import")
	}
}

func TestImportWrongPassphrase(t *testing.T) {
	id, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	data, err := ExportKey(id, []byte("correct"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = ImportKey(data, []byte("wrong"))
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}

func TestExportProducesDistinctFiles(t *testing.T) {
	id, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	passphrase := []byte("same")

	d1, err := ExportKey(id, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := ExportKey(id, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(d1, d2) {
		t.Fatal("two exports must produce different files (distinct salts)")
	}
}