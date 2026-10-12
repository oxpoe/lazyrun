package shortcut

import (
	"errors"
	"os"

	"lazyrun/internal/store"
)

// RemoveOptions mengatur apa saja yang ikut dihapus saat uninstall.
type RemoveOptions struct {
	// Binary true = binary hasil salinan di ~/.local/bin juga dihapus.
	Binary bool
}

// Uninstall menghapus .desktop (menu + desktop), ikon, dan opsional binary.
func Uninstall(sc store.Shortcut, o RemoveOptions) error {
	var errs []error

	for _, path := range []string{sc.DesktopCopyPath, sc.DesktopPath} {
		if path == "" {
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}
	if sc.IconPath != "" {
		if err := os.Remove(sc.IconPath); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}
	if o.Binary && sc.Type == store.TypeExecutable && sc.BinaryPath != "" {
		if err := os.Remove(sc.BinaryPath); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}

	go RefreshCache()
	return errors.Join(errs...)
}
