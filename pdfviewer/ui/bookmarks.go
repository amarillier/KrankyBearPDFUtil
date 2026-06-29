package ui

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// bookmarkTitlePrefix marks an outline entry as a user "bookmark" (vs the document's own
// table-of-contents) when written into the PDF. A single BMP glyph keeps it portable: pdfcpu
// stores titles as UTF-16BE so it round-trips cleanly, and it renders in other readers too.
// On load, entries with this prefix are shown as 🔖 bookmarks (the prefix is stripped for
// display); entries without it are 📖 TOC. Change this one constant to use a different marker.
const bookmarkTitlePrefix = "⚑ "

// Bookmark represents a PDF bookmark/outline entry
type Bookmark struct {
	Title       string
	PageNo      int
	YOffset     float32 // Vertical scroll position as a fraction (0..1) for region bookmarks
	Children    []*Bookmark
	Level       int
	Bold        bool
	Italic      bool
	IsUserAdded bool // true = added in this app (a "bookmark"); false = the PDF's own outline (TOC)
}

// BookmarkManager handles PDF bookmark operations
type BookmarkManager struct {
	pdfPath   string
	bookmarks []*Bookmark
}

// NewBookmarkManager creates a new bookmark manager
func NewBookmarkManager() *BookmarkManager {
	return &BookmarkManager{
		bookmarks: make([]*Bookmark, 0),
	}
}

