package ui

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// PDFViewer handles PDF viewing functionality
type PDFViewer struct {
	window        fyne.Window
	app           fyne.App
	
	// PDF state
	pdfPath       string
	currentPage   int
	totalPages    int
	zoomLevel     float32
	
	// Rendering
	renderer      *PDFRenderer
	currentImage  *canvas.Image
	imageScroll   *container.Scroll
	
	// UI components
	pageEntry     *widget.Entry
	pageLabel     *widget.Label
	zoomSelect    *widget.Select
	statusBar     *widget.Label
	toolbar       *fyne.Container
	bookmarkPanel *BookmarkPanel
	split         *container.Split
	mainContent   *fyne.Container
	
	// State management
	mu              sync.RWMutex
	lastDir         string
	showBookmarks   bool
}

// NewPDFViewer creates a new PDF viewer instance
func NewPDFViewer(w fyne.Window, a fyne.App) *PDFViewer {
	v := &PDFViewer{
		window:        w,
		app:           a,
		currentPage:   1,
		zoomLevel:     1.0,
		showBookmarks: false,
	}
	
	v.renderer = NewPDFRenderer()
	
	return v
}

// Build constructs the viewer UI
func (v *PDFViewer) Build() *fyne.Container {
	// Create toolbar
	v.toolbar = v.createToolbar()
	
	// Create main viewing area
	v.currentImage = canvas.NewImageFromResource(theme.DocumentIcon())
	v.currentImage.FillMode = canvas.ImageFillContain
	v.currentImage.SetMinSize(fyne.NewSize(600, 800))
	
	v.imageScroll = container.NewScroll(v.currentImage)
	v.imageScroll.SetMinSize(fyne.NewSize(600, 600))
	
	// Create status bar
	v.statusBar = widget.NewLabel("Open a PDF to begin viewing")
	v.statusBar.Alignment = fyne.TextAlignCenter
	
	// Create bookmark panel
	v.bookmarkPanel = NewBookmarkPanel(v)
	
	// Setup tree selection callback
	v.bookmarkPanel.tree.OnSelected = func(uid string) {
		selectedBookmarkUID = uid
		v.bookmarkPanel.OnBookmarkSelected(uid)
	}
	
	// PDF viewer content
	pdfContent := container.NewBorder(
		v.toolbar,    // top
		v.statusBar,  // bottom
		nil,          // left
		nil,          // right
		v.imageScroll, // center
	)
	
	// Create split view with bookmarks (initially hidden)
	v.split = container.NewHSplit(v.bookmarkPanel.GetContainer(), pdfContent)
	v.split.SetOffset(0.2) // Bookmarks take 20% of width
	
	// Main content starts without bookmarks visible
	v.mainContent = container.NewMax(pdfContent)
	
	return v.mainContent
}

// createToolbar builds the toolbar with navigation and zoom controls
func (v *PDFViewer) createToolbar() *fyne.Container {
	// Navigation buttons
	firstBtn := widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), func() {
		v.JumpToPage(1)
	})
	firstBtn.Importance = widget.LowImportance
	
	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		v.PreviousPage()
	})
	
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		v.NextPage()
	})
	
	lastBtn := widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), func() {
		v.JumpToPage(v.totalPages)
	})
	lastBtn.Importance = widget.LowImportance
	
	// Page entry
	v.pageEntry = widget.NewEntry()
	v.pageEntry.SetPlaceHolder("Page")
	v.pageEntry.OnSubmitted = func(s string) {
		if pageNum, err := strconv.Atoi(s); err == nil {
			v.JumpToPage(pageNum)
		}
	}
	
	v.pageLabel = widget.NewLabel("of 0")
	
	// Zoom controls
	v.zoomSelect = widget.NewSelect([]string{
		"Fit Width",
		"50%",
		"75%",
		"100%",
		"125%",
		"150%",
		"200%",
		"300%",
	}, func(s string) {
		v.handleZoomChange(s)
	})
	v.zoomSelect.SetSelected("Fit Width")
	
	// Layout toolbar
	navSection := container.NewHBox(
		firstBtn,
		prevBtn,
		layout.NewSpacer(),
		container.New(layout.NewFormLayout(),
			widget.NewLabel("Page:"),
			container.NewGridWithColumns(2,
				v.pageEntry,
				v.pageLabel,
			),
		),
		layout.NewSpacer(),
		nextBtn,
		lastBtn,
	)
	
	zoomSection := container.NewHBox(
		widget.NewLabel("Zoom:"),
		v.zoomSelect,
	)
	
	toolbar := container.NewBorder(
		nil, nil,
		navSection,
		zoomSection,
		nil,
	)
	
	return toolbar
}

