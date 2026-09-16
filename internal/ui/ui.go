package ui

import (
	"fightogm/internal/input"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	Background   = rl.NewColor(12, 18, 27, 255)
	Panel        = rl.NewColor(23, 32, 44, 245)
	Line         = rl.NewColor(62, 75, 91, 255)
	Text         = rl.NewColor(235, 238, 238, 255)
	Muted        = rl.NewColor(145, 159, 178, 255)
	Amber        = rl.NewColor(239, 181, 80, 255)
	PlayerColors = [2]rl.Color{rl.NewColor(245, 105, 82, 255), rl.NewColor(78, 191, 225, 255)}
)

type UI struct{ Font, Bold rl.Font }

func New() *UI {
	codepoints := make([]rune, 0, 512)
	for r := rune(32); r < 127; r++ {
		codepoints = append(codepoints, r)
	}
	for r := rune(0x400); r <= 0x45f; r++ {
		codepoints = append(codepoints, r)
	}
	codepoints = append(codepoints, '—', '–', '→', '←', '↑', '↓', '×', '·', '…', '№', '✓')
	u := &UI{Font: rl.LoadFontFromMemory(".ttf", goregular.TTF, 40, codepoints), Bold: rl.LoadFontFromMemory(".ttf", gobold.TTF, 80, codepoints)}
	rl.SetTextureFilter(u.Font.Texture, rl.FilterBilinear)
	rl.SetTextureFilter(u.Bold.Texture, rl.FilterBilinear)
	return u
}
func (u *UI) Close() { rl.UnloadFont(u.Font); rl.UnloadFont(u.Bold) }
func (u *UI) Text(s string, x, y, size float32, c rl.Color) {
	rl.DrawTextEx(u.Font, s, rl.NewVector2(x, y), size, 0, c)
}
func (u *UI) Heading(s string, x, y, size float32, c rl.Color) {
	rl.DrawTextEx(u.Bold, s, rl.NewVector2(x, y), size, .5, c)
}
func (u *UI) Center(s string, y, size float32, c rl.Color) {
	w := rl.MeasureTextEx(u.Bold, s, size, .5).X
	u.Heading(s, (1600-w)/2, y, size, c)
}
func (u *UI) Right(s string, x, y, size float32, c rl.Color) {
	w := rl.MeasureTextEx(u.Font, s, size, 0).X
	u.Text(s, x-w, y, size, c)
}
func (u *UI) Fit(s string, x, y, width, size float32, c rl.Color) {
	w := rl.MeasureTextEx(u.Font, s, size, 0).X
	if w > width {
		size *= width / w
	}
	u.Text(s, x, y, size, c)
}
func Rect(x, y, w, h float32, c rl.Color) { rl.DrawRectangleRec(rl.NewRectangle(x, y, w, h), c) }
func Border(x, y, w, h float32, c rl.Color) {
	rl.DrawRectangleLinesEx(rl.NewRectangle(x, y, w, h), 1, c)
}
func (u *UI) Header(section string) {
	u.Heading("FIGHTOGM", 64, 34, 32, Text)
	u.Text("/   "+section, 298, 44, 20, Muted)
	Rect(64, 92, 1472, 1, Line)
	u.Right("ОГМ  /  2D-ФАЙТИНГ", 1536, 46, 18, Amber)
}
func (u *UI) Footer(s string) { Rect(64, 849, 1472, 1, Line); u.Text(s, 64, 863, 18, Muted) }

type Layout struct{ X, Y, W, H, Gap float32 }

var MainLayout = Layout{96, 390, 500, 54, 12}
var OverlayLayout = Layout{555, 335, 490, 58, 14}

func (l Layout) Rect(i int) rl.Rectangle {
	return rl.NewRectangle(l.X, l.Y+float32(i)*(l.H+l.Gap), l.W, l.H)
}
func (l Layout) Handle(frame input.Frame, selected *int, count int) int {
	if frame.Up {
		*selected = (*selected + count - 1) % count
	}
	if frame.Down {
		*selected = (*selected + 1) % count
	}
	for i := 0; i < count; i++ {
		if rl.CheckCollisionPointRec(rl.NewVector2(frame.MouseX, frame.MouseY), l.Rect(i)) && frame.MouseClick {
			*selected = i
			return i
		}
	}
	if frame.Accept {
		return *selected
	}
	return -1
}
func (u *UI) Buttons(labels []string, selected int, l Layout) {
	for i, label := range labels {
		r := l.Rect(i)
		c := Panel
		fg := Text
		if i == selected {
			c = Amber
			fg = Background
		}
		rl.DrawRectangleRec(r, c)
		rl.DrawRectangleLinesEx(r, 1, Line)
		u.Text(label, r.X+22, r.Y+16, 24, fg)
		if i == selected {
			u.Right("→", r.X+r.Width-22, r.Y+15, 25, fg)
		}
	}
}
func (u *UI) Overlay(title, subtitle string, labels []string, selected int) {
	Rect(0, 0, 1600, 900, rl.Fade(Background, .87))
	u.Center(title, 195, 64, Text)
	u.Center(subtitle, 276, 21, Muted)
	u.Buttons(labels, selected, OverlayLayout)
}
