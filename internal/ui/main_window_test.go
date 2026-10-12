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

func TestDescribe(t *testing.T) {
	sc := store.Shortcut{Type: store.TypeWebApp, Targets: store.Targets{Menu: true}, DesktopPath: "/nonexistent/x.desktop"}
	got := describe(sc)
	if got != "Web · Application menu · missing" && got != "Web · Menu aplikasi · hilang" {
		t.Errorf("describe() = %q", got)
	}
}
