package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/internal/update"
)

// recheckInterval adalah jeda pemeriksaan update berkala selama aplikasi terbuka.
const recheckInterval = 4 * time.Hour

// appImagePath mengembalikan lokasi AppImage yang sedang dipakai.
// $APPIMAGE dipakai kalau ada; kalau tidak (mis. dijalankan lewat wrapper),
// berkas executable-nya sendiri diperiksa apakah benar AppImage.
func appImagePath() string {
	if path := os.Getenv("APPIMAGE"); path != "" {
		return path
	}
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if update.IsAppImage(exe) {
		return exe
	}
	return ""
}

// checkUpdates dipanggil tombol footer: hasilnya selalu diberitahukan.
func (ui *app) checkUpdates() { ui.checkForUpdate(true) }

// scheduleUpdateCheck memeriksa rilis terbaru setelah aplikasi tampil,
// lalu mengulanginya berkala selama aplikasi terbuka.
func (ui *app) scheduleUpdateCheck() {
	if appImagePath() == "" {
		return // dijalankan dari source / hasil ekstrak: tidak bisa memasang update
	}
	time.AfterFunc(2*time.Second, func() { ui.checkForUpdate(false) })
	ticker := time.NewTicker(recheckInterval)
	go func() {
		for range ticker.C {
			ui.checkForUpdate(false)
		}
	}()
}

// checkForUpdate mengambil rilis terbaru dari GitHub. manual=false hanya
// memberitahu kalau memang ada versi baru.
func (ui *app) checkForUpdate(manual bool) {
	if ui.version == "" || ui.version == "dev" {
		if manual {
			dialog.ShowInformation(T("check_updates"), fmt.Sprintf(T("dev_build"), ui.version), ui.win)
		}
		return
	}

	go func() {
		rel, err := update.Latest()
		fyne.Do(func() {
			switch {
			case err != nil:
				if manual {
					dialog.ShowError(err, ui.win)
				}
			case update.Newer(ui.version, rel):
				ui.markUpdatePending(rel)
				ui.showUpdateDialog(rel)
			case manual:
				dialog.ShowInformation(T("check_updates"), fmt.Sprintf(T("up_to_date"), ui.version), ui.win)
			}
		})
	}()
}

// markUpdatePending menampilkan versi baru di footer, jadi tetap terlihat
// walau dialognya ditutup dengan "Nanti".
func (ui *app) markUpdatePending(rel update.Release) {
	if ui.versionText == nil {
		return
	}
	ui.versionText.Text = fmt.Sprintf("v%s → v%s", ui.version, rel.Version)
	ui.versionText.Refresh()
}

// showUpdateDialog menawarkan update: unduh sekarang, buka halaman rilis, atau nanti.
func (ui *app) showUpdateDialog(rel update.Release) {
	// catatan rilis GitHub berformat markdown; render supaya blok kode/daftar
	// tidak tampil sebagai teks mentah
	notes := widget.NewRichTextFromMarkdown(rel.Notes)

	// catatan rilis di tengah (bukan top): VScroll di Fyne punya MinSize kecil,
	// jadi kalau dipasang sebagai top dari Border isinya terpotong.
	head := heading(fmt.Sprintf(T("update_available"), rel.Version, ui.version))
	body := container.NewVScroll(container.NewVBox(muted(T("update_hint")), notes))

	var d *dialog.CustomDialog
	buttons := container.NewHBox(
		logoBox(newLogo(), 20),
		spacer(),
		widget.NewButton(T("update_later"), func() { d.Hide() }),
		widget.NewButton(T("update_page"), func() { ui.openURL(rel.PageURL) }),
	)

	if updates := ui.updaterBinary(); updates != "" {
		// AppImageUpdate yang memasang (dan memverifikasi) sendiri
		buttons.Add(widget.NewButtonWithIcon(T("update_now"), theme.DownloadIcon(), func() {
			d.Hide()
			ui.runUpdater(updates, rel)
		}))
	} else {
		buttons.Add(widget.NewButtonWithIcon(T("update_now"), theme.DownloadIcon(), func() {
			d.Hide()
			ui.downloadUpdate(rel)
		}))
	}

	d = dialog.NewCustomWithoutButtons(T("check_updates"), container.NewBorder(head, buttons, nil, nil, body), ui.win)
	showDialog(d, ui.win, 560, 440)
}

// updaterBinary mengembalikan path AppImageUpdate kalau terpasang.
func (ui *app) updaterBinary() string {
	if appImagePath() == "" {
		return ""
	}
	path, err := exec.LookPath("AppImageUpdate")
	if err != nil {
		return ""
	}
	return path
}

// downloadUpdate mengunduh AppImage terbaru dan menukarnya dengan yang sedang dipakai.
func (ui *app) downloadUpdate(rel update.Release) {
	appImage := appImagePath()
	if appImage == "" {
		ui.openURL(rel.PageURL)
		return
	}

	progress := dialog.NewProgressInfinite(T("check_updates"), T("update_downloading"), ui.win)
	progress.Show()
	go func() {
		err := update.Download(rel, appImage)
		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(err, ui.win)
				return
			}
			ui.offerRestart()
		})
	}()
}

// runUpdater memakai AppImageUpdate untuk memasang update (pakai zsync, unduhannya kecil).
func (ui *app) runUpdater(updater string, rel update.Release) {
	appImage := appImagePath()
	if appImage == "" {
		ui.openURL(rel.PageURL)
		return
	}

	progress := dialog.NewProgressInfinite(T("check_updates"), T("update_downloading"), ui.win)
	progress.Show()
	go func() {
		err := exec.Command(updater, appImage).Run()
		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(err, ui.win)
				return
			}
			ui.offerRestart()
		})
	}()
}

// offerRestart menawarkan restart supaya versi baru yang dipakai.
func (ui *app) offerRestart() {
	var d *dialog.CustomDialog
	content := container.NewVBox(
		widget.NewLabel(T("update_installed")),
		container.NewHBox(spacer(),
			widget.NewButton(T("update_later"), func() { d.Hide() }),
			widget.NewButtonWithIcon(T("update_restart"), theme.ViewRefreshIcon(), ui.restart),
		),
	)
	d = dialog.NewCustomWithoutButtons(T("check_updates"), content, ui.win)
	showDialog(d, ui.win, 460, 200)
}

// restart menjalankan AppImage yang baru lalu menutup proses lama.
func (ui *app) restart() {
	appImage := appImagePath()
	if appImage == "" {
		return
	}
	cmd := exec.Command(appImage)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		dialog.ShowError(err, ui.win)
		return
	}
	ui.win.Close()
}
