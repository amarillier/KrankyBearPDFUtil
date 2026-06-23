# Setting Up the GUI Version - Quick Start Guide

This guide will help you set up and build the GUI version of Kranky Bear PDF Utility.

## What's Been Created

Your project now has both CLI and GUI versions that share the same core PDF functionality:

```
KrankyBearPDF/
├── CLI Version (existing)
│   ├── main.go              # CLI entry point
│   ├── util.go              # CLI utilities
│   └── compile.sh           # Build CLI
│
├── GUI Version (new)
│   ├── gui/
│   │   ├── main.go         # GUI entry point
│   │   ├── ui/             # UI components
│   │   │   ├── encrypt_decrypt.go
│   │   │   ├── merge_split.go
│   │   │   ├── pages.go
│   │   │   ├── permissions.go
│   │   │   └── properties.go
│   │   └── go.mod          # GUI dependencies
│   └── compile-gui.sh      # Build GUI
│
└── Shared Code (new)
    ├── shared/
    │   ├── pdfops.go       # Shared PDF operations
    │   └── go.mod          # Shared dependencies
    └── build-all.sh        # Build both versions
```

## Step 1: Install Fyne

First, install the Fyne package tool:

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH.

## Step 2: Install Platform Dependencies

### macOS
```bash
xcode-select --install
```

### Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install gcc libgl1-mesa-dev xorg-dev
```

### Linux (Fedora)
```bash
sudo dnf install gcc mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel
```

### Windows
Download and install TDM-GCC from: https://jmeubank.github.io/tdm-gcc/

## Step 3: Initialize Go Modules

From the project root:

```bash
# Initialize shared module
cd shared
go mod tidy
cd ..

# Initialize GUI module
cd gui
go mod tidy
cd ..
```

## Step 4: Test the GUI

Quick test to make sure everything works:

```bash
cd gui
go run .
```

This should open a window with the GUI application. If it works, you're ready to build!

## Step 5: Build the GUI

From the project root:

```bash
# Build GUI only
./compile-gui.sh

# Or build both CLI and GUI
./build-all.sh
```

Your binaries will be in:
- `bin/` - CLI binaries
- `bin-gui/` - GUI binaries

## GUI Features

The GUI provides all your existing CLI features with a visual interface:

### 🔐 Encrypt/Decrypt Tab
- Select PDF file
- Enter user/owner passwords
- Choose encryption mode (AES/RC4) and bits (40/128/256)
- One-click encrypt or decrypt

### 🔗 Merge/Split Tab
- **Merge**: Add multiple PDFs and merge into one
- **Split**: Split a PDF into individual pages with optional zero-padding

### 📑 Pages Tab
- **Extract**: Extract specific pages (e.g., 1,3-5,10-l)
- **Remove**: Remove pages from PDF
- **Rotate**: Rotate pages by 90/180/270 degrees
- **Reverse**: Reverse all pages

### 🔒 Permissions Tab
- **View**: See detailed permission breakdown
- **Set**: Apply permission profiles (all, none, print, readonly, forms, annotate, modify)
- Modify in-place or save to new file

### 📄 Properties Tab
- **View**: See all document properties
- **Add/Set**: Set title, author, subject, keywords, creator, producer

## Usage Tips

1. **File Selection**: All tabs have "Select PDF File" buttons that open native file dialogs
2. **Passwords**: Password fields are masked for security
3. **Progress**: Status messages appear after operations complete
4. **Errors**: Clear error messages if something goes wrong
5. **Encrypted Files**: Most operations require decryption first (except decrypt and view info)

## Distribution

### macOS App Bundle
```bash
cd gui
fyne package -os darwin -icon ../KrankyBearBeret.png -name "KrankyBearPDF"
```
Creates `KrankyBearPDF.app`

### Windows Installer
The `.exe` in `bin-gui/` is standalone and can be distributed as-is or packaged with an installer.

### Linux
Distribute the binary or package as AppImage, Flatpak, Snap, etc.

## Maintaining Both Versions

### Workflow
1. Add new PDF features to `shared/pdfops.go`
2. Add CLI interface in `main.go` (if needed)
3. Add GUI interface in `gui/ui/*.go` (if needed)
4. Both CLI and GUI automatically use the shared code

### Build Process
```bash
# During development - quick test
cd gui && go run .

# For release - build all platforms
./build-all.sh
```

## Troubleshooting

### "cannot find package pdfutil/shared"
```bash
cd shared && go mod tidy
cd ../gui && go mod tidy
```

### GUI window doesn't open
- Check that graphics libraries are installed (Linux)
- On macOS, make sure Xcode Command Line Tools are installed
- On Windows, verify TDM-GCC is installed

### Build errors on Linux
```bash
# Make sure you have all required libraries
sudo apt-get install gcc libgl1-mesa-dev xorg-dev libx11-dev
```

### macOS "damaged" or "unverified developer" warnings
For local testing:
```bash
xattr -cr KrankyBearPDF.app
```

For distribution, sign the app:
```bash
codesign --deep --force --verify --verbose --sign "Developer ID" KrankyBearPDF.app
```

## Next Steps

1. **Test thoroughly**: Try all features in the GUI
2. **Customize**: Adjust colors, layouts, or add features in `gui/ui/`
3. **Package**: Create installers for distribution
4. **Document**: Update user documentation with GUI screenshots

## Getting Help

- Fyne documentation: https://developer.fyne.io/
- Your shared code is in: `shared/pdfops.go`
- GUI components are in: `gui/ui/*.go`

## License

Same as main project - 100% free to use!

