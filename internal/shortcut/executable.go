package shortcut

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lazyrun/internal/store"
	"lazyrun/internal/xdg"
)

// ExecOptions adalah input Fungsi A: shortcut dari file executable.
type ExecOptions struct {
	BinaryPath string
	IconPath   string // opsional: file ikon
	IconURL    string // opsional: URL ikon
	IconData   []byte // opsional: ikon yang sudah diunduh UI
	Name       string
	Terminal   bool // jalankan di terminal (script interaktif)
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

	kind := DetectKind(o.BinaryPath)
	if err := installBinary(kind, o.BinaryPath, binDst); err != nil {
		return sc, err
	}
	iconPath, iconSource, err := installIconFor(iconOptions{
		IconPath:   o.IconPath,
		IconURL:    o.IconURL,
		IconData:   o.IconData,
		NearbyDir:  filepath.Dir(o.BinaryPath),
		NearbyBase: strings.TrimSuffix(filepath.Base(o.BinaryPath), filepath.Ext(o.BinaryPath)),
		Query:      name,
		Slug:       sl,
	})
	if err != nil {
		return sc, err
	}

	// .exe Windows dijalankan lewat wine; script lewat wrapper (PATH & cwd)
	execArgs := []string{binDst}
	if kind == KindWindows {
		wine, err := exec.LookPath("wine")
		if err != nil {
			return sc, errors.New("this is a Windows .exe; install wine first so the shortcut can run it")
		}
		execArgs = []string{wine, binDst}
	}

	entry := Entry{
		Name:       name,
		ExecArgs:   execArgs,
		Icon:       iconPath,
		Categories: category(o.Category, "Utility;"),
		Terminal:   o.Terminal && kind != KindWindows,
		NoDisplay:  !o.Menu,
	}
	if kind == KindScript {
		// desktop environment menjalankan dengan cwd = home, jadi folder asli
		// script dipakai sebagai Path= kalau masih ada
		if _, err := os.Stat(filepath.Dir(o.BinaryPath)); err == nil {
			entry.WorkingDir = filepath.Dir(o.BinaryPath)
		}
	}
	if err := WriteDesktop(deskDst, entry); err != nil {
		return sc, err
	}

	sc = store.Shortcut{
		Slug:        sl,
		Name:        name,
		Type:        store.TypeExecutable,
		Kind:        kind,
		ExecArgs:    entry.ExecArgs,
		Categories:  entry.Categories,
		Terminal:    entry.Terminal,
		BinaryPath:  binDst,
		IconPath:    iconPath,
		IconSource:  iconSource,
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
	IconPath string // opsional: file ikon
	IconURL  string // opsional: URL ikon
	IconData []byte // opsional: ikon yang sudah diunduh UI
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

	iconPath, iconSource, err := installIconFor(iconOptions{
		IconPath: o.IconPath,
		IconURL:  o.IconURL,
		IconData: o.IconData,
		PageURL:  pageURL,
		Query:    name,
		Slug:     sl,
	})
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
		Kind:        KindPWA,
		URL:         pageURL,
		Categories:  entry.Categories,
		Browser:     strings.Join(o.Browser.Exec, " "),
		ExecArgs:    entry.ExecArgs,
		IconPath:    iconPath,
		IconSource:  iconSource,
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
