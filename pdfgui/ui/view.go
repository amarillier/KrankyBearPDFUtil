package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// prefViewerMode persists the user's choice of viewer: "builtin" or "system".
const prefViewerMode = "viewer.mode"

// goosLabel maps runtime.GOOS to the label used in our binary names (darwin -> macos).
func goosLabel() string {
	if runtime.GOOS == "darwin" {
		return "macos"
	}
	return runtime.GOOS
}

// isExecutableFile reports whether p is a regular, runnable file.
func isExecutableFile(p string) bool {
	fi, err := os.Stat(p)
	if err != nil || fi.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return fi.Mode()&0o111 != 0
}

// locateBundledViewer finds the pdfviewer binary that shipped alongside pdfgui, or "".
// It looks next to the running executable (covers the macOS .app/Contents/MacOS and the
// Windows install dir), the Linux /opt install dir, and finally PATH and dev build outputs.
func locateBundledViewer() string {
	name := "pdfviewer"
	archName := "pdfviewer-" + goosLabel() + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		name += ".exe"
		archName += ".exe"
	}

	var dirs []string
	if exe, err := os.Executable(); err == nil {
		if resolved, err2 := filepath.EvalSymlinks(exe); err2 == nil {
			exe = resolved
		}
		dirs = append(dirs, filepath.Dir(exe))
	}
	dirs = append(dirs, "/opt/KrankyBearPDF")
	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(wd, "bin"), wd)
	}

	for _, d := range dirs {
		for _, n := range []string{name, archName} {
			p := filepath.Join(d, n)
			if isExecutableFile(p) {
				return p
			}
		}
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return ""
}

// launchBundledViewer starts the bundled pdfviewer on the given file (detached).
func launchBundledViewer(viewerPath, file string) error {
	cmd := exec.Command(viewerPath, file)
	return cmd.Start()
}

// openWithSystemDefault opens the file in the OS default PDF application.
func openWithSystemDefault(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// createViewerSection builds the "View PDF" page: open the selected PDF in either the
// bundled basic viewer or the system default viewer (user-selectable, persisted).
func createViewerSection(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	prefs := fyne.CurrentApp().Preferences()
	bundled := locateBundledViewer()

	statusLabel := widget.NewLabel("")
	if bundled != "" {
		statusLabel.SetText("Built-in viewer found: " + filepath.Base(bundled))
	} else {
		statusLabel.SetText("Built-in viewer not found — will use the system default viewer.")
	}
	statusLabel.Wrapping = fyne.TextWrapWord

	const optBuiltin = "Built-in viewer (basic)"
	const optSystem = "System default viewer"
	modeRadio := widget.NewRadioGroup([]string{optBuiltin, optSystem}, func(s string) {
		if s == optSystem {
			prefs.SetString(prefViewerMode, "system")
		} else {
			prefs.SetString(prefViewerMode, "builtin")
		}
	})
	if prefs.StringWithFallback(prefViewerMode, "builtin") == "system" {
		modeRadio.SetSelected(optSystem)
	} else {
		modeRadio.SetSelected(optBuiltin)
	}

	openButton := widget.NewButton("👁️  Open in Viewer", func() {
		file := fileState.GetFile()
		if file == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}

		useSystem := prefs.StringWithFallback(prefViewerMode, "builtin") == "system"

		// System default requested, or no built-in viewer available.
		if useSystem || bundled == "" {
			if err := openWithSystemDefault(file); err != nil {
				dialog.ShowError(fmt.Errorf("could not open with the system viewer: %w", err), w)
			}
			return
		}

		// Try the bundled viewer; on failure (e.g. an amd64 fallback binary on arm64
		// Linux), fall back to the system default rather than leaving the user stuck.
		if err := launchBundledViewer(bundled, file); err != nil {
			if err2 := openWithSystemDefault(file); err2 != nil {
				dialog.ShowError(fmt.Errorf("built-in viewer failed (%v) and system viewer failed (%v)", err, err2), w)
			} else {
				dialog.ShowInformation("Viewer",
					"The built-in viewer could not run here, so the PDF was opened in your system default viewer.", w)
			}
		}
	})

	return container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("View PDF", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Open the selected PDF in a viewer."),
		fileSelector,
		widget.NewSeparator(),
		widget.NewForm(widget.NewFormItem("Open with", modeRadio)),
		statusLabel,
		openButton,
	))
}
