package shortcut

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lazyrun/internal/icon"
	"lazyrun/internal/store"
	"lazyrun/internal/xdg"
)

// Restore memasang ulang shortcut dari data cadangan ke mesin ini:
// binary dan ikon ditulis dari isi cadangan (kalau ada), lalu .desktop dibuat
// dengan path lokal. Metadata yang menunjuk ke komputer lain tidak dipakai
// mentah-mentah, jadi hasil restore langsung bisa dipakai.
func Restore(item store.Shortcut, binary, iconData []byte) (store.Shortcut, error) {
	name := strings.TrimSpace(item.Name)
	if name == "" {
		return item, errors.New("nama shortcut kosong")
	}
	sl := slugify(name)
	if sl == "" {
		return item, errors.New("nama shortcut tidak valid")
	}

	restored := item
	restored.Slug = sl
	restored.Name = name
	restored.DesktopCopyPath = ""
	restored.IconPath = ""

	if restored.Type == store.TypeExecutable {
		if err := restoreBinary(&restored, sl, binary); err != nil {
			return restored, err
		}
	}
	if restored.Type == store.TypeWebApp && len(restored.ExecArgs) == 0 && restored.Browser != "" {
		restored.ExecArgs = webAppArgs(restored.Browser, restored.URL)
	}

	iconPath, err := restoreIcon(restored, sl, iconData)
	if err != nil {
		return restored, err
	}
	restored.IconPath = iconPath

	entry := Entry{
		Name:       name,
		ExecArgs:   restored.ExecArgs,
		Icon:       iconPath,
		Categories: restored.Categories,
		Terminal:   restored.Terminal && restored.Kind != KindWindows,
		NoDisplay:  !restored.Targets.Menu,
	}
	if entry.Categories == "" {
		entry.Categories = "Utility;"
	}
	if restored.Kind == KindScript {
		if dir := scriptOriginDir(restored); dir != "" {
			entry.WorkingDir = dir
		}
	}

	deskDst := filepath.Join(xdg.ApplicationsDir(), sl+".desktop")
	if err := WriteDesktop(deskDst, entry); err != nil {
		return restored, err
	}
	restored.DesktopPath = deskDst

	if restored.Targets.Desktop {
		copyPath := filepath.Join(xdg.DesktopDir(), sl+".desktop")
		if err := copyFile(deskDst, copyPath, 0o755); err != nil {
			return restored, err
		}
		restored.DesktopCopyPath = copyPath
	}

	go RefreshCache()
	return restored, nil
}

// restoreBinary menulis binary/script dari cadangan ke ~/.local/bin.
// Kalau cadangan tidak memuat isinya, berkas lokal yang masih ada dipakai.
func restoreBinary(restored *store.Shortcut, slug string, binary []byte) error {
	binDst := filepath.Join(xdg.BinDir(), slug)
	originDir := scriptOriginDir(*restored)

	if len(binary) == 0 {
		// tidak ada isi cadangan: pakai berkas lama kalau masih ada di mesin ini
		if _, err := os.Stat(restored.BinaryPath); err != nil {
			restored.BinaryPath = ""
			restored.ExecArgs = nil
			return nil
		}
		binary, _ = os.ReadFile(restored.BinaryPath)
	}

	if err := writeFileMode(binDst, binary, 0o755); err != nil {
		return err
	}

	kind := DetectKind(binDst)
	if kind == KindScript {
		// script dijalankan lewat wrapper: script aslinya disimpan sebagai .real
		if err := os.Rename(binDst, binDst+scriptSuffix); err != nil {
			return err
		}
		if err := writeScriptWrapper(binDst, binDst+scriptSuffix, originDir); err != nil {
			return err
		}
	}

	restored.Kind = kind
	restored.BinaryPath = binDst
	restored.ExecArgs = []string{binDst}
	if kind == KindWindows {
		if wine, err := exec.LookPath("wine"); err == nil {
			restored.ExecArgs = []string{wine, binDst}
		}
	}
	return nil
}

// restoreIcon memakai ikon dari cadangan; kalau tidak ada, coba unduh ulang
// dari sumber yang tercatat, lalu jatuh ke ikon default.
func restoreIcon(item store.Shortcut, slug string, iconData []byte) (string, error) {
	if len(iconData) > 0 {
		if path, err := icon.Save(iconData, slug); err == nil {
			return path, nil
		}
	}

	source := item.IconSource
	if source != "" {
		options := iconOptions{Slug: slug}
		switch {
		case strings.HasPrefix(source, iconSourceFavicon):
			options.PageURL = strings.TrimPrefix(source, iconSourceFavicon)
		case strings.HasPrefix(source, iconSourceSearch):
			options.Query = strings.TrimPrefix(source, iconSourceSearch)
		case strings.HasPrefix(source, "http"):
			options.IconURL = source
		}
		if path, _, err := installIconFor(options); err == nil && path != "" {
			return path, nil
		}
	}
	return icon.Default(slug)
}

// scriptOriginDir menebak folder asal script dari metadata (dipakai sebagai
// Path= dan folder kerja wrapper).
func scriptOriginDir(item store.Shortcut) string {
	dir := filepath.Dir(item.BinaryPath)
	if dir == "" || dir == "." || dir == "/" {
		return ""
	}
	return dir
}

// webAppArgs menyusun argv PWA dari perintah browser yang tersimpan.
func webAppArgs(browser, pageURL string) []string {
	args := strings.Fields(browser)
	if len(args) == 0 {
		return nil
	}
	if strings.Contains(strings.ToLower(args[0]), "firefox") {
		return append(args, "--new-window", pageURL)
	}
	return append(args, "--app="+pageURL)
}
