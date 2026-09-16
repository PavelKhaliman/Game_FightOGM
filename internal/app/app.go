package app

import (
	"fightogm/internal/assets"
	"fightogm/internal/audio"
	"fightogm/internal/characters"
	"fightogm/internal/combat"
	"fightogm/internal/config"
	"fightogm/internal/fighter"
	"fightogm/internal/game"
	"fightogm/internal/input"
	"fightogm/internal/render"
	"fightogm/internal/ui"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"os"
	"path/filepath"
)

type Screen uint8

const (
	Menu Screen = iota
	Select
	Versus
	Battle
	Pause
	Controls
	Settings
	Results
)

type Options struct {
	Root   string
	Smoke  bool
	Output string
}
type App struct {
	Root                          string
	Screen, ReturnScreen          Screen
	Settings                      config.Settings
	UI                            *ui.UI
	Renderer                      *render.Renderer
	Audio                         *audio.Manager
	Target                        rl.RenderTexture2D
	Roster                        []fighter.Definition
	Match                         *game.Match
	Selection                     [2]int
	Ready                         [2]bool
	MenuIndex                     int
	SelectionPlayer               int
	Clock, Accumulator, ToastTime float32
	Toast, Message                string
	Debug, Quit                   bool
	Bot                           bool
	CPU                           game.CPU
}

