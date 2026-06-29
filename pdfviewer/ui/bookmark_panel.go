package ui

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BookmarkPanel manages the bookmark sidebar UI
type BookmarkPanel struct {
	viewer          *PDFViewer
	container       *fyne.Container
	tree            *widget.Tree
	bookmarkManager *BookmarkManager
	treeData        map[string]*Bookmark
	rootIDs         []string
	modeSelect      *widget.Select
	addBtn          *widget.Button
	selected        *Bookmark // currently selected entry (stable across tree rebuilds)
}

// Labels for the in-panel mode switch.
const (
	modeLabelTOC       = "📖 Table of Contents"
	modeLabelBookmarks = "🔖 Bookmarks"
)

// NewBookmarkPanel creates a new bookmark panel
func NewBookmarkPanel(viewer *PDFViewer) *BookmarkPanel {
	panel := &BookmarkPanel{
		viewer:          viewer,
		bookmarkManager: NewBookmarkManager(),
		treeData:        make(map[string]*Bookmark),
		rootIDs:         make([]string, 0),
	}

	panel.buildUI()
	return panel
}

// buildUI constructs the bookmark panel UI
func (bp *BookmarkPanel) buildUI() {
	// Create tree widget for bookmarks
	bp.tree = widget.NewTree(
		bp.childUIDs,
		bp.isBranch,
		bp.createTemplate,
		bp.updateItem,
	)

	// Add button — label switches with the mode (Bookmark vs TOC entry).
	bp.addBtn = widget.NewButton("Add Bookmark", func() {
		bp.showAddDialog(bp.viewer.panelMode == PanelTOC)
	})

	// Delete-selected (orange/warning) vs Delete-all (red/danger) to signal severity.
	deleteBtn := widget.NewButton("Delete Selected", func() {
		bp.deleteSelectedBookmark()
	})
	deleteBtn.Importance = widget.WarningImportance

	// Delete-all button — clears every entry in the current view (TOC or Bookmarks).
	deleteAllBtn := widget.NewButton("Delete All", func() {
		bp.deleteAll()
	})
	deleteAllBtn.Importance = widget.DangerImportance

	// Save bookmarks button
	saveBtn := widget.NewButton("Save to PDF", func() {
		bp.showSaveBookmarksDialog()
	})

	// Refresh button
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		bp.LoadBookmarks(bp.viewer.pdfPath)
	})
	refreshBtn.Importance = widget.LowImportance

	// In-panel switch between the document's Table of Contents and the user's Bookmarks.
	// Shows the current mode and lets the user flip without using the menu.
	bp.modeSelect = widget.NewSelect([]string{modeLabelTOC, modeLabelBookmarks}, func(s string) {
		if s == modeLabelTOC {
			bp.viewer.SetPanelMode(PanelTOC)
		} else {
			bp.viewer.SetPanelMode(PanelBookmarks)
		}
	})

	toolbar := container.NewBorder(
		nil, nil,
		nil, refreshBtn,
		container.NewVBox(bp.addBtn, deleteBtn, deleteAllBtn, saveBtn),
	)

	// Main container: mode switch + toolbar on top, scrollable tree below.
	bp.container = container.NewBorder(
		container.NewVBox(bp.modeSelect, toolbar),
		nil, nil, nil,
		container.NewScroll(bp.tree),
	)
}

// syncModeSelect updates the in-panel switch to reflect the current mode without firing its
// OnChanged handler (setting .Selected directly + Refresh, rather than SetSelected).
func (bp *BookmarkPanel) syncModeSelect() {
	if bp.modeSelect == nil {
		return
	}
	switch bp.viewer.panelMode {
	case PanelTOC:
		bp.modeSelect.Selected = modeLabelTOC
	case PanelBookmarks:
		bp.modeSelect.Selected = modeLabelBookmarks
	}
	bp.modeSelect.Refresh()
	bp.updateAddButtonLabel()
}

