package ui

import (
	"fmt"
	"os"
	"os/exec"

	"fyne.io/fyne/v2/dialog"
)

// checkUpdates menjalankan AppImageUpdate kalau LazyRun berjalan sebagai AppImage.
func (ui *app) checkUpdates() {
	image := os.Getenv("APPIMAGE")
	if image == "" {
		dialog.ShowInformation(T("check_updates"), fmt.Sprintf(T("not_appimage"), ui.version), ui.win)
		return
	}
	update, err := exec.LookPath("AppImageUpdate")
	if err != nil {
		dialog.ShowInformation(T("check_updates"), T("no_appimageupdate"), ui.win)
		return
	}
	if err := exec.Command(update, image).Start(); err != nil {
		dialog.ShowError(err, ui.win)
	}
}
