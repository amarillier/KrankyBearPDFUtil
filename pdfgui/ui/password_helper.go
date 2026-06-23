package ui

import (
	"fyne.io/fyne/v2/widget"
)

// CreatePasswordEntries creates password entries that auto-populate from FileState
func CreatePasswordEntries(fileState *FileState) (*widget.Entry, *widget.Entry) {
	userPassEntry := widget.NewPasswordEntry()
	userPassEntry.SetPlaceHolder("User password (if needed)")
	ownerPassEntry := widget.NewPasswordEntry()
	ownerPassEntry.SetPlaceHolder("Owner password (if needed)")

	// Pre-populate with stored passwords
	userPw, ownerPw := fileState.GetPasswords()
	if userPw != "" {
		userPassEntry.SetText(userPw)
	}
	if ownerPw != "" {
		ownerPassEntry.SetText(ownerPw)
	}

	// Listen for password changes and update FileState
	userPassEntry.OnChanged = func(s string) {
		_, ownerPw := fileState.GetPasswords()
		fileState.SetPasswords(s, ownerPw)
	}
	ownerPassEntry.OnChanged = func(s string) {
		userPw, _ := fileState.GetPasswords()
		fileState.SetPasswords(userPw, s)
	}

	// Listen for FileState password changes (when cleared on file change)
	fileState.OnPasswordChange(func(userPw, ownerPw string) {
		userPassEntry.SetText(userPw)
		ownerPassEntry.SetText(ownerPw)
	})

	return userPassEntry, ownerPassEntry
}

// GetPasswordsFromEntries gets passwords from entries and updates FileState
func GetPasswordsFromEntries(userEntry, ownerEntry *widget.Entry, fileState *FileState) (string, string) {
	userPw := userEntry.Text
	ownerPw := ownerEntry.Text
	
	// Store for future use
	fileState.SetPasswords(userPw, ownerPw)
	
	return userPw, ownerPw
}

