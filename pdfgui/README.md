# KrankyBear PDF Utility - GUI (pdfgui)

![KrankyBear Icon](Resources/Images/KrankyBearBeanieMultiColor.png)

Graphical user interface for comprehensive PDF management. Built with [Fyne](https://fyne.io) for a modern, cross-platform experience.

---

## Features

### Smart & Intuitive
- ✨ **Auto-Loading** - Properties and permissions load automatically
- 🔒 **Password Persistence** - Enter once, use everywhere (per file)
- 📁 **Smart Dialogs** - Remembers last location, starts in current directory
- 💾 **In-Place Editing** - Default option updates files directly
- 🎯 **Clear Layout** - Sidebar navigation with organized operations

### Professional UI
- 🎨 **Theme Support** - Light, Dark, or System theme
- 🖼️ **System Tray** - Quick access, show/hide window
- 📖 **Help System** - Comprehensive built-in documentation
- ℹ️ **About Dialog** - Version info and links
- 🔄 **Update Checker** - Manual update checking

### Complete PDF Operations
- **Encryption** - Add/remove password protection (AES 256/128/40, RC4)
- **Properties** - View and edit metadata (auto-loads!)
- **Permissions** - View and set access controls (auto-loads!)
- **Page Operations** - Extract, Remove, Rotate, Reverse
- **Merge/Split** - Combine or separate PDFs

---

## Quick Start

### Launch

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

### First Time Use

1. **Select an Operation** - Click any button in the sidebar (e.g., "View Properties")
2. **Choose Your PDF** - Browse to select your PDF file
3. **Enter Passwords** (if needed) - Enter user/owner passwords
4. **Done!** - The operation completes or data displays

**Pro Tip:** After you've opened a file and entered passwords, they work for all operations until you change files!

---

## Interface Guide

### Main Window Layout

```
┌─────────────────────────────────────────────────────┐
│ File  Help  Settings                    [_][□][X]  │
├──────────┬──────────────────────────────────────────┤
│          │  Current File Info                       │
│ Encrypt  │  File: document.pdf                      │
│ Decrypt  │  Location: /Users/.../Documents          │
│          ├──────────────────────────────────────────┤
│ Properties│                                          │
│ - View   │                                          │
│ - Edit   │         Operation Content Area           │
│          │                                          │
│ Permissions│                                        │
│ - View   │                                          │
│ - Set    │                                          │
│          │                                          │
│ Pages    │                                          │
│ - Extract│                                          │
│ - Remove │                                          │
│ - Rotate │                                          │
│ - Reverse│                                          │
│          │                                          │
│ Merge    │                                          │
│ Split    │                                          │
└──────────┴──────────────────────────────────────────┘
```

### Sidebar Operations

**Encryption:**
- **Encrypt** - Add password protection
- **Decrypt** - Remove passwords

**Properties:**
- **View Properties** - Display metadata (auto-loads!)
- **Edit Properties** - Modify metadata (auto-populates!)

**Permissions:**
- **View Permissions** - Show access controls (auto-loads!)
- **Set Permissions** - Change restrictions

**Page Operations:**
- **Extract** - Save specific pages
- **Remove** - Delete pages
- **Rotate** - Turn pages
- **Reverse** - Reverse page order

**File Operations:**
- **Merge** - Combine PDFs
- **Split** - Separate into pages

---

## Smart Features Explained

### Auto-Loading

**View Properties:**
1. Click "View Properties" in sidebar
2. Select PDF (if not already selected)
3. Enter passwords (if needed)
4. **Properties display automatically** - No "Load" button needed!
5. Click "View Properties (Refresh)" to reload

**View Permissions:**
- Same as properties - automatically loads and displays
- No extra clicks required

**Edit Properties:**
1. Click "Edit Properties" in sidebar  
2. If you've already viewed properties, **fields auto-populate** with current values
3. Edit as needed
4. Save (in-place or as new file)

### Password Persistence

```
Workflow Example:
1. Click "View Properties"
2. Select PDF → Enter user password: "mypass"
3. Properties display ✓
4. Click "View Permissions" (sidebar)
5. Permissions display - NO password prompt! ✓
6. Click "Edit Properties"
7. Fields populated - NO password prompt! ✓
8. Click "Set Permissions"
9. Works without re-entering password! ✓

Change file → Passwords cleared, ready for new file
```

### Smart File Dialogs

**Features:**
- **Remember Location** - Starts where you last browsed
- **Current Directory** - Defaults to current file's location
- **PDF Filtering** - Shows only PDF files
- **Large Size** - 850x700 for easy browsing
- **Fast Navigation** - Responsive directory changes

**Behavior:**
```
First operation: Starts in ~/Documents (or current working dir)
Browse to: /Users/you/Projects/Reports/
Select file: annual-report.pdf
Next operation: Starts in /Users/you/Projects/Reports/
```

### In-Place Editing

**Default Behavior:**
- Saves to the same file automatically
- Behind the scenes: creates temp file → applies changes → replaces original
- Original preserved if operation fails

**How It Works:**
```
You select: document.pdf
You edit properties and click "Save"
GUI does:
  1. Create: document.pdf.tmp
  2. Apply changes to temp file
  3. Success? Replace original
  4. You see: document.pdf (updated)
  5. Failure? Original untouched
```

**Alternative:**
- Check "Save as new file" option
- Choose new name/location
- Original remains unchanged

---

## Menu Reference

### File → Operations
- **Show** - Show the main window
- **Hide** - Hide to system tray
- **Quit** - Exit application

### Help
- **About** - Version, copyright, links
- **Check for Updates** - Manual update check (no auto-updates)
- **Help** - Comprehensive guide (this information + more)

### Settings
- **Light Theme** - Bright, clean interface
- **Dark Theme** - Easy on the eyes
- **System Theme** - Follows your OS setting

**Theme Persistence:**
Your choice is saved and restored on next launch.

---

## System Tray

Right-click the system tray icon for quick access:

```
Kranky Bear PDF Utility
├── Show
├── Hide
├── ─────────────
├── About
├── Help
├── Check for Updates
├── ─────────────
├── Light Theme
├── Dark Theme
├── System Theme
├── ─────────────
└── Quit
```

**Platform Notes:**
- **macOS**: Icon in menu bar (top right)
- **Windows**: Icon in notification area (bottom right)
- **Linux**: Icon in system tray (depends on desktop environment)

---

## Permission Presets

When setting permissions, choose from these presets:

| Preset | What It Allows |
|--------|----------------|
| **all** | Full permissions - Everything allowed |
| **none** | Complete lockdown - View only, nothing else |
| **print** | Allow printing only (most common) |
| **readonly** | Allow copy/extract text only (NO print) |
| **forms** | Fill forms and annotate |
| **annotate** | Add comments only |
| **modify** | Full editing (but NO print) |

**Important Notes:**
- Most presets DON'T include print permission
- Want printing? Use "print" or "all"
- "readonly" means text extraction, NOT printing
- File must already be encrypted to set permissions

---

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl+Q` | Quit application |
| `Cmd/Ctrl+W` | Close window |
| `Cmd/Ctrl+M` | Minimize window |

Standard shortcuts work as expected on each platform.

---

## Tips & Tricks

### Efficient Workflows

**Working with Multiple Properties:**
```
1. View Properties → Auto-loads ✓
2. Edit Properties → Auto-populates ✓
3. Make changes → Save ✓
4. View Properties → See updates immediately ✓
```

**Secure PDF Workflow:**
```
1. Encrypt → Add passwords
2. View Permissions → Check current settings
3. Set Permissions → Apply restrictions
4. View Properties → Verify encryption
All using the same passwords you entered once!
```

### Page Range Syntax

When specifying pages (Extract, Remove, Rotate):
- **Single**: `5` (page 5)
- **Multiple**: `1,3,5` (pages 1, 3, and 5)
- **Range**: `1-10` (pages 1 through 10)
- **To End**: `10-l` (page 10 to last)
- **Combined**: `1,5-10,15-l`

### Merge PDFs Efficiently

The merge dialog remembers your last directory:
```
1. Click "Add PDF Files"
2. Browse to /your/pdfs/directory
3. Select file1.pdf
4. Click "Add PDF Files" again
5. Already in /your/pdfs/directory!
6. Select file2.pdf
7. Repeat as needed
```

---

## Building the GUI

### Prerequisites

- Go 1.21+
- C compiler (gcc, clang, TDM-GCC)
- Platform-specific graphics libraries

**macOS:**
```bash
xcode-select --install
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt install libgl1-mesa-dev libx11-dev libxcursor-dev \
  libxinerama-dev libxi-dev libxrandr-dev libxss-dev libxxf86vm-dev
```

**Windows:**
- Install TDM-GCC or MinGW-w64
- Ensure gcc is in PATH

### Build Commands

**Simple build:**
```bash
cd gui
go build -o pdfgui
./pdfgui
```

**Fyne package (creates proper app bundle):**
```bash
# macOS
fyne package -os darwin -icon Resources/Images/KrankyBearBeanieMultiColor.png

# Windows  
fyne package -os windows -icon Resources/Images/KrankyBearBeanieMultiColor.png

# Linux
fyne package -os linux -icon Resources/Images/KrankyBearBeanieMultiColor.png
```

**Using build script:**
```bash
cd ..
./compile-gui.sh
```

---

## Known Issues

### General
- Encrypted PDFs cannot be split, merged, or have page operations
  - **Solution**: Decrypt first, operate, then re-encrypt
- Setting permissions requires file to already be encrypted
  - **Solution**: Encrypt first, then set permissions

### Platform-Specific

**Linux:**
- System tray may not work on all desktop environments
- Depends on DE support for system tray protocol

**Windows:**
- First launch may be slow (Windows security scan)
- Icon requires winres.json to be built with `go generate`

**macOS:**
- App bundle needs to be created with `fyne package` for proper icon
- Simple `go build` won't include icon in dock

---

## Troubleshooting

### "Failed to create window"
- **Cause**: Graphics libraries not installed
- **Solution**: Install platform-specific graphics libs (see Prerequisites)

### "Permission denied" on save
- **Cause**: File is read-only or locked
- **Solution**: Check file permissions, close other programs using the file

### Passwords not working
- **Cause**: Incorrect password or wrong password type
- **Solution**: Try both user and owner passwords, check caps lock

### System tray icon not showing (Linux)
- **Cause**: Desktop environment doesn't support system tray
- **Solution**: Use main menu instead, all features accessible there

---

## Project Structure

```
gui/
├── main.go              # Application entry point
├── about.go             # About dialog (reusable)
├── help.go              # Help dialog (reusable)
├── dialogs.go           # Update checker dialog
├── theme.go             # Theme switching
├── util.go              # Update checker function
├── bundled.go           # Embedded icons (generated)
│
├── ui/
│   ├── dashboard.go     # Main layout
│   ├── sidebar.go       # Navigation
│   ├── shared_state.go  # File/password state
│   ├── encrypt_decrypt_split.go
│   ├── properties.go
│   ├── permissions.go
│   ├── pages.go
│   ├── merge_split.go
│   ├── file_helpers.go  # In-place editing
│   └── password_helper.go
│
├── Resources/
│   └── Images/
│       ├── KrankyBearBeanieMultiColor.png
│       └── KrankyBearBeanieMultiColor64.png
│
├── winres/              # Windows resources
│   └── winres.json
│
└── Inno/                # Windows installer
    └── KrankyBearPDFGui.iss
```

---

## Reusable Components

The GUI uses reusable dialog patterns that can be copied to other projects:

- `about.go` - Professional about dialog
- `help.go` - Comprehensive help dialog
- `theme.go` - Theme switching system

See **[REUSABLE-DIALOGS-GUIDE.md](REUSABLE-DIALOGS-GUIDE.md)** for how to use these in your own Fyne projects.

---

## Version History

See [ReleaseNotes.txt](../ReleaseNotes.txt) for complete version history.

---

## Contributing

GUI contributions welcome! Areas of interest:
- Additional file operations
- UI/UX improvements
- Theme customizations
- Localization/internationalization
- Platform-specific enhancements

---

## License

Same as main project - 100% free to use. See [LICENSE](../LICENSE) for details.

---

**Enjoy the graphical interface! 🐻🖥️**
