// Package characters defines the ten fighters. Shared combat stays in game.
package characters

import (
	"fightogm/internal/combat"
	"fightogm/internal/fighter"
)

func Roster() []fighter.Definition {
	roster := []fighter.Definition{
		{ID: "ivanov", Name: "ИВАНОВ", Archetype: "Тяжёлый напор", Passive: "Плохой слух: каждый третий удар оглушает слабее", VictoryText: "ПРОИЗВОДСТВО НЕ ОСТАНОВИТЬ", MaxHP: 1100, Speed: .9, Power: 1.15, Defense: 1.1, Color: [3]uint8{239, 84, 63}},
		{ID: "gagloev", Name: "ГАГЛОЕВ", Archetype: "Быстрый техник", Passive: "Быстрые руки: ускоренный возврат после джеба", VictoryText: "ШАМПУР ЗАКРЫТ", MaxHP: 950, Speed: 1.12, Power: 1, Defense: .95, Color: [3]uint8{246, 182, 74}},
		{ID: "kiselik", Name: "КИСЕЛИК", Archetype: "Суета и натиск", Passive: "Не отказываюсь: попадания и блок дают до 5 зарядов суеты", MaxHP: 980, Speed: 1.2, Power: .95, Defense: 1, Color: [3]uint8{122, 207, 132}},
		{ID: "novatskiy", Name: "НОВАЦКИЙ", Archetype: "Техническое давление", Passive: "Овертайм: серии атак дают больше энергии", MaxHP: 1020, Speed: 1, Power: 1.05, Defense: 1, Color: [3]uint8{244, 157, 79}},
		{ID: "zhirnov", Name: "ЖИРНОВ", Archetype: "Автоматизация боя", Passive: "ИИ-помощник: каждый третий удар вызывает дрона", MaxHP: 960, Speed: .96, Power: 1, Defense: 1, Color: [3]uint8{103, 217, 238}},
		{ID: "elkhimov", Name: "ЕЛХИМОВ", Archetype: "Мощь и захваты", Passive: "Механик: на 20% меньше урона от тяжёлых ударов", MaxHP: 1200, Speed: .82, Power: 1.3, Defense: 1.15, Color: [3]uint8{232, 181, 93}},
		{ID: "kalachev", Name: "КАЛАЧЕВ", Archetype: "Уклонение и обман", Passive: "Отлынить: успешное уклонение даёт энергию", MaxHP: 940, Speed: 1.1, Power: 1, Defense: .97, Color: [3]uint8{71, 180, 249}},
		{ID: "fedoseev", Name: "ФЕДОСЕЕВ", Archetype: "Контратаки и сарказм", Passive: "Сарказм: контратака отнимает энергию противника", MaxHP: 970, Speed: 1.02, Power: 1, Defense: 1, Color: [3]uint8{185, 145, 237}},
		{ID: "khaliman", Name: "ХАЛИМАН", Archetype: "Стойки и усиления", Passive: "Ожидание заряжает энергию; джеб → сильный удар усиливается", MaxHP: 1050, Speed: .96, Power: 1.1, Defense: 1, Color: [3]uint8{81, 217, 171}},
		{ID: "shuev", Name: "ШУЕВ", Archetype: "Дисциплина карате", Passive: "Бесконечная смена: ускоренное восстановление", MaxHP: 1030, Speed: 1.08, Power: 1.08, Defense: 1, Color: [3]uint8{218, 219, 231}},
	}
	for i := range roster {
		d := &roster[i]
		d.ModelPath = "assets/models/" + d.ID + ".glb"
		d.Moves = fighter.CommonMoves()
		if d.VictoryText == "" {
			d.VictoryText = d.Name + " ПОБЕЖДАЕТ"
		}
		if d.ID == "gagloev" {
			m := d.Moves["light"]
			m.Recovery *= .7
			d.Moves["light"] = m
		}
		if d.ID == "shuev" {
			for k, m := range d.Moves {
				m.Recovery *= .85
				d.Moves[k] = m
			}
		}
		configure(d)
	}
	return roster
}
func attack(id, name, anim, effect string, damage, reach, cooldown float32) combat.MoveDefinition {
	return combat.MoveDefinition{ID: id, Name: name, Animation: anim, Effect: effect, Startup: .28, Active: .19, Recovery: .38, Damage: damage, Chip: .12, HitStun: .55, BlockStun: .24, KnockbackX: .75, MeterGain: 8, Cooldown: cooldown, Hits: 1, Hitboxes: []combat.HitboxDefinition{{Forward: reach, Height: 1.1, Radius: .46}}, Events: []combat.MoveEvent{{At: 0, Type: combat.EventProp, Value: effect}, {At: .28, Type: combat.EventVFX, Value: effect}}}
}
func defense(name, anim, buff string, cooldown float32) combat.MoveDefinition {
	return combat.MoveDefinition{ID: "secondary", Name: name, Animation: anim, Effect: buff, Startup: .06, Active: .18, Recovery: .22, Cooldown: cooldown, Defense: true, Events: []combat.MoveEvent{{At: 0, Type: combat.EventBuff, Value: buff}, {At: 0, Type: combat.EventVFX, Value: buff}}}
}
func ultimate(name, anim, effect string, damage, reach float32) combat.MoveDefinition {
	m := attack("ultimate", name, anim, effect, damage, reach, 0)
	m.Cost = 100
	m.Ultimate = true
	m.Heavy = true
	m.Startup = .52
	m.Active = .8
	m.Recovery = .65
	m.Knockdown = true
	m.KnockbackX = 1.8
	m.KnockbackY = 2
	m.Events[1].At = .52
	return m
}
func projectile(m combat.MoveDefinition, kind string, count int) combat.MoveDefinition {
	m.Projectile = true
	m.Events = append(m.Events, combat.MoveEvent{At: m.Startup, Type: combat.EventProjectile, Value: kind, Count: count})
	return m
}
func configure(d *fighter.Definition) {
	var s, b, u combat.MoveDefinition
	switch d.ID {
	case "ivanov":
		s = defense("ПЛОХО СЛЫШУ · А?", "bad_hearing", "hearing", 10)
		s.ID = "special"
		b = defense("Я НА ПРОИЗВОДСТВО", "production_invisible", "invisible", 15)
		u = ultimate("ОМНИ ИВАНОВ", "omni_transform", "omni", 0, 0)
		u.Active = .5
		u.Events = append(u.Events, combat.MoveEvent{At: .2, Type: combat.EventBuff, Value: "omni"})
		rush := attack("rush", "ОМНИ-РЫВОК", "omni_rush", "omni", 125, .9, 1.2)
		rush.Dash = 7
		rush.Knockdown = true
		rush.KnockbackX = 2
		rush.Heavy = true
		d.Moves["rush"] = rush
	case "gagloev":
		s = attack("special", "ШАМПУР КОНТРОЛЬ", "skewer_control", "skewer", 70, 1.45, 6)
		b = defense("БЫСТРЫЕ РУКИ", "block", "quick", 7)
		u = ultimate("ШАШЛЫК ГОТОВ", "shashlik", "skewer", 220, .62)
		u.Grab = true
		u.Active = .16
		u.Recovery = 1.2
	case "kiselik":
		s = attack("special", "СУЕТА", "sueta_dash", "afterimage", 32, .7, 4)
		s.Startup = .06
		s.Active = .2
		s.Recovery = .1
		s.Dash = 8
		b = defense("ДА, СДЕЛАЮ!", "yes_boss", "quick", 8)
		u = ultimate("ВСЁ БЕРУ", "everything_at_once", "afterimage", 52, 1.02)
		u.Hits = 5
		u.Dash = 3.8
	case "novatskiy":
		s = projectile(attack("special", "ШНЕК", "screw_throw", "auger", 72, 1, 6), "auger", 1)
		b = defense("Я ЗАНЯТ.", "snap_back", "counter", 8)
		b.CounterWindow = .7
		u = projectile(ultimate("ПАРТИЯ ШНЕКОВ", "screw_batch", "auger", 62, 1), "auger", 5)
	case "zhirnov":
		s = projectile(attack("special", "ИИ-АГЕНТ", "ai_agent", "bot", 65, 1, 7), "bot", 1)
		b = defense("АВТОПИЛОТ", "autopilot", "autoblock", 9)
		u = projectile(ultimate("СДЕЛАЙ ЗА МЕНЯ", "do_it_for_me", "bot", 52, 1), "bot", 5)
	case "elkhimov":
		s = attack("special", "КЛЮЧ НА 32", "wrench_32", "wrench", 110, 1.2, 7)
		s.Heavy = true
		s.KnockbackX = 2
		s.Knockdown = true
		b = defense("ПОЧИНИЛ", "repair_armor", "armor", 10)
		u = ultimate("КАПРЕМОНТ", "overhaul", "wrench", 250, .65)
		u.Grab = true
		u.Active = .16
		u.Recovery = 1.2
	case "kalachev":
		s = projectile(attack("special", "ВОДОПОДГОТОВКА", "water_burst", "water", 60, 1, 6), "water", 1)
		b = defense("МЕНЯ НЕ БЫЛО", "was_not_here", "evade", 8)
		u = ultimate("ПЕРЕРЫВ", "break_return", "water", 215, .95)
		u.Events = append(u.Events, combat.MoveEvent{At: 0, Type: combat.EventBuff, Value: "vanish"}, combat.MoveEvent{At: .42, Type: combat.EventTeleport, Value: "behind"})
	case "fedoseev":
		s = projectile(attack("special", "НУ ДА, КОНЕЧНО", "sarcasm_wave", "wave", 58, 1, 6), "wave", 1)
		b = defense("МЕДЛЕННЫЕ АПЛОДИСМЕНТЫ", "slow_clap", "counter", 8)
		b.CounterWindow = .9
		u = ultimate("ГЕНИАЛЬНЫЙ ПЛАН", "genius_plan", "wave", 58, 1.4)
		u.Hits = 4
		u.Dash = 2.5
	case "khaliman":
		s = projectile(attack("special", "ИМПУЛЬС ПЛК", "plc_pulse", "electric", 62, 1, 6), "electric", 1)
		s.HitStun = .85
		b = defense("РЕЖИМ КАЧАЛКИ", "gym_mode", "gym", 12)
		u = projectile(ultimate("АСУ ТП", "asutp", "panel", 65, 1), "electric", 4)
	case "shuev":
		s = attack("special", "СЕРИЯ КАРАТЕ", "karate_flurry", "trail", 27, .97, 6)
		s.Hits = 4
		s.Active = .64
		s.Dash = 1.8
		s.KnockbackX = .16
		b = defense("СТОЙКА", "stance_counter", "counter", 7)
		b.CounterWindow = .8
		u = ultimate("БЕЗ ВЫХОДНЫХ", "endless_shift", "focus", 0, 0)
		u.Events = append(u.Events, combat.MoveEvent{At: 0, Type: combat.EventBuff, Value: "focus"})
		finish := attack("finisher", "КОНЕЦ СМЕНЫ", "karate_flurry", "trail", 170, 1.2, 0)
		finish.Heavy = true
		finish.Knockdown = true
		finish.Dash = 5
		finish.KnockbackX = 2
		d.Moves["finisher"] = finish
	}
	d.Moves["special"] = s
	d.Moves["secondary"] = b
	d.Moves["ultimate"] = u
}
