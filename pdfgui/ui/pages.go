package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"pdfutil/shared"
)

// CreatePagesTab creates the page operations UI tab
func CreatePagesTab(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Extract", createExtractSection(w, fileState)),
		container.NewTabItem("Remove", createRemoveSection(w, fileState)),
		container.NewTabItem("Rotate", createRotateSection(w, fileState)),
		container.NewTabItem("Reverse", createReverseSection(w, fileState)),
	)
	return tabs
}

func createExtractSection(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	pagesEntry := widget.NewEntry()
	pagesEntry.SetPlaceHolder("e.g., 1,3-5,10-l")

	padCheck := widget.NewCheck("Zero-pad filenames", func(checked bool) {})
	padCheck.SetChecked(true)

	extractButton := widget.NewButton("Extract Pages", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}
		if pagesEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("please specify pages to extract"), w)
			return
		}

		if shared.IsEncrypted(selectedFile) {
			dialog.ShowError(fmt.Errorf("encrypted files cannot be extracted"), w)
			return
		}

		ShowLargeFolderOpen(func(folder fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if folder == nil {
				return
			}

			pages := strings.Split(strings.ReplaceAll(pagesEntry.Text, " ", ""), ",")
			err = shared.ExtractPages(selectedFile, folder.Path(), pages, padCheck.Checked, "", "")
			if err != nil {
				dialog.ShowError(fmt.Errorf("extraction failed: %v", err), w)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Pages extracted to:\n%s", folder.Path()), w)
		}, w, fileState)
	})

	return container.NewPadded(
		container.NewVBox(
			widget.NewLabelWithStyle("Extract Pages", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			fileSelector,
			widget.NewForm(
				widget.NewFormItem("Pages (e.g., 1,3-5,10-l)", pagesEntry),
			),
			padCheck,
			extractButton,
		),
	)
}

func createRemoveSection(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	pagesEntry := widget.NewEntry()
	pagesEntry.SetPlaceHolder("e.g., 1,3-5,10-l")

	// Helper text
	helpText := widget.NewLabel("💡 Specify pages to remove:")
	helpText.Wrapping = fyne.TextWrapWord
	
	examplesText := widget.NewLabel("Examples: '1' (page 1), '1,3,5' (pages 1,3,5), '1-5' (pages 1-5), '10-l' (page 10 to last)")
	examplesText.Wrapping = fyne.TextWrapWord

	removeButton := widget.NewButton("Remove Pages", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}
		if pagesEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("please specify pages to remove"), w)
			return
		}

		if shared.IsEncrypted(selectedFile) {
			dialog.ShowError(fmt.Errorf("encrypted files cannot be modified"), w)
			return
		}

		ShowLargeFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			pages := strings.Split(strings.ReplaceAll(pagesEntry.Text, " ", ""), ",")
			err = shared.RemovePages(selectedFile, writer.URI().Path(), pages)
			if err != nil {
				dialog.ShowError(fmt.Errorf("removal failed: %v", err), w)
				return
			}

			dialog.ShowInformation("Success", "Pages removed successfully!", w)
		}, w, fileState)
	})

	return container.NewPadded(
		container.NewVBox(
			widget.NewLabelWithStyle("Remove Pages", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Select a PDF and specify which pages to remove"),
			widget.NewSeparator(),
			fileSelector,
			widget.NewSeparator(),
			helpText,
			widget.NewForm(
				widget.NewFormItem("Pages to remove", pagesEntry),
			),
			examplesText,
			widget.NewSeparator(),
			removeButton,
		),
	)
}

func createRotateSection(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	pagesEntry := widget.NewEntry()
	pagesEntry.SetPlaceHolder("e.g., 1,3-5,10-l (or 1-l for all)")

	rotationSelect := widget.NewSelect([]string{"90", "180", "270"}, func(s string) {})
	rotationSelect.SetSelected("90")

	rotateButton := widget.NewButton("Rotate Pages", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}
		if pagesEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("please specify pages to rotate"), w)
			return
		}

		if shared.IsEncrypted(selectedFile) {
			dialog.ShowError(fmt.Errorf("encrypted files cannot be modified"), w)
			return
		}

		ShowLargeFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			pages := strings.Split(strings.ReplaceAll(pagesEntry.Text, " ", ""), ",")
			rotation := 90
			if rotationSelect.Selected == "180" {
				rotation = 180
			} else if rotationSelect.Selected == "270" {
				rotation = 270
			}

			err = shared.RotatePages(selectedFile, writer.URI().Path(), pages, rotation)
			if err != nil {
				dialog.ShowError(fmt.Errorf("rotation failed: %v", err), w)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Pages rotated %d degrees!", rotation), w)
		}, w, fileState)
	})

	return container.NewPadded(
		container.NewVBox(
			widget.NewLabelWithStyle("Rotate Pages", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			fileSelector,
			widget.NewForm(
				widget.NewFormItem("Pages (e.g., 1-l for all)", pagesEntry),
				widget.NewFormItem("Rotation angle", rotationSelect),
			),
			rotateButton,
		),
	)
}

func createReverseSection(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	reverseButton := widget.NewButton("Reverse All Pages", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}

		if shared.IsEncrypted(selectedFile) {
			dialog.ShowError(fmt.Errorf("encrypted files cannot be modified"), w)
			return
		}

		ShowLargeFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			err = shared.ReversePages(selectedFile, writer.URI().Path(), false, "", "")
			if err != nil {
				dialog.ShowError(fmt.Errorf("reverse failed: %v", err), w)
				return
			}

			dialog.ShowInformation("Success", "Pages reversed successfully!", w)
		}, w, fileState)
	})

	return container.NewPadded(
		container.NewVBox(
			widget.NewLabelWithStyle("Reverse Pages", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("This will reverse the order of all pages in the PDF"),
			fileSelector,
			reverseButton,
		),
	)
}
