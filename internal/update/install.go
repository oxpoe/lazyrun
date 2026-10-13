package update

import (
	"errors"
	"io"
)

// maxDownload membatasi ukuran unduhan (AppImage LazyRun ~12 MB).
const maxDownload = 300 << 20

// copyWithLimit menyalin sampai batas tertentu supaya unduhan tidak menghabiskan disk.
func copyWithLimit(dst io.Writer, src io.Reader, limit int64) (int64, error) {
	n, err := io.Copy(dst, io.LimitReader(src, limit+1))
	if err != nil {
		return n, err
	}
	if n > limit {
		return n, errors.New("berkas rilis terlalu besar")
	}
	return n, nil
}
