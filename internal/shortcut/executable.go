package shortcut

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lazyrun/internal/icon"
	"lazyrun/internal/store"
	"lazyrun/internal/xdg"
)

// ExecOptions adalah input Fungsi A: shortcut dari file executable.
type ExecOptions struct {
	BinaryPath string
	IconPath   string // opsional
	Name       string
	Category   string
	Menu       bool
	Desktop    bool
}

// InstallExecutable menyalin binary + ikon, menulis .desktop, lalu refresh cache.
func InstallExecutable(o ExecOptions) (store.Shortcut, error) {
	var sc store.Shortcut

	name := strings.TrimSpace(o.Name)
	if err := validate(name, o.Menu, o.Desktop); err != nil {
		return sc, err
	}
	if err := checkRunnable(o.BinaryPath); err != nil {
		return sc, fmt.Errorf("%s: %w", filepath.Base(o.BinaryPath), err)
	}

	sl := slugify(name)
	binDst := filepath.Join(xdg.BinDir(), sl)
	deskDst := filepath.Join(xdg.ApplicationsDir(), sl+".desktop")
	if _, err := os.Stat(deskDst); err == nil {
		return sc, fmt.Errorf("shortcut %q already exists, uninstall it first", name)
	}
	if st, err := os.Stat(binDst); err == nil && !st.IsDir() && !samePath(o.BinaryPath, binDst) {
		return sc, fmt.Errorf("%s already exists, rename the app or remove that file first", binDst)
	}

	if !samePath(o.BinaryPath, binDst) {
		if err := copyFile(o.BinaryPath, binDst, 0o755); err != nil {
			return sc, err
		}
	}
	iconPath, err := icon.Save(o.IconPath, filepath.Dir(o.BinaryPath), strings.TrimSuffix(filepath.Base(o.BinaryPath), filepath.Ext(o.BinaryPath)), sl)
	if err != nil {
		return sc, err
	}

	entry := Entry{
		Name:       name,
		ExecArgs:   []string{binDst},
		Icon:       iconPath,
		Categories: category(o.Category, "Utility;"),
		NoDisplay:  !o.Menu,
	}
	if err := WriteDesktop(deskDst, entry); err != nil {
		return sc, err
	}

	sc = store.Shortcut{
		Slug:        sl,
		Name:        name,
		Type:        store.TypeExecutable,
		ExecArgs:    entry.ExecArgs,
		BinaryPath:  binDst,
		IconPath:    iconPath,
		DesktopPath: deskDst,
		Targets:     store.Targets{Menu: o.Menu, Desktop: o.Desktop},
		InstalledAt: time.Now().UTC(),
	}
	if o.Desktop {
		cp, err := desktopCopy(deskDst, sl)
		if err != nil {
			return sc, err
		}
		sc.DesktopCopyPath = cp
	}
	go RefreshCache()
	return sc, nil
}

// WebOptions adalah input Fungsi B: shortcut dari URL (PWA).
type WebOptions struct {
	URL      string
	Name     string
	Browser  BrowserArg
	IconPath string // opsional; kosong = auto-fetch favicon
	Menu     bool
	Desktop  bool
}

// BrowserArg adalah browser target yang dipilih di UI.
type BrowserArg struct {
	Name string
	Exec []string
	// Firefox membuka dengan --new-window, sisanya (Chromium-based) dengan --app=.
	Firefox bool
}

// InstallWebApp menulis .desktop yang membuka URL sebagai aplikasi.
func InstallWebApp(o WebOptions) (store.Shortcut, error) {
	var sc store.Shortcut

	name := strings.TrimSpace(o.Name)
	if err := validate(name, o.Menu, o.Desktop); err != nil {
		return sc, err
	}
	if len(o.Browser.Exec) == 0 {
		return sc, errors.New("no browser selected")
	}
	pageURL, err := normalizeURL(o.URL)
	if err != nil {
		return sc, err
	}

	sl := slugify(name)
	deskDst := filepath.Join(xdg.ApplicationsDir(), sl+".desktop")
	if _, err := os.Stat(deskDst); err == nil {
		return sc, fmt.Errorf("shortcut %q already exists, uninstall it first", name)
	}

	iconPath, err := installWebIcon(o.IconPath, pageURL, sl)
	if err != nil {
		return sc, err
	}

	appArgs := o.Browser.Exec[1:]
	if o.Browser.Firefox {
		appArgs = append(appArgs, "--new-window", pageURL)
	} else {
		appArgs = append(appArgs, "--app="+pageURL)
	}
	entry := Entry{
		Name:       name,
		ExecArgs:   append([]string{o.Browser.Exec[0]}, appArgs...),
		Icon:       iconPath,
		Categories: "Network;WebBrowser;",
		NoDisplay:  !o.Menu,
	}
	if err := WriteDesktop(deskDst, entry); err != nil {
		return sc, err
	}

	sc = store.Shortcut{
		Slug:        sl,
		Name:        name,
		Type:        store.TypeWebApp,
		URL:         pageURL,
		Browser:     strings.Join(o.Browser.Exec, " "),
		ExecArgs:    entry.ExecArgs,
		IconPath:    iconPath,
		DesktopPath: deskDst,
		Targets:     store.Targets{Menu: o.Menu, Desktop: o.Desktop},
		InstalledAt: time.Now().UTC(),
	}
	if o.Desktop {
		cp, err := desktopCopy(deskDst, sl)
		if err != nil {
			return sc, err
		}
		sc.DesktopCopyPath = cp
	}
	go RefreshCache()
	return sc, nil
}

func installWebIcon(userIcon, pageURL, slug string) (string, error) {
	if userIcon != "" {
		return icon.Save(userIcon, "", "", slug)
	}
	if data, ext, err := icon.Fetch(pageURL); err == nil {
		return icon.SaveBytes(data, ext, slug)
	}
	return icon.Default(slug)
}

func category(c, fallback string) string {
	c = strings.TrimSpace(c)
	if c == "" {
		return fallback
	}
	return strings.TrimSuffix(c, ";") + ";"
}

// normalizeURL menambah skema https kalau user menulis domain saja.
func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("URL is required")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("only http/https URLs are supported")
	}
	if u.Host == "" {
		return "", errors.New("URL has no host")
	}
	return u.String(), nil
}
