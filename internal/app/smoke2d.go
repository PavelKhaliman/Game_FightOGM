package app

import (
	"encoding/json"
	"fightogm/internal/combat"
	"fightogm/internal/game"
	"fightogm/internal/input"
	"fightogm/internal/ui"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func (a *App) exportTarget(dir, name string) error {
	img := rl.LoadImageFromTexture(a.Target.Texture)
	defer rl.UnloadImage(img)
	rl.ImageFlipVertical(img)
	if !rl.ExportImage(*img, filepath.Join(dir, name+".png")) {
		return fmt.Errorf("не удалось сохранить %s", name)
	}
	return nil
}
func (a *App) snapshot(dir, name string) error {
	a.draw()
	a.present()
	return a.exportTarget(dir, name)
}

// smoke runs against the actual native renderer, including shader pixel checks.
func (a *App) smoke(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	rl.SetTargetFPS(0)
	report := map[string]any{"renderer": "raylib native 2D sprites", "fighters": len(a.Roster), "3d_models_loaded": 0}
	if len(a.Roster) != 12 || a.Roster[0].ID != "ivanov" {
		return fmt.Errorf("ожидались двенадцать 2D-бойцов")
	}
	atlas := a.Renderer.Atlases["ivanov"]
	report["sprite_frames"] = len(atlas.Frames)
	report["clips"] = len(atlas.Clips)
	if len(atlas.Frames) < 36 {
		return fmt.Errorf("неполный набор кадров")
	}
	ids := make([]string, 0, len(atlas.Frames))
	for id := range atlas.Frames {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rl.BeginTextureMode(a.Target)
	rl.ClearBackground(ui.Background)
	for i, id := range ids {
		x, y := float32(i%6)*266, float32(i/6)*150
		a.Renderer.DrawFrame(atlas, id, rl.NewVector2(x+130, y+128), 50, i%2 == 1, rl.White)
		a.UI.Text(id, x+12, y+131, 15, ui.Text)
	}
	rl.EndTextureMode()
	img := rl.LoadImageFromTexture(a.Target.Texture)
	pixels := rl.LoadImageColors(img)
	green, drawn := 0, 0
	for _, p := range pixels {
		if int(p.G)-max(int(p.R), int(p.B)) > 80 {
			green++
		}
		if p.R > 50 || p.B > 50 {
			drawn++
		}
	}
	rl.UnloadImage(img)
	if green > 20 || drawn < 20000 {
		return fmt.Errorf("ошибка спрайтов/хромакея: зелёных %d, видимых %d", green, drawn)
	}
	report["chroma_key_green_pixels"] = green
	if e := a.exportTarget(dir, "00-sprite-atlas"); e != nil {
		return e
	}
	if e := a.snapshot(dir, "01-main-menu"); e != nil {
		return e
	}
	a.update(input.Frame{Accept: true}, game.Step)
	if a.Screen != Select || !a.Bot || !a.Ready[1] {
		return fmt.Errorf("меню боя с компьютером не работает")
	}
	a.update(input.Frame{SelectionX: [2]int{-1, 1}, SelectionY: [2]int{1, -1}}, game.Step)
	if a.Selection != [2]int{11, 8} {
		return fmt.Errorf("выбор по сетке не работает: %v", a.Selection)
	}
	// Select both new fighters through the same mouse path as the player.
	for p, index := range [2]int{10, 11} {
		header, card := ui.PlayerRect(p), ui.CardRect(index)
		a.update(input.Frame{MouseClick: true, MouseX: header.X + 10, MouseY: header.Y + 10}, game.Step)
		a.update(input.Frame{MouseClick: true, MouseX: card.X + 10, MouseY: card.Y + 10}, game.Step)
	}
	if a.Selection != [2]int{10, 11} {
		return fmt.Errorf("мышь не выбирает новых бойцов: %v", a.Selection)
	}
	if e := a.snapshot(dir, "02-character-select"); e != nil {
		return e
	}
	a.Selection = [2]int{0, 0}
	a.update(input.Frame{Accept: true}, game.Step)
	if a.Screen != Versus {
		return fmt.Errorf("подтверждение бойца не работает")
	}
	if e := a.snapshot(dir, "03-versus"); e != nil {
		return e
	}
	a.update(input.Frame{Accept: true}, game.Step)
	for i := 0; i < 280; i++ {
		a.update(input.Frame{}, game.Step)
	}
	if a.Match.Phase != game.Fighting {
		return fmt.Errorf("раунд не начался")
	}
	if e := a.snapshot(dir, "04-arena"); e != nil {
		return e
	}
	for i := 0; i < 1000; i++ {
		a.update(input.Frame{}, game.Step)
	}
	if a.Match.Fighters[0].HP >= a.Roster[0].MaxHP {
		return fmt.Errorf("компьютер не атакует")
	}
	report["cpu_match"] = "passed"
	a.update(input.Frame{Back: true}, game.Step)
	timer := a.Match.Timer
	a.update(input.Frame{}, .5)
	if a.Screen != Pause || a.Match.Timer != timer {
		return fmt.Errorf("пауза не останавливает бой")
	}
	if e := a.snapshot(dir, "05-pause"); e != nil {
		return e
	}
	a.setScreen(Menu)
	a.MenuIndex = 1
	a.update(input.Frame{Accept: true}, game.Step)
	if a.Bot || a.Ready != [2]bool{} {
		return fmt.Errorf("не включился режим двух игроков")
	}
	a.update(input.Frame{Confirm: [2]bool{true, false}}, game.Step)
	if a.Screen != Select || !a.Ready[0] || a.Ready[1] {
		return fmt.Errorf("игроки должны подтверждать отдельно")
	}
	a.update(input.Frame{Confirm: [2]bool{false, true}}, game.Step)
	if a.Screen != Versus {
		return fmt.Errorf("локальный бой не запустился")
	}
	a.Selection = [2]int{0, 0}
	reset := func() {
		a.startMatch()
		a.setScreen(Battle)
		a.Match.Phase = game.Fighting
		a.Match.Fighters[0].Position = combat.Vec3{X: -.5}
		a.Match.Fighters[1].Position = combat.Vec3{X: .5}
	}
	for _, scenario := range []struct {
		Name          string
		Action        input.Action
		Block, Crouch bool
	}{
		{"06-punch", input.Light, false, false}, {"07-block", input.Heavy, true, false},
		{"08-kick", input.Kick, false, false}, {"09-sweep", input.Kick, false, true},
		{"10-grab", input.Grab, true, false}, {"11-hearing", input.Special, false, false},
		{"12-invisible", input.Secondary, false, false}, {"13-omni", input.Ultimate, false, false},
	} {
		reset()
		f := a.Match.Fighters[0]
		f.Meter = 100
		frame := input.Frame{}
		frame.Players[0] = input.InputState{Crouch: scenario.Crouch, Pressed: []input.Action{scenario.Action}}
		frame.Players[1].Block = scenario.Block
		a.update(frame, game.Step)
		if f.Move == nil {
			return fmt.Errorf("приём не запущен: %s", scenario.Name)
		}
		impact := f.Move.Startup + f.Move.Active*.45
		if scenario.Action == input.Secondary || scenario.Action == input.Ultimate || scenario.Action == input.Special {
			impact = .5
		}
		for step := 0; step < 180 && f.Move != nil && f.MoveTime < impact; step++ {
			frame.Players[0].Pressed = nil
			a.update(frame, game.Step)
		}
		if scenario.Action <= input.Grab && scenario.Action != input.Special && !scenario.Block && a.Match.Fighters[1].HP >= a.Roster[0].MaxHP {
			return fmt.Errorf("приём не попал: %s", scenario.Name)
		}
		if scenario.Action == input.Heavy && (a.Match.Fighters[1].HP >= a.Roster[0].MaxHP || a.Match.Fighters[1].HP < a.Roster[0].MaxHP-20) {
			return fmt.Errorf("неверный урон через блок")
		}
		if e := a.snapshot(dir, scenario.Name); e != nil {
			return e
		}
	}
	reset()
	a.Match.Fighters[1].Position.X = 2.0
	frame := input.Frame{}
	frame.Players[0] = input.InputState{X: 1, Jump: true}
	a.update(frame, game.Step)
	frame.Players[0].Jump = false
	for i := 0; i < 28; i++ {
		a.update(frame, game.Step)
	}
	if a.Match.Fighters[0].Position.Y <= 0 || a.Match.Fighters[0].Position.Z != 0 {
		return fmt.Errorf("прыжок вышел из 2D-плоскости")
	}
	if a.Renderer.Actors[0].Controller.FrameID == a.Renderer.Actors[1].Controller.FrameID {
		return fmt.Errorf("состояния анимаций двух Ивановых связаны")
	}
	if e := a.snapshot(dir, "14-jump"); e != nil {
		return e
	}
	for i := 0; i < 30; i++ {
		a.update(frame, game.Step)
	}
	for i, f := range a.Match.Fighters {
		actor := a.Renderer.Actors[i]
		h := actor.Atlas.Frames[actor.Controller.FrameID].WorldHeight
		if a.Renderer.Project(f.Position.Add(combat.Vec3{Y: h})).Y < 180 {
			return fmt.Errorf("прыжок перекрыт интерфейсом")
		}
	}
	if e := a.snapshot(dir, "14-jump-apex"); e != nil {
		return e
	}
	a.Debug = true
	if e := a.snapshot(dir, "15-debug"); e != nil {
		return e
	}
	a.Debug = false
	report["mirror_match"] = "independent sprite animation and combat state"
	reset()
	for rounds := 0; rounds < 2; rounds++ {
		a.Match.Fighters[1].HP = 1
		for step := 0; step < 2200 && a.Screen != Results; step++ {
			frame := input.Frame{}
			if a.Match.Phase == game.Fighting {
				a.Match.Fighters[0].Position = combat.Vec3{X: -.48}
				a.Match.Fighters[1].Position = combat.Vec3{X: .48}
				if step%80 == 0 {
					frame.Players[0].Pressed = []input.Action{input.Light}
				}
			}
			a.update(frame, game.Step)
			if a.Match.Phase == game.RoundOver && a.Match.PhaseTime > 1.5 && a.Match.PhaseTime < 1.5+game.Step {
				if e := a.snapshot(dir, fmt.Sprintf("16-ko-%d", rounds+1)); e != nil {
					return e
				}
			}
			if a.Match.Phase == game.Intro && a.Match.Wins[0] == rounds+1 {
				break
			}
		}
	}
	if a.Screen != Results || a.Match.Wins[0] != 2 {
		return fmt.Errorf("матч не завершён после двух побед: %v", a.Match.Wins)
	}
	if e := a.snapshot(dir, "17-results"); e != nil {
		return e
	}
	a.update(input.Frame{Accept: true}, game.Step)
	if a.Screen != Versus || a.Match.Wins != [2]int{} {
		return fmt.Errorf("реванш не сбросил счёт")
	}
	report["local_match_rounds_pause_rematch"] = "passed"
	a.setScreen(Settings)
	if e := a.snapshot(dir, "18-settings"); e != nil {
		return e
	}
	a.setScreen(Controls)
	if e := a.snapshot(dir, "19-controls"); e != nil {
		return e
	}
	a.setScreen(Menu)
	a.MenuIndex = 2
	a.update(input.Frame{Accept: true}, game.Step)
	if a.Screen != Controls {
		return fmt.Errorf("пункт управления не открывается")
	}
	a.update(input.Frame{Back: true}, game.Step)
	a.MenuIndex = 3
	a.update(input.Frame{Accept: true}, game.Step)
	if a.Screen != Settings {
		return fmt.Errorf("пункт настроек не открывается")
	}
	a.setScreen(Menu)
	a.MenuIndex = 4
	a.update(input.Frame{Accept: true}, game.Step)
	if !a.Quit {
		return fmt.Errorf("пункт выхода не работает")
	}
	a.Quit = false
	if err := a.smokeRoster(dir, report); err != nil {
		return err
	}
	if err := a.smokeCloseRange(dir, report); err != nil {
		return err
	}
	a.Selection = [2]int{0, 0}
	reset()
	start := time.Now()
	for frame := 0; frame < 180; frame++ {
		a.update(input.Frame{}, 1.0/60)
		a.draw()
		a.present()
	}
	report["render_benchmark_fps"] = 180 / time.Since(start).Seconds()
	report["status"] = "PASS"
	data, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(dir, "report.json"), data, 0644); e != nil {
		return e
	}
	log.Printf("[FightOGM] SMOKE PASS: 2D, %d бойцов, все приёмы, компьютер, локальный бой, матч и реванш", len(a.Roster))
	return nil
}
