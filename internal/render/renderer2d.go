// Package render draws a side-view arena and frame-animated 2D fighters.
package render

import (
	"fightogm/internal/assets"
	"fightogm/internal/combat"
	"fightogm/internal/fighter"
	"fightogm/internal/game"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"path/filepath"
)

const GroundY float32 = 752

type SpriteController struct {
	Name           string
	Time, Duration float32
	Serial         uint64
	Frame          int
	FrameID        string
}

func (c *SpriteController) NormalizedTime() float32 {
	if c.Duration <= 0 {
		return 0
	}
	return min(1, c.Time/c.Duration)
}

type SpriteActor struct {
	Atlas      *assets.SpriteAtlas
	Controller SpriteController
}

func (a *SpriteActor) Update(f *fighter.Fighter, dt float32) {
	name := f.Animation
	if f.Move != nil && f.Move.ID == "low" {
		name = "kick_low"
	}
	clip, ok := a.Atlas.Clips[name]
	if !ok {
		name = "idle"
		clip = a.Atlas.Clips[name]
	}
	c := &a.Controller
	if c.Name != name || c.Serial != f.AnimationSerial {
		*c = SpriteController{Name: name, Serial: f.AnimationSerial, Duration: f.AnimationDuration}
	}
	if c.Duration <= 0 {
		c.Duration = float32(len(clip.Frames)) / clip.FPS
	}
	c.Time += dt
	startup, active := float32(-1), float32(0)
	if f.Move != nil {
		c.Time = f.MoveTime
		c.Duration = f.Move.Duration()
		startup = f.Move.Startup
		active = f.Move.Active
	}
	c.Frame = clip.FrameAt(c.Time, c.Duration, startup, active)
	if clip.Attack && f.Move != nil && f.Move.Hits > 1 && c.Time >= startup && c.Time < startup+active {
		// Recoil between individual contacts in flurries; hit-stop holds contact.
		phase := float32(math.Mod(float64((c.Time-startup)/active*float32(f.Move.Hits)), 1))
		c.Frame = 1
		if phase > .68 {
			c.Frame = 2
		}
	}
	c.FrameID = clip.Frames[c.Frame]
}

type spark struct {
	Position, Velocity combat.Vec3
	Life, MaxLife      float32
	Color              rl.Color
	Strength           float32
}

type Renderer struct {
	Atlases                          map[string]*assets.SpriteAtlas
	Roster                           []fighter.Definition
	Actors                           [2]*SpriteActor
	Stage                            rl.Texture2D
	Shader                           rl.Shader
	Font                             rl.Font
	Time, Unit, Center, Shake, Flash float32
	Sparks                           []spark
	Seed                             uint32
}

func New(root string, roster []fighter.Definition, spec assets.SpriteRoster, font rl.Font, progress func(int, string)) (*Renderer, error) {
	r := &Renderer{Atlases: make(map[string]*assets.SpriteAtlas), Roster: roster, Font: font, Unit: 210, Seed: 17}
	r.Shader = rl.LoadShaderFromMemory("", `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
uniform sampler2D texture0;
uniform vec4 colDiffuse;
out vec4 finalColor;
void main() {
 vec4 c=texture(texture0,fragTexCoord);
 float green=c.g-max(c.r,c.b);
 float matte=1.0-smoothstep(0.08,0.38,green);
 if(matte<0.015) discard;
 if(green>0.015) c.g=min(c.g,max(c.r,c.b)+0.015);
 finalColor=vec4(c.rgb,c.a*matte)*fragColor*colDiffuse;
}`)
	if r.Shader.ID == 0 {
		return nil, fmt.Errorf("не удалось создать шейдер 2D-спрайтов")
	}
	for i, d := range roster {
		progress(i, d.Name)
		atlas, err := assets.LoadSpriteAtlas(filepath.Join(root, "assets/sprites", d.ID, "animations.json"))
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("%s: %w", d.Name, err)
		}
		r.Atlases[d.ID] = atlas
		for _, move := range d.Moves {
			if _, ok := atlas.Clips[move.Animation]; !ok {
				r.Close()
				return nil, fmt.Errorf("%s: нет анимации приёма %s (%s)", d.Name, move.Name, move.Animation)
			}
		}
	}
	r.Stage = rl.LoadTexture(filepath.Join(root, spec.Stage))
	if r.Stage.ID == 0 {
		r.Close()
		return nil, fmt.Errorf("не удалось загрузить фон 2D-арены")
	}
	rl.SetTextureFilter(r.Stage, rl.FilterBilinear)
	return r, nil
}