func FindRoot(explicit string) (string, error) {
	if explicit != "" {
		p, e := filepath.Abs(explicit)
		if e != nil {
			return "", e
		}
		if _, e = os.Stat(filepath.Join(p, "assets/sprites/roster.json")); e != nil {
			return "", e
		}
		return p, nil
	}
	cwd, _ := os.Getwd()
	exe, _ := os.Executable()
	for _, p := range []string{cwd, filepath.Dir(exe), filepath.Dir(filepath.Dir(exe))} {
		if _, e := os.Stat(filepath.Join(p, "assets/sprites/roster.json")); e == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("не найден assets/sprites/roster.json; запустите из папки игры или укажите -assets-root")
}
func Run(opts Options) error {
	root, e := FindRoot(opts.Root)
	if e != nil {
		return e
	}
	manifest, e := assets.ReadSpriteRoster(root)
	if e != nil {
		return fmt.Errorf("манифест: %w", e)
	}
	a := &App{Root: root, Settings: config.Load(filepath.Join(root, "config/settings.json"))}
	for _, id := range manifest.Characters {
		found := false
		for _, d := range characters.Roster() {
			if d.ID == id {
				a.Roster = append(a.Roster, d)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("неизвестный боец в списке 2D: %s", id)
		}
	}
	flags := uint32(rl.FlagWindowResizable | rl.FlagMsaa4xHint | rl.FlagVsyncHint)
	if opts.Smoke {
		flags = uint32(rl.FlagWindowHidden)
	}
	rl.SetConfigFlags(flags)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.InitWindow(int32(a.Settings.Width), int32(a.Settings.Height), "FIGHTOGM 2D — Производство не останавливается")
	if !rl.IsWindowReady() {
		return fmt.Errorf("не удалось открыть окно OpenGL")
	}
	defer rl.CloseWindow()
	rl.SetExitKey(rl.KeyNull)
	rl.SetWindowMinSize(960, 540)
	rl.SetTargetFPS(60)
	if a.Settings.Fullscreen && !opts.Smoke {
		rl.ToggleFullscreen()
	}
	a.UI = ui.New()
	defer a.UI.Close()
	a.Target = rl.LoadRenderTexture(1600, 900)
	defer rl.UnloadRenderTexture(a.Target)
	log.Printf("[FightOGM] Загрузка 2D-спрайтов: %d бойцов...", len(a.Roster))
	a.Renderer, e = render.New(root, a.Roster, manifest, a.UI.Font, func(i int, name string) {
		rl.BeginDrawing()
		rl.ClearBackground(ui.Background)
		scale := float32(rl.GetScreenWidth()) / 1600
		rl.DrawTextEx(a.UI.Bold, "FIGHTOGM", rl.NewVector2(64*scale, 190*scale), 72*scale, 1, ui.Text)
		rl.DrawTextEx(a.UI.Font, "Загрузка: "+name, rl.NewVector2(70*scale, 320*scale), 27*scale, 0, ui.Amber)
		rl.DrawRectangle(70, int32(420*scale), int32(float32(i)*100*scale), 6, ui.Amber)
		rl.EndDrawing()
	})
	if e != nil {
		return e
	}
	defer a.Renderer.Close()
	a.Audio = audio.New(root)
	defer a.Audio.Close()
	a.Audio.Volumes(a.Settings.Master, a.Settings.Music, a.Settings.SFX)
	if opts.Smoke {
		return a.smoke(opts.Output)
	}
	for !rl.WindowShouldClose() && !a.Quit {
		dt := min(.1, rl.GetFrameTime())
		frame := input.Poll(a.Settings.Mappings)
		a.mapMouse(&frame)
		a.update(frame, dt)
		a.draw()
		a.present()
		a.Audio.Update()
	}
	if e = a.Settings.Save(filepath.Join(root, "config/settings.json")); e != nil {
		log.Printf("[ПРЕДУПРЕЖДЕНИЕ] Не удалось сохранить настройки: %v", e)
	}
	return nil
}
func (a *App) mapMouse(f *input.Frame) {
	scale := min(float32(rl.GetScreenWidth())/1600, float32(rl.GetScreenHeight())/900)
	f.MouseX = (f.MouseX - (float32(rl.GetScreenWidth())-1600*scale)*.5) / scale
	f.MouseY = (f.MouseY - (float32(rl.GetScreenHeight())-900*scale)*.5) / scale
}
func (a *App) setScreen(s Screen) { a.Screen = s; a.MenuIndex = 0; a.Clock = 0; a.Accumulator = 0 }
func (a *App) startMatch() bool {
	m := game.New(a.Roster[a.Selection[0]], a.Roster[a.Selection[1]])
	if e := a.Renderer.SetMatch(m); e != nil {
		a.Message = e.Error()
		log.Print(e)
		a.setScreen(Select)
		return false
	}
	a.Match = m
	a.CPU = game.CPU{}
	a.Toast = ""
	a.ToastTime = 0
	a.setScreen(Versus)
	return true
}
func (a *App) fullscreen() { rl.ToggleFullscreen(); a.Settings.Fullscreen = rl.IsWindowFullscreen() }
func (a *App) update(f input.Frame, dt float32) {
	a.Clock += dt
	a.ToastTime = max(0, a.ToastTime-dt)
	if f.Fullscreen {
		a.fullscreen()
	}
	if f.Debug {
		a.Debug = !a.Debug
	}
	oldIndex := a.MenuIndex
	switch a.Screen {
	case Menu:
		a.Renderer.UpdateShowcase(dt, [2]int{0, 0})
		choice := ui.MainLayout.Handle(f, &a.MenuIndex, len(ui.MainLabels))
		switch choice {
		case 0, 1:
			a.Bot = choice == 0
			a.Ready = [2]bool{false, a.Bot}
			a.Selection = [2]int{0, min(1, len(a.Roster)-1)}
			a.SelectionPlayer = 0
			a.Message = ""
			a.setScreen(Select)
		case 2:
			a.ReturnScreen = Menu
			a.setScreen(Controls)
		case 3:
			a.setScreen(Settings)
		case 4:
			a.Quit = true
		}
	case Select:
		a.Renderer.UpdateShowcase(dt, a.Selection)
		if f.Back {
			a.setScreen(Menu)
			break
		}
		for i := 0; i < 2; i++ {
			if i == 1 && a.Bot {
				a.Ready[i] = true
			}
			if f.Cancel[i] && !(i == 1 && a.Bot) {
				a.Ready[i] = false
			}
			if a.Ready[i] && !(i == 1 && a.Bot) {
				continue
			}
			dx, dy := f.SelectionX[i], f.SelectionY[i]
			if dx != 0 || dy != 0 {
				a.Selection[i] = gridSelection(a.Selection[i], dx, dy, len(a.Roster))
				a.SelectionPlayer = i
				a.Audio.Play("menu")
			}
			if f.Confirm[i] || f.Accept {
				a.Ready[i] = true
				a.Message = ""
				a.Audio.Play("menu")
				if i == 0 && !a.Bot {
					a.SelectionPlayer = 1
				}
			}
		}
		if f.MouseClick {
			point := rl.NewVector2(f.MouseX, f.MouseY)
			for i := 0; i < 2; i++ {
				if rl.CheckCollisionPointRec(point, ui.PlayerRect(i)) {
					a.SelectionPlayer = i
				}
				if rl.CheckCollisionPointRec(point, ui.ReadyRect(i)) {
					a.Ready[i] = true
					if i == 0 && !a.Bot {
						a.SelectionPlayer = 1
					}
				}
			}
			for index := range a.Roster {
				if rl.CheckCollisionPointRec(point, ui.CardRect(index)) {
					p := a.SelectionPlayer
					if !a.Ready[p] || (p == 1 && a.Bot) {
						a.Selection[p] = index
						a.Audio.Play("menu")
					}
				}
			}
		}
		if a.Ready[0] && a.Ready[1] {
			a.startMatch()
		}
	case Versus:
		a.Renderer.Update(a.Match, dt, dt)
		if a.Clock > 1.8 || f.Accept {
			a.setScreen(Battle)
		}
		if f.Back {
			a.Ready = [2]bool{false, a.Bot}
			a.setScreen(Select)
		}
	case Battle:
		if f.Back {
			a.setScreen(Pause)
			break
		}
		if a.Bot {
			f.Players[1] = a.CPU.Input(a.Match, dt)
		}
		a.Match.Submit(f.Players)
		before := a.Match.Time
		a.Accumulator += dt
		for a.Accumulator >= game.Step {
			a.Match.Tick(game.Step)
			a.Accumulator -= game.Step
		}
		animationDT := a.Match.Time - before
		if a.Match.Phase != game.Fighting {
			animationDT = dt
			if a.Match.Phase == game.RoundOver && a.Match.PhaseTime < 1 {
				animationDT *= .2
			}
		}
		if a.Match.HitStop > 0 {
			animationDT = 0
		}
		a.Renderer.Update(a.Match, dt, animationDT)
		a.events()
		if a.Match.Phase == game.MatchOver {
			a.setScreen(Results)
		}
	case Pause:
		if f.Back {
			a.setScreen(Battle)
			break
		}
		switch ui.OverlayLayout.Handle(f, &a.MenuIndex, len(ui.PauseLabels)) {
		case 0:
			a.setScreen(Battle)
		case 1:
			a.Match.ResetRound()
			a.setScreen(Battle)
		case 2:
			a.ReturnScreen = Pause
			a.setScreen(Controls)
		case 3:
			a.setScreen(Menu)
		}
	case Controls:
		if f.Back || f.Accept || f.MouseClick {
			a.setScreen(a.ReturnScreen)
		}
	case Settings:
		choice := (ui.Layout{X: 288, Y: 246, W: 1024, H: 63, Gap: 15}).Handle(f, &a.MenuIndex, 6)
		delta := 0
		if f.Left {
			delta = -1
		}
		if f.Right || choice >= 0 {
			delta = 1
		}
		if a.MenuIndex < 5 && delta != 0 {
			a.changeSetting(a.MenuIndex, delta)
		}
		if f.Back || choice == 5 {
			if e := a.Settings.Save(filepath.Join(a.Root, "config/settings.json")); e != nil {
				a.Toast = "Не удалось сохранить настройки"
				a.ToastTime = 3
				log.Print(e)
			}
			a.setScreen(Menu)
		}
	case Results:
		a.Renderer.Update(a.Match, dt, dt)
		switch ui.OverlayLayout.Handle(f, &a.MenuIndex, len(ui.ResultLabels)) {
		case 0:
			a.startMatch()
		case 1:
			a.Ready = [2]bool{false, a.Bot}
			a.setScreen(Select)
		case 2:
			a.setScreen(Menu)
		}
	}
	if oldIndex != a.MenuIndex || f.Accept {
		a.Audio.Play("menu")
	}
}
func (a *App) changeSetting(index, delta int) {
	switch index {
	case 0:
		sizes := [][2]int{{1280, 720}, {1600, 900}, {1920, 1080}, {2560, 1440}}
		at := 1
		for i, s := range sizes {
			if a.Settings.Width == s[0] {
				at = i
			}
		}
		at = (at + delta + len(sizes)) % len(sizes)
		a.Settings.Width = sizes[at][0]
		a.Settings.Height = sizes[at][1]
		if !a.Settings.Fullscreen {
			rl.SetWindowSize(a.Settings.Width, a.Settings.Height)
		}
	case 1:
		a.fullscreen()
	case 2:
		a.Settings.Master = combat.Clamp(a.Settings.Master+float32(delta)*.05, 0, 1)
	case 3:
		a.Settings.Music = combat.Clamp(a.Settings.Music+float32(delta)*.05, 0, 1)
	case 4:
		a.Settings.SFX = combat.Clamp(a.Settings.SFX+float32(delta)*.05, 0, 1)
	}
	a.Audio.Volumes(a.Settings.Master, a.Settings.Music, a.Settings.SFX)
}
func (a *App) events() {
	a.Renderer.Events(a.Match.Events)
	for _, e := range a.Match.Events {
		a.Audio.Play(e.Kind)
		if e.Text != "" && e.Kind != "ko" && e.Kind != "victory" {
			a.Toast = e.Text
			a.ToastTime = 1.55
		}
	}
	a.Match.Events = a.Match.Events[:0]
}
func (a *App) draw() {
	rl.BeginTextureMode(a.Target)
	rl.ClearBackground(ui.Background)
	switch a.Screen {
	case Menu:
		a.Renderer.DrawShowcase([2]int{0, 0}, true)
		a.UI.Main(a.MenuIndex)
	case Select:
		a.Renderer.DrawSelection(a.Selection)
		a.UI.Select(a.Roster, a.Selection, a.Ready, a.Bot, a.SelectionPlayer, a.Message, a.Settings, a.Renderer.DrawPortrait)
	case Versus:
		a.Renderer.DrawShowcase(a.Selection, false)
		a.UI.Versus(a.Match.Fighters[0].Definition, a.Match.Fighters[1].Definition)
	case Battle, Pause, Results:
		a.Renderer.DrawBattle(a.Match, a.Debug)
		a.UI.HUD(a.Match, a.Settings, a.Bot)
		if a.ToastTime > 0 && a.Match.Phase == game.Fighting {
			a.UI.Center(a.Toast, 185, 27, ui.Amber)
		}
		if a.Debug {
			a.debugHUD()
		}
		if a.Screen == Pause {
			a.UI.Overlay("ПАУЗА", "СМЕНА ПОДОЖДЁТ", ui.PauseLabels, a.MenuIndex)
		}
		if a.Screen == Results {
			winner := a.Match.Winner
			if winner >= 0 {
				tag := fmt.Sprintf("ИГРОК %d", winner+1)
				if winner == 1 && a.Bot {
					tag = "КОМПЬЮТЕР"
				}
				a.UI.Overlay(tag+" ПОБЕЖДАЕТ", a.Match.Fighters[winner].Definition.VictoryText, ui.ResultLabels, a.MenuIndex)
			}
		}
	case Controls:
		a.UI.Controls(a.Settings)
	case Settings:
		a.UI.Settings(a.Settings, a.MenuIndex)
	}
	rl.EndTextureMode()
}
func (a *App) present() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	scale := min(float32(rl.GetScreenWidth())/1600, float32(rl.GetScreenHeight())/900)
	dest := rl.NewRectangle((float32(rl.GetScreenWidth())-1600*scale)/2, (float32(rl.GetScreenHeight())-900*scale)/2, 1600*scale, 900*scale)
	rl.DrawTexturePro(a.Target.Texture, rl.NewRectangle(0, 0, 1600, -900), dest, rl.Vector2{}, 0, rl.White)
	rl.EndDrawing()
}
func (a *App) debugHUD() {
	ui.Rect(16, 330, 500, 270, rl.Fade(ui.Background, .87))
	a.UI.Text(fmt.Sprintf("ОТЛАДКА  ·  %d кадров/с  ·  120 шагов/с", rl.GetFPS()), 30, 340, 18, ui.Amber)
	for i, f := range a.Match.Fighters {
		model := a.Renderer.Actors[i]
		y := float32(373 + i*106)
		a.UI.Text(fmt.Sprintf("ИГР.%d  %s   HP %.0f   Энергия %.0f", i+1, f.State, f.HP, f.Meter), 30, y, 17, ui.Text)
		a.UI.Text(fmt.Sprintf("Анимация: %s  %.2f", model.Controller.Name, model.Controller.NormalizedTime()), 30, y+23, 17, ui.Muted)
		a.UI.Text(fmt.Sprintf("X %.2f   Высота %.2f   Кадр %s", f.Position.X, f.Position.Y, model.Controller.FrameID), 30, y+46, 15, ui.Muted)
		move := "—"
		if f.Move != nil {
			move = f.Move.Name
		}
		a.UI.Fit("Приём: "+move, 30, y+70, 470, 17, ui.Text)
	}
}
