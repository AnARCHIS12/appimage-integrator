# Aurémi — AppImage Integrator

[![Latest release](https://img.shields.io/github/v/release/AnARCHIS12/appimage-integrator?style=flat-square&color=dc2626)](https://github.com/AnARCHIS12/appimage-integrator/releases/latest)
[![Release build](https://img.shields.io/github/actions/workflow/status/AnARCHIS12/appimage-integrator/release.yml?style=flat-square&label=build)](https://github.com/AnARCHIS12/appimage-integrator/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-black?style=flat-square)](LICENSE)
![Linux AMD64 and ARM64](https://img.shields.io/badge/Linux-AMD64%20%7C%20ARM64-black?style=flat-square&logo=linux&logoColor=white)

**Aurémi** turns a double-click on an AppImage into a clean desktop installation. It stays out of the way during normal use; only the initial setup displays a small confirmation dialog.

<p align="center">
  <img src="assets/auremi-logo.png" alt="Aurémi logo" width="220">
</p>

It follows Linux/Freedesktop standards and works with KDE Plasma, GNOME, Cinnamon, XFCE, MATE, LXQt, and most other Linux desktop environments.

## What it does

When you double-click an AppImage, Aurémi:

1. verifies that the file is an actual AppImage;
2. reads its name and embedded icon without launching the application;
3. stores an executable copy in your home directory;
4. creates an application-menu entry;
5. creates a desktop shortcut;
6. displays a notification when the installation is complete.

The original downloaded file is never deleted. The AppImage itself is never launched during installation.

## Double-click installation — easiest method

1. Download the archive for your computer from the [latest release](https://github.com/AnARCHIS12/appimage-integrator/releases/latest):
   - `Auremi-Installer-linux-amd64.tar.gz` for most 64-bit Intel and AMD computers;
   - `Auremi-Installer-linux-arm64.tar.gz` for 64-bit ARM computers.
2. Extract the archive with your file manager.
3. Double-click `Auremi-Installer-linux-amd64` or `Auremi-Installer-linux-arm64`.
4. Select **Install**.

The installer uses a native KDialog window on KDE, Zenity on GNOME and compatible desktops, or a desktop notification as a fallback. It installs Aurémi only for your account and does not request an administrator password.

## Command-line installation

### Current user only — recommended

This method does not require `sudo`:

```bash
chmod +x appimage-integrator-linux-amd64
./appimage-integrator-linux-amd64 setup --user
```

The manager is copied to `~/.local/bin/` and registered only for your account.

### All users

This command writes to `/usr/local/bin`, `/usr/share`, and `/etc/xdg`, so administrator privileges are required:

```bash
chmod +x appimage-integrator-linux-amd64
sudo ./appimage-integrator-linux-amd64 setup --system
```

After this one-time setup, every user can integrate AppImages without `sudo`. Each user's applications remain isolated in their own home directory.

## Usage

After installing Aurémi, double-click any file with the `.AppImage` extension.

The installed application will appear:

- in the application menu;
- on the desktop;
- under `~/.local/share/appimage-integrator/apps/`.

The first click installs the AppImage. Launch it afterward from the application menu or its desktop shortcut.

### Manual integration

```bash
appimage-integrator integrate MyApplication.AppImage
```

Without a desktop shortcut:

```bash
appimage-integrator integrate --no-desktop MyApplication.AppImage
```

### List managed applications

```bash
appimage-integrator list
```

### Remove an application

First find its identifier with `list`, then run:

```bash
appimage-integrator remove appimage-my-application
```

This removes only the managed copy, icon, and launchers. The originally downloaded AppImage remains untouched.

## Linux compatibility

The supplied binaries are self-contained and do not depend on GTK, Qt, or a particular distribution:

- `linux-amd64` supports 64-bit Intel and AMD computers;
- `linux-arm64` supports 64-bit ARM computers.

The binaries are statically linked and work on distributions based on either glibc or musl.

Verify downloaded release files with:

```bash
sha256sum -c SHA256SUMS
```

The optional `unsquashfs` command, provided by `squashfs-tools`, lets Aurémi extract the actual application name and embedded icon. Without it, integration still works using the filename and a generic icon.

Optional `squashfs-tools` installation:

```bash
# Debian, Ubuntu, Linux Mint
sudo apt install squashfs-tools

# Fedora
sudo dnf install squashfs-tools

# Arch Linux, Manjaro
sudo pacman -S squashfs-tools

# openSUSE
sudo zypper install squashfs

# Alpine Linux
sudo apk add squashfs-tools
```

The utilities `update-mime-database`, `update-desktop-database`, `gtk-update-icon-cache`, `kbuildsycoca6`, and `notify-send` are used when available, but none is mandatory.

## Security

Aurémi reduces installation-time risk:

- it rejects symbolic links and files that are not valid AppImages;
- it never launches an AppImage while inspecting or installing it;
- it uses `unsquashfs` to read metadata when available;
- it writes managed files atomically;
- it never requires `sudo` to integrate an application for one user;
- it never deletes the original download.

These checks do not prove that an AppImage itself is trustworthy. Always download applications from their official website or repository and verify published SHA-256 checksums when available.

## File locations

Per-user manager installation:

- binary: `~/.local/bin/appimage-integrator`
- file-manager handler: `~/.local/share/applications/appimage-integrator-handler.desktop`
- MIME definition: `~/.local/share/mime/packages/appimage-integrator.xml`

Integrated applications:

- AppImages and metadata: `~/.local/share/appimage-integrator/apps/`
- launchers: `~/.local/share/applications/`
- icons: `~/.local/share/icons/hicolor/`
- shortcuts: the desktop directory defined by `XDG_DESKTOP_DIR`

The standard `XDG_DATA_HOME` and `XDG_CONFIG_HOME` environment variables are respected.

## Building from source

Go 1.23 or newer is required:

```bash
make test
make
```

Static AMD64 and ARM64 binaries and installer archives are created under `dist/`.

To build the core binary for another architecture:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/appimage-integrator-linux-arm64 .
```

## Known limitations

- Type 1 and Type 2 AppImages are recognized, but automatic icon extraction requires `unsquashfs` and readable SquashFS content.
- Some file managers ask which application should open an AppImage the first time. Select **Install an AppImage with Aurémi** and enable the option to remember your choice.
- A system-wide setup makes Aurémi available to everyone, but every AppImage is deliberately installed within the account of the user who clicked it. This avoids requesting an administrator password for every application.

## License

MIT — see [LICENSE](LICENSE).
