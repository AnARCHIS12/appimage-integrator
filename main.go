package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	programName = "appimage-integrator"
	productName = "Aurémi"
	version     = "0.1.0"
	handlerID   = "appimage-integrator-handler.desktop"
)

//go:embed assets/auremi-logo-512.png
var auremiLogo []byte

type appMetadata struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Comment     string `json:"comment,omitempty"`
	Source      string `json:"source"`
	Installed   string `json:"installed"`
	SHA256      string `json:"sha256"`
	Integrated  string `json:"integrated_at"`
	Icon        string `json:"icon,omitempty"`
	DesktopFile string `json:"desktop_file"`
}

type desktopMetadata struct {
	Name           string
	Comment        string
	Categories     string
	StartupWMClass string
	Icon           string
}

type paths struct {
	home       string
	dataHome   string
	configHome string
	desktopDir string
	appsRoot   string
	appEntries string
	iconsRoot  string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}
	switch args[0] {
	case "integrate", "install":
		return integrateCommand(args[1:])
	case "setup":
		return setupCommand(args[1:])
	case "list":
		return listCommand()
	case "remove", "uninstall":
		return removeCommand(args[1:])
	case "version", "--version", "-v":
		fmt.Println(programName, version)
		return nil
	case "help", "--help", "-h":
		printUsage(os.Stdout)
		return nil
	default:
		if strings.HasSuffix(strings.ToLower(args[0]), ".appimage") {
			return integrateCommand(args)
		}
		return usageError()
	}
}

func usageError() error {
	printUsage(os.Stderr)
	return errors.New("commande manquante ou inconnue")
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `%s %s — intégration silencieuse des AppImage

Utilisation :
  %s setup --user              Installer le gestionnaire pour l'utilisateur
  sudo %s setup --system       Installer le gestionnaire pour tous les utilisateurs
  %s integrate fichier.AppImage
  %s list
  %s remove IDENTIFIANT

L'intégration ne lance jamais l'AppImage et ne supprime jamais le fichier source.
`, programName, version, programName, programName, programName, programName, programName)
}

func integrateCommand(args []string) error {
	set := flag.NewFlagSet("integrate", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	noDesktop := set.Bool("no-desktop", false, "ne pas créer de raccourci sur le bureau")
	quiet := set.Bool("quiet", false, "ne pas afficher de notification")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return errors.New("indiquez exactement un fichier AppImage")
	}
	result, err := integrate(set.Arg(0), !*noDesktop)
	if err != nil {
		if !*quiet {
			notify("Échec de l’installation AppImage", err.Error(), "dialog-error")
		}
		return err
	}
	fmt.Printf("%s installé\nIdentifiant : %s\nEmplacement : %s\n", result.Name, result.ID, result.Installed)
	if !*quiet {
		notify("AppImage installée", result.Name+" est maintenant disponible dans le menu des applications.", result.Icon)
	}
	return nil
}

