package game

import (
	"fightogm/internal/combat"
	"fightogm/internal/input"
	"math"
	"testing"
)

func TestApproachStopsWithoutMovingOpponent(t *testing.T) {
	for player := 0; player < 2; player++ {
		for _, side := range []float32{-1, 1} {
			for _, anchor := range []float32{0, -side * 3.6} {
				m := setup(0, 0)
				walker, still := m.Fighters[player], m.Fighters[1-player]
				still.Position = combat.Vec3{X: anchor}
				walker.Position = combat.Vec3{X: anchor + side*1.4}
				var states [2]input.InputState
				states[player].X = -side
				m.Submit(states)
				advance(m, 1.5)
				if !near(still.Position.X, anchor) || !near(walker.Position.X, anchor+side*PushboxWidth) {
					t.Fatalf("player %d side %v: walking moved opponent or crossed body: %v / %v", player, side, walker.Position, still.Position)
				}
				before := walker.Position.X
				m.Submit([2]input.InputState{})
				advance(m, .4)
				if !near(walker.Position.X, before) || !near(still.Position.X, anchor) {
					t.Fatal("fighters drifted after releasing movement")
				}
				states[player].X = side
				m.Submit(states)
				advance(m, .2)
				if (walker.Position.X-before)*side <= 0 || !near(still.Position.X, anchor) {
					t.Fatal("retreat was blocked or dragged the opponent")
				}
			}
		}
	}
}

func TestGroundDashCannotTunnelThroughOrCarryOpponent(t *testing.T) {
	m := setup(2, 0)
	a, b := m.Fighters[0], m.Fighters[1]
	a.Position.X, b.Position.X = -1, 0
	b.Status.Invulnerable = 3
	press(m, 0, input.Special)
	advance(m, .5)
	if !near(b.Position.X, 0) || a.Position.X > -PushboxWidth+.001 {
		t.Fatalf("dash crossed or carried opponent: %v / %v", a.Position, b.Position)
	}
}

func TestTouchingIdleFightersDoNotAttract(t *testing.T) {
	m := setup(0, 0)
	m.Fighters[0].Position.X, m.Fighters[1].Position.X = -.43, .43
	advance(m, 2)
	if math.Abs(float64(m.Fighters[0].Position.X+.43)) > .0001 || math.Abs(float64(m.Fighters[1].Position.X-.43)) > .0001 {
		t.Fatal("neutral close-range fighters moved without input")
	}
}
