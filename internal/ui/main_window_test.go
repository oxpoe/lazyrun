package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"lazyrun/internal/store"
)

func TestBuildMainWindow(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui, err := build(a, "test")
	if err != nil {
		t.Fatalf("build(): %v", err)
	}
	if ui.win.Content() == nil {
		t.Fatal("window tanpa content")
	}

	// dialog harus bisa dibuka tanpa panic
	ui.showExecutableDialog()
	ui.showWebAppDialog()
	ui.checkUpdates()
}

func TestMetaText(t *testing.T) {
	sc := store.Shortcut{Type: store.TypeWebApp, URL: "https://example.com/x", Targets: store.Targets{Menu: true}}
	if got := metaText(sc); got != "Web · example.com" {
		t.Errorf("metaText() = %q", got)
	}
	exe := store.Shortcut{Type: store.TypeExecutable, BinaryPath: "/home/u/.local/bin/winbox"}
	if got := metaText(exe); got != "Executable · winbox" {
		t.Errorf("metaText() = %q", got)
	}
}
