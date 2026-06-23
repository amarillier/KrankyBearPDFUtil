package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"pdfutil/shared"
)

// CreateEncryptDecryptTab creates the encryption/decryption UI tab
func CreateEncryptDecryptTab(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	// File selection using shared state
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	// Password fields
	userPassEntry := widget.NewPasswordEntry()
	userPassEntry.SetPlaceHolder("User password (optional)")
	ownerPassEntry := widget.NewPasswordEntry()
	ownerPassEntry.SetPlaceHolder("Owner password (required)")

	// Encryption options
	encryptionBits := widget.NewSelect([]string{"40", "128", "256"}, func(s string) {})
	encryptionBits.SetSelected("256")

	encryptionMode := widget.NewSelect([]string{"AES", "RC4"}, func(s string) {})
	encryptionMode.SetSelected("AES")

	// Status label
	statusLabel := widget.NewLabel("")

	// Encrypt button
	encryptButton := widget.NewButton("Encrypt PDF", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file first"), w)
			return
		}
		if ownerPassEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("owner password is required for encryption"), w)
			return
		}

		// Check if already encrypted
		if shared.IsEncrypted(selectedFile) {
			dialog.ShowError(fmt.Errorf("file is already encrypted"), w)
			return
		}

		// Get encryption bits
		bits := 256
		if encryptionBits.Selected == "128" {
			bits = 128
		} else if encryptionBits.Selected == "40" {
			bits = 40
		}

		// Encrypt the file
		err := shared.EncryptFile(selectedFile, userPassEntry.Text, ownerPassEntry.Text, bits, encryptionMode.Selected)
		if err != nil {
			dialog.ShowError(fmt.Errorf("encryption failed: %v", err), w)
			return
		}

		statusLabel.SetText("✓ File encrypted successfully!")
		dialog.ShowInformation("Success", "PDF file encrypted successfully!", w)
	})

	// Decrypt button
	decryptButton := widget.NewButton("Decrypt PDF", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file first"), w)
			return
		}
		if userPassEntry.Text == "" && ownerPassEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("at least one password is required for decryption"), w)
			return
		}

		// Check if encrypted
		if !shared.IsEncrypted(selectedFile) {
			dialog.ShowError(fmt.Errorf("file is not encrypted"), w)
			return
		}

		// Decrypt the file
		err := shared.DecryptFile(selectedFile, userPassEntry.Text, ownerPassEntry.Text, 0)
		if err != nil {
			dialog.ShowError(fmt.Errorf("decryption failed: %v", err), w)
			return
		}

		statusLabel.SetText("✓ File decrypted successfully!")
		dialog.ShowInformation("Success", "PDF file decrypted successfully!", w)
	})

	// Layout
	form := container.NewVBox(
		fileSelector,
		widget.NewSeparator(),
		
		widget.NewLabelWithStyle("Passwords", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("User Password", userPassEntry),
			widget.NewFormItem("Owner Password", ownerPassEntry),
		),
		widget.NewSeparator(),
		
		widget.NewLabelWithStyle("Encryption Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Encryption Bits", encryptionBits),
			widget.NewFormItem("Encryption Mode", encryptionMode),
		),
		widget.NewSeparator(),
		
		container.NewHBox(
			encryptButton,
			decryptButton,
		),
		statusLabel,
	)

	return container.NewPadded(form)
}

