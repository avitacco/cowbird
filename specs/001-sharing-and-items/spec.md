# Feature Specification: Item Sharing, Item Types, and Key Recovery

**Feature Branch**: `001-sharing-and-items`
**Status**: Draft
**Created**: 2026-06-02

## Summary

Cowbird users need to store a variety of credential types (not just logins),
share individual items with other users so recipients can read and use them,
and have a way to recover access if they lose their device. All item contents
must remain readable only by their owner and explicitly authorized recipients,
never by the operator of the storage backend.

## User Scenarios

### User Story 1 — Store different kinds of items (Priority: P1)

A user saves not only website logins but also payment cards, secure notes,
identity details, and standalone passwords, and can add arbitrary extra fields
to any of them or create a fully freeform item.

**Why this priority**: Core value of a password manager beyond a single login
type; everything else builds on having items to store and share.

**Acceptance Scenarios**:

1. **Given** a user is signed in, **When** they create a login with username,
   password, and one or more URLs, **Then** the item is saved and reappears
   with the same fields on next launch.
2. **Given** a user is creating any item type, **When** they add a custom field
   with a chosen label and value, **Then** that field is stored and displayed
   alongside the type's standard fields.
3. **Given** a user needs to store something that fits no standard type,
   **When** they create a custom item with their own fields, **Then** it is
   saved and listed like any other item.

### User Story 2 — Share an item with another user (Priority: P1)

A user shares one of their items with another cowbird user. The recipient sees
the shared item in their own list and can read its contents.

**Why this priority**: Sharing is the headline capability distinguishing this
work from single-user storage.

**Acceptance Scenarios**:

1. **Given** users Paul and Aubry both use cowbird, **When** Paul shares an
   item with Aubry, **Then** Aubry's client shows the item in her list and she
   can read its contents.
2. **Given** Paul has shared an item with Aubry, **When** Paul edits the item,
   **Then** Aubry sees the updated contents without Paul re-sharing.
3. **Given** an item shared with Aubry, **When** a third user who was not given
   access inspects the storage backend, **Then** they cannot read the item's
   contents.

### User Story 3 — Revoke a shared item (Priority: P2)

A user who previously shared an item removes the recipient's access.

**Acceptance Scenarios**:

1. **Given** Paul shared an item with Aubry, **When** Paul revokes access,
   **Then** the item is removed from Aubry's list.
2. **Given** access has been revoked, **When** Aubry attempts to read the
   previously shared content, **Then** she can no longer decrypt it.

### User Story 4 — Recover access after device loss (Priority: P2)

A user exports their private key to a safe location and later restores it on a
new device to regain access to all their items.

**Acceptance Scenarios**:

1. **Given** a signed-in user, **When** they choose to export their key and
   provide a passphrase, **Then** they receive a passphrase-protected key file.
2. **Given** a user on a new device with their exported key file, **When** they
   import it with the correct passphrase, **Then** they regain access to their
   own items and any items still shared with them.
3. **Given** a user who never exported their key and lost their device,
   **When** they attempt recovery, **Then** the system clearly indicates there
   is no recovery path. (No operator-side reset exists by design.)

### User Story 5 — Rotate keys or change unlock password (Priority: P3)

A user changes the password that unlocks their key, or rotates their underlying
key after a suspected compromise.

**Acceptance Scenarios**:

1. **Given** a signed-in user, **When** they change their unlock password,
   **Then** their items remain accessible and no item contents are re-encrypted
   unnecessarily.
2. **Given** a user who suspects key compromise, **When** they rotate their key,
   **Then** all their items are re-secured under the new key and their published
   public key is updated for future shares.

## Requirements

### Functional Requirements

- **FR-001**: System MUST support distinct item types: login, card, note,
  identity, password, and custom.
- **FR-002**: System MUST allow arbitrary user-defined custom fields (label,
  type, value) on any item type.
- **FR-003**: System MUST encrypt all item contents such that only the owner and
  explicitly authorized recipients can read them; the storage operator MUST NOT
  be able to read item contents.
- **FR-004**: System MUST allow a user to share an individual item with another
  user, after which the recipient can read it.
- **FR-005**: Edits to a shared item MUST become visible to recipients without
  re-sharing.
- **FR-006**: System MUST allow the owner to revoke a recipient's access, after
  which the recipient can no longer read the item.
- **FR-007**: A recipient's client MUST be able to discover items newly shared
  with it without scanning all shared data in the system.
- **FR-008**: System MUST allow a user to export their private key in a
  passphrase-protected form.
- **FR-009**: System MUST allow a user to restore access on a new device by
  importing their exported key.
- **FR-010**: System MUST allow a user to change their unlock password without
  re-encrypting all item contents.
- **FR-011**: System MUST allow a user to rotate their key, re-securing their
  items and updating their published public key.
- **FR-012**: System MUST make clear to the user that exporting their key is the
  only recovery mechanism and that no operator-side reset exists.

### Key Entities

- **Item**: A stored secret of a given type, with standard and custom fields.
- **Share**: An authorization granting a specific recipient read access to a
  specific item.
- **Recipient**: A user who has been granted access to a shared item.
- **Identity / Key material**: The per-user keys establishing who a user is and
  enabling encryption to and from them.

## Assumptions

- The storage operator is semi-trusted: technically prevented from reading item
  contents, but assumed not to actively forge user identities. (See research.md
  for the trust-model decision.)
- Users are a known, finite set within a deployment (a company or a family).
- Each deployment runs its own storage backend.

## Out of Scope

- Browser extension clients (future work).
- Email alias generation (depends on external services cowbird does not run).
- File attachments on items.
- Defending against an actively malicious operator forging public keys
  (explicitly deferred; see research.md).
