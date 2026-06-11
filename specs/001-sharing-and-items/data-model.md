# Data Model: Item Sharing, Item Types, and Key Recovery

**Branch**: `001-sharing-and-items` | **Plan**: [plan.md](./plan.md)

All structs are JSON-serialized. Item content is serialized, then encrypted into
the envelope ciphertext. Field tags use `omitempty` where a field is optional.

## Item content types (`internal/items`)

The chosen approach is one `Content` interface with concrete typed structs
(option "B"), each carrying a `CustomFields` slice for arbitrary extra fields.
A `Custom` type covers fully freeform items.

```go
type ItemType string

const (
	TypeLogin    ItemType = "login"
	TypeCard     ItemType = "card"
	TypeNote     ItemType = "note"
	TypeIdentity ItemType = "identity"
	TypePassword ItemType = "password"
	TypeCustom   ItemType = "custom"
)

type FieldType string

const (
	FieldText   FieldType = "text"
	FieldHidden FieldType = "hidden"
	FieldTOTP   FieldType = "totp"
	FieldURL    FieldType = "url"
)

type Field struct {
	Type  FieldType `json:"type"`
	Label string    `json:"label"`
	Value string    `json:"value"`
}

// Content is the decrypted payload. Concrete types implement it.
type Content interface {
	Kind() ItemType
}

type Login struct {
	Title        string   `json:"title"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	URLs         []string `json:"urls,omitempty"`
	TOTP         string   `json:"totp,omitempty"`
	Note         string   `json:"note,omitempty"`
	CustomFields []Field  `json:"custom_fields,omitempty"`
}

type Card struct {
	Title          string  `json:"title"`
	Cardholder     string  `json:"cardholder"`
	Number         string  `json:"number"`
	ExpirationDate string  `json:"expiration_date"`
	CVV            string  `json:"cvv,omitempty"`
	PIN            string  `json:"pin,omitempty"`
	Note           string  `json:"note,omitempty"`
	CustomFields   []Field `json:"custom_fields,omitempty"`
}

type Note struct {
	Title        string  `json:"title"`
	Body         string  `json:"body"`
	CustomFields []Field `json:"custom_fields,omitempty"`
}

type Identity struct {
	Title        string  `json:"title"`
	FirstName    string  `json:"first_name,omitempty"`
	LastName     string  `json:"last_name,omitempty"`
	Email        string  `json:"email,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	Address      string  `json:"address,omitempty"`
	Company      string  `json:"company,omitempty"`
	JobTitle     string  `json:"job_title,omitempty"`
	Note         string  `json:"note,omitempty"`
	CustomFields []Field `json:"custom_fields,omitempty"`
}

type Password struct {
	Title        string  `json:"title"`
	Password     string  `json:"password"`
	Note         string  `json:"note,omitempty"`
	CustomFields []Field `json:"custom_fields,omitempty"`
}

type Custom struct {
	Title        string  `json:"title"`
	CustomFields []Field `json:"custom_fields,omitempty"`
}

func (Login) Kind() ItemType    { return TypeLogin }
func (Card) Kind() ItemType     { return TypeCard }
func (Note) Kind() ItemType     { return TypeNote }
func (Identity) Kind() ItemType { return TypeIdentity }
func (Password) Kind() ItemType { return TypePassword }
func (Custom) Kind() ItemType   { return TypeCustom }
```

**Note**: JSON cannot unmarshal into the `Content` interface directly. A decode
helper reads the type, then unmarshals into the matching concrete struct. This
is the one piece of friction inherent to approach B. (Helper not yet written.)

## Encryption envelope (`internal/sharing` or `internal/crypto`)

```go
type WrappedKey struct {
	RecipientID  string `json:"recipient_id"`  // recipient entity ID / key fingerprint
	EphemeralPub []byte `json:"ephemeral_pub"` // X25519 box ephemeral public key
	Nonce        []byte `json:"nonce"`
	Wrapped      []byte `json:"wrapped"`       // item key encrypted to recipient
}

type Envelope struct {
	ID         string       `json:"id"`
	Type       ItemType     `json:"type"`
	OwnerID    string       `json:"owner_id"`
	Recipients []WrappedKey `json:"recipients,omitempty"` // see note below
	Nonce      []byte       `json:"nonce"`
	Ciphertext []byte       `json:"ciphertext"`           // content encrypted with item key
	Signature  []byte       `json:"signature,omitempty"`  // optional Ed25519, deferred
}
```

**Design note**: For shared items, the recipient's wrapped key is delivered via
the inbox message and stored in the recipient's SharedLink, NOT placed in the
shared envelope's `Recipients`. This keeps "who has access" out of the
shared-readable namespace. `Recipients` on the envelope is therefore primarily
the owner's own access for non-shared items.

## Inbox message (`internal/sharing`)

Consume-and-delete. Written once by sender (create-only), read then deleted by
recipient.

```go
type MessageType string

const (
	MessageShare  MessageType = "share"
	MessageRevoke MessageType = "revoke"
)

type Message struct {
	Type       MessageType `json:"type"`
	ShareID    string      `json:"share_id"`    // opaque UUID
	SenderID   string      `json:"sender_id"`   // informational
	EnvVersion int64       `json:"env_version"` // KV v2 version of shared envelope; ordering tiebreaker
	Timestamp  time.Time   `json:"timestamp"`   // display only, not authoritative

	Share *SharePayload `json:"share,omitempty"` // present only for share messages
}

type SharePayload struct {
	SharePath  string `json:"share_path"`  // location of the shared envelope
	WrappedKey []byte `json:"wrapped_key"` // JSON-encoded WrappedKey struct (ephemeral_pub + nonce + wrapped)
	ItemType   string `json:"item_type"`   // for list display before decrypt
	OwnerID    string `json:"owner_id"`
}

// InboxEntry pairs a Message with the Vault path key required to delete it.
// Message has no ID field; the path component is external to the struct.
type InboxEntry struct {
	ID  string  // Vault KV key (UUID); passed to DeleteInboxMessage
	Msg Message
}
```

## Shared link (`internal/sharing`)

Durable record written into the recipient's own subtree after consuming a share
message. The recipient's standing record of an item shared with them.

```go
type SharedLink struct {
	ShareID    string `json:"share_id"`
	SharePath  string `json:"share_path"`  // where the envelope lives
	WrappedKey []byte `json:"wrapped_key"` // JSON-encoded WrappedKey struct (ephemeral_pub + nonce + wrapped)
	OwnerID    string `json:"owner_id"`
	ItemType   string `json:"item_type"`   // for list display
	EnvVersion int64  `json:"env_version"` // last-seen version acted on
}
```

## Per-user key material (`internal/crypto`)

```go
type Identity struct {
	SigningPub     ed25519.PublicKey  // verify authorship (optional, deferred)
	SigningPriv    ed25519.PrivateKey
	EncryptionPub  [32]byte           // X25519, wrap item keys to this user
	EncryptionPriv [32]byte
	Fingerprint    string             // hex SHA-256 of EncryptionPub
}

// LockedIdentity is the at-rest form stored in Vault at users/<entityID>/identity.
type LockedIdentity struct {
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

// ExportedKey is the passphrase-protected export produced by ExportKey.
type ExportedKey struct {
	Version    int    `json:"version"`
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}
```

Private keys are held in memory only after unlock; at rest they are stored as
`LockedIdentity` encrypted under an Argon2id-derived key (HKDF info
`"cowbird-unlock-v1"`). Export produces a passphrase-protected `ExportedKey`
JSON blob; the current version is `1`.
