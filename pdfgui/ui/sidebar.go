package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ViewType represents different views in the application
type ViewType int

const (
	ViewEncrypt ViewType = iota
	ViewDecrypt
	ViewMerge
	ViewSplit
	ViewExtract
	ViewRemove
	ViewRotate
	ViewReverse
	ViewPermissionsView
	ViewPermissionsSet
	ViewPropertiesView
	ViewPropertiesEdit
	ViewOpenViewer
)

// CreateSidebar creates the navigation sidebar with grouped buttons
func CreateSidebar(onViewChange func(ViewType)) fyne.CanvasObject {
	var selectedButton *widget.Button
	
	// Helper to create a button with selection state
	createNavButton := func(label string, view ViewType) *widget.Button {
		btn := widget.NewButton(label, nil)
		
		// Set up the tap handler properly
		btn.OnTapped = func() {
			// Deselect previous button
			if selectedButton != nil {
				selectedButton.Importance = widget.MediumImportance
				selectedButton.Refresh()
			}
			// Select this button
			btn.Importance = widget.HighImportance
			selectedButton = btn
			btn.Refresh()
			
			// Call the view change handler
			onViewChange(view)
		}
		
		return btn
	}
	
	// File Operations Group
	fileOpsLabel := widget.NewLabelWithStyle("File Operations", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	encryptBtn := createNavButton("🔐 Encrypt PDF", ViewEncrypt)
	decryptBtn := createNavButton("🔓 Decrypt PDF", ViewDecrypt)
	mergeBtn := createNavButton("🔗 Merge PDFs", ViewMerge)
	splitBtn := createNavButton("✂️  Split PDF", ViewSplit)
	
	// Page Operations Group
	pageOpsLabel := widget.NewLabelWithStyle("Page Operations", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	extractBtn := createNavButton("📄 Extract Pages", ViewExtract)
	removeBtn := createNavButton("🗑️  Remove Pages", ViewRemove)
	rotateBtn := createNavButton("🔄 Rotate Pages", ViewRotate)
	reverseBtn := createNavButton("↕️  Reverse Pages", ViewReverse)
	
	// Security Group
	securityLabel := widget.NewLabelWithStyle("Security", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	viewPermsBtn := createNavButton("👁️  View Permissions", ViewPermissionsView)
	setPermsBtn := createNavButton("🔒 Set Permissions", ViewPermissionsSet)
	
	// Metadata Group
	metadataLabel := widget.NewLabelWithStyle("Metadata", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	viewPropsBtn := createNavButton("📋 View Properties", ViewPropertiesView)
	editPropsBtn := createNavButton("✏️  Edit Properties", ViewPropertiesEdit)

	// View Group
	viewLabel := widget.NewLabelWithStyle("View", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	openViewerBtn := createNavButton("👁️  View PDF", ViewOpenViewer)

	// Set initial selection
	encryptBtn.Importance = widget.HighImportance
	selectedButton = encryptBtn
	
	// Create sidebar content
	sidebar := container.NewVBox(
		canvas.NewText("Kranky Bear PDF", theme.ForegroundColor()),
		widget.NewSeparator(),
		
		fileOpsLabel,
		encryptBtn,
		decryptBtn,
		mergeBtn,
		splitBtn,
		widget.NewSeparator(),
		
		pageOpsLabel,
		extractBtn,
		removeBtn,
		rotateBtn,
		reverseBtn,
		widget.NewSeparator(),
		
		securityLabel,
		viewPermsBtn,
		setPermsBtn,
		widget.NewSeparator(),
		
		metadataLabel,
		viewPropsBtn,
		editPropsBtn,
		widget.NewSeparator(),

		viewLabel,
		openViewerBtn,
	)
	
	// Wrap in scroll for smaller screens
	return container.NewVScroll(sidebar)
}

