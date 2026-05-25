package vault

import (
	"context"
	"cowbird/internal/config"
	"cowbird/internal/credentials"
	"log"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

type Vault struct {
	Vault      *vault.Client
	SecretPath string
	EntityID   string
}

func NewVault(
	cfg config.Vault,
	creds credentials.CredentialStore,
) *Vault {
	ctx := context.Background()

	client, err := vault.New(
		vault.WithAddress(cfg.Address),
		vault.WithRequestTimeout(cfg.RequestTimeout),
	)
	if err != nil {
		panic(err)
	}

	username, err := creds.Get("username")
	if err != nil {
	}

	password, err := creds.Get("password")
	if err != nil {
	}

	resp, err := client.Auth.UserpassLogin(
		ctx,
		username,
		schema.UserpassLoginRequest{
			Password: password,
		},
	)
	if err != nil {
		log.Fatalf("error logging in: %v", err)
	}

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		log.Fatalf("error setting token: %v", err)
	}

	return &Vault{
		Vault:      client,
		SecretPath: "cowbird",
		EntityID:   resp.Auth.EntityID,
	}
}

func (v *Vault) Put(path string, data map[string]interface{}) error {
	ctx := context.Background()
	_, err := v.Vault.Secrets.KvV2Write(
		ctx,
		"users/"+v.EntityID+"/path",
		schema.KvV2WriteRequest{
			Data: data,
		},
		vault.WithMountPath(v.SecretPath),
	)
	return err
}

func Get(path string) (map[string]interface{}, error) {
	return nil, nil
}

func Delete(path string) error {
	return nil
}

func PutEncrypted(path string, data map[string]interface{}) error {
	return nil
}

func GetEncrypted(path string) (map[string]interface{}, error) {
	return nil, nil
}
