package shortcut

import (
	"debug/elf"
	"strings"
)

// guiLibs adalah library yang menandakan aplikasi berjendela sendiri.
// Kalau binary tidak memakai satu pun dari ini, biasanya ia program terminal.
var guiLibs = []string{
	"libX11", "libxcb", "libwayland-client", "libgtk", "libgdk",
	"libQt5", "libQt6", "libSDL", "libGL", "libEGL", "libglfw", "libglib",
}

// RunsInTerminal menebak apakah program perlu jendela terminal.
//
// Script dan aplikasi baris perintah butuh Terminal=true di .desktop,
// sedangkan aplikasi GUI tidak (kalau dipaksa, jendela terminal ikut terbuka).
// Tebakannya dari daftar library dinamis: tanpa library GUI dianggap CLI.
func RunsInTerminal(path string) bool {
	kind := DetectKind(path)
	if kind == KindScript {
		return true
	}
	if kind == KindWindows {
		return false // .exe Windows (lewat wine) umumnya GUI
	}

	f, err := elf.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	libs, err := f.DynString(elf.DT_NEEDED)
	if err != nil {
		// binary statis (Go murni) tanpa dependency dinamis: kemungkinan CLI
		return true
	}
	for _, lib := range libs {
		for _, gui := range guiLibs {
			if strings.Contains(lib, gui) {
				return false
			}
		}
	}
	return true
}
