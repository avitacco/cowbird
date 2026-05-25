package ui

import (
	"cowbird/internal/encryption"
	"cowbird/internal/vault"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewMainWindow(a fyne.App, v *vault.Vault) fyne.Window {
	w := a.NewWindow("Cowbird")
	w.Resize(fyne.NewSize(800, 600))

	masterpassword := widget.NewPasswordEntry()
	masterpassword.SetPlaceHolder("Enter your master password")

	content := container.NewBorder(
		nil,
		container.NewVBox(
			widget.NewLabel("Welcome to Cowbird"),
			masterpassword,
			widget.NewButton(
				"Login",
				func() {
					fmt.Printf("Password is %s\n", masterpassword.Text)
					salt, err := encryption.GenerateSalt()
					if err != nil {
						fmt.Println("Error generating salt:", err)
						return
					}
					key := encryption.DeriveEncKey([]byte(masterpassword.Text), salt)
					fmt.Printf("Encryption key: %x\n", key)

					err = v.Put("users/"+v.EntityID+"/details", map[string]interface{}{
						"salt": fmt.Sprintf("%x", salt),
					})
				},
			),
		),
		nil, nil, nil,
	)

	w.SetContent(content)

	return w
}
