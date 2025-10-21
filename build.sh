#!/bin/bash
# Build script for KrankyBear PDFUtil
# Builds binaries for Windows, Linux, and macOS

set -e

echo "Building KrankyBear PDFUtil..."
echo ""

# Create bin directory
mkdir -p bin

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-windows-amd64.exe

# Build for Linux (amd64)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-linux-amd64

# Build for macOS Intel (amd64)
echo "Building for macOS Intel (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-darwin-amd64

# Build for macOS Apple Silicon (arm64)
echo "Building for macOS Apple Silicon (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o bin/pdfutil-darwin-arm64

echo ""
echo "Build complete! Binaries are in the 'bin/' directory:"
ls -lh bin/
cp bin/pdfutil-darwin-arm64 ./pdfutil
