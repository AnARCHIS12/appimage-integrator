Name:           auremi
Version:        %{auremi_version}
Release:        1%{?dist}
Summary:        Simple AppImage integration for Linux desktops
License:        MIT
URL:            https://github.com/AnARCHIS12/appimage-integrator
%description
Aurémi installs AppImages into the application menu and desktop with a
double-click. It preserves the original download and does not execute the
AppImage while inspecting it.

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/usr/share/applications
mkdir -p %{buildroot}/usr/share/mime/packages
mkdir -p %{buildroot}/usr/share/icons/hicolor/512x512/apps
mkdir -p %{buildroot}/usr/share/metainfo
install -m 0755 %{auremi_binary} %{buildroot}/usr/bin/appimage-integrator
install -m 0644 %{auremi_packaging}/appimage-integrator-handler.desktop %{buildroot}/usr/share/applications/
install -m 0644 %{auremi_packaging}/io.github.anarchis12.auremi.desktop %{buildroot}/usr/share/applications/
install -m 0644 %{auremi_packaging}/appimage-integrator.xml %{buildroot}/usr/share/mime/packages/
install -m 0644 %{auremi_packaging}/io.github.anarchis12.auremi.metainfo.xml %{buildroot}/usr/share/metainfo/
install -m 0644 %{auremi_logo} %{buildroot}/usr/share/icons/hicolor/512x512/apps/auremi.png

%post
command -v update-mime-database >/dev/null 2>&1 && update-mime-database /usr/share/mime || true
command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database /usr/share/applications || true
command -v appstreamcli >/dev/null 2>&1 && appstreamcli refresh-cache --force || true

%postun
command -v update-mime-database >/dev/null 2>&1 && update-mime-database /usr/share/mime || true
command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database /usr/share/applications || true
command -v appstreamcli >/dev/null 2>&1 && appstreamcli refresh-cache --force || true

%files
/usr/bin/appimage-integrator
/usr/share/applications/appimage-integrator-handler.desktop
/usr/share/applications/io.github.anarchis12.auremi.desktop
/usr/share/mime/packages/appimage-integrator.xml
/usr/share/icons/hicolor/512x512/apps/auremi.png
/usr/share/metainfo/io.github.anarchis12.auremi.metainfo.xml

%changelog
* Wed Oct 07 2026 AnARCHIS12 <noreply@github.com> - %{auremi_version}-1
- Add native double-click installation packages.
