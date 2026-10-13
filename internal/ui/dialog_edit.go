package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/internal/browser"
	"lazyrun/internal/db"
	"lazyrun/internal/icon"
	"lazyrun/internal/shortcut"
	"lazyrun/internal/store"
)

var lazyFilter = storage.NewExtensionFileFilter([]string{".lazy"})

// exportShortcuts menyimpan daftar shortcut ke satu berkas SQLite (.lazy),
// mis. "db.lazy", supaya bisa dipindah ke komputer lain.
func (ui *app) exportShortcuts() {
	save := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.win)
			return
		}
		if w == nil {
			return
		}
		path := w.URI().Path()
		w.Close() // berkas kosong dari dialog akan diganti basis data
		if !strings.HasSuffix(strings.ToLower(path), ".lazy") {
			path += ".lazy"
		}

		if err := db.Export(path, ui.version, ui.store.Items); err != nil {
			dialog.ShowError(err, ui.win)
			return
		}
		dialog.ShowInformation(T("export"), fmt.Sprintf(T("export_done"), len(ui.store.Items), path), ui.win)
	}, ui.win)
	save.SetFileName(db.DefaultName)
	save.SetFilter(lazyFilter)
	save.Show()
}

// importShortcuts menambahkan shortcut dari berkas .lazy tanpa menimpa yang sudah ada.
func (ui *app) importShortcuts() {
	open := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.win)
			return
		}
		if r == nil {
			return
		}
		path := r.URI().Path()
		r.Close()

		items, err := db.Import(path)
		if err != nil {
			dialog.ShowError(err, ui.win)
			return
		}
		if len(items) == 0 {
			dialog.ShowInformation(T("import"), T("import_empty"), ui.win)
			return
		}

		added, skipped := ui.restoreImported(items)
		if err := ui.store.Save(); err != nil {
			dialog.ShowError(err, ui.win)
			return
		}
		ui.reload()
		dialog.ShowInformation(T("import"), fmt.Sprintf(T("import_done"), added, skipped), ui.win)
	}, ui.win)
	open.SetFilter(lazyFilter)
	open.Show()
}

