package main

import (
	"encoding/binary"
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
