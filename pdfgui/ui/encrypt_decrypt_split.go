package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"pdfutil/shared"
)

// createEncryptDecryptContent creates content for encrypt or decrypt view
func createEncryptDecryptContent(w fyne.Window, fileState *FileState, isEncrypt bool) fyne.CanvasObject {
	var title string
	var buttonLabel string
	var instructions string
	
	if isEncrypt {
		title = "Encrypt PDF"
		buttonLabel = "🔐 Encrypt File"
		instructions = "Encrypt a PDF file with password protection.\nOwner password is required, user password is optional."
	} else {
		title = "Decrypt PDF"
		buttonLabel = "🔓 Decrypt File"
		instructions = "Remove password protection from a PDF file.\nProvide either user or owner password."
	}
	
	// Password fields with auto-population
	userPassEntry, ownerPassEntry := CreatePasswordEntries(fileState)
	
	// Override placeholders for context
	if isEncrypt {
		userPassEntry.PlaceHolder = "User password (optional)"
		ownerPassEntry.PlaceHolder = "Owner password (required)"
	} else {
		userPassEntry.PlaceHolder = "User password (if available)"
		ownerPassEntry.PlaceHolder = "Owner password (if available)"
	}

	// Encryption options (only for encrypt)
	var encryptionBits *widget.Select
	var encryptionMode *widget.Select
	var encryptOptions fyne.CanvasObject
	
	if isEncrypt {
		encryptionBits = widget.NewSelect([]string{"256", "128", "40"}, func(s string) {})
		encryptionBits.SetSelected("256")
		
		encryptionMode = widget.NewSelect([]string{"AES", "RC4"}, func(s string) {})
		encryptionMode.SetSelected("AES")
		
		encryptOptions = widget.NewForm(
			widget.NewFormItem("Encryption Bits", encryptionBits),
			widget.NewFormItem("Encryption Mode", encryptionMode),
		)
	}

	// Status label
	statusLabel := widget.NewLabel("")

	// Action button
	actionButton := widget.NewButton(buttonLabel, func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file first"), w)
			return
		}

		if isEncrypt {
			// Encrypt logic
			if ownerPassEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("owner password is required for encryption"), w)
				return
			}

			if shared.IsEncrypted(selectedFile) {
				dialog.ShowError(fmt.Errorf("file is already encrypted"), w)
				return
			}

			bits := 256
			if encryptionBits.Selected == "128" {
				bits = 128
			} else if encryptionBits.Selected == "40" {
				bits = 40
			}

			// Get and store passwords
			userPw, ownerPw := GetPasswordsFromEntries(userPassEntry, ownerPassEntry, fileState)
			
			err := shared.EncryptFile(selectedFile, userPw, ownerPw, bits, encryptionMode.Selected)
			if err != nil {
				dialog.ShowError(fmt.Errorf("encryption failed: %v", err), w)
				return
			}

			statusLabel.SetText("✓ File encrypted successfully!")
			dialog.ShowInformation("Success", "PDF file encrypted successfully!", w)
		} else {
			// Decrypt logic
			if userPassEntry.Text == "" && ownerPassEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("at least one password is required for decryption"), w)
				return
			}

			if !shared.IsEncrypted(selectedFile) {
				dialog.ShowError(fmt.Errorf("file is not encrypted"), w)
				return
			}

			// Get and store passwords
			userPw, ownerPw := GetPasswordsFromEntries(userPassEntry, ownerPassEntry, fileState)
			
			err := shared.DecryptFile(selectedFile, userPw, ownerPw, 0)
			if err != nil {
				dialog.ShowError(fmt.Errorf("decryption failed: %v", err), w)
				return
			}

			statusLabel.SetText("✓ File decrypted successfully!")
			dialog.ShowInformation("Success", "PDF file decrypted successfully!", w)
		}
	})

	// Layout
	var content *fyne.Container
	if isEncrypt {
		content = container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(instructions),
			widget.NewSeparator(),
			
			widget.NewLabelWithStyle("Passwords", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewForm(
				widget.NewFormItem("User Password", userPassEntry),
				widget.NewFormItem("Owner Password", ownerPassEntry),
			),
			widget.NewSeparator(),
			
			widget.NewLabelWithStyle("Encryption Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			encryptOptions,
			widget.NewSeparator(),
			
			actionButton,
			statusLabel,
		)
	} else {
		content = container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(instructions),
			widget.NewSeparator(),
			
			widget.NewLabelWithStyle("Passwords", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewForm(
				widget.NewFormItem("User Password", userPassEntry),
				widget.NewFormItem("Owner Password", ownerPassEntry),
			),
			widget.NewSeparator(),
			
			actionButton,
			statusLabel,
		)
	}

	return container.NewPadded(container.NewVScroll(content))
}