// showEditDialog mengubah shortcut yang sudah ada: binary/URL, ikon, nama, lokasi.
func (ui *app) showEditDialog(sc store.Shortcut) {
	isWeb := sc.Type == store.TypeWebApp

	var iconPath, iconURL, binPath string
	var iconData []byte

	nameEntry := widget.NewEntry()
	nameEntry.SetText(sc.Name)
	iconLabel := muted(iconLabelOf(sc))
	menu := widget.NewCheck(T("check_menu"), nil)
	menu.SetChecked(sc.Targets.Menu)
	desktopChk := widget.NewCheck(T("check_desktop"), nil)
	desktopChk.SetChecked(sc.Targets.Desktop)
	terminal := widget.NewCheck(T("check_terminal"), nil)
	terminal.SetChecked(sc.Terminal)
	if sc.Kind != shortcut.KindScript {
		terminal.Hide()
	}

	browseIcon := newButton(T("browse"), theme.FolderOpenIcon(), T("tip_browse"), widget.MediumImportance, func() {
		pickFile(ui.win, imageFilter, func(path string) {
			iconPath, iconURL, iconData = path, "", nil
			iconLabel.SetText(filepath.Base(path))
		})
	})
	onlineIcon := newButton(T("online"), theme.SearchIcon(), T("tip_online"), widget.MediumImportance, func() {
		ui.showIconSearch(nameEntry.Text, func(result icon.Result, data []byte) {
			iconURL, iconPath, iconData = result.URL, "", data
			iconLabel.SetText(result.Name)
		})
	})

	// widget khusus executable
	binLabel := muted(sc.BinaryPath)
	browseBin := newButton(T("browse"), theme.FolderOpenIcon(), T("tip_browse"), widget.MediumImportance, func() {
		pickFile(ui.win, nil, func(path string) {
			binPath = path
			binLabel.SetText(filepath.Base(path) + " · " + detectedKind(path))
			if shortcut.DetectKind(path) == shortcut.KindScript {
				terminal.Show()
			} else {
				terminal.SetChecked(false)
				terminal.Hide()
			}
		})
	})
	category := widget.NewSelect(categories, nil)
	category.SetSelected(categoryOf(sc))

	// widget khusus web app
	urlEntry := widget.NewEntry()
	urlEntry.SetText(sc.URL)
	browserSel := widget.NewSelect(browserNames(ui.browsers), nil)
	browserSel.SetSelected(currentBrowser(ui.browsers, sc.Browser))

	iconItem := widget.NewFormItem(T("field_icon"), container.NewBorder(nil, nil, nil,
		container.NewHBox(onlineIcon, browseIcon), iconLabel))

	items := []*widget.FormItem{}
	if isWeb {
		items = append(items,
			widget.NewFormItem(T("field_url"), urlEntry),
			widget.NewFormItem(T("field_name"), nameEntry),
			iconItem,
			widget.NewFormItem(T("field_browser"), browserSel),
		)
	} else {
		items = append(items,
			widget.NewFormItem(T("field_binary"), container.NewBorder(nil, nil, nil, browseBin, binLabel)),
			widget.NewFormItem(T("field_name"), nameEntry),
			iconItem,
			widget.NewFormItem(T("field_category"), category),
		)
	}

	var d *dialog.CustomDialog
	save := newButton(T("save"), theme.DocumentSaveIcon(), T("tip_save"), widget.HighImportance, func() {
		switch {
		case strings.TrimSpace(nameEntry.Text) == "":
			dialog.ShowInformation(T("err_title"), T("need_name"), ui.win)
			return
		case !menu.Checked && !desktopChk.Checked:
			dialog.ShowInformation(T("err_title"), T("need_target"), ui.win)
			return
		}

		ui.install(func() (store.Shortcut, error) {
			if isWeb {
				b := ui.browsers[indexOf(ui.browsers, browserSel.Selected)]
				return shortcut.EditWebApp(sc, shortcut.WebOptions{
					URL:      urlEntry.Text,
					Name:     nameEntry.Text,
					Browser:  shortcut.BrowserArg{Name: b.Name, Exec: b.Exec, Firefox: b.Kind == browser.Firefox},
					IconPath: iconPath,
					IconURL:  iconURL,
					IconData: iconData,
					Menu:     menu.Checked,
					Desktop:  desktopChk.Checked,
				})
			}
			path := binPath
			if path == "" {
				path = sc.BinaryPath
			}
			return shortcut.EditExecutable(sc, shortcut.ExecOptions{
				BinaryPath: path,
				IconPath:   iconPath,
				IconURL:    iconURL,
				IconData:   iconData,
				Name:       nameEntry.Text,
				Category:   category.Selected,
				Terminal:   terminal.Checked,
				Menu:       menu.Checked,
				Desktop:    desktopChk.Checked,
			})
		}, func(updated store.Shortcut, err error) {
			if err != nil {
				dialog.ShowError(err, ui.win)
				return
			}
			ui.store.Add(updated)
			if err := ui.store.Save(); err != nil {
				dialog.ShowError(err, ui.win)
				return
			}
			d.Hide()
			ui.reload()
			if card, ok := ui.cards[updated.Slug]; ok {
				card.flash(0)
			}
			dialog.ShowInformation(T("edit_title"), fmt.Sprintf(T("edit_saved"), updated.Name), ui.win)
		})
	})

	hint := textLine(T("edit_hint"), 12, false, colorNameMuted)
	content := container.NewVScroll(container.NewVBox(
		closeRow(func() { d.Hide() }),
		widget.NewForm(items...),
		widget.NewSeparator(),
		sectionTitle(T("section_location")),
		menu,
		desktopChk,
		terminal,
		hint,
		container.NewHBox(logoBox(newLogo(), 20), spacer(),
			newButton(T("cancel"), nil, T("tip_cancel"), widget.MediumImportance, func() { d.Hide() }),
			save),
	))

	d = dialog.NewCustomWithoutButtons(T("edit_title"), content, ui.win)
	d.SetOnClosed(pushHintTarget(hint))
	showDialog(d, ui.win, 560, 520)
}

// restoreImported memasang ulang setiap entri cadangan ke mesin ini
// (binary, ikon, .desktop) lalu menambahkannya ke daftar, tanpa menimpa
// shortcut yang sudah ada.
func (ui *app) restoreImported(items []db.Backup) (added, skipped int) {
	for _, backup := range items {
		item := backup.Item
		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" {
			skipped++
			continue
		}
		if _, exists := ui.store.Get(shortcut.Slug(item.Name)); exists {
			skipped++
			continue
		}

		restored, err := shortcut.Restore(item, backup.Binary, backup.Icon)
		if err != nil {
			skipped++
			continue
		}
		ui.store.Add(restored)
		added++
	}
	return added, skipped
}

// iconLabelOf menampilkan ikon yang sedang dipakai shortcut.
func iconLabelOf(sc store.Shortcut) string {
	if sc.IconPath == "" {
		return T("icon_auto")
	}
	return filepath.Base(sc.IconPath)
}

// currentBrowser mengembalikan nama browser yang cocok dengan yang tersimpan.
func currentBrowser(list []browser.Browser, saved string) string {
	for _, b := range list {
		if strings.Join(b.Exec, " ") == saved {
			return b.Name
		}
	}
	if len(list) > 0 {
		return list[0].Name
	}
	return ""
}

// categoryOf mengambil kategori pertama yang tersimpan.
func categoryOf(sc store.Shortcut) string {
	first, _, _ := strings.Cut(strings.TrimSuffix(sc.Categories, ";"), ";")
	for _, c := range categories {
		if c == first {
			return c
		}
	}
	return "Utility"
}
