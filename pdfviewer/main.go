package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pdfutil/pdfviewer/ui"
)

//go:embed assets/images/KrankyBearBeanieMultiColor.png
var appIconBytes []byte

// appIcon is the embedded window/system-tray icon.
var appIcon = fyne.NewStaticResource("KrankyBearBeanieMultiColor.png", appIconBytes)

const (
	appVersion = "0.3.1"
	appName    = "KrankyBear PDF Viewer"
	appID      = "com.krankybear.pdfviewer"
)

// ---- Window size persistence (size only; Fyne can't restore position) ----
const (
	prefWindowWidth  = "windowWidth"
	prefWindowHeight = "windowHeight"

	mainWindowDefaultWidth  = float32(1000)
	mainWindowDefaultHeight = float32(800)

	mainWindowMinWidth  = float32(500)
	mainWindowMinHeight = float32(400)
	mainWindowMaxWidth  = float32(8000)
	mainWindowMaxHeight = float32(8000)
)

// mainWindowLaunchSize returns the saved window size if present and sane, else the default.
func mainWindowLaunchSize(a fyne.App) fyne.Size {
	sw := float32(a.Preferences().FloatWithFallback(prefWindowWidth, float64(mainWindowDefaultWidth)))
	sh := float32(a.Preferences().FloatWithFallback(prefWindowHeight, float64(mainWindowDefaultHeight)))
	if sw < mainWindowMinWidth || sh < mainWindowMinHeight ||
		sw > mainWindowMaxWidth || sh > mainWindowMaxHeight {
		return fyne.NewSize(mainWindowDefaultWidth, mainWindowDefaultHeight)
	}
	return fyne.NewSize(sw, sh)
}

// saveMainWindowGeometry persists the current window size (skips a too-small/hidden size).
func saveMainWindowGeometry(a fyne.App, w fyne.Window) {
	if w == nil {
		return
	}
	sz := w.Canvas().Size()
	if sz.Width < mainWindowMinWidth || sz.Height < mainWindowMinHeight {
		return
	}
	a.Preferences().SetFloat(prefWindowWidth, float64(sz.Width))
	a.Preferences().SetFloat(prefWindowHeight, float64(sz.Height))
}

// quitViewer saves the window geometry before quitting.
func quitViewer(a fyne.App, w fyne.Window) {
	saveMainWindowGeometry(a, w)
	a.Quit()
}

func main() {
	a := app.NewWithID(appID)
	a.Settings().SetTheme(theme.DefaultTheme())

	w := a.NewWindow(appName)
	w.SetIcon(appIcon)
	w.Resize(mainWindowLaunchSize(a)) // restore last saved size (or default)
	w.CenterOnScreen()

	// With a system tray, the app keeps running after the window closes unless we say
	// otherwise. SetMaster + a close intercept ensure every close path fully quits (and
	// removes the tray icon), matching pdfgui. Save geometry on the way out.
	w.SetMaster()
	w.SetCloseIntercept(func() { quitViewer(a, w) })

	// Create the PDF viewer
	viewer := ui.NewPDFViewer(w, a)

	// Create menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open PDF...", func() {
			viewer.OpenPDFDialog()
		}),
		fyne.NewMenuItem("Open Recent...", func() {
			showRecentDialog(w, viewer)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			quitViewer(a, w)
		}),
	)

	// View menu: choose what the side panel shows (persisted across launches).
	// NOTE: no checkmarks/Menu.Refresh() here — calling Menu.Refresh() rebuilds the native
	// macOS menu and crashes in Fyne 2.5.4 (insertDarwinMenuItem on an empty array). The
	// active mode is reflected by the panel contents and remembered across launches.
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Table of Contents", func() { viewer.SetPanelMode(ui.PanelTOC) }),
		fyne.NewMenuItem("Bookmarks", func() { viewer.SetPanelMode(ui.PanelBookmarks) }),
		fyne.NewMenuItem("Hide Panel", func() { viewer.SetPanelMode(ui.PanelNone) }),
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

	// System tray (desktop) with icon + menu, mirroring pdfgui.
	if desk, ok := a.(desktop.App); ok {
		setupSystemTray(desk, a, w, viewer)
	}

	// Set content
	w.SetContent(viewer.Build())

	// Drag & drop a PDF onto the window to open it (desktop).
	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, u := range uris {
			if u == nil {
				continue
			}
			if strings.EqualFold(filepath.Ext(u.Path()), ".pdf") {
				if err := viewer.LoadPDF(u.Path()); err != nil {
					dialog.ShowError(fmt.Errorf("Failed to open PDF: %v", err), w)
				}
				return
			}
		}
	})

	// Handle command line arguments
	if len(os.Args) > 1 {
		pdfPath := os.Args[1]
		if err := viewer.LoadPDF(pdfPath); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to open PDF: %v", err), w)
		}
	}

	w.ShowAndRun()
}

