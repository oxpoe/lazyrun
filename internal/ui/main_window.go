// Package ui berisi window utama dan dialog LazyRun.
package ui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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
	ui := &app{version: version}
	a.SetIcon(fyne.NewStaticResource("lazyrun.png", assets.Icon))

	ui.browsers = browser.Detect()
	ui.rows = container.NewVBox()
	ui.win = a.NewWindow(T("app_title"))
	ui.win.Resize(fyne.NewSize(760, 540))

	top := container.NewHBox(
		widget.NewButton(T("btn_executable"), ui.showExecutableDialog),
		widget.NewButton(T("btn_webapp"), ui.showWebAppDialog),
	)
	bottom := container.NewBorder(nil, nil,
		widget.NewButton(T("check_updates"), ui.checkUpdates),
		widget.NewLabel("v"+version))
	body := container.NewBorder(heading(T("installed")), nil, nil, nil, container.NewVScroll(ui.rows))
	ui.win.SetContent(container.NewBorder(top, bottom, nil, nil, body))

	st, err := store.Load(xdg.MetadataPath())
	ui.store = st
	ui.reload()
	return ui, err
}

func (ui *app) reload() {
	if st, err := store.Load(ui.store.Path); err == nil {
		ui.store = st
	}
	ui.rows.RemoveAll()
	if len(ui.store.Items) == 0 {
		ui.rows.Add(widget.NewLabel(T("empty")))
	}
	for _, sc := range ui.store.Items {
		ui.rows.Add(ui.row(sc))
	}
}

func (ui *app) row(sc store.Shortcut) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(sc.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	info := widget.NewLabel(describe(sc))

	open := widget.NewButton(T("folder"), func() {
		if err := exec.Command("xdg-open", filepath.Dir(sc.DesktopPath)).Start(); err != nil {
			dialog.ShowError(err, ui.win)
		}
	})
	remove := widget.NewButton(T("delete"), func() { ui.confirmDelete(sc) })

	return container.NewBorder(nil, widget.NewSeparator(), nil,
		container.NewHBox(open, remove),
		container.NewVBox(title, info))
}

func describe(sc store.Shortcut) string {
	targets := make([]string, 0, 2)
	if sc.Targets.Menu {
		targets = append(targets, T("check_menu"))
	}
	if sc.Targets.Desktop {
		targets = append(targets, T("check_desktop"))
	}
	loc := strings.Join(targets, " + ")

	kind := "Executable"
	if sc.Type == store.TypeWebApp {
		kind = "Web"
	}
	line := kind + " · " + loc
	if sc.Broken() {
		line += " · " + T("broken")
	}
	return line
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

func heading(text string) *widget.Label {
	return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

func spacer() fyne.CanvasObject { return layout.NewSpacer() }
