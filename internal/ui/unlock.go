package ui

import (
	"context"
	"errors"
	"fmt"

	"cowbird/internal/core"
	"cowbird/internal/sharing"
	"cowbird/internal/vault"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// UnlockDoneFunc is called after the identity is successfully created or
// unlocked, with the fully initialised App.
type UnlockDoneFunc func(app *core.App)

// NewUnlockWindow creates the identity unlock (or first-run setup) window.
//
// It first checks whether an identity already exists in Vault and presents
// either a "set password" form (first run, with confirmation field) or an
// "enter password" form (returning user).
func NewUnlockWindow(a fyne.App, v *vault.Vault, onUnlock UnlockDoneFunc) fyne.Window {
	w := a.NewWindow("Cowbird")
	w.Resize(fyne.NewSize(360, 260))
	w.CenterOnScreen()

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")
	confirmEntry := widget.NewPasswordEntry()
	confirmEntry.SetPlaceHolder("Confirm password")

	statusLabel := widget.NewLabel("")
	submitBtn := widget.NewButton("Please wait…", nil)
	submitBtn.Disable()

	// body is a single-slot container; we swap its content after the first-run
	// check completes.
	body := container.NewMax(container.NewCenter(widget.NewLabel("Connecting to Vault…")))
	w.SetContent(container.NewPadded(body))

	setStatus := func(msg string) {
		fyne.Do(func() { statusLabel.SetText(msg) })
	}

	var isFirstRun bool

	submitBtn.OnTapped = func() {
		password := []byte(passwordEntry.Text)
		if isFirstRun && passwordEntry.Text != confirmEntry.Text {
			statusLabel.SetText("Passwords do not match")
			return
		}

		submitBtn.Disable()
		statusLabel.SetText("Please wait…")

		go func() {
			defer fyne.Do(func() { submitBtn.Enable() })

			id, err := core.InitIdentity(context.Background(), v, password)
			if err != nil {
				setStatus(fmt.Sprintf("Error: %v", err))
				return
			}

			app := core.NewApp(v, id)
			fyne.Do(func() {
				onUnlock(app)
				w.Close()
			})
		}()
	}

	// Check first-run status asynchronously to avoid blocking the main thread.
	go func() {
		_, err := v.GetLockedIdentity(context.Background())
		firstRun := errors.Is(err, sharing.ErrNotFound)

		fyne.Do(func() {
			isFirstRun = firstRun

			var heading string
			var fields fyne.CanvasObject
			if firstRun {
				w.SetTitle("Set Unlock Password")
				heading = "Choose an unlock password for your vault data.\nYou will need this password every time you open Cowbird."
				submitBtn.SetText("Create")
				fields = container.NewVBox(passwordEntry, confirmEntry)
			} else {
				w.SetTitle("Unlock Cowbird")
				heading = "Enter your unlock password."
				submitBtn.SetText("Unlock")
				fields = container.NewVBox(passwordEntry)
			}

			form := container.NewVBox(
				widget.NewLabel(heading),
				widget.NewSeparator(),
				fields,
				submitBtn,
				statusLabel,
			)
			body.Objects = []fyne.CanvasObject{form}
			body.Refresh()
			submitBtn.Enable()
		})
	}()

	return w
}