// setupSystemTray adds a tray icon + menu (Show/Hide, Open, About/Help, theme, Quit).
func setupSystemTray(desk desktop.App, a fyne.App, w fyne.Window, viewer *ui.PDFViewer) {
	showItem := fyne.NewMenuItem("Show", func() { w.Show(); w.RequestFocus() })
	hideItem := fyne.NewMenuItem("Hide", func() { w.Hide() })
	openItem := fyne.NewMenuItem("Open PDF...", func() {
		w.Show()
		w.RequestFocus()
		viewer.OpenPDFDialog()
	})
	openRecentItem := fyne.NewMenuItem("Open Recent...", func() {
		w.Show()
		w.RequestFocus()
		showRecentDialog(w, viewer)
	})
	// View section — mirrors the main menu's View options.
	tocItem := fyne.NewMenuItem("Table of Contents", func() {
		w.Show()
		w.RequestFocus()
		viewer.SetPanelMode(ui.PanelTOC)
	})
	bookmarksItem := fyne.NewMenuItem("Bookmarks", func() {
		w.Show()
		w.RequestFocus()
		viewer.SetPanelMode(ui.PanelBookmarks)
	})
	hidePanelItem := fyne.NewMenuItem("Hide Panel", func() { viewer.SetPanelMode(ui.PanelNone) })

	aboutItem := fyne.NewMenuItem("About", func() { showAboutDialog(w) })
	helpItem := fyne.NewMenuItem("Help", func() { showHelpDialog(w) })
	lightItem := fyne.NewMenuItem("Light Theme", func() { a.Settings().SetTheme(theme.LightTheme()) })
	darkItem := fyne.NewMenuItem("Dark Theme", func() { a.Settings().SetTheme(theme.DarkTheme()) })
	systemItem := fyne.NewMenuItem("System Theme", func() { a.Settings().SetTheme(theme.DefaultTheme()) })
	quitItem := fyne.NewMenuItem("Quit", func() { quitViewer(a, w) })

	menu := fyne.NewMenu(appName,
		showItem,
		hideItem,
		fyne.NewMenuItemSeparator(),
		openItem,
		openRecentItem,
		fyne.NewMenuItemSeparator(),
		tocItem,
		bookmarksItem,
		hidePanelItem,
		fyne.NewMenuItemSeparator(),
		aboutItem,
		helpItem,
		fyne.NewMenuItemSeparator(),
		lightItem,
		darkItem,
		systemItem,
		fyne.NewMenuItemSeparator(),
		quitItem,
	)

	desk.SetSystemTrayMenu(menu)
	desk.SetSystemTrayIcon(appIcon)
}

// showRecentDialog lists recently opened PDFs (read fresh from preferences each time, so no
// native-menu rebuild is needed). Clicking one opens it; "Clear Recent List" empties it.
func showRecentDialog(w fyne.Window, viewer *ui.PDFViewer) {
	recent := viewer.RecentFiles()
	if len(recent) == 0 {
		dialog.ShowInformation("Recent Files", "No recent files yet.", w)
		return
	}

	var d dialog.Dialog
	list := container.NewVBox()
	for _, p := range recent {
		p := p // capture
		btn := widget.NewButton(filepath.Base(p)+"  —  "+filepath.Dir(p), func() {
			if d != nil {
				d.Hide()
			}
			if err := viewer.LoadPDF(p); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to open %s: %v", filepath.Base(p), err), w)
			}
		})
		btn.Alignment = widget.ButtonAlignLeading
		list.Add(btn)
	}

	clearBtn := widget.NewButton("Clear Recent List", func() {
		viewer.ClearRecent()
		if d != nil {
			d.Hide()
		}
	})
	clearBtn.Importance = widget.DangerImportance

	content := container.NewBorder(nil, clearBtn, nil, nil, container.NewVScroll(list))
	d = dialog.NewCustom("Recent Files", "Close", content, w)
	d.Resize(fyne.NewSize(640, 420))
	d.Show()
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