func (r *Renderer) SetMatch(m *game.Match) error {
	for i, f := range m.Fighters {
		atlas, ok := r.Atlases[f.Definition.ID]
		if !ok {
			return fmt.Errorf("нет спрайтов %s", f.Definition.Name)
		}
		r.Actors[i] = &SpriteActor{Atlas: atlas}
		r.Actors[i].Update(f, 0)
	}
	r.Center = 0
	r.Unit = 210
	r.Shake = 0
	r.Flash = 0
	r.Sparks = nil
	return nil
}

func (r *Renderer) random() float32 {
	r.Seed = r.Seed*1664525 + 1013904223
	return float32(r.Seed>>8) / 16777216
}

func (r *Renderer) Events(events []game.Event) {
	for _, e := range events {
		if e.Kind == "round" || e.Kind == "start" || e.Kind == "victory" {
			continue
		}
		color := rl.NewColor(255, 205, 112, 255)
		if e.Kind == "block" {
			color = rl.NewColor(135, 219, 255, 255)
		}
		if e.Kind == "ultimate" || e.Kind == "omni" {
			color = rl.NewColor(255, 93, 62, 255)
			r.Flash = .10
		}
		if e.Kind == "hit" || e.Kind == "heavy" || e.Kind == "kick" || e.Kind == "ko" {
			r.Shake = max(r.Shake, .1+e.Strength*.18)
			r.Flash = max(r.Flash, e.Strength*.055)
		}
		count := 12 + int(e.Strength*12)
		for j := 0; j < count; j++ {
			angle := float64(r.random()) * math.Pi * 2
			speed := 1 + r.random()*4
			life := .15 + r.random()*.32
			r.Sparks = append(r.Sparks, spark{e.Position, combat.Vec3{X: float32(math.Cos(angle)) * speed, Y: float32(math.Sin(angle)) * speed}, life, life, color, e.Strength})
		}
	}
	if len(r.Sparks) > 384 {
		r.Sparks = r.Sparks[len(r.Sparks)-384:]
	}
}

func (r *Renderer) Update(m *game.Match, dt, animationDT float32) {
	r.Time += animationDT
	r.Shake = max(0, r.Shake-dt)
	r.Flash = max(0, r.Flash-dt)
	if m == nil {
		return
	}
	distance := float32(math.Abs(float64(m.Fighters[0].Position.X - m.Fighters[1].Position.X)))
	tallest := float32(2)
	for i, f := range m.Fighters {
		r.Actors[i].Update(f, animationDT)
		frame := r.Actors[i].Atlas.Frames[r.Actors[i].Controller.FrameID]
		tallest = max(tallest, f.Position.Y+frame.WorldHeight)
	}
	// Leave room below the HUD even at the apex of a jump or during a launcher.
	verticalLimit := (GroundY - 188) / tallest
	wanted := min(float32(210), 1240/(distance+2.2), verticalLimit)
	r.Unit += (wanted - r.Unit) * min(1, dt*8)
	r.Unit = min(r.Unit, verticalLimit)
	center := cameraCenter(r.Center, r.Unit, m.Fighters[0].Position.X, m.Fighters[1].Position.X)
	r.Center += (center - r.Center) * min(1, dt*9)
	active := r.Sparks[:0]
	for _, p := range r.Sparks {
		p.Life -= animationDT
		if p.Life <= 0 {
			continue
		}
		p.Position = p.Position.Add(p.Velocity.Mul(animationDT))
		p.Velocity.Y -= animationDT * 5
		active = append(active, p)
	}
	r.Sparks = active
}

// Pan only when a fighter reaches the safe edge. Following the midpoint at
// close range made a stationary opponent appear to slide toward the player.
func cameraCenter(current, unit, a, b float32) float32 {
	halfWidth := float32(620) / unit
	lo, hi := max(a, b)-halfWidth, min(a, b)+halfWidth
	if lo > hi {
		return (a + b) * .5
	}
	return combat.Clamp(current, lo, hi)
}

func (r *Renderer) Project(pos combat.Vec3) rl.Vector2 {
	x := 800 + (pos.X-r.Center)*r.Unit
	y := GroundY - pos.Y*r.Unit
	if r.Shake > 0 {
		x += float32(math.Sin(float64(r.Time)*113)) * r.Shake * 27
		y += float32(math.Cos(float64(r.Time)*83)) * r.Shake * 12
	}
	return rl.NewVector2(x, y)
}

