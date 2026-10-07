package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	got := slugify(" GitHub Copilot — édition 2026 ")
	if got != "github-copilot-dition-2026" {
		t.Fatalf("slugify() = %q", got)
	}
}

func TestUpdateINIListPreservesExistingValues(t *testing.T) {
	in := "[Default Applications]\ntext/plain=editor.desktop;\napplication/vnd.appimage=old.desktop;\n"
	out := updateINIList(in, "Default Applications", "application/vnd.appimage", handlerID)
	if !strings.Contains(out, "application/vnd.appimage="+handlerID+";old.desktop;") {
		t.Fatalf("association incorrecte:\n%s", out)
	}
	if !strings.Contains(out, "text/plain=editor.desktop;") {
		t.Fatal("l’association existante a été perdue")
	}
}

func TestValidSquashSuperblock(t *testing.T) {
	b := make([]byte, 96)
	copy(b, "hsqs")
	binary.LittleEndian.PutUint32(b[12:16], 131072)
	binary.LittleEndian.PutUint16(b[28:30], 4)
	binary.LittleEndian.PutUint16(b[30:32], 0)
	binary.LittleEndian.PutUint64(b[40:48], 4096)
	if !validSquashSuperblock(b, 8192) {
		t.Fatal("superbloc valide refusé")
	}
	binary.LittleEndian.PutUint64(b[40:48], 9000)
	if validSquashSuperblock(b, 8192) {
		t.Fatal("superbloc hors limites accepté")
	}
}

func TestValidateAppImageHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.AppImage")
	b := make([]byte, 4096)
	copy(b[:4], "\x7fELF")
	copy(b[8:11], []byte{'A', 'I', 2})
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateAppImageHeader(path); err != nil {
		t.Fatal(err)
	}
}

func TestPathContainment(t *testing.T) {
	root := "/tmp/apps"
	if !isWithin("/tmp/apps/demo", root) {
		t.Fatal("chemin enfant refusé")
	}
	if isWithin("/tmp/apps-evil/demo", root) || isWithin("/tmp/apps/../secret", root) {
		t.Fatal("sortie du dossier racine acceptée")
	}
}

func TestFindBestIconsUsesDeclaredNameAndKeepsSizes(t *testing.T) {
	root := t.TempDir()
	writeTestPNG(t, filepath.Join(root, "usr/share/icons/hicolor/64x64/apps/right-logo.png"), 64)
	writeTestPNG(t, filepath.Join(root, "usr/share/icons/hicolor/256x256/apps/right-logo.png"), 256)
	writeTestPNG(t, filepath.Join(root, "unrelated-huge.png"), 512)

	icons := findBestIcons(root, desktopMetadata{Icon: "right-logo", Name: "Example"})
	if len(icons) != 2 {
		t.Fatalf("got %d matching icons, want 2", len(icons))
	}
	for _, icon := range icons {
		if filepath.Base(icon.path) != "right-logo.png" {
			t.Fatalf("selected unrelated icon %q", icon.path)
		}
	}
}

func TestFindBestIconsRecognizesDirIconWithoutExtension(t *testing.T) {
	root := t.TempDir()
	icon := filepath.Join(root, ".DirIcon")
	writeTestPNG(t, icon, 128)

	icons := findBestIcons(root, desktopMetadata{Name: "No matching filename"})
	if len(icons) != 1 || icons[0].path != icon || icons[0].ext != ".png" {
		t.Fatalf(".DirIcon not recognized: %#v", icons)
	}
}

func TestPreferredDesktopEntryPreservesWaylandAppID(t *testing.T) {
	apps := t.TempDir()
	got := preferredDesktopEntry(apps, "com.example.Editor.desktop", "appimage-editor", "/tmp/Editor.AppImage")
	want := filepath.Join(apps, "com.example.Editor.desktop")
	if got != want {
		t.Fatalf("preferredDesktopEntry() = %q, want %q", got, want)
	}

	if err := os.WriteFile(want, []byte("[Desktop Entry]\nName=Other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got = preferredDesktopEntry(apps, "com.example.Editor.desktop", "appimage-editor", "/tmp/Editor.AppImage")
	if got != filepath.Join(apps, "appimage-editor.desktop") {
		t.Fatalf("existing launcher was not protected: %q", got)
	}
}

func TestFindAndParseDesktopKeepsDesktopID(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "org.example.App.desktop")
	if err := os.WriteFile(path, []byte("[Desktop Entry]\nName=Example\nIcon=example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := findAndParseDesktop(root)
	if err != nil {
		t.Fatal(err)
	}
	if m.DesktopID != "org.example.App.desktop" {
		t.Fatalf("DesktopID = %q", m.DesktopID)
	}
}

func writeTestPNG(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
