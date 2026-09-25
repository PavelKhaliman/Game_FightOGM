package game

import (
	"fightogm/internal/characters"
	"fightogm/internal/combat"
	"fightogm/internal/fighter"
	"fightogm/internal/input"
	"testing"
)

func definitionByID(t *testing.T, id string) fighter.Definition {
	t.Helper()
	for _, d := range characters.Roster() {
		if d.ID == id {
			return d
		}
	}
	t.Fatalf("missing fighter %s", id)
	return fighter.Definition{}
}

func TestTsvetkovRoadArmorAffectsChipOnly(t *testing.T) {
	tsvetkov := definitionByID(t, "tsvetkov")
	baseline := tsvetkov
	baseline.ID = "without_passive"
	attacker := definitionByID(t, "ivanov")
	for _, moveID := range []string{"heavy", "grab"} {
		loss := [2]float32{}
		for i, defender := range []fighter.Definition{tsvetkov, baseline} {
			m := New(attacker, defender)
			m.Fighters[1].Input.Block = true
			m.strike(0, attacker.Moves[moveID], 0)
			loss[i] = defender.MaxHP - m.Fighters[1].HP
		}
		if loss[1] <= 0 {
			t.Fatal("attack must connect")
		}
		if moveID == "heavy" && !near(loss[0], loss[1]*.8) {
			t.Fatalf("road armor did not reduce chip: %v", loss)
		}
		if moveID == "grab" && !near(loss[0], loss[1]) {
			t.Fatalf("road armor must not protect from grabs: %v", loss)
		}
	}
}

func TestKozerodRechargeConsumesOnceAndExpires(t *testing.T) {
	d := definitionByID(t, "kozerod")
	m := New(d, definitionByID(t, "ivanov"))
	m.Phase = Fighting
	f := m.Fighters[0]
	press(m, 0, input.Secondary)
	if !near(f.Meter, 12) || f.Status.Capacitor <= 0 {
		t.Fatal("recharge must grant 12 meter and a stored charge")
	}
	advance(m, .6)
	m.startMove(f, d.Moves["light"])
	if f.Status.Capacitor <= 0 {
		t.Fatal("a normal attack consumed projectile charge")
	}
	m.startMove(f, d.Moves["special"])
	if f.Status.Capacitor != 0 || !near(f.Move.Damage, d.Moves["special"].Damage*1.35) {
		t.Fatal("charge must empower the next projectile exactly once")
	}
	f.MoveTime = f.Move.Startup
	m.moveEvents(f)
	if len(m.Projectiles) != 1 || !near(m.Projectiles[0].Move.Damage, d.Moves["special"].Damage*1.35) {
		t.Fatal("spawned projectile lost the stored charge")
	}
	m.startMove(f, d.Moves["special"])
	if !near(f.Move.Damage, d.Moves["special"].Damage) {
		t.Fatal("charge remained active after consumption")
	}
	if !near(d.Moves["special"].Damage, 68) {
		t.Fatal("per-fighter boost changed shared move definitions")
	}
	f.Status.Capacitor = 6
	f.Status.Tick(6.1)
	if f.Status.Capacitor != 0 {
		t.Fatal("unused charge did not expire")
	}
	f.Status.Capacitor = 6
	m.ResetRound()
	if f.Status.Capacitor != 0 {
		t.Fatal("charge leaked into the next round")
	}
}

func TestKozerodAmbitionRequiresThreeUnblockedHits(t *testing.T) {
	d := definitionByID(t, "kozerod")
	m := New(d, definitionByID(t, "ivanov"))
	f, enemy := m.Fighters[0], m.Fighters[1]
	move := d.Moves["light"]
	move.MeterGain = 0
	enemy.Input.Block = true
	m.strike(0, move, 0)
	if f.HitCount != 0 || f.Meter != 0 {
		t.Fatal("blocked attack must not count toward ambition")
	}
	enemy.Input.Block = false
	for i := 1; i <= 3; i++ {
		m.strike(0, move, 0)
		want := float32(0)
		if i == 3 {
			want = 6
		}
		if !near(f.Meter, want) {
			t.Fatalf("hit %d: meter %v, want %v", i, f.Meter, want)
		}
	}
}

func TestNewFighterProjectilesHitFromEitherSide(t *testing.T) {
	for _, id := range []string{"tsvetkov", "kozerod"} {
		for _, side := range []float32{-1, 1} {
			d := definitionByID(t, id)
			m := New(d, definitionByID(t, "ivanov"))
			m.Phase = Fighting
			m.Fighters[0].Position = combat.Vec3{X: side * 1.4}
			m.Fighters[1].Position = combat.Vec3{X: -side * 1.4}
			press(m, 0, input.Special)
			advance(m, 2)
			if m.Fighters[1].HP >= m.Fighters[1].Definition.MaxHP {
				t.Fatalf("%s missed from side %v", id, side)
			}
		}
	}
}
