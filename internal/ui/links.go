package ui

import (
	"net/url"
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

// launch menjalankan shortcut: pakai argv hasil install, fallback ke binary/URL.
func (ui *app) launch(sc store.Shortcut) {
	args := sc.ExecArgs
	if len(args) == 0 {
		switch {
		case sc.Type == store.TypeWebApp && sc.URL != "":
			args = []string{"xdg-open", sc.URL}
		case sc.BinaryPath != "":
			args = []string{sc.BinaryPath}
		}
	}
	if len(args) == 0 {
		return
	}

	cmd := exec.Command(args[0], args[1:]...)
	if err := cmd.Start(); err != nil {
		dialog.ShowError(err, ui.win)
		return
	}
	go cmd.Wait() // supaya proses anak tidak jadi zombie
}