func integrate(input string, desktopShortcut bool) (*appMetadata, error) {
	source, err := filepath.Abs(input)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(source)
	if err != nil {
		return nil, fmt.Errorf("fichier inaccessible : %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("les liens symboliques sont refusés; sélectionnez le fichier AppImage réel")
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("la source n’est pas un fichier ordinaire")
	}
	if info.Size() < 4096 {
		return nil, errors.New("fichier trop petit pour être une AppImage valide")
	}
	if err := validateAppImageHeader(source); err != nil {
		return nil, err
	}

	p, err := userPaths()
	if err != nil {
		return nil, err
	}
	meta := desktopMetadata{
		Name:       cleanDisplayName(strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))),
		Comment:    "Application AppImage intégrée localement",
		Categories: "Utility;",
	}
	tmp, err := os.MkdirTemp("", "appimage-integrator-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	extracted, extractErr := extractMetadata(source, tmp)
	if extractErr == nil {
		if parsed, err := findAndParseDesktop(extracted); err == nil {
			mergeDesktopMetadata(&meta, parsed)
		}
	}
	if meta.Name == "" {
		meta.Name = "AppImage"
	}
	id := uniqueID(slugify(meta.Name), source, p.appsRoot)
	appDir := filepath.Join(p.appsRoot, id)
	installed := filepath.Join(appDir, safeAppImageFilename(meta.Name))
	entry := filepath.Join(p.appEntries, id+".desktop")

	for _, dir := range []string{appDir, p.appEntries, p.iconsRoot} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	if err := copyFileAtomic(source, installed, 0o755); err != nil {
		return nil, fmt.Errorf("copie de l’AppImage : %w", err)
	}

	iconName := id
	iconPath := ""
	if extracted != "" {
		if candidate := findBestIcon(extracted, meta.Icon); candidate != "" {
			ext := strings.ToLower(filepath.Ext(candidate))
			if ext == ".png" || ext == ".svg" || ext == ".xpm" {
				iconDir := filepath.Join(p.dataHome, "icons", "hicolor", iconSizeDir(candidate), "apps")
				if err := os.MkdirAll(iconDir, 0o755); err == nil {
					iconPath = filepath.Join(iconDir, id+ext)
					if err := copyFileAtomic(candidate, iconPath, 0o644); err != nil {
						iconPath = ""
					}
				}
			}
		}
	}
	if iconPath == "" {
		iconDir := filepath.Join(p.dataHome, "icons", "hicolor", "scalable", "apps")
		if err := os.MkdirAll(iconDir, 0o755); err != nil {
			return nil, err
		}
		iconPath = filepath.Join(iconDir, id+".svg")
		if err := writeAtomic(iconPath, []byte(genericIconSVG), 0o644); err != nil {
			return nil, err
		}
	}

	desktopBody := installedDesktopEntry(meta, installed, iconName, source)
	if err := writeAtomic(entry, []byte(desktopBody), 0o755); err != nil {
		return nil, err
	}
	if desktopShortcut && p.desktopDir != "" {
		if err := os.MkdirAll(p.desktopDir, 0o755); err == nil {
			shortcut := filepath.Join(p.desktopDir, safeDesktopFilename(meta.Name)+".desktop")
			_ = writeAtomic(shortcut, []byte(desktopBody), 0o755)
		}
	}

	hash, err := fileSHA256(installed)
	if err != nil {
		return nil, err
	}
	result := &appMetadata{
		ID:          id,
		Name:        meta.Name,
		Comment:     meta.Comment,
		Source:      source,
		Installed:   installed,
		SHA256:      hash,
		Integrated:  time.Now().Format(time.RFC3339),
		Icon:        iconPath,
		DesktopFile: entry,
	}
	encoded, _ := json.MarshalIndent(result, "", "  ")
	if err := writeAtomic(filepath.Join(appDir, "metadata.json"), append(encoded, '\n'), 0o644); err != nil {
		return nil, err
	}
	refreshDesktop(p)
	return result, nil
}

func validateAppImageHeader(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	header := make([]byte, 12)
	if _, err := io.ReadFull(f, header); err != nil {
		return errors.New("en-tête AppImage illisible")
	}
	if string(header[:4]) != "\x7fELF" {
		return errors.New("le fichier n’est pas un exécutable ELF")
	}
	if header[8] != 'A' || header[9] != 'I' || (header[10] != 1 && header[10] != 2) {
		return errors.New("signature AppImage absente ou format non pris en charge")
	}
	return nil
}

func extractMetadata(source, tempRoot string) (string, error) {
	unsquashfs, err := exec.LookPath("unsquashfs")
	if err != nil {
		return "", errors.New("squashfs-tools absent; utilisation de l’icône générique")
	}
	offset, err := findSquashFSOffset(source)
	if err != nil {
		return "", err
	}
	payload := filepath.Join(tempRoot, "payload.squashfs")
	if err := copyFromOffset(source, payload, offset); err != nil {
		return "", err
	}
	out := filepath.Join(tempRoot, "metadata")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	patterns := []string{
		"*.desktop", ".DirIcon", "*.png", "*.svg", "*.xpm",
		"usr/share/applications/*.desktop",
		"usr/share/icons/hicolor/*/apps/*.png",
		"usr/share/icons/hicolor/*/apps/*.svg",
		"usr/share/pixmaps/*",
	}
	args := append([]string{"-no-progress", "-d", out, payload}, patterns...)
	cmd := exec.CommandContext(ctx, unsquashfs, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", errors.New("délai dépassé pendant la lecture des métadonnées")
		}
		return "", fmt.Errorf("lecture SquashFS impossible : %w", err)
	}
	return out, nil
}

