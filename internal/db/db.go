// Package db menyimpan daftar shortcut LazyRun dalam satu berkas SQLite (.lazy)
// supaya bisa dipindah antar komputer.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"lazyrun/internal/store"
)

// FormatVersion dipakai untuk mengenali berkas .lazy buatan LazyRun.
// Versi 1 hanya metadata; versi 2 menyertakan isi binary dan ikon.
const FormatVersion = "2"

// DefaultName adalah nama berkas cadangan yang disarankan.
const DefaultName = "lazyrun-shortcuts.lazy"

const schema = `
CREATE TABLE meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE shortcuts (
	slug              TEXT PRIMARY KEY,
	name              TEXT NOT NULL,
	type              TEXT NOT NULL,
	kind              TEXT,
	url               TEXT,
	browser           TEXT,
	exec_args         TEXT,
	categories        TEXT,
	binary_path       TEXT,
	icon_path         TEXT,
	icon_source       TEXT,
	desktop_path      TEXT,
	desktop_copy_path TEXT,
	terminal          INTEGER NOT NULL DEFAULT 0,
	menu              INTEGER NOT NULL,
	desktop           INTEGER NOT NULL,
	installed_at      TEXT,
	binary_data       BLOB,
	icon_data         BLOB
);`

// Export menulis seluruh daftar shortcut ke berkas SQLite baru.
func Export(path string, appVersion string, items []store.Shortcut) error {
	if path == "" {
		return errors.New("lokasi berkas ekspor kosong")
	}
	_ = os.Remove(path) // mulai dari berkas baru

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Exec(schema); err != nil {
		return err
	}

	meta := map[string]string{
		"format":      FormatVersion,
		"app_version": appVersion,
		"exported_at": time.Now().UTC().Format(time.RFC3339),
	}
	for key, value := range meta {
		if _, err := conn.Exec("INSERT INTO meta(key, value) VALUES(?, ?)", key, value); err != nil {
			return err
		}
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	insert, err := tx.Prepare(`INSERT INTO shortcuts(
		slug, name, type, kind, url, browser, exec_args, categories,
		binary_path, icon_path, icon_source, desktop_path, desktop_copy_path,
		terminal, menu, desktop, installed_at, binary_data, icon_data)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer insert.Close()

	for _, item := range items {
		_, err := insert.Exec(
			item.Slug, item.Name, item.Type, item.Kind, item.URL, item.Browser,
			strings.Join(item.ExecArgs, "\x00"), item.Categories,
			item.BinaryPath, item.IconPath, item.IconSource, item.DesktopPath,
			item.DesktopCopyPath, boolToInt(item.Terminal), boolToInt(item.Targets.Menu), boolToInt(item.Targets.Desktop),
			item.InstalledAt.Format(time.RFC3339), readBlob(item.BinaryPath), readBlob(item.IconPath),
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// Backup adalah satu shortcut beserta isi berkasnya dari .lazy.
type Backup struct {
	Item   store.Shortcut
	Binary []byte // isi binary/script (kalau ada)
	Icon   []byte // isi berkas ikon (kalau ada)
}

// maxBlobFile adalah batas ukuran berkas yang ikut disimpan ke cadangan.
const maxBlobFile = 512 << 20

// readBlob membaca isi berkas untuk disimpan ke cadangan; gagal/kebesaran = kosong.
func readBlob(path string) []byte {
	if path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() > maxBlobFile {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

// Import membaca daftar shortcut dari berkas .lazy.
func Import(path string) ([]Backup, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var format string
	if err := conn.QueryRow("SELECT value FROM meta WHERE key = 'format'").Scan(&format); err != nil {
		return nil, fmt.Errorf("bukan basis data LazyRun: %w", err)
	}
	switch format {
	case FormatVersion:
		// versi sekarang: ada blobs
	case "1":
		// cadangan lama: hanya metadata, berkas dipulihkan dari sumber lain
	default:
		return nil, fmt.Errorf("versi basis data %s tidak didukung (mendukung %s)", format, FormatVersion)
	}

	rows, err := conn.Query(`SELECT slug, name, type, kind, url, browser, exec_args,
		categories, binary_path, icon_path, icon_source, desktop_path,
		desktop_copy_path, terminal, menu, desktop, installed_at,
		binary_data, icon_data FROM shortcuts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Backup
	for rows.Next() {
		var item store.Shortcut
		var args, installed string
		var terminal, menu, desktop int
		var binary, iconBlob []byte
		if err := rows.Scan(&item.Slug, &item.Name, &item.Type, &item.Kind, &item.URL,
			&item.Browser, &args, &item.Categories, &item.BinaryPath, &item.IconPath,
			&item.IconSource, &item.DesktopPath, &item.DesktopCopyPath,
			&terminal, &menu, &desktop, &installed, &binary, &iconBlob); err != nil {
			return nil, err
		}
		if args != "" {
			item.ExecArgs = strings.Split(args, "\x00")
		}
		item.Terminal = terminal != 0
		item.Targets = store.Targets{Menu: menu != 0, Desktop: desktop != 0}
		if t, err := time.Parse(time.RFC3339, installed); err == nil {
			item.InstalledAt = t
		}
		items = append(items, Backup{Item: item, Binary: binary, Icon: iconBlob})
	}
	return items, rows.Err()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
