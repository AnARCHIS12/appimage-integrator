# Changelog

All notable changes to Aurémi are documented here.

The project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.3.0] — 2026-10-07

### Added

- Validated AppStream metadata for Discover, GNOME Software, and other Linux software centers.
- A visible Aurémi application-menu entry using the official logo.
- A localized onboarding window explaining how to install an AppImage after setup.
- Software-center metadata for the project name, description, license, links, release, and supported architecture-independent features.

## [0.2.0] — 2026-10-07

### Added

- Double-clickable DEB packages for AMD64 and ARM64.
- Double-clickable RPM packages for AMD64 and ARM64.
- A universal graphical installer bundle with a dedicated desktop launcher, icon, and instructions.
- Yad and XMessage graphical-dialog fallbacks.

### Changed

- Made installation instructions distribution-specific and beginner-friendly.
- Ensured the graphical installer always produces visible success or error feedback.
- Reorganized and translated the project documentation into English.
- Added contribution, security, and issue-reporting documentation.

## [0.1.0] — 2026-10-07

### Added

- AppImage header validation without launching the target application.
- Metadata and embedded-icon extraction through `unsquashfs` when available.
- Atomic per-user AppImage integration.
- Application-menu entries and desktop shortcuts.
- MIME registration for double-click installation.
- Managed application listing and removal.
- Graphical one-click installers for AMD64 and ARM64.
- Statically linked distribution-independent binaries.
- Automated tagged releases through GitHub Actions.
- SHA-256 checksums for release assets.

[Unreleased]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/AnARCHIS12/appimage-integrator/releases/tag/v0.1.0
