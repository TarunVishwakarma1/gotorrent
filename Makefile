VERSION = 1.0.0
BINARY  = gotorrent
CMD     = ./cmd/gotorrent

LDFLAGS = -ldflags="-s -w"

.PHONY: run build clean deps icon package-mac package-linux package-windows install-linux

## deps: download and tidy Go modules
deps:
	go mod tidy

## icon: convert SVG to PNG using rsvg-convert or Inkscape (required before packaging)
icon:
	@if command -v rsvg-convert >/dev/null 2>&1; then \
		rsvg-convert -w 256 -h 256 assets/icon.svg -o assets/icon.png; \
	elif command -v inkscape >/dev/null 2>&1; then \
		inkscape assets/icon.svg --export-width=256 --export-height=256 --export-filename=assets/icon.png; \
	else \
		echo "ERROR: install rsvg-convert (brew install librsvg) or Inkscape to generate the PNG icon"; \
		exit 1; \
	fi

## run: run the application in development mode
run:
	go run $(CMD)

## build: build a stripped binary for the current platform
build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

## clean: remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf GoTorrent.app GoTorrent.exe
	rm -f $(BINARY)-*.tar.gz $(BINARY)-*.zip

## package-mac: build a macOS .app bundle
package-mac: icon
	fyne package -os darwin -icon $(CURDIR)/assets/icon.png -name GoTorrent --app-id com.gotorrent.app -src $(CMD)
	tar -czf $(BINARY)-$(VERSION)-mac.tar.gz GoTorrent.app

## package-linux: build a Linux executable package
package-linux: icon
	fyne package -os linux -icon assets/icon.png -name GoTorrent --app-id com.gotorrent.app -src $(CMD)
	tar -czf $(BINARY)-$(VERSION)-linux.tar.gz $(BINARY)

## package-windows: build a Windows executable
package-windows: icon
	fyne package -os windows -icon assets/icon.png -name GoTorrent --app-id com.gotorrent.app -src $(CMD)
	zip $(BINARY)-$(VERSION)-windows.zip GoTorrent.exe

## install-linux: install the binary and .desktop file on Linux
install-linux: package-linux
	sudo cp $(BINARY) /usr/local/bin/
	@mkdir -p ~/.local/share/applications ~/.local/share/icons/hicolor/256x256/apps
	cp assets/icon.png ~/.local/share/icons/hicolor/256x256/apps/gotorrent.png
	cp assets/gotorrent.desktop ~/.local/share/applications/
	xdg-mime default gotorrent.desktop application/x-bittorrent
	update-desktop-database ~/.local/share/applications/ 2>/dev/null || true

## vet: run go vet
vet:
	go vet ./...

## test: run unit tests
test:
	go test ./internal/... -v