func findSquashFSOffset(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	const chunkSize = 4 << 20
	buf := make([]byte, chunkSize+95)
	var base int64
	var overlap int
	for {
		n, readErr := f.Read(buf[overlap:])
		total := overlap + n
		for start := 0; start+96 <= total; {
			i := bytesIndex(buf[start:total], []byte("hsqs"))
			if i < 0 {
				break
			}
			i += start
			candidate := base - int64(overlap) + int64(i)
			if validSquashSuperblock(buf[i:i+96], info.Size()-candidate) {
				return candidate, nil
			}
			start = i + 4
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return 0, readErr
		}
		if total > 95 {
			overlap = 95
			copy(buf[:overlap], buf[total-overlap:total])
		} else {
			overlap = total
		}
		base += int64(n)
	}
	return 0, errors.New("système SquashFS AppImage introuvable")
}

func bytesIndex(haystack, needle []byte) int {
	if len(needle) == 0 {
		return 0
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func validSquashSuperblock(b []byte, remaining int64) bool {
	if len(b) < 96 || string(b[:4]) != "hsqs" {
		return false
	}
	blockSize := binary.LittleEndian.Uint32(b[12:16])
	major := binary.LittleEndian.Uint16(b[28:30])
	minor := binary.LittleEndian.Uint16(b[30:32])
	bytesUsed := binary.LittleEndian.Uint64(b[40:48])
	return major == 4 && minor == 0 && blockSize >= 4096 && blockSize <= 1<<20 && blockSize&(blockSize-1) == 0 && bytesUsed >= 96 && bytesUsed <= uint64(remaining)
}

func copyFromOffset(source, dest string, offset int64) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, io.NewSectionReader(in, offset, info.Size()-offset))
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func findAndParseDesktop(root string) (desktopMetadata, error) {
	var candidates []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(d.Name()), ".desktop") {
			candidates = append(candidates, path)
		}
		return nil
	})
	if err != nil || len(candidates) == 0 {
		return desktopMetadata{}, errors.New("aucun fichier .desktop embarqué")
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return desktopRank(candidates[i]) < desktopRank(candidates[j])
	})
	return parseDesktopFile(candidates[0])
}

func desktopRank(path string) int {
	clean := filepath.ToSlash(path)
	if strings.Contains(clean, "/usr/share/applications/") {
		return 1
	}
	return 0
}

func parseDesktopFile(path string) (desktopMetadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return desktopMetadata{}, err
	}
	defer f.Close()
	var m desktopMetadata
	inMain := false
	s := bufio.NewScanner(io.LimitReader(f, 1<<20))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inMain = line == "[Desktop Entry]"
			continue
		}
		if !inMain || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "Name":
			if m.Name == "" {
				m.Name = cleanDisplayName(val)
			}
		case "Comment":
			if m.Comment == "" {
				m.Comment = cleanDesktopValue(val)
			}
		case "Categories":
			m.Categories = cleanCategories(val)
		case "StartupWMClass":
			m.StartupWMClass = cleanDesktopValue(val)
		case "Icon":
			m.Icon = cleanDesktopValue(val)
		}
	}
	return m, s.Err()
}

func mergeDesktopMetadata(dst *desktopMetadata, src desktopMetadata) {
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.Comment != "" {
		dst.Comment = src.Comment
	}
	if src.Categories != "" {
		dst.Categories = src.Categories
	}
	if src.StartupWMClass != "" {
		dst.StartupWMClass = src.StartupWMClass
	}
	dst.Icon = src.Icon
}

func findBestIcon(root, desktopIcon string) string {
	target := strings.TrimSuffix(filepath.Base(desktopIcon), filepath.Ext(desktopIcon))
	type candidate struct {
		path  string
		score int64
	}
	var all []candidate
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".png" && ext != ".svg" && ext != ".xpm" {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > 10<<20 {
			return nil
		}
		score := info.Size()
		base := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		if target != "" && strings.EqualFold(base, target) {
			score += 1 << 40
		}
		if d.Name() == ".DirIcon" || strings.EqualFold(base, "icon") {
			score += 1 << 39
		}
		if strings.Contains(filepath.ToSlash(path), "/usr/share/icons/") {
			score += 1 << 38
		}
		all = append(all, candidate{path, score})
		return nil
	})
	if len(all) == 0 {
		return ""
	}
	sort.Slice(all, func(i, j int) bool { return all[i].score > all[j].score })
	return all[0].path
}

func iconSizeDir(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".svg") {
		return "scalable"
	}
	f, err := os.Open(path)
	if err != nil {
		return "512x512"
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(io.LimitReader(f, 16<<20))
	if err != nil || cfg.Width < 16 || cfg.Height < 16 {
		return "512x512"
	}
	return strconv.Itoa(cfg.Width) + "x" + strconv.Itoa(cfg.Height)
}

