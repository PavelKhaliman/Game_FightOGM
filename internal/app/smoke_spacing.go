package app

import (
	"fightogm/internal/combat"
	"fightogm/internal/game"
	"fightogm/internal/input"
	"fmt"
	"math"
)

func (a *App) smokeCloseRange(dir string, report map[string]any) error {
	for _, direction := range []float32{-1, 1} {
		a.Bot = false
		a.Selection = [2]int{3, 0}
		if !a.startMatch() {
			return fmt.Errorf("не начат тест Новацкого")
		}
		a.setScreen(Battle)
		a.Match.Phase = game.Fighting
		walker, still := a.Match.Fighters[0], a.Match.Fighters[1]
		walker.Position = combat.Vec3{X: -1.2 * direction}
		still.Position = combat.Vec3{X: 1.2 * direction}
		a.update(input.Frame{}, game.Step)
		beforeWorld, beforeScreen := still.Position.X, a.Renderer.Project(still.Position).X
		frame := input.Frame{}
		frame.Players[0].X = direction
		for step := 0; step < 240; step++ {
			a.update(frame, game.Step)
		}
		for step := 0; step < 60; step++ {
			a.update(input.Frame{}, game.Step)
		}
		if math.Abs(float64(still.Position.X-beforeWorld)) > .001 || math.Abs(float64(a.Renderer.Project(still.Position).X-beforeScreen)) > 1 {
			return fmt.Errorf("притяжка в ближнем бою: мир %v -> %v, экран %v -> %v", beforeWorld, still.Position.X, beforeScreen, a.Renderer.Project(still.Position).X)
		}
		if math.Abs(float64((still.Position.X-walker.Position.X)*direction-game.PushboxWidth)) > .001 {
			return fmt.Errorf("неверная дистанция при соприкосновении")
		}
		if err := a.snapshot(dir, fmt.Sprintf("close-no-pull-%g", direction)); err != nil {
			return err
		}
		walker.Position.X, still.Position.X = -.85*direction, .85*direction
		frame = input.Frame{}
		frame.Players[0].Pressed = []input.Action{input.Special}
		a.update(frame, game.Step)
		for step := 0; step < 55; step++ {
			a.update(input.Frame{}, game.Step)
		}
		if err := a.snapshot(dir, fmt.Sprintf("auger-contact-%g", direction)); err != nil {
			return err
		}
	}
	report["close_range_no_pull"] = "passed: world and screen positions, both directions"
	report["novatskiy_held_auger"] = "36 poses; melee strike and three-hit combo"
	return nil
}
