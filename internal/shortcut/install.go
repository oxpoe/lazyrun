package shortcut

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"lazyrun/internal/xdg"
)

func validate(name string, menu, desktop bool) error {
	if name == "" {
		return errors.New("name is required")
	}
	if strings.ContainsAny(name, "\n\r") {
		return errors.New("name must not contain line breaks")
	}
	if !menu && !desktop {
		return errors.New("pick at least one location (menu or desktop)")
	}
	if slugify(name) == "" {
		return errors.New("name must contain at least one letter or digit")
	}
	return nil
}

// checkRunnable memastikan file adalah ELF (binary/AppImage) atau script dengan shebang.
func checkRunnable(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return err
	}
	if st.IsDir() {
		return errors.New("this is a directory, not a file")
	}

	head := make([]byte, 4)
	n, _ := io.ReadFull(f, head)
	head = head[:n]
	if bytes.HasPrefix(head, []byte{0x7f, 'E', 'L', 'F'}) ||
		bytes.HasPrefix(head, []byte("#!")) ||
		bytes.HasPrefix(head, []byte("MZ")) { // .exe Windows (dijalankan lewat wine)
		return nil
	}
	return errors.New("not an executable (ELF, AppImage, script, or Windows .exe)")
}

// copyFile menyalin src ke dst lewat file sementara supaya dst tidak pernah setengah tertulis.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".lazyrun-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}

func samePath(a, b string) bool {
	sa, err := os.Stat(a)
	if err != nil {
		return false
	}
	sb, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(sa, sb)
}

// desktopCopy menyalin .desktop ke ~/Desktop agar bisa diklik ganda.
func desktopCopy(desktopPath, slug string) (string, error) {
	dst := filepath.Join(xdg.DesktopDir(), slug+".desktop")
	if _, err := os.Stat(dst); err == nil && !samePath(dst, desktopPath) {
		return "", errors.New(dst + " already exists")
	}
	return dst, copyFile(desktopPath, dst, 0o755)
}

// installBinary menyalin berkas target ke binDst.
//
// Script dapat perlakuan khusus: yang disalin adalah script aslinya (dengan
// akhiran .real) plus wrapper kecil. Desktop environment menjalankan shortcut
// dengan cwd = home dan PATH sesi, sehingga script yang memakai path relatif
// atau tool di ~/.local/bin gagal; wrapper yang membereskan keduanya.
func installBinary(kind, src, binDst string) error {
	if kind == KindScript {
		if err := copyFile(src, binDst+scriptSuffix, 0o755); err != nil {
			return err
		}
		return writeScriptWrapper(binDst, binDst+scriptSuffix, src)
	}
	if samePath(src, binDst) {
		return nil
	}
	return copyFile(src, binDst, 0o755)
}

// writeFileMode menulis berkas dengan mode tertentu (dipakai saat restore).
func writeFileMode(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, mode)
}

// scriptSuffix adalah akhiran berkas script asli di samping wrapper.
const scriptSuffix = ".real"

// writeScriptWrapper menulis wrapper yang menyiapkan PATH dan folder kerja
// sebelum menjalankan script.
//
// Script asli di folder asalnya dipakai lebih dulu kalau masih ada (supaya
// `dirname $0` dan berkas di sebelahnya tetap mengarah ke sana); salinan di
// ~/.local/bin jadi cadangan kalau script aslinya sudah pindah/dihapus.
func writeScriptWrapper(path, real, originFile string) error {
	body := "#!/bin/sh\n" +
		"# dibuat oleh LazyRun: siapkan PATH & folder kerja lalu jalankan script\n" +
		`PATH="$HOME/.local/bin:$PATH"` + "\n" +
		"export PATH\n"

	originDir := filepath.Dir(originFile)
	if originDir != "" && originDir != "." && originDir != "/" {
		quoted := shellQuote(originDir)
		body += "if [ -d " + quoted + " ]; then cd " + quoted + " || exit 1; fi\n"
	}

	if originFile != "" && originFile != real {
		body += "if [ -f " + shellQuote(originFile) + " ]; then exec " + shellQuote(originFile) + ` "$@"; fi` + "\n"
	}
	body += "exec " + shellQuote(real) + ` "$@"` + "\n"
	return os.WriteFile(path, []byte(body), 0o755)
}

// shellQuote mengutip nilai untuk dipakai di dalam script shell.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
