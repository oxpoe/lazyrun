package shortcut

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunsInTerminal(t *testing.T) {
	dir := t.TempDir()

	script := filepath.Join(dir, "tool.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho hai\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !RunsInTerminal(script) {
		t.Error("script harus dianggap butuh terminal")
	}

	exe := filepath.Join(dir, "win.exe")
	if err := os.WriteFile(exe, append([]byte("MZ\x90\x00"), make([]byte, 32)...), 0o755); err != nil {
		t.Fatal(err)
	}
	if RunsInTerminal(exe) {
		t.Error("aplikasi Windows (wine) tidak perlu terminal")
	}

	// binary CLI tanpa library GUI: pakai binary tes ini sendiri
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if !RunsInTerminal(self) {
		t.Error("binary CLI (tanpa lib GUI) harus dianggap butuh terminal")
	}

	// berkas tidak ada / bukan ELF
	if RunsInTerminal(filepath.Join(dir, "tidak-ada")) {
		t.Error("berkas tidak ada tidak boleh dianggap butuh terminal")
	}
}

func TestRunsInTerminalOnGUIBinary(t *testing.T) {
	// pakai aplikasi GUI nyata kalau ada di mesin ini
	candidates := []string{"/opt/helium-browser-bin/helium", "/usr/bin/vivaldi"}
	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if RunsInTerminal(path) {
			t.Errorf("%s memakai lib GUI, tidak boleh dianggap butuh terminal", path)
		}
		return
	}
	t.Skip("tidak ada binary GUI untuk diuji")
}

func TestRunsInTerminalOnRealCLITool(t *testing.T) {
	// lazydocker (TUI Go) adalah kasus nyata yang dilaporkan user
	if _, err := os.Stat("/home/oxpoe/.local/bin/lazydocker"); err != nil {
		t.Skip("lazydocker tidak ada di mesin ini")
	}
	if !RunsInTerminal("/home/oxpoe/.local/bin/lazydocker") {
		t.Error("lazydocker adalah TUI, harus butuh terminal")
	}
}

var _ = exec.LookPath
