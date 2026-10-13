package cursor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIniParsesSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.theme")
	body := "# komentar\n[Icon Theme]\nName=default\nInherits=Breeze_Cursors\n\n[Other]\nX=1\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	got := ini(path)
	if got["icon theme"]["inherits"] != "Breeze_Cursors" {
		t.Errorf("Inherits salah: %q", got["icon theme"]["inherits"])
	}
	if got["other"]["x"] != "1" {
		t.Errorf("section lain salah: %+v", got)
	}
}

func TestDesktopCursorFromKDE(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kcminputrc"), []byte("[Mouse]\ncursorTheme=material_cursors\ncursorSize=36\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	theme, size := desktopCursor()
	if theme != "material_cursors" || size != "36" {
		t.Errorf("desktopCursor() = %q, %q", theme, size)
	}
}

func TestApplySetsEnvWhenEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XCURSOR_THEME", "")
	t.Setenv("XCURSOR_SIZE", "")

	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kcminputrc"), []byte("[Mouse]\ncursorTheme=material_cursors\ncursorSize=36\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	Apply()
	if got := os.Getenv("XCURSOR_THEME"); got != "material_cursors" {
		t.Errorf("XCURSOR_THEME = %q", got)
	}
	if got := os.Getenv("XCURSOR_SIZE"); got != "36" {
		t.Errorf("XCURSOR_SIZE = %q", got)
	}
}

func TestApplyKeepsExistingEnv(t *testing.T) {
	t.Setenv("XCURSOR_THEME", "Breeze_Cursors")
	t.Setenv("XCURSOR_SIZE", "24")

	Apply()
	if got := os.Getenv("XCURSOR_THEME"); got != "Breeze_Cursors" {
		t.Errorf("env yang sudah ada tidak boleh ditimpa, dapat %q", got)
	}
}