func installedDesktopEntry(m desktopMetadata, executable, iconName, source string) string {
	wm := ""
	if m.StartupWMClass != "" {
		wm = "StartupWMClass=" + cleanDesktopValue(m.StartupWMClass) + "\n"
	}
	return fmt.Sprintf(`[Desktop Entry]
Version=1.0
Type=Application
Name=%s
Comment=%s
Exec=%s %%U
TryExec=%s
Icon=%s
Terminal=false
StartupNotify=true
%sCategories=%s
X-AppImage-Source=%s
X-AppImage-Integrator-Version=%s
`, cleanDesktopValue(m.Name), cleanDesktopValue(m.Comment), quoteExec(executable), cleanDesktopValue(executable), cleanDesktopValue(iconName), wm, cleanCategories(m.Categories), cleanDesktopValue(source), version)
}

func setupCommand(args []string) error {
	set := flag.NewFlagSet("setup", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	user := set.Bool("user", false, "installation utilisateur")
	system := set.Bool("system", false, "installation système")
	if err := set.Parse(args); err != nil {
		return err
	}
	if *user == *system {
		return errors.New("choisissez exactement --user ou --system")
	}
	if *system {
		if os.Geteuid() != 0 {
			return errors.New("l’installation système doit être lancée avec sudo")
		}
		return setupSystem()
	}
	if os.Geteuid() == 0 {
		return errors.New("n’utilisez pas sudo avec --user")
	}
	return setupUser()
}

func setupUser() error {
	p, err := userPaths()
	if err != nil {
		return err
	}
	binDir := filepath.Join(p.home, ".local", "bin")
	bin := filepath.Join(binDir, programName)
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	if err := installSelf(bin); err != nil {
		return err
	}
	if err := writeHandlerFiles(bin, p.dataHome); err != nil {
		return err
	}
	if err := setMIMEAssociation(filepath.Join(p.configHome, "mimeapps.list")); err != nil {
		return err
	}
	refreshMIME(filepath.Join(p.dataHome, "mime"), p.appEntries)
	fmt.Println("Gestionnaire AppImage installé pour", p.home)
	fmt.Println("Un double-clic sur une AppImage l’intégrera désormais au menu des applications.")
	return nil
}

func setupSystem() error {
	bin := "/usr/local/bin/" + programName
	if err := installSelf(bin); err != nil {
		return err
	}
	if err := writeHandlerFiles(bin, "/usr/share"); err != nil {
		return err
	}
	if err := setMIMEAssociation("/etc/xdg/mimeapps.list"); err != nil {
		return err
	}
	refreshMIME("/usr/share/mime", "/usr/share/applications")
	fmt.Println("Gestionnaire AppImage installé pour tous les utilisateurs.")
	return nil
}

func installSelf(destination string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return err
	}
	if sameFile(self, destination) {
		return os.Chmod(destination, 0o755)
	}
	return copyFileAtomic(self, destination, 0o755)
}