// updateAddButtonLabel makes the Add button match the active view: it adds a TOC entry in
// Table-of-Contents mode, or a bookmark otherwise.
func (bp *BookmarkPanel) updateAddButtonLabel() {
	if bp.addBtn == nil {
		return
	}
	if bp.viewer.panelMode == PanelTOC {
		bp.addBtn.SetText("Add TOC Entry")
	} else {
		bp.addBtn.SetText("Add Bookmark")
	}
}

// Tree functions for Fyne tree widget
func (bp *BookmarkPanel) childUIDs(uid string) []string {
	if uid == "" {
		return bp.rootIDs
	}

	bookmark, exists := bp.treeData[uid]
	if !exists || len(bookmark.Children) == 0 {
		return []string{}
	}

	childIDs := make([]string, len(bookmark.Children))
	for i, child := range bookmark.Children {
		childID := fmt.Sprintf("%p", child) // stable per-entry id (pointer identity)
		bp.treeData[childID] = child
		childIDs[i] = childID
	}

	return childIDs
}

func (bp *BookmarkPanel) isBranch(uid string) bool {
	if uid == "" {
		return true
	}

	bookmark, exists := bp.treeData[uid]
	return exists && len(bookmark.Children) > 0
}

func (bp *BookmarkPanel) createTemplate(branch bool) fyne.CanvasObject {
	return widget.NewLabel("Template")
}

func (bp *BookmarkPanel) updateItem(uid string, branch bool, item fyne.CanvasObject) {
	label := item.(*widget.Label)
	bookmark, exists := bp.treeData[uid]

	if !exists {
		label.SetText("Unknown")
		return
	}

	// Distinct icon per entry type:
	//   📖 document outline (table of contents)
	//   📍 user region bookmark (a specific position on the page)
	//   🔖 user page bookmark (top of page)
	switch {
	case !bookmark.IsUserAdded:
		label.SetText(fmt.Sprintf("📖 %s (p.%d)", bookmark.Title, bookmark.PageNo))
	case bookmark.YOffset > 0:
		label.SetText(fmt.Sprintf("📍 %s (p.%d)", bookmark.Title, bookmark.PageNo))
	default:
		label.SetText(fmt.Sprintf("🔖 %s (p.%d)", bookmark.Title, bookmark.PageNo))
	}
}

// LoadBookmarks loads bookmarks from the current PDF
func (bp *BookmarkPanel) LoadBookmarks(pdfPath string) error {
	if pdfPath == "" {
		return fmt.Errorf("no PDF loaded")
	}

	err := bp.bookmarkManager.LoadBookmarks(pdfPath)
	if err != nil {
		return err
	}

	// Rebuild tree data
	bp.rebuildTreeData()
	bp.tree.Refresh()

	return nil
}

// rebuildTreeData reconstructs the tree data map from bookmarks
func (bp *BookmarkPanel) rebuildTreeData() {
	bp.treeData = make(map[string]*Bookmark)
	bp.rootIDs = make([]string, 0)

	// Display bookmarks in page order (stable, so same-page bookmarks keep add-order).
	// This matches the order they're written on Save to PDF.
	bookmarks := bp.bookmarkManager.GetBookmarks()
	sorted := make([]*Bookmark, len(bookmarks))
	copy(sorted, bookmarks)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].PageNo < sorted[j].PageNo })

	// Filter by the current panel mode: TOC shows the document outline, Bookmarks shows
	// user-added entries. (When the panel is hidden the filter is irrelevant.)
	mode := bp.viewer.panelMode
	for _, bookmark := range sorted {
		switch mode {
		case PanelTOC:
			if bookmark.IsUserAdded {
				continue
			}
		case PanelBookmarks:
			if !bookmark.IsUserAdded {
				continue
			}
		}
		id := fmt.Sprintf("%p", bookmark) // stable per-entry id (pointer identity)
		bp.treeData[id] = bookmark
		bp.rootIDs = append(bp.rootIDs, id)
	}
}

