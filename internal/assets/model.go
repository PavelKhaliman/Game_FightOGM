package assets

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"math"
)

type AnimationSet struct {
	ByName    map[string]int
	Durations map[string]float32
}
type Model struct {
	Model      rl.Model
	Animations []rl.ModelAnimation
	Set        AnimationSet
	Scale      float32
	Available  bool
	Error      string
	Controller AnimationController
	textures   []rl.Texture2D
}

func Load(path string, entry ManifestEntry, shader rl.Shader) *Model {
	out := &Model{Set: AnimationSet{ByName: make(map[string]int), Durations: make(map[string]float32)}}
	meta, err := ReadGLB(path)
	if err != nil {
		out.Error = err.Error()
		log.Printf("[ОШИБКА] %s", err)
		return out
	}
	out.Model = rl.LoadModel(path)
	out.prepareTextures()
	if out.Model.MeshCount == 0 || out.Model.BoneCount == 0 {
		out.Error = "Не удалось загрузить модель / скелет"
		log.Printf("[ОШИБКА] %s: %s", path, out.Error)
		if out.Model.MeshCount > 0 {
			rl.UnloadModel(out.Model)
		}
		out.closeTextures()
		return out
	}
	out.Animations = rl.LoadModelAnimations(path)
	for i, a := range out.Animations {
		name := a.GetName()
		if name == "" && i < len(meta.Clips) {
			name = meta.Clips[i].Name
			log.Printf("[ПРЕДУПРЕЖДЕНИЕ] %s: имя анимации #%d восстановлено из GLB", path, i)
		}
		if !rl.IsModelAnimationValid(out.Model, a) || a.FrameCount <= 0 {
			log.Printf("[ОШИБКА] %s: анимация %s несовместима", path, name)
			continue
		}
		out.Set.ByName[name] = i
		for _, c := range meta.Clips {
			if c.Name == name {
				out.Set.Durations[name] = c.Duration
			}
		}
	}
	requiredOK := true
	for _, name := range entry.Base {
		if _, ok := out.Set.ByName[name]; !ok {
			log.Printf("[ОШИБКА] %s: отсутствует основная анимация %q", path, name)
			requiredOK = false
		}
	}
	for _, name := range entry.Special {
		if _, ok := out.Set.ByName[name]; !ok {
			log.Printf("[ПРЕДУПРЕЖДЕНИЕ] %s: отсутствует специальная анимация %q", path, name)
		}
	}
	if len(out.Set.ByName) == 0 || !requiredOK {
		out.Error = "Не хватает боевых анимаций"
		rl.UnloadModelAnimations(out.Animations)
		rl.UnloadModel(out.Model)
		out.closeTextures()
		out.Animations = nil
		return out
	}
	bounds := rl.GetModelBoundingBox(out.Model)
	height := bounds.Max.Y - bounds.Min.Y
	if meta.ZUp {
		// Explicit basis: source (x, y, z) -> world (x, z, -y).
		// MatrixRotateX in raylib-go 0.55.1 uses the opposite sign to
		// raymath's general axis rotation; keep the asset convention explicit.
		out.Model.Transform = rl.Matrix{M0: 1, M6: -1, M9: 1, M15: 1}
		height = bounds.Max.Z - bounds.Min.Z
	}
	out.Scale = 2 / max(.1, height)
	for i := range out.Model.GetMaterials() {
		out.Model.GetMaterials()[i].Shader = shader
	}
	out.Available = true
	out.Controller.Play(out, "idle", true, 0, 1)
	log.Printf("[FightOGM] %s: модель OK, скелет %d костей, анимации OK (%d)", path, out.Model.BoneCount, len(out.Set.ByName))
	return out
}
func (m *Model) Close() {
	if m.Available {
		rl.UnloadModelAnimations(m.Animations)
		rl.UnloadModel(m.Model)
		m.closeTextures()
		m.Available = false
	}
}

// raylib leaves model textures owned by the caller. Track each uploaded image
// once, even when several material maps reference it, and filter fine fabric
// patterns with mipmaps so they do not flicker at gameplay distance.
func (m *Model) prepareTextures() {
	seen := make(map[uint32]rl.Texture2D)
	for _, material := range m.Model.GetMaterials() {
		for i := int32(0); i < rl.MaxMaterialMaps; i++ {
			texture := &material.GetMap(i).Texture
			if texture.ID == 0 || texture.ID == rl.GetTextureIdDefault() {
				continue
			}
			if existing, ok := seen[texture.ID]; ok {
				*texture = existing
				continue
			}
			rl.GenTextureMipmaps(texture)
			rl.SetTextureFilter(*texture, rl.FilterTrilinear)
			seen[texture.ID] = *texture
			m.textures = append(m.textures, *texture)
		}
	}
}

func (m *Model) closeTextures() {
	for _, texture := range m.textures {
		rl.UnloadTexture(texture)
	}
	m.textures = nil
}
func (m *Model) ValidateNames() error {
	if !m.Available {
		return fmt.Errorf("%s", m.Error)
	}
	return nil
}

type AnimationController struct {
	Name           string
	Time, Duration float32
	Looping        bool
	Serial         uint64
	Frame          int32
}

func (c *AnimationController) Play(m *Model, name string, loop bool, duration float32, serial uint64) {
	if c.Name == name && c.Serial == serial {
		return
	}
	if _, ok := m.Set.ByName[name]; !ok {
		log.Printf("[ПРЕДУПРЕЖДЕНИЕ] анимация %q недоступна, используется special_pose", name)
		name = "special_pose"
		if _, ok = m.Set.ByName[name]; !ok {
			name = "idle"
		}
	}
	c.Name = name
	c.Looping = loop
	c.Time = 0
	c.Serial = serial
	c.Duration = duration
	if duration <= 0 {
		c.Duration = m.Set.Durations[name]
	}
	if c.Duration <= 0 {
		c.Duration = 1
	}
	c.Apply(m)
}
func (c *AnimationController) Update(m *Model, dt float32) {
	if !m.Available {
		return
	}
	c.Time += dt
	if c.Looping {
		c.Time = float32(math.Mod(float64(c.Time), float64(c.Duration)))
	}
	c.Apply(m)
}
func (c *AnimationController) Apply(m *Model) {
	idx, ok := m.Set.ByName[c.Name]
	if !ok {
		return
	}
	a := m.Animations[idx]
	c.Frame = int32(c.NormalizedTime() * float32(a.FrameCount-1))
	rl.UpdateModelAnimation(m.Model, a, c.Frame)
}
func (c *AnimationController) IsPlaying(name string) bool { return c.Name == name && !c.Finished() }
func (c *AnimationController) NormalizedTime() float32 {
	if c.Duration <= 0 {
		return 0
	}
	return min(1, c.Time/c.Duration)
}
func (c *AnimationController) Finished() bool { return !c.Looping && c.Time >= c.Duration }