func writeHandlerFiles(binary, dataRoot string) error {
	apps := filepath.Join(dataRoot, "applications")
	mimePackages := filepath.Join(dataRoot, "mime", "packages")
	iconDir := filepath.Join(dataRoot, "icons", "hicolor", "512x512", "apps")
	if err := os.MkdirAll(apps, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(mimePackages, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(iconDir, 0o755); err != nil {
		return err
	}
	if err := writeAtomic(filepath.Join(iconDir, "auremi.png"), auremiLogo, 0o644); err != nil {
		return err
	}
	desktop := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Installer une AppImage avec Aurémi
Comment=Intègre silencieusement une AppImage au bureau Linux avec Aurémi
Exec=%s integrate %%f
TryExec=%s
Icon=auremi
Terminal=false
NoDisplay=true
MimeType=application/vnd.appimage;application/x-iso9660-appimage;
`, quoteExec(binary), cleanDesktopValue(binary))
	if err := writeAtomic(filepath.Join(apps, handlerID), []byte(desktop), 0o644); err != nil {
		return err
	}
	return writeAtomic(filepath.Join(mimePackages, "appimage-integrator.xml"), []byte(mimeXML), 0o644)
}

func setMIMEAssociation(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, _ := os.ReadFile(path)
	updated := updateINIList(string(data), "Default Applications", "application/vnd.appimage", handlerID)
	updated = updateINIList(updated, "Default Applications", "application/x-iso9660-appimage", handlerID)
	updated = updateINIList(updated, "Added Associations", "application/vnd.appimage", handlerID)
	updated = updateINIList(updated, "Added Associations", "application/x-iso9660-appimage", handlerID)
	return writeAtomic(path, []byte(updated), 0o644)
}

func updateINIList(content, section, key, value string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	header := "[" + section + "]"
	start, end := -1, len(lines)
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			start = i
			for j := i + 1; j < len(lines); j++ {
				if strings.HasPrefix(strings.TrimSpace(lines[j]), "[") {
					end = j
					break
				}
			}
			break
		}
	}
	entry := key + "=" + value + ";"
	if start == -1 {
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, header, entry)
		return strings.TrimLeft(strings.Join(lines, "\n"), "\n") + "\n"
	}
	for i := start + 1; i < end; i++ {
		k, old, ok := strings.Cut(lines[i], "=")
		if ok && strings.TrimSpace(k) == key {
			values := []string{value}
			for _, item := range strings.Split(old, ";") {
				item = strings.TrimSpace(item)
				if item != "" && item != value {
					values = append(values, item)
				}
			}
			lines[i] = key + "=" + strings.Join(values, ";") + ";"
			return strings.Join(lines, "\n")
		}
	}
	lines = append(lines[:end], append([]string{entry}, lines[end:]...)...)
	return strings.Join(lines, "\n")
}

func listCommand() error {
	p, err := userPaths()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(p.appsRoot)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Aucune AppImage intégrée.")
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(p.appsRoot, entry.Name(), "metadata.json"))
		var m appMetadata
		if err == nil && json.Unmarshal(data, &m) == nil {
			fmt.Printf("%-28s %s\n", m.ID, m.Name)
		} else {
			fmt.Println(entry.Name())
		}
	}
	return nil
}

func removeCommand(args []string) error {
	if len(args) != 1 {
		return errors.New("indiquez l’identifiant affiché par la commande list")
	}
	id := args[0]
	if slugify(id) != id || id == "" {
		return errors.New("identifiant invalide")
	}
	p, err := userPaths()
	if err != nil {
		return err
	}
	appDir := filepath.Join(p.appsRoot, id)
	if !isWithin(appDir, p.appsRoot) {
		return errors.New("chemin de désinstallation refusé")
	}
	data, err := os.ReadFile(filepath.Join(appDir, "metadata.json"))
	if err != nil {
		return errors.New("application gérée introuvable")
	}
	var m appMetadata
	if err := json.Unmarshal(data, &m); err != nil || m.ID != id {
		return errors.New("métadonnées de désinstallation invalides")
	}
	if !isWithin(m.DesktopFile, p.appEntries) {
		return errors.New("chemin du lanceur invalide dans les métadonnées")
	}
	_ = os.Remove(m.DesktopFile)
	if m.Icon != "" && isWithin(m.Icon, filepath.Join(p.dataHome, "icons")) {
		_ = os.Remove(m.Icon)
	}
	if p.desktopDir != "" {
		_ = os.Remove(filepath.Join(p.desktopDir, safeDesktopFilename(m.Name)+".desktop"))
	}
	if err := os.RemoveAll(appDir); err != nil {
		return err
	}
	refreshDesktop(p)
	fmt.Println(m.Name, "désinstallé. Le fichier source n’a pas été supprimé.")
	return nil
}

func userPaths() (paths, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return paths{}, errors.New("dossier personnel introuvable")
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local", "share")
	}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}
	return paths{
		home:       home,
		dataHome:   dataHome,
		configHome: configHome,
		desktopDir: discoverDesktopDir(home, configHome),
		appsRoot:   filepath.Join(dataHome, programName, "apps"),
		appEntries: filepath.Join(dataHome, "applications"),
		iconsRoot:  filepath.Join(dataHome, "icons"),
	}, nil
}

func discoverDesktopDir(home, configHome string) string {
	data, _ := os.ReadFile(filepath.Join(configHome, "user-dirs.dirs"))
	re := regexp.MustCompile(`(?m)^XDG_DESKTOP_DIR="([^"]*)"`)
	if match := re.FindStringSubmatch(string(data)); len(match) == 2 {
		value := strings.ReplaceAll(match[1], "$HOME", home)
		if filepath.IsAbs(value) {
			return filepath.Clean(value)
		}
	}
	for _, name := range []string{"Desktop", "Bureau"} {
		candidate := filepath.Join(home, name)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return filepath.Join(home, "Desktop")
}

func uniqueID(base, source, root string) string {
	if base == "" {
		base = "appimage"
	}
	candidate := "appimage-" + base
	metaPath := filepath.Join(root, candidate, "metadata.json")
	if data, err := os.ReadFile(metaPath); err == nil {
		var m appMetadata
		if json.Unmarshal(data, &m) == nil && m.Source == source {
			return candidate
		}
	}
	if _, err := os.Stat(filepath.Join(root, candidate)); errors.Is(err, os.ErrNotExist) {
		return candidate
	}
	sum := sha256.Sum256([]byte(source))
	return candidate + "-" + hex.EncodeToString(sum[:3])
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	dash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if b.Len() > 0 && !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func cleanDisplayName(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "\x00", ""), "_", " "))
	s = strings.TrimSuffix(s, ".AppImage")
	if len(s) > 100 {
		s = s[:100]
	}
	return cleanDesktopValue(s)
}

