# Aurémi 0.3.4

This release introduces official signed package repositories so Linux software centers can identify Aurémi and trust its publication source.

## Highlights

- Publishes official APT and DNF repositories for AMD64 and ARM64 through GitHub Pages.
- Signs APT release metadata, DNF repository metadata, and RPM packages with a dedicated GPG publication key.
- Provides AppStream catalogs and cached icons to software centers.
- Identifies the application as Aurémi, with its official logo, precise version, description, and release history.
- Displays the installed Aurémi version in the welcome dialog and keeps its onboarding text in English.
- Keeps the precise AppImage version detection, corrected icons, and Wayland launcher matching from earlier releases.
- Signs every executable, installer archive, DEB, RPM, and checksum file with keyless Sigstore signing.
- Publishes a `.sigstore.json` verification bundle beside every signed download.

## Downloads

- **Debian, Ubuntu, Linux Mint:** choose the `.deb` file for your processor.
- **Fedora, openSUSE:** choose the `.rpm` file for your processor.
- **Other distributions:** choose the universal `Auremi-Installer` archive for your processor.

Most computers use AMD64. ARM64 packages are also provided.

After installation, open **Aurémi** from the application menu for a short getting-started guide, or simply double-click an AppImage.

Use `SHA256SUMS` to verify downloaded assets.

For the trusted software-center experience, install Aurémi from the signed repository by following the README. Direct release downloads remain available and continue to include Sigstore provenance bundles.
