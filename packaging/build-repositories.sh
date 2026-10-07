#!/bin/sh
set -eu

version=${1:?version is required}
output=${2:?output directory is required}
project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist_dir="$project_dir/dist"
packaging_dir="$project_dir/packaging"
work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT HUP INT TERM

for command in appstreamcli createrepo_c dpkg-deb dpkg-scanpackages gpg modifyrepo_c xmlstarlet; do
    command -v "$command" >/dev/null 2>&1 || { echo "$command is required" >&2; exit 1; }
done

rm -rf "$output"
mkdir -p "$output"
install -m 0644 "$packaging_dir/auremi-signing-key.asc" "$output/"
install -m 0644 "$packaging_dir/auremi.repo" "$output/"
install -m 0644 "$packaging_dir/auremi.sources" "$output/"
install -m 0644 "$packaging_dir/repository-index.html" "$output/index.html"
gpg --batch --yes --dearmor --output "$output/auremi-signing-key.gpg" "$packaging_dir/auremi-signing-key.asc"

package_root="$work_dir/package-root"
compose_root="$work_dir/compose"
mkdir -p "$package_root" "$compose_root"
dpkg-deb -x "$dist_dir/Auremi-$version-linux-amd64.deb" "$package_root"
appstreamcli compose --no-net --origin=auremi --result-root="$compose_root" "$package_root"

catalog_gz="$compose_root/usr/share/swcatalog/xml/auremi.xml.gz"
catalog_xml="$work_dir/auremi.xml"
gzip -dc "$catalog_gz" > "$catalog_xml"
xmlstarlet ed --inplace \
    --insert '/components/component[not(@origin)]' --type attr --name origin --value auremi \
    --subnode '/components/component[not(pkgname)]' --type elem --name pkgname --value auremi \
    "$catalog_xml"
gzip -n -9 -c "$catalog_xml" > "$work_dir/auremi.xml.gz"
appstreamcli convert "$catalog_xml" "$work_dir/Components.yml"
sed -i '/^Version:/a Origin: Auremi' "$work_dir/Components.yml"

icons_dir="$compose_root/usr/share/swcatalog/icons/auremi"
icons_tar="$work_dir/auremi-icons.tar.gz"
tar -C "$icons_dir/128x128" -cf - . | gzip -n -9 > "$icons_tar"

apt_root="$output/apt"
mkdir -p "$apt_root/pool/main/a/auremi"
for go_arch in amd64 arm64; do
    install -m 0644 \
        "$dist_dir/Auremi-$version-linux-$go_arch.deb" \
        "$apt_root/pool/main/a/auremi/auremi_${version}_${go_arch}.deb"
done

for arch in amd64 arm64; do
    binary_dir="$apt_root/dists/stable/main/binary-$arch"
    dep11_dir="$apt_root/dists/stable/main/dep11"
    mkdir -p "$binary_dir" "$dep11_dir"
    (cd "$apt_root" && dpkg-scanpackages --arch "$arch" pool /dev/null) > "$binary_dir/Packages"
    grep -q '^Package: auremi$' "$binary_dir/Packages"
    gzip -n -9 -c "$binary_dir/Packages" > "$binary_dir/Packages.gz"
    cp "$work_dir/Components.yml" "$dep11_dir/Components-$arch.yml"
    gzip -n -9 -c "$dep11_dir/Components-$arch.yml" > "$dep11_dir/Components-$arch.yml.gz"
done
cp "$icons_tar" "$apt_root/dists/stable/main/dep11/icons-128x128.tar.gz"
cp "$icons_tar" "$apt_root/dists/stable/main/dep11/icons-128x128@2.tar.gz"

release_file="$apt_root/dists/stable/Release"
apt-ftparchive \
    -o APT::FTPArchive::Release::Origin=Aurémi \
    -o APT::FTPArchive::Release::Label=Aurémi \
    -o APT::FTPArchive::Release::Suite=stable \
    -o APT::FTPArchive::Release::Codename=stable \
    -o APT::FTPArchive::Release::Architectures='amd64 arm64' \
    -o APT::FTPArchive::Release::Components=main \
    -o APT::FTPArchive::Release::Description='Official Aurémi package repository' \
    release "$apt_root/dists/stable" > "$release_file"
gpg --batch --yes --armor --detach-sign --output "$release_file.gpg" "$release_file"
gpg --batch --yes --clearsign --output "$apt_root/dists/stable/InRelease" "$release_file"

for mapping in 'amd64 x86_64' 'arm64 aarch64'; do
    set -- $mapping
    go_arch=$1
    rpm_arch=$2
    rpm_root="$output/rpm/$rpm_arch"
    mkdir -p "$rpm_root"
    install -m 0644 "$dist_dir/Auremi-$version-linux-$go_arch.rpm" "$rpm_root/"
    createrepo_c --no-database "$rpm_root"
    modifyrepo_c --mdtype=appstream "$work_dir/auremi.xml.gz" "$rpm_root/repodata"
    modifyrepo_c --mdtype=appstream-icons "$icons_tar" "$rpm_root/repodata"
    gpg --batch --yes --armor --detach-sign \
        --output "$rpm_root/repodata/repomd.xml.asc" "$rpm_root/repodata/repomd.xml"
done

gpg --batch --verify "$release_file.gpg" "$release_file"
gpg --batch --verify "$apt_root/dists/stable/InRelease"
for arch in x86_64 aarch64; do
    gpg --batch --verify "$output/rpm/$arch/repodata/repomd.xml.asc" "$output/rpm/$arch/repodata/repomd.xml"
done
