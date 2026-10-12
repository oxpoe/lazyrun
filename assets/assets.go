// Package assets menyimpan aset yang di-embed ke dalam binary.
package assets

import _ "embed"

//go:embed icon.png
var Icon []byte
