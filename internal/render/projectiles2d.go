package render

import (
	"fightogm/internal/combat"
	"fightogm/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func (r *Renderer) DrawProjectile(p game.Projectile) {
	if p.Delay > 0 || p.Hit {
		return
	}
	center := r.Project(p.Position)
	forward := rl.NewVector2(p.Direction.X, -p.Direction.Y)
	if forward.X == 0 && forward.Y == 0 {
		forward.X = 1
	}
	point := func(x, y float32) rl.Vector2 {
		return rl.NewVector2(center.X+(forward.X*x-forward.Y*y)*r.Unit, center.Y+(forward.Y*x+forward.X*y)*r.Unit)
	}
	color := rl.NewColor(110, 222, 255, 255)
	if p.Kind == "wave" {
		color = rl.NewColor(205, 145, 255, 255)
	}
	if p.Kind == "auger" {
		color = rl.NewColor(242, 197, 128, 255)
	}
	if p.Kind == "foam" {
		color = rl.NewColor(255, 187, 65, 255)
	}
	if p.Kind == "arc" {
		color = rl.NewColor(156, 178, 255, 255)
	}
	for i := 7; i > 0; i-- {
		rl.DrawCircleV(point(-float32(i)*.06, 0), r.Unit*(.035+float32(7-i)*.012), rl.Fade(color, .028*float32(8-i)))
	}
	switch p.Kind {
	case "auger":
		rl.DrawLineEx(point(-.27, 0), point(.3, 0), r.Unit*.042, rl.NewColor(216, 224, 233, 255))
		for i := 0; i < 5; i++ {
			x := float32(i)*.09 - .24
			shift := float32(math.Sin(float64(r.Time)*28+float64(i))) * .045
			rl.DrawLineEx(point(x, -.08+shift), point(x+.07, .08+shift), r.Unit*.035, rl.NewColor(133, 153, 177, 255))
			rl.DrawLineEx(point(x, -.08+shift), point(x+.04, -.04+shift), r.Unit*.012, rl.White)
		}
		rl.DrawTriangle(point(.34, 0), point(.22, -.065), point(.22, .065), rl.White)
	case "bot":
		s := r.Unit * .17
		rl.DrawCircleV(center, s*1.5, rl.Fade(color, .12))
		rect := rl.NewRectangle(center.X-s, center.Y-s*.72, s*2, s*1.44)
		rl.DrawRectangleRounded(rect, .4, 6, rl.NewColor(12, 34, 57, 255))
		rl.DrawRectangleRoundedLinesEx(rect, .4, 6, 2, color)
		for _, x := range []float32{-.42, .42} {
			rl.DrawCircleV(rl.NewVector2(center.X+x*s, center.Y-s*.06), s*.16, rl.White)
		}
		rl.DrawLineEx(rl.NewVector2(center.X, center.Y-s*.7), rl.NewVector2(center.X, center.Y-s*1.1), 2, color)
		rl.DrawCircleV(rl.NewVector2(center.X, center.Y-s*1.2), 3, rl.White)
	case "water":
		for i := 0; i < 5; i++ {
			radius := r.Unit * (.075 + float32(i)*.025)
			at := point(-float32(i)*.035, float32(math.Sin(float64(r.Time)*14+float64(i)))*.08)
			rl.DrawCircleV(at, radius, rl.Fade(color, .10))
			rl.DrawCircleLinesV(at, radius, rl.Fade(color, .5))
		}
		rl.DrawCircleV(center, r.Unit*.065, rl.NewColor(222, 247, 255, 220))
	case "foam":
		rl.DrawEllipse(int32(center.X), int32(center.Y), r.Unit*.25, r.Unit*.14, rl.Fade(color, .8))
		for i := 0; i < 7; i++ {
			x := float32(i)*.067 - .21
			y := -.08 + float32(math.Sin(float64(i)*2+float64(r.Time)*12))*.055
			rl.DrawCircleV(point(x, y), r.Unit*(.055+float32(i%3)*.011), rl.NewColor(255, 249, 222, 240))
		}
		rl.DrawCircleV(point(.24, .025), r.Unit*.045, rl.White)
	case "electric", "arc":
		if p.Kind == "arc" {
			rl.DrawCircleV(center, r.Unit*.20, rl.Fade(color, .14))
			rl.DrawCircleLinesV(center, r.Unit*.16, color)
			for _, x := range []float32{-.27, .27} {
				rl.DrawCircleV(point(x, 0), r.Unit*.037, color)
			}
		}
		for j := 0; j < 2; j++ {
			last := point(-.29, 0)
			for i := 1; i <= 7; i++ {
				y := float32(math.Sin(float64(i*3+j)+float64(r.Time)*25)) * .13
				next := point(-.29+float32(i)*.085, y)
				rl.DrawLineEx(last, next, 6, rl.Fade(color, .3))
				rl.DrawLineEx(last, next, 2, rl.White)
				last = next
			}
		}
	default:
		for i := 0; i < 3; i++ {
			at := point(-float32(i)*.09, 0)
			rl.DrawCircleLinesV(at, r.Unit*(.16+float32(i)*.035), rl.Fade(color, 1-float32(i)*.22))
		}
		rl.DrawCircleV(center, r.Unit*.095, rl.Fade(color, .55))
	}
}

func (r *Renderer) DrawStatusRing(pos combat.Vec3, color rl.Color) {
	c := r.Project(pos.Add(combat.Vec3{Y: 1.05}))
	rl.DrawEllipseLines(int32(c.X), int32(c.Y), r.Unit*.43, r.Unit*.62, rl.Fade(color, .65))
}

func (r *Renderer) DrawEngineTrail(pos combat.Vec3, facing float32) {
	for i := 0; i < 5; i++ {
		p := pos.Add(combat.Vec3{X: -facing * (.35 + float32(i)*.14), Y: .05 + float32(i%2)*.07})
		at := r.Project(p)
		rl.DrawCircleV(at, r.Unit*(.065+float32(i)*.016), rl.NewColor(184, 161, 135, uint8(90-i*13)))
		end := r.Project(p.Add(combat.Vec3{X: -facing * .35}))
		rl.DrawLineEx(at, end, 2, rl.NewColor(255, 193, 88, uint8(160-i*25)))
	}
}
