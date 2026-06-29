#!/usr/bin/env bash
# Central project identifiers for compile / sync / package scripts.
# Single source of truth so the build scripts never drift apart again.
# Bump the shipping version with ./setver.sh X.Y.Z (updates KB_VERSION_DEFAULT and all consumer files).

# Human-readable project name (window titles, .desktop Name, etc.)
export KB_PROJECT_TITLE="KrankyBear PDF"

# macOS .app bundle / Windows install dir / Linux /opt dir name
export KB_APP_NAME="KrankyBearPDF"

# Bundle identifier (macOS Info.plist)
export KB_BUNDLE_ID="com.krankybear.pdf"

# fpm / package base name: produces installers/${KB_FPM_NAME}_<ver>-<iter>_<arch>.{pkg,deb,rpm}
export KB_FPM_NAME="krankybear-pdf"

# The three binaries this project ships. Each is built into ./bin as
#   <base>-<os>-<arch>[.exe]   e.g. bin/pdfgui-macos-arm64, bin/pdfutil-windows-amd64.exe
# Build dirs (Go modules) relative to repo root:
#   CLI    -> .            (module pdfutil)
#   GUI    -> ./pdfgui     (module pdfutil/pdfgui)  -- primary launchable app
#   VIEWER -> ./pdfviewer  (module pdfutil/pdfviewer) -- bundled helper for now
export KB_CLI_BIN="pdfutil"
export KB_CLI_DIR="."
export KB_GUI_BIN="pdfgui"
export KB_GUI_DIR="pdfgui"
export KB_VIEWER_BIN="pdfviewer"
export KB_VIEWER_DIR="pdfviewer"

# App icon (PNG). winres embeds it on Windows; setIcon.sh stamps mac binaries/pkg.
export KB_ICON="assets/images/KrankyBearBeanieMultiColor.png"

export KB_VERSION_DEFAULT="0.3.1"

export KB_HOMEPAGE="https://github.com/amarillier/KrankyBearPDF"
export KB_MAINTAINER_DEFAULT="amarillier@gmail.com"
export KB_VENDOR_DEFAULT="Allan Marillier"
export KB_LICENSE_DEFAULT="GNU GPL v3"

# Windows Inno Setup script (run on Windows; produces one installer for all binaries)
export KB_INNO_ISS="Inno/KrankyBearPDF.iss"

# Linux .desktop launcher metadata (package.sh installs a launcher for the GUI only).
export KB_DESKTOP_COMMENT="Cross-platform PDF utility, editor and viewer"
export KB_DESKTOP_CATEGORIES="Office;Utility;"

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
