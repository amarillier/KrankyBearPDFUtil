#! /bin/sh

if [ $# -ge 1 ]
then
    ver=$1
else
    echo "Enter a version number"
    cur=$(cat main.go | grep -i "Version =")
    echo "    current: $cur"
    read ver
    if [ -z "$ver" ]
    then
        echo "Enter a version!"
        echo "No version change detected, continuing to allow compile to continue"
        exit
    else
        echo "Version: $ver"
        # exit
    fi
fi

echo "version: $ver"
echo "main.go"
sed -i '' "s/Version = \".*\"/Version = \"$ver\"/" main.go

echo "FyneApp.toml"
sed -i '' "s/Version = \".*\"/Version = \"$ver\"/" FyneApp.toml

echo "Inno Setup Inno/KrankyBearTimer.iss"
sed -i '' "s/MyAppVersion \".*\"/MyAppVersion \"$ver\"/" ./Inno/KrankyBearTimer.iss

echo "Inno Setup winres/winres.json"
sed -i '' "s/file_version\":.*/file_version\": \"$ver\",/" ./winres/winres.json
sed -i '' "s/product_version\":.*/product_version\": \"$ver\"/" ./winres/winres.json
sed -i '' "s/FileVersion\":.*/FileVersion\": \"$ver\",/" ./winres/winres.json
sed -i '' "s/ProductVersion\":.*/ProductVersion\": \"$ver\",/" ./winres/winres.json

echo "No Info.plist updates"
# sed -i '' "s/<string>v .*<\/string>/<string>v $ver<\/string>/" ./KrankyBearTimer.app/Contents/Info.plist

echo "mkAllZip.sh"
sed -i '' "s/version=\".*\"/version=\"$ver\"/" mkAllZip.sh
sed -i '' "s/\/.*{flag=1}\//\/$ver\/{flag=1}\//" mkAllZip.sh

echo "Update license.txt and ReleaseNotes.txt"
cp license.txt Resources
cp ReleaseNotes.txt Resources
cp license.txt KrankyBearTimer.app/Contents/Resources
cp ReleaseNotes.txt KrankyBearTimer.app/Contents/Resources