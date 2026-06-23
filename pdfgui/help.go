package main

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var helpWindow fyne.Window

// showHelp displays comprehensive help documentation
// Reusable pattern from KrankyBearClock - customize these for your app:
//   - appName: Your application name
//   - resourceKrankyBearBeanieMultiColorPng: Your embedded icon resource
//   - helpText: Your application's help content (see below for structure)
//   - GitHub and License URLs
//
// Help text structure recommendation:
//   - Use section headers with visual separators (━━━)
//   - Group related features together
//   - Include tips, tricks, and known limitations
//   - Add keyboard shortcuts
//   - Provide links to external resources
func showHelp(a fyne.App) {
	if helpWindow != nil && helpWindow.Content().Visible() {
		helpWindow.Show()
		helpWindow.RequestFocus()
		return
	}

	helpWindow = a.NewWindow(appName + " - Help")
	helpWindow.SetIcon(resourceKrankyBearBeanieMultiColorPng)

	// Customize this help text for your application
	helpText := `Kranky Bear PDF Utility - Comprehensive PDF Management

FILE OPERATIONS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Encrypt/Decrypt PDFs
  - AES 256/128/40 bit encryption
  - User and Owner password protection
  - Automatic password persistence during session

• Merge PDFs
  - Combine multiple PDF files into one
  - Maintains original quality
  - Smart directory navigation

• Split PDF
  - Split into individual pages
  - Optional zero-padded filenames
  - Batch processing

PAGE OPERATIONS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Extract Pages
  - Specify individual pages or ranges (e.g., 1,3-5,10-l)
  - Zero-padding option for filenames
  - Preserves page quality

• Remove Pages
  - Remove specific pages or ranges
  - Keeps original untouched (save-as)
  - Fast processing

• Rotate Pages
  - Rotate by 90, 180, or 270 degrees
  - Apply to specific pages or all pages
  - Maintains document structure

• Reverse Pages
  - Reverse the entire page order
  - Useful for scanning corrections
  - Creates new file

SECURITY & PERMISSIONS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• View Permissions (Auto-loads!)
  - See all current PDF permissions
  - Detailed permission breakdown
  - Encryption status

• Set Permissions
  - Control print, copy, modify access
  - Choose from common presets
  - In-place or save-as options

Common Permission Presets:
  - all: Full permissions (everything allowed)
  - none: Complete lockdown (view only)
  - print: Allow printing only (most common)
  - readonly: Allow copy/extract text (no print)
  - forms: Fill forms and annotate
  - modify: Full editing (but no print)

METADATA & PROPERTIES:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• View Properties (Auto-loads!)
  - Title, Author, Subject, Keywords
  - Creator, Producer
  - Creation and modification dates
  - PDF version and page count
  - Encryption status

• Edit Properties (Auto-populates!)
  - Modify document metadata
  - In-place or save-as options
  - Clear or reset anytime

SMART FEATURES:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✨ Auto-Load Properties: Properties display automatically when you
   click "View Properties" - no extra button click needed!

✨ Auto-Populate Editing: When editing properties, fields auto-fill
   with current values - ready to edit immediately!

✨ Password Persistence: Enter passwords once, they work across all
   operations until you change files

✨ Smart Dialogs: File/folder dialogs remember your last location
   and start there next time

✨ In-Place Editing: Default option updates the current file directly
   (transparent temp file handling behind the scenes)

✨ Theme Support: Light, Dark, or System theme - matches your preference

TIPS & TRICKS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💡 Workflow Optimization:
   1. Select PDF → Enter passwords once
   2. View Properties → Auto-loads! ✓
   3. Edit Properties → Auto-populates! ✓
   4. Set Permissions → Passwords already there! ✓
   5. All operations share the same file context

💡 Page Ranges:
   - Single: "5" (page 5)
   - Multiple: "1,3,5" (pages 1, 3, and 5)
   - Range: "1-10" (pages 1 through 10)
   - To End: "10-l" (page 10 to last)
   - Combined: "1,5-10,15-l"

💡 Permission Gotcha:
   Most presets DON'T include print permission!
   - Want print only? Use "print"
   - Want everything? Use "all"
   - "readonly" = copy only (NO print)
   - "modify" = edit only (NO print)

KNOWN LIMITATIONS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  Encrypted files cannot be:
   - Split into pages
   - Merged with other PDFs
   - Have page operations (extract, remove, rotate, reverse)
   
   Solution: Decrypt first, then perform operations

⚠️  Setting permissions requires:
   - File must already be encrypted
   - Owner password is required
   
   Solution: Encrypt file first, then set permissions

⚠️  In-place editing:
   - Uses temporary files behind the scenes
   - Original replaced only on success
   - Failed operations leave original untouched

KEYBOARD SHORTCUTS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Standard system shortcuts apply:
• Cmd/Ctrl+Q - Quit
• Cmd/Ctrl+W - Close window
• Cmd/Ctrl+M - Minimize

MORE INFORMATION:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
For detailed documentation, bug reports, or feature requests:
📦 GitHub: https://github.com/amarillier/KrankyBearPDFUtil
📄 License: https://github.com/amarillier/KrankyBearPDFUtil/blob/main/LICENSE
📝 Release Notes: Check "Help → Check for Updates"

FREE SOFTWARE - Use anywhere, anytime, any purpose!
No registration, no tracking, no phone-home (except manual update checks).
`

	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	// Links - update URLs for your project
	githubURL, _ := url.Parse("https://github.com/amarillier/KrankyBearPDFUtil")
	githubLink := widget.NewHyperlink("Visit GitHub Repository", githubURL)
	githubLink.Alignment = fyne.TextAlignCenter

	licenseURL, _ := url.Parse("https://github.com/amarillier/KrankyBearPDFUtil/blob/allanm/LICENSE")
	licenseLink := widget.NewHyperlink("View License", licenseURL)
	licenseLink.Alignment = fyne.TextAlignCenter

	// Create scrollable area with minimum size for better readability
	scrollContent := container.NewScroll(helpLabel)
	scrollContent.SetMinSize(fyne.NewSize(750, 550))

	// Layout with better proportions
	header := container.NewVBox(
		widget.NewLabelWithStyle("Kranky Bear PDF Utility - Help", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(container.NewHBox(githubLink, licenseLink)),
	)

	content := container.NewBorder(header, footer, nil, nil, scrollContent)

	helpWindow.SetContent(container.NewPadded(content))
	helpWindow.Resize(fyne.NewSize(850, 700))

	helpWindow.SetCloseIntercept(func() {
		helpWindow.Hide()
	})

	helpWindow.Show()
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