// OnBookmarkSelected is called when an entry is selected: remember it (by pointer, so it
// stays valid across tree rebuilds) and jump to its page/position.
func (bp *BookmarkPanel) OnBookmarkSelected(uid string) {
	bookmark, exists := bp.treeData[uid]
	if !exists {
		return
	}
	bp.selected = bookmark
	bp.viewer.JumpToPageAndPosition(bookmark.PageNo, bookmark.YOffset)
}

// showAddDialog adds a new entry. toTOC=true adds a Table-of-Contents entry (page-level);
// otherwise a user Bookmark (with an optional region position).
func (bp *BookmarkPanel) showAddDialog(toTOC bool) {
	if bp.viewer.pdfPath == "" {
		dialog.ShowInformation("No PDF", "Please open a PDF first", bp.viewer.window)
		return
	}

	kind := "Bookmark"
	dlgTitle := "Add Bookmark"
	if toTOC {
		kind = "TOC entry"
		dlgTitle = "Add to Table of Contents"
	}

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder(kind + " title")

	pageEntry := widget.NewEntry()
	pageEntry.SetText(strconv.Itoa(bp.viewer.currentPage))
	pageEntry.SetPlaceHolder("Page number")

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Title:", titleEntry),
			widget.NewFormItem("Page:", pageEntry),
		),
	)

	// Region (current scroll position) is offered for bookmarks only; TOC entries are
	// page-level, like a normal document outline.
	currentFrac := bp.viewer.GetScrollFraction()
	var typeRadio *widget.RadioGroup
	if !toTOC {
		typeRadio = widget.NewRadioGroup([]string{
			"Page bookmark (jump to top of page)",
			"Region bookmark (jump to current position)",
		}, nil)
		if currentFrac > 0.01 {
			typeRadio.SetSelected("Region bookmark (jump to current position)")
		} else {
			typeRadio.SetSelected("Page bookmark (jump to top of page)")
		}
		form.Add(typeRadio)
	}

	d := dialog.NewCustomConfirm(dlgTitle, "Add", "Cancel", form, func(ok bool) {
		if !ok {
			return
		}

		title := titleEntry.Text
		if title == "" {
			dialog.ShowError(fmt.Errorf("title cannot be empty"), bp.viewer.window)
			return
		}

		pageNo, err := strconv.Atoi(pageEntry.Text)
		if err != nil || pageNo < 1 || pageNo > bp.viewer.totalPages {
			dialog.ShowError(fmt.Errorf("invalid page number"), bp.viewer.window)
			return
		}

		var yOffset float32
		if typeRadio != nil && typeRadio.Selected == "Region bookmark (jump to current position)" {
			yOffset = currentFrac
		}

		// userAdded=false for a TOC entry (📖), true for a bookmark (🔖/📍).
		bp.bookmarkManager.AddBookmark(title, pageNo, yOffset, !toTOC)

		// Show the new entry in its matching view.
		if toTOC {
			bp.viewer.SetPanelMode(PanelTOC)
		} else {
			bp.viewer.SetPanelMode(PanelBookmarks)
		}

		label := kind
		if !toTOC && yOffset > 0 {
			label = "Region bookmark"
		}
		bp.showTransientInfo("Success", fmt.Sprintf("%s '%s' added", label, title))
	}, bp.viewer.window)

	d.Resize(fyne.NewSize(460, 250))
	d.Show()
	// Put the cursor in the title field so the user can type immediately.
	bp.viewer.window.Canvas().Focus(titleEntry)
}

