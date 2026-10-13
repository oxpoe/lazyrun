package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Animasi di sini sengaja pendek dan tidak menyentuh posisi/ukuran objek,
// supaya tidak bertabrakan dengan layout container.

// pulseGradient membuat garis aksen bergerak bolak-balik antara indigo dan cyan.
func pulseGradient(bar *canvas.LinearGradient) {
	anim := canvas.NewColorRGBAAnimation(brandAmber, brandMagenta, 3*time.Second, func(c color.Color) {
		bar.EndColor = c
		canvas.Refresh(bar)
	})
	anim.AutoReverse = true
	anim.RepeatCount = fyne.AnimationRepeatForever
	anim.Start()
}

// hoverCard adalah kartu yang warnanya berubah halus saat disorot mouse,
// dan bisa dipakai untuk animasi kedatangan baris baru.
type hoverCard struct {
	widget.BaseWidget
	bg      *canvas.Rectangle
	content fyne.CanvasObject
	hover   bool
}

func newHoverCard(content fyne.CanvasObject) *hoverCard {
	c := &hoverCard{
		bg:      canvas.NewRectangle(theme.Color(colorNameCard)),
		content: content,
	}
	c.bg.CornerRadius = 12
	c.ExtendBaseWidget(c)
	return c
}

func (c *hoverCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(c.bg, container.NewPadded(c.content)))
}

func (c *hoverCard) MouseIn(*desktop.MouseEvent)    { c.setHover(true) }
func (c *hoverCard) MouseMoved(*desktop.MouseEvent) {}
func (c *hoverCard) MouseOut()                      { c.setHover(false) }

func (c *hoverCard) setHover(on bool) {
	if c.hover == on {
		return
	}
	c.hover = on
	c.animate(theme.Color(pick(on, colorNameCardHover, colorNameCard)), 150*time.Millisecond)
}

// flash membuat kartu berkedip warna brand sebentar (baris baru muncul).
func (c *hoverCard) flash(delay time.Duration) {
	start := func() { c.animate(theme.Color(colorNameCard), 450*time.Millisecond) }
	if delay <= 0 {
		c.bg.FillColor = theme.Color(colorNameChipMenu)
		start()
		return
	}
	time.AfterFunc(delay, func() {
		fyne.Do(func() {
			c.bg.FillColor = theme.Color(colorNameChipMenu)
			start()
		})
	})
}

// animate memindahkan warna latar kartu secara halus dari warna sekarang.
func (c *hoverCard) animate(goal color.Color, d time.Duration) {
	from := c.bg.FillColor
	canvas.NewColorRGBAAnimation(from, goal, d, func(col color.Color) {
		c.bg.FillColor = col
		canvas.Refresh(c.bg)
	}).Start()
}
