//go:build (darwin && !ios) || linux || windows

package credentials

import "github.com/99designs/keyring"

type KeyringStore struct {
	ring keyring.Keyring
}

func NewStore(appName string) (CredentialStore, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: appName,
	})
	if err != nil {
		return nil, err
	}
	return &KeyringStore{ring: ring}, nil
}

func (k *KeyringStore) Get(key string) (string, error) {
	item, err := k.ring.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Data), nil
}

func (k *KeyringStore) Set(key, value string) error {
	return k.ring.Set(keyring.Item{
		Key:  key,
		Data: []byte(value),
	})
}

func (k *KeyringStore) Delete(key string) error {
	return k.ring.Remove(key)
}
