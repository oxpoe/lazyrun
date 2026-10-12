package xdg

import (
	"path/filepath"
	"testing"
)

func TestDataHomeOverride(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("XDG_DATA_HOME", "/data")
	if got := ApplicationsDir(); got != "/data/applications" {
		t.Errorf("ApplicationsDir() = %s", got)
	}
	if got := IconsDir(); got != filepath.Join("/data", "icons/hicolor/512x512/apps") {
		t.Errorf("IconsDir() = %s", got)
	}
}

func TestFallbacks(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if got := ApplicationsDir(); got != "/home/tester/.local/share/applications" {
		t.Errorf("ApplicationsDir() = %s", got)
	}
	if got := BinDir(); got != "/home/tester/.local/bin" {
		t.Errorf("BinDir() = %s", got)
	}
	if got := MetadataPath(); got != "/home/tester/.config/lazyrun/shortcuts.json" {
		t.Errorf("MetadataPath() = %s", got)
	}

	// xdg-user-dir tidak ada di $PATH → fallback ke ~/Desktop.
	t.Setenv("PATH", "")
	if got := DesktopDir(); got != "/home/tester/Desktop" {
		t.Errorf("DesktopDir() = %s", got)
	}
}
