package game

import (
	"fightogm/internal/characters"
	"fightogm/internal/combat"
	"fightogm/internal/fighter"
	"fightogm/internal/input"
	"math"
	"testing"
)

func setup(a, b int) *Match {
	r := characters.Roster()
	m := New(r[a], r[b])
	m.Phase = Fighting
	m.Fighters[0].Position = combat.Vec3{X: -.48}
	m.Fighters[1].Position = combat.Vec3{X: .48}
	return m
}
func advance(m *Match, seconds float32) {
	for i := 0; i < int(seconds/Step)+1; i++ {
		m.Tick(Step)
	}
}
func press(m *Match, player int, action input.Action) {
	var states [2]input.InputState
	states[player].Pressed = []input.Action{action}
	m.Submit(states)
	m.Tick(Step)
}
func near(a, b float32) bool { return math.Abs(float64(a-b)) < .01 }

func TestNormalsDamageOnceAndCooldown(t *testing.T) {
	for _, action := range []input.Action{input.Light, input.Heavy, input.Kick, input.Grab} {
		t.Run(string(rune('0'+action)), func(t *testing.T) {
			m := setup(0, 1)
			press(m, 0, action)
			advance(m, 1.5)
			a, b := m.Fighters[0], m.Fighters[1]
			if b.HP >= b.Definition.MaxHP {
				t.Fatal("attack missed at close range")
			}
			if a.HitCount != 1 {
				t.Fatalf("single move hit %d times", a.HitCount)
			}
			if a.Meter <= 0 || b.Meter <= 0 {
				t.Fatal("meter did not increase")
			}
		})
	}
	m := setup(1, 0)
	press(m, 0, input.Special)
	serial := m.Serial
	advance(m, 1)
	press(m, 0, input.Special)
	if m.Serial != serial {
		t.Fatal("special ignored cooldown")
	}
	advance(m, 6)
	press(m, 0, input.Special)
	if m.Serial == serial {
		t.Fatal("special did not recover")
	}
}
func TestBlockingChipGrabAndRecovery(t *testing.T) {
	for _, action := range []input.Action{input.Light, input.Heavy, input.Grab} {
		m := setup(1, 0)
		var states [2]input.InputState
		states[0].Pressed = []input.Action{action}
		states[1].Block = true
		m.Submit(states)
		advance(m, 1)
		loss := m.Fighters[1].Definition.MaxHP - m.Fighters[1].HP
		switch action {
		case input.Light:
			if loss != 0 {
				t.Fatalf("light should be fully blocked, got %f", loss)
			}
		case input.Heavy:
			if loss <= 0 || loss > 20 {
				t.Fatalf("expected small chip damage, got %f", loss)
			}
		case input.Grab:
			if loss < 60 || m.Fighters[1].State != fighter.Knockdown {
				t.Fatal("grab must break block and knock down")
			}
		}
		if action == input.Grab {
			advance(m, 2)
			if m.Fighters[1].State != fighter.Blocking {
				t.Fatalf("get-up failed: %v", m.Fighters[1].State)
			}
		}
	}
}
func TestBufferNearRecoveryAndExpiry(t *testing.T) {
	m := setup(1, 0)
	m.Fighters[1].Position.X = 5
	press(m, 0, input.Heavy)
	advance(m, .62)
	press(m, 0, input.Light)
	advance(m, .2)
	if m.Fighters[0].Move == nil || m.Fighters[0].Move.ID != "light" {
		t.Fatal("near-recovery input was lost")
	}
	m = setup(1, 0)
	m.Fighters[1].Position.X = 5
	press(m, 0, input.Heavy)
	advance(m, .1)
	press(m, 0, input.Light)
	advance(m, 1)
	if m.Serial != 1 {
		t.Fatal("stale queued input should expire")
	}
}
func TestComboCancelAndTimeout(t *testing.T) {
	m := setup(1, 0)
	press(m, 0, input.Light)
	advance(m, .21)
	press(m, 0, input.Heavy)
	advance(m, .55)
	if m.Fighters[0].Combo.Hits < 2 {
		t.Fatalf("combo did not connect: %#v", m.Fighters[0].Combo)
	}
	advance(m, 2)
	if m.Fighters[0].Combo.Hits != 0 {
		t.Fatal("combo did not expire")
	}
}
func TestTimerTieKOAndBestOfThree(t *testing.T) {
	m := setup(0, 1)
	m.Timer = Step
	m.Fighters[0].HP = 400
	m.Fighters[1].HP = 300
	m.Tick(Step)
	if m.Winner != 0 || m.Wins[0] != 1 || m.Phase != RoundOver {
		t.Fatal("time win failed")
	}
	advance(m, 4)
	if m.Round != 2 || m.Phase != Intro {
		t.Fatal("round two did not begin")
	}
	advance(m, 2.2)
	m.Fighters[1].HP = 0
	m.Tick(Step)
	advance(m, 4)
	if m.Phase != MatchOver || m.Wins[0] != 2 {
		t.Fatal("best of three failed")
	}
	if m.Fighters[1].DisplayHP != 0 {
		t.Fatal("KO health bar did not empty during the round result")
	}
	tie := setup(0, 1)
	tie.Fighters[0].HP = 50
	tie.Fighters[1].HP = 50
	tie.Timer = 0
	tie.Tick(Step)
	if tie.Winner != -1 || tie.Wins != [2]int{} {
		t.Fatal("tie awarded a win")
	}
}
func TestAllSpecialsAndUltimateGameplay(t *testing.T) {
	roster := characters.Roster()
	for i, d := range roster {
		t.Run(d.ID, func(t *testing.T) {
			m := setup(i, (i+1)%len(roster))
			a, b := m.Fighters[0], m.Fighters[1]
			press(m, 0, input.Ultimate)
			if a.Move != nil {
				t.Fatal("ultimate allowed without meter")
			}
			a.Meter = 100
			press(m, 0, input.Ultimate)
			if a.Move == nil || !a.Move.Ultimate || a.Meter != 0 {
				t.Fatal("ultimate activation/cost failed")
			}
			advance(m, 3.5)
			switch d.ID {
			case "ivanov":
				if a.Status.Omni <= 0 {
					t.Fatal("omni mode missing")
				}
			case "shuev":
				if a.Status.Focus <= 0 {
					t.Fatal("focus missing")
				}
			default:
				if b.HP == b.Definition.MaxHP {
					t.Fatal("ultimate produced no damage")
				}
			}
			m = setup(i, (i+1)%len(roster))
			a, b = m.Fighters[0], m.Fighters[1]
			press(m, 0, input.Special)
			advance(m, 1.8)
			if d.ID == "ivanov" {
				if a.Status.Resistance <= 0 {
					t.Fatal("hearing resistance missing")
				}
			} else if b.HP == b.Definition.MaxHP {
				t.Fatal("special produced no gameplay effect")
			}
			m = setup(i, (i+1)%len(roster))
			a = m.Fighters[0]
			press(m, 0, input.Secondary)
			s := a.Status
			if s.Invisible+s.QuickRecovery+s.Armor+s.Counter+s.AutoBlock+s.Invulnerable+s.Gym+s.Capacitor <= 0 {
				t.Fatal("secondary produced no gameplay effect")
			}
		})
	}
}
func TestCounterAutoblockInvisibilityArmorAndEvasion(t *testing.T) {
	for _, i := range []int{0, 3, 4, 5, 6, 7, 9} {
		m := setup(1, i)
		a, d := m.Fighters[0], m.Fighters[1]
		press(m, 1, input.Secondary)
		advance(m, .08)
		// A real incoming hit must trigger the defensive state, even during its animation.
		move := a.Definition.Moves["heavy"]
		before := d.HP
		m.strike(0, move, 0)
		switch i {
		case 0:
			if d.Status.Invisible > 0 {
				t.Fatal("getting hit must end invisibility")
			}
		case 3, 7, 9:
			if a.HP == a.Definition.MaxHP || d.HP != before {
				t.Fatalf("counter failed for %s", d.Definition.ID)
			}
		case 4, 6:
			if d.HP < before-10 {
				t.Fatalf("autoblock/evade failed for %s", d.Definition.ID)
			}
		case 5:
			if d.Move == nil {
				t.Fatal("armor was interrupted")
			}
		}
	}
	m := setup(0, 1)
	press(m, 0, input.Secondary)
	advance(m, .6)
	press(m, 0, input.Light)
	if m.Fighters[0].Status.Invisible > 0 {
		t.Fatal("attacking must reveal Ivanov")
	}
}
func Test2DMovementJumpAndArenaBounds(t *testing.T) {
	m := setup(0, 1)
	var in [2]input.InputState
	in[0] = input.InputState{X: -1, Z: 1, Jump: true}
	m.Submit(in)
	advance(m, .25)
	if m.Fighters[0].Position.Y <= 0 || m.Fighters[0].Position.Z != 0 {
		t.Fatal("jump must stay in the side-view plane")
	}
	in[0].Jump = false
	m.Submit(in)
	advance(m, 6)
	p := m.Fighters[0].Position
	if p.Y != 0 || p.X < -3.6 || p.Z != 0 {
		t.Fatalf("invalid bounded movement: %#v", p)
	}
}

