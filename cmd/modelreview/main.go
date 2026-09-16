// Modelreview renders the actual skinned GLB through the game's native shader.
// It is an offline art review tool, not part of the shipped game executable.
package main

import (
	"fightogm/internal/art3d"
	"fightogm/internal/assets"
	"fightogm/internal/ui"
	"flag"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	runtime.LockOSThread()
	path := flag.String("model", "assets/models/ivanov.glb", "GLB to review")
	out := flag.String("output", "test-output/ivanov", "PNG output folder")
	clip := flag.String("animation", "", "Animation name; empty uses bind pose")
	frame := flag.Int("frame", 0, "Animation frame")
	compare := flag.String("compare", "", "Optional original GLB for a before/after render")
	flag.Parse()
	if err := os.MkdirAll(*out, 0755); err != nil {
		panic(err)
	}
	rl.SetConfigFlags(rl.FlagWindowHidden | rl.FlagMsaa4xHint)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.InitWindow(1200, 1000, "FightOGM model review")
	defer rl.CloseWindow()
	shader := art3d.LoadLighting()
	defer rl.UnloadShader(shader)
	model := assets.Load(*path, assets.ManifestEntry{}, shader)
	if !model.Available {
		panic(model.Error)
	}
	defer model.Close()
	if *clip != "" {
		idx, ok := model.Set.ByName[*clip]
		if !ok {
			panic("Unknown animation: " + *clip)
		}
		anim := model.Animations[idx]
		rl.UpdateModelAnimation(model.Model, anim, min(int32(*frame), anim.FrameCount-1))
	}
	target := rl.LoadRenderTexture(1200, 1000)
	defer rl.UnloadRenderTexture(target)
	type shot struct {
		name  string
		angle float32
		close bool
	}
	for _, s := range []shot{{"front", 0, false}, {"three-quarter", 35, false}, {"side", 90, false}, {"back", 180, false}, {"face-front", 0, true}, {"face-three-quarter", 35, true}, {"face-side", 90, true}} {
		camera := rl.Camera3D{Position: rl.NewVector3(0, 1.12, 4), Target: rl.NewVector3(0, 1.08, 0), Up: rl.NewVector3(0, 1, 0), Fovy: 2.32, Projection: rl.CameraOrthographic}
		if s.close {
			camera.Position = rl.NewVector3(0, 1.81, 4)
			camera.Target = rl.NewVector3(0, 1.81, 0)
			camera.Fovy = .58
		}
		rl.BeginTextureMode(target)
		rl.ClearBackground(rl.NewColor(36, 42, 49, 255))
		rl.BeginMode3D(camera)
		if !s.close {
			rl.DrawPlane(rl.Vector3{}, rl.NewVector2(6, 6), rl.NewColor(49, 55, 60, 255))
		}
		rl.DrawModelEx(model.Model, rl.Vector3{}, rl.NewVector3(0, 1, 0), s.angle, rl.NewVector3(model.Scale, model.Scale, model.Scale), rl.White)
		rl.EndMode3D()
		rl.EndTextureMode()
		img := rl.LoadImageFromTexture(target.Texture)
		rl.ImageFlipVertical(img)
		if !rl.ExportImage(*img, filepath.Join(*out, s.name+".png")) {
			panic("PNG export failed")
		}
		rl.UnloadImage(img)
	}
	if *compare != "" {
		original := assets.Load(*compare, assets.ManifestEntry{}, shader)
		if !original.Available {
			panic(original.Error)
		}
		defer original.Close()
		fonts := ui.New()
		defer fonts.Close()
		camera := rl.Camera3D{Position: rl.NewVector3(0, 1.12, 4), Target: rl.NewVector3(0, 1.08, 0), Up: rl.NewVector3(0, 1, 0), Fovy: 2.45, Projection: rl.CameraOrthographic}
		rl.BeginTextureMode(target)
		rl.ClearBackground(rl.NewColor(22, 29, 39, 255))
		rl.BeginMode3D(camera)
		for i, m := range []*assets.Model{original, model} {
			pos := rl.NewVector3(-.66+float32(i)*1.32, 0, 0)
			rl.DrawModelEx(m.Model, pos, rl.NewVector3(0, 1, 0), -12, rl.NewVector3(m.Scale, m.Scale, m.Scale), rl.White)
		}
		rl.EndMode3D()
		rl.DrawTextEx(fonts.Bold, "ИСХОДНАЯ МОДЕЛЬ", rl.NewVector2(130, 35), 30, 0, ui.Text)
		rl.DrawTextEx(fonts.Bold, "ИВАНОВ ПО РЕФЕРЕНСАМ", rl.NewVector2(680, 35), 30, 0, ui.Amber)
		rl.DrawTextEx(fonts.Font, "Один скелет и те же 18 анимаций. Кадр из нативного 3D-рендера игры.", rl.NewVector2(145, 945), 23, 0, ui.Text)
		rl.EndTextureMode()
		img := rl.LoadImageFromTexture(target.Texture)
		rl.ImageFlipVertical(img)
		if !rl.ExportImage(*img, filepath.Join(*out, "comparison.png")) {
			panic("Comparison export failed")
		}
		rl.UnloadImage(img)
	}
	fmt.Printf("Review saved: %s\n", *out)
}
