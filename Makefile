.PHONY: all binaries installers packages checksums test clean install-user install-system

all: checksums

binaries: dist/appimage-integrator-linux-amd64 dist/appimage-integrator-linux-arm64

dist/appimage-integrator-linux-amd64: main.go go.mod assets/auremi-logo-512.png
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $@ .

dist/appimage-integrator-linux-arm64: main.go go.mod assets/auremi-logo-512.png
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $@ .

installers: binaries installer/main.go installer/assets/auremi-logo-512.png
	mkdir -p installer/payload
	cp dist/appimage-integrator-linux-amd64 installer/payload/appimage-integrator
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/Auremi-Installer-linux-amd64 ./installer
	cp dist/appimage-integrator-linux-arm64 installer/payload/appimage-integrator
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/Auremi-Installer-linux-arm64 ./installer
	$(RM) installer/payload/appimage-integrator

packages: installers
	tar -czf dist/Auremi-Installer-linux-amd64.tar.gz -C dist Auremi-Installer-linux-amd64
	tar -czf dist/Auremi-Installer-linux-arm64.tar.gz -C dist Auremi-Installer-linux-arm64

checksums: packages
	cd dist && sha256sum appimage-integrator-linux-amd64 appimage-integrator-linux-arm64 Auremi-Installer-linux-amd64 Auremi-Installer-linux-arm64 Auremi-Installer-linux-amd64.tar.gz Auremi-Installer-linux-arm64.tar.gz > SHA256SUMS

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
