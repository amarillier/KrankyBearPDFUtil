# KrankyBear PDF Utility - CLI, GUI & Viewer

![KrankyBearBeret](https://github.com/user-attachments/assets/95aef02f-a72c-4b82-aad2-5d2e4b30315a)

A powerful, easy-to-use PDF toolkit in three flavours: a **command-line tool (pdfutil)**, a **graphical app (pdfgui)**, and a **lightweight page viewer (pdfviewer)**. Built with [pdfcpu](https://github.com/pdfcpu/pdfcpu) by Horst H Rutter, [MuPDF](https://mupdf.com/) (via [go-fitz](https://github.com/gen2brain/go-fitz)) for high-quality page rendering, and [Fyne](https://fyne.io) for the GUIs.

The three ship together: a single installer per OS puts all of them in place, and pdfgui can launch pdfviewer (or your system default viewer) to preview a file.

---

## 🎯 Choose Your Interface

### 🖥️ **GUI Version (pdfgui)** - NEW!
**Perfect for:**
- Visual file selection and management
- Quick access to all features via sidebar
- Real-time property and permission viewing
- Password persistence across operations
- Drag-and-drop file handling
- System tray integration

**Highlights:**
- ✨ Auto-loading properties & permissions
- 🔒 Password persistence during session
- 🎨 Light/Dark/System theme support
- 📁 Smart file dialogs (remembers locations)
- 💾 In-place editing by default
- 🔄 Professional Help & About dialogs

→ **[GUI Documentation](pdfgui/README.md)**

### 💻 **CLI Version (pdfutil)**
**Perfect for:**
- Automation and scripting
- Batch processing
- Server environments
- Power users who prefer terminal
- Integration with other tools

**Highlights:**
- ⚡ Fast command-line operations
- 🔄 Easy integration with scripts
- 📊 Clear, formatted output
- 🎯 Simple command structure
- 🔗 Shell aliases support

### 👁️ **Viewer (pdfviewer)**
A deliberately minimalist, fast page viewer for reading PDFs (rendered with MuPDF). Launchable on its own, or from pdfgui's **View** button.

**Perfect for:**
- Quick reading without a heavyweight app
- Navigating long documents (keyboard, continuous scroll)
- Building or browsing a document's outline / personal bookmarks

**Highlights:**
- 📖🔖 **Table of Contents vs Bookmarks, clearly distinguished** — a genuinely unique touch: 📖 = the document's own outline (TOC), 🔖 = your page bookmarks, 📍 = a saved position on a page. An in-panel switch flips between them, and the **Add** button / shortcuts adapt to whichever you're viewing.
- 🏷️ **Bookmarks survive a round-trip** — saved into the standard PDF outline (portable to Preview/Acrobat/Foxit) with a small marker so they reload *as bookmarks*, not collapsed into the TOC — no sidecar file needed.
- ⌨️ **Keyboard-first** — PgUp/PgDn/←/→/Home/End to navigate; Cmd/Ctrl+D add a bookmark, Cmd/Ctrl+T add a TOC entry
- 🧭 **Continuous scroll** (lazy-rendered) or single-page; **Fit Width / Fit Page / 50–300%** zoom
- 💾 **Save options** — new file or overwrite original; "Delete All" / save-empty to strip an outline for cleanup
- 🕘 **Open Recent** (last 10) · 🖼️ drag-and-drop to open · 🎨 themes · system tray
- 🧠 Remembers your window size, zoom, panel mode, and continuous-scroll choice

→ **[Viewer Documentation](pdfviewer/README.md)**

---

## Highlights - All Three

- **🔐 Encryption & Security** - Protect PDFs with passwords and granular permissions (pdfutil/pdfgui)
- **📑 Page Manipulation** - Extract, remove, split, rotate, reverse, and merge pages (pdfutil/pdfgui)
- **🔒 Permission Control** - Set precise access restrictions (print, modify, extract, forms, etc.)
- **📊 PDF Analysis** - View detailed metadata and permission breakdowns
- **👁️ Reading & Outlines** - Minimalist viewer with high-quality rendering, continuous scroll, and a unique TOC-vs-bookmarks distinction (pdfviewer)
- **⚡ Simple & Intuitive** - Straightforward interface (GUI, CLI, or viewer)
- **🆓 100% Free** - Open source, no restrictions
- **🌍 Cross-Platform** - Windows, macOS, Linux

### Why KrankyBear PDF Utility?

This tool provides **simple, focused PDF operations** with your choice of interface. While [pdfcpu](https://pdfcpu.io/) offers comprehensive PDF manipulation, KrankyBear PDF Utility is designed for:

- Quick, common PDF tasks without complexity
- Easy-to-use interface (graphical or command-line)
- Visual permission displays and clear feedback
- Unique features like full page reversal and in-place editing
- Professional GUI with modern UX patterns

---

## Table of Contents
- [Installation](#installation)
- [GUI Quick Start](#gui-quick-start-pdfgui)
- [CLI Quick Start](#cli-quick-start-pdfutil)
- [Features](#features)
- [CLI Command Examples](#cli-command-examples)
- [GUI Features](#gui-features-details)
- [Building from Source](#building-from-source)
- [Advanced Usage](#advanced-usage-tips)
- [Quick Reference](#quick-reference)
- [License](#license)

---

## Installation

### Pre-built Binaries

Download the latest release for your platform:
- **Windows**: `KrankyBearPDFGuiSetup.exe` (GUI) or `pdfutil-windows-amd64.exe` (CLI)
- **macOS**: `KrankyBearPDF.app` (GUI) or `pdfutil-darwin-{arch}` (CLI)
- **Linux**: `.deb` or `.rpm` packages (GUI + CLI)

### From Source

See [Building from Source](#building-from-source) section below.

---

## GUI Quick Start (pdfgui)

### Launch the GUI

**macOS:**
```bash
./pdfgui
# or double-click KrankyBearPDF.app
```

**Windows:**
```cmd
pdfgui.exe
```

**Linux:**
```bash
./pdfgui
```

### Using the GUI

1. **Select a PDF** - Click any operation button, then browse to select your PDF
2. **Enter Passwords** - If needed, enter user/owner passwords (remembered during session)
3. **Perform Operations**:
   - **Encrypt/Decrypt** - Add or remove password protection
   - **Properties** - View/Edit metadata (auto-loads when you switch to it!)
   - **Permissions** - View/Set access controls (auto-loads too!)
   - **Page Operations** - Extract, Remove, Rotate, Reverse pages
   - **Merge/Split** - Combine or separate PDFs

4. **Change Theme** - Settings → Light/Dark/System Theme
5. **Get Help** - Help → Help (comprehensive guide)

**Pro Tips:**
- Once you open a PDF and enter passwords, they work for all operations until you change files
- File dialogs remember your last location
- Properties and permissions auto-load when you view them
- Edit operations auto-populate with current values
- Use system tray for quick access (right-click icon)

→ **[Full GUI Documentation](pdfgui/README.md)**

---

## CLI Quick Start (pdfutil)

### Get Help

```bash
pdfutil --help
```

### Check for Updates

```bash
pdfutil checkupdate
# or
pdfutil cu
```

### Common Operations

```bash
# View PDF info
pdfutil info -in document.pdf

# Encrypt a PDF
pdfutil encrypt -in document.pdf -up userpass -op ownerpass

# Decrypt a PDF
pdfutil decrypt -in encrypted.pdf -up userpass -out decrypted.pdf

# Merge PDFs
pdfutil merge -in file1.pdf file2.pdf file3.pdf -out combined.pdf

# Extract pages
pdfutil extract -in document.pdf -pages "1,5-10,15" -od output/

# View permissions
pdfutil permissions -in document.pdf -up userpass
```

---

## Features

Both CLI and GUI versions provide the following PDF operations:

* **✅ Check for updates** - Manual update checking (no auto-updates)
* **🔑 Change passwords** - Change user or owner passwords on encrypted PDFs
* **🔓 Decrypt files** - Remove password protection from PDFs
* **🔐 Encrypt files** - Add password protection with AES or RC4 encryption
* **📄 Extract pages** - Extract specific pages or ranges to separate files
* **ℹ️ File information** - View detailed PDF metadata and properties
* **➕ Insert pages** - Insert one PDF into another at any position
* **🔗 Merge files** - Combine multiple PDFs into a single file
* **❌ Remove pages** - Delete specific pages or ranges from a PDF
* **🔄 Reverse pages** - Reverse the order of all pages (last becomes first)
* **↻ Rotate pages** - Rotate pages by 90, 180, or 270 degrees
* **👁️ View permissions** - Display detailed permission breakdown with visual formatting
* **🔒 Set permissions** - Apply granular access controls (print, modify, extract, forms, etc.)
* **📝 Document properties** - List, add, and remove PDF metadata (title, author, subject, etc.)
* **✂️ Split pages** - Split a PDF into individual page files

### Future Additions
* **Watermark** - Apply watermarks to pages
* **Batch Operations** - Process multiple files at once (CLI already supports via scripting)

---

## CLI Command Examples

### 📄 File Information
View PDF metadata, encryption status, and permissions:
```bash
# Basic info (unencrypted file)
pdfutil info -in document.pdf

# Info on encrypted file
pdfutil fileinfo -in encrypted.pdf -up userpass -op ownerpass

# Aliases: info, fi, fileinfo, i, inf
```

### 🔐 Encryption & Decryption

**Encrypt a PDF:**
```bash
# AES-256 encryption (recommended)
pdfutil encrypt -in document.pdf -up userpass -op ownerpass -eb 256

# AES-128 encryption
pdfutil en -in document.pdf -up myuser -op myowner -eb 128

# RC4 encryption
pdfutil e -in document.pdf -up usr -op own -eb 40

# Aliases: encrypt, en, e, enc
```

**Decrypt a PDF:**
```bash
# Decrypt with user password
pdfutil decrypt -in encrypted.pdf -up userpass -out decrypted.pdf

# Decrypt with owner password
pdfutil de -in encrypted.pdf -op ownerpass -o decrypted.pdf

# Aliases: decrypt, de, d, dec
```

**Change passwords:**
```bash
# Change both passwords
pdfutil changepass -in encrypted.pdf -up olduser -op oldowner \
  -upnew newuser -opnew newowner -out updated.pdf

# Aliases: changepass, cp, chpass, changepw
```

### 🔒 Permissions Management

**View permissions:**
```bash
# Unencrypted PDF
pdfutil permissions -in document.pdf

# Encrypted PDF
pdfutil perm -in secured.pdf -up userpass

# Aliases: permissions, perm, p, perms
```

**Set permissions:**
```bash
# All permissions
pdfutil setpermissions -in document.pdf -pt all \
  -up userpass -op ownerpass -out secured.pdf

# Read-only (no modifications)
pdfutil sp -in document.pdf -pt none -up usr -op own -o readonly.pdf

# Allow printing only
pdfutil setp -in document.pdf -pt print -up usr -op own -o printonly.pdf

# Permission types: all, none, print, readonly, forms, annotate, modify

# Aliases: setpermissions, sp, setp, setperms
```

### 📝 Document Properties Management

**List properties:**
```bash
# Unencrypted PDF
pdfutil listproperties -in document.pdf

# Encrypted PDF
pdfutil lp -in secured.pdf -up userpass

# Aliases: listproperties, lp, listprop, list, ls
```

**Add properties:**
```bash
# Set title and author
pdfutil addproperty -in document.pdf \
  -title "My Document" -author "John Doe" -out updated.pdf

# Set all common properties
pdfutil ap -in document.pdf \
  -title "Report 2025" \
  -author "Jane Smith" \
  -subject "Annual Report" \
  -keywords "business, finance, 2025" \
  -o updated.pdf

# Aliases: addproperty, ap, addprop, add
```

**Remove properties:**
```bash
# Remove specific properties
pdfutil removeproperty -in document.pdf -title -author -out cleaned.pdf

# Remove all custom properties
pdfutil rp -in document.pdf -title -author -subject -keywords -o clean.pdf

# Aliases: removeproperty, rp, remprop, remove, rem
```

### 📄 Page Operations

**Extract pages:**
```bash
# Extract specific pages
pdfutil extract -in document.pdf -pages "1,3,5-10" -od extracted/

# Extract with zero-padding
pdfutil ex -in document.pdf -pages "1-100" -zeropad -od output/

# Aliases: extract, ex, e
```

**Remove pages:**
```bash
# Remove specific pages
pdfutil remove -in document.pdf -pages "2,4,6-8" -out cleaned.pdf

# Aliases: remove, rm, rem, r
```

**Rotate pages:**
```bash
# Rotate all pages 90 degrees clockwise
pdfutil rotate -in document.pdf -rotate 90 -out rotated.pdf

# Rotate specific pages 180 degrees
pdfutil ro -in document.pdf -rotate 180 -pages "1-5" -o flipped.pdf

# Rotation: 90, 180, 270, -90

# Aliases: rotate, ro, rot, r
```

**Reverse page order:**
```bash
# Reverse all pages
pdfutil reverse -in document.pdf -out reversed.pdf

# Aliases: reverse, rev, rv
```

**Split into individual pages:**
```bash
# Split PDF
pdfutil split -in document.pdf -od pages/

# Split with zero-padding
pdfutil sp -in document.pdf -zeropad -od pages/

# Aliases: split, sp, spl, s
```

### 🔗 Merge & Insert

**Merge PDFs:**
```bash
# Merge multiple files
pdfutil merge -in file1.pdf file2.pdf file3.pdf -out combined.pdf

# Merge with wildcard (shell expansion)
pdfutil m -in chapter*.pdf -o book.pdf

# Aliases: merge, m, mrg, combine
```

**Insert pages:**
```bash
# Insert at specific position (0-based)
pdfutil insert -in base.pdf -insert addition.pdf -position 5 -out combined.pdf

# Insert at beginning
pdfutil ins -in base.pdf -insert header.pdf -position 0 -o result.pdf

# Aliases: insert, ins, in, i
```

---

## GUI Features (Details)

### Smart Features

**Auto-Loading:**
- Properties auto-load when you click "View Properties"
- Permissions auto-load when you click "View Permissions"  
- Edit operations auto-populate with current values
- No extra clicks needed!

**Password Persistence:**
- Enter passwords once per file
- Work across all operations
- Cleared only when you change files

**Smart File Dialogs:**
- Remember last browsed location
- Start in current file's directory
- PDF filtering applied automatically
- Large, readable dialog windows

**In-Place Editing:**
- Default save option updates current file
- Transparent temporary file handling
- Original preserved on failure
- Or choose "save as" for new file

### Theme Support

**Three Themes:**
- **Light Theme** - Bright, clean interface
- **Dark Theme** - Easy on the eyes
- **System Theme** - Follows your OS setting

**Preference Persistence:**
- Theme choice saved automatically
- Restored on next launch
- Change anytime via Settings menu

### Professional UI

**Main Window:**
- Sidebar navigation for all operations
- Current file info panel
- Password entry with reveal toggles
- Clear operation sections

**Dialogs:**
- **About** - App info, version, links
- **Help** - Comprehensive guide (850x700 window)
- **Update Checker** - Manual update checking

**System Integration:**
- System tray icon
- Show/Hide from tray
- Quick access to all features
- Quit from tray or menu

### Permission Presets (GUI)

Easy-to-understand permission options:
- **all** - Full permissions (everything allowed)
- **none** - Complete lockdown (view only)
- **print** - Allow printing only
- **readonly** - Allow copy/extract text (no print)
- **forms** - Fill forms and annotate
- **annotate** - Add comments only
- **modify** - Full editing (but no print)

Each preset clearly labeled with what it allows!

---

## Building from Source

### Prerequisites

**All Platforms:**
- Go 1.21 or later
- Git

**For GUI:**
- C compiler (gcc, clang, or TDM-GCC on Windows)
- Platform-specific graphics libraries:
  - **macOS**: Xcode Command Line Tools
  - **Linux**: `libgl1-mesa-dev libx11-dev libxcursor-dev libxinerama-dev libxi-dev libxrandr-dev`
  - **Windows**: TDM-GCC or MinGW-w64

**Fyne CLI (for GUI):**
```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```

### Quick Build

**CLI Only (macOS):**
```bash
git clone https://github.com/amarillier/KrankyBearPDFUtil.git
cd KrankyBearPDFUtil
./compile.sh
./bin/pdfutil-darwin-arm64 --help
```

**GUI Only (macOS):**
```bash
cd gui
go build -o pdfgui
./pdfgui
```

**Everything (CLI + GUI, all platforms):**
```bash
./compile-all.sh
```

### Complete Build System

See **[BUILD-SYSTEM-GUIDE.md](BUILD-SYSTEM-GUIDE.md)** for comprehensive build documentation:
- Version management with `setver.sh`
- Cross-platform compilation
- Remote builds (Ubuntu, Linux Mint, Windows)
- Packaging for all platforms
- Complete workflow examples

**Quick commands:**
```bash
# Update version everywhere
./setver.sh 0.2.0

# Build everything locally
./compile-all.sh --local-only

# Build CLI only
./compile-all.sh --cli-only

# Build GUI only
./compile-all.sh --gui-only
```

---

## Advanced Usage Tips

### CLI Scripting

**Batch processing:**
```bash
# Process all PDFs in a directory
for pdf in *.pdf; do
  pdfutil encrypt -in "$pdf" -up pass -op admin -eb 256
done

# Extract first page from all PDFs
for pdf in *.pdf; do
  pdfutil extract -in "$pdf" -pages "1" -od first-pages/
done
```

**Pipeline integration:**
```bash
# Find and encrypt all unencrypted PDFs
find . -name "*.pdf" -type f | while read pdf; do
  if ! pdfutil info -in "$pdf" | grep -q "Encrypted: true"; then
    pdfutil encrypt -in "$pdf" -up user -op owner -eb 256
  fi
done
```

### GUI Workflows

**Typical workflow for editing a secured PDF:**
1. Open GUI
2. Click "View Properties"
3. Select your PDF, enter passwords
4. Properties auto-load → Review info ✓
5. Click "Edit Properties" (sidebar)
6. Fields auto-populate → Edit as needed ✓
7. Save (in-place by default)
8. Click "View Permissions"
9. Permissions auto-load → No re-entering passwords! ✓
10. Modify if needed via "Set Permissions"

**Password-protected workflow:**
- Enter passwords once when viewing properties
- Passwords remembered for all subsequent operations
- Work through permissions, properties, page operations
- No need to re-enter until you change files

---

## Quick Reference

### CLI Command Aliases

| Command | Aliases |
|---------|---------|
| `checkupdate` | `cu`, `update`, `u` |
| `changepass` | `cp`, `chpass`, `changepw` |
| `decrypt` | `de`, `d`, `dec` |
| `encrypt` | `en`, `e`, `enc` |
| `extract` | `ex`, `e` |
| `fileinfo` | `fi`, `info`, `i`, `inf` |
| `insert` | `ins`, `in`, `i` |
| `listproperties` | `lp`, `listprop`, `list`, `ls` |
| `addproperty` | `ap`, `addprop`, `add` |
| `removeproperty` | `rp`, `remprop`, `remove`, `rem` |
| `merge` | `m`, `mrg`, `combine` |
| `permissions` | `perm`, `p`, `perms` |
| `setpermissions` | `sp`, `setp`, `setperms` |
| `remove` | `rm`, `rem`, `r` |
| `reverse` | `rev`, `rv` |
| `rotate` | `ro`, `rot`, `r` |
| `split` | `sp`, `spl`, `s` |

### CLI Common Flag Aliases

| Full Flag | Short Aliases |
|-----------|---------------|
| `-infile` | `-in`, `-i` |
| `-outfile` | `-out`, `-o` |
| `-outdir` | `-od`, `-o` |
| `-userpass` | `-up`, `-u` |
| `-ownerpass` | `-op`, `-o` |
| `-pages` | `-pg`, `-p`, `-page`, `-pagelist` |
| `-encryptionbits` | `-eb`, `-e` |
| `-permtype` | `-pt`, `-perm` |
| `-rotate` | `-ro`, `-rot`, `-r` |
| `-zeropad` | `-pad`, `-z` |

### GUI Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl+Q` | Quit application |
| `Cmd/Ctrl+W` | Close window |
| `Cmd/Ctrl+M` | Minimize |

### GUI Menu Structure

**File → Operations:**
- Show
- Hide
- Quit

**Help:**
- About
- Check for Updates
- Help

**Settings:**
- Light Theme
- Dark Theme
- System Theme

### GUI System Tray

Right-click the tray icon for:
- Show/Hide window
- About, Help, Updates
- Theme selection
- Quit

---

## Known Issues / Limitations

### Both Versions
- **Encrypted files**: File operations (split, merge, page operations) do not work on encrypted PDFs
  - **Solution**: Decrypt first, perform operations, then re-encrypt
- **Setting permissions**: Requires file to already be encrypted
  - **Solution**: Encrypt file first, then set permissions

### Windows Specific
- Icon may not display properly in some contexts (cosmetic only)
- GUI requires CGO and C compiler (TDM-GCC or MinGW-w64)

### Linux Specific
- System tray may not work on all desktop environments (depends on DE support)

---

## Documentation

- **[BUILD-SYSTEM-GUIDE.md](BUILD-SYSTEM-GUIDE.md)** - Complete build system documentation
- **[pdfgui/README.md](pdfgui/README.md)** - GUI-specific documentation
- **[pdfviewer/README.md](pdfviewer/README.md)** - Viewer-specific documentation
- **[ReleaseNotes.txt](ReleaseNotes.txt)** - Version history

---

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## Credits

- **pdfcpu** by Horst H Rutter - The excellent PDF library powering this tool
- **Fyne** - The cross-platform GUI toolkit
- **KrankyBear Icons** - © Allan Marillier

---

## License

This is 100% free for anyone to use or misuse any way you like with no warranty as to suitability or anything else, other than it has no viruses when I compile and commit to git. But you should always check and scan anything you download from the internet for viruses anyway. Don't be reckless.

All KrankyBear icons, images, logos used are copyright (c) Allan Marillier, 2024, 2025...

Keep copies of your files and test features until you know how they work and trust the application.

In other words, I take no responsibility for how you use this, protect yourself.

See [LICENSE](LICENSE) file for full details.

---

## Support

- **Issues**: [GitHub Issues](https://github.com/amarillier/KrankyBearPDFUtil/issues)
- **Discussions**: [GitHub Discussions](https://github.com/amarillier/KrankyBearPDFUtil/discussions)

---

**Choose your interface, manage your PDFs with ease! 🐻📄**
