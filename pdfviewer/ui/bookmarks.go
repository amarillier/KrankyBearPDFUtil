package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Bookmark represents a PDF bookmark/outline entry
type Bookmark struct {
	Title    string
	PageNo   int
	YOffset  float32        // Vertical scroll position (region bookmark)
	Children []*Bookmark
	Level    int
	Bold     bool
	Italic   bool
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
		bookmark := &Bookmark{
			Title:    pb.Title,
			PageNo:   pb.PageFrom,
			Level:    level,
			Bold:     pb.Bold,
			Italic:   pb.Italic,
			Children: make([]*Bookmark, 0),
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

// AddBookmark adds a new bookmark at the top level with optional region position
func (bm *BookmarkManager) AddBookmark(title string, pageNo int, yOffset float32) *Bookmark {
	bookmark := &Bookmark{
		Title:    title,
		PageNo:   pageNo,
		YOffset:  yOffset,
		Level:    1,
		Children: make([]*Bookmark, 0),
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

// SaveBookmarks saves bookmarks back to the PDF
func (bm *BookmarkManager) SaveBookmarks(outputPath string) error {
	if bm.pdfPath == "" {
		return fmt.Errorf("no PDF loaded")
	}

	conf := model.NewDefaultConfiguration()
	
	// Convert our bookmarks to pdfcpu bookmark format
	pdfcpuBookmarks := bm.convertToPdfcpuFormat()
	
	// Create bookmarks in PDF (replace=true to overwrite existing)
	err := api.AddBookmarksFile(bm.pdfPath, outputPath, pdfcpuBookmarks, true, conf)
	if err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}

	log.Printf("[INFO] Saved %d bookmarks to %s", len(bm.bookmarks), outputPath)
	return nil
}

// convertToPdfcpuFormat converts our bookmarks to pdfcpu bookmark format
func (bm *BookmarkManager) convertToPdfcpuFormat() []pdfcpu.Bookmark {
	return bm.convertBookmarksRecursive(bm.bookmarks)
}

func (bm *BookmarkManager) convertBookmarksRecursive(bookmarks []*Bookmark) []pdfcpu.Bookmark {
	result := make([]pdfcpu.Bookmark, 0)
	
	for _, b := range bookmarks {
		pb := pdfcpu.Bookmark{
			Title:    b.Title,
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
