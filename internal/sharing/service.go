package sharing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cowbird/internal/crypto"
	"cowbird/internal/items"
)

// Service coordinates item creation, sharing, and revocation for a single user.
type Service struct {
	entityID string
	identity *crypto.Identity
	store    Store
}

// NewService returns a Service for entityID backed by store.
// identity must be unlocked (EncryptionPriv populated).
func NewService(entityID string, identity *crypto.Identity, store Store) *Service {
	return &Service{entityID: entityID, identity: identity, store: store}
}

// CreateItem encrypts content and stores it in the owner's own subtree.
func (s *Service) CreateItem(ctx context.Context, content items.Content) (Envelope, error) {
	env, _, err := NewEnvelope(s.entityID, s.identity.EncryptionPub, content)
	if err != nil {
		return Envelope{}, fmt.Errorf("creating envelope: %w", err)
	}
	if err := s.store.PutItem(ctx, env.ID, env); err != nil {
		return Envelope{}, fmt.Errorf("storing item: %w", err)
	}
	return env, nil
}

// OpenOwnItem decrypts an item from the owner's own subtree.
func (s *Service) OpenOwnItem(_ context.Context, env Envelope) (items.Content, error) {
	wk, ok := findOwnerKey(env, s.entityID)
	if !ok {
		return nil, fmt.Errorf("no wrapped key for %s in item %s", s.entityID, env.ID)
	}
	return OpenEnvelope(env, s.identity.EncryptionPriv, wk)
}

// Share shares itemID with recipientID.
//
// It decrypts the owner's item key, wraps it for the recipient, writes a shared
// envelope to the shared namespace (retaining the owner's wrapped key so the
// owner can still decrypt it), and drops a share message in the recipient's inbox.
func (s *Service) Share(ctx context.Context, itemID, recipientID string) error {
	env, err := s.store.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("getting item %s: %w", itemID, err)
	}

	ownerWK, ok := findOwnerKey(env, s.entityID)
	if !ok {
		return fmt.Errorf("no wrapped key for owner in item %s", itemID)
	}
	itemKey, err := crypto.UnwrapKey(s.identity.EncryptionPriv, ownerWK.EphemeralPub, ownerWK.Nonce, ownerWK.Wrapped)
	if err != nil {
		return fmt.Errorf("unwrapping item key: %w", err)
	}

	recipientPub, err := s.store.GetPublicKey(ctx, recipientID)
	if err != nil {
		return fmt.Errorf("getting public key for %s: %w", recipientID, err)
	}
	recipientWK, err := WrapKeyForRecipient(itemKey, recipientID, recipientPub)
	if err != nil {
		return fmt.Errorf("wrapping key for recipient: %w", err)
	}
	recipientWKBytes, err := marshalWrappedKey(recipientWK)
	if err != nil {
		return err
	}

	// Write the shared envelope. It is a shallow copy of the original — same
	// ciphertext, owner's wrapped key in Recipients; recipient's key is NOT here.
	shareID := newID()
	sharedEnv := env
	sharedEnv.ID = shareID

	version, err := s.store.PutSharedEnvelope(ctx, shareID, sharedEnv)
	if err != nil {
		return fmt.Errorf("writing shared envelope: %w", err)
	}

	msg := Message{
		Type:       MessageShare,
		ShareID:    shareID,
		SenderID:   s.entityID,
		EnvVersion: version,
		Timestamp:  time.Now().UTC(),
		Share: &SharePayload{
			SharePath:  sharePath(s.entityID, shareID),
			WrappedKey: recipientWKBytes,
			ItemType:   string(env.Type),
			OwnerID:    s.entityID,
		},
	}
	if err := s.store.SendMessage(ctx, recipientID, newID(), msg); err != nil {
		return fmt.Errorf("sending share message to %s: %w", recipientID, err)
	}
	return nil
}

// Revoke removes a recipient's access to a shared item.
//
// It deletes the shared envelope (the real security action — the ciphertext is
// gone) and drops a revoke message so the recipient's client can remove the
// stale SharedLink. recipientID is required to route the revoke message.
func (s *Service) Revoke(ctx context.Context, shareID, recipientID string) error {
	if err := s.store.DeleteSharedEnvelope(ctx, shareID); err != nil {
		return fmt.Errorf("deleting shared envelope %s: %w", shareID, err)
	}

	msg := Message{
		Type:      MessageRevoke,
		ShareID:   shareID,
		SenderID:  s.entityID,
		Timestamp: time.Now().UTC(),
	}
	if err := s.store.SendMessage(ctx, recipientID, newID(), msg); err != nil {
		return fmt.Errorf("sending revoke message to %s: %w", recipientID, err)
	}
	return nil
}

