#!/usr/bin/env bash
# Bump the app version across all three KrankyBear PDF binaries and packaging metadata.
# Canonical default lives in build-config.sh (KB_VERSION_DEFAULT); this script keeps every
# consumer file in lock-step so the CLI (pdfutil), GUI (pdfgui) and viewer (pdfviewer) ship
# with one version.
#
# Usage: ./setver.sh 0.3.1
#
# macOS/BSD sed: sed -i '' …   Linux GNU sed: sed -i …

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

# shellcheck disable=SC1091
source "${ROOT}/build-config.sh"

if [[ "$(uname -s)" == Darwin ]]; then
  sed_i() { sed -i '' "$@"; }
else
  sed_i() { sed -i "$@"; }
fi

# Pull the current appVersion out of a Go file, for the interactive prompt.
go_ver() { grep -E '^[[:space:]]*appVersion[[:space:]]*=' "$1" 2>/dev/null | head -1 | sed -E 's/.*"([^"]+)".*/\1/'; }

if [[ $# -ge 1 ]]; then
  ver=$1
else
  echo "Enter a version number."
  echo "    packaging default (build-config.sh): ${KB_VERSION_DEFAULT}"
  echo "    pdfutil (main.go):                   $(go_ver main.go)"
  echo "    pdfgui  (pdfgui/main.go):            $(go_ver pdfgui/main.go)"
  echo "    pdfviewer (pdfviewer/main.go):       $(go_ver pdfviewer/main.go)"
  read -r ver
  if [[ -z "${ver}" ]]; then
    echo "No version change; exiting."
    exit 0
  fi
fi

echo "Setting version: ${ver}"
echo ""

echo "build-config.sh (KB_VERSION_DEFAULT)"
sed_i "s/^export KB_VERSION_DEFAULT=.*/export KB_VERSION_DEFAULT=\"${ver}\"/" ./build-config.sh

# appVersion = "..." in each app's main.go
for gofile in main.go pdfgui/main.go pdfviewer/main.go; do
  if [[ -f "$gofile" ]]; then
    echo "${gofile} (appVersion)"
    sed_i "s/^\([[:space:]]*appVersion[[:space:]]*=[[:space:]]*\"\)[^\"]*\(\".*\)/\1${ver}\2/" "$gofile"
  fi
done

# FyneApp.toml files (root = CLI/GUI metadata, pdfviewer has its own)
for toml in FyneApp.toml pdfviewer/FyneApp.toml; do
  if [[ -f "$toml" ]]; then
    echo "${toml} (Version)"
    sed_i "s/^[[:space:]]*Version = \".*\"/  Version = \"${ver}\"/" "$toml"
  fi
done

# Windows Inno Setup
if [[ -f "./${KB_INNO_ISS}" ]]; then
  echo "${KB_INNO_ISS} (MyAppVersion)"
  sed_i "s/#define MyAppVersion \".*\"/#define MyAppVersion \"${ver}\"/" "./${KB_INNO_ISS}"
fi

# Windows version resources (go-winres) — GUI is the only binary with embedded version info
if [[ -f "pdfgui/winres/winres.json" ]]; then
  echo "pdfgui/winres/winres.json"
  sed_i "s/\"file_version\": \"[^\"]*\"/\"file_version\": \"${ver}\"/" pdfgui/winres/winres.json
  sed_i "s/\"product_version\": \"[^\"]*\"/\"product_version\": \"${ver}\"/" pdfgui/winres/winres.json
  sed_i "s/\"FileVersion\": \"[^\"]*\"/\"FileVersion\": \"${ver}\"/" pdfgui/winres/winres.json
  sed_i "s/\"ProductVersion\": \"[^\"]*\"/\"ProductVersion\": \"${ver}\"/" pdfgui/winres/winres.json
fi

# macOS Info.plist sample (CFBundleShortVersionString only)
if [[ -f "Info-plist.txt" ]]; then
  echo "Info-plist.txt (CFBundleShortVersionString)"
  sed_i "/<key>CFBundleShortVersionString<\\/key>/,/<string>/ s/<string>[^<]*<\\/string>/<string>${ver}<\\/string>/" Info-plist.txt
fi

# ReleaseNotes.txt: prepend a header for this version if not already present
echo "ReleaseNotes.txt"
if ! head -5 ReleaseNotes.txt | grep -q "Version ${ver} "; then
  echo "  adding 'Version ${ver}' header"
  {
    echo "Version ${ver} - $(date '+%Y-%m-%d')"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    cat ReleaseNotes.txt
  } > ReleaseNotes.txt.new
  mv ReleaseNotes.txt.new ReleaseNotes.txt
fi
# keep the copy the apps read at runtime in sync
cp ReleaseNotes.txt assets/ReleaseNotes.txt 2>/dev/null || true

echo ""
echo "Version updated to: ${ver}"
echo ""
echo "Files updated:"
echo "  - build-config.sh (KB_VERSION_DEFAULT)"
echo "  - main.go, pdfgui/main.go, pdfviewer/main.go (appVersion)"
echo "  - FyneApp.toml, pdfviewer/FyneApp.toml"
echo "  - ${KB_INNO_ISS}"
echo "  - pdfgui/winres/winres.json"
echo "  - Info-plist.txt"
echo "  - ReleaseNotes.txt (+ assets/ReleaseNotes.txt)"
echo ""
echo "Don't forget to flesh out ReleaseNotes.txt with the actual changes!"

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
