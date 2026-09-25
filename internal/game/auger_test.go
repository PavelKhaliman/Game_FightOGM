package game

import (
	"fightogm/internal/combat"
	"fightogm/internal/fighter"
	"fightogm/internal/input"
	"testing"
)

func TestNovatskiyHeldAugerTimingRangeAndBlock(t *testing.T) {
	for _, side := range []float32{-1, 1} {
		for _, scenario := range []struct {
			name     string
			distance float32
			block    bool
		}{{"close", .9, false}, {"tip", 1.7, false}, {"miss", 2.3, false}, {"block", 1.7, true}} {
			t.Run(scenario.name, func(t *testing.T) {
				m := setup(3, 0)
				a, b := m.Fighters[0], m.Fighters[1]
				a.Position = combat.Vec3{}
				b.Position = combat.Vec3{X: side * scenario.distance}
				states := [2]input.InputState{{Pressed: []input.Action{input.Special}}, {Block: scenario.block}}
				m.Submit(states)
				m.Tick(Step)
				m.Submit([2]input.InputState{{}, {Block: scenario.block}})
				advance(m, .3)
				if b.HP != b.Definition.MaxHP || len(m.Projectiles) != 0 {
					t.Fatal("windup caused damage or spawned a thrown auger")
				}
				for tick := 0; tick < 160; tick++ {
					m.Tick(Step)
					if len(m.Projectiles) != 0 {
						t.Fatal("held auger became a projectile")
					}
				}
				if scenario.name == "miss" {
					if b.HP != b.Definition.MaxHP {
						t.Fatal("auger hit outside its reach")
					}
				} else {
					damage := a.Definition.Moves["special"].Damage * a.DamageMultiplier() / b.Definition.Defense
					if scenario.block {
						damage *= a.Definition.Moves["special"].Chip
					}
					if !near(b.Definition.MaxHP-b.HP, damage) {
						t.Fatalf("expected one contact for %v damage, got %v", damage, b.Definition.MaxHP-b.HP)
					}
					if b.Position.X*side <= scenario.distance {
						t.Fatal("contact did not push defender away")
					}
				}
				if !near(a.Position.X, 0) {
					t.Fatal("held-weapon strike pulled the attacker toward opponent")
				}
			})
		}
	}
}

func TestNovatskiyAugerComboHasThreeContactsAndFinalKnockdown(t *testing.T) {
	m := setup(3, 0)
	a, b := m.Fighters[0], m.Fighters[1]
	a.Meter = 100
	press(m, 0, input.Ultimate)
	for tick := 0; tick < 200 && b.State != fighter.Knockdown; tick++ {
		m.Tick(Step)
		if len(m.Projectiles) != 0 {
			t.Fatal("combo released projectiles")
		}
	}
	if a.HitCount != 3 || b.State != fighter.Knockdown || a.Meter >= 100 {
		t.Fatalf("combo: hits=%d state=%v meter=%v", a.HitCount, b.State, a.Meter)
	}
}
