.PHONY: all binaries installers bundles native-packages packages checksums checksums-only print-version test clean install-user install-system

VERSION := 0.3.4

all: checksums

print-version:
	@printf '%s\n' '$(VERSION)'

binaries: dist/appimage-integrator-linux-amd64 dist/appimage-integrator-linux-arm64

dist/appimage-integrator-linux-amd64: main.go go.mod assets/auremi-logo-512.png packaging/io.github.anarchis12.auremi.metainfo.xml
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $@ .

dist/appimage-integrator-linux-arm64: main.go go.mod assets/auremi-logo-512.png packaging/io.github.anarchis12.auremi.metainfo.xml
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $@ .

installers: binaries installer/main.go installer/assets/auremi-logo-512.png
	mkdir -p installer/payload
	cp dist/appimage-integrator-linux-amd64 installer/payload/appimage-integrator
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/Auremi-Installer-bin-linux-amd64 ./installer
	cp dist/appimage-integrator-linux-arm64 installer/payload/appimage-integrator
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/Auremi-Installer-bin-linux-arm64 ./installer
	$(RM) installer/payload/appimage-integrator

bundles: installers
	$(RM) -r dist/Auremi-Installer-linux-amd64 dist/Auremi-Installer-linux-arm64
	for arch in amd64 arm64; do \
		bundle="dist/Auremi-Installer-linux-$$arch"; \
		mkdir -p "$$bundle"; \
		cp "dist/Auremi-Installer-bin-linux-$$arch" "$$bundle/Auremi-Installer"; \
		cp packaging/Install-Auremi.desktop "$$bundle/Install Aurémi.desktop"; \
		cp packaging/INSTALL.txt "$$bundle/README.txt"; \
		cp assets/auremi-logo-512.png "$$bundle/auremi-logo.png"; \
		chmod 755 "$$bundle/Auremi-Installer" "$$bundle/Install Aurémi.desktop"; \
		tar -czf "dist/Auremi-Installer-linux-$$arch.tar.gz" -C dist "Auremi-Installer-linux-$$arch"; \
	done

native-packages: binaries
	./packaging/build-native-packages.sh $(VERSION)

packages: bundles native-packages

checksums: packages
	$(MAKE) checksums-only

checksums-only:
	cd dist && sha256sum \
		appimage-integrator-linux-amd64 \
		appimage-integrator-linux-arm64 \
		Auremi-Installer-linux-amd64.tar.gz \
		Auremi-Installer-linux-arm64.tar.gz \
		Auremi-$(VERSION)-linux-amd64.deb \
		Auremi-$(VERSION)-linux-arm64.deb \
		Auremi-$(VERSION)-linux-amd64.rpm \
		Auremi-$(VERSION)-linux-arm64.rpm \
		> SHA256SUMS

test: binaries
	mkdir -p installer/payload
	cp dist/appimage-integrator-linux-amd64 installer/payload/appimage-integrator
	go test ./...
	$(RM) installer/payload/appimage-integrator

install-user: all
	./dist/appimage-integrator-linux-amd64 setup --user

install-system: all
	sudo ./dist/appimage-integrator-linux-amd64 setup --system

clean:
	$(RM) -r dist installer/payload/appimage-integrator