func (r *Renderer) DrawStage() {
	rl.DrawTexturePro(r.Stage, rl.NewRectangle(0, 0, float32(r.Stage.Width), float32(r.Stage.Height)), rl.NewRectangle(0, 0, 1600, 900), rl.Vector2{}, 0, rl.White)
	rl.DrawRectangleGradientV(0, 0, 1600, 225, rl.NewColor(7, 10, 17, 165), rl.Blank)
	rl.DrawRectangleGradientV(0, 735, 1600, 165, rl.Blank, rl.NewColor(5, 9, 15, 220))
	for i := 0; i < 24; i++ {
		x := float32((i*173+127)%1600) + float32(math.Sin(float64(r.Time)*.25+float64(i)))*30
		y := 180 + float32(math.Mod(float64(i*97)+float64(r.Time)*float64(3+i%4), 530))
		rl.DrawCircleV(rl.NewVector2(x, y), float32(1+i%2), rl.NewColor(246, 205, 140, 35))
	}
}

// DrawFrame mirrors around a stable ground pivot, not the changing image centre.
func (r *Renderer) DrawFrame(atlas *assets.SpriteAtlas, id string, foot rl.Vector2, unit float32, flip bool, tint rl.Color) {
	f, ok := atlas.Frames[id]
	if !ok {
		return
	}
	scale := f.WorldHeight * unit / float32(f.Rect[3])
	pivot := f.Pivot[0]
	src := rl.NewRectangle(float32(f.Rect[0]), float32(f.Rect[1]), float32(f.Rect[2]), float32(f.Rect[3]))
	if flip {
		src.Width = -src.Width
		pivot = float32(f.Rect[2]) - pivot
	}
	dst := rl.NewRectangle(foot.X-pivot*scale, foot.Y-f.Pivot[1]*scale, float32(f.Rect[2])*scale, float32(f.Rect[3])*scale)
	rl.BeginShaderMode(r.Shader)
	if len(f.Regions) == 0 {
		rl.DrawTexturePro(atlas.Textures[f.Sheet], src, dst, rl.Vector2{}, 0, tint)
	} else {
		for _, region := range f.Regions {
			sx, sy, sw, sh := float32(region[0]), float32(region[1]), float32(region[2]), float32(region[3])
			source := rl.NewRectangle(float32(f.Rect[0])+sx, float32(f.Rect[1])+sy, sw, sh)
			dx := sx
			if flip {
				source.Width = -sw
				dx = float32(f.Rect[2]) - sx - sw
			}
			target := rl.NewRectangle(dst.X+dx*scale, dst.Y+sy*scale, sw*scale, sh*scale)
			rl.DrawTexturePro(atlas.Textures[f.Sheet], source, target, rl.Vector2{}, 0, tint)
		}
	}
	rl.EndShaderMode()
}

