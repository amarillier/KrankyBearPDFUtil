package ui

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Side-panel display modes, persisted across launches.
const (
	prefPanelMode  = "pdfviewer.panel"
	PanelNone      = "none"      // panel hidden
	PanelTOC       = "toc"       // show the PDF's own outline (table of contents)
	PanelBookmarks = "bookmarks" // show bookmarks added in this app
)

// Other persisted preferences.
const (
	prefContinuous  = "pdfviewer.continuous"
	prefRecentFiles = "pdfviewer.recent"
	prefZoom        = "pdfviewer.zoom"
	maxRecentFiles  = 10

	zoomFitWidth = "Fit Width"
	zoomFitPage  = "Fit Page"
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
	viewport      *fyne.Container

	// UI components
	pageEntry       *widget.Entry
	pageLabel       *widget.Label
	zoomSelect      *widget.Select
	continuousCheck *widget.Check
	statusBar       *widget.Label
	toolbar         *fyne.Container
	bookmarkPanel   *BookmarkPanel
	split           *container.Split
	mainContent     *fyne.Container

	// Continuous (lazy) scroll mode
	continuous bool
	pagesBox   *fyne.Container
	pagesL     *pagesLayout
	pageImages []*canvas.Image // one slot per page; image is nil until rendered into view
	contRowW   float32         // per-page display width
	contRowH   float32         // per-page display height (uniform, from page 1 aspect)
	contAspect float32         // page height / width (from page 1)

	// State management
	mu        sync.RWMutex
	lastDir   string
	panelMode string // PanelNone / PanelTOC / PanelBookmarks
}

// (fitWatcher is defined below)

// pagesLayout stacks page slots vertically with no gaps and uniform height, so scroll-offset
// math maps cleanly to page index. Slots are centred when the viewport is wider than a page.
type pagesLayout struct{ rowW, rowH float32 }

func (l *pagesLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(l.rowW, l.rowH*float32(len(objs)))
}

func (l *pagesLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	x := (size.Width - l.rowW) / 2
	if x < 0 {
		x = 0
	}
	for i, o := range objs {
		o.Resize(fyne.NewSize(l.rowW, l.rowH))
		o.Move(fyne.NewPos(x, float32(i)*l.rowH))
	}
}

// fitWatcher is a passthrough layout that fills its single child (the scroll) and reports
// the viewport size whenever it changes — used to re-fit the page live on window resize
// (Fit Width needs width; Fit Page needs width and height).
type fitWatcher struct {
	onResize func(fyne.Size)
	lastSize fyne.Size
}

func (f *fitWatcher) MinSize(objs []fyne.CanvasObject) fyne.Size {
	if len(objs) > 0 {
		return objs[0].MinSize()
	}
	return fyne.NewSize(0, 0)
}

func (f *fitWatcher) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objs {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
	if f.onResize != nil && size != f.lastSize {
		f.lastSize = size
		f.onResize(size)
	}
}

