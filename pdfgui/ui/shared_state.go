package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// FileState holds the currently selected file and notifies listeners of changes
type FileState struct {
	currentFile     string
	lastDirectory   string // Remember last browsed directory
	userPassword    string // Remember user password
	ownerPassword   string // Remember owner password
	listeners       []func(string)
	passwordListeners []func(string, string) // Notify when passwords change
}

// NewFileState creates a new FileState
func NewFileState() *FileState {
	// Start in current working directory
	startDir, err := os.Getwd()
	if err != nil {
		// Fallback to home directory if can't get current working directory
		startDir, _ = os.UserHomeDir()
	}
	
	return &FileState{
		currentFile:       "",
		lastDirectory:     startDir,
		userPassword:      "",
		ownerPassword:     "",
		listeners:         make([]func(string), 0),
		passwordListeners: make([]func(string, string), 0),
	}
}

// GetFile returns the current file
func (fs *FileState) GetFile() string {
	return fs.currentFile
}

// SetFile updates the current file and notifies listeners
func (fs *FileState) SetFile(file string) {
	fs.currentFile = file
	// Remember the directory for next time
	if file != "" {
		fs.lastDirectory = filepath.Dir(file)
	}
	// Clear passwords when file changes
	fs.userPassword = ""
	fs.ownerPassword = ""
	for _, listener := range fs.listeners {
		listener(file)
	}
	// Notify password listeners to clear fields
	for _, listener := range fs.passwordListeners {
		listener("", "")
	}
}

// GetLastDirectory returns the last browsed directory
func (fs *FileState) GetLastDirectory() string {
	return fs.lastDirectory
}

// SetPasswords updates the stored passwords
func (fs *FileState) SetPasswords(userPw, ownerPw string) {
	fs.userPassword = userPw
	fs.ownerPassword = ownerPw
	// Notify listeners
	for _, listener := range fs.passwordListeners {
		listener(userPw, ownerPw)
	}
}

// GetPasswords returns the stored passwords
func (fs *FileState) GetPasswords() (string, string) {
	return fs.userPassword, fs.ownerPassword
}

// OnPasswordChange registers a listener for password changes
func (fs *FileState) OnPasswordChange(listener func(string, string)) {
	fs.passwordListeners = append(fs.passwordListeners, listener)
}

// OnFileChange registers a listener for file changes
func (fs *FileState) OnFileChange(listener func(string)) {
	fs.listeners = append(fs.listeners, listener)
}

// CreateFileSelector creates a standard file selector widget with the current file
func CreateFileSelector(w fyne.Window, fileState *FileState, label *widget.Label) *fyne.Container {
	// Update label with current file or default text
	updateLabel := func(file string) {
		if file == "" {
			label.SetText("No file selected")
		} else {
			label.SetText(filepath.Base(file) + " (" + filepath.Dir(file) + ")")
		}
	}

	// Initialize label with current file
	updateLabel(fileState.GetFile())

	// Listen for file changes
	fileState.OnFileChange(updateLabel)

	fileButton := widget.NewButton("Select PDF File", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			fileState.SetFile(reader.URI().Path())
			reader.Close()
		}, w)
		
		// Make dialog larger
		fd.Resize(fyne.NewSize(800, 600))
		
		// Filter to show only PDF files (much faster in large directories)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
		
		// Start in last used directory
		if lastDir := fileState.GetLastDirectory(); lastDir != "" {
			if uri := storage.NewFileURI(lastDir); uri != nil {
				lister, err := storage.ListerForURI(uri)
				if err == nil {
					fd.SetLocation(lister)
				}
			}
		}
		
		fd.Show()
	})

	clearButton := widget.NewButton("Clear", func() {
		fileState.SetFile("")
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("File Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, nil, container.NewHBox(fileButton, clearButton), label),
	)
}

// ShowLargeFileOpen shows a file open dialog with larger size and PDF filter
func ShowLargeFileOpen(callback func(fyne.URIReadCloser, error), w fyne.Window, fileState *FileState) {
	fd := dialog.NewFileOpen(callback, w)
	fd.Resize(fyne.NewSize(800, 600))
	// Filter to show only PDF files
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	
	// Default to current working directory or last browsed directory
	var startDir string
	if fileState != nil {
		lastDir := fileState.GetLastDirectory()
		if lastDir != "" {
			startDir = lastDir
		}
	}
	
	// Fallback to current working directory if no last directory
	if startDir == "" {
		cwd, err := os.Getwd()
		if err == nil {
			startDir = cwd
		}
	}
	
	if startDir != "" {
		if uri := storage.NewFileURI(startDir); uri != nil {
			lister, err := storage.ListerForURI(uri)
			if err == nil {
				fd.SetLocation(lister)
			}
		}
	}
	
	fd.Show()
}

// ShowLargeFileSave shows a file save dialog with larger size
func ShowLargeFileSave(callback func(fyne.URIWriteCloser, error), w fyne.Window, fileState *FileState) {
	fd := dialog.NewFileSave(callback, w)
	fd.Resize(fyne.NewSize(800, 600))
	// Suggest .pdf extension
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	fd.SetFileName("output.pdf")
	
	// Default to current file's directory if available
	if fileState != nil && fileState.GetFile() != "" {
		currentDir := filepath.Dir(fileState.GetFile())
		if uri := storage.NewFileURI(currentDir); uri != nil {
			lister, err := storage.ListerForURI(uri)
			if err == nil {
				fd.SetLocation(lister)
			}
		}
	}
	
	fd.Show()
}

// ShowLargeFolderOpen shows a folder open dialog with larger size
func ShowLargeFolderOpen(callback func(fyne.ListableURI, error), w fyne.Window, fileState *FileState) {
	fd := dialog.NewFolderOpen(callback, w)
	fd.Resize(fyne.NewSize(800, 600))
	
	// Default to current file's directory if available
	if fileState != nil && fileState.GetFile() != "" {
		currentDir := filepath.Dir(fileState.GetFile())
		if uri := storage.NewFileURI(currentDir); uri != nil {
			lister, err := storage.ListerForURI(uri)
			if err == nil {
				fd.SetLocation(lister)
			}
		}
	}
	
	fd.Show()
}

// ShowLargeFileOpenAt shows a file open dialog starting at a specific location
func ShowLargeFileOpenAt(callback func(fyne.URIReadCloser, error), w fyne.Window, startDir string) {
	fd := dialog.NewFileOpen(callback, w)
	fd.Resize(fyne.NewSize(800, 600))
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	
	// Set starting location
	if startDir != "" {
		if uri := storage.NewFileURI(startDir); uri != nil {
			lister, err := storage.ListerForURI(uri)
			if err == nil {
				fd.SetLocation(lister)
			}
		}
	}
	
	fd.Show()
}

