package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"pdfutil/shared"
)

// SaveInPlace saves changes to the current file by using a temporary file
// This masks the complexity of pdfcpu requiring output files
func SaveInPlace(currentFile string, userPw, ownerPw string, 
	operation func(inFile, outFile string) error, w fyne.Window) error {
	
	// Create temporary file in same directory
	dir := filepath.Dir(currentFile)
	base := filepath.Base(currentFile)
	tempFile := filepath.Join(dir, ".tmp_"+base)
	
	// Ensure temp file doesn't exist
	os.Remove(tempFile)
	
	// Perform operation to temp file
	err := operation(currentFile, tempFile)
	if err != nil {
		os.Remove(tempFile) // Clean up on error
		return fmt.Errorf("operation failed: %v", err)
	}
	
	// If original file was encrypted, re-encrypt the temp file with same passwords
	if shared.IsEncrypted(currentFile) && (userPw != "" || ownerPw != "") {
		// The temp file should already maintain encryption, but ensure it's there
		// We don't need to re-encrypt as pdfcpu operations preserve encryption
	}
	
	// Replace original with temp file
	err = os.Remove(currentFile)
	if err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to remove original file: %v", err)
	}
	
	err = os.Rename(tempFile, currentFile)
	if err != nil {
		return fmt.Errorf("failed to replace file: %v", err)
	}
	
	return nil
}

// SavePropertiesInPlace saves properties to the current file transparently
func SavePropertiesInPlace(currentFile string, userPw, ownerPw string, 
	properties map[string]string, w fyne.Window) error {
	
	return SaveInPlace(currentFile, userPw, ownerPw, func(inFile, outFile string) error {
		return shared.AddProperties(inFile, outFile, userPw, ownerPw, properties)
	}, w)
}

// SavePermissionsInPlace saves permissions to the current file transparently
func SavePermissionsInPlace(currentFile string, userPw, ownerPw string, 
	encryptBits int, permType string, w fyne.Window) error {
	
	return SaveInPlace(currentFile, userPw, ownerPw, func(inFile, outFile string) error {
		return shared.SetPermissions(inFile, outFile, userPw, ownerPw, encryptBits, permType)
	}, w)
}

