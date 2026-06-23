package ui

import (
	"fmt"
	"strconv"

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
}

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

	// Add bookmark button
	addBtn := widget.NewButton("Add Bookmark", func() {
		bp.showAddBookmarkDialog()
	})

	// Delete bookmark button
	deleteBtn := widget.NewButton("Delete Selected", func() {
		bp.deleteSelectedBookmark()
	})
	deleteBtn.Importance = widget.DangerImportance

	// Save bookmarks button
	saveBtn := widget.NewButton("Save to PDF", func() {
		bp.showSaveBookmarksDialog()
	})

	// Refresh button
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		bp.LoadBookmarks(bp.viewer.pdfPath)
	})
	refreshBtn.Importance = widget.LowImportance

	toolbar := container.NewBorder(
		nil, nil,
		nil, refreshBtn,
		container.NewVBox(addBtn, deleteBtn, saveBtn),
	)

	// Main container
	bp.container = container.NewBorder(
		toolbar,
		nil, nil, nil,
		container.NewScroll(bp.tree),
	)
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
		childID := fmt.Sprintf("%s-%d", uid, i)
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

	// Show region indicator if bookmark has Y offset
	if bookmark.YOffset > 0 {
		label.SetText(fmt.Sprintf("📍 %s (p.%d)", bookmark.Title, bookmark.PageNo))
	} else {
		label.SetText(fmt.Sprintf("%s (p.%d)", bookmark.Title, bookmark.PageNo))
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

	bookmarks := bp.bookmarkManager.GetBookmarks()
	for i, bookmark := range bookmarks {
		id := fmt.Sprintf("root-%d", i)
		bp.treeData[id] = bookmark
		bp.rootIDs = append(bp.rootIDs, id)
	}
}

// OnBookmarkSelected is called when a bookmark is selected
func (bp *BookmarkPanel) OnBookmarkSelected(uid string) {
	bookmark, exists := bp.treeData[uid]
	if !exists {
		return
	}

	// Jump to the bookmarked page and position
	bp.viewer.JumpToPageAndPosition(bookmark.PageNo, bookmark.YOffset)
}

// showAddBookmarkDialog shows dialog to add a new bookmark
func (bp *BookmarkPanel) showAddBookmarkDialog() {
	if bp.viewer.pdfPath == "" {
		dialog.ShowInformation("No PDF", "Please open a PDF first", bp.viewer.window)
		return
	}

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Bookmark title")

	// Get current scroll position
	currentYOffset := bp.viewer.GetScrollYOffset()
	
	// Radio buttons for bookmark type
	typeRadio := widget.NewRadioGroup([]string{
		"Page bookmark (jump to top of page)",
		"Region bookmark (jump to current position)",
	}, nil)
	
	// Default to region if scrolled down, otherwise page
	if currentYOffset > 10 {
		typeRadio.SetSelected("Region bookmark (jump to current position)")
	} else {
		typeRadio.SetSelected("Page bookmark (jump to top of page)")
	}

	pageEntry := widget.NewEntry()
	pageEntry.SetText(strconv.Itoa(bp.viewer.currentPage))
	pageEntry.SetPlaceHolder("Page number")

	form := container.NewVBox(
		widget.NewLabel("Add Bookmark"),
		widget.NewForm(
			widget.NewFormItem("Title:", titleEntry),
			widget.NewFormItem("Page:", pageEntry),
		),
		typeRadio,
	)

	d := dialog.NewCustomConfirm("Add Bookmark", "Add", "Cancel", form, func(ok bool) {
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

		// Determine Y offset based on selection
		var yOffset float32
		if typeRadio.Selected == "Region bookmark (jump to current position)" {
			yOffset = currentYOffset
		} else {
			yOffset = 0 // Page bookmark
		}

		bp.bookmarkManager.AddBookmark(title, pageNo, yOffset)
		bp.rebuildTreeData()
		bp.tree.Refresh()

		bookmarkType := "Bookmark"
		if yOffset > 0 {
			bookmarkType = "Region bookmark"
		}
		dialog.ShowInformation("Success", fmt.Sprintf("%s '%s' added", bookmarkType, title), bp.viewer.window)
	}, bp.viewer.window)

	d.Resize(fyne.NewSize(450, 250))
	d.Show()
}

// showSaveBookmarksDialog shows dialog to save bookmarks to PDF
func (bp *BookmarkPanel) showSaveBookmarksDialog() {
	if bp.viewer.pdfPath == "" {
		dialog.ShowInformation("No PDF", "Please open a PDF first", bp.viewer.window)
		return
	}

	if !bp.bookmarkManager.HasBookmarks() {
		dialog.ShowInformation("No Bookmarks", "Add some bookmarks first", bp.viewer.window)
		return
	}

	dialog.ShowConfirm("Save Bookmarks", 
		"This will save bookmarks to a new PDF file.\nOriginal file will not be modified.",
		func(ok bool) {
			if !ok {
				return
			}

			// Create output filename
			outputPath := bp.viewer.pdfPath + "-bookmarked.pdf"

			err := bp.bookmarkManager.SaveBookmarks(outputPath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to save: %v", err), bp.viewer.window)
				return
			}

			dialog.ShowInformation("Success", 
				fmt.Sprintf("Bookmarks saved to:\n%s", outputPath), 
				bp.viewer.window)
		}, bp.viewer.window)
}

// selectedBookmarkUID stores the last selected bookmark UID
var selectedBookmarkUID string

// deleteSelectedBookmark deletes the currently selected bookmark
func (bp *BookmarkPanel) deleteSelectedBookmark() {
	if selectedBookmarkUID == "" {
		dialog.ShowInformation("No Selection", "Please select a bookmark to delete", bp.viewer.window)
		return
	}

	bookmark, exists := bp.treeData[selectedBookmarkUID]
	if !exists {
		dialog.ShowError(fmt.Errorf("bookmark not found"), bp.viewer.window)
		return
	}

	// Confirm deletion
	dialog.ShowConfirm("Delete Bookmark", 
		fmt.Sprintf("Delete bookmark '%s'?", bookmark.Title),
		func(ok bool) {
			if !ok {
				return
			}

			if bp.bookmarkManager.DeleteBookmark(bookmark) {
				bp.rebuildTreeData()
				bp.tree.Refresh()
				selectedBookmarkUID = ""
				dialog.ShowInformation("Success", "Bookmark deleted", bp.viewer.window)
			} else {
				dialog.ShowError(fmt.Errorf("failed to delete bookmark"), bp.viewer.window)
			}
		}, bp.viewer.window)
}

// GetContainer returns the bookmark panel container
func (bp *BookmarkPanel) GetContainer() *fyne.Container {
	return bp.container
}