// OpenPDFDialog shows a file picker to open a PDF
func (v *PDFViewer) OpenPDFDialog() {
	fd := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, v.window)
			return
		}
		if uc == nil {
			return
		}
		defer uc.Close()
		
		path := uc.URI().Path()
		v.lastDir = uc.URI().Path()
		
		if err := v.LoadPDF(path); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to load PDF: %v", err), v.window)
		}
	}, v.window)
	
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	fd.Resize(fyne.NewSize(850, 700))
	
	if v.lastDir != "" {
		if uri := storage.NewFileURI(v.lastDir); uri != nil {
			lister, err := storage.ListerForURI(uri)
			if err == nil {
				fd.SetLocation(lister)
			}
		}
	}
	
	fd.Show()
}

// LoadPDF loads a PDF file
func (v *PDFViewer) LoadPDF(path string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	// Load the PDF
	if err := v.renderer.LoadPDF(path); err != nil {
		return fmt.Errorf("failed to load PDF: %w", err)
	}
	
	v.pdfPath = path
	v.totalPages = v.renderer.GetPageCount()
	v.currentPage = 1
	
	// Update UI
	v.updatePageDisplay()
	v.statusBar.SetText(fmt.Sprintf("Loaded: %s (%d pages)", path, v.totalPages))
	
	// Load bookmarks
	if err := v.bookmarkPanel.LoadBookmarks(path); err != nil {
		// Non-fatal error - PDF might not have bookmarks
		v.statusBar.SetText(fmt.Sprintf("Loaded: %s (%d pages, no bookmarks)", path, v.totalPages))
	} else if v.bookmarkPanel.bookmarkManager.HasBookmarks() {
		v.statusBar.SetText(fmt.Sprintf("Loaded: %s (%d pages, %d bookmarks)", 
			path, v.totalPages, len(v.bookmarkPanel.bookmarkManager.GetBookmarks())))
	}
	
	// Render first page
	return v.renderCurrentPage()
}

// NextPage advances to the next page
func (v *PDFViewer) NextPage() {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.currentPage < v.totalPages {
		v.currentPage++
		v.updatePageDisplay()
		if err := v.renderCurrentPage(); err != nil {
			if v.statusBar != nil {
				v.statusBar.SetText(fmt.Sprintf("Error: %v", err))
			}
			dialog.ShowError(err, v.window)
		}
	}
}

// PreviousPage goes back to the previous page
func (v *PDFViewer) PreviousPage() {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.currentPage > 1 {
		v.currentPage--
		v.updatePageDisplay()
		if err := v.renderCurrentPage(); err != nil {
			if v.statusBar != nil {
				v.statusBar.SetText(fmt.Sprintf("Error: %v", err))
			}
			dialog.ShowError(err, v.window)
		}
	}
}

// JumpToPage jumps to a specific page number
func (v *PDFViewer) JumpToPage(page int) {
	v.JumpToPageAndPosition(page, 0)
}

// JumpToPageAndPosition jumps to a specific page and scroll position
func (v *PDFViewer) JumpToPageAndPosition(page int, yOffset float32) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if page < 1 || page > v.totalPages {
		return
	}
	
	v.currentPage = page
	v.updatePageDisplay()
	if err := v.renderCurrentPage(); err != nil {
		if v.statusBar != nil {
			v.statusBar.SetText(fmt.Sprintf("Error: %v", err))
		}
		dialog.ShowError(err, v.window)
		return
	}
	
	// Set scroll position after rendering
	if yOffset > 0 {
		v.SetScrollYOffset(yOffset)
	}
}

