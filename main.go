package main

import (
	"cowbird/internal/auth"
	"cowbird/internal/config"
	"cowbird/internal/core"
	"cowbird/internal/credentials"
	"cowbird/internal/ui"
	"cowbird/internal/vault"
	"log"

	"fyne.io/fyne/v2/app"
	"github.com/99designs/keyring"
)

func main() {
	a := app.NewWithID("co.avitac.cowbird")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	needsSetup := cfg.Vault.Address == "" || cfg.Vault.AuthMethod == ""

	if !needsSetup {
		// Config exists — check that credentials are actually present.
		store, err := credentials.NewStore("cowbird")
		if err != nil {
			log.Fatalf("error opening credential store: %v", err)
		}
		method := auth.ByName(cfg.Vault.AuthMethod)
		if method == nil {
			log.Fatalf("unknown auth method %q in config — delete config and re-run setup", cfg.Vault.AuthMethod)
		}
		if len(method.Fields()) > 0 {
			if _, err := store.Get(method.Fields()[0].Key); err == keyring.ErrKeyNotFound {
				needsSetup = true
			}
		}
	}

	if needsSetup {
		w := ui.NewSetupWindow(a, func(cfg config.Config, method auth.Method, store credentials.CredentialStore) error {
			v, err := vault.NewVault(cfg.Vault, store, method)
			if err != nil {
				return err
			}
			unlockW := ui.NewUnlockWindow(a, v, func(coreApp *core.App) {
				mainW := ui.NewMainWindow(a, coreApp)
				mainW.Show()
			})
			unlockW.Show()
			return nil
		})
		w.ShowAndRun()
		return
	}

	store, err := credentials.NewStore("cowbird")
	if err != nil {
		log.Fatalf("error opening credential store: %v", err)
	}

	method := auth.ByName(cfg.Vault.AuthMethod)
	v, err := vault.NewVault(cfg.Vault, store, method)
	if err != nil {
		log.Fatalf("error connecting to vault: %v", err)
	}
	defer v.Close()

	w := ui.NewUnlockWindow(a, v, func(coreApp *core.App) {
		mainW := ui.NewMainWindow(a, coreApp)
		mainW.Show()
	})
	w.ShowAndRun()
}
