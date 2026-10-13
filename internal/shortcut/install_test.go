package shortcut

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lazyrun/internal/store"
)

var pngMagic = []byte("\x89PNG\r\n\x1a\n")

// testEnv mengarahkan semua path XDG ke tempdir dan mematikan tool eksternal,
// supaya tes tidak menyentuh ~/.local atau ~/Desktop milik user.
func testEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("PATH", "")
	return home
}

func fakeBinary(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, append([]byte{0x7f, 'E', 'L', 'F'}, []byte("fake")...), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestInstallExecutable(t *testing.T) {
	home := testEnv(t)
	src := fakeBinary(t, home, "winbox")

	sc, err := InstallExecutable(ExecOptions{
		BinaryPath: src,
		Name:       "Winbox",
		Category:   "Network",
		Menu:       true,
		Desktop:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if sc.BinaryPath != filepath.Join(home, ".local", "bin", "winbox") {
		t.Errorf("binary tersalin ke %s", sc.BinaryPath)
	}
	if st, err := os.Stat(sc.BinaryPath); err != nil || st.Mode().Perm()&0o111 == 0 {
		t.Errorf("binary harus executable: %v %v", st, err)
	}

	content, err := os.ReadFile(sc.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Name=Winbox\n",
		"Exec=" + sc.BinaryPath + "\n",
		"Categories=Network;\n",
		"Icon=" + sc.IconPath + "\n",
	} {
		if !strings.Contains(string(content), want) {
			t.Errorf("isi .desktop tidak memuat %q:\n%s", want, content)
		}
	}
	if strings.Contains(string(content), "NoDisplay") {
		t.Errorf("menu tercentang tidak boleh NoDisplay:\n%s", content)
	}

	if sc.DesktopCopyPath != filepath.Join(home, "Desktop", "winbox.desktop") {
		t.Errorf("salinan desktop di %s", sc.DesktopCopyPath)
	}
	if st, err := os.Stat(sc.DesktopCopyPath); err != nil || st.Mode().Perm()&0o111 == 0 {
		t.Errorf("file di Desktop harus executable: %v %v", st, err)
	}

	if _, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Winbox", Menu: true, Desktop: true}); err == nil {
		t.Error("install kedua dengan nama sama harus ditolak")
	}
}

func TestInstallExecutableStoresExecArgs(t *testing.T) {
	home := testEnv(t)
	src := fakeBinary(t, home, "args")

	sc, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Args", Menu: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.ExecArgs) != 1 || sc.ExecArgs[0] != sc.BinaryPath {
		t.Errorf("exec_args untuk tombol Open salah: %v", sc.ExecArgs)
	}
}

func TestInstallExecutableDesktopOnly(t *testing.T) {
	home := testEnv(t)
	src := fakeBinary(t, home, "tool")

	sc, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Tool", Menu: false, Desktop: true})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(sc.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "NoDisplay=true") {
		t.Errorf("desktop-only harus disembunyikan dari menu:\n%s", content)
	}
}

func TestInstallExecutableRejectsNonBinary(t *testing.T) {
	home := testEnv(t)
	path := filepath.Join(home, "notes.txt")
	if err := os.WriteFile(path, []byte("halo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallExecutable(ExecOptions{BinaryPath: path, Name: "Notes", Menu: true}); err == nil {
		t.Error("file teks biasa harus ditolak")
	}
}

func TestInstallExecutableRequiresTarget(t *testing.T) {
	home := testEnv(t)
	src := fakeBinary(t, home, "tool")
	if _, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Tool"}); err == nil {
		t.Error("tanpa target harus ditolak")
	}
}

func TestInstallWebApp(t *testing.T) {
	testEnv(t)

	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><head><link rel="icon" type="image/png" sizes="128x128" href="/fav.png"></head></html>`))
		case "/fav.png":
			_, _ = w.Write(append(pngMagic, []byte("fake")...))
		default:
			http.NotFound(w, r)
		}
	}))
	defer page.Close()

	sc, err := InstallWebApp(WebOptions{
		URL:     page.URL,
		Name:    "WhatsApp Web",
		Browser: BrowserArg{Name: "Chromium", Exec: []string{"/usr/bin/chromium"}},
		Menu:    true,
		Desktop: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(sc.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Exec=/usr/bin/chromium --app="+page.URL+"\n") {
		t.Errorf("Exec PWA salah:\n%s", content)
	}
	if filepath.Ext(sc.IconPath) != ".png" {
		t.Errorf("favicon harus tersimpan sebagai .png, dapat %s", sc.IconPath)
	}
	if st, err := os.Stat(sc.IconPath); err != nil || st.Size() == 0 {
		t.Errorf("ikon favicon tidak tertulis: %v %v", st, err)
	}
	if sc.Type != store.TypeWebApp {
		t.Errorf("type = %s", sc.Type)
	}
	if sc.URL != page.URL {
		t.Errorf("url = %s", sc.URL)
	}
}

func TestInstallWebAppFirefoxUsesNewWindow(t *testing.T) {
	testEnv(t)
	sc, err := InstallWebApp(WebOptions{
		URL:     "https://example.com",
		Name:    "Example",
		Browser: BrowserArg{Name: "Firefox", Exec: []string{"/usr/bin/firefox"}, Firefox: true},
		Menu:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(sc.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Exec=/usr/bin/firefox --new-window https://example.com\n") {
		t.Errorf("Exec Firefox salah:\n%s", content)
	}
	if len(sc.ExecArgs) != 3 || sc.ExecArgs[2] != "https://example.com" {
		t.Errorf("exec_args PWA salah: %v", sc.ExecArgs)
	}
}

func TestInstallWebAppFlatpakBrowser(t *testing.T) {
	testEnv(t)
	sc, err := InstallWebApp(WebOptions{
		URL:     "https://example.com",
		Name:    "Example",
		Browser: BrowserArg{Name: "Chromium (Flatpak)", Exec: []string{"flatpak", "run", "org.chromium.Chromium"}},
		Menu:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(sc.DesktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Exec=flatpak run org.chromium.Chromium --app=https://example.com\n") {
		t.Errorf("Exec flatpak salah:\n%s", content)
	}
}

func TestUninstall(t *testing.T) {
	home := testEnv(t)
	src := fakeBinary(t, home, "gone")

	sc, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Gone", Menu: true, Desktop: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := Uninstall(sc, RemoveOptions{Binary: true}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{sc.DesktopPath, sc.DesktopCopyPath, sc.IconPath, sc.BinaryPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s masih ada setelah uninstall", path)
		}
	}

	// uninstall kedua harus aman (file sudah tidak ada)
	if err := Uninstall(sc, RemoveOptions{Binary: true}); err != nil {
		t.Errorf("uninstall ulang: %v", err)
	}
}

func TestUninstallKeepsBinaryByDefault(t *testing.T) {
	home := testEnv(t)
	src := fakeBinary(t, home, "keep")

	sc, err := InstallExecutable(ExecOptions{BinaryPath: src, Name: "Keep", Menu: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := Uninstall(sc, RemoveOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sc.BinaryPath); err != nil {
		t.Errorf("binary harus tetap ada: %v", err)
	}
}
