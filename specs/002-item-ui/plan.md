# Implementation Plan: Item List and Editor UI

**Branch**: `002-item-ui` | **Spec**: [spec.md](./spec.md)

## Technical Context

**Language**: Go 1.26
**GUI**: Fyne (`fyne.io/fyne/v2`), app ID `co.avitac.cowbird`
**Builds on**: spec set [001-sharing-and-items](../001-sharing-and-items/plan.md) —
crypto, items codec, sharing service, Vault store, and unlock flow are all in
place. The main window (`internal/ui/app.go`) is a fingerprint placeholder.

## Architecture Overview

### A discovered gap: owners have no record of outgoing shares

`Service.Share` writes a *copy* of the item's envelope (new `shareID`, same
ciphertext and item key) into the shared namespace and sends the inbox
message — and records nothing on the owner's side. Sharing the same item with
two recipients produces two independent shared envelopes. Consequently:

- An owner edit cannot propagate to shared copies (001's FR-005 "edits visible
  without re-sharing" is currently unsatisfiable after the fact).
- Delete cannot clean up shared envelopes for the deleted item.
- A future share/revoke UI has nothing to list ("who did I share this with?").

**Resolution (this feature)**: introduce an owner-side **ShareRecord**, written
by `Share` and stored in the owner's own subtree:

```
users/<entityID>/shares/<shareID>   # owner's durable record of an outgoing share
```

```go
type ShareRecord struct {
	ShareID     string `json:"share_id"`
	ItemID      string `json:"item_id"`      // owner's item the share was made from
	RecipientID string `json:"recipient_id"`
	ItemType    string `json:"item_type"`
}
```

Stored plaintext-JSON like `SharedLink` (same `{"v": ...}` wrapper). It reveals
sharing relationships only to the owner and the operator — the operator can
already observe inbox writes and shared-namespace paths, so this leaks nothing
new. The 001 policy's self-subtree rule (`users/{{identity.entity.id}}/*`)
already covers the new path.

`Revoke` deletes the matching ShareRecord after deleting the envelope.

### Service-layer additions (`internal/sharing`)

The UI talks only to `sharing.Service`; the Store interface stays an internal
detail. New methods:

| Method | Behavior |
|---|---|
| `ListItems(ctx)` | Pass-through to `store.ListItems` — owner's envelopes. |
| `ListSharedLinks(ctx)` | Pass-through to `store.ListSharedLinks`. |
| `UpdateItem(ctx, itemID, content)` | Unwrap the owner's existing item key, re-encrypt content with it (fresh random nonce — never reuse a nonce under the same key), write the owned envelope back, then rewrite every shared envelope listed in the item's ShareRecords. Recipients' wrapped keys stay valid because the item key is unchanged. |
| `DeleteItem(ctx, itemID)` | Delete the owned envelope; for each ShareRecord on the item, delete the shared envelope, send a revoke message to the recipient, and delete the record. |
| `ListShareRecords(ctx, itemID)` | Owner's outgoing shares for an item (needed by UpdateItem/DeleteItem now, share UI later). |

Store interface additions (implemented in `internal/vault/store.go` and the
test in-memory store): `PutShareRecord`, `ListShareRecords`,
`DeleteShareRecord`.

`Share` gains one step: write the ShareRecord after the envelope write
succeeds. Existing shares made before this feature have no records; that is
acceptable pre-release (no production data).

### Stale-link self-healing

Per 001's design, a missed revoke degrades to a dead link. The list-loading
path makes this concrete: when resolving a `SharedLink`, an `ErrNotFound` on
the shared envelope means the share is gone — delete the link and omit the row
(spec scenario 5.3). Any other error keeps the link and surfaces an unreadable
row.

### UI structure (`internal/ui`)

Single window, master-detail layout. The empty `internal/ui/screens` package
stub is deleted; the `ui` package stays flat:

```
internal/ui/
├── app.go       # NewMainWindow: HSplit (list | detail), toolbar (new / refresh), status bar
├── model.go     # row view-model + async load: process inbox, list+decrypt both sources
├── list.go      # widget.List bound to view-models, search entry, type filter, empty state
├── detail.go    # read-only field grid: mask/reveal toggles, per-field copy buttons
├── editor.go    # create/edit form for all six types + custom-fields repeater
├── fields.go    # field descriptors: per-type mapping content struct ↔ form rows
├── setup.go     # existing
└── unlock.go    # existing
```

**Row view-model** (`model.go`):

```go
type itemRow struct {
	ID      string         // itemID (owned) or shareID (shared)
	Title   string
	Type    items.ItemType
	Shared  bool           // true → read-only, show owner badge
	OwnerID string         // set for shared rows
	Content items.Content  // decrypted; nil if Err != nil
	Err     error          // decrypt/decode failure → "unreadable" row (FR-010)
}
```

Titles live inside the encrypted payload, so loading the list decrypts every
envelope (owned) and every link's shared envelope. Acceptable at target scale
(spec assumption); rows with decryption errors render as unreadable entries
without sinking the rest of the list.

**Per-type editors without six hand-written forms** (`fields.go`): a small
descriptor table maps each `items.Content` type to its field list (label, kind,
getter/setter). `editor.go` renders descriptors as form rows generically and
appends the custom-fields repeater (add/remove row, kind selector). `detail.go`
renders the same descriptors read-only. One place defines what a "Card" looks
like; viewer and editor cannot drift apart.

Sensitive-by-default fields: `Login.Password`, `Card.Number/CVV/PIN`,
`Password.Password`, and any custom field with kind `hidden` — masked in the
detail view behind a reveal toggle; copy works without reveal (Fyne
`w.Clipboard().SetContent`).

### Concurrency rules (from project gotchas)

- All Vault I/O (load, save, delete, inbox processing) runs in goroutines —
  never on the Fyne main thread.
- Every widget mutation from a goroutine goes through `fyne.Do()`.
- Widget values are captured on the main thread *before* launching goroutines.
- The load path is: spinner on → goroutine {ProcessInbox → ListItems +
  ListSharedLinks → decrypt all} → `fyne.Do` {rebuild rows, spinner off}.
  Errors land in a status bar with a Retry action (FR-011), never a silent
  empty list.

### Validation

Editors require a non-empty Title for every type (it is the list's display
key). Custom fields require a non-empty label. Everything else is optional —
a password manager should not refuse to store a partial record. Validation
failures keep the editor open with an inline message; nothing is written
(spec scenario 2.3).

## Build Order (risk-first)

1. **Service layer**: `ShareRecord` + store methods (vault + in-memory test
   store), `ListItems` / `ListSharedLinks` / `ListShareRecords` /
   `UpdateItem` / `DeleteItem`, `Share`/`Revoke` record bookkeeping. Extend
   `service_test.go`: update propagates to shared envelopes, delete revokes,
   nonce freshness, stale-link cleanup.
2. **List + load path**: view-model, async load with inbox processing, list
   pane, empty state, unreadable rows, refresh, status-bar errors.
3. **Detail view**: descriptor-driven read-only rendering, mask/reveal, copy.
4. **Editors**: descriptor-driven forms, custom-fields repeater, create and
   edit flows wired to `CreateItem`/`UpdateItem`.
5. **Delete**: confirmation dialog → `DeleteItem` → list refresh.
6. **Search/filter** (P3): title substring + type filter over loaded rows.

Steps 1 is fully testable without a UI; steps 2–6 are exercised against the
live Vault.

## Open Items

- Owner-side ShareRecord is also the foundation the share/revoke UI (feature
  003) will list from; confirm its shape covers that before freezing.
- `internal/ui/screens` stub package: delete.
- TOTP custom-field kind is stored and displayed as text; live code generation
  remains out of scope (spec).
- Re-confirm live Vault policy covers `users/<eid>/identity`,
  `users/<eid>/links/*`, and now `users/<eid>/shares/*` (one rule:
  `users/{{identity.entity.id}}/*` covers all three).
