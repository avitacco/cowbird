# Feature Specification: Item List and Editor UI

**Feature Branch**: `002-item-ui`
**Status**: Draft
**Created**: 2026-06-11

## Summary

Cowbird's main window is currently a placeholder showing the user's key
fingerprint. Users need a real working surface: a list of all items they can
read (their own and those shared with them), the ability to create items of
every supported type, a detail view that protects sensitive values by default,
and the ability to edit and delete items. This feature replaces the placeholder
with that surface, making cowbird usable as a password manager for the first
time.

The share/revoke actions themselves remain service-level only; their UI is a
follow-on feature.

## User Scenarios

### User Story 1 — See all my items in one list (Priority: P1)

A user unlocks cowbird and sees a single list of every item they can read:
items they own and items shared with them, each showing its title and type,
with shared items visibly distinguished from owned ones.

**Why this priority**: The list is the home screen and the entry point to every
other interaction; nothing else in this feature is reachable without it.

**Acceptance Scenarios**:

1. **Given** a user with stored items, **When** they unlock cowbird, **Then**
   the main window lists each item with its title and type.
2. **Given** another user has shared an item with this user, **When** the user
   unlocks cowbird, **Then** the shared item appears in the list, visibly
   marked as shared with them rather than owned.
3. **Given** a share was revoked while the user was offline, **When** they next
   unlock cowbird, **Then** the revoked item no longer appears in the list.
4. **Given** a user with no items, **When** they unlock cowbird, **Then** the
   list shows an empty state inviting them to create their first item rather
   than a blank screen.

### User Story 2 — Create an item of any type (Priority: P1)

A user creates a new item, choosing from the six supported types (login, card,
note, identity, password, custom), fills in the type's standard fields, and
optionally adds custom fields.

**Why this priority**: Without creation the list is permanently empty;
creation and listing together form the minimum usable product.

**Acceptance Scenarios**:

1. **Given** the main window, **When** the user chooses to create a login and
   fills in title, username, password, and a URL, **Then** the item is saved
   and appears in the list immediately and on next launch.
2. **Given** an item editor for any type, **When** the user adds a custom field
   with a label, value, and field kind (text / hidden / TOTP / URL), **Then**
   the field is saved with the item and shown in the detail view.
3. **Given** an item editor with a required field left empty (e.g. no title),
   **When** the user attempts to save, **Then** they get a clear validation
   message and nothing is written to storage.
4. **Given** the user cancels out of the editor, **Then** nothing is written to
   storage.

### User Story 3 — View details without leaking secrets (Priority: P1)

A user selects an item from the list and sees its full contents, with sensitive
values (passwords, CVV, PIN, hidden custom fields) masked by default. They can
reveal a value or copy it to the clipboard without revealing it.

**Why this priority**: Reading a stored credential back is the core daily-use
loop of a password manager; masking-by-default is table stakes for using it in
view of other people.

**Acceptance Scenarios**:

1. **Given** an item in the list, **When** the user selects it, **Then** a
   detail view shows all standard and custom fields, with sensitive values
   masked.
2. **Given** a masked field, **When** the user toggles reveal, **Then** the
   plaintext value is shown until toggled back or the view is left.
3. **Given** any field, **When** the user activates copy, **Then** the value is
   placed on the system clipboard without needing to reveal it first.
4. **Given** an item shared with the user (not owned), **When** they view it,
   **Then** all fields are readable but no edit or delete action is offered.

### User Story 4 — Edit an item (Priority: P2)

A user edits an item they own. Changes persist, and if the item is shared, the
recipients see the updated contents without any re-sharing step.

**Acceptance Scenarios**:

1. **Given** an item the user owns, **When** they edit a field and save,
   **Then** the change is visible in the detail view and survives a restart.
2. **Given** an item the user has shared with someone, **When** the owner edits
   it, **Then** the recipient sees the updated contents on their next refresh
   without the owner re-sharing.
3. **Given** an open editor with unsaved changes, **When** the user cancels,
   **Then** the stored item is unchanged.

### User Story 5 — Delete an item (Priority: P2)

A user permanently deletes an item they own, confirming first since there is no
undo.

**Acceptance Scenarios**:

1. **Given** an item the user owns, **When** they delete it and confirm,
   **Then** it disappears from the list and from storage.
2. **Given** the confirmation prompt, **When** the user declines, **Then** the
   item is untouched.
3. **Given** a deleted item that had been shared, **When** a recipient next
   refreshes, **Then** the item no longer appears in their list (a stale link
   to a missing envelope is cleaned up, not shown as an error).

### User Story 6 — Find an item quickly (Priority: P3)

A user with many items filters the list by typing part of a title, or narrows
it to a single type.

