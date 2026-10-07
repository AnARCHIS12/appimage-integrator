<div align="center">

<img src="assets/auremi-logo.png" alt="Aurémi logo" width="180">

# Aurémi

### Install any AppImage with a double-click.

No terminal. No administrator password. No deleted downloads.

[![Latest release](https://img.shields.io/github/v/release/AnARCHIS12/appimage-integrator?style=for-the-badge&color=dc2626)](https://github.com/AnARCHIS12/appimage-integrator/releases/latest)
[![Release build](https://img.shields.io/github/actions/workflow/status/AnARCHIS12/appimage-integrator/release.yml?style=for-the-badge&label=build)](https://github.com/AnARCHIS12/appimage-integrator/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-black?style=for-the-badge)](LICENSE)
[![Linux](https://img.shields.io/badge/Linux-AMD64%20%7C%20ARM64-black?style=for-the-badge&logo=linux&logoColor=white)](#linux-compatibility)

[Download](https://github.com/AnARCHIS12/appimage-integrator/releases/latest) ·
[How it works](#how-it-works) ·
[Security](#security) ·
[Build from source](#build-from-source) ·
[Report a bug](https://github.com/AnARCHIS12/appimage-integrator/issues/new?template=bug_report.yml)

</div>

---

## Why Aurémi?

AppImages are portable, but integrating them into a Linux desktop is still unnecessarily manual. Users often need to make files executable, find a permanent location, extract an icon, and create a `.desktop` launcher themselves.

Aurémi handles that workflow automatically while keeping the original download untouched.

- **One-time setup:** install Aurémi once, then double-click AppImages.
- **Desktop integration:** application-menu entry, icon, and desktop shortcut.
- **No AppImage execution during inspection:** metadata is read without launching the application.
- **No root access for daily use:** applications are installed inside the current user's home directory.
- **Distribution-independent:** static AMD64 and ARM64 binaries, with no GTK or Qt runtime dependency.
- **Reversible:** list and remove managed applications without touching the original download.

## Quick start

### Graphical installation

1. Open the [latest release](https://github.com/AnARCHIS12/appimage-integrator/releases/latest).
2. Download the archive for your computer:
   - **Most Intel/AMD PCs:** `Auremi-Installer-linux-amd64.tar.gz`
   - **ARM64 systems:** `Auremi-Installer-linux-arm64.tar.gz`
3. Extract the archive.
4. Double-click the `Auremi-Installer` executable.
5. Select **Install**.

The installer uses KDialog on KDE Plasma, Zenity on GNOME and compatible desktops, or a desktop notification as a fallback. It installs only for the current account and does not request an administrator password.

### Command-line installation

For the current user, without `sudo`:

```bash
chmod +x appimage-integrator-linux-amd64
./appimage-integrator-linux-amd64 setup --user
```

For every user on the computer:

```bash
chmod +x appimage-integrator-linux-amd64
sudo ./appimage-integrator-linux-amd64 setup --system
```

The system-wide setup installs only the handler globally. AppImages are still integrated separately inside each user's home directory.

## How it works

```text
Double-click an AppImage
          │
          ▼
Validate the ELF/AppImage header
          │
          ▼
Read embedded desktop metadata and icon
          │
          ▼
Copy the AppImage into the user's data directory
          │
          ▼
Create the menu entry, icon, and desktop shortcut
```

During integration, Aurémi:

1. rejects symbolic links and invalid AppImage files;
2. reads metadata without launching the AppImage;
3. copies the file atomically to a managed directory;
4. preserves the original downloaded file;
5. registers a Freedesktop-compatible launcher and icon;
6. refreshes available desktop caches when their tools are installed.

## Usage

After setup, double-click any `.AppImage` file. The integrated application will appear in the application menu and on the desktop.

Manual integration:

```bash
appimage-integrator integrate MyApplication.AppImage
```

Integrate without creating a desktop shortcut:

```bash
appimage-integrator integrate --no-desktop MyApplication.AppImage
```

List managed applications:

```bash
appimage-integrator list
```

Remove a managed application:

```bash
appimage-integrator remove appimage-my-application
```

Removal deletes only the managed copy, icon, and launchers. The original AppImage remains untouched.

## Linux compatibility

Aurémi follows Freedesktop standards and is designed for:

- KDE Plasma
- GNOME
- Cinnamon
- XFCE
- MATE
- LXQt
- other Freedesktop-compatible environments

The release provides statically linked binaries for:

- **AMD64:** 64-bit Intel and AMD computers
- **ARM64:** 64-bit ARM computers

The binaries do not depend on glibc, GTK, or Qt and can run on both glibc- and musl-based distributions.

### Optional metadata support

The optional `unsquashfs` command, supplied by `squashfs-tools`, lets Aurémi extract the real application name and embedded icon. Without it, integration still works with the filename and a generic icon.

<details>
<summary>Install squashfs-tools</summary>

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

</details>

## Security

Aurémi is an installer and desktop integrator, not a malware scanner or code-signing authority.

It reduces installation-time risk by:

- refusing symbolic links and non-AppImage files;
- never launching an AppImage while inspecting or installing it;
- constraining managed paths to the user's application directories;
- writing files atomically;
- avoiding `sudo` for per-user integration;
- preserving the original download.

These checks do **not** prove that an AppImage is trustworthy. Download applications from their official website or repository and verify published checksums whenever possible.

Please report security issues using the instructions in [SECURITY.md](SECURITY.md), not through a public issue.

## Verify a release

Every release includes a `SHA256SUMS` file:

```bash
sha256sum -c SHA256SUMS
```

## Managed file locations

<details>
<summary>Show paths</summary>

Per-user Aurémi installation:

- Binary: `~/.local/bin/appimage-integrator`
- File handler: `~/.local/share/applications/appimage-integrator-handler.desktop`
- MIME definition: `~/.local/share/mime/packages/appimage-integrator.xml`

Integrated applications:

- AppImages and metadata: `~/.local/share/appimage-integrator/apps/`
- Launchers: `~/.local/share/applications/`
- Icons: `~/.local/share/icons/hicolor/`
- Shortcuts: desktop directory defined by `XDG_DESKTOP_DIR`

Aurémi respects `XDG_DATA_HOME` and `XDG_CONFIG_HOME`.

</details>

## Build from source

Requirements:

- Go 1.23 or newer
- GNU Make
- `tar` and `sha256sum` for release packages

Build and test:

```bash
make test
make
```

Generated binaries, installer archives, and checksums are written to `dist/`.

The automated GitHub workflow runs the same test and build process whenever a `v*` tag is pushed.

## Project documentation

- [Changelog](CHANGELOG.md)
- [Contributing guide](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Release notes](RELEASE_NOTES.md)

## Contributing

Bug reports, compatibility feedback, documentation improvements, and code contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

## License

Aurémi is available under the [MIT License](LICENSE).

---

<div align="center">

Built for a simpler Linux desktop.

</div>
