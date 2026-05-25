package core

import "cowbird/internal/vault"

type App struct {
	vault *vault.Vault
}

func NewApp(v *vault.Vault) *App {
	return &App{vault: v}
}
