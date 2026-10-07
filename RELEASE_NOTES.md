# Aurémi 0.3.2

This release fixes incorrect or missing application icons and improves launcher matching on Wayland and other Freedesktop-compatible desktops. Release downloads are now cryptographically signed with Sigstore.

## Highlights

- Preserves the embedded desktop-file ID so Wayland compositors can associate running windows with the correct launcher and icon.
- Selects the icon declared by the AppImage instead of an unrelated large image from the bundle.
- Recognizes extensionless `.DirIcon` files, symlinked icons, SVG, PNG, and XPM assets.
- Installs all useful resolutions and uses a reliable absolute icon path across desktop environments.
- Cleans obsolete icon variants when an application is reintegrated or removed.
- Uses the installed Aurémi logo directly in generated and native-package launchers.
- Signs every executable, installer archive, DEB, RPM, and checksum file with keyless Sigstore signing.
- Publishes a `.sigstore.json` verification bundle beside every signed download.

## Downloads

- **Debian, Ubuntu, Linux Mint:** choose the `.deb` file for your processor.
- **Fedora, openSUSE:** choose the `.rpm` file for your processor.
- **Other distributions:** choose the universal `Auremi-Installer` archive for your processor.

Most computers use AMD64. ARM64 packages are also provided.

After installation, open **Aurémi** from the application menu for a short getting-started guide, or simply double-click an AppImage.

Use `SHA256SUMS` to verify downloaded assets.

For cryptographic provenance verification, download the matching `.sigstore.json` file and follow the Cosign command in the README. The signature proves that the artifact was produced by Aurémi's tagged GitHub release workflow.
