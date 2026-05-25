//go:build android || ios

package credentials

import "errors"

type stubStore struct{}

func NewStore(appName string) (CredentialStore, error) {
	return &stubStore{}, nil
}

func (s *stubStore) Get(key string) (string, error) {
	return "", errors.New("credential storage not yet implemented on mobile")
}

func (s *stubStore) Set(key, value string) error {
	return errors.New("credential storage not yet implemented on mobile")
}

func (s *stubStore) Delete(key string) error {
	return errors.New("credential storage not yet implemented on mobile")
}