// showSaveBookmarksDialog shows dialog to save bookmarks to PDF
func (bp *BookmarkPanel) showSaveBookmarksDialog() {
	if bp.viewer.pdfPath == "" {
		dialog.ShowInformation("No PDF", "Please open a PDF first", bp.viewer.window)
		return
	}

	// Note: an empty list is allowed — saving then strips the outline (cleanup).
	empty := !bp.bookmarkManager.HasBookmarks()

	// Let the user choose between a new file or overwriting the original. pdfcpu writes to
	// a temp file and atomically renames when out == in, so overwriting is safe.
	const optNew = "Save as a new file (…-bookmarked.pdf)"
	const optOverwrite = "Overwrite the original file"
	choice := widget.NewRadioGroup([]string{optNew, optOverwrite}, nil)
	choice.SetSelected(optNew)

	headline := "Save bookmarks & TOC to:"
	if empty {
		headline = "No entries — this will REMOVE all bookmarks/TOC from the file. Save to:"
	}
	content := container.NewVBox(
		widget.NewLabel(headline),
		choice,
	)

	d := dialog.NewCustomConfirm("Save to PDF", "Save", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}

		overwrite := choice.Selected == optOverwrite
		outputPath := bp.viewer.pdfPath + "-bookmarked.pdf"
		if overwrite {
			outputPath = bp.viewer.pdfPath
		}

		if err := bp.bookmarkManager.SaveBookmarks(outputPath); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save: %v", err), bp.viewer.window)
			return
		}

		verb := "Saved to"
		if empty {
			verb = "Removed all entries; saved to"
		}
		if overwrite {
			bp.showTransientInfo("Success", fmt.Sprintf("%s the original file:\n%s", verb, outputPath))
		} else {
			bp.showTransientInfo("Success", fmt.Sprintf("%s:\n%s", verb, outputPath))
		}
	}, bp.viewer.window)

	d.Resize(fyne.NewSize(460, 220))
	d.Show()
}

// showTransientInfo shows a confirmation dialog that auto-dismisses after 3s, so the user
// needn't click OK (they can still click it to close sooner). fyne.Do marshals the Hide
// back onto the UI thread (Fyne >= 2.6).
func (bp *BookmarkPanel) showTransientInfo(title, msg string) {
	d := dialog.NewInformation(title, msg, bp.viewer.window)
	d.Show()
	time.AfterFunc(2*time.Second, func() {
		fyne.Do(d.Hide)
	})
}

// deleteSelectedBookmark deletes the currently selected entry.
func (bp *BookmarkPanel) deleteSelectedBookmark() {
	bookmark := bp.selected
	if bookmark == nil {
		dialog.ShowInformation("No Selection", "Please select an entry to delete", bp.viewer.window)
		return
	}

	dialog.ShowConfirm("Delete",
		fmt.Sprintf("Delete '%s'?", bookmark.Title),
		func(ok bool) {
			if !ok {
				return
			}

			if bp.bookmarkManager.DeleteBookmark(bookmark) {
				bp.selected = nil
				bp.tree.UnselectAll()
				bp.rebuildTreeData()
				bp.tree.Refresh()
				// Auto-dismissing confirmation (no click needed).
				bp.showTransientInfo("Deleted", fmt.Sprintf("Deleted '%s'", bookmark.Title))
			} else {
				dialog.ShowError(fmt.Errorf("failed to delete entry"), bp.viewer.window)
			}
		}, bp.viewer.window)
}

// deleteAll removes every entry in the current view (TOC entries in TOC mode, bookmarks
// otherwise), after a confirmation.
func (bp *BookmarkPanel) deleteAll() {
	userAdded := bp.viewer.panelMode != PanelTOC
	what := "bookmarks"
	if !userAdded {
		what = "table-of-contents entries"
	}

	dialog.ShowConfirm("Delete All", fmt.Sprintf("Delete ALL %s?", what), func(ok bool) {
		if !ok {
			return
		}
		n := bp.bookmarkManager.DeleteByType(userAdded)
		bp.selected = nil
		bp.tree.UnselectAll()
		bp.rebuildTreeData()
		bp.tree.Refresh()
		bp.showTransientInfo("Deleted", fmt.Sprintf("Deleted %d %s", n, what))
	}, bp.viewer.window)
}

// GetContainer returns the bookmark panel container
func (bp *BookmarkPanel) GetContainer() *fyne.Container {
	return bp.container
}
