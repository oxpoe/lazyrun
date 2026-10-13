package ui

import (
	"net/url"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lazyrun/internal/browser"
	"lazyrun/internal/icon"
	"lazyrun/internal/shortcut"
	"lazyrun/internal/store"
)

func (ui *app) showWebAppDialog() {
	if len(ui.browsers) == 0 {
		dialog.ShowInformation(T("err_title"), T("no_browser"), ui.win)
		return
	}

	var iconPath, iconURL string
	var iconData []byte
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

	var d *dialog.CustomDialog
	install := newButton(T("install"), theme.DownloadIcon(), T("tip_install"), widget.HighImportance, func() {
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
			IconURL:  iconURL,
			IconData: iconData,
			Menu:     menu.Checked,
			Desktop:  desktop.Checked,
		}
		ui.install(
			func() (store.Shortcut, error) { return shortcut.InstallWebApp(opts) },
			func(sc store.Shortcut, err error) { ui.added(sc, err, d.Hide) },
		)
	})
	form := widget.NewForm(
		widget.NewFormItem(T("field_url"), urlEntry),
		widget.NewFormItem(T("field_name"), nameEntry),
		widget.NewFormItem(T("field_browser"), picker),
		widget.NewFormItem(T("field_icon"), container.NewBorder(nil, nil, nil,
			container.NewHBox(onlineIcon, browseIcon), iconLabel)),
	)
	hint := textLine(T("web_hint"), 12, false, colorNameMuted)
	content := container.NewVScroll(container.NewVBox(
		closeRow(func() { d.Hide() }),
		form,
		widget.NewSeparator(),
		sectionTitle(T("section_location")),
		menu,
		desktop,
		hint,
		container.NewHBox(logoBox(newLogo(), 20), spacer(),
			newButton(T("cancel"), nil, T("tip_cancel"), widget.MediumImportance, func() { d.Hide() }),
			install),
	))

	d = dialog.NewCustomWithoutButtons(T("new_web_title"), content, ui.win)
	d.SetOnClosed(pushHintTarget(hint))
	showDialog(d, ui.win, 560, 520)
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
