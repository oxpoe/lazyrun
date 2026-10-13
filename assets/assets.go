// Package assets menyimpan aset yang di-embed ke dalam binary.
package assets

import _ "embed"

// Icon adalah ikon aplikasi LazyRun.
//
//go:embed icon.png
var Icon []byte

// Heart adalah ikon kecil untuk tombol donasi.
//
//go:embed heart.png
var Heart []byte
