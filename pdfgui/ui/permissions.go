package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"pdfutil/shared"
)

// PermissionsState holds auto-load function for permissions
type PermissionsState struct {
	autoLoadViewFunc func() // Function to auto-load and display permissions
}

// CreatePermissionsTab creates the permissions management UI tab
func CreatePermissionsTab(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	// Use horizontal split for better layout
	permsState := &PermissionsState{}
	viewSection := createViewPermissionsSection(w, fileState, permsState)
	setSection := createSetPermissionsSection(w, fileState)
	
	return container.NewHSplit(viewSection, setSection)
}

func createViewPermissionsSection(w fyne.Window, fileState *FileState, permsState *PermissionsState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	// Password fields with auto-population
	userPassEntry, ownerPassEntry := CreatePasswordEntries(fileState)

	// Make result text much larger and more visible
	resultText := widget.NewLabel("Select a file and permissions will load automatically")
	resultText.Wrapping = fyne.TextWrapWord
	resultScroll := container.NewScroll(resultText)
	resultScroll.SetMinSize(fyne.NewSize(400, 300))

	// Clear permissions display when file changes
	fileState.OnFileChange(func(file string) {
		resultText.SetText("File changed. Will auto-load permissions if passwords available.")
	})

	// Function to load and display permissions
	loadAndDisplayPermissions := func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}

		// Get passwords from entries
		userPw, ownerPw := GetPasswordsFromEntries(userPassEntry, ownerPassEntry, fileState)

		perms, err := shared.GetPermissions(selectedFile, userPw, ownerPw)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to get permissions: %v", err), w)
			resultText.SetText("Error reading permissions. Check passwords if file is encrypted.")
			return
		}

		if perms == nil {
			resultText.SetText("Document is NOT encrypted\n\nFull access - all permissions granted")
		} else {
			details := shared.DecodePermissions(*perms)
			result := fmt.Sprintf("Raw Value: 0x%04X (%d)\n\n", details.RawValue, details.RawValue)
			result += "═══════════════════════════════════\n\n"
			result += fmt.Sprintf("Print Document:       %s\n", formatPerm(details.Print))
			result += fmt.Sprintf("Print High Quality:   %s\n", formatPerm(details.PrintHighRes))
			result += fmt.Sprintf("Modify Document:      %s\n", formatPerm(details.Modify))
			result += fmt.Sprintf("Copy/Extract:         %s\n", formatPerm(details.Extract))
			result += fmt.Sprintf("Extract (Rev 3+):     %s\n", formatPerm(details.ExtractRev3))
			result += fmt.Sprintf("Annotate/Comment:     %s\n", formatPerm(details.Annotate))
			result += fmt.Sprintf("Fill Forms:           %s\n", formatPerm(details.FillForms))
			result += fmt.Sprintf("Assemble Document:    %s\n", formatPerm(details.Assemble))
			resultText.SetText(result)
		}
	}

	viewButton := widget.NewButton("View Permissions (Refresh)", loadAndDisplayPermissions)

	// Store the load function so dashboard can trigger it
	permsState.autoLoadViewFunc = loadAndDisplayPermissions

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("View Permissions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			fileSelector,
			widget.NewForm(
				widget.NewFormItem("User Password", userPassEntry),
				widget.NewFormItem("Owner Password", ownerPassEntry),
			),
			viewButton,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		resultScroll,
	)
}

func createSetPermissionsSection(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	// Password fields with auto-population (need both for encrypted files)
	userPassEntry, ownerPassEntry := CreatePasswordEntries(fileState)

	// Permission presets with descriptions (matching actual behavior)
	permTypeSelect := widget.NewSelect(
		[]string{
			"all - Full permissions (everything allowed)",
			"none - No permissions (complete lockdown)",
			"print - Allow printing only (print + print high quality)",
			"readonly - Allow copy/extract text only (no print, no modify)",
			"forms - Fill forms and annotate (no print, no modify)",
			"annotate - Add annotations and fill forms (no print, no modify)",
			"modify - Full editing (modify, assemble, forms, but no print)",
		},
		func(s string) {},
	)
	permTypeSelect.SetSelected("print - Allow printing only (print + print high quality)")

	encryptBitsSelect := widget.NewSelect([]string{"256", "128", "40"}, func(s string) {})
	encryptBitsSelect.SetSelected("256")

	inPlaceCheck := widget.NewCheck("Modify file in-place", func(checked bool) {})
	inPlaceCheck.SetChecked(true)

	// Helper text explaining permissions
	helpText := widget.NewLabel("Set permissions to control what users can do with this PDF:")
	helpText.Wrapping = fyne.TextWrapWord
	
	setButton := widget.NewButton("Set Permissions", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}
		
		// Get passwords - need both user and owner
		userPw, ownerPw := GetPasswordsFromEntries(userPassEntry, ownerPassEntry, fileState)
		
		if ownerPw == "" {
			dialog.ShowError(fmt.Errorf("owner password is required to set permissions"), w)
			return
		}

		if !shared.IsEncrypted(selectedFile) {
			dialog.ShowError(fmt.Errorf("file must be encrypted first. Use Encrypt/Decrypt tab to encrypt the file."), w)
			return
		}

		bits := 256
		if encryptBitsSelect.Selected == "128" {
			bits = 128
		} else if encryptBitsSelect.Selected == "40" {
			bits = 40
		}

		// Extract permission type from selection (before the dash)
		permType := permTypeSelect.Selected
		if idx := findString(permType, " - "); idx > 0 {
			permType = permType[:idx]
		}

		if inPlaceCheck.Checked {
			// In-place modification (transparent temp file handling)
			err := SavePermissionsInPlace(selectedFile, userPw, ownerPw, bits, permType, w)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to set permissions: %v", err), w)
				return
			}
			dialog.ShowInformation("Success", "Permissions updated successfully!", w)
		} else {
			// Save to new file
			ShowLargeFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				if writer == nil {
					return
				}
				defer writer.Close()

				err = shared.SetPermissions(selectedFile, writer.URI().Path(), userPw, ownerPw, bits, permType)
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to set permissions: %v", err), w)
					return
				}
				dialog.ShowInformation("Success", "Permissions set successfully!", w)
			}, w, fileState)
		}
	})

	scrollContent := container.NewVBox(
		widget.NewLabelWithStyle("Set Permissions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("⚠️  Note: File must be encrypted first (use Encrypt tab)"),
		widget.NewSeparator(),
		fileSelector,
		widget.NewSeparator(),
		helpText,
		widget.NewForm(
			widget.NewFormItem("User Password", userPassEntry),
			widget.NewFormItem("Owner Password (required)", ownerPassEntry),
		),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Permission Settings:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Permission Level", permTypeSelect),
			widget.NewFormItem("Encryption Strength", encryptBitsSelect),
		),
		widget.NewLabel("💡 Common choices: 'print' (printing only), 'all' (full access), 'none' (locked down)"),
		widget.NewLabel("⚠️  Note: Most presets don't include print permission except 'print' and 'all'"),
		widget.NewSeparator(),
		inPlaceCheck,
		setButton,
	)
	
	return container.NewPadded(container.NewScroll(scrollContent))
}

func formatPerm(allowed bool) string {
	if allowed {
		return "✓ ALLOWED"
	}
	return "✗ DENIED"
}

// Helper function to find substring
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
