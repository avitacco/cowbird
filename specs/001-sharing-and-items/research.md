# Research & Decisions: Item Sharing, Item Types, and Key Recovery

**Branch**: `001-sharing-and-items` | **Plan**: [plan.md](./plan.md)

This document records the design decisions reached during discussion, with the
alternatives considered and the rationale for each choice.

## Decision 1: Sharing via envelope encryption with per-item content keys

**Decision**: Each item has its own random symmetric key encrypting its content.
That item key is wrapped per recipient using the recipient's public key.

**Rationale**: A password-derived symmetric key is single-recipient and cannot
support sharing. Envelope encryption lets content be encrypted once and access
granted by wrapping the item key to additional public keys, without
re-encrypting content. This is the standard approach (used by Proton, 1Password,
etc.).

**Alternatives rejected**: Re-encrypting content per recipient (wasteful, does
not scale); sharing the derivation password (defeats security entirely).

## Decision 2: X25519 for wrapping, Ed25519 (optional) for signing

**Decision**: Use X25519 (Diffie-Hellman) keys to wrap item keys. Ed25519 is for
optional authorship signing, deferred.

**Rationale**: Ed25519 is a signing key, not an encryption key. Wrapping needs a
DH-capable key (X25519). NaCl `box` / `age` implement the wrap-to-public-key
primitive directly. Signing proves authorship, which only matters under an
adversarial-operator model (deferred), so it is optional for now.

## Decision 3: Soft (semi-trusted operator) trust model, not strict zero-trust

**Decision**: Treat the storage operator as semi-trusted: technically unable to
read contents, but assumed not to forge public keys. Drop mandatory out-of-band
key verification.

**Rationale**: Both target use cases have an accountable, non-adversarial
operator:
- Corporate: a dedicated Vault team whose job is secure operation, and who
  already hold more valuable secrets than cowbird stores.
- Personal: the operator is the user themselves; defending against a malicious
  operator means defending against oneself.

Strict zero-trust requires mandatory fingerprint verification, whose friction
(e.g. "call your coworker to read four words before sharing") is exactly what
causes users to skip it, yielding the appearance of zero-trust without the
substance. The protections that cost users nothing (content encryption) are
kept; the one that imposes friction for a threat that does not exist
(verification) is dropped.

**Consequence**: A malicious operator could substitute a public key on first
contact. This is explicitly accepted, not solved. The underlying machinery is
the same as strict mode, so an optional "verify this contact" affordance can be
added later without redesign.

## Decision 4: Public keys stored in Vault; private keys never leave the client

**Decision**: Publish public keys to a Vault directory (source of truth,
distributed to all instances). Private keys are stored encrypted (Argon2id-
derived key) and never sent to the operator in plaintext.

**Rationale**: Public keys are not secret; storing them in Vault gives
persistence and cross-instance distribution. Storing them in Vault does NOT make
them tamper-proof against the operator (the operator controls the store), but
under the soft trust model that is acceptable.

**Clarification recorded**: Encrypting a bare public key buys no confidentiality
(it is meant to be public). However, a user's *private record of which keys they
have pinned* IS sensitive (it reveals sharing relationships) and is therefore
stored encrypted-to-self. Authenticated encryption (XChaCha20-Poly1305) on that record also
makes post-pinning tampering detectable (it fails to decrypt). This protects the
pin after it is made; it does not vouch for the key's authenticity at first
fetch (see Decision 3).

## Decision 5: Local TOFU change detection deferred

**Decision**: Do not keep a local pinned-key copy for change detection in v1.

**Rationale**: TOFU change detection needs an independent local reference to
compare against; storing pins only in Vault means the reference and the checked
value share a source, so silent changes cannot be detected. This is consistent
with the soft trust model. Deferred as a possible later enhancement (store the
authoritative key in Vault plus a tiny local fingerprint note purely for change
detection).

## Decision 6: Shared items in a shared namespace (not copied per recipient)

**Decision**: Shared item envelopes live in a dedicated shared namespace, not
duplicated into each recipient's subtree.

**Rationale**: Copying into recipient trees creates N copies to keep in sync on
every edit, which is error-prone. A single shared envelope means edits are
visible to all recipients with no propagation. The cost (a recipient must read a
path outside their own subtree) is handled by broad read access on the shared
namespace, which is safe because contents are encrypted and access is gated by
the wrapped key, not the path ACL.

