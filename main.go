package main

import (
	"cowbird/internal/config"
	"cowbird/internal/credentials"
	"cowbird/internal/ui"
	"cowbird/internal/vault"
	"log"

	"fyne.io/fyne/v2/app"
)

const (
	kvMount = "cowbird"
)

//func deriveEncKey(password, salt []byte) []byte {
//	master := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
//	key := make([]byte, 32)
//	r := hkdf.New(sha256.New, master, salt, []byte("cowbird-enc-v1"))
//	if _, err := io.ReadFull(r, key); err != nil {
//		panic(err)
//	}
//	return key
//}

//func encrypt(key, plaintext []byte) (string, error) {
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return "", err
//	}
//	gcm, err := cipher.NewGCM(block)
//	if err != nil {
//		return "", err
//	}
//	nonce := make([]byte, gcm.NonceSize())
//	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
//		return "", err
//	}
//	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, plaintext, nil)), nil
//}

func main() {
	creds, err := credentials.NewStore("cowbird")
	if err != nil {
		log.Fatalf("error creating credentials store: %v", err)
		return
	}

	if _, err := creds.Get("username"); err != nil {
		log.Fatalf("error retrieving username: %v", err)
		return
	}

	config, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
		return
	}

	v := vault.NewVault(
		config.Vault,
		creds,
	)
	a := app.NewWithID("co.avitac.cowbird")
	w := ui.NewMainWindow(a, v)
	w.ShowAndRun()
	/*
		ctx := context.Background()

		client, err := vault.New(
			vault.WithAddress("https://vaultserver:8200"),
			vault.WithRequestTimeout(10*time.Second),
		)
		if err != nil {
			panic(err)
		}

		resp, err := client.Auth.UserpassLogin(
			ctx,
			"somevaliduser",
			schema.UserpassLoginRequest{
				Password: "definitelynotafakepassword",
			},
		)
		if err != nil {
			log.Fatalf("error logging in: %v", err)
		}

		if err := client.SetToken(resp.Auth.ClientToken); err != nil {
			log.Fatalf("error setting token: %v", err)
		}

		err = initializeKeys(client, resp.Auth.EntityID, []byte("definitelynotafakepassword"))
		if err != nil {
			log.Fatalf("error initializing keys: %v", err)
		}

		log.Println("Keys initialized successfully")
	*/
}

//func initializeKeys(client *vault.Client, entityID string, password []byte) error {
//	// Generate Ed25519 keypair -- private key never leaves the client
//	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
//	if err != nil {
//		return err
//	}
//
//	// In production the salt is retrieved from Vault here.
//	// Using a fixed salt for this example only -- never do this in production.
//	salt := make([]byte, 16)
//	encKey := deriveEncKey(password, salt)
//
//	// Encrypt the private key before it touches the network
//	encPrivateKey, err := encrypt(encKey, privateKey)
//	if err != nil {
//		return err
//	}
//
//	// Zero out the plaintext private key now that we have the ciphertext
//	copy(privateKey, make([]byte, len(privateKey)))
//
//	_, err = client.Secrets.KvV2Write(
//		context.Background(),
//		"users/"+entityID+"/keys",
//		schema.KvV2WriteRequest{
//			Data: map[string]interface{}{
//				"private_enc": encPrivateKey,
//				"public":      base64.StdEncoding.EncodeToString(publicKey),
//			},
//		},
//		vault.WithMountPath(kvMount),
//	)
//	return err
//}
