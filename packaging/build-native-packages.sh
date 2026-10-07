#!/bin/sh
set -eu

version=${1:?version is required}
project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist_dir="$project_dir/dist"
packaging_dir="$project_dir/packaging"

build_deb() {
    go_arch=$1
    deb_arch=$2
    root="$dist_dir/package-deb-$go_arch"
    rm -rf "$root"
    mkdir -p "$root/DEBIAN" "$root/usr/bin" "$root/usr/share/applications" \
        "$root/usr/share/mime/packages" "$root/usr/share/icons/hicolor/512x512/apps"
    install -m 0755 "$dist_dir/appimage-integrator-linux-$go_arch" "$root/usr/bin/appimage-integrator"
    install -m 0644 "$packaging_dir/appimage-integrator-handler.desktop" "$root/usr/share/applications/"
    install -m 0644 "$packaging_dir/appimage-integrator.xml" "$root/usr/share/mime/packages/"
    install -m 0644 "$project_dir/assets/auremi-logo-512.png" "$root/usr/share/icons/hicolor/512x512/apps/auremi.png"
    install -m 0755 "$packaging_dir/debian-postinst" "$root/DEBIAN/postinst"
    install -m 0755 "$packaging_dir/debian-postrm" "$root/DEBIAN/postrm"
    cat > "$root/DEBIAN/control" <<EOF
Package: auremi
Version: $version
Section: utils
Priority: optional
Architecture: $deb_arch
Maintainer: AnARCHIS12 <noreply@github.com>
Homepage: https://github.com/AnARCHIS12/appimage-integrator
Description: Simple AppImage integration for Linux desktops
 Aurémi installs AppImages into the application menu and desktop with a
 double-click, while preserving the original downloaded file.
EOF
    dpkg-deb --root-owner-group --build "$root" "$dist_dir/Auremi-$version-linux-$go_arch.deb"
    rm -rf "$root"
}

build_rpm() {
    go_arch=$1
    rpm_arch=$2
    topdir="$dist_dir/rpmbuild-$go_arch"
    rm -rf "$topdir"
    mkdir -p "$topdir/BUILD" "$topdir/BUILDROOT" "$topdir/RPMS" "$topdir/SOURCES" "$topdir/SPECS" "$topdir/SRPMS"
    rpmbuild -bb --target "$rpm_arch" \
        --define "_topdir $topdir" \
        --define "auremi_version $version" \
        --define "auremi_binary $dist_dir/appimage-integrator-linux-$go_arch" \
        --define "auremi_packaging $packaging_dir" \
        --define "auremi_logo $project_dir/assets/auremi-logo-512.png" \
        "$packaging_dir/auremi.spec"
    rpm_file=$(find "$topdir/RPMS" -type f -name '*.rpm' -print -quit)
    cp "$rpm_file" "$dist_dir/Auremi-$version-linux-$go_arch.rpm"
    rm -rf "$topdir"
}

command -v dpkg-deb >/dev/null 2>&1 || { echo "dpkg-deb is required" >&2; exit 1; }
command -v rpmbuild >/dev/null 2>&1 || { echo "rpmbuild is required" >&2; exit 1; }

build_deb amd64 amd64
build_deb arm64 arm64
build_rpm amd64 x86_64
build_rpm arm64 aarch64
