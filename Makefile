VERSION = 1.0.0
BINARY  = gotorrent
CMD     = ./cmd/gotorrent

LDFLAGS = -ldflags="-s -w"

.PHONY: run build clean deps icon package-mac package-linux package-windows install-linux installer-dmg installer-pkg installer-nsis installer-deb

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
	rm -f $(BINARY)-*.dmg $(BINARY)-*.pkg $(BINARY)-*.deb
	rm -f GoTorrent-*-Setup.exe
	rm -rf dmg_staging pkg_root deb_staging

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

## installer-dmg: create a macOS .dmg drag-to-Applications disk image
installer-dmg: package-mac
	@echo "Creating DMG..."
	mkdir -p dmg_staging
	cp -R GoTorrent.app dmg_staging/
	ln -sf /Applications dmg_staging/Applications
	hdiutil create -volname "GoTorrent" -srcfolder dmg_staging \
		-ov -format UDZO "$(BINARY)-$(VERSION)-mac.dmg"
	rm -rf dmg_staging

## installer-pkg: create a macOS .pkg system installer
installer-pkg: package-mac
	@echo "Creating PKG..."
	mkdir -p pkg_root/Applications
	cp -R GoTorrent.app pkg_root/Applications/
	pkgbuild --root pkg_root \
		--identifier com.gotorrent.app \
		--version $(VERSION) \
		--install-location / \
		"$(BINARY)-$(VERSION)-mac.pkg"
	rm -rf pkg_root

## installer-nsis: create a Windows NSIS Setup.exe (requires NSIS installed)
installer-nsis: package-windows
	makensis /DVERSION=$(VERSION) installer/windows/gotorrent.nsi

## installer-deb: create a Linux .deb package
installer-deb: package-linux
	@echo "Creating DEB..."
	mkdir -p deb_staging/DEBIAN
	mkdir -p deb_staging/usr/local/bin
	mkdir -p deb_staging/usr/share/applications
	mkdir -p deb_staging/usr/share/icons/hicolor/256x256/apps
	cp $(BINARY) deb_staging/usr/local/bin/
	cp assets/gotorrent.desktop deb_staging/usr/share/applications/
	cp assets/icon.png deb_staging/usr/share/icons/hicolor/256x256/apps/gotorrent.png
	printf 'Package: gotorrent\nVersion: $(VERSION)\nSection: net\nPriority: optional\nArchitecture: amd64\nMaintainer: Tarun Vishwakarma\nDescription: GoTorrent - A BitTorrent client written in Go\n' \
		> deb_staging/DEBIAN/control
	dpkg-deb --build deb_staging "$(BINARY)-$(VERSION)-linux-amd64.deb"
	rm -rf deb_staging