func (r *Renderer) DrawBattle(m *game.Match, debug bool) {
	r.DrawStage()
	if m.Fighters[0].Status.Omni > 0 || m.Fighters[1].Status.Omni > 0 {
		rl.DrawRectangle(0, 155, 1600, 610, rl.NewColor(56, 2, 8, 65))
	}
	for i, f := range m.Fighters {
		foot := r.Project(f.Position)
		ground := r.Project(combat.Vec3{X: f.Position.X})
		rl.DrawEllipse(int32(ground.X), int32(ground.Y+3), r.Unit*.51, r.Unit*.064, rl.NewColor(0, 0, 0, 140))
		marker := rl.NewColor(246, 128, 99, 155)
		if i == 1 {
			marker = rl.NewColor(93, 207, 244, 155)
		}
		rl.DrawEllipseLines(int32(ground.X), int32(ground.Y+5), r.Unit*.49, r.Unit*.064, marker)
		actor := r.Actors[i]
		flip := f.Facing.X < 0
		if f.Move != nil && f.Move.Effect == "engine" {
			r.DrawEngineTrail(f.Position, f.Facing.X)
		}
		if f.Status.Omni > 0 || (f.Move != nil && (f.Move.Effect == "afterimage" || f.Move.Effect == "trail")) {
			ghost := rl.NewColor(255, 97, 64, 255)
			if f.Status.Omni <= 0 {
				c := f.Definition.Color
				ghost = rl.NewColor(c[0], c[1], c[2], 255)
			}
			for j := 3; j >= 1; j-- {
				ghost.A = uint8(54 - j*10)
				r.DrawFrame(actor.Atlas, actor.Controller.FrameID, rl.NewVector2(foot.X-f.Facing.X*float32(j)*12, foot.Y), r.Unit, flip, ghost)
			}
		}
		tint := rl.White
		if i == 1 {
			tint = rl.NewColor(218, 238, 255, 255)
		}
		if f.Status.Invisible > 0 {
			tint.A = 55
		}
		if f.State == fighter.HitStun && f.StateTime < .07 {
			tint = rl.NewColor(255, 145, 125, 255)
		}
		r.DrawFrame(actor.Atlas, actor.Controller.FrameID, foot, r.Unit, flip, tint)
		if f.Status.Armor > 0 || f.Status.Gym > 0 {
			r.DrawStatusRing(f.Position, rl.NewColor(255, 207, 98, 255))
		}
		if f.Status.Counter > 0 || f.Status.AutoBlock > 0 {
			r.DrawStatusRing(f.Position, rl.NewColor(132, 213, 255, 255))
		}
		if f.Status.Capacitor > 0 {
			r.DrawStatusRing(f.Position, rl.NewColor(171, 187, 255, 255))
		}
		if f.Status.Resistance > 0 {
			pos := r.Project(f.Position.Add(combat.Vec3{Y: 1.85}))
			rl.DrawCircleLinesV(pos, 22+float32(math.Sin(float64(r.Time)*4))*4, rl.NewColor(255, 212, 110, 180))
		}
		if debug {
			for _, h := range f.Hurtboxes() {
				rl.DrawCircleLinesV(r.Project(h.Center), h.Radius*r.Unit, rl.Green)
			}
			for _, h := range f.Hitboxes() {
				rl.DrawCircleLinesV(r.Project(h.Center), h.Radius*r.Unit, rl.Red)
			}
		}
	}
	for _, p := range m.Projectiles {
		r.DrawProjectile(p)
	}
	for _, p := range r.Sparks {
		a := r.Project(p.Position)
		end := r.Project(p.Position.Sub(p.Velocity.Mul(.018)))
		alpha := p.Life / p.MaxLife
		rl.DrawLineEx(a, end, 2+p.Strength*2, rl.Fade(p.Color, alpha))
		rl.DrawCircleV(a, 2, rl.Fade(rl.White, alpha))
	}
	if r.Flash > 0 {
		rl.DrawRectangle(0, 160, 1600, 610, rl.Fade(rl.NewColor(255, 217, 156, 255), r.Flash*1.1))
	}
}

func (r *Renderer) DrawShowcase(selected [2]int, menu bool) {
	r.DrawStage()
	if menu {
		atlas := r.Atlases[r.Roster[0].ID]
		clip := atlas.Clips["idle"]
		r.DrawFrame(atlas, clip.Frames[int(r.Time*clip.FPS)%len(clip.Frames)], rl.NewVector2(1130, 825), 330, false, rl.White)
		return
	}
	rl.DrawRectangle(0, 130, 1600, 605, rl.NewColor(8, 13, 22, 100))
	for p, index := range selected {
		atlas := r.Atlases[r.Roster[index].ID]
		clip := atlas.Clips["idle"]
		tint := rl.White
		if p == 1 {
			tint = rl.NewColor(218, 238, 255, 255)
		}
		r.DrawFrame(atlas, clip.Frames[int(r.Time*clip.FPS)%len(clip.Frames)], rl.NewVector2(445+float32(p)*710, 704), 230, p == 1, tint)
	}
}

func (r *Renderer) UpdateShowcase(dt float32, selected [2]int) { r.Time += dt }

func (r *Renderer) DrawSelection(selected [2]int) {
	r.DrawStage()
	rl.DrawRectangle(0, 95, 1600, 465, rl.NewColor(8, 13, 22, 110))
	for p, index := range selected {
		atlas := r.Atlases[r.Roster[index].ID]
		clip := atlas.Clips["idle"]
		r.DrawFrame(atlas, clip.Frames[int(r.Time*clip.FPS)%len(clip.Frames)], rl.NewVector2(420+float32(p)*760, 538), 164, p == 1, rl.White)
	}
}

func (r *Renderer) DrawPortrait(index int, rect rl.Rectangle) {
	atlas := r.Atlases[r.Roster[index].ID]
	clip := atlas.Clips["idle"]
	rl.BeginScissorMode(int32(rect.X), int32(rect.Y), int32(rect.Width), int32(rect.Height))
	r.DrawFrame(atlas, clip.Frames[0], rl.NewVector2(rect.X+rect.Width*.45, rect.Y+190), 90, false, rl.White)
	rl.EndScissorMode()
}

func (r *Renderer) Close() {
	for _, atlas := range r.Atlases {
		atlas.Close()
	}
	if r.Stage.ID != 0 {
		rl.UnloadTexture(r.Stage)
	}
	if r.Shader.ID != 0 {
		rl.UnloadShader(r.Shader)
	}
}
