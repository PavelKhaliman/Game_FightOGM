package ui

import (
	"fightogm/internal/config"
	"fightogm/internal/fighter"
	"fightogm/internal/game"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

var MainLabels = []string{"БОЙ С КОМПЬЮТЕРОМ", "ДВА ИГРОКА", "КАК ИГРАТЬ", "НАСТРОЙКИ", "ВЫХОД"}
var PauseLabels = []string{"ПРОДОЛЖИТЬ", "НАЧАТЬ РАУНД ЗАНОВО", "УПРАВЛЕНИЕ", "ГЛАВНОЕ МЕНЮ"}
var ResultLabels = []string{"РЕВАНШ", "ПОДГОТОВКА К БОЮ", "ГЛАВНОЕ МЕНЮ"}

func (u *UI) Main(selected int) {
	Rect(0, 0, 680, 900, rl.Fade(Background, .94))
	Rect(680, 0, 2, 900, rl.Fade(Amber, .35))
	u.Text("ОТДЕЛ ГЛАВНОГО МЕХАНИКА", 100, 119, 18, Amber)
	u.Heading("FIGHTOGM", 90, 175, 85, Text)
	u.Text("ПРОИЗВОДСТВО НЕ ОСТАНАВЛИВАЕТСЯ", 100, 284, 21, Muted)
	Rect(100, 340, 72, 5, Amber)
	u.Buttons(MainLabels, selected, MainLayout)
	u.Text("2D-ФАЙТИНГ   ·   10 БОЙЦОВ   ·   СМЕНА 01", 100, 760, 18, Muted)
	u.Text("↑ ↓  Выбор     Enter  Подтвердить", 100, 802, 18, Muted)
	u.Right("ЦЕХ ОГМ  /  СМЕНА 01", 1532, 842, 20, Text)
}
func CardRect(i int) rl.Rectangle {
	return rl.NewRectangle(64+float32(i%5)*298, 560+float32(i/5)*120, 280, 108)
}
func ReadyRect(p int) rl.Rectangle  { return rl.NewRectangle(64+float32(p)*758, 805, 714, 36) }
func PlayerRect(p int) rl.Rectangle { return rl.NewRectangle(64+float32(p)*758, 128, 714, 82) }
func (u *UI) Select(roster []fighter.Definition, selected [2]int, ready [2]bool, bot bool, focus int, message string, settings config.Settings, portrait func(int, rl.Rectangle)) {
	Rect(0, 0, 1600, 208, rl.Fade(Background, .88))
	u.Header("ВЫБОР БОЙЦОВ")
	mode := "ДВА ИГРОКА · ОДНА КЛАВИАТУРА"
	if bot {
		mode = "БОЙ С КОМПЬЮТЕРОМ"
	}
	u.Center(mode, 104, 18, Amber)
	u.Center("ПРОТИВ", 320, 44, Text)
	for p := 0; p < 2; p++ {
		x := float32(64 + p*758)
		d := roster[selected[p]]
		c := PlayerColors[p]
		tag := fmt.Sprintf("ИГРОК %d", p+1)
		if p == 1 && bot {
			tag = "КОМПЬЮТЕР"
		}
		u.Heading(d.Name, x+16, 135, 34, Text)
		u.Right(tag, x+698, 144, 18, c)
		u.Fit(d.Passive, x+16, 182, 682, 17, Muted)
		if focus == p {
			Rect(x+16, 174, 60, 2, c)
		}
		r := ReadyRect(p)
		rl.DrawRectangleRec(r, rl.Fade(Panel, .97))
		Rect(x, r.Y, 4, r.Height, c)
		label := keyName(settings.Mappings[p].Light) + " — ГОТОВ К БОЮ"
		if ready[p] {
			label = "ГОТОВ К БОЮ"
		}
		u.Text(label, x+16, r.Y+8, 19, c)
		keys := "W A S D — выбор"
		if p == 1 {
			keys = "Стрелки — выбор соперника"
		}
		u.Right(keys, x+696, r.Y+9, 17, Muted)
	}
	for i, d := range roster {
		r := CardRect(i)
		rl.DrawRectangleRec(r, rl.Fade(Panel, .96))
		rl.DrawRectangleLinesEx(r, 1, Line)
		portrait(i, rl.NewRectangle(r.X+2, r.Y+2, 94, r.Height-4))
		u.Fit(d.Name, r.X+103, r.Y+29, 169, 22, Text)
		u.Fit(d.Archetype, r.X+103, r.Y+65, 164, 16, Muted)
		for p := 0; p < 2; p++ {
			if selected[p] == i {
				o := float32(p) * 3
				rl.DrawRectangleLinesEx(rl.NewRectangle(r.X+o, r.Y+o, r.Width-o*2, r.Height-o*2), 3, PlayerColors[p])
				u.Text(fmt.Sprintf("%d", p+1), r.X+103+float32(p)*26, r.Y+7, 17, PlayerColors[p])
			}
		}
	}
	if message != "" {
		u.Center(message, 525, 18, PlayerColors[0])
	}
	u.Footer("Enter — бой   ·   R / O — отмена готовности   ·   Клик по имени игрока выбирает его для мыши   ·   Esc — меню")
}
func (u *UI) Versus(a, b fighter.Definition) {
	Rect(0, 0, 1600, 160, rl.Fade(Background, .85))
	Rect(0, 680, 1600, 220, rl.Fade(Background, .9))
	u.Center("СМЕНА НАЧИНАЕТСЯ", 57, 30, Amber)
	u.Center("ПРОТИВ", 342, 57, Text)
	u.Heading(a.Name, 130, 715, 52, PlayerColors[0])
	u.Right(b.Name, 1470, 724, 48, PlayerColors[1])
	u.Text(a.Archetype, 133, 783, 23, Muted)
	u.Right(b.Archetype, 1470, 783, 23, Muted)
}
func status(f *fighter.Fighter) string {
	s := ""
	if f.Status.Omni > 0 {
		s += "ОМНИ  "
	}
	if f.Status.Invisible > 0 {
		s += "НЕВИДИМОСТЬ  "
	}
	if f.Status.Sueta > 0 {
		s += fmt.Sprintf("СУЕТА ×%d  ", f.Status.Sueta)
	}
	if f.Status.Gym > 0 {
		s += "КАЧАЛКА  "
	}
	if f.Status.Focus > 0 {
		s += "БЕЗ ВЫХОДНЫХ  "
	}
	if f.Status.Armor > 0 {
		s += "БРОНЯ  "
	}
	if f.Status.Counter > 0 {
		s += "КОНТРАТАКА  "
	}
	if f.Status.Resistance > 0 {
		s += "ПЛОХО СЛЫШУ  "
	}
	if f.Status.AutoBlock > 0 {
		s += "АВТОПИЛОТ  "
	}
	if f.Status.Slippery > 0 {
		s += "СКОЛЬЗКО  "
	}
	return s
}
func (u *UI) HUD(m *game.Match, settings config.Settings, bot bool) {
	Rect(0, 0, 1600, 164, rl.Fade(Background, .86))
	for i, f := range m.Fighters {
		x := float32(64 + i*836)
		c := PlayerColors[i]
		w := float32(636)
		u.Heading(f.Definition.Name, x, 27, 30, Text)
		tag := fmt.Sprintf("ИГРОК %d", i+1)
		if i == 1 && bot {
			tag = "КОМПЬЮТЕР"
		}
		u.Right(tag, x+w, 36, 18, c)
		Rect(x, 73, w, 28, Line)
		// Right player's bar drains toward the outside edge.
		bar := func(amount float32, color rl.Color) {
			width := w * amount / f.Definition.MaxHP
			bx := x
			if i == 1 {
				bx = x + w - width
			}
			Rect(bx, 73, width, 28, color)
		}
		bar(f.DamageHP, rl.NewColor(179, 79, 67, 255))
		bar(f.DisplayHP, c)
		u.Text(fmt.Sprintf("%d / %d", int(f.HP), int(f.Definition.MaxHP)), x, 108, 17, Muted)
		Rect(x+132, 115, 358, 6, Line)
		Rect(x+132, 115, 358*f.Meter/100, 6, Amber)
		u.Right(fmt.Sprintf("%d%%", int(f.Meter)), x+545, 108, 17, Amber)
		for r := 0; r < 2; r++ {
			color := Line
			if r < m.Wins[i] {
				color = Amber
			}
			rl.DrawCircle(int32(x+w-38+float32(r)*25), 117, 7, color)
		}
		u.Fit(status(f), x, 139, w, 17, c)
		if f.Combo.Hits > 1 {
			cx := x
			if i == 1 {
				cx = x + w - 206
			}
			word := "УДАРОВ"
			if f.Combo.Hits%100 < 11 || f.Combo.Hits%100 > 14 {
				if f.Combo.Hits%10 == 1 {
					word = "УДАР"
				} else if f.Combo.Hits%10 >= 2 && f.Combo.Hits%10 <= 4 {
					word = "УДАРА"
				}
			}
			u.Heading(fmt.Sprintf("%d %s", f.Combo.Hits, word), cx, 228, 36, c)
			u.Text(fmt.Sprintf("Урон серии: %d", int(f.Combo.Damage)), cx, 275, 19, Text)
		}
		bx := float32(64 + i*836)
		Rect(bx, 769, w, 61, rl.Fade(Panel, .95))
		special := f.Definition.Moves["special"]
		keys := settings.Mappings[i]
		key := keyName(keys.Special)
		second := keyName(keys.Block) + "+" + key
		ult := keyName(keys.Ultimate)
		cooldown := f.Cooldowns["special"]
		if f.Status.Omni > 0 {
			special = f.Definition.Moves["rush"]
			cooldown = f.Cooldowns["rush"]
		}
		label := "ГОТОВ"
		if cooldown > 0 {
			label = fmt.Sprintf("%.1f с", cooldown)
		}
		u.Fit(fmt.Sprintf("%s  %s  /  %s", key, special.Name, label), bx+14, 780, w-28, 18, c)
		secondary := f.Definition.Moves["secondary"].Name
		if f.Cooldowns["secondary"] > 0 {
			secondary += fmt.Sprintf(" [%.0f с]", f.Cooldowns["secondary"])
		}
		u.Fit(fmt.Sprintf("%s  %s     %s  Суперприём: 100%%", second, secondary, ult), bx+14, 807, w-28, 16, Muted)
	}
	u.Center(fmt.Sprintf("%02d", int(math.Ceil(float64(m.Timer)))), 32, 57, Text)
	u.Center(fmt.Sprintf("РАУНД %d", m.Round), 104, 17, Muted)
	u.Text("ЦЕХ ОГМ  ·  РАБОТАЕМ", 64, 859, 17, Muted)
	u.Right("Esc  Пауза    F3  Отладка    Alt+Enter  Полный экран", 1536, 859, 17, Muted)
	if m.Phase == game.Intro {
		title := fmt.Sprintf("РАУНД %d", m.Round)
		if m.Wins == [2]int{1, 1} {
			title = "РЕШАЮЩИЙ РАУНД"
		}
		if m.PhaseTime > 1.2 {
			title = "НА ПРОИЗВОДСТВО!"
		}
		Rect(0, 354, 1600, 126, rl.Fade(Background, .75))
		u.Center(title, 381, 60, Amber)
	}
	if m.Phase == game.RoundOver {
		title := "НОКАУТ"
		if m.Timer <= 0 {
			title = "ВРЕМЯ ВЫШЛО"
		}
		if m.Winner < 0 {
			title = "НИЧЬЯ"
		}
		u.Center(title, 324, 100, Amber)
		if m.PhaseTime > 1.1 {
			s := "ПОВТОР РАУНДА"
			if m.Winner >= 0 {
				tag := fmt.Sprintf("ИГРОК %d", m.Winner+1)
				if m.Winner == 1 && bot {
					tag = "КОМПЬЮТЕР"
				}
				s = tag + " ЗАБИРАЕТ РАУНД"
			}
			u.Center(s, 446, 26, Text)
		}
	}
}
func (u *UI) Settings(s config.Settings, selected int) {
	Rect(0, 0, 1600, 900, Background)
	u.Header("НАСТРОЙКИ")
	u.Heading("ПОДГОТОВКА К СМЕНЕ", 288, 151, 43, Text)
	fullscreen := "ВЫКЛ."
	if s.Fullscreen {
		fullscreen = "ВКЛ."
	}
	labels := []string{fmt.Sprintf("Разрешение                         %d × %d", s.Width, s.Height), "Полный экран                        " + fullscreen, fmt.Sprintf("Общая громкость                  %d%%", int(s.Master*100+.5)), fmt.Sprintf("Музыка                                    %d%%", int(s.Music*100+.5)), fmt.Sprintf("Эффекты                                 %d%%", int(s.SFX*100+.5)), "СОХРАНИТЬ И ВЕРНУТЬСЯ"}
	u.Buttons(labels, selected, Layout{288, 246, 1024, 63, 15})
	u.Text("← →  Изменить значение     ↑ ↓  Выбрать пункт", 288, 760, 22, Muted)
	u.Footer("Alt+Enter — переключить полный экран. Привязки клавиш можно изменить в config/settings.json.")
}
func keyName(k int32) string {
	switch k {
	case rl.KeyLeft:
		return "←"
	case rl.KeyRight:
		return "→"
	case rl.KeyUp:
		return "↑"
	case rl.KeyDown:
		return "↓"
	case rl.KeySpace:
		return "Пробел"
	}
	if k >= 32 && k < 127 {
		return string(rune(k))
	}
	return fmt.Sprintf("[%d]", k)
}
func (u *UI) Controls(s config.Settings) {
	Rect(0, 0, 1600, 900, Background)
	u.Header("КАК ИГРАТЬ")
	u.Heading("ДО ДВУХ ПОБЕД. БЕЗ ПЕРЕРЫВА.", 64, 133, 42, Text)
	for i, m := range s.Mappings {
		x := float32(64 + i*758)
		Rect(x, 215, 714, 438, Panel)
		u.Heading(fmt.Sprintf("ИГРОК %d", i+1), x+28, 239, 30, PlayerColors[i])
		rows := [][2]string{{keyName(m.Left) + " / " + keyName(m.Right), "Влево / вправо"}, {keyName(m.Jump) + " / " + keyName(m.Crouch), "Прыжок / присед"}, {keyName(m.Light) + " / " + keyName(m.Heavy) + " / " + keyName(m.Kick), "Джеб / сильный удар / нога"}, {keyName(m.Light) + " + " + keyName(m.Heavy), "Захват (вблизи)"}, {keyName(m.Block), "Блок"}, {keyName(m.Special), "Специальный приём"}, {keyName(m.Block) + " + " + keyName(m.Special), "Защитный / второй приём"}, {keyName(m.Ultimate), "Суперприём (100% энергии)"}}
		for r, row := range rows {
			y := float32(296 + r*37)
			u.Text(row[0], x+28, y, 23, Amber)
			u.Text(row[1], x+213, y, 21, Text)
		}
	}
	u.Text("Серии: джеб → сильный удар → нога. Следующий удар можно нажать немного заранее.", 64, 698, 23, Text)
	u.Text("Присед + нога — подсечка. Движение к сопернику + сильный удар — хук. Захват пробивает блок.", 64, 739, 22, Muted)
	u.Text("99 секунд на раунд. При равном здоровье — переигровка. Побеждает первый, кто взял 2 раунда.", 64, 780, 22, Muted)
	u.Footer("Esc / Enter — вернуться      ·      F3 — состояние бойцов, анимации и зоны ударов")
}