## Decision 7: Discovery via consume-and-delete inbox, durable state in own store

**Decision**: A per-recipient inbox carries transient share/revoke messages. The
recipient's client consumes each message and projects it into a durable
SharedLink in its own subtree, then deletes the message.

**Rationale**: Separates notification (event) from state (standing access
record). The inbox stays small and write-once; the authoritative list of a
recipient's shares lives in their own locked, fully-controlled subtree and is
read on startup like their own items. Avoids scanning the whole shared namespace
(which would leak share count/existence to everyone).

**Robustness requirements**:
- Handlers idempotent; apply link-first-then-delete (and remove-first-then-delete
  on revoke) so partial failures self-heal on reprocess.
- Message stream is advisory UI state; the real security boundary is the shared
  envelope's existence and keying. A missed revoke message degrades to a dead
  link, not retained access.
- Out-of-order/duplicate delivery resolved by the envelope's KV v2 version.

## Decision 8: Ordering tiebreaker = KV v2 version, not a counter or timestamp

**Decision**: Use the shared envelope's Vault-maintained KV v2 version number to
order events about the same share.

**Rationale**: A sender-maintained counter needs durable storage and races
across devices. A timestamp depends on the sender's (possibly skewed) clock.
Vault already assigns a monotonic version per path server-side, so the client
reads it rather than inventing one, and the recipient can cross-check it against
the envelope's actual current version.

## Decision 9: Vault policy verified empirically

**Decision**: The inbox access model relies on KV v2 honoring `create` vs
`update`, and on templated-vs-wildcard precedence. These were tested against the
running Vault and confirmed.

**Confirmed behaviors**:
- `create` without `update` lets a sender drop a new inbox message but not
  overwrite an existing one.
- A sender with `create` only cannot read or list the recipient's inbox.
- The recipient's templated own-inbox rule (read/delete) takes precedence over
  the create-only sender wildcard for their own inbox.

**Constraint**: KV v2 list output is not ACL-filtered and exposes key names, so
all key names (entity IDs, share IDs, message IDs) are opaque UUIDs with no
meaningful content. Owner entity IDs appearing in shared paths are acceptable
because they are opaque UUIDs, not human-readable names.

## Decision 10: Item content model = typed structs + custom fields (option B)

**Decision**: One `Content` interface with concrete typed structs (Login, Card,
Note, Identity, Password), each with a `CustomFields` slice, plus a `Custom`
type for fully freeform items.

**Rationale**: Preserves compile-time guarantees and easy autofill on known
fields, while the custom-fields slice provides Proton-like flexibility. A pure
generic sections/fields model (option C) maximizes flexibility but loses type
safety and makes autofill convention-based. Option B with custom fields captures
most of C's flexibility while keeping B's ergonomics.

**Cost accepted**: JSON cannot unmarshal into the interface directly; a small
type-dispatch decode helper is required.

## Decision 11: Recovery via user-exported key; no operator reset

**Decision**: Recovery is a user-initiated, passphrase-protected export of the
private key. Import on a new device restores access. No operator-side reset.

**Rationale**: Simpler and sufficient for loss recovery versus an
auto-generated second "emergency" identity (which would require wrapping every
item to two keys and maintaining that invariant). Matches the trust model: no
trusted party exists to reset keys. The export must be passphrase-protected so a
leaked file is useless without the passphrase, and the UI must be honest that
this is the only recovery path.

**Alternative considered**: Independent emergency keypair with wrap-to-both. More
robust for emergency *access* scenarios but unnecessary complexity for *loss*
recovery; deferred.

## Decision 12: Password change vs key rotation are distinct (Bitwarden model)

**Decision**: Separate the unlock password from the underlying encryption key.
Changing the password re-wraps key material only; rotating the key re-secures
all items.

**Rationale**: Following Bitwarden's separation, changing the password should not
require re-encrypting all content (cheap, frequent operation), while true key
rotation (for compromise) re-wraps all item keys (expensive, rare). Sessions
must be coordinated during rotation so a stale client does not write under the
old key, which Bitwarden documents as a corruption risk.
