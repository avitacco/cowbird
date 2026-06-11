package items

// ItemType identifies the concrete type of a Content value.
type ItemType string

const (
	TypeLogin    ItemType = "login"
	TypeCard     ItemType = "card"
	TypeNote     ItemType = "note"
	TypeIdentity ItemType = "identity"
	TypePassword ItemType = "password"
	TypeCustom   ItemType = "custom"
)

// FieldType classifies the presentation and handling of a custom field value.
type FieldType string

const (
	FieldText   FieldType = "text"
	FieldHidden FieldType = "hidden"
	FieldTOTP   FieldType = "totp"
	FieldURL    FieldType = "url"
)

// Field is a user-defined key/value pair that can be attached to any item type.
type Field struct {
	Type  FieldType `json:"type"`
	Label string    `json:"label"`
	Value string    `json:"value"`
}

// Content is the decrypted payload of an item. Concrete types implement it.
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