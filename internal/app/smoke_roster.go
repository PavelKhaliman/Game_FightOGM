package app

import (
	"fightogm/internal/combat"
	"fightogm/internal/game"
	"fightogm/internal/input"
	"fightogm/internal/ui"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"sort"
)

func (a *App) smokeRoster(dir string, report map[string]any) error {
	totalFrames, totalClips, actionsPlayed := 0, 0, 0
	projectiles := map[string]bool{}
	for index, d := range a.Roster {
		atlas := a.Renderer.Atlases[d.ID]
		totalFrames += len(atlas.Frames)
		totalClips += len(atlas.Clips)
		if len(atlas.Frames) != 36 {
			return fmt.Errorf("%s: ожидались 36 кадров, получено %d", d.Name, len(atlas.Frames))
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
		image := rl.LoadImageFromTexture(a.Target.Texture)
		green, visible := 0, 0
		for _, p := range rl.LoadImageColors(image) {
			if int(p.G)-max(int(p.R), int(p.B)) > 80 {
				green++
			}
			if p.R > 50 || p.B > 50 {
				visible++
			}
		}
		rl.UnloadImage(image)
		if green > 20 || visible < 20000 {
			return fmt.Errorf("%s: хромакей/спрайты: зелёных %d, видимых %d", d.Name, green, visible)
		}
		if err := a.exportTarget(dir, "atlas-"+d.ID); err != nil {
			return err
		}
		a.Bot = false
		a.Selection = [2]int{index, index}
		if !a.startMatch() {
			return fmt.Errorf("не удалось начать зеркальный бой %s", d.Name)
		}
		if a.Renderer.Actors[0] == a.Renderer.Actors[1] {
			return fmt.Errorf("общий контроллер зеркальных бойцов")
		}
		a.setScreen(Battle)
		a.Match.Phase = game.Fighting
		frame := input.Frame{}
		frame.Players[0].Jump = true
		a.update(frame, game.Step)
		a.update(input.Frame{}, game.Step)
		if a.Renderer.Actors[0].Controller.Name == a.Renderer.Actors[1].Controller.Name {
			return fmt.Errorf("связаны анимации %s", d.Name)
		}
		for _, action := range []input.Action{input.Special, input.Secondary, input.Ultimate} {
			a.Selection = [2]int{index, 0}
			a.startMatch()
			a.setScreen(Battle)
			a.Match.Phase = game.Fighting
			f, enemy := a.Match.Fighters[0], a.Match.Fighters[1]
			f.Position = combat.Vec3{X: -.48}
			enemy.Position = combat.Vec3{X: .48}
			f.Meter = 100
			moveID := "special"
			if action == input.Secondary {
				moveID = "secondary"
			}
			if action == input.Ultimate {
				moveID = "ultimate"
			}
			move := d.Moves[moveID]
			if d.ID == "novatskiy" && move.Damage > 0 {
				f.Position.X = -.85
				enemy.Position.X = .85
			}
			if move.Projectile {
				f.Position.X = -1.4
				enemy.Position.X = 1.4
			}
			frame := input.Frame{}
			frame.Players[0].Pressed = []input.Action{action}
			a.update(frame, game.Step)
			if f.Move == nil || f.Move.ID != moveID || a.Renderer.Actors[0].Controller.Name != move.Animation {
				return fmt.Errorf("%s: не запущен %s", d.Name, moveID)
			}
			captured := false
			for step := 0; step < 420; step++ {
				a.update(input.Frame{}, game.Step)
				for _, fighter := range a.Match.Fighters {
					if fighter.Position.Z != 0 || fighter.Velocity.Z != 0 {
						return fmt.Errorf("%s: %s вышел из 2D", d.Name, moveID)
					}
				}
				visibleProjectile := false
				for _, p := range a.Match.Projectiles {
					if p.Delay <= 0 && !p.Hit {
						visibleProjectile = true
						projectiles[p.Kind] = true
					}
				}
				if !captured && ((move.Projectile && visibleProjectile) || (!move.Projectile && ((move.Defense && step == 12) || (!move.Defense && f.Move != nil && f.MoveTime >= move.Startup+game.Step)))) {
					if err := a.snapshot(dir, "move-"+d.ID+"-"+moveID); err != nil {
						return err
					}
					captured = true
				}
			}
			if !captured {
				return fmt.Errorf("%s: не отображён %s", d.Name, moveID)
			}
			if move.Damage > 0 && enemy.HP >= enemy.Definition.MaxHP {
				return fmt.Errorf("%s: %s не нанёс урон", d.Name, moveID)
			}
			if action == input.Ultimate && f.Meter >= 100 {
				return fmt.Errorf("%s: суперприём не расходует энергию", d.Name)
			}
			actionsPlayed++
		}
	}
	if len(projectiles) != 6 || !projectiles["foam"] || !projectiles["arc"] {
		return fmt.Errorf("не проверены все типы снарядов: %v", projectiles)
	}
	report["sprite_frames"] = totalFrames
	report["clips"] = totalClips
	report["special_actions_played"] = actionsPlayed
	report["projectile_types_rendered"] = projectiles
	report["all_mirror_matches"] = "passed"
	report["all_atlases_chroma_key"] = "passed"
	return nil
}
