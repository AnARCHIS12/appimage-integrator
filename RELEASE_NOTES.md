# Aurémi 0.3.3

This release lets Aurémi identify the precise version of an integrated AppImage while retaining the icon and Wayland improvements from 0.3.2. Release downloads remain cryptographically signed with Sigstore.

## Highlights

- Reads the official `X-AppImage-Version` field when supplied by the publisher.
- Falls back to matching embedded AppStream release metadata when needed.
- Recognizes a structured version in the filename only when no authoritative metadata exists.
- Stores the application version in `metadata.json` and the generated desktop entry.
- Shows the detected version after integration and in the managed-application list.
- Keeps the corrected icon selection, multiple icon resolutions, and Wayland launcher matching introduced in 0.3.2.
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
