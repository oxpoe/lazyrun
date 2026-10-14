package main

import (
	"fyne.io/fyne/v2/app"

	"lazyrun/internal/cursor"
	"lazyrun/internal/ui"
)

// version diisi saat build: -ldflags "-X main.version=1.0.0".
var version = "dev"

func main() {
	cursor.Apply() // sebelum app dibuat: GLFW membaca env cursor saat init
	ui.Run(app.NewWithID("io.oxpoe.lazyrun.LazyRun"), version)
}
