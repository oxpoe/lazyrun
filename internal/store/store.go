// Package store menyimpan daftar shortcut milik LazyRun dalam satu file JSON.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	TypeExecutable = "executable"
	TypeWebApp     = "webapp"
)

// Targets adalah lokasi shortcut yang dipilih user.
type Targets struct {
	Menu    bool `json:"menu"`
	Desktop bool `json:"desktop"`
}

// Shortcut adalah satu baris metadata di shortcuts.json.
type Shortcut struct {
	Slug            string    `json:"slug"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	URL             string    `json:"url,omitempty"`
	Browser         string    `json:"browser,omitempty"`
	BinaryPath      string    `json:"binary_path,omitempty"`
	IconPath        string    `json:"icon_path,omitempty"`
	DesktopPath     string    `json:"desktop_path"`
	DesktopCopyPath string    `json:"desktop_copy_path,omitempty"`
	Targets         Targets   `json:"targets"`
	InstalledAt     time.Time `json:"installed_at"`
}

// Broken berarti file .desktop sudah hilang (dihapus manual oleh user).
func (s Shortcut) Broken() bool {
	_, err := os.Stat(s.DesktopPath)
	return err != nil
}

// Store adalah isi shortcuts.json (array di level atas).
type Store struct {
	Path  string
	Items []Shortcut
}

// Load membaca store dari path. File yang belum ada bukan error.
// File yang rusak di-rename ke *.bak supaya tidak menimpa data lama.
func Load(path string) (*Store, error) {
	s := &Store{Path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s.Items); err != nil {
		bak := path + ".bak"
		_ = os.Rename(path, bak)
		return s, fmt.Errorf("shortcuts.json rusak, dipindahkan ke %s: %w", bak, err)
	}
	return s, nil
}

// Save menulis store secara atomik supaya file tidak pernah setengah tertulis.
func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.Items, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.Path), ".shortcuts-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.Path)
}

// Add menambahkan shortcut; slug yang sama menimpa entry lama.
func (s *Store) Add(sc Shortcut) {
	if i, ok := s.Get(sc.Slug); ok {
		s.Items[i] = sc
		return
	}
	s.Items = append(s.Items, sc)
}

// Get mencari index shortcut berdasarkan slug.
func (s *Store) Get(slug string) (int, bool) {
	for i, sc := range s.Items {
		if sc.Slug == slug {
			return i, true
		}
	}
	return 0, false
}

// Remove menghapus shortcut berdasarkan slug.
func (s *Store) Remove(slug string) bool {
	i, ok := s.Get(slug)
	if !ok {
		return false
	}
	s.Items = append(s.Items[:i], s.Items[i+1:]...)
	return true
}
