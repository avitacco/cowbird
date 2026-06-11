package vault

import (
	"context"

	"cowbird/internal/crypto"
)

// GetLockedIdentity retrieves the user's encrypted private key from Vault.
// Returns sharing.ErrNotFound if no identity has been stored yet (first run).
func (v *Vault) GetLockedIdentity(ctx context.Context) (*crypto.LockedIdentity, error) {
	var locked crypto.LockedIdentity
	if _, err := v.kvRead(ctx, "users/"+v.EntityID+"/identity", &locked); err != nil {
		return nil, err
	}
	return &locked, nil
}

// PutLockedIdentity stores the user's encrypted private key in Vault.
func (v *Vault) PutLockedIdentity(ctx context.Context, locked *crypto.LockedIdentity) error {
	_, err := v.kvWrite(ctx, "users/"+v.EntityID+"/identity", locked)
	return err
}