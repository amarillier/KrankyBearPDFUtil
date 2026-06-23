#!/bin/sh
# Version updater for KrankyBearPDF (CLI and GUI)
# Updates version numbers in all relevant files

if [ $# -ge 1 ]
then
    ver=$1
else
    echo "Enter a version number"
    cur=$(cat main.go | grep -i "Version =" | head -1)
    echo "    current CLI: $cur"
    cur_gui=$(cat gui/main.go | grep -i "appVersion =" | head -1)
    echo "    current GUI: $cur_gui"
    read ver
    if [ -z "$ver" ]
    then
        echo "Enter a version!"
        echo "No version change detected, exiting"
        exit 1
    else
        echo "Version: $ver"
    fi
fi

echo "Setting version: $ver"
echo ""

# CLI main.go
echo "CLI main.go"
sed -i '' "s/Version = \".*\"/Version = \"$ver\"/" main.go

# GUI main.go
echo "GUI main.go"
sed -i '' "s/appVersion = \".*\"/appVersion = \"$ver\"/" gui/main.go

# CLI FyneApp.toml (if exists, for CLI Fyne builds)
if [ -f "FyneApp.toml" ]; then
    echo "FyneApp.toml"
    sed -i '' "s/Version = \".*\"/Version = \"$ver\"/" FyneApp.toml
fi

# GUI FyneApp.toml
if [ -f "gui/FyneApp.toml" ]; then
    echo "GUI FyneApp.toml"
    sed -i '' "s/Version = \".*\"/Version = \"$ver\"/" gui/FyneApp.toml
fi

# GUI Inno Setup
echo "GUI Inno Setup gui/Inno/KrankyBearPDFGui.iss"
sed -i '' "s/MyAppVersion \".*\"/MyAppVersion \"$ver\"/" ./gui/Inno/KrankyBearPDFGui.iss

# GUI winres.json
echo "GUI winres/winres.json"
sed -i '' "s/\"file_version\": \".*\"/\"file_version\": \"$ver\"/" ./gui/winres/winres.json
sed -i '' "s/\"product_version\": \".*\"/\"product_version\": \"$ver\"/" ./gui/winres/winres.json
sed -i '' "s/\"FileVersion\": \".*\"/\"FileVersion\": \"$ver\"/" ./gui/winres/winres.json
sed -i '' "s/\"ProductVersion\": \".*\"/\"ProductVersion\": \"$ver\"/" ./gui/winres/winres.json

# Update README if version is mentioned
if [ -f "README.md" ] && grep -q "Version:" README.md; then
    echo "README.md (version reference)"
    sed -i '' "s/Version: .*/Version: $ver/" README.md
fi

# Update ReleaseNotes.txt header
echo "ReleaseNotes.txt"
# Add new version header if it doesn't exist
if ! head -5 ReleaseNotes.txt | grep -q "$ver"; then
    echo "Adding version $ver header to ReleaseNotes.txt"
    echo "Version $ver - $(date '+%Y-%m-%d')" > ReleaseNotes.txt.new
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >> ReleaseNotes.txt.new
    echo "" >> ReleaseNotes.txt.new
    cat ReleaseNotes.txt >> ReleaseNotes.txt.new
    mv ReleaseNotes.txt.new ReleaseNotes.txt
fi

echo ""
echo "Version updated to: $ver"
echo ""
echo "Files updated:"
echo "  - main.go (CLI)"
echo "  - gui/main.go (GUI)"
echo "  - gui/Inno/KrankyBearPDFGui.iss"
echo "  - gui/winres/winres.json"
echo "  - ReleaseNotes.txt"
echo ""
echo "Don't forget to update ReleaseNotes.txt with actual changes!"

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
