package ui

import (
	"net/url"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/internal/browser"
	"lazyrun/internal/shortcut"
	"lazyrun/internal/store"
)

func (ui *app) showWebAppDialog() {
	if len(ui.browsers) == 0 {
		dialog.ShowInformation(T("err_title"), T("no_browser"), ui.win)
		return
	}

	var iconPath string
	var nameTouched, filling bool

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://web.whatsapp.com")
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("WhatsApp Web")
	iconLabel := muted(T("icon_auto"))
	picker := widget.NewSelect(browserNames(ui.browsers), nil)
	picker.SetSelected(ui.browsers[0].Name)
	menu := widget.NewCheck(T("check_menu"), nil)
	menu.SetChecked(true)
	desktop := widget.NewCheck(T("check_desktop"), nil)
	desktop.SetChecked(true)

	nameEntry.OnChanged = func(string) {
		if !filling {
			nameTouched = true
		}
	}
	urlEntry.OnChanged = func(value string) {
		if nameTouched {
			return
		}
		if host := hostOf(value); host != "" {
			filling = true
			nameEntry.SetText(host)
			filling = false
		}
	}

	browseIcon := widget.NewButtonWithIcon(T("browse"), theme.FolderOpenIcon(), func() {
		pickFile(ui.win, imageFilter, func(path string) {
			iconPath = path
			iconLabel.SetText(filepath.Base(path))
		})
	})

	var d *dialog.CustomDialog
	install := widget.NewButtonWithIcon(T("install"), theme.DownloadIcon(), func() {
		switch {
		case strings.TrimSpace(urlEntry.Text) == "":
			dialog.ShowInformation(T("err_title"), T("need_url"), ui.win)
			return
		case strings.TrimSpace(nameEntry.Text) == "":
			dialog.ShowInformation(T("err_title"), T("need_name"), ui.win)
			return
		case !menu.Checked && !desktop.Checked:
			dialog.ShowInformation(T("err_title"), T("need_target"), ui.win)
			return
		}
		b := ui.browsers[indexOf(ui.browsers, picker.Selected)]
		opts := shortcut.WebOptions{
			URL:      urlEntry.Text,
			Name:     nameEntry.Text,
			Browser:  shortcut.BrowserArg{Name: b.Name, Exec: b.Exec, Firefox: b.Kind == browser.Firefox},
			IconPath: iconPath,
			Menu:     menu.Checked,
			Desktop:  desktop.Checked,
		}
		ui.install(
			func() (store.Shortcut, error) { return shortcut.InstallWebApp(opts) },
			func(sc store.Shortcut, err error) { ui.added(sc, err, d.Hide) },
		)
	})
	install.Importance = widget.HighImportance

	form := widget.NewForm(
		widget.NewFormItem(T("field_url"), urlEntry),
		widget.NewFormItem(T("field_name"), nameEntry),
		widget.NewFormItem(T("field_browser"), picker),
		widget.NewFormItem(T("field_icon"), container.NewBorder(nil, nil, nil, browseIcon, iconLabel)),
	)
	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		sectionTitle(T("section_location")),
		menu,
		desktop,
		muted(T("web_hint")),
		container.NewHBox(logoBox(newLogo(), 20), spacer(),
			widget.NewButton(T("cancel"), func() { d.Hide() }),
			install),
	)

	d = dialog.NewCustom(T("new_web_title"), "", content, ui.win)
	d.Resize(fyne.NewSize(560, 470))
	d.Show()
}

func browserNames(list []browser.Browser) []string {
	names := make([]string, len(list))
	for i, b := range list {
		names[i] = b.Name
	}
	return names
}

func indexOf(list []browser.Browser, name string) int {
	for i, b := range list {
		if b.Name == name {
			return i
		}
	}
	return 0
}

// hostOf mengambil hostname dari URL yang mungkin ditulis tanpa skema.
func hostOf(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
