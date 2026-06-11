# Implementation Plan: Item Sharing, Item Types, and Key Recovery

**Branch**: `001-sharing-and-items` | **Spec**: [spec.md](./spec.md)

## Technical Context

**Language**: Go 1.26
**GUI**: Fyne (`fyne.io/fyne/v2`), app ID `co.avitac.cowbird`
**Storage backend**: HashiCorp Vault, KV v2 engine, mount `cowbird`
**Auth**: userpass / token / approle via existing `internal/auth` package
**Local credential store**: OS keyring (`99designs/keyring`)
**Config**: TOML via Viper at `~/.config/cowbird/config.toml`
**Crypto**: `golang.org/x/crypto`; envelope encryption with X25519 key
wrapping and XChaCha20-Poly1305 content encryption; Ed25519 reserved for optional
authorship signing; Argon2id for password-derived key.

## Trust Model

Soft / semi-trusted operator (see research.md for the decision record). Item
contents are end-to-end encrypted so the operator cannot read them. The
operator is trusted not to forge public keys, so mandatory out-of-band
verification is NOT required. Trust-on-first-use change detection is deferred,
not adopted.

## Architecture Overview

### Encryption (envelope model)

- Each item has a randomly generated symmetric **item key**.
- Item content is encrypted once with the item key (XChaCha20-Poly1305).
- The item key is wrapped (encrypted) to each authorized recipient's X25519
  public key. Adding a recipient = one more wrapped key; content is not
  re-encrypted.
- Private keys never leave the client in plaintext. A user's private key is
  stored encrypted (under an Argon2id-derived key) and never visible to the
  operator.

### Identity / keys

- Each user has an X25519 encryption keypair (for wrapping) and may have an
  Ed25519 signing keypair (optional authorship signing, deferred).
- Public keys are published to a directory in Vault (source of truth,
  distributed to all instances).
- The unlock password is separate from the Vault auth credential, so an operator
  credential reset cannot decrypt data.

### Vault storage layout

```
cowbird/data/users/<entityID>/items/<itemID>   # owner's own items
cowbird/data/users/<entityID>/identity         # locked (encrypted) keypair
cowbird/data/users/<entityID>/links/<shareID>  # durable SharedLink records
cowbird/data/users/<entityID>/pinned           # owner's encrypted pinned-keys record (not yet used)
cowbird/data/pubkeys/<entityID>                # public-key directory (read-all, write-own)
cowbird/data/shared/<ownerEntityID>/<shareID>  # shared item envelopes
cowbird/data/inbox/<recipientEntityID>/<msgID> # transient share/revoke messages
```

All key names are opaque UUIDs. (Vault list output is not ACL-filtered, so no
meaningful data may appear in key names.)

### Vault policy (templated per entity)

- Self subtree: full CRUD on `users/{{identity.entity.id}}/*` (data) and
  list/read/delete on the matching metadata path.
- Pubkey directory: read-all on `pubkeys/*`; create+update only on own entry.
- Shared namespace: read-all on `shared/*`; owner-templated CRUD on
  `shared/{{identity.entity.id}}/*`.
- Inbox: recipient reads/deletes own (`inbox/{{identity.entity.id}}/*`); senders
  get `create` only on `inbox/+/*` (cannot read, list, or overwrite).

Path-ACL is NOT the security boundary for shared items; the wrapped item key is.
Broad read on the shared namespace is safe only because contents are encrypted.

### Sharing protocol

- Share envelopes live in the shared namespace; each recipient's wrapped item key
  travels privately via the recipient's inbox, not in the shared blob.
- Inbox is a consume-and-delete message channel (not a durable store).
- On share: owner writes the shared envelope, drops a share message into the
  recipient's inbox. Recipient's client consumes it and writes a durable
  **SharedLink** into its own subtree.
- On revoke: owner re-keys/removes the shared envelope (the real security action)
  and drops a revoke message; recipient's client removes the link.
- Message handlers MUST be idempotent and ordered: write/remove the link first,
  then delete the message, so partial failures self-heal on reprocess.
- Ordering tiebreaker is the shared envelope's KV v2 version number, not a
  client-generated counter or timestamp.

### Key recovery and rotation

- **Recovery**: user-initiated export of the private key, passphrase-protected.
  Import on a new device restores access. No operator-side reset (by design).
- **Password change**: re-wrap key material under a new Argon2id-derived key;
  item contents are not re-encrypted.
- **Key rotation** (compromise): generate new keypair, re-wrap all item keys to
  the new key, re-publish the public key, re-distribute shares. Coordinate
  sessions so a stale client cannot write under the old key.