func Test2DWallPushboxesAndFacingAfterCrossing(t *testing.T) {
	m := setup(0, 0)
	a, b := m.Fighters[0], m.Fighters[1]
	a.Position = combat.Vec3{X: 3.55, Z: 1}
	b.Position = combat.Vec3{X: 3.6, Z: -1}
	a.Velocity.Z = 2
	m.Tick(Step)
	if b.Position.X-a.Position.X < .679 || b.Position.X > 3.6 || a.Position.Z != 0 || a.Velocity.Z != 0 {
		t.Fatalf("fighters overlap at wall or left plane: %v / %v", a.Position, b.Position)
	}
	a.Position = combat.Vec3{X: -.5}
	b.Position = combat.Vec3{X: .5}
	in := [2]input.InputState{{X: 1, Jump: true}, {}}
	m.Submit(in)
	m.Tick(Step)
	in[0].Jump = false
	m.Submit(in)
	advance(m, .72)
	if a.Position.X <= b.Position.X || a.Facing.X >= 0 || b.Facing.X <= 0 {
		t.Fatalf("jump-over did not switch sides and facing: %v / %v", a.Position, b.Position)
	}
}

func TestCPUApproachesAttacksAndWaitsForRound(t *testing.T) {
	r := characters.Roster()[0]
	m := New(r, r)
	cpu := CPU{}
	if s := cpu.Input(m, Step); s.X != 0 || len(s.Pressed) > 0 {
		t.Fatal("CPU acts before the round")
	}
	m.Phase = Fighting
	for i := 0; i < 1200; i++ {
		m.Submit([2]input.InputState{{}, cpu.Input(m, Step)})
		m.Tick(Step)
	}
	if m.Fighters[0].HP >= r.MaxHP || m.Fighters[1].HitCount == 0 {
		t.Fatal("CPU never approached and hit the idle player")
	}
	if m.Fighters[1].Position.Z != 0 || m.Fighters[1].Meter < 0 {
		t.Fatal("CPU bypassed movement or meter rules")
	}
}