// ProcessInbox reads all pending inbox messages and applies them.
//
// For share messages: writes a SharedLink, then deletes the message.
// For revoke messages: removes the SharedLink, then deletes the message.
//
// Each handler writes/removes the link before deleting the message so that a
// crash between the two steps is self-healing — the next ProcessInbox call
// will re-process the still-present message.
func (s *Service) ProcessInbox(ctx context.Context) error {
	entries, err := s.store.ListInboxMessages(ctx)
	if err != nil {
		return fmt.Errorf("listing inbox: %w", err)
	}
	for _, entry := range entries {
		if err := s.processEntry(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

// OpenSharedItem fetches the shared envelope identified by link and decrypts it
// using the recipient's wrapped key stored in the link.
func (s *Service) OpenSharedItem(ctx context.Context, link SharedLink) (items.Content, error) {
	wk, err := unmarshalWrappedKey(link.WrappedKey)
	if err != nil {
		return nil, fmt.Errorf("parsing wrapped key in link %s: %w", link.ShareID, err)
	}

	ownerID, shareID, err := parseSharePath(link.SharePath)
	if err != nil {
		return nil, fmt.Errorf("parsing share path %q: %w", link.SharePath, err)
	}

	env, _, err := s.store.GetSharedEnvelope(ctx, ownerID, shareID)
	if err != nil {
		return nil, fmt.Errorf("getting shared envelope: %w", err)
	}

	return OpenEnvelope(env, s.identity.EncryptionPriv, wk)
}

func (s *Service) processEntry(ctx context.Context, entry InboxEntry) error {
	switch entry.Msg.Type {
	case MessageShare:
		return s.processShare(ctx, entry)
	case MessageRevoke:
		return s.processRevoke(ctx, entry)
	default:
		return fmt.Errorf("unknown message type %q (share %s)", entry.Msg.Type, entry.Msg.ShareID)
	}
}

func (s *Service) processShare(ctx context.Context, entry InboxEntry) error {
	if entry.Msg.Share == nil {
		return fmt.Errorf("share message %s missing payload", entry.Msg.ShareID)
	}

	// Idempotency: if we already have a link at this version or newer, skip writing
	// and just clean up the message.
	existing, err := s.store.GetSharedLink(ctx, entry.Msg.ShareID)
	if err == nil && existing.EnvVersion >= entry.Msg.EnvVersion {
		return s.store.DeleteInboxMessage(ctx, entry.ID)
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("checking existing link for share %s: %w", entry.Msg.ShareID, err)
	}

	link := SharedLink{
		ShareID:    entry.Msg.ShareID,
		SharePath:  entry.Msg.Share.SharePath,
		WrappedKey: entry.Msg.Share.WrappedKey,
		OwnerID:    entry.Msg.Share.OwnerID,
		ItemType:   entry.Msg.Share.ItemType,
		EnvVersion: entry.Msg.EnvVersion,
	}
	// Write link first — if we crash before deleting the message, the next
	// ProcessInbox call will re-enter and the idempotency check will skip the write.
	if err := s.store.PutSharedLink(ctx, link); err != nil {
		return fmt.Errorf("writing shared link for share %s: %w", entry.Msg.ShareID, err)
	}
	return s.store.DeleteInboxMessage(ctx, entry.ID)
}

func (s *Service) processRevoke(ctx context.Context, entry InboxEntry) error {
	// Remove link first — if we crash before deleting the message, the next
	// ProcessInbox call will re-enter; ErrNotFound on the link is swallowed.
	if err := s.store.DeleteSharedLink(ctx, entry.Msg.ShareID); err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("removing shared link for revoke %s: %w", entry.Msg.ShareID, err)
	}
	return s.store.DeleteInboxMessage(ctx, entry.ID)
}

func sharePath(ownerID, shareID string) string {
	return ownerID + "/" + shareID
}

func parseSharePath(path string) (ownerID, shareID string, err error) {
	i := strings.IndexByte(path, '/')
	if i < 0 {
		return "", "", fmt.Errorf("invalid share path %q: expected ownerID/shareID", path)
	}
	return path[:i], path[i+1:], nil
}