## Project Structure (packages)

```
internal/
├── auth/         # existing: Method interface, Userpass/Token/AppRole
├── config/       # existing: Viper TOML config
├── credentials/  # existing: OS keyring CredentialStore
├── vault/
│   ├── vault.go      # existing: Vault client, auth, token renewal, VerifyMount
│   ├── store.go      # NEW: implements sharing.Store (all 16 methods)
│   └── identity.go   # NEW: GetLockedIdentity / PutLockedIdentity
├── crypto/
│   ├── kdf.go        # Argon2id + HKDF DeriveUnlockKey, GenerateSalt
│   ├── aes.go        # XChaCha20-Poly1305 Seal / Open
│   ├── item.go       # NewItemKey (random 32-byte key)
│   ├── identity.go   # Identity, LockedIdentity, NewIdentity, Lock/Unlock
│   ├── wrap.go       # WrapKey / UnwrapKey (X25519 ECDH + HKDF + XChaCha20-Poly1305)
│   ├── export.go     # ExportKey / ImportKey (passphrase-protected JSON)
│   └── *_test.go     # 17 tests covering all functions
├── items/
│   ├── types.go      # ItemType, FieldType, Field, Content interface, all 6 concrete types
│   ├── codec.go      # Encode / Decode with type-dispatch envelope
│   └── codec_test.go # 10 tests
├── sharing/
│   ├── types.go      # WrappedKey, Envelope, Message, SharePayload, SharedLink, InboxEntry
│   ├── store.go      # Store interface (16 methods) + ErrNotFound
│   ├── envelope.go   # NewEnvelope, OpenEnvelope, WrapKeyForRecipient
│   ├── service.go    # Service, CreateItem, OpenOwnItem, Share, Revoke, ProcessInbox, OpenSharedItem
│   └── service_test.go # 5 integration tests using in-memory Store
├── core/
│   ├── app.go        # App{Vault, Identity, Service}, NewApp
│   └── core.go       # InitIdentity (first-run and returning-user paths)
└── ui/
    ├── setup.go      # existing: NewSetupWindow (first-run Vault config)
    ├── unlock.go     # NEW: NewUnlockWindow (identity create/unlock)
    └── app.go        # NewMainWindow (placeholder; shows key fingerprint)
```

## Build Order (risk-first)

1. ✅ `crypto` package in isolation, with tests.
2. ✅ `items` types + codec; `vault/store.go` + `vault/identity.go` for own-item CRUD.
3. ✅ `sharing` (envelopes, inbox, links, service); `core` and `ui/unlock.go` wired end-to-end.

## Implementation Notes

### Deviations from data-model.md

- **`InboxEntry` type added**: `Message` has no `ID` field — the Vault path
  component is the only identifier for deletion. `InboxEntry{ID string, Msg
  Message}` was introduced so `ListInboxMessages` can return the path key
  alongside each message.
- **`SharePayload.WrappedKey` and `SharedLink.WrappedKey`**: Documented as
  `[]byte`; they contain JSON-marshaled `WrappedKey` structs (carrying all three
  ECDH fields: `ephemeral_pub`, `nonce`, `wrapped`). Internal helpers
  `marshalWrappedKey`/`unmarshalWrappedKey` handle the translation.
- **`internal/encryption` deleted**: Replaced entirely by `internal/crypto`.
  The old `argon2.go` stub is gone.

### KV v2 storage format

All KV writes use `{"v": "<json-string>"}`. The `json-string` is the
JSON-marshaled form of the actual struct (Envelope, SharedLink, etc.).
This single-key wrapper avoids Vault's metadata-vs-data split leaking field
names and keeps the read/write path uniform.

### Version number type

vault-client-go's KV response parser returns raw `map[string]interface{}` for
metadata. Without `decoder.UseNumber()`, `version` arrives as `float64` and
loses precision for large integers. `versionFromMetadata` type-asserts to
`json.Number` then calls `.Int64()`.

## Open Items

- Confirmed via testing: `create`-without-`update` on inbox wildcard;
  templated-vs-wildcard precedence on own inbox; `create`-only grants no list.
- Vault policy needs updating: add `users/<eid>/identity` and
  `users/<eid>/links/*`; confirm inbox hard-delete on metadata path.
- Full item list UI (main window is a fingerprint placeholder).
- Password change and key rotation flows (crypto primitives done; service
  layer not yet written).
- Key export/import UI (`crypto.ExportKey`/`ImportKey` implemented; no UI yet).
- Local TOFU change detection: deliberately deferred.
- Optional Ed25519 authorship signing: deferred; slot reserved in `Envelope.Signature`.
