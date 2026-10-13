package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Palet brand LazyRun, diambil dari logo (dream sky: ungu malam, senja pink, bulan amber).
var (
	brandPurple  = color.NRGBA{R: 0x6A, G: 0x3D, B: 0x9A, A: 0xFF}
	brandMagenta = color.NRGBA{R: 0xC2, G: 0x5C, B: 0x9E, A: 0xFF}
	brandAmber   = color.NRGBA{R: 0xFF, G: 0xD1, B: 0x66, A: 0xFF}
)

// Nama warna tambahan di luar tema standar Fyne.
const (
	colorNameCard       fyne.ThemeColorName = "lazyrunCard"
	colorNameCardHover  fyne.ThemeColorName = "lazyrunCardHover"
	colorNameMuted      fyne.ThemeColorName = "lazyrunMuted"
	colorNameChipMenu   fyne.ThemeColorName = "lazyrunChipMenu"
	colorNameChipDesk   fyne.ThemeColorName = "lazyrunChipDesktop"
	colorNameChipBroken fyne.ThemeColorName = "lazyrunChipBroken"
	colorNameChipKind   fyne.ThemeColorName = "lazyrunChipKind"
)

// brand menimpa sebagian warna/ukuran tema bawaan; terang/gelap tetap ikut sistem.
type brand struct{ fyne.Theme }

func (b brand) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	dark := variant == theme.VariantDark
	switch name {
	case theme.ColorNamePrimary:
		// senja pink di malam hari, ungu tua di tema terang
		return pick(dark, brandMagenta, color.NRGBA{R: 0x8E, G: 0x3F, B: 0xA8, A: 0xFF})
	case theme.ColorNameBackground:
		// langit malam surreal
		return pick(dark, color.NRGBA{R: 0x14, G: 0x0E, B: 0x24, A: 0xFF}, color.NRGBA{R: 0xF5, G: 0xEF, B: 0xF9, A: 0xFF})
	case colorNameCard:
		return pick(dark, color.NRGBA{R: 0x1E, G: 0x16, B: 0x36, A: 0xFF}, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	case colorNameCardHover:
		return withAlpha(brandAmber, pick[uint8](dark, 0x1F, 0x18))
	case theme.ColorNameInputBackground:
		return pick(dark, color.NRGBA{R: 0x19, G: 0x11, B: 0x35, A: 0xFF}, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	case theme.ColorNameButton, theme.ColorNameMenuBackground:
		return pick(dark, color.NRGBA{R: 0x26, G: 0x1B, B: 0x45, A: 0xFF}, color.NRGBA{R: 0xED, G: 0xE3, B: 0xF5, A: 0xFF})
	case theme.ColorNameOverlayBackground:
		return pick(dark, color.NRGBA{R: 0x1B, G: 0x13, B: 0x33, A: 0xFF}, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	case theme.ColorNameInputBorder, theme.ColorNameSeparator:
		return pick(dark, color.NRGBA{R: 0x3A, G: 0x2A, B: 0x63, A: 0xFF}, color.NRGBA{R: 0xDC, G: 0xCF, B: 0xE9, A: 0xFF})
	case colorNameMuted:
		return pick(dark, color.NRGBA{R: 0xA7, G: 0x9A, B: 0xC8, A: 0xFF}, color.NRGBA{R: 0x6E, G: 0x61, B: 0x83, A: 0xFF})
	case colorNameChipMenu:
		return withAlpha(brandPurple, pick[uint8](dark, 0x66, 0x2E))
	case colorNameChipDesk:
		return withAlpha(brandMagenta, pick[uint8](dark, 0x4E, 0x30))
	case colorNameChipKind:
		return withAlpha(color.NRGBA{R: 0xFF, G: 0xD1, B: 0x66, A: 0xFF}, pick[uint8](dark, 0x24, 0x30))
	case colorNameChipBroken:
		// merah senja: peringatan yang tetap terasa "seram" tapi terbaca
		return withAlpha(color.NRGBA{R: 0xD9, G: 0x46, B: 0x3C, A: 0xFF}, pick[uint8](dark, 0x50, 0x30))
	}
	return b.Theme.Color(name, variant)
}

// Size membuat sudut tombol/kartu lebih membulat dari tema bawaan.
func (b brand) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameButtonRadius, theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameCardRadius, theme.SizeNamePopupRadius:
		return 14
	case theme.SizeNamePadding:
		return 5
	}
	return b.Theme.Size(name)
}

func pick[T any](cond bool, yes, no T) T {
	if cond {
		return yes
	}
	return no
}

func withAlpha(c color.NRGBA, a uint8) color.Color {
	c.A = a
	return c
}
