#! /bin/sh

GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/pdfutil-MacOSARM64
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/pdfutil-MacOSAMD64
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="x86_64-w64-mingw32-gcc" go build -ldflags="-w -s" -o bin/pdfutil.exe #  -H windowsgui -r img2icons.rc" -o bin/pdfutil.exe
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/pdfutil-LinuxAMD64
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/pdfutil-LinuxARM64
# GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/jiggler-LinuxAMD64
setIcon.sh KrankyBearBeret.png bin/pdfutil-MacOSARM64
setIcon.sh KrankyBearBeret.png bin/pdfutil-MacOSAMD64
setIcon.sh KrankyBearBeret.png bin/pdfutil.exe
setIcon.sh KrankyBearBeret.png bin/pdfutil-LinuxAMD64
setIcon.sh KrankyBearBeret.png bin/pdfutil-LinuxARM64