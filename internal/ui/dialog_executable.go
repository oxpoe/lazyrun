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

	"lazyrun/internal/shortcut"
	"lazyrun/internal/store"
)

var categories = []string{
	"Utility", "Network", "Development", "Graphics",
	"AudioVideo", "Office", "Game", "Education", "System", "Settings",
}

var (
	imageFilter = storage.NewExtensionFileFilter([]string{".png", ".svg", ".jpg", ".jpeg", ".xpm", ".ico"})
	stripExts   = []string{".AppImage", ".appimage", ".sh", ".bin", ".run", ".desktop"}
)

func (ui *app) showExecutableDialog() {
	var binPath, iconPath string

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
			iconPath = path
			iconLabel.SetText(filepath.Base(path))
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
		widget.NewFormItem(T("field_icon"), container.NewBorder(nil, nil, nil, browseIcon, iconLabel)),
		widget.NewFormItem(T("field_name"), nameEntry),
		widget.NewFormItem(T("field_category"), category),
	)
	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		sectionTitle(T("section_location")),
		menu,
		desktop,
		muted(T("exec_hint")),
		container.NewHBox(logoBox(newLogo(), 20), spacer(),
			widget.NewButton(T("cancel"), func() { d.Hide() }),
			install),
	)

	d = dialog.NewCustom(T("new_exec_title"), "", content, ui.win)
	d.Resize(fyne.NewSize(560, 470))
	d.Show()
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
