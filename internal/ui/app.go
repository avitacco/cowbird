package ui

import (
	"cowbird/internal/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewMainWindow(a fyne.App, app *core.App) fyne.Window {
	w := a.NewWindow("Cowbird")
	w.Resize(fyne.NewSize(800, 600))

	content := container.NewBorder(
		widget.NewToolbar(
			widget.NewToolbarSpacer(),
		),
		nil, nil, nil,
		container.NewCenter(
			widget.NewLabel("Key: "+app.Identity.Fingerprint[:16]+"…"),
		),
	)

	w.SetContent(content)
	return w
}