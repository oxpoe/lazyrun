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
const FormatVersion = "1"

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
	menu              INTEGER NOT NULL,
	desktop           INTEGER NOT NULL,
	installed_at      TEXT
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
		menu, desktop, installed_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
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
			item.DesktopCopyPath, boolToInt(item.Targets.Menu), boolToInt(item.Targets.Desktop),
			item.InstalledAt.Format(time.RFC3339),
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// Import membaca daftar shortcut dari berkas .lazy.
func Import(path string) ([]store.Shortcut, error) {
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
	if format != FormatVersion {
		return nil, fmt.Errorf("versi basis data %s tidak didukung (mendukung %s)", format, FormatVersion)
	}

	rows, err := conn.Query(`SELECT slug, name, type, kind, url, browser, exec_args,
		categories, binary_path, icon_path, icon_source, desktop_path,
		desktop_copy_path, menu, desktop, installed_at FROM shortcuts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []store.Shortcut
	for rows.Next() {
		var item store.Shortcut
		var args, installed string
		var menu, desktop int
		if err := rows.Scan(&item.Slug, &item.Name, &item.Type, &item.Kind, &item.URL,
			&item.Browser, &args, &item.Categories, &item.BinaryPath, &item.IconPath,
			&item.IconSource, &item.DesktopPath, &item.DesktopCopyPath,
			&menu, &desktop, &installed); err != nil {
			return nil, err
		}
		if args != "" {
			item.ExecArgs = strings.Split(args, "\x00")
		}
		item.Targets = store.Targets{Menu: menu != 0, Desktop: desktop != 0}
		if t, err := time.Parse(time.RFC3339, installed); err == nil {
			item.InstalledAt = t
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
