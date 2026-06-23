package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"pdfutil/shared"
)

// PropertiesState holds the currently loaded properties
type PropertiesState struct {
	title            string
	author           string
	subject          string
	keywords         string
	creator          string
	producer         string
	loaded           bool
	autoLoadViewFunc func() // Function to auto-load view properties
	autoLoadEditFunc func() // Function to auto-load edit properties
}

// CreatePropertiesTab creates the properties management UI tab
func CreatePropertiesTab(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	// Shared properties state between view and add/set tabs
	propsState := &PropertiesState{loaded: false}
	
	tabs := container.NewAppTabs(
		container.NewTabItem("View", createViewPropertiesSection(w, fileState, propsState)),
		container.NewTabItem("Add/Set", createSetPropertiesSection(w, fileState, propsState)),
	)
	
	return tabs
}

func createViewPropertiesSection(w fyne.Window, fileState *FileState, propsState *PropertiesState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	// Password fields with auto-population
	userPassEntry, ownerPassEntry := CreatePasswordEntries(fileState)

	resultText := widget.NewLabel("Select a file and click 'View Properties' or it will load automatically")
	resultText.Wrapping = fyne.TextWrapWord
	resultScroll := container.NewScroll(resultText)
	resultScroll.SetMinSize(fyne.NewSize(400, 300))

	// Clear properties display when file changes
	fileState.OnFileChange(func(file string) {
		resultText.SetText("File changed. Will auto-load properties if passwords available.")
		propsState.loaded = false
	})

	// Function to load and display properties
	loadAndDisplayProperties := func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}

		// Get passwords from entries
		userPw, ownerPw := GetPasswordsFromEntries(userPassEntry, ownerPassEntry, fileState)

		ctx, err := shared.GetFileInfo(selectedFile, userPw, ownerPw)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to get file info: %v", err), w)
			resultText.SetText("Error reading file properties. Check passwords if file is encrypted.")
			propsState.loaded = false
			return
		}

		// Store properties in shared state
		propsState.title = ctx.Title
		propsState.author = ctx.Author
		propsState.subject = ctx.Subject
		propsState.keywords = ctx.Keywords
		propsState.creator = ctx.Creator
		propsState.producer = ctx.Producer
		propsState.loaded = true

		// Display properties
		result := fmt.Sprintf("═══ Document Properties ═══\n\n")
		result += fmt.Sprintf("Title:          %s\n", getStringValue(ctx.Title))
		result += fmt.Sprintf("Author:         %s\n", getStringValue(ctx.Author))
		result += fmt.Sprintf("Subject:        %s\n", getStringValue(ctx.Subject))
		result += fmt.Sprintf("Keywords:       %s\n", getStringValue(ctx.Keywords))
		result += fmt.Sprintf("Creator:        %s\n", getStringValue(ctx.Creator))
		result += fmt.Sprintf("Producer:       %s\n", getStringValue(ctx.Producer))
		result += fmt.Sprintf("Creation Date:  %s\n", getStringValue(ctx.Configuration.CreationDate))
		result += fmt.Sprintf("Modified Date:  %s\n", getStringValue(ctx.ModDate))
		result += fmt.Sprintf("\nPDF Version:    %s\n", ctx.HeaderVersion)
		result += fmt.Sprintf("Page Count:     %d\n", ctx.PageCount)

		if ctx.Encrypt != nil {
			result += "\nEncryption:     Yes"
		} else {
			result += "\nEncryption:     No"
		}

		resultText.SetText(result)
	}

	viewButton := widget.NewButton("View Properties (Refresh)", loadAndDisplayProperties)

	// Store the load function so dashboard can trigger it
	propsState.autoLoadViewFunc = loadAndDisplayProperties

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("View Document Properties", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
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

func createSetPropertiesSection(w fyne.Window, fileState *FileState, propsState *PropertiesState) fyne.CanvasObject {
	fileLabel := widget.NewLabel("No file selected")
	fileSelector := CreateFileSelector(w, fileState, fileLabel)

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Document title")
	authorEntry := widget.NewEntry()
	authorEntry.SetPlaceHolder("Document author")
	subjectEntry := widget.NewEntry()
	subjectEntry.SetPlaceHolder("Document subject")
	keywordsEntry := widget.NewEntry()
	keywordsEntry.SetPlaceHolder("Document keywords")
	creatorEntry := widget.NewEntry()
	creatorEntry.SetPlaceHolder("Document creator")
	producerEntry := widget.NewEntry()
	producerEntry.SetPlaceHolder("Document producer")

	// Password fields with auto-population
	userPassEntry, ownerPassEntry := CreatePasswordEntries(fileState)

	// Function to populate fields from properties state
	populateFields := func() {
		titleEntry.SetText(propsState.title)
		authorEntry.SetText(propsState.author)
		subjectEntry.SetText(propsState.subject)
		keywordsEntry.SetText(propsState.keywords)
		creatorEntry.SetText(propsState.creator)
		producerEntry.SetText(propsState.producer)
	}

	// Function to load properties from file
	loadPropertiesFromFile := func(showDialog bool) {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			if showDialog {
				dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			}
			return
		}

		// Get passwords from entries
		userPw, ownerPw := GetPasswordsFromEntries(userPassEntry, ownerPassEntry, fileState)

		// Load properties
		ctx, err := shared.GetFileInfo(selectedFile, userPw, ownerPw)
		if err != nil {
			if showDialog {
				dialog.ShowError(fmt.Errorf("failed to read properties: %v", err), w)
			}
			return
		}

		// Store and populate
		propsState.title = ctx.Title
		propsState.author = ctx.Author
		propsState.subject = ctx.Subject
		propsState.keywords = ctx.Keywords
		propsState.creator = ctx.Creator
		propsState.producer = ctx.Producer
		propsState.loaded = true

		populateFields()
		if showDialog {
			dialog.ShowInformation("Loaded", "Current properties loaded into fields for editing", w)
		}
	}

	// Auto-load button (manual refresh/reset)
	autoLoadButton := widget.NewButton("Load Current Properties (Reset)", func() {
		loadPropertiesFromFile(true)
	})

	clearButton := widget.NewButton("Clear All Fields", func() {
		titleEntry.SetText("")
		authorEntry.SetText("")
		subjectEntry.SetText("")
		keywordsEntry.SetText("")
		creatorEntry.SetText("")
		producerEntry.SetText("")
	})

	// Store the load function so dashboard can trigger it
	propsState.autoLoadEditFunc = func() {
		// Auto-load silently when switching to this view
		loadPropertiesFromFile(false)
	}

	// Listen for file changes and clear fields
	lastLoadedFile := ""
	fileState.OnFileChange(func(file string) {
		if file != lastLoadedFile {
			propsState.loaded = false
			lastLoadedFile = file
			// Clear fields when file changes
			titleEntry.SetText("")
			authorEntry.SetText("")
			subjectEntry.SetText("")
			keywordsEntry.SetText("")
			creatorEntry.SetText("")
			producerEntry.SetText("")
		}
	})

	inPlaceCheck := widget.NewCheck("Update current file (in-place)", func(checked bool) {})
	inPlaceCheck.SetChecked(true)

	setButton := widget.NewButton("Save Properties", func() {
		selectedFile := fileState.GetFile()
		if selectedFile == "" {
			dialog.ShowError(fmt.Errorf("please select a PDF file"), w)
			return
		}

		// Build properties map from entry fields
		properties := make(map[string]string)
		if titleEntry.Text != "" {
			properties["Title"] = titleEntry.Text
		}
		if authorEntry.Text != "" {
			properties["Author"] = authorEntry.Text
		}
		if subjectEntry.Text != "" {
			properties["Subject"] = subjectEntry.Text
		}
		if keywordsEntry.Text != "" {
			properties["Keywords"] = keywordsEntry.Text
		}
		if creatorEntry.Text != "" {
			properties["Creator"] = creatorEntry.Text
		}
		if producerEntry.Text != "" {
			properties["Producer"] = producerEntry.Text
		}

		if len(properties) == 0 {
			dialog.ShowError(fmt.Errorf("please enter at least one property"), w)
			return
		}

		// Get passwords
		userPw, ownerPw := GetPasswordsFromEntries(userPassEntry, ownerPassEntry, fileState)

		if inPlaceCheck.Checked {
			// In-place modification (transparent temp file handling)
			err := SavePropertiesInPlace(selectedFile, userPw, ownerPw, properties, w)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to set properties: %v", err), w)
				return
			}
			
			dialog.ShowInformation("Success", "Properties updated successfully!", w)
			
			// Reload properties if they were previously loaded
			if propsState.loaded {
				propsState.title = properties["Title"]
				propsState.author = properties["Author"]
				propsState.subject = properties["Subject"]
				propsState.keywords = properties["Keywords"]
				propsState.creator = properties["Creator"]
				propsState.producer = properties["Producer"]
			}
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

				err = shared.AddProperties(selectedFile, writer.URI().Path(), userPw, ownerPw, properties)
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to set properties: %v", err), w)
					return
				}

				dialog.ShowInformation("Success", fmt.Sprintf("Properties saved to:\n%s", writer.URI().Path()), w)
			}, w, fileState)
		}
	})

	scrollContent := container.NewVBox(
		widget.NewLabelWithStyle("Add/Edit Document Properties", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fileSelector,
		widget.NewSeparator(),
		container.NewHBox(autoLoadButton, clearButton),
		widget.NewLabel("Tip: Properties auto-load when you open this view. Use 'Reset' button to reload."),
		widget.NewSeparator(),
		widget.NewLabel("Enter properties:"),
		widget.NewForm(
			widget.NewFormItem("Title", titleEntry),
			widget.NewFormItem("Author", authorEntry),
			widget.NewFormItem("Subject", subjectEntry),
			widget.NewFormItem("Keywords", keywordsEntry),
			widget.NewFormItem("Creator", creatorEntry),
			widget.NewFormItem("Producer", producerEntry),
		),
		widget.NewSeparator(),
		widget.NewLabel("If file is encrypted:"),
		widget.NewForm(
			widget.NewFormItem("User Password", userPassEntry),
			widget.NewFormItem("Owner Password", ownerPassEntry),
		),
		inPlaceCheck,
		setButton,
	)

	return container.NewPadded(container.NewScroll(scrollContent))
}

func getStringValue(value string) string {
	if value == "" {
		return "(not set)"
	}
	return value
}
