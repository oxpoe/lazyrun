// Package cursor meneruskan cursor theme desktop (KDE/GNOME) ke Fyne/GLFW.
package cursor

import (
	"os"
	"path/filepath"
	"strings"
)

// Apply men-set XCURSOR_THEME dan XCURSOR_SIZE kalau sesi belum men-set-nya.
// Fyne (GLFW) membaca env ini waktu memuat cursor, jadi panggil sebelum app dibuat.
func Apply() {
	if os.Getenv("XCURSOR_SIZE") == "" {
		if _, size := desktopCursor(); size != "" {
			os.Setenv("XCURSOR_SIZE", size)
		}
	}
	if os.Getenv("XCURSOR_THEME") != "" {
		return
	}
	if theme, _ := desktopCursor(); theme != "" {
		os.Setenv("XCURSOR_THEME", theme)
	}
}

func home() string {
	h, _ := os.UserHomeDir()
	return h
}

func configHome() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v
	}
	return filepath.Join(home(), ".config")
}

// desktopCursor mengambil nama & ukuran cursor dari config KDE, GTK, atau ~/.icons/default.
func desktopCursor() (theme, size string) {
	// KDE Plasma: ~/.config/kcminputrc
	kde := ini(filepath.Join(configHome(), "kcminputrc"))
	if t := kde["mouse"]["cursortheme"]; t != "" {
		return t, kde["mouse"]["cursorsize"]
	}

	for _, gtk := range []string{"gtk-4.0", "gtk-3.0"} {
		cfg := ini(filepath.Join(configHome(), gtk, "settings.ini"))
		if t := cfg["settings"]["gtk-cursor-theme-name"]; t != "" {
			return t, cfg["settings"]["gtk-cursor-theme-size"]
		}
	}

	// fallback: theme default di ~/.icons/default/index.theme (format "Inherits=")
	fallback := filepath.Join(home(), ".icons", "default", "index.theme")
	if t := ini(fallback)["icon theme"]["inherits"]; t != "" {
		return t, ""
	}
	return "", ""
}

// ini membaca file config berformat INI sederhana; key disimpan huruf kecil.
func ini(path string) map[string]map[string]string {
	out := map[string]map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}

	section := ""
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if out[section] == nil {
			out[section] = map[string]string{}
		}
		out[section][strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	return out
}
