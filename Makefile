.DEFAULT_GOAL := hello
.ONESHELL:

hello:
	echo "Hello KrankyBear!"
	echo ""
	echo "Common targets:"
	echo "  make all             - Build for all platforms (Win, Linux, macOS Intel & ARM) to bin/"
	echo "  make build           - Build for the current system"
	echo "  make clean           - Remove compiled files from bin/*"
	echo ""
	echo "Development:"
	echo "  make fmt             - Format the code"
	echo "  make lint            - Run golint"
	echo "  make vet             - Vet the code"
	echo "  make tidy            - Tidy and vendor dependencies"
	echo "  make run             - Run the main.go"
	echo "  make doc             - Generate docs based on func names"
	echo ""
	echo "Platform-specific targets:"
	echo "  make macamd64        - Build macOS Intel (with installer)"
	echo "  make macarm64        - Build macOS ARM (with installer)"
	echo "  make winamd64        - Build Windows AMD64 (with installer)"
	echo "  make buildsupported  - Build for all supported systems with installers"
.PHONY:hello

fmt:
	go fmt ./...
.PHONY:fmt

lint:
	~/go/bin/golint ./...
.PHONY:lint

tidy:
	go mod tidy
	go mod vendor
	go mod verify
.PHONY:tidy

vet: 
	fmt
	go vet ...
.PHONY:vet

run:
	go run .
.PHONY:run

# Supported cross compile GOOS and GOARCH https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
build:
	./setver.sh
	go build -ldflags="-w -s" -o pdfutil .
	./setIcon.sh Resources/Images/KrankyBearBeret.png pdfutil
.PHONY:build

ios:
	GOOS=ios CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/ios/
	./setIcon.sh Resources/Images/KrankyBearBeret.png bin/ios/pdfutil
.PHONY:ios

linuxamd64:
	echo "This doesn't work right now on Mac ARM or Win AMD64 - no action"
 	# GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/LinuxAMD64/
	# ./setIcon.sh Resources/Images/KrankyBearBeret.png bin/LinuxAMD64/pdfutil
.PHONY:linuxamd64

linuxarm64:
	echo "This doesn't work right now on Mac ARM or Win AMD64 - no action"
 	# GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/LinuxARM64/
	# ./setIcon.sh Resources/Images/KrankyBearBeret.png bin/LinuxAMD64/pdfutil
.PHONY:linuxarm64

macamd64:
	# GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/MacOSAMD64/
	# ./setIcon.sh KrankyBear.png bin/MacOSAMD64/KrankyBearTimer
	./mkicns.sh Resources/Images/KrankyBearFedoraRed.png
	./dmgbuildIntel.sh
.PHONY:macamd64

macarm64:
	# GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/MacOSARM64/
	# ./setIcon.sh KrankyBear.png bin/MacOSARM64/KrankyBearTimer
	./mkicns.sh Resources/Images/KrankyBearFedoraRed.png
	./dmgbuildARM.sh
.PHONY:macarm64

winamd64:
	go-winres make
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="x86_64-w64-mingw32-gcc" go build -ldflags="-w -s -H windowsgui -r KrankyBearTimer.rc" -o bin/WinAMD64/
.PHONY:winamd64

winarm64:
	echo "This doesn't work right now on Mac ARM or Win AMD64 - no action"
	# go-winres make
	# GOOS=windows GOARCH=arm64 CGO_ENABLED=1 CC="x86_64-w64-mingw32-gcc" go build -ldflags="-w -s -H windowsgui -r KrankyBearTimer.rc" -o bin/WinARM64/
.PHONY:winarm64


all:
	@echo "Building KrankyBear PDFUtil for all platforms..."
	@echo ""
	@mkdir -p bin
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-windows-amd64.exe
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-linux-amd64
	@echo "Building for macOS Intel (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o bin/pdfutil-darwin-amd64
	@echo "Building for macOS Apple Silicon (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o bin/pdfutil-darwin-arm64
	@echo ""
	@echo "Build complete! Binaries are in the 'bin/' directory:"
	@ls -lh bin/
	@cp bin/pdfutil-darwin-arm64 ./pdfutil
.PHONY:all

buildall: linuxamd64 linuxarm64 macamd64 macarm64 winamd64 winarm64
.PHONY:buildall

buildsupported supported: macamd64 macarm64 winamd64
.PHONY:buildsupported

dmg: 
	./dmgbuildIntel.sh
	./dmgbuildARM.sh
.PHONY:dmg

clean:
	@echo "Cleaning build artifacts..."
	rm -f bin/*
	rm -f installers/*.dmg
	rm -f installers/*.exe
	@echo "Clean complete!"
.PHONY:clean

doc:
	grep -e "^// .* \.\.\. .*" -e "^.. .* \.\.\." -e "^func .*" *.go | sed 's/ {//g' | tee doc.md
.PHONY:doc

