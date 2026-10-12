// Package browser mendeteksi browser yang terinstall untuk dipakai shortcut PWA.
package browser

import (
	"os"
	"os/exec"
	"strings"
)

// Kind menentukan cara membuka URL dalam mode aplikasi.
type Kind int

const (
	// Chromium membuka PWA dengan --app=<url>.
	Chromium Kind = iota
	// Firefox tidak punya mode app; dipakai --new-window.
	Firefox
)

// Browser adalah browser siap dipanggil.
type Browser struct {
	Name string
	Exec []string // argv lengkap, sudah termasuk argumen flatpak
	Kind Kind
}

// AppArgs adalah argumen yang membuka url sebagai aplikasi.
func (b Browser) AppArgs(url string) []string {
	if b.Kind == Firefox {
		return []string{"--new-window", url}
	}
	return []string{"--app=" + url}
}

type candidate struct {
	name string
	kind Kind
	bin  []string // nama di $PATH (tanpa "/") atau path absolut
}

var candidates = []candidate{
	{"Google Chrome", Chromium, []string{"google-chrome", "google-chrome-stable", "/opt/google/chrome/chrome"}},
	{"Chromium", Chromium, []string{"chromium", "chromium-browser"}},
	{"Brave", Chromium, []string{"brave-browser", "brave", "/opt/brave.com/brave/brave"}},
	{"Microsoft Edge", Chromium, []string{"microsoft-edge", "microsoft-edge-stable"}},
	{"Vivaldi", Chromium, []string{"vivaldi", "vivaldi-stable", "/opt/vivaldi/vivaldi"}},
	{"Helium", Chromium, []string{"helium", "/opt/helium-browser-bin/helium"}},
	{"Firefox", Firefox, []string{"firefox", "firefox-esr", "/usr/lib/firefox/firefox"}},
}

var flatpakApps = []struct {
	id, name string
	kind     Kind
}{
	{"com.google.Chrome", "Google Chrome (Flatpak)", Chromium},
	{"org.chromium.Chromium", "Chromium (Flatpak)", Chromium},
	{"com.brave.Browser", "Brave (Flatpak)", Chromium},
	{"com.microsoft.Edge", "Microsoft Edge (Flatpak)", Chromium},
	{"org.mozilla.firefox", "Firefox (Flatpak)", Firefox},
}

// Detect mengembalikan browser yang ada di sistem, native dulu lalu flatpak.
func Detect() []Browser {
	var out []Browser
	seen := map[string]bool{}

	add := func(b Browser) {
		if !seen[b.Name] {
			seen[b.Name] = true
			out = append(out, b)
		}
	}

	for _, c := range candidates {
		for _, bin := range c.bin {
			path := bin
			if !strings.Contains(bin, "/") {
				p, err := exec.LookPath(bin)
				if err != nil {
					continue
				}
				path = p
			} else if st, err := os.Stat(bin); err != nil || st.IsDir() {
				continue
			}
			add(Browser{Name: c.name, Exec: []string{path}, Kind: c.kind})
			break
		}
	}

	installed := flatpakList()
	for _, f := range flatpakApps {
		if installed[f.id] {
			add(Browser{Name: f.name, Exec: []string{"flatpak", "run", f.id}, Kind: f.kind})
		}
	}
	return out
}

func flatpakList() map[string]bool {
	out := map[string]bool{}
	if _, err := exec.LookPath("flatpak"); err != nil {
		return out
	}
	data, err := exec.Command("flatpak", "list", "--app", "--columns=application").Output()
	if err != nil {
		return out
	}
	for _, line := range strings.Split(string(data), "\n") {
		if id := strings.TrimSpace(line); id != "" {
			out[id] = true
		}
	}
	return out
}
