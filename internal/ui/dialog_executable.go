package ui

import (
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/internal/icon"
	"lazyrun/internal/shortcut"
	"lazyrun/internal/store"
)

var categories = []string{
	"Utility", "Network", "Development", "Graphics",
	"AudioVideo", "Office", "Game", "Education", "System", "Settings",
}

var (
	imageFilter = storage.NewExtensionFileFilter([]string{".png", ".svg", ".jpg", ".jpeg", ".webp", ".gif", ".bmp", ".ico", ".xpm"})
	stripExts   = []string{".AppImage", ".appimage", ".sh", ".bin", ".run", ".desktop"}
)

func (ui *app) showExecutableDialog() {
	var binPath, iconPath, iconURL string
	var iconData []byte

	binLabel := muted(T("none"))
	iconLabel := muted(T("icon_auto"))
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Winbox")
	category := widget.NewSelect(categories, nil)
	category.SetSelected("Utility")
	menu := widget.NewCheck(T("check_menu"), nil)
	menu.SetChecked(true)
	desktop := widget.NewCheck(T("check_desktop"), nil)
	desktop.SetChecked(true)

	browseBin := widget.NewButtonWithIcon(T("browse"), theme.FolderOpenIcon(), func() {
		pickFile(ui.win, nil, func(path string) {
			binPath = path
			binLabel.SetText(filepath.Base(path))
			if strings.TrimSpace(nameEntry.Text) == "" {
				nameEntry.SetText(appNameFromFile(path))
			}
		})
	})
	browseIcon := widget.NewButtonWithIcon(T("browse"), theme.FolderOpenIcon(), func() {
		pickFile(ui.win, imageFilter, func(path string) {
			iconPath, iconURL, iconData = path, "", nil
			iconLabel.SetText(filepath.Base(path))
		})
	})
	onlineIcon := widget.NewButtonWithIcon(T("online"), theme.SearchIcon(), func() {
		ui.showIconSearch(nameEntry.Text, func(result icon.Result, data []byte) {
			iconURL, iconPath, iconData = result.URL, "", data
			iconLabel.SetText(result.Name)
		})
	})

	var d *dialog.CustomDialog
	install := widget.NewButtonWithIcon(T("install"), theme.DownloadIcon(), func() {
		switch {
		case binPath == "":
			dialog.ShowInformation(T("err_title"), T("need_binary"), ui.win)
			return
		case strings.TrimSpace(nameEntry.Text) == "":
			dialog.ShowInformation(T("err_title"), T("need_name"), ui.win)
			return
		case !menu.Checked && !desktop.Checked:
			dialog.ShowInformation(T("err_title"), T("need_target"), ui.win)
			return
		}
		opts := shortcut.ExecOptions{
			BinaryPath: binPath,
			IconPath:   iconPath,
			IconURL:    iconURL,
			IconData:   iconData,
			Name:       nameEntry.Text,
			Category:   category.Selected,
			Menu:       menu.Checked,
			Desktop:    desktop.Checked,
		}
		ui.install(
			func() (store.Shortcut, error) { return shortcut.InstallExecutable(opts) },
			func(sc store.Shortcut, err error) { ui.added(sc, err, d.Hide) },
		)
	})
	install.Importance = widget.HighImportance

	form := widget.NewForm(
		widget.NewFormItem(T("field_binary"), container.NewBorder(nil, nil, nil, browseBin, binLabel)),
		widget.NewFormItem(T("field_icon"), container.NewBorder(nil, nil, nil,
			container.NewHBox(onlineIcon, browseIcon), iconLabel)),
		widget.NewFormItem(T("field_name"), nameEntry),
		widget.NewFormItem(T("field_category"), category),
	)
	content := container.NewVScroll(container.NewVBox(
		closeRow(func() { d.Hide() }),
		form,
		widget.NewSeparator(),
		sectionTitle(T("section_location")),
		menu,
		desktop,
		muted(T("exec_hint")),
		container.NewHBox(logoBox(newLogo(), 20), spacer(),
			widget.NewButton(T("cancel"), func() { d.Hide() }),
			install),
	))

	d = dialog.NewCustomWithoutButtons(T("new_exec_title"), content, ui.win)
	showDialog(d, ui.win, 560, 520)
}

// pickFile membuka file picker dan mengirim path absolut ke onPick.
func pickFile(win fyne.Window, filter storage.FileFilter, onPick func(path string)) {
	fd := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		if r == nil {
			return
		}
		defer r.Close()
		onPick(r.URI().Path())
	}, win)
	if filter != nil {
		fd.SetFilter(filter)
	}
	fd.Show()
}

func appNameFromFile(path string) string {
	base := filepath.Base(path)
	for _, ext := range stripExts {
		if strings.HasSuffix(base, ext) {
			return strings.TrimSuffix(base, ext)
		}
	}
	return base
}
