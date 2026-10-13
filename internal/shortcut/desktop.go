package shortcut

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Entry adalah isi file .desktop yang dihasilkan LazyRun.
type Entry struct {
	Name       string
	ExecArgs   []string // argv lengkap; tiap arg di-quote sesuai spec Exec
	Icon       string   // path absolut atau nama ikon
	Categories string   // sudah diakhiri ";" atau kosong
	Terminal   bool     // true = jalankan di terminal (untuk script interaktif)
	WorkingDir string   // folder kerja saat dijalankan (Path=), opsional
	// NoDisplay menyembunyikan entry dari menu aplikasi. Dipakai kalau user
	// hanya memilih Desktop: file tetap ada di applications/ sebagai sumber kebenaran.
	NoDisplay bool
}

// Render menghasilkan isi file .desktop.
func (e Entry) Render() string {
	var b strings.Builder
	b.WriteString("[Desktop Entry]\n")
	b.WriteString("Type=Application\n")
	fmt.Fprintf(&b, "Name=%s\n", escapeValue(e.Name))
	fmt.Fprintf(&b, "Exec=%s\n", execArgs(e.ExecArgs))
	if e.Icon != "" {
		fmt.Fprintf(&b, "Icon=%s\n", escapeValue(e.Icon))
	}
	fmt.Fprintf(&b, "Terminal=%t\n", e.Terminal)
	if e.WorkingDir != "" {
		fmt.Fprintf(&b, "Path=%s\n", escapeValue(e.WorkingDir))
	}
	if e.Categories != "" {
		fmt.Fprintf(&b, "Categories=%s\n", escapeValue(e.Categories))
	}
	if e.NoDisplay {
		b.WriteString("NoDisplay=true\n")
	}
	b.WriteString("Comment=Created by LazyRun\n")
	return b.String()
}

// WriteDesktop menulis entry .desktop ke path.
func WriteDesktop(path string, e Entry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(e.Render()), 0o644)
}

func execArgs(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = execQuote(a)
	}
	return strings.Join(quoted, " ")
}

// execQuote mengutip satu argumen sesuai aturan field Exec (spec Desktop Entry).
func execQuote(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\n\"'\\><~|&;$*?#()`") {
		return s
	}
	r := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "`", "\\`", "$", "\\$", "\n", "\\n")
	return "\"" + r.Replace(s) + "\""
}

// escapeValue meng-escape nilai string biasa (Name, Icon, Categories).
func escapeValue(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	return r.Replace(s)
}

// Slug mengembalikan nama berkas yang aman dari nama aplikasi (dipakai UI).
func Slug(name string) string { return slugify(name) }

// slugify membuat nama file yang aman dari nama aplikasi.
func slugify(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
			dash = false
		case b.Len() > 0 && !dash:
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
