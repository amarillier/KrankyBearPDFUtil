package ui

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"sync"

	"github.com/gen2brain/go-fitz"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// PDFRenderer handles PDF page rendering using pdfcpu
type PDFRenderer struct {
	pdfPath   string
	pageCount int
	mu        sync.RWMutex

	// Cache for rendered pages
	cache    map[string]image.Image
	cacheMu  sync.RWMutex
	maxCache int
}

// NewPDFRenderer creates a new PDF renderer
func NewPDFRenderer() *PDFRenderer {
	return &PDFRenderer{
		cache:    make(map[string]image.Image),
		maxCache: 10, // Cache up to 10 pages
	}
}

// LoadPDF loads a PDF file for rendering
func (r *PDFRenderer) LoadPDF(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Quick validation with go-fitz to get page count
	doc, err := fitz.New(path)
	if err != nil {
		// Fall back to pdfcpu validation if fitz fails
		conf := model.NewDefaultConfiguration()
		if err := api.ValidateFile(path, conf); err != nil {
			return fmt.Errorf("invalid PDF: %w", err)
		}
		
		// Get page count from pdfcpu
		ctx, err := api.ReadContextFile(path)
		if err != nil {
			return fmt.Errorf("failed to read PDF: %w", err)
		}
		r.pageCount = ctx.PageCount
	} else {
		r.pageCount = doc.NumPage()
		doc.Close()
	}

	r.pdfPath = path

	// Clear cache
	r.cacheMu.Lock()
	r.cache = make(map[string]image.Image)
	r.cacheMu.Unlock()

	return nil
}

// GetPageCount returns the total number of pages
func (r *PDFRenderer) GetPageCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pageCount
}

// RenderPage renders a specific page at the given zoom level
func (r *PDFRenderer) RenderPage(pageNum int, zoom float32) (image.Image, error) {
	r.mu.RLock()
	pdfPath := r.pdfPath
	pageCount := r.pageCount
	r.mu.RUnlock()

	if pdfPath == "" {
		return nil, fmt.Errorf("no PDF loaded")
	}

	if pageNum < 1 || pageNum > pageCount {
		return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, pageCount)
	}

	// Check cache
	cacheKey := fmt.Sprintf("%s:%d:%.2f", pdfPath, pageNum, zoom)
	r.cacheMu.RLock()
	if img, ok := r.cache[cacheKey]; ok {
		r.cacheMu.RUnlock()
		return img, nil
	}
	r.cacheMu.RUnlock()

	// Render the page
	img, err := r.renderPageToPNG(pdfPath, pageNum, zoom)
	if err != nil {
		return nil, err
	}

	// Add to cache
	r.cacheMu.Lock()
	// Simple cache eviction: if cache is full, clear it
	if len(r.cache) >= r.maxCache {
		r.cache = make(map[string]image.Image)
	}
	r.cache[cacheKey] = img
	r.cacheMu.Unlock()

	return img, nil
}

// renderPageToPNG renders a PDF page to a PNG image using MuPDF (via go-fitz)
func (r *PDFRenderer) renderPageToPNG(pdfPath string, pageNum int, zoom float32) (image.Image, error) {
	// Use go-fitz (MuPDF) for high-quality PDF rendering
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF with MuPDF: %w", err)
	}
	defer doc.Close()

	totalPages := doc.NumPage()
	if pageNum < 1 || pageNum > totalPages {
		return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, totalPages)
	}

	// Calculate DPI based on zoom level
	// Base DPI is 150, scaled by zoom factor
	baseDPI := 150.0
	dpi := baseDPI * float64(zoom)
	
	// Clamp DPI to reasonable values (min 50, max 600)
	if dpi < 50 {
		dpi = 50
	}
	if dpi > 600 {
		dpi = 600
	}

	// Render the page (go-fitz uses 0-based page indexing)
	img, err := doc.ImageDPI(pageNum-1, dpi)
	if err != nil {
		return nil, fmt.Errorf("failed to render page %d at %.1f DPI: %w", pageNum, dpi, err)
	}

	// Verify image was created
	if img == nil {
		return nil, fmt.Errorf("rendered image is nil for page %d", pageNum)
	}

	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return nil, fmt.Errorf("rendered image has zero size for page %d", pageNum)
	}

	return img, nil
}

// createPlaceholderWithInfo creates a placeholder with PDF information
func (r *PDFRenderer) createPlaceholderWithInfo(pdfPath string, pageNum, totalPages int) image.Image {
	return r.createPlaceholderImage(pageNum, fmt.Sprintf("PDF Viewer - Placeholder Mode\n\nPage %d of %d\n\n%s\n\nNote: Full PDF rendering requires\nMuPDF or Poppler libraries.", pageNum, totalPages, filepath.Base(pdfPath)))
}

// createPlaceholderImage creates a placeholder image with text
func (r *PDFRenderer) createPlaceholderImage(pageNum int, message string) image.Image {
	// Create a simple white image with message
	width, height := 612, 792 // Standard letter size in points
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with light gray background
	lightGray := image.NewUniform(image.White)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, lightGray)
		}
	}

	// Note: Actual text rendering would require additional libraries
	// For now, this creates a blank placeholder page
	// In a real implementation, you'd draw the text here

	return img
}
