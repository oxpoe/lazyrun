package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileIsEmpty(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "shortcuts.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Items) != 0 {
		t.Errorf("store baru harus kosong, dapat %d item", len(s.Items))
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "shortcuts.json")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Add(Shortcut{Slug: "winbox", Name: "Winbox", Type: TypeExecutable, Targets: Targets{Menu: true}})
	s.Add(Shortcut{Slug: "wa", Name: "WhatsApp Web", Type: TypeWebApp, Targets: Targets{Menu: true, Desktop: true}})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[1].Name != "WhatsApp Web" {
		t.Fatalf("hasil load salah: %+v", got.Items)
	}

	if _, ok := got.Get("wa"); !ok {
		t.Error("Get by slug gagal")
	}
	if !got.Remove("wa") || len(got.Items) != 1 {
		t.Error("Remove gagal")
	}
}

func TestLoadCorruptFileIsBackedUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shortcuts.json")
	if err := os.WriteFile(path, []byte("{bukan json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("file rusak harus dilaporkan sebagai error")
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Errorf("file rusak harus di-rename ke .bak: %v", err)
	}
}