// NewPDFViewer creates a new PDF viewer instance
func NewPDFViewer(w fyne.Window, a fyne.App) *PDFViewer {
	v := &PDFViewer{
		window:      w,
		app:         a,
		currentPage: 1,
		zoomLevel:   1.0,
		panelMode:   a.Preferences().StringWithFallback(prefPanelMode, PanelNone),
		continuous:  a.Preferences().BoolWithFallback(prefContinuous, false),
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
	// In continuous mode, (un)render pages as they scroll into/out of view.
	v.imageScroll.OnScrolled = func(_ fyne.Position) {
		if v.continuous {
			v.lazyRenderVisible()
			v.updateCurrentPageFromScroll()
		}
	}

	// Wrap the scroll in a width-watching layout so "Fit Width" re-fits live on resize.
	v.viewport = container.New(&fitWatcher{onResize: v.onViewportResize}, v.imageScroll)

	// Create status bar
	v.statusBar = widget.NewLabel("Open a PDF to begin viewing")
	v.statusBar.Alignment = fyne.TextAlignCenter

	// Create bookmark panel
	v.bookmarkPanel = NewBookmarkPanel(v)

	// Setup tree selection callback
	v.bookmarkPanel.tree.OnSelected = func(uid string) {
		v.bookmarkPanel.OnBookmarkSelected(uid)
	}

	// PDF viewer content
	pdfContent := container.NewBorder(
		v.toolbar,    // top
		v.statusBar,  // bottom
		nil,          // left
		nil,          // right
		v.viewport,   // center
	)
	
	// Create split view with bookmarks (initially hidden)
	v.split = container.NewHSplit(v.bookmarkPanel.GetContainer(), pdfContent)
	v.split.SetOffset(0.2) // Bookmarks take 20% of width
	
	// Main content; the persisted panel mode is applied below.
	v.mainContent = container.NewMax(pdfContent)

	v.setupShortcuts()
	v.applyPanelMode() // restore last-used panel state (TOC / Bookmarks / hidden)

	return v.mainContent
}

// setupShortcuts wires keyboard navigation: PgDn/Right = next page, PgUp/Left = previous,
// Home/End = first/last. While the page-number entry is focused, only PgUp/PgDn act so the
// arrow keys still move the text cursor.
func (v *PDFViewer) setupShortcuts() {
	if v.window == nil {
		return
	}
	c := v.window.Canvas()
	c.SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if _, editing := c.Focused().(*widget.Entry); editing {
			switch ev.Name {
			case fyne.KeyPageUp, fyne.KeyPageDown: // allow page flips while typing
			default:
				return
			}
		}
		switch ev.Name {
		case fyne.KeyPageDown, fyne.KeyRight:
			v.NextPage()
		case fyne.KeyPageUp, fyne.KeyLeft:
			v.PreviousPage()
		case fyne.KeyHome:
			v.JumpToPage(1)
		case fyne.KeyEnd:
			v.JumpToPage(v.totalPages)
		}
	})

	// Cmd/Ctrl+D = add bookmark, Cmd/Ctrl+T = add TOC entry (KeyModifierShortcutDefault is
	// Cmd on macOS, Ctrl elsewhere — browser-like).
	c.AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyD, Modifier: fyne.KeyModifierShortcutDefault},
		func(fyne.Shortcut) {
			if v.bookmarkPanel != nil {
				v.bookmarkPanel.showAddDialog(false)
			}
		},
	)
	c.AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierShortcutDefault},
		func(fyne.Shortcut) {
			if v.bookmarkPanel != nil {
				v.bookmarkPanel.showAddDialog(true)
			}
		},
	)
}

// ---- Continuous (lazy) scroll mode ----

// contZoom is the render zoom for continuous mode (Fit Width/Page render at 100% and scale).
func (v *PDFViewer) contZoom() float32 {
	if v.zoomLevel <= 0 {
		return 1.0
	}
	return v.zoomLevel
}

// computeRowSize sets the per-page display size from page 1's aspect and the current zoom.
func (v *PDFViewer) computeRowSize(imgW1, imgH1 float32) {
	if imgW1 < 1 || imgH1 < 1 {
		return
	}
	v.contAspect = imgH1 / imgW1
	if v.zoomLevel <= 0 { // Fit Width / Fit Page (continuous fits to width)
		vpW := v.imageScroll.Size().Width
		if vpW < 1 {
			vpW = 600
		}
		v.contRowW = vpW - 4
	} else {
		const displayScale = 72.0 / baseRenderDPI
		v.contRowW = imgW1 * v.zoomLevel * displayScale
	}
	v.contRowH = v.contRowW * v.contAspect
}

// setContinuous switches between single-page and continuous (scroll-all-pages) layout.
func (v *PDFViewer) setContinuous(on bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.continuous = on
	if v.app != nil {
		v.app.Preferences().SetBool(prefContinuous, on)
	}
	if v.pdfPath == "" {
		return
	}
	if on {
		if err := v.buildContinuousPages(); err != nil {
			v.continuous = false
			if v.continuousCheck != nil {
				v.continuousCheck.SetChecked(false)
			}
			return
		}
		v.scrollToPageContinuous(v.currentPage)
	} else {
		v.imageScroll.Content = v.currentImage
		v.imageScroll.Refresh()
		v.renderCurrentPage()
	}
}

