package ui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateDashboard creates the main dashboard with file info and content area
func CreateDashboard(w fyne.Window, fileState *FileState, propsState *PropertiesState) fyne.CanvasObject {
	// File info panel at top
	fileInfoLabel := widget.NewLabel("No file selected")
	fileInfoLabel.Wrapping = fyne.TextWrapWord
	
	fileSelector := CreateFileSelector(w, fileState, widget.NewLabel(""))
	
	// Update file info when file changes
	fileState.OnFileChange(func(file string) {
		if file == "" {
			fileInfoLabel.SetText("No file selected - Select a file to begin")
		} else {
			fileInfoLabel.SetText(fmt.Sprintf("Current file: %s\nLocation: %s", 
				filepath.Base(file), 
				filepath.Dir(file)))
		}
	})
	
	fileInfoPanel := container.NewVBox(
		widget.NewLabelWithStyle("Current File", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fileInfoLabel,
		fileSelector,
		widget.NewSeparator(),
	)
	
	// Content area (will be swapped based on sidebar selection)
	contentArea := container.NewStack()
	
	// Create permissions state for auto-load
	permsState := &PermissionsState{}
	
	// Create all view contents
	views := map[ViewType]fyne.CanvasObject{
		ViewEncrypt:         createEncryptView(w, fileState),
		ViewDecrypt:         createDecryptView(w, fileState),
		ViewMerge:           createMergeView(w, fileState),
		ViewSplit:           createSplitView(w, fileState),
		ViewExtract:         createExtractSection(w, fileState),
		ViewRemove:          createRemoveSection(w, fileState),
		ViewRotate:          createRotateSection(w, fileState),
		ViewReverse:         createReverseSection(w, fileState),
		ViewPermissionsView: createViewPermissionsSection(w, fileState, permsState),
		ViewPermissionsSet:  createSetPermissionsSection(w, fileState),
		ViewPropertiesView:  createViewPropertiesSection(w, fileState, propsState),
		ViewPropertiesEdit:  createSetPropertiesSection(w, fileState, propsState),
		ViewOpenViewer:      createViewerSection(w, fileState),
	}
	
	// Set initial view
	currentView := ViewEncrypt
	contentArea.Objects = []fyne.CanvasObject{views[currentView]}
	
	// Create sidebar
	sidebar := CreateSidebar(func(view ViewType) {
		currentView = view
		contentArea.Objects = []fyne.CanvasObject{views[currentView]}
		contentArea.Refresh()
		
		// Auto-trigger actions when switching to certain views
		switch view {
		case ViewPermissionsView:
			// Auto-load and display permissions when viewing
			if permsState.autoLoadViewFunc != nil && fileState.GetFile() != "" {
				permsState.autoLoadViewFunc()
			}
		case ViewPropertiesView:
			// Auto-load and display properties when viewing
			if propsState.autoLoadViewFunc != nil && fileState.GetFile() != "" {
				propsState.autoLoadViewFunc()
			}
		case ViewPropertiesEdit:
			// Auto-load properties into edit fields when editing
			if propsState.autoLoadEditFunc != nil && fileState.GetFile() != "" {
				propsState.autoLoadEditFunc()
			}
		}
	})
	
	// Main layout: sidebar | (file info + content)
	mainContent := container.NewBorder(fileInfoPanel, nil, nil, nil, contentArea)
	
	// Create split with fixed narrow sidebar width
	split := container.NewHSplit(sidebar, mainContent)
	split.SetOffset(0.20) // Sidebar takes 20% of width (narrower)
	
	return split
}

// Create individual view components (split from existing tabs)
func createEncryptView(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	return createEncryptDecryptContent(w, fileState, true)
}

func createDecryptView(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	return createEncryptDecryptContent(w, fileState, false)
}

func createMergeView(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	return createMergeSection(w, fileState)
}

func createSplitView(w fyne.Window, fileState *FileState) fyne.CanvasObject {
	return createSplitSection(w, fileState)
}

