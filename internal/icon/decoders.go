package icon

// decoder gambar tambahan: webp & bmp tidak ada di stdlib,
// jpeg/gif perlu di-import supaya terdaftar di image.Decode.
import (
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	_ "image/gif"
	_ "image/jpeg"
)
