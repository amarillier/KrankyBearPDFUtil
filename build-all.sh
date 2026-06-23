#!/bin/bash
# Master build script - builds both CLI and GUI versions
# Builds binaries for Windows, Linux, and macOS

set -e

echo "========================================"
echo "Building Kranky Bear PDFUtil"
echo "Building both CLI and GUI versions"
echo "========================================"
echo ""

# Build CLI version
echo "Building CLI version..."
./compile.sh
echo ""

# Build GUI version
echo "Building GUI version..."
./compile-gui.sh
echo ""

echo "========================================"
echo "All builds complete!"
echo "========================================"
echo ""
echo "CLI binaries are in: bin/"
echo "GUI binaries are in: bin-gui/"
echo ""
ls -lh bin/
echo ""
ls -lh bin-gui/

