// Package ui berisi window utama dan dialog LazyRun.
package ui

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/assets"
	"lazyrun/internal/browser"
	"lazyrun/internal/shortcut"
	"lazyrun/internal/store"
	"lazyrun/internal/xdg"
)

type app struct {
	win      fyne.Window
	store    *store.Store
	browsers []browser.Browser
	rows     *fyne.Container
	version  string
}

// Run menampilkan window utama LazyRun sampai user menutupnya.
func Run(a fyne.App, version string) {
	ui, loadErr := build(a, version)
	ui.win.Show()
	if loadErr != nil {
		dialog.ShowError(loadErr, ui.win)
	}
	a.Run()
}

// build menyiapkan window utama tanpa menampilkannya (dipakai Run dan tes headless).
func build(a fyne.App, version string) (*app, error) {
	a.Settings().SetTheme(brand{theme.DefaultTheme()})

	ui := &app{version: version}
	a.SetIcon(fyne.NewStaticResource("lazyrun.png", assets.Icon))

	ui.browsers = browser.Detect()
	ui.rows = container.NewVBox()
	ui.win = a.NewWindow(T("app_title"))
	ui.win.Resize(fyne.NewSize(840, 600))

	st, err := store.Load(xdg.MetadataPath())
	ui.store = st

	addExec := widget.NewButtonWithIcon(T("btn_executable"), theme.ContentAddIcon(), ui.showExecutableDialog)
	addExec.Importance = widget.HighImportance
	addWeb := widget.NewButtonWithIcon(T("btn_webapp"), theme.ContentAddIcon(), ui.showWebAppDialog)

	leftHeader := container.NewHBox(logoImage(40), container.NewVBox(
		heading("LazyRun"),
		caption(T("tagline")),
	))
	rightHeader := container.NewHBox(addExec, addWeb)
	header := container.NewBorder(nil, nil, leftHeader, rightHeader)

	update := widget.NewButtonWithIcon(T("check_updates"), theme.DownloadIcon(), ui.checkUpdates)
	footer := container.NewBorder(
		widget.NewSeparator(), nil,
		update,
		caption("v"+version),
	)

	top := container.NewVBox(header, widget.NewSeparator(), sectionLabel(T("installed")))
	body := container.NewBorder(
		top,
		nil, nil, nil,
		container.NewVScroll(ui.rows))

	content := container.NewBorder(
		nil,
		container.NewPadded(footer),
		nil, nil,
		container.NewPadded(body),
	)
	ui.win.SetContent(content)
	ui.reload()
	return ui, err
}

func (ui *app) reload() {
	if st, err := store.Load(ui.store.Path); err == nil {
		ui.store = st
	}
	ui.rows.RemoveAll()
	if len(ui.store.Items) == 0 {
		ui.rows.Add(emptyState())
	}
	for _, sc := range ui.store.Items {
		ui.rows.Add(ui.row(sc))
	}
}

func (ui *app) row(sc store.Shortcut) fyne.CanvasObject {
	info := container.NewVBox(
		subheading(sc.Name),
		caption(metaText(sc)),
	)

	chips := make([]fyne.CanvasObject, 0, 3)
	if sc.Targets.Menu {
		chips = append(chips, chip(T("chip_menu"), colorNameChipMenu))
	}
	if sc.Targets.Desktop {
		chips = append(chips, chip(T("chip_desktop"), colorNameChipDesk))
	}
	if sc.Broken() {
		chips = append(chips, chip(T("chip_broken"), colorNameChipBroken))
	}

	actions := container.NewHBox(chips...)
	actions.Add(iconButton(theme.FolderOpenIcon(), func() {
		if err := exec.Command("xdg-open", filepath.Dir(sc.DesktopPath)).Start(); err != nil {
			dialog.ShowError(err, ui.win)
		}
	}))
	actions.Add(iconButton(theme.DeleteIcon(), func() { ui.confirmDelete(sc) }))

	return container.NewVBox(
		card(container.NewBorder(nil, nil,
			container.NewHBox(thumb(sc.IconPath, fallbackIcon(sc)), info),
			actions)),
	)
}

// metaText menampilkan tipe shortcut dan penanda singkatnya.
func metaText(sc store.Shortcut) string {
	if sc.Type == store.TypeWebApp {
		if host := hostOf(sc.URL); host != "" {
			return "Web · " + host
		}
		return "Web"
	}
	if sc.BinaryPath != "" {
		return "Executable · " + filepath.Base(sc.BinaryPath)
	}
	return "Executable"
}

func fallbackIcon(sc store.Shortcut) fyne.Resource {
	if sc.Type == store.TypeWebApp {
		return theme.ComputerIcon()
	}
	return theme.FileApplicationIcon()
}

func emptyState() fyne.CanvasObject {
	hint := muted(T("empty_hint"))
	hint.Alignment = fyne.TextAlignCenter
	return container.NewCenter(container.NewVBox(
		container.NewCenter(logoImage(88)),
		subheading(T("empty_title")),
		hint,
	))
}

func (ui *app) confirmDelete(sc store.Shortcut) {
	content := container.NewVBox(widget.NewLabel(fmt.Sprintf(T("confirm_delete_msg"), sc.Name)))

	var removeBinary *widget.Check
	if sc.Type == store.TypeExecutable && sc.BinaryPath != "" {
		removeBinary = widget.NewCheck(T("also_binary"), nil)
		content.Add(removeBinary)
	}

	d := dialog.NewCustomConfirm(T("confirm_delete_title"), T("delete"), T("cancel"), content, func(ok bool) {
		if !ok {
			return
		}
		opts := shortcut.RemoveOptions{}
		if removeBinary != nil {
			opts.Binary = removeBinary.Checked
		}
		if err := shortcut.Uninstall(sc, opts); err != nil {
			dialog.ShowError(err, ui.win)
		}
		ui.store.Remove(sc.Slug)
		if err := ui.store.Save(); err != nil {
			dialog.ShowError(err, ui.win)
		}
		ui.reload()
	}, ui.win)
	d.Show()
}

// install menjalankan pekerjaan berat (salin file besar / fetch favicon) di luar UI thread.
func (ui *app) install(work func() (store.Shortcut, error), done func(store.Shortcut, error)) {
	progress := dialog.NewProgressInfinite(T("install"), T("working"), ui.win)
	progress.Show()
	go func() {
		sc, err := work()
		fyne.Do(func() {
			progress.Hide()
			done(sc, err)
		})
	}()
}

// added menyimpan metadata shortcut baru; dipanggil dari callback install.
func (ui *app) added(sc store.Shortcut, err error, closeDialog func()) {
	if err != nil {
		dialog.ShowError(err, ui.win)
		return
	}
	ui.store.Add(sc)
	if err := ui.store.Save(); err != nil {
		ui.store.Remove(sc.Slug)
		dialog.ShowError(err, ui.win)
		return
	}
	closeDialog()
	ui.reload()
	dialog.ShowInformation(T("success_title"), fmt.Sprintf(T("success_msg"), sc.Name), ui.win)
}

// sectionTitle adalah judul bagian di dalam dialog.
func sectionTitle(text string) fyne.CanvasObject {
	return container.NewVBox(sectionLabel(text), widget.NewSeparator())
}