// handleZoomChange handles zoom level changes
func (v *PDFViewer) handleZoomChange(zoom string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	switch zoom {
	case "Fit Width":
		v.zoomLevel = -1.0 // Special value for fit width
	case "50%":
		v.zoomLevel = 0.5
	case "75%":
		v.zoomLevel = 0.75
	case "100%":
		v.zoomLevel = 1.0
	case "125%":
		v.zoomLevel = 1.25
	case "150%":
		v.zoomLevel = 1.5
	case "200%":
		v.zoomLevel = 2.0
	case "300%":
		v.zoomLevel = 3.0
	}
	
	// Only render if there's a PDF loaded
	if v.pdfPath == "" {
		return
	}
	
	if err := v.renderCurrentPage(); err != nil {
		// Check if statusBar exists before using it (might be called during initialization)
		if v.statusBar != nil {
			v.statusBar.SetText(fmt.Sprintf("Error: %v", err))
		}
		dialog.ShowError(err, v.window)
	}
}

// updatePageDisplay updates the page number display
func (v *PDFViewer) updatePageDisplay() {
	v.pageEntry.SetText(fmt.Sprintf("%d", v.currentPage))
	v.pageLabel.SetText(fmt.Sprintf("of %d", v.totalPages))
}

// renderCurrentPage renders the current page
func (v *PDFViewer) renderCurrentPage() error {
	if v.pdfPath == "" {
		return fmt.Errorf("no PDF loaded")
	}
	
	// Calculate effective zoom
	effectiveZoom := v.zoomLevel
	if effectiveZoom == -1.0 {
		// Fit to width - use scroll container width
		effectiveZoom = 1.0 // Will be adjusted by Fyne's ImageFillContain
	}
	
	// Render the page
	img, err := v.renderer.RenderPage(v.currentPage, effectiveZoom)
	if err != nil {
		return fmt.Errorf("failed to render page: %w", err)
	}
	
	// Update the image - use canvas.NewImageFromImage to properly create from image.Image
	// Clear any existing resource first
	v.currentImage.Resource = nil
	v.currentImage.Image = img
	v.currentImage.FillMode = canvas.ImageFillContain
	
	// Force refresh of the image and scroll container
	v.currentImage.Refresh()
	v.imageScroll.Refresh()
	
	// Scroll to top
	v.imageScroll.ScrollToTop()
	
	return nil
}

// ToggleBookmarks shows or hides the bookmark panel
func (v *PDFViewer) ToggleBookmarks() {
	v.showBookmarks = !v.showBookmarks
	
	if v.showBookmarks {
		// Show bookmarks - switch to split view
		v.mainContent.Objects = []fyne.CanvasObject{v.split}
	} else {
		// Hide bookmarks - show only PDF viewer
		pdfContent := v.split.Trailing
		v.mainContent.Objects = []fyne.CanvasObject{pdfContent}
	}
	
	v.mainContent.Refresh()
}

// ShowBookmarks shows the bookmark panel
func (v *PDFViewer) ShowBookmarks() {
	if !v.showBookmarks {
		v.ToggleBookmarks()
	}
}

// HideBookmarks hides the bookmark panel
func (v *PDFViewer) HideBookmarks() {
	if v.showBookmarks {
		v.ToggleBookmarks()
	}
}

// GetScrollYOffset returns the current vertical scroll position
func (v *PDFViewer) GetScrollYOffset() float32 {
	return v.imageScroll.Offset.Y
}

// SetScrollYOffset sets the vertical scroll position
func (v *PDFViewer) SetScrollYOffset(yOffset float32) {
	v.imageScroll.Offset.Y = yOffset
	v.imageScroll.Refresh()
}

