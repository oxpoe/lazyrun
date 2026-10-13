package shortcut

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lazyrun/internal/store"
	"lazyrun/internal/xdg"
)

// EditExecutable memperbarui shortcut executable yang sudah ada:
// binary bisa diganti (disalin ulang), begitu juga ikon, nama, dan lokasinya.
func EditExecutable(old store.Shortcut, o ExecOptions) (store.Shortcut, error) {
	name := strings.TrimSpace(o.Name)
	if err := validate(name, o.Menu, o.Desktop); err != nil {
		return old, err
	}

	src := o.BinaryPath
	if src == "" {
		src = old.BinaryPath
	}
	if err := checkRunnable(src); err != nil {
		return old, fmt.Errorf("%s: %w", filepath.Base(src), err)
	}

	sl := slugify(name)
	binDst := filepath.Join(xdg.BinDir(), sl)
	if sl != old.Slug {
		if _, err := os.Stat(filepath.Join(xdg.ApplicationsDir(), sl+".desktop")); err == nil {
			return old, fmt.Errorf("shortcut %q already exists", name)
		}
	}

	// salin kalau sumbernya bukan berkas yang sudah terpasang
	if !samePath(src, binDst) {
		if err := copyFile(src, binDst, 0o755); err != nil {
			return old, err
		}
	}

	kind := DetectKind(binDst)
	execArgs := []string{binDst}
	if kind == KindWindows {
		wine, err := exec.LookPath("wine")
		if err != nil {
			return old, errors.New("this is a Windows .exe; install wine first so the shortcut can run it")
		}
		execArgs = []string{wine, binDst}
	}

	iconPath, iconSource := old.IconPath, old.IconSource
	if o.IconPath != "" || o.IconURL != "" || len(o.IconData) > 0 {
		path, source, err := installIconFor(iconOptions{
			IconPath: o.IconPath, IconURL: o.IconURL, IconData: o.IconData,
			Query: name, Slug: sl,
		})
		if err != nil {
			return old, err
		}
		iconPath, iconSource = path, source
	}

	category := category(o.Category, "Utility;")
	if o.Category == "" {
		category = old.Categories
		if category == "" {
			category = "Utility;"
		}
	}

	updated := old
	updated.Slug = sl
	updated.Name = name
	updated.Kind = kind
	updated.ExecArgs = execArgs
	updated.Categories = category
	updated.BinaryPath = binDst
	updated.IconPath = iconPath
	updated.IconSource = iconSource
	updated.Targets = store.Targets{Menu: o.Menu, Desktop: o.Desktop}

	if err := rewriteShortcutFiles(old, &updated, Entry{
		Name:       name,
		ExecArgs:   execArgs,
		Icon:       iconPath,
		Categories: category,
		NoDisplay:  !o.Menu,
	}); err != nil {
		return old, err
	}
	return updated, nil
}

// EditWebApp memperbarui shortcut web app (URL, browser, ikon, nama, lokasi).
func EditWebApp(old store.Shortcut, o WebOptions) (store.Shortcut, error) {
	name := strings.TrimSpace(o.Name)
	if err := validate(name, o.Menu, o.Desktop); err != nil {
		return old, err
	}
	if len(o.Browser.Exec) == 0 {
		return old, errors.New("no browser selected")
	}
	pageURL, err := normalizeURL(o.URL)
	if err != nil {
		return old, err
	}

	sl := slugify(name)
	if sl != old.Slug {
		if _, err := os.Stat(filepath.Join(xdg.ApplicationsDir(), sl+".desktop")); err == nil {
			return old, fmt.Errorf("shortcut %q already exists", name)
		}
	}

	iconPath, iconSource := old.IconPath, old.IconSource
	if o.IconPath != "" || o.IconURL != "" || len(o.IconData) > 0 {
		path, source, err := installIconFor(iconOptions{
			IconPath: o.IconPath, IconURL: o.IconURL, IconData: o.IconData,
			PageURL: pageURL, Query: name, Slug: sl,
		})
		if err != nil {
			return old, err
		}
		iconPath, iconSource = path, source
	}

	appArgs := o.Browser.Exec[1:]
	if o.Browser.Firefox {
		appArgs = append(appArgs, "--new-window", pageURL)
	} else {
		appArgs = append(appArgs, "--app="+pageURL)
	}
	execArgs := append([]string{o.Browser.Exec[0]}, appArgs...)

	updated := old
	updated.Slug = sl
	updated.Name = name
	updated.Kind = KindPWA
	updated.URL = pageURL
	updated.Browser = strings.Join(o.Browser.Exec, " ")
	updated.ExecArgs = execArgs
	updated.IconPath = iconPath
	updated.IconSource = iconSource
	updated.Targets = store.Targets{Menu: o.Menu, Desktop: o.Desktop}

	if err := rewriteShortcutFiles(old, &updated, Entry{
		Name:       name,
		ExecArgs:   execArgs,
		Icon:       iconPath,
		Categories: old.Categories,
		NoDisplay:  !o.Menu,
	}); err != nil {
		return old, err
	}
	return updated, nil
}

// rewriteShortcutFiles menulis ulang .desktop (dan salinannya) untuk shortcut
// yang diedit, termasuk membereskan berkas lama kalau nama/slug berubah.
// updated diisi lewat pointer karena path-nya berubah di sini.
func rewriteShortcutFiles(old store.Shortcut, updated *store.Shortcut, entry Entry) error {
	if entry.Categories == "" {
		entry.Categories = "Utility;"
	}

	newDesk := filepath.Join(xdg.ApplicationsDir(), updated.Slug+".desktop")
	if err := WriteDesktop(newDesk, entry); err != nil {
		return err
	}
	updated.DesktopPath = newDesk

	// salinan Desktop: dibuat kalau diminta, dihapus kalau tidak
	newCopy := ""
	if updated.Targets.Desktop {
		newCopy = filepath.Join(xdg.DesktopDir(), updated.Slug+".desktop")
		if err := copyFile(newDesk, newCopy, 0o755); err != nil {
			return err
		}
	}
	updated.DesktopCopyPath = newCopy

	// bersihkan berkas lama kalau lokasinya berubah
	if old.DesktopPath != "" && old.DesktopPath != newDesk {
		_ = os.Remove(old.DesktopPath)
	}
	if old.DesktopCopyPath != "" && old.DesktopCopyPath != newCopy {
		_ = os.Remove(old.DesktopCopyPath)
	}
	if old.IconPath != "" && old.IconPath != updated.IconPath {
		_ = os.Remove(old.IconPath)
	}
	// binary lama dibuang kalau nama berkasnya berubah
	if old.Type == store.TypeExecutable && old.BinaryPath != "" && old.BinaryPath != updated.BinaryPath {
		if !samePath(old.BinaryPath, updated.BinaryPath) {
			_ = os.Remove(old.BinaryPath)
		}
	}

	go RefreshCache()
	return nil
}
