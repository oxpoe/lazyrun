package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Palet brand LazyRun, diambil dari ikon aplikasi.
var (
	brandIndigo = color.NRGBA{R: 0x4F, G: 0x46, B: 0xE5, A: 0xFF}
	brandCyan   = color.NRGBA{R: 0x06, G: 0xB6, B: 0xD4, A: 0xFF}
)

// Nama warna tambahan di luar tema standar Fyne.
const (
	colorNameCard       fyne.ThemeColorName = "lazyrunCard"
	colorNameCardHover  fyne.ThemeColorName = "lazyrunCardHover"
	colorNameMuted      fyne.ThemeColorName = "lazyrunMuted"
	colorNameChipMenu   fyne.ThemeColorName = "lazyrunChipMenu"
	colorNameChipDesk   fyne.ThemeColorName = "lazyrunChipDesktop"
	colorNameChipBroken fyne.ThemeColorName = "lazyrunChipBroken"
)

// brand menimpa sebagian warna/ukuran tema bawaan; terang/gelap tetap ikut sistem.
type brand struct{ fyne.Theme }

func (b brand) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	dark := variant == theme.VariantDark
	switch name {
	case theme.ColorNamePrimary:
		return pick(dark, color.NRGBA{R: 0x7C, G: 0x7A, B: 0xFF, A: 0xFF}, brandIndigo)
	case theme.ColorNameBackground:
		return pick(dark, color.NRGBA{R: 0x13, G: 0x13, B: 0x1B, A: 0xFF}, color.NRGBA{R: 0xF5, G: 0xF5, B: 0xFA, A: 0xFF})
	case colorNameCard:
		return pick(dark, color.NRGBA{R: 0x1D, G: 0x1D, B: 0x28, A: 0xFF}, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	case colorNameCardHover:
		return withAlpha(brandIndigo, pick[uint8](dark, 0x2E, 0x14))
	case theme.ColorNameInputBackground:
		return pick(dark, color.NRGBA{R: 0x1A, G: 0x1A, B: 0x24, A: 0xFF}, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	case theme.ColorNameButton, theme.ColorNameMenuBackground:
		return pick(dark, color.NRGBA{R: 0x22, G: 0x22, B: 0x2E, A: 0xFF}, color.NRGBA{R: 0xEC, G: 0xEC, B: 0xF4, A: 0xFF})
	case theme.ColorNameOverlayBackground:
		return pick(dark, color.NRGBA{R: 0x1B, G: 0x1B, B: 0x25, A: 0xFF}, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	case theme.ColorNameInputBorder, theme.ColorNameSeparator:
		return pick(dark, color.NRGBA{R: 0x32, G: 0x32, B: 0x44, A: 0xFF}, color.NRGBA{R: 0xDC, G: 0xDC, B: 0xE8, A: 0xFF})
	case colorNameMuted:
		return pick(dark, color.NRGBA{R: 0x9C, G: 0x9C, B: 0xB4, A: 0xFF}, color.NRGBA{R: 0x6B, G: 0x6B, B: 0x80, A: 0xFF})
	case colorNameChipMenu:
		return withAlpha(brandIndigo, pick[uint8](dark, 0x44, 0x26))
	case colorNameChipDesk:
		return withAlpha(brandCyan, pick[uint8](dark, 0x3D, 0x2C))
	case colorNameChipBroken:
		return withAlpha(color.NRGBA{R: 0xE0, G: 0x4F, B: 0x4F, A: 0xFF}, pick[uint8](dark, 0x4A, 0x2E))
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
