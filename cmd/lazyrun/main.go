package main

import (
	"fyne.io/fyne/v2/app"

	"lazyrun/internal/ui"
)

// version diisi saat build: -ldflags "-X main.version=1.0.0".
var version = "dev"

func main() {
	ui.Run(app.NewWithID("io.github.lazyrun.LazyRun"), version)
}
