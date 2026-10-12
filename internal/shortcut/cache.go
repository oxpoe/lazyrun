package shortcut

import (
	"context"
	"os/exec"
	"time"

	"lazyrun/internal/xdg"
)

// RefreshCache memberi tahu desktop environment supaya .desktop baru langsung terbaca.
// Best-effort: kalau tool tidak ada atau gagal, menu tetap refresh saat login berikutnya.
func RefreshCache() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if p, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.CommandContext(ctx, p, xdg.ApplicationsDir()).Run()
	}
	for _, tool := range []string{"kbuildsycoca6", "kbuildsycoca5"} {
		if p, err := exec.LookPath(tool); err == nil {
			_ = exec.CommandContext(ctx, p, "--noincremental").Run()
			return
		}
	}
}