// LoadBookmarks loads bookmarks from a PDF file
func (bm *BookmarkManager) LoadBookmarks(pdfPath string) error {
	bm.pdfPath = pdfPath
	bm.bookmarks = make([]*Bookmark, 0)

	conf := model.NewDefaultConfiguration()
	
	// Open PDF file for reading
	f, err := os.Open(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	// Get bookmarks using API
	bookmarkList, err := api.Bookmarks(f, conf)
	if err != nil {
		log.Printf("[WARN] No bookmarks found or error reading bookmarks: %v", err)
		return nil // Not an error if PDF has no bookmarks
	}

	// Convert pdfcpu bookmarks to our format
	bm.bookmarks = convertPdfcpuBookmarks(bookmarkList)
	log.Printf("[INFO] Loaded %d top-level bookmarks", len(bm.bookmarks))
	
	return nil
}

// convertPdfcpuBookmarks converts pdfcpu bookmark format to our bookmark structure
func convertPdfcpuBookmarks(pdfcpuBookmarks []pdfcpu.Bookmark) []*Bookmark {
	return convertPdfcpuBookmarksRecursive(pdfcpuBookmarks, 1)
}

func convertPdfcpuBookmarksRecursive(pdfcpuBookmarks []pdfcpu.Bookmark, level int) []*Bookmark {
	bookmarks := make([]*Bookmark, 0)
	
	for _, pb := range pdfcpuBookmarks {
		// A title carrying the bookmark prefix was added in this app as a bookmark; strip the
		// marker for display and flag it. Everything else is the document's TOC.
		title := pb.Title
		userAdded := false
		if strings.HasPrefix(title, bookmarkTitlePrefix) {
			userAdded = true
			title = strings.TrimPrefix(title, bookmarkTitlePrefix)
		}

		bookmark := &Bookmark{
			Title:       title,
			PageNo:      pb.PageFrom,
			Level:       level,
			Bold:        pb.Bold,
			Italic:      pb.Italic,
			Children:    make([]*Bookmark, 0),
			IsUserAdded: userAdded,
		}
		
		// Recursively convert children
		if len(pb.Kids) > 0 {
			bookmark.Children = convertPdfcpuBookmarksRecursive(pb.Kids, level+1)
		}
		
		bookmarks = append(bookmarks, bookmark)
	}
	
	return bookmarks
}

// GetBookmarks returns all bookmarks
func (bm *BookmarkManager) GetBookmarks() []*Bookmark {
	return bm.bookmarks
}

// AddBookmark adds a new entry at the top level with optional region position.
// userAdded=true marks it as a user Bookmark (🔖/📍); false makes it a Table-of-Contents
// entry (📖). Both are written to the standard PDF outline on save.
func (bm *BookmarkManager) AddBookmark(title string, pageNo int, yOffset float32, userAdded bool) *Bookmark {
	bookmark := &Bookmark{
		Title:       title,
		PageNo:      pageNo,
		YOffset:     yOffset,
		Level:       1,
		Children:    make([]*Bookmark, 0),
		IsUserAdded: userAdded,
	}
	
	bm.bookmarks = append(bm.bookmarks, bookmark)
	
	if yOffset > 0 {
		log.Printf("[INFO] Added region bookmark '%s' to page %d at position %.0f", title, pageNo, yOffset)
	} else {
		log.Printf("[INFO] Added bookmark '%s' to page %d", title, pageNo)
	}
	
	return bookmark
}

// DeleteBookmark removes a bookmark
func (bm *BookmarkManager) DeleteBookmark(bookmark *Bookmark) bool {
	return bm.deleteBookmarkRecursive(&bm.bookmarks, bookmark)
}

// DeleteByType removes all top-level entries of the given kind (userAdded=true → bookmarks,
// false → TOC) and returns how many were removed.
func (bm *BookmarkManager) DeleteByType(userAdded bool) int {
	kept := make([]*Bookmark, 0, len(bm.bookmarks))
	removed := 0
	for _, b := range bm.bookmarks {
		if b.IsUserAdded == userAdded {
			removed++
			continue
		}
		kept = append(kept, b)
	}
	bm.bookmarks = kept
	return removed
}

func (bm *BookmarkManager) deleteBookmarkRecursive(bookmarks *[]*Bookmark, target *Bookmark) bool {
	for i, b := range *bookmarks {
		if b == target {
			*bookmarks = append((*bookmarks)[:i], (*bookmarks)[i+1:]...)
			return true
		}
		if bm.deleteBookmarkRecursive(&b.Children, target) {
			return true
		}
	}
	return false
}

// SaveBookmarks writes the current entries into the PDF's outline. With zero entries it
// strips the outline entirely (lets you clean a file), via RemoveBookmarks.
func (bm *BookmarkManager) SaveBookmarks(outputPath string) error {
	if bm.pdfPath == "" {
		return fmt.Errorf("no PDF loaded")
	}

	conf := model.NewDefaultConfiguration()

	// No entries: remove the outline. pdfcpu's Add path nil-panics on an empty list, so use
	// RemoveBookmarks. "No outline present" is treated as success (already the desired state).
	if len(bm.bookmarks) == 0 {
		err := api.RemoveBookmarksFile(bm.pdfPath, outputPath, conf)
		if errors.Is(err, api.ErrNoOutlines) {
			// Nothing to strip. For a new-file save, copy the source so an output still exists.
			if outputPath != "" && outputPath != bm.pdfPath {
				return copyFile(bm.pdfPath, outputPath)
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to clear bookmarks: %w", err)
		}
		log.Printf("[INFO] Removed all outline entries -> %s", outputPath)
		return nil
	}

	// Create bookmarks in PDF (replace=true to overwrite existing)
	pdfcpuBookmarks := bm.convertToPdfcpuFormat()
	err := api.AddBookmarksFile(bm.pdfPath, outputPath, pdfcpuBookmarks, true, conf)
	if err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}

	log.Printf("[INFO] Saved %d entries to %s", len(bm.bookmarks), outputPath)
	return nil
}

// copyFile copies src to dst (used when an empty-save targets a new file that already has no outline).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// convertToPdfcpuFormat converts our bookmarks to pdfcpu bookmark format
func (bm *BookmarkManager) convertToPdfcpuFormat() []pdfcpu.Bookmark {
	return bm.convertBookmarksRecursive(bm.bookmarks)
}

func (bm *BookmarkManager) convertBookmarksRecursive(bookmarks []*Bookmark) []pdfcpu.Bookmark {
	// pdfcpu requires sibling bookmarks in non-decreasing page order (it returns
	// "invalid bookmark" otherwise). Bookmarks are stored in add-order, so sort a copy
	// by page before writing. Stable sort keeps add-order among same-page bookmarks.
	sorted := make([]*Bookmark, len(bookmarks))
	copy(sorted, bookmarks)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].PageNo < sorted[j].PageNo })

	result := make([]pdfcpu.Bookmark, 0, len(sorted))

	for _, b := range sorted {
		// Tag user bookmarks with the marker so they reload as bookmarks (not TOC). Strip any
		// existing marker first to avoid doubling it on repeated saves.
		title := strings.TrimPrefix(b.Title, bookmarkTitlePrefix)
		if b.IsUserAdded {
			title = bookmarkTitlePrefix + title
		}

		pb := pdfcpu.Bookmark{
			Title:    title,
			PageFrom: b.PageNo,
			Bold:     b.Bold,
			Italic:   b.Italic,
		}

		// Recursively convert children
		if len(b.Children) > 0 {
			pb.Kids = bm.convertBookmarksRecursive(b.Children)
		}

		result = append(result, pb)
	}

	return result
}

// HasBookmarks returns true if there are any bookmarks
func (bm *BookmarkManager) HasBookmarks() bool {
	return len(bm.bookmarks) > 0
}
