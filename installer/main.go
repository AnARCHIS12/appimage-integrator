package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed payload/appimage-integrator
var integratorBinary []byte

//go:embed assets/auremi-logo-512.png
var auremiLogo []byte

func main() {
	if err := run(); err != nil {
		showMessage("error", "Aurémi installation failed", err.Error())
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	if os.Geteuid() == 0 {
		return errors.New("run the graphical installer from your normal account, without sudo")
	}
	if os.Getenv("AUREMI_INSTALLER_ASSUME_YES") != "1" && !confirmInstall() {
		return nil
	}
	tempDir, err := os.MkdirTemp("", "auremi-installer-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	payload := filepath.Join(tempDir, "appimage-integrator")
	if err := os.WriteFile(payload, integratorBinary, 0o755); err != nil {
		return err
	}
	cmd := exec.Command(payload, "setup", "--user")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return errors.New(message)
	}
	showMessage("info", "Aurémi is installed", "You can now double-click an AppImage to add it to your application menu and desktop.")
	return nil
}

func confirmInstall() bool {
	question := "Install Aurémi for your account?\n\nNo administrator password is required."
	if os.Getenv("AUREMI_INSTALLER_NO_DIALOG") == "1" {
		return true
	}
	logo := temporaryLogo()
	defer os.Remove(logo)
	if path, err := exec.LookPath("kdialog"); err == nil {
		args := []string{"--title", "Aurémi", "--yesno", question, "--yes-label", "Install", "--no-label", "Cancel"}
		if logo != "" {
			args = append([]string{"--icon", logo}, args...)
		}
		return exec.Command(path, args...).Run() == nil
	}
	if path, err := exec.LookPath("zenity"); err == nil {
		args := []string{"--question", "--title=Aurémi", "--text=" + question, "--ok-label=Install", "--cancel-label=Cancel"}
		if logo != "" {
			args = append(args, "--window-icon="+logo)
		}
		return exec.Command(path, args...).Run() == nil
	}
	if path, err := exec.LookPath("yad"); err == nil {
		args := []string{"--question", "--title=Aurémi", "--text=" + question, "--button=Install:0", "--button=Cancel:1"}
		if logo != "" {
			args = append(args, "--window-icon="+logo)
		}
		return exec.Command(path, args...).Run() == nil
	}
	if path, err := exec.LookPath("xmessage"); err == nil {
		return exec.Command(path, "-center", "-title", "Aurémi", "-buttons", "Install:0,Cancel:1", question).Run() == nil
	}
	return true
}

func showMessage(kind, title, message string) {
	if os.Getenv("AUREMI_INSTALLER_NO_DIALOG") == "1" {
		fmt.Println(title + ": " + message)
		return
	}
	logo := temporaryLogo()
	defer os.Remove(logo)
	if path, err := exec.LookPath("kdialog"); err == nil {
		mode := "--msgbox"
		if kind == "error" {
			mode = "--error"
		}
		args := []string{"--title", "Aurémi", mode, message}
		if logo != "" {
			args = append([]string{"--icon", logo}, args...)
		}
		_ = exec.Command(path, args...).Run()
		return
	}
	if path, err := exec.LookPath("zenity"); err == nil {
		mode := "--info"
		if kind == "error" {
			mode = "--error"
		}
		args := []string{mode, "--title=" + title, "--text=" + message}
		if logo != "" {
			args = append(args, "--window-icon="+logo)
		}
		_ = exec.Command(path, args...).Run()
		return
	}
	if path, err := exec.LookPath("yad"); err == nil {
		mode := "--info"
		if kind == "error" {
			mode = "--error"
		}
		args := []string{mode, "--title=" + title, "--text=" + message, "--button=OK:0"}
		if logo != "" {
			args = append(args, "--window-icon="+logo)
		}
		_ = exec.Command(path, args...).Run()
		return
	}
	if path, err := exec.LookPath("xmessage"); err == nil {
		_ = exec.Command(path, "-center", "-title", title, "-buttons", "OK:0", message).Run()
		return
	}
	if path, err := exec.LookPath("notify-send"); err == nil {
		args := []string{"--app-name", "Aurémi"}
		if logo != "" {
			args = append(args, "--icon", logo)
		}
		args = append(args, title, message)
		_ = exec.Command(path, args...).Run()
		return
	}
	fmt.Fprintln(os.Stderr, title+":", message)
}

func temporaryLogo() string {
	file, err := os.CreateTemp("", "auremi-logo-*.png")
	if err != nil {
		return ""
	}
	name := file.Name()
	if _, err := file.Write(auremiLogo); err != nil {
		file.Close()
		os.Remove(name)
		return ""
	}
	if err := file.Close(); err != nil {
		os.Remove(name)
		return ""
	}
	return name
}
