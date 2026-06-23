# Build Instructions for Kranky Bear PDF Utility

This document explains how to build both the CLI and GUI versions of the application.

## Project Structure

```
KrankyBearPDF/
├── main.go              # CLI application
├── util.go              # CLI utilities
├── compile.sh           # Build CLI binaries
├── gui/                 # GUI application
│   ├── main.go         # GUI entry point
│   ├── ui/             # GUI components
│   └── go.mod          # GUI dependencies
├── shared/             # Shared PDF operations
│   ├── pdfops.go      # PDF functions (used by both CLI and GUI)
│   └── go.mod         # Shared dependencies
├── compile-gui.sh      # Build GUI binaries
└── build-all.sh        # Build both CLI and GUI
```

## Building

### Prerequisites

**For CLI:**
- Go 1.24.2 or later
- No additional dependencies

**For GUI (additional requirements):**
- Fyne framework: `go install fyne.io/fyne/v2/cmd/fyne@latest`
- Platform-specific C compiler and graphics libraries

#### Platform-Specific GUI Requirements

**macOS:**
```bash
xcode-select --install
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install gcc libgl1-mesa-dev xorg-dev
```

**Linux (Fedora):**
```bash
sudo dnf install gcc mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel
```

**Windows:**
- Install TDM-GCC or MinGW-w64
- Or use MSYS2: `pacman -S mingw-w64-x86_64-gcc`

### Build Commands

**Build CLI only:**
```bash
./compile.sh
```
Output: `bin/pdfutil-{os}-{arch}`

**Build GUI only:**
```bash
./compile-gui.sh
```
Output: `bin-gui/pdfutil-gui-{os}-{arch}`

**Build both CLI and GUI:**
```bash
./build-all.sh
```

### Development Builds

**CLI:**
```bash
go run main.go util.go [command]
```

**GUI:**
```bash
cd gui
go run .
```

## Cross-Platform Compilation

### CLI Cross-Compilation

The CLI can easily cross-compile for all platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-windows-amd64.exe

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-linux-amd64

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-darwin-amd64

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o bin/pdfutil-darwin-arm64
```

### GUI Cross-Compilation

GUI cross-compilation is more complex due to CGO requirements. Best practice:
- Build on the target platform
- Or use Docker/CI for cross-platform builds

For macOS .app bundles:
```bash
cd gui
fyne package -os darwin -icon ../KrankyBearBeret.png -name "KrankyBearPDF"
```

## Distribution

### CLI Distribution
- Single binary, no dependencies
- Just copy the appropriate `pdfutil-{os}-{arch}` binary

### GUI Distribution

**macOS:**
- Use `fyne package` to create `.app` bundle
- Sign for distribution: `codesign --deep --force --verify --verbose --sign "Developer ID" KrankyBearPDF.app`

**Windows:**
- `.exe` file is standalone
- Optionally add to installer (NSIS, WiX, etc.)

**Linux:**
- Binary distribution
- Or package as AppImage, Flatpak, Snap, .deb, .rpm

## Module Management

The project uses Go modules with local replacements for shared code:

**Root go.mod:** Original CLI dependencies
**gui/go.mod:** GUI dependencies + references to `../shared`
**shared/go.mod:** Core PDF operation dependencies

To update dependencies:
```bash
# Update CLI
go get -u ./...
go mod tidy

# Update shared module
cd shared
go get -u ./...
go mod tidy

# Update GUI
cd ../gui
go get -u ./...
go mod tidy
```

## Troubleshooting

**"cannot find package" errors:**
```bash
# Ensure modules are initialized
cd shared && go mod tidy
cd ../gui && go mod tidy
cd .. && go mod tidy
```

**Fyne build errors on Linux:**
```bash
# Install missing graphics libraries
sudo apt-get install libgl1-mesa-dev xorg-dev
```

**macOS code signing issues:**
```bash
# For local testing, disable Gatekeeper temporarily
sudo spctl --master-disable
# Re-enable after testing
sudo spctl --master-enable
```

## Clean Build

To clean all build artifacts:
```bash
rm -rf bin/ bin-gui/
cd gui && go clean -cache
cd ../shared && go clean -cache
cd .. && go clean -cache
```

## CI/CD Considerations

For automated builds:
1. Use GitHub Actions or similar
2. Set up matrix builds for each platform
3. Use Docker for Linux builds
4. Use macOS runners for macOS builds
5. Use Windows runners for Windows builds

Example GitHub Actions workflow structure:
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    target: [cli, gui]
```

## License

See LICENSE file in project root.