**Acceptance Scenarios**:

1. **Given** a list of many items, **When** the user types a search string,
   **Then** the list narrows to items whose title contains it
   (case-insensitive).
2. **Given** an active filter, **When** the user clears it, **Then** the full
   list returns.

## Requirements

### Functional Requirements

- **FR-001**: The main window MUST list all items the user can read — owned
  items and shared links — showing at minimum each item's title and type.
- **FR-002**: The client MUST process the user's inbox before first populating
  the list, so newly shared and newly revoked items are reflected at startup,
  and MUST provide a manual refresh that re-processes the inbox and reloads
  the list.
- **FR-003**: Shared items MUST be visibly distinguished from owned items, and
  MUST be read-only: no edit or delete affordance is offered on them.
- **FR-004**: Users MUST be able to create items of all six types (login, card,
  note, identity, password, custom) with each type's standard fields.
- **FR-005**: Users MUST be able to add, edit, and remove custom fields (label,
  kind, value) on any item type in the editor.
- **FR-006**: The detail view MUST mask sensitive values (password-like
  standard fields and custom fields of the hidden kind) by default, with a
  per-field reveal toggle.
- **FR-007**: Users MUST be able to copy any field value to the system
  clipboard without revealing it on screen.
- **FR-008**: Users MUST be able to edit items they own; saved edits MUST be
  visible to existing recipients of a shared item without re-sharing.
- **FR-009**: Users MUST be able to delete items they own, behind an explicit
  confirmation that states deletion is permanent.
- **FR-010**: An item whose content fails to decrypt or decode MUST be shown as
  an unreadable entry (with its ID/type if known) without crashing the app or
  hiding the rest of the list.
- **FR-011**: Storage errors (e.g. Vault unreachable) MUST surface as a
  user-readable message with the option to retry, never as a silent empty
  list.
- **FR-012**: The list SHOULD support filtering by title substring and by item
  type (P3).

### Service-layer gaps to close

The UI consumes `sharing.Service`, which currently lacks list/update/delete.
The underlying `sharing.Store` already provides most primitives
(`ListItems`, `PutItem`, `DeleteItem`, `ListSharedLinks`,
`PutSharedEnvelope`, `DeleteSharedEnvelope`), so this feature adds:

- `ListItems` — list the user's own envelopes.
- `ListSharedLinks` — list the user's durable shared links.
- `UpdateItem` — re-encrypt content under the existing item key (fresh nonce)
  and write the owned envelope back, then rewrite each shared envelope made
  from the item, so recipients see edits without re-sharing and their wrapped
  keys remain valid.
- `DeleteItem` — delete the owned envelope; for each outstanding share of the
  item, delete the shared envelope (the revoke security action) and notify the
  recipient.

`Share` currently records nothing on the owner's side (each share is an
independent envelope copy), so update/delete cannot find the shared envelopes
they must rewrite or remove. This feature therefore also introduces an
owner-side **ShareRecord** (`users/<eid>/shares/<shareID>`, written by `Share`,
removed by `Revoke`/`DeleteItem`) — see [plan.md](./plan.md) for the design.

### Key Entities

- **Item list entry**: A row representing either an owned envelope or a
  SharedLink; carries title, type, and ownership for display. Titles require
  decrypting content, since title lives inside the encrypted payload.
- **Editor form**: A per-type form covering the type's standard fields plus a
  repeating custom-field section; produces an `items.Content` value.
- **ShareRecord** (new): the owner's durable record of one outgoing share —
  share ID, source item ID, recipient — enabling edit propagation, delete
  cleanup, and the future share-management UI.
- Existing entities are reused unchanged: `items.Content` and its six concrete
  types, `sharing.Envelope`, `sharing.SharedLink` (see
  [001 data-model.md](../001-sharing-and-items/data-model.md)).

## Assumptions

- Desktop only; a single Fyne window with in-window navigation (no separate
  windows per item). Mobile layout is out of scope.
- Listing requires decrypting every envelope to obtain titles. Acceptable at
  the target scale (personal/team vaults, not thousands of items); revisit
  with cached metadata if it ever isn't.
- The deployed Vault policy grants the paths the spec set 001 requires
  (`users/<eid>/items/*`, `users/<eid>/links/*`, `users/<eid>/identity`);
  verifying the live policy is a deployment task, not part of this feature.

## Out of Scope

- Share and revoke UI (service methods exist; surfacing them is the next
  feature).
- Password generator.
- TOTP code generation/display (the TOTP field stores the secret as text; live
  codes are future work).
- Clipboard auto-clear timers.
- Browser autofill, favicons/site icons, folders/tags, item history, sorting
  preferences.
- Key export/import UI, password change, key rotation (separate features).