func cleanDesktopValue(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func cleanCategories(s string) string {
	var out []string
	seen := map[string]bool{}
	for _, part := range strings.Split(s, ";") {
		part = regexp.MustCompile(`[^A-Za-z0-9-]`).ReplaceAllString(part, "")
		if part != "" && !seen[part] {
			out = append(out, part)
			seen[part] = true
		}
	}
	if len(out) == 0 {
		out = []string{"Utility"}
	}
	return strings.Join(out, ";") + ";"
}

func quoteExec(path string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "`", "\\`", "$", "\\$")
	return "\"" + replacer.Replace(path) + "\""
}

func safeAppImageFilename(name string) string {
	base := safeDesktopFilename(name)
	if base == "" {
		base = "Application"
	}
	return base + ".AppImage"
}

func safeDesktopFilename(name string) string {
	name = cleanDisplayName(name)
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r < 32 {
			return '-'
		}
		return r
	}, name)
	return strings.Trim(name, " .")
}

func copyFileAtomic(source, destination string, mode fs.FileMode) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(destination), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, destination)
}

func writeAtomic(path string, data []byte, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sameFile(a, b string) bool {
	ai, errA := os.Stat(a)
	bi, errB := os.Stat(b)
	return errA == nil && errB == nil && os.SameFile(ai, bi)
}

func isWithin(path, root string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func refreshDesktop(p paths) {
	runOptional("update-desktop-database", p.appEntries)
	runOptional("gtk-update-icon-cache", "-q", "-t", "-f", filepath.Join(p.dataHome, "icons", "hicolor"))
	runOptional("kbuildsycoca6", "--noincremental")
}

func refreshMIME(mimeDir, appDir string) {
	runOptional("update-mime-database", mimeDir)
	runOptional("update-desktop-database", appDir)
}

func runOptional(name string, args ...string) {
	path, err := exec.LookPath(name)
	if err != nil {
		return
	}
	cmd := exec.Command(path, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	_ = cmd.Run()
}

func notify(title, body, icon string) {
	path, err := exec.LookPath("notify-send")
	if err != nil || os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		return
	}
	args := []string{"--app-name", productName}
	if icon != "" {
		args = append(args, "--icon", icon)
	}
	args = append(args, title, body)
	cmd := exec.Command(path, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	_ = cmd.Start()
}

const mimeXML = `<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
  <mime-type type="application/vnd.appimage">
    <comment>Application AppImage</comment>
    <glob pattern="*.AppImage" weight="90"/>
    <glob pattern="*.appimage" weight="90"/>
    <magic priority="80">
      <match type="string" offset="8" value="AI\x01"/>
      <match type="string" offset="8" value="AI\x02"/>
    </magic>
  </mime-type>
  <mime-type type="application/x-iso9660-appimage">
    <comment>Application AppImage (compatibilité)</comment>
    <sub-class-of type="application/vnd.appimage"/>
  </mime-type>
</mime-info>
`

const genericIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
  <defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop stop-color="#7c3aed"/><stop offset="1" stop-color="#2563eb"/></linearGradient></defs>
  <rect width="512" height="512" rx="104" fill="url(#g)"/>
  <path d="M256 88 402 172v168L256 424 110 340V172z" fill="none" stroke="#fff" stroke-width="30" stroke-linejoin="round"/>
  <path d="M256 88v168m146-84-146 84-146-84m146 84v168" fill="none" stroke="#fff" stroke-width="24" stroke-linejoin="round" opacity=".9"/>
</svg>
`
