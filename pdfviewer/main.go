package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pdfutil/pdfviewer/ui"
)

const (
	appVersion = "1.0.0"
	appName    = "KrankyBear PDF Viewer"
	appID      = "com.krankybear.pdfviewer"
)

func main() {
	a := app.NewWithID(appID)
	a.Settings().SetTheme(theme.DefaultTheme())

	w := a.NewWindow(appName)
	w.Resize(fyne.NewSize(1000, 800))
	w.CenterOnScreen()

	// Create the PDF viewer
	viewer := ui.NewPDFViewer(w, a)

	// Create menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open PDF...", func() {
			viewer.OpenPDFDialog()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			a.Quit()
		}),
	)

	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Toggle Bookmarks", func() {
			viewer.ToggleBookmarks()
		}),
	)

	settingsMenu := fyne.NewMenu("Settings",
		fyne.NewMenuItem("Light Theme", func() {
			a.Settings().SetTheme(theme.LightTheme())
		}),
		fyne.NewMenuItem("Dark Theme", func() {
			a.Settings().SetTheme(theme.DarkTheme())
		}),
		fyne.NewMenuItem("System Theme", func() {
			a.Settings().SetTheme(theme.DefaultTheme())
		}),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			showAboutDialog(w)
		}),
		fyne.NewMenuItem("Help", func() {
			showHelpDialog(w)
		}),
	)

	mainMenu := fyne.NewMainMenu(fileMenu, viewMenu, settingsMenu, helpMenu)
	w.SetMainMenu(mainMenu)

	// Set content
	w.SetContent(viewer.Build())

	// Handle command line arguments
	if len(os.Args) > 1 {
		pdfPath := os.Args[1]
		if err := viewer.LoadPDF(pdfPath); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to open PDF: %v", err), w)
		}
	}

	w.ShowAndRun()
}

func showAboutDialog(w fyne.Window) {
	about := widget.NewRichTextFromMarkdown(fmt.Sprintf(`# %s

**Version:** %s

A modern, feature-rich PDF viewer built with Go, Fyne, and MuPDF.

**Features:**
- View PDF documents (full rendering)
- Bookmark management (view, add, delete, save)
- Page navigation (keyboard & mouse)
- Zoom controls (50%% - 300%%)
- Jump to page
- Light/dark themes
- Cross-platform

**Powered by:**
- MuPDF (via go-fitz) - High-quality rendering
- pdfcpu - PDF manipulation & bookmarks
- Fyne - Modern Go GUI framework

---

© 2026 KrankyBear Software
100%% Free to Use

---

**Website:** https://krankybear.com  
**GitHub:** https://github.com/krankybear/pdfviewer
`, appName, appVersion))

	aboutDialog := dialog.NewCustom("About", "Close", container.NewScroll(about), w)
	aboutDialog.Resize(fyne.NewSize(500, 600))
	aboutDialog.Show()
}

func showHelpDialog(w fyne.Window) {
	help := widget.NewRichTextFromMarkdown(`# PDF Viewer Help

## Opening a PDF

**Method 1:** File → Open PDF...  
**Method 2:** Command line: ` + "`pdfviewer document.pdf`" + `

## Navigation

### Keyboard Shortcuts
- **Page Down** / **Space** - Next page
- **Page Up** / **Shift+Space** - Previous page
- **Home** - First page
- **End** - Last page
- **Ctrl/Cmd +** - Zoom in
- **Ctrl/Cmd -** - Zoom out
- **Ctrl/Cmd 0** - Fit to width

### Mouse
- **Scroll wheel** - Navigate pages
- **Click buttons** - Use toolbar controls

## Bookmarks

### Viewing Bookmarks
- **View → Toggle Bookmarks** - Show/hide bookmark panel
- **Click a bookmark** - Jump to that page (or region)
- PDFs with existing bookmarks will load automatically
- 📍 icon indicates a **region bookmark** (position on page)

### Adding Bookmarks
1. Navigate to the page (and scroll position) you want to bookmark
2. Click **"Add Bookmark"** in the bookmark panel
3. Enter a title
4. Choose bookmark type:
   - **Page bookmark** - Jumps to top of page
   - **Region bookmark** - Jumps to exact scroll position
5. Confirm to add

**Tip:** Region bookmarks are perfect for marking specific paragraphs or sections!

### Managing Bookmarks
- **Delete Selected** - Remove the selected bookmark
- **Save to PDF** - Save bookmarks to a new PDF file
  - Original file is never modified
  - Creates new file with "-bookmarked.pdf" suffix
  - Region positions are saved in memory only

## Toolbar Controls

### Navigation
- **◄ First** - Jump to first page
- **◄ Prev** - Previous page
- **► Next** - Next page
- **Last ►** - Jump to last page

### Page Input
Type a page number and press Enter to jump directly to that page.

### Zoom
- **Fit Width** - Scale to fit window width
- **100%** - Actual size
- **150%** - 1.5x zoom
- **200%** - 2x zoom
- **Custom** - Use +/- buttons

---

## Tips & Tricks

### Efficient Navigation
- Use **Space/Shift+Space** for quick page flipping
- Type page numbers for direct jumps
- Use **Home/End** for document boundaries
- Use **bookmarks** for quick access to chapters/sections

### Zoom Levels
- **Fit Width** is best for reading
- **100%** for accurate size
- Higher zoom for detailed viewing

### Bookmarks
- Add bookmarks to frequently accessed pages
- Organize long documents with descriptive bookmarks
- Export bookmarks by saving to a new PDF

### Performance
- Large PDFs may take a moment to load
- Pages are rendered on-demand with MuPDF
- Zoom changes re-render the current page
- Bookmarks are loaded automatically

---

## Keyboard Reference

| Key | Action |
|-----|--------|
| PgDn, Space | Next page |
| PgUp, Shift+Space | Previous page |
| Home | First page |
| End | Last page |
| Ctrl/Cmd + | Zoom in |
| Ctrl/Cmd - | Zoom out |
| Ctrl/Cmd 0 | Fit to width |
| Ctrl/Cmd O | Open PDF |
| Ctrl/Cmd Q | Quit |

---

**Enjoy viewing! 🐻📄**
`)

	helpDialog := dialog.NewCustom("Help", "Close", container.NewScroll(help), w)
	helpDialog.Resize(fyne.NewSize(700, 700))
	helpDialog.Show()
}

