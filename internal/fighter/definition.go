package fighter

import "fightogm/internal/combat"

type Definition struct {
	ID, Name, Archetype, Passive, VictoryText string
	ModelPath                                 string
	MaxHP, Speed, Power, Defense              float32
	Color                                     [3]uint8
	Moves                                     map[string]combat.MoveDefinition
}

func CommonMoves() map[string]combat.MoveDefinition {
	mk := func(id, name, anim string, start, active, recovery, damage, reach, height float32) combat.MoveDefinition {
		return combat.MoveDefinition{ID: id, Name: name, Animation: anim, Startup: start, Active: active, Recovery: recovery, Damage: damage, HitStun: .31, BlockStun: .17, KnockbackX: .18, MeterGain: 5, Hits: 1, Hitboxes: []combat.HitboxDefinition{{Forward: reach, Height: height, Radius: .32}}}
	}
	light := mk("light", "Прямой удар", "punch_light", .105, .095, .22, 38, .73, 1.4)
	heavy := mk("heavy", "Сильный удар", "punch_heavy", .25, .13, .34, 82, .91, 1.35)
	heavy.Heavy = true
	heavy.Chip = .08
	heavy.HitStun = .53
	heavy.KnockbackX = .7
	kick := mk("kick", "Высокий удар ногой", "kick_high", .23, .16, .3, 64, 1.06, 1.45)
	kick.HitStun = .43
	kick.KnockbackX = .45
	low := mk("low", "Подсечка", "kick_high", .2, .12, .34, 51, .95, .38)
	low.Knockdown = true
	low.KnockbackY = .4
	hook := mk("hook", "Производственный хук", "punch_heavy", .29, .13, .38, 94, 1.05, 1.3)
	hook.Heavy = true
	hook.Knockdown = true
	hook.KnockbackX = .9
	hook.KnockbackY = 1.6
	grab := mk("grab", "Захват", "grab", .18, .12, .5, 95, .48, 1.05)
	grab.Grab = true
	grab.Knockdown = true
	grab.KnockbackX = 1.1
	grab.KnockbackY = 1.7
	return map[string]combat.MoveDefinition{"light": light, "heavy": heavy, "kick": kick, "low": low, "hook": hook, "grab": grab}
}
