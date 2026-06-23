package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"pdfutil/shared"
)

// CreateMergeSplitTab creates the merge/split UI tab
func CreateMergeSplitTab(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	return container.NewVBox(
		createMergeSection(w, fileState),
		widget.NewSeparator(),
		createSplitSection(w, fileState),
	)
}

func createMergeSection(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	var selectedFiles []string
	var selectedIndex int = -1
	fileCountLabel := widget.NewLabel("0 files selected")
	
	fileList := widget.NewList(
		func() int { return len(selectedFiles) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(selectedFiles[id])
		},
	)
	
	// Track which item is selected
	fileList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
	}
	
	fileList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
	}

	updateCount := func() {
		fileCountLabel.SetText(fmt.Sprintf("%d file(s) selected", len(selectedFiles)))
	}

	addFilesButton := widget.NewButton("➕ Add PDF File", func() {
		ShowLargeFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			filePath := reader.URI().Path()
			selectedFiles = append(selectedFiles, filePath)
			fileList.Refresh()
			updateCount()
			reader.Close()
			
			// Update fileState to remember this directory for next time
			// (SetFile does this automatically by setting lastDirectory)
			if fileState != nil && len(selectedFiles) > 0 {
				fileState.SetFile(selectedFiles[len(selectedFiles)-1])
			}
		}, w, fileState)
	})

	clearButton := widget.NewButton("Clear List", func() {
		selectedFiles = []string{}
		fileList.Refresh()
		updateCount()
	})
	
	removeButton := widget.NewButton("Remove Selected", func() {
		if selectedIndex >= 0 && selectedIndex < len(selectedFiles) {
			selectedFiles = append(selectedFiles[:selectedIndex], selectedFiles[selectedIndex+1:]...)
			selectedIndex = -1
			fileList.UnselectAll()
			fileList.Refresh()
			updateCount()
		} else {
			dialog.ShowInformation("No Selection", "Please select a file from the list first, then click Remove", w)
		}
	})

	mergeButton := widget.NewButton("Merge PDFs", func() {
		if len(selectedFiles) < 2 {
			dialog.ShowError(fmt.Errorf("please select at least 2 PDF files to merge"), w)
			return
		}

		// Check if any files are encrypted
		for _, file := range selectedFiles {
			if shared.IsEncrypted(file) {
				dialog.ShowError(fmt.Errorf("encrypted files cannot be merged: %s", file), w)
				return
			}
		}

		// Show save dialog
		ShowLargeFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			outFile := writer.URI().Path()
			err = shared.MergePDFs(selectedFiles, outFile)
			if err != nil {
				dialog.ShowError(fmt.Errorf("merge failed: %v", err), w)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Merged %d files into:\n%s", len(selectedFiles), outFile), w)
		}, w, fileState)
	})

	helpText := widget.NewLabel("💡 Tip: Click 'Add PDF File' multiple times to add files in order. The dialog remembers your location!")
	helpText.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Merge PDFs", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			helpText,
			widget.NewSeparator(),
			fileCountLabel,
			container.NewHBox(addFilesButton, removeButton, clearButton),
		),
		container.NewPadded(mergeButton),
		nil, nil,
		fileList,
	)
}

func createSplitSection(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	padCheck := widget.NewCheck("Zero-pad filenames", func(checked bool) {})
	padCheck.SetChecked(true)

	splitButton := widget.NewButton("Split into Individual Pages", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file first"), w)
			return
		}

		// Check if encrypted
		if shared.IsEncrypted(selectedFile) {
			dialog.ShowError(fmt.Errorf("encrypted files cannot be split"), w)
			return
		}

		// Show folder selection dialog
		ShowLargeFolderOpen(func(folder fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if folder == nil {
				return
			}

			outDir := folder.Path()
			err = shared.SplitPDF(selectedFile, outDir, padCheck.Checked, "", "")
			if err != nil {
				dialog.ShowError(fmt.Errorf("split failed: %v", err), w)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("PDF split successfully into:\n%s", outDir), w)
		}, w, fileState)
	})

	return container.NewPadded(
		container.NewVBox(
			widget.NewLabelWithStyle("Split PDF", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			fileSelector,
			padCheck,
			splitButton,
		),
	)
}
