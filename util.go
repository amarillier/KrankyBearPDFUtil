package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	updatechecker "github.com/amarillier/go-update-checker"
)

// pathExists checks if a file or directory exists at the given path and returns true or false
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // File or directory exists
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil // File or directory does not exist
	}
	return false, err // An error occurred other than "not exist"
}

// makeTempDir creates a temporary directory for output files using the current pid
// and a simple, random number based on the pid
func makeTempDir() string {
	// Create a temporary directory for output files
	pid := strconv.Itoa(os.Getpid())
	seed := rand.NewSource(time.Now().UnixNano())
	random := strconv.Itoa(rand.New(seed).Intn(1000))
	// Use PID and random number to create a unique directory name
	tempDir := pid + "-temp-" + random
	err = os.MkdirAll(tempDir, 0700) // permissions)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return ""
	}
	return tempDir
}

// readFilesInDirectory reads the content of all files in a given directory
// into a string array.
func readFilesInDir(dirPath string) ([]string, error) {
	var contents []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}
		filePath := filepath.Join(dirPath, entry.Name())
		contents = append(contents, filePath)
	}
	return contents, nil
}

func updateChecker(repoOwner string, repo string, repoName string, repodl string) (string, bool) {
	// uc := updatechecker.New("amarillier", "pdfutil", "pdfutil", "", 1, false)
	uc := updatechecker.New(repoOwner, repo, repoName, repodl, 0, false)
	uc.CheckForUpdate(appVersion)
	// uc.PrintMessage()
	updtmsg := uc.Message
	return updtmsg, uc.UpdateAvailable
}

// zeroPadNames renames files in the specified directory to have zero-padded numbers
// in their names. For example, it renames file_1.pdf to file_000
// for better sorting
func zeroPadNames(dir string, pad int) error {
	// Regex to match filenames like file_1.pdf, file_15.pdf, etc.
	re := regexp.MustCompile(`^.*_(\d+)\.pdf$`)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		name := info.Name()
		matches := re.FindStringSubmatch(name)
		if len(matches) == 2 {
			num, err := strconv.Atoi(matches[1])
			if err != nil {
				return err
			}

			paddedNum := fmt.Sprintf("%0*d", pad, num)
			// Create the new name with padded number
			// e.g., if num is 1 and pad is 4, paddedNum will be "0001"
			newName := fmt.Sprintf("file_%s.pdf", paddedNum)
			// newName := fmt.Sprintf("file_%03d.pdf", num)
			oldPath := filepath.Join(dir, name)
			newPath := filepath.Join(dir, newName)

			if oldPath != newPath {
				// fmt.Printf("Renaming %s → %s\n", name, newName)
				return os.Rename(oldPath, newPath)
			}
		}

		return nil
	})
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
