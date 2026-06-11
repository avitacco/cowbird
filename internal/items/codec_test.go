package items

import (
	"encoding/json"
	"testing"
)

func TestKindMethods(t *testing.T) {
	cases := []struct {
		c    Content
		want ItemType
	}{
		{Login{}, TypeLogin},
		{Card{}, TypeCard},
		{Note{}, TypeNote},
		{Identity{}, TypeIdentity},
		{Password{}, TypePassword},
		{Custom{}, TypeCustom},
	}
	for _, tc := range cases {
		if tc.c.Kind() != tc.want {
			t.Errorf("%T.Kind() = %q, want %q", tc.c, tc.c.Kind(), tc.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	cases := []Content{
		Login{
			Title:    "My Bank",
			Username: "alice",
			Password: "s3cr3t",
			URLs:     []string{"https://bank.example.com", "https://m.bank.example.com"},
			TOTP:     "JBSWY3DPEHPK3PXP",
			Note:     "personal account",
			CustomFields: []Field{
				{Type: FieldText, Label: "Account Number", Value: "123456"},
			},
		},
		Card{
			Title:          "Visa",
			Cardholder:     "Alice Smith",
			Number:         "4111111111111111",
			ExpirationDate: "12/28",
			CVV:            "123",
			PIN:            "9999",
			Note:           "everyday card",
		},
		Note{
			Title: "Secret Note",
			Body:  "Remember to water the plants.",
			CustomFields: []Field{
				{Type: FieldHidden, Label: "Extra", Value: "hidden"},
			},
		},
		Identity{
			Title:     "Alice Smith",
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     "alice@example.com",
			Phone:     "+1 555 1234",
			Address:   "1 Main St",
			Company:   "ACME",
			JobTitle:  "Engineer",
			Note:      "primary identity",
		},
		Password{
			Title:    "WiFi Password",
			Password: "CorrectHorseBatteryStaple",
			Note:     "home network",
		},
		Custom{
			Title: "Custom Item",
			CustomFields: []Field{
				{Type: FieldHidden, Label: "Secret", Value: "hidden value"},
				{Type: FieldURL, Label: "Link", Value: "https://example.com"},
				{Type: FieldTOTP, Label: "OTP", Value: "JBSWY3DPEHPK3PXP"},
			},
		},
	}

	for _, original := range cases {
		t.Run(string(original.Kind()), func(t *testing.T) {
			encoded, err := Encode(original)
			if err != nil {
				t.Fatal(err)
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Fatal(err)
			}

			if decoded.Kind() != original.Kind() {
				t.Fatalf("Kind: got %q, want %q", decoded.Kind(), original.Kind())
			}

			// Re-encode both and compare; canonical for these simple structs.
			origJSON, _ := json.Marshal(original)
			gotJSON, _ := json.Marshal(decoded)
			if string(origJSON) != string(gotJSON) {
				t.Fatalf("content mismatch\ngot:  %s\nwant: %s", gotJSON, origJSON)
			}
		})
	}
}

func TestDecodePartialFields(t *testing.T) {
	// Only required fields present; optional fields should decode to zero values.
	b, err := Encode(Login{Title: "minimal", Username: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	login, ok := decoded.(Login)
	if !ok {
		t.Fatalf("expected Login, got %T", decoded)
	}
	if login.Title != "minimal" {
		t.Fatalf("title: got %q, want %q", login.Title, "minimal")
	}
	if login.URLs != nil {
		t.Fatalf("URLs should be nil, got %v", login.URLs)
	}
}

func TestDecodeUnknownType(t *testing.T) {
	_, err := Decode([]byte(`{"type":"unknown","data":{}}`))
	if err == nil {
		t.Fatal("expected error for unknown item type")
	}
}

func TestDecodeMalformedJSON(t *testing.T) {
	_, err := Decode([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestDecodeMalformedData(t *testing.T) {
	_, err := Decode([]byte(`{"type":"login","data":"not an object"}`))
	if err == nil {
		t.Fatal("expected error for malformed data field")
	}
}