// buildContinuousPages creates one slot per page (images rendered lazily) and shows it.
func (v *PDFViewer) buildContinuousPages() error {
	if v.totalPages < 1 {
		return fmt.Errorf("no pages")
	}
	img, err := v.renderer.RenderPage(1, 1.0)
	if err != nil {
		return err
	}
	b := img.Bounds()
	v.computeRowSize(float32(b.Dx()), float32(b.Dy()))

	v.pageImages = make([]*canvas.Image, v.totalPages)
	objs := make([]fyne.CanvasObject, v.totalPages)
	for i := 0; i < v.totalPages; i++ {
		im := canvas.NewImageFromImage(nil)
		im.FillMode = canvas.ImageFillContain
		v.pageImages[i] = im
		objs[i] = im
	}
	v.pagesL = &pagesLayout{rowW: v.contRowW, rowH: v.contRowH}
	v.pagesBox = container.New(v.pagesL, objs...)
	v.imageScroll.Content = v.pagesBox
	v.imageScroll.Refresh()
	v.lazyRenderVisible()
	return nil
}

// relayoutContinuous recomputes slot sizes (after zoom or resize) and re-renders.
func (v *PDFViewer) relayoutContinuous() {
	if !v.continuous || v.pagesL == nil {
		return
	}
	// Recompute from page-1 aspect we already stored (no re-render needed for sizing).
	if v.zoomLevel <= 0 { // Fit Width / Fit Page (continuous fits to width)
		vpW := v.imageScroll.Size().Width
		if vpW < 1 {
			vpW = 600
		}
		v.contRowW = vpW - 4
	} else {
		const displayScale = 72.0 / baseRenderDPI
		// derive page-1 width from the stored aspect via a fresh render-free estimate:
		// reuse current rendered slot if available, else re-render page 1.
		if img, err := v.renderer.RenderPage(1, 1.0); err == nil {
			b := img.Bounds()
			v.contRowW = float32(b.Dx()) * v.zoomLevel * displayScale
		}
	}
	v.contRowH = v.contRowW * v.contAspect
	v.pagesL.rowW = v.contRowW
	v.pagesL.rowH = v.contRowH
	v.pagesBox.Refresh()
	v.lazyRenderVisible()
}

// scrollToPageContinuous scrolls so page p (1-based) is at the top.
func (v *PDFViewer) scrollToPageContinuous(p int) {
	if v.contRowH <= 0 || p < 1 {
		return
	}
	v.imageScroll.Offset.Y = float32(p-1) * v.contRowH
	v.imageScroll.Refresh()
	v.lazyRenderVisible()
}

// lazyRenderVisible renders pages within (and just around) the viewport, freeing the rest.
func (v *PDFViewer) lazyRenderVisible() {
	if !v.continuous || len(v.pageImages) == 0 || v.contRowH <= 0 {
		return
	}
	off := v.imageScroll.Offset.Y
	vh := v.imageScroll.Size().Height
	if vh <= 0 {
		vh = 600
	}
	const buffer = 1
	lo := int(off/v.contRowH) - buffer
	hi := int((off+vh)/v.contRowH) + buffer
	if lo < 0 {
		lo = 0
	}
	if hi > len(v.pageImages)-1 {
		hi = len(v.pageImages) - 1
	}
	for i, im := range v.pageImages {
		if i >= lo && i <= hi {
			if im.Image == nil {
				if rendered, err := v.renderer.RenderPage(i+1, v.contZoom()); err == nil {
					im.Image = rendered
					im.Refresh()
				}
			}
		} else if im.Image != nil {
			im.Image = nil // free off-screen page to cap memory
			im.Refresh()
		}
	}
}

// updateCurrentPageFromScroll keeps the page indicator in sync while scrolling.
func (v *PDFViewer) updateCurrentPageFromScroll() {
	if v.contRowH <= 0 {
		return
	}
	off := v.imageScroll.Offset.Y
	vh := v.imageScroll.Size().Height
	p := int((off+vh*0.3)/v.contRowH) + 1
	if p < 1 {
		p = 1
	}
	if p > v.totalPages {
		p = v.totalPages
	}
	if p != v.currentPage {
		v.currentPage = p
		v.updatePageDisplay()
	}
}

