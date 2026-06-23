// Package main provides update checking dialog
// Note: About and Help dialogs have been moved to separate files:
//   - about.go: About dialog (reusable)
//   - help.go: Help dialog (reusable)
//   - dialogs.go: Update checker dialog (this file)
package main

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var updateWindow fyne.Window

func showUpdateDialog(a fyne.App, message string, updateAvailable bool) {
	if updateWindow != nil && updateWindow.Content().Visible() {
		updateWindow.Show()
		updateWindow.RequestFocus()
		return
	}

	updateWindow = a.NewWindow(appName + " - Update Check")
	updateWindow.SetIcon(resourceKrankyBearBeanieMultiColorPng)

	icon := canvas.NewImageFromResource(resourceKrankyBearBeanieMultiColorPng)
	icon.FillMode = canvas.ImageFillOriginal
	icon.SetMinSize(fyne.NewSize(64, 64))

	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord
	messageLabel.Alignment = fyne.TextAlignCenter

	var content *fyne.Container
	if updateAvailable {
		releaseURL, _ := url.Parse("https://github.com/amarillier/KrankyBearPDF/releases/latest")
		releaseLink := widget.NewHyperlink("Download Latest Release", releaseURL)
		releaseLink.Alignment = fyne.TextAlignCenter

		notesURL, _ := url.Parse("https://github.com/amarillier/KrankyBearPDF/blob/main/ReleaseNotes.txt")
		notesLink := widget.NewHyperlink("View Release Notes", notesURL)
		notesLink.Alignment = fyne.TextAlignCenter

		content = container.NewVBox(
			container.NewCenter(icon),
			widget.NewSeparator(),
			messageLabel,
			widget.NewSeparator(),
			container.NewCenter(releaseLink),
			container.NewCenter(notesLink),
		)
	} else {
		content = container.NewVBox(
			container.NewCenter(icon),
			widget.NewSeparator(),
			messageLabel,
		)
	}

	updateWindow.SetContent(container.NewPadded(content))
	updateWindow.Resize(fyne.NewSize(450, 300))

	updateWindow.SetCloseIntercept(func() {
		updateWindow.Hide()
	})

	updateWindow.Show()
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
