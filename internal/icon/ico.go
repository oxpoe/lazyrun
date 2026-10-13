package icon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/png"
)

// decodeICO mengambil gambar terbesar di dalam file .ico.
//
// ICO bisa berisi PNG (modern) atau DIB/BMP (format lama, masih sering dipakai
// favicon). Keduanya didukung: BMP 32-bit pakai alpha, 24-bit pakai AND mask.
// Entri dengan palet (<= 8-bit) tidak didukung.
func decodeICO(data []byte) (image.Image, error) {
	if pngData := largestPNGFromICO(data); pngData != nil {
		return png.Decode(bytes.NewReader(pngData))
	}

	if len(data) < 22 || data[0] != 0 || data[1] != 0 || data[2] != 1 || data[3] != 0 {
		return nil, errors.New("bukan file ICO")
	}
	count := int(data[4]) | int(data[5])<<8

	var best image.Image
	bestSize := 0
	for i := 0; i < count; i++ {
		off := 6 + i*16
		if off+16 > len(data) {
			break
		}
		width := int(data[off])
		if width == 0 {
			width = 256
		}
		size := int(data[off+8]) | int(data[off+9])<<8 | int(data[off+10])<<16 | int(data[off+11])<<24
		start := int(data[off+12]) | int(data[off+13])<<8 | int(data[off+14])<<16 | int(data[off+15])<<24
		if size <= 0 || start <= 0 || start+size > len(data) {
			continue
		}
		if width*width <= bestSize {
			continue
		}
		img, err := decodeDIB(data[start : start+size])
		if err != nil {
			continue
		}
		best, bestSize = img, width*width
	}
	if best == nil {
		return nil, errors.New("tidak ada gambar yang bisa dibaca di dalam .ico")
	}
	return best, nil
}

// decodeDIB membaca gambar DIB/BMP di dalam entri ICO.
func decodeDIB(dib []byte) (image.Image, error) {
	if len(dib) < 40 {
		return nil, errors.New("entri ICO terlalu pendek")
	}
	headerSize := int(binary.LittleEndian.Uint32(dib[0:4]))
	width := int(int32(binary.LittleEndian.Uint32(dib[4:8])))
	height := int(int32(binary.LittleEndian.Uint32(dib[8:12]))) / 2 // tinggi DIB = XOR + AND mask
	bitCount := int(binary.LittleEndian.Uint16(dib[14:16]))
	compression := int(binary.LittleEndian.Uint32(dib[16:20]))

	if compression != 0 {
		return nil, errors.New("entri ICO terkompresi tidak didukung")
	}
	if width <= 0 || height <= 0 || width > 512 || height > 512 {
		return nil, errors.New("ukuran entri ICO tidak wajar")
	}
	if bitCount != 24 && bitCount != 32 {
		return nil, errors.New("hanya ICO 24/32-bit yang didukung")
	}

	stride := (width*bitCount/8 + 3) & ^3
	start := headerSize
	if start+stride*height > len(dib) {
		return nil, errors.New("data piksel ICO terpotong")
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	allTransparent := bitCount == 32
	for y := 0; y < height; y++ {
		row := dib[start+(height-1-y)*stride:] // DIB disimpan dari bawah ke atas
		for x := 0; x < width; x++ {
			px := row[x*bitCount/8:]
			var alpha uint8 = 0xFF
			if bitCount == 32 {
				alpha = px[3]
				if alpha != 0 {
					allTransparent = false
				}
			}
			img.SetNRGBA(x, y, color.NRGBA{R: px[2], G: px[1], B: px[0], A: alpha})
		}
	}

	// sebagian ICO 32-bit tidak mengisi kanal alpha sama sekali
	if allTransparent {
		for i := 3; i < len(img.Pix); i += 4 {
			img.Pix[i] = 0xFF
		}
	}
	return img, nil
}
