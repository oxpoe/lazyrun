// Package xdg menghitung path XDG yang dipakai LazyRun.
package xdg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func home() string {
	h, _ := os.UserHomeDir()
	return h
}

func dataHome() string {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v
	}
	return filepath.Join(home(), ".local", "share")
}

func configHome() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v
	}
	return filepath.Join(home(), ".config")
}

// BinDir adalah lokasi binary target setelah di-install.
func BinDir() string { return filepath.Join(home(), ".local", "bin") }

// ApplicationsDir menampung file .desktop untuk menu aplikasi.
func ApplicationsDir() string { return filepath.Join(dataHome(), "applications") }

// IconsDir adalah direktori ikon aplikasi XDG.
func IconsDir() string {
	return filepath.Join(dataHome(), "icons", "hicolor", "512x512", "apps")
}

// DesktopDir mengikuti xdg-user-dir, fallback ke ~/Desktop.
func DesktopDir() string {
	if out, err := exec.Command("xdg-user-dir", "DESKTOP").Output(); err == nil {
		if p := strings.TrimSpace(string(out)); p != "" {
			return p
		}
	}
	return filepath.Join(home(), "Desktop")
}

// MetadataPath adalah file daftar shortcut milik LazyRun.
func MetadataPath() string {
	return filepath.Join(configHome(), "lazyrun", "shortcuts.json")
}
