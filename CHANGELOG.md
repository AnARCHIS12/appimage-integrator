# Changelog

All notable changes to Aurémi are documented here.

The project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.3.4] — 2026-10-07

### Added

- Published official APT and DNF repositories through GitHub Pages.
- Added a dedicated GPG repository key and signed APT release metadata, DNF repository metadata, and RPM packages.
- Added AppStream repository catalogs and cached icons so software centers can identify Aurémi by its name, logo, version, and release notes.

### Changed

- Displayed the Aurémi version in the welcome dialog and kept its onboarding text in English.
- Kept English as the default AppStream description while retaining localized French summary metadata.

## [0.3.3] — 2026-10-07

### Added

- Extracted the precise application version from `X-AppImage-Version` or matching embedded AppStream metadata.
- Added a conservative filename fallback for versioned AppImages that omit structured version metadata.
- Stored the detected application version in managed metadata, propagated it to generated desktop entries, and displayed it in integration and list output.

## [0.3.2] — 2026-10-07

### Fixed

- Preserved safe embedded desktop-file IDs so Wayland compositors can associate running windows with their launchers and icons.
- Selected icons by the name declared in AppImage metadata instead of unrelated image size, including extensionless `.DirIcon` files and symlinks.
- Installed every available size of the selected icon and used absolute icon paths for reliable display across desktop environments.
- Removed obsolete icon variants when reintegrating or uninstalling an application and refreshed icon caches after native package removal.

### Security

- Added keyless Sigstore signatures and public verification bundles for every release executable, archive, native package, and checksum file.
- Made the release workflow verify every signature before publishing and pinned the Cosign installer action to an immutable commit.

## [0.3.1] — 2026-10-07

### Changed

- Made English the default language for all CLI help, errors, notifications, generated launchers, and integration output.
- Kept French only as an explicit locale-aware translation for supported graphical metadata and onboarding.
- Standardized generated MIME descriptions and AppImage handler metadata in English.

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

[Unreleased]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.3.4...HEAD
[0.3.4]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.3.3...v0.3.4
[0.3.3]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/AnARCHIS12/appimage-integrator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/AnARCHIS12/appimage-integrator/releases/tag/v0.1.0