func TestEveryCPUCharacterFightsIn2D(t *testing.T) {
	for index, d := range characters.Roster() {
		t.Run(d.ID, func(t *testing.T) {
			m := setup(0, index)
			m.Fighters[0].Position.X = -1.35
			m.Fighters[1].Position.X = 1.35
			cpu := CPU{}
			for tick := 0; tick < 1440; tick++ {
				m.Submit([2]input.InputState{{}, cpu.Input(m, Step)})
				m.Tick(Step)
				for _, f := range m.Fighters {
					if f.Position.Z != 0 || f.Velocity.Z != 0 {
						t.Fatalf("%s left the fighting plane", f.Definition.ID)
					}
				}
			}
			if m.Fighters[0].HP >= m.Fighters[0].Definition.MaxHP {
				t.Fatal("CPU never damaged the player")
			}
		})
	}
}
func TestPassives(t *testing.T) {
	m := setup(0, 2)
	for i := 0; i < 8; i++ {
		m.Fighters[1].Stun = 0
		m.strike(0, m.Fighters[0].Definition.Moves["light"], 0)
	}
	if m.Fighters[1].Status.Sueta != 5 {
		t.Fatal("sueta must cap at five")
	}
	m = setup(8, 0)
	advance(m, 3)
	if m.Fighters[0].Meter < 5 || m.Fighters[0].Animation != "standby" {
		t.Fatal("standby regeneration missing")
	}
	m = setup(4, 0)
	for i := 0; i < 3; i++ {
		m.strike(0, m.Fighters[0].Definition.Moves["light"], 0)
	}
	if len(m.Projectiles) == 0 {
		t.Fatal("AI followup missing")
	}
	m = setup(9, 0)
	m.Fighters[0].Meter = 100
	press(m, 0, input.Ultimate)
	advance(m, 8.3)
	if m.Fighters[0].Move == nil || m.Fighters[0].Move.ID != "finisher" {
		t.Fatal("endless shift should finish with a strike")
	}
}
func TestMirrorIsolationAndIndependentCommands(t *testing.T) {
	m := setup(0, 0)
	m.Fighters[0].Meter = 100
	press(m, 0, input.Ultimate)
	if m.Fighters[1].Meter != 0 || m.Fighters[1].Move != nil || !near(m.Fighters[1].HP, 1100) {
		t.Fatal("mirror fighter state is shared")
	}
}

func TestFocusFinisherWaitsForRecovery(t *testing.T) {
	m := setup(9, 0)
	a := m.Fighters[0]
	a.Status.Focus = .05
	m.startMove(a, a.Definition.Moves["heavy"])
	advance(m, .9)
	if a.Move == nil || a.Move.ID != "finisher" {
		t.Fatal("finisher was lost while another move was recovering")
	}
}
