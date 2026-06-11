package items

import (
	"encoding/json"
	"fmt"
)

// envelope is the at-rest JSON representation of any Content value.
// The type tag lets Decode reconstruct the correct concrete type.
type envelope struct {
	Type ItemType        `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Encode serializes a Content value to JSON, embedding a type tag so that
// Decode can reconstruct the correct concrete type without any out-of-band hint.
func Encode(c Content) ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("encoding %s content: %w", c.Kind(), err)
	}
	return json.Marshal(envelope{Type: c.Kind(), Data: data})
}

// Decode deserializes a JSON envelope produced by Encode into the appropriate
// concrete Content type.
func Decode(b []byte) (Content, error) {
	var env envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("parsing item envelope: %w", err)
	}
	switch env.Type {
	case TypeLogin:
		return decodeAs[Login](env.Data, "login")
	case TypeCard:
		return decodeAs[Card](env.Data, "card")
	case TypeNote:
		return decodeAs[Note](env.Data, "note")
	case TypeIdentity:
		return decodeAs[Identity](env.Data, "identity")
	case TypePassword:
		return decodeAs[Password](env.Data, "password")
	case TypeCustom:
		return decodeAs[Custom](env.Data, "custom")
	default:
		return nil, fmt.Errorf("unknown item type %q", env.Type)
	}
}

func decodeAs[T Content](data json.RawMessage, label string) (Content, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", label, err)
	}
	return v, nil
}