// ---- Recent files (persisted, most-recent first, capped at maxRecentFiles) ----

// RecentFiles returns the saved recent-file paths (most recent first).
func (v *PDFViewer) RecentFiles() []string {
	if v.app == nil {
		return nil
	}
	return v.app.Preferences().StringList(prefRecentFiles)
}

// ClearRecent empties the recent-files list.
func (v *PDFViewer) ClearRecent() {
	if v.app != nil {
		v.app.Preferences().SetStringList(prefRecentFiles, []string{})
	}
}

// addRecent records path at the front of the recent list (de-duplicated, capped).
func (v *PDFViewer) addRecent(path string) {
	if v.app == nil || path == "" {
		return
	}
	out := []string{path}
	for _, p := range v.app.Preferences().StringList(prefRecentFiles) {
		if p == path {
			continue
		}
		out = append(out, p)
		if len(out) >= maxRecentFiles {
			break
		}
	}
	v.app.Preferences().SetStringList(prefRecentFiles, out)
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
	zoomOptions := []string{
		zoomFitWidth,
		zoomFitPage,
		"50%",
		"75%",
		"100%",
		"125%",
		"150%",
		"200%",
		"300%",
	}
	v.zoomSelect = widget.NewSelect(zoomOptions, func(s string) {
		v.handleZoomChange(s)
	})
	// Restore the last-used zoom (validate it's still a known option).
	saved := v.app.Preferences().StringWithFallback(prefZoom, zoomFitWidth)
	valid := false
	for _, o := range zoomOptions {
		if o == saved {
			valid = true
			break
		}
	}
	if !valid {
		saved = zoomFitWidth
	}
	v.zoomSelect.SetSelected(saved)

	// Continuous (scroll all pages) toggle — reflect the persisted choice.
	v.continuousCheck = widget.NewCheck("Continuous", func(b bool) {
		v.setContinuous(b)
	})
	v.continuousCheck.SetChecked(v.continuous)

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
		v.continuousCheck,
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
	v.addRecent(path)
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
	
	// Render: continuous rebuilds the page stack; single renders the first page.
	if v.continuous {
		return v.buildContinuousPages()
	}
	return v.renderCurrentPage()
}

// NextPage advances to the next page
func (v *PDFViewer) NextPage() {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.currentPage < v.totalPages {
		v.currentPage++
		v.updatePageDisplay()
		if v.continuous {
			v.scrollToPageContinuous(v.currentPage)
		} else if err := v.renderCurrentPage(); err != nil {
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
		if v.continuous {
			v.scrollToPageContinuous(v.currentPage)
		} else if err := v.renderCurrentPage(); err != nil {
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

// JumpToPageAndPosition jumps to a specific page, then scrolls to frac (0..1) of the page's
// scrollable height. Using a fraction (rather than an absolute pixel offset) keeps region
// bookmarks accurate across zoom levels and window sizes.
func (v *PDFViewer) JumpToPageAndPosition(page int, frac float32) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if page < 1 || page > v.totalPages {
		return
	}

	v.currentPage = page
	v.updatePageDisplay()

	// Continuous: scroll to the page (plus the region fraction within it).
	if v.continuous {
		y := float32(page-1) * v.contRowH
		if frac > 0 {
			y += frac * v.contRowH
		}
		v.imageScroll.Offset.Y = y
		v.imageScroll.Refresh()
		v.lazyRenderVisible()
		return
	}

	if err := v.renderCurrentPage(); err != nil {
		if v.statusBar != nil {
			v.statusBar.SetText(fmt.Sprintf("Error: %v", err))
		}
		dialog.ShowError(err, v.window)
		return
	}

	// Restore the scroll position after rendering (renderCurrentPage scrolls to top first).
	if frac > 0 {
		if r := v.scrollableHeight(); r > 0 {
			v.imageScroll.Offset.Y = frac * r
			v.imageScroll.Refresh()
		}
	}
}

// scrollableHeight is how far the page can scroll vertically in the current view (page
// height minus the viewport height); 0 when the page fits entirely.
func (v *PDFViewer) scrollableHeight() float32 {
	r := v.currentImage.MinSize().Height - v.imageScroll.Size().Height
	if r < 1 {
		return 0
	}
	return r
}

// GetScrollFraction returns the current vertical scroll position as a fraction (0..1) of the
// scrollable height — used to capture a zoom-independent region-bookmark location.
func (v *PDFViewer) GetScrollFraction() float32 {
	r := v.scrollableHeight()
	if r <= 0 {
		return 0
	}
	f := v.imageScroll.Offset.Y / r
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	return f
}

// handleZoomChange handles zoom level changes
func (v *PDFViewer) handleZoomChange(zoom string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	switch zoom {
	case zoomFitWidth:
		v.zoomLevel = -1.0 // Special value for fit width
	case zoomFitPage:
		v.zoomLevel = -2.0 // Special value for fit page (whole page in view)
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

	// Remember the choice for next launch.
	if v.app != nil {
		v.app.Preferences().SetString(prefZoom, zoom)
	}

	// Only render if there's a PDF loaded
	if v.pdfPath == "" {
		return
	}

	// Continuous: resize the page slots for the new zoom and keep the current page in view.
	if v.continuous {
		page := v.currentPage
		v.relayoutContinuous()
		v.scrollToPageContinuous(page)
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
	
	// Fit Width (-1) / Fit Page (-2) render at 100% and are scaled to the viewport below.
	fitWidth := v.zoomLevel == -1.0
	fitPage := v.zoomLevel == -2.0
	effectiveZoom := v.zoomLevel
	if fitWidth || fitPage {
		effectiveZoom = 1.0
	}

	// Render the page
	img, err := v.renderer.RenderPage(v.currentPage, effectiveZoom)
	if err != nil {
		return fmt.Errorf("failed to render page: %w", err)
	}

	b := img.Bounds()
	imgW, imgH := float32(b.Dx()), float32(b.Dy())
	if imgW < 1 || imgH < 1 {
		return fmt.Errorf("rendered page has zero size")
	}

	// Clear any existing resource first, then set the new image
	v.currentImage.Resource = nil
	v.currentImage.Image = img
	v.currentImage.FillMode = canvas.ImageFillContain

	// The display size must scale with the zoom level. The image lives inside a scroll
	// container, so we drive the displayed size via MinSize (the scroll provides bars when
	// the page is larger than the viewport). A fixed MinSize was the reason zoom did
	// nothing and the page always appeared small.
	switch {
	case fitWidth:
		// Scale the page to the width of the scroll viewport; scroll vertically.
		v.applyFitWidth(v.imageScroll.Size().Width, imgW, imgH)
	case fitPage:
		// Scale so the whole page fits within the viewport (width and height).
		v.applyFitPage(imgW, imgH)
	default:
		// Render DPI already scales with zoom, so a constant divisor keeps the page
		// crisp while honouring the selected zoom (100% ≈ natural 72-DPI page size).
		const displayScale = 72.0 / baseRenderDPI
		v.currentImage.SetMinSize(fyne.NewSize(imgW*displayScale, imgH*displayScale))
	}

	// Force a refresh, then reset to the top of the newly rendered page. Setting Offset
	// directly + Refresh is more reliable than ScrollToTop(), which didn't reset when
	// consecutive pages share the same dimensions (the scroll keeps its prior position).
	// JumpToPageAndPosition re-applies a region offset after this when needed.
	v.currentImage.Refresh()
	v.imageScroll.Offset = fyne.NewPos(0, 0)
	v.imageScroll.Refresh()

	return nil
}

// applyFitWidth sizes the page image to the given viewport width, preserving aspect ratio.
func (v *PDFViewer) applyFitWidth(vpW, imgW, imgH float32) {
	if imgW < 1 || imgH < 1 {
		return
	}
	if vpW < 1 {
		vpW = 600 // before the first layout pass
	}
	vpW -= 4 // small margin so the vertical scrollbar doesn't overlap content
	v.currentImage.SetMinSize(fyne.NewSize(vpW, vpW*imgH/imgW))
}

// applyFitPage sizes the page so the whole page fits within the viewport (both dimensions),
// preserving aspect ratio.
func (v *PDFViewer) applyFitPage(imgW, imgH float32) {
	if imgW < 1 || imgH < 1 {
		return
	}
	vpW := v.imageScroll.Size().Width
	vpH := v.imageScroll.Size().Height
	if vpW < 1 {
		vpW = 600
	}
	if vpH < 1 {
		vpH = 800
	}
	scale := (vpW - 4) / imgW
	if sh := (vpH - 4) / imgH; sh < scale {
		scale = sh
	}
	v.currentImage.SetMinSize(fyne.NewSize(imgW*scale, imgH*scale))
}

// onViewportResize is called when the viewport size changes (window resize). In Fit Width /
// Fit Page modes it re-fits the already-rendered page to the new size — no re-render needed,
// just a rescale, so it stays smooth.
func (v *PDFViewer) onViewportResize(_ fyne.Size) {
	// Continuous + Fit Width/Page: re-fit slots to the new width, preserving the proportional
	// scroll position (rendered images rescale via Contain, no re-render).
	if v.continuous {
		if v.zoomLevel <= 0 && v.pagesL != nil && len(v.pageImages) > 0 {
			oldTotal := v.contRowH * float32(len(v.pageImages))
			var frac float32
			if oldTotal > 0 {
				frac = v.imageScroll.Offset.Y / oldTotal
			}
			v.relayoutContinuous()
			newTotal := v.contRowH * float32(len(v.pageImages))
			v.imageScroll.Offset.Y = frac * newTotal
			v.imageScroll.Refresh()
			v.lazyRenderVisible()
			v.updateCurrentPageFromScroll()
		}
		return
	}

	if v.currentImage == nil || v.currentImage.Image == nil {
		return
	}
	b := v.currentImage.Image.Bounds()
	imgW, imgH := float32(b.Dx()), float32(b.Dy())
	switch v.zoomLevel {
	case -1.0: // Fit Width
		v.applyFitWidth(v.imageScroll.Size().Width, imgW, imgH)
	case -2.0: // Fit Page
		v.applyFitPage(imgW, imgH)
	default:
		return // fixed zoom: nothing to re-fit on resize
	}
	v.currentImage.Refresh()
	v.imageScroll.Refresh()
}

// PanelMode returns the current side-panel mode (PanelNone/PanelTOC/PanelBookmarks).
func (v *PDFViewer) PanelMode() string {
	return v.panelMode
}

// SetPanelMode switches the side panel between the document's Table of Contents, the user's
// Bookmarks, or hidden — and remembers the choice for next launch.
func (v *PDFViewer) SetPanelMode(mode string) {
	switch mode {
	case PanelTOC, PanelBookmarks, PanelNone:
	default:
		mode = PanelNone
	}
	v.panelMode = mode
	if v.app != nil {
		v.app.Preferences().SetString(prefPanelMode, mode)
	}
	v.applyPanelMode()
}

// ToggleBookmarks cycles Bookmarks → hidden (kept for the menu/back-compat).
func (v *PDFViewer) ToggleBookmarks() {
	if v.panelMode == PanelNone {
		v.SetPanelMode(PanelBookmarks)
	} else {
		v.SetPanelMode(PanelNone)
	}
}

// applyPanelMode shows/hides the side panel and refreshes its (filtered) contents.
func (v *PDFViewer) applyPanelMode() {
	if v.mainContent == nil || v.split == nil {
		return
	}
	if v.panelMode == PanelNone {
		v.mainContent.Objects = []fyne.CanvasObject{v.split.Trailing}
	} else {
		if v.bookmarkPanel != nil {
			// The visible list changes with the mode, so clear any stale selection.
			v.bookmarkPanel.selected = nil
			v.bookmarkPanel.tree.UnselectAll()
			v.bookmarkPanel.rebuildTreeData()
			v.bookmarkPanel.tree.Refresh()
			v.bookmarkPanel.syncModeSelect()
		}
		v.mainContent.Objects = []fyne.CanvasObject{v.split}
	}
	v.mainContent.Refresh()
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

