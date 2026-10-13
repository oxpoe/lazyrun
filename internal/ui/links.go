package ui

import (
	"errors"
	"net/url"
	"os"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"lazyrun/internal/store"
)

// Link aplikasi: repo GitHub LazyRun dan halaman donasi.
const (
	repoURL   = "https://github.com/oxpoe/lazyrun"
	donateURL = "https://donasi.oxpoe.qzz.io/"
)

// openURL membuka link di browser default.
func (ui *app) openURL(raw string) {
	u, err := url.Parse(raw)
	if err == nil {
		err = fyne.CurrentApp().OpenURL(u)
	}
	if err != nil {
		dialog.ShowError(err, ui.win)
	}
}

// launch menjalankan shortcut.
//
// Kalau ada, `gio launch <file .desktop>` dipakai supaya aturannya sama dengan
// menu desktop: Terminal=true membuka jendela terminal (script yang butuh sudo
// atau input bisa jalan) dan Path= ikut dihormati.
func (ui *app) launch(sc store.Shortcut) {
	args, err := launchCommand(sc)
	if err != nil {
		dialog.ShowError(err, ui.win)
		return
	}

	cmd := exec.Command(args[0], args[1:]...)
	if err := cmd.Start(); err != nil {
		dialog.ShowError(err, ui.win)
		return
	}
	go cmd.Wait() // supaya proses anak tidak jadi zombie
}

// launchCommand menyusun perintah untuk menjalankan shortcut.
func launchCommand(sc store.Shortcut) ([]string, error) {
	if sc.DesktopPath != "" {
		if gio, err := exec.LookPath("gio"); err == nil {
			if _, err := os.Stat(sc.DesktopPath); err == nil {
				return []string{gio, "launch", sc.DesktopPath}, nil
			}
		}
	}
	switch {
	case len(sc.ExecArgs) > 0:
		return sc.ExecArgs, nil
	case sc.Type == store.TypeWebApp && sc.URL != "":
		return []string{"xdg-open", sc.URL}, nil
	case sc.BinaryPath != "":
		return []string{sc.BinaryPath}, nil
	}
	return nil, errors.New("shortcut ini tidak punya perintah yang bisa dijalankan")
}
