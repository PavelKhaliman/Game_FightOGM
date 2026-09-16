// Package game runs a fixed-step simulation independent of the window and renderer.
package game

import (
	"fightogm/internal/combat"
	"fightogm/internal/fighter"
	"fightogm/internal/input"
	"math"
)

const Step float32 = 1.0 / 120

type Phase uint8

const (
	Intro Phase = iota
	Fighting
	RoundOver
	MatchOver
)

type Event struct {
	Kind, Text string
	Position   combat.Vec3
	Player     int
	Strength   float32
}
type Projectile struct {
	Position, Direction combat.Vec3
	Owner               int
	Move                combat.MoveDefinition
	Kind                string
	Life, Delay, Speed  float32
	Hit                 bool
}
type Match struct {
	Fighters                        [2]*fighter.Fighter
	Wins                            [2]int
	Round                           int
	Timer, Time, PhaseTime, HitStop float32
	Phase                           Phase
	Winner                          int
	Events                          []Event
	Projectiles                     []Projectile
	Serial                          uint64
}

func New(a, b fighter.Definition) *Match {
	m := &Match{Fighters: [2]*fighter.Fighter{fighter.New(a, 0), fighter.New(b, 1)}, Events: make([]Event, 0, 64), Projectiles: make([]Projectile, 0, 32)}
	m.Round = 1
	m.ResetRound()
	return m
}
func (m *Match) Emit(kind, text string, p combat.Vec3, player int, strength float32) {
	m.Events = append(m.Events, Event{kind, text, p, player, strength})
}
func (m *Match) ResetRound() {
	for _, f := range m.Fighters {
		f.Reset()
	}
	m.Timer = 99
	m.Phase = Intro
	m.PhaseTime = 0
	m.Winner = -1
	m.HitStop = 0
	m.Projectiles = m.Projectiles[:0]
	m.Emit("round", "", combat.Vec3{}, -1, 1)
}
func (m *Match) Submit(states [2]input.InputState) {
	for i, f := range m.Fighters {
		f.Input = states[i]
		f.Input.Z = 0
		f.Input.Pressed = nil
		if m.Phase != Fighting {
			continue
		}
		for _, a := range states[i].Pressed {
			f.Buffer.Push(a, m.Time)
		}
		if states[i].Jump {
			f.Input.Jump = true
		}
	}
}
func (m *Match) Tick(dt float32) {
	if m.HitStop > 0 {
		m.HitStop = max(0, m.HitStop-dt)
		return
	}
	m.PhaseTime += dt
	if m.Phase == Intro {
		if m.PhaseTime >= 2.1 {
			m.Phase = Fighting
			m.PhaseTime = 0
			m.Emit("start", "НА ПРОИЗВОДСТВО!", combat.Vec3{}, -1, 1)
		}
		return
	}
	if m.Phase == MatchOver {
		return
	}
	if m.Phase == RoundOver {
		slow := dt
		if m.PhaseTime < 1 {
			slow *= .2
		}
		for _, f := range m.Fighters {
			m.physics(f, slow)
			f.StateTime += slow
			f.DisplayHP = combat.Approach(f.DisplayHP, f.HP, dt*1800)
			f.DamageHP = combat.Approach(f.DamageHP, f.HP, dt*200)
		}
		if m.PhaseTime > 1.05 {
			for i, f := range m.Fighters {
				if i == m.Winner && f.State != fighter.Victory {
					f.SetState(fighter.Victory)
					f.SetAnimation("victory", true, 0)
				}
			}
		}
		if m.PhaseTime > 3.3 {
			if m.Winner >= 0 && m.Wins[m.Winner] >= 2 {
				m.Phase = MatchOver
				m.Emit("victory", m.Fighters[m.Winner].Definition.VictoryText, combat.Vec3{}, m.Winner, 1)
			} else {
				m.Round++
				m.ResetRound()
			}
		}
		return
	}
	m.Time += dt
	m.Timer = max(0, m.Timer-dt)
	for i, f := range m.Fighters {
		target := m.Fighters[1-i].Position.Sub(f.Position)
		target.Y = 0
		target.Z = 0
		if f.Move == nil && target.Len() > .01 {
			f.Facing = target.Unit()
		}
		m.updateFighter(f, dt)
	}
	m.separate()
	// Collect both collisions before resolving damage, allowing simultaneous trades.
	type contact struct {
		owner int
		move  combat.MoveDefinition
		part  int
	}
	var contacts [2]contact
	n := 0
	for i, f := range m.Fighters {
		if f.Move == nil || f.Move.Projectile {
			continue
		}
		part := f.Move.HitIndex(f.MoveTime)
		if part < 0 || f.HitMask&(1<<part) != 0 {
			continue
		}
		opp := m.Fighters[1-i]
		hit := false
		for _, box := range f.Hitboxes() {
			for _, hurt := range opp.Hurtboxes() {
				if combat.Overlap(box, hurt) {
					hit = true
				}
			}
		}
		if hit {
			f.HitMask |= 1 << part
			contacts[n] = contact{i, *f.Move, part}
			n++
		}
	}
	for _, c := range contacts[:n] {
		m.strike(c.owner, c.move, c.part)
	}
	m.updateProjectiles(dt)
	for i, f := range m.Fighters {
		f.Combo.Update(dt, m.Fighters[1-i].Stun > 0)
		f.DisplayHP = combat.Approach(f.DisplayHP, f.HP, dt*1800)
		f.DamageHP = combat.Approach(f.DamageHP, f.HP, dt*200)
	}
	if m.Fighters[0].HP <= 0 || m.Fighters[1].HP <= 0 || m.Timer <= 0 {
		m.finishRound()
	}
}
func (m *Match) finishRound() {
	a, b := m.Fighters[0], m.Fighters[1]
	m.Winner = -1
	if a.HP > b.HP {
		m.Winner = 0
	}
	if b.HP > a.HP {
		m.Winner = 1
	}
	if m.Winner >= 0 {
		m.Wins[m.Winner]++
	}
	m.Phase = RoundOver
	m.PhaseTime = 0
	m.HitStop = .16
	m.Projectiles = m.Projectiles[:0]
	for _, f := range m.Fighters {
		f.Buffer.Clear()
		f.Move = nil
		if f.HP <= 0 {
			f.SetState(fighter.KO)
			f.SetAnimation("knockdown", false, 1.2)
		}
	}
	text := "НОКАУТ"
	if m.Timer <= 0 {
		text = "ВРЕМЯ ВЫШЛО"
	}
	if m.Winner < 0 {
		text = "НИЧЬЯ"
	}
	m.Emit("ko", text, a.Position.Add(b.Position).Mul(.5), m.Winner, 1.3)
}
func (m *Match) updateFighter(f *fighter.Fighter, dt float32) {
	f.StateTime += dt
	if f.Definition.ID == "shuev" && f.Status.Focus > 0 && f.Status.Focus <= dt {
		f.FinisherPending = true
	}
	f.Status.Tick(dt)
	for k, v := range f.Cooldowns {
		f.Cooldowns[k] = max(0, v-dt)
	}
	if f.Input.Block {
		f.BlockAge += dt
	} else {
		f.BlockAge = 0
	}
	m.physics(f, dt)
	if f.Stun > 0 {
		f.Stun = max(0, f.Stun-dt)
		return
	}
	if f.State == fighter.Knockdown {
		if f.StateTime > .82 {
			f.SetState(fighter.GettingUp)
			f.SetAnimation("get_up", false, .65)
			f.Status.Invulnerable = .65
		}
		return
	}
	if f.State == fighter.GettingUp {
		if f.StateTime < .65 {
			return
		}
		f.SetState(fighter.Idle)
	}
	if f.State == fighter.KO || f.State == fighter.Victory {
		return
	}
	if f.Move != nil {
		f.MoveTime += dt
		m.moveEvents(f)
		if f.Move.Dash != 0 && f.MoveTime < f.Move.Startup+f.Move.Active {
			dir := f.Facing
			if f.Move.Animation == "sueta_dash" && (f.Input.X != 0 || f.Input.Z != 0) {
				dir = combat.Vec3{X: f.Input.X, Z: f.Input.Z}.Unit()
			}
			f.Position = f.Position.Add(dir.Mul(f.Move.Dash * dt))
			m.clamp(f)
		}
		a := f.Buffer.Peek(m.Time)
		canCancel := f.Connected && f.MoveTime >= f.Move.Startup+f.Move.Active*.6 && (f.Move.ID == "light" && (a == input.Heavy || a == input.Kick) || f.Move.ID == "heavy" && a == input.Kick)
		if f.Move.Animation == "sueta_dash" && f.MoveTime > .12 && a == input.Light {
			canCancel = true
		}
		if canCancel {
			f.PreviousMove = f.Move.ID
			f.Move = nil
			m.tryInput(f, a)
			f.Buffer.Pop()
			return
		}
		if f.MoveTime < f.Move.Duration() {
			return
		}
		f.PreviousMove = f.Move.ID
		f.Move = nil
	}
	if f.FinisherPending && f.Free() {
		f.FinisherPending = false
		m.startMove(f, f.Definition.Moves["finisher"])
		return
	}
	a := f.Buffer.Peek(m.Time)
	if a != input.None {
		m.tryInput(f, a)
		f.Buffer.Pop()
		if f.Move != nil {
			return
		}
	}
	if f.Input.Jump && f.Position.Y <= .001 {
		f.Velocity.Y = 6.5
		f.SetAnimation("jump", false, .85)
		f.Input.Jump = false
	}
	airborne := f.Position.Y > .01 || f.Velocity.Y > 0
	if airborne {
		f.SetState(fighter.Jumping)
		f.SetAnimation("jump", false, 1)
	} else if f.Input.Block {
		f.SetState(fighter.Blocking)
		f.SetAnimation("block", true, 0)
		return
	} else if f.Input.Crouch {
		f.SetState(fighter.Crouching)
		f.SetAnimation("crouch", false, .35)
		return
	}
	movement := combat.Vec3{X: f.Input.X}
	if movement.Len() > 0 {
		if movement.Len() > 1 {
			movement = movement.Unit()
		}
		speed := 2.55 * f.Speed()
		if airborne {
			speed *= .9
		}
		f.Position = f.Position.Add(movement.Mul(speed * dt))
		m.clamp(f)
		f.Status.Stationary = 0
		if !airborne {
			f.SetState(fighter.Walking)
			f.SetAnimation("walk", true, 0)
		}
	} else if !airborne {
		f.Status.Stationary += dt
		anim := "idle"
		state := fighter.Idle
		if f.Status.Invisible > 0 {
			state = fighter.Invisible
		}
		if f.Definition.ID == "khaliman" && f.Status.Stationary > .8 {
			anim = "standby"
			f.Meter = min(100, f.Meter+dt*3)
		}
		f.SetState(state)
		f.SetAnimation(anim, true, 0)
	}
}
func (m *Match) physics(f *fighter.Fighter, dt float32) {
	f.Position.Z = 0
	f.Velocity.Z = 0
	f.Position = f.Position.Add(f.Velocity.Mul(dt))
	f.Velocity.Y -= 13 * dt
	drag := float32(10)
	if f.Status.Slippery > 0 {
		drag = 2
	}
	f.Velocity.X = combat.Approach(f.Velocity.X, 0, drag*dt)
	f.Velocity.Z = combat.Approach(f.Velocity.Z, 0, drag*dt)
	if f.Position.Y <= 0 {
		f.Position.Y = 0
		f.Velocity.Y = 0
	}
	m.clamp(f)
}
func (m *Match) clamp(f *fighter.Fighter) {
	f.Position.X = combat.Clamp(f.Position.X, -3.6, 3.6)
	f.Position.Z = 0
	f.Velocity.Z = 0
}
func (m *Match) separate() {
	a, b := m.Fighters[0], m.Fighters[1]
	d := b.Position.Sub(a.Position)
	d.Y = 0
	d.Z = 0
	l := d.Len()
	if l < .68 && float32(math.Abs(float64(a.Position.Y-b.Position.Y))) < 1.2 {
		if l < .0001 {
			d.X = a.Facing.X
			if d.X == 0 {
				d.X = 1
			}
		}
		delta := d.Unit().Mul((.68 - l) * .5)
		a.Position = a.Position.Sub(delta)
		b.Position = b.Position.Add(delta)
		m.clamp(a)
		m.clamp(b)
		// Keep the pushboxes separated when either fighter is against a wall.
		remaining := .68 - float32(math.Abs(float64(b.Position.X-a.Position.X)))
		if remaining > 0 {
			if math.Abs(float64(a.Position.X)) >= 3.599 {
				b.Position.X += d.Unit().X * remaining
			} else {
				a.Position.X -= d.Unit().X * remaining
			}
			m.clamp(a)
			m.clamp(b)
		}
	}
}
func (m *Match) tryInput(f *fighter.Fighter, a input.Action) {
	id := ""
	switch a {
	case input.Light:
		id = "light"
	case input.Heavy:
		id = "heavy"
		if f.Input.X*f.Facing.X > .5 {
			id = "hook"
		}
	case input.Kick:
		id = "kick"
		if f.Input.Crouch {
			id = "low"
		}
	case input.Grab:
		id = "grab"
	case input.Special:
		id = "special"
		if f.Status.Omni > 0 {
			id = "rush"
		}
	case input.Secondary:
		id = "secondary"
	case input.Ultimate:
		id = "ultimate"
	}
	move, ok := f.Definition.Moves[id]
	if !ok || f.Cooldowns[id] > 0 || f.Meter < move.Cost {
		return
	}
	if move.Grab && (f.Position.Sub(m.Fighters[1-f.Index].Position).Len() > 1.4 || f.Position.Y > .1) {
		return
	}
	m.startMove(f, move)
}
func (m *Match) startMove(f *fighter.Fighter, move combat.MoveDefinition) {
	if f.Status.Focus > 0 {
		move.Recovery *= .55
		move.Startup *= .8
	}
	if f.Status.QuickRecovery > 0 {
		move.Recovery *= .7
	}
	if f.Definition.ID == "khaliman" && f.PreviousMove == "light" && move.ID == "heavy" {
		move.Damage *= 1.15
		move.HitStun += .12
	}
	if move.Damage > 0 {
		f.Status.Invisible = 0
	}
	m.Serial++
	f.MoveSerial = m.Serial
	f.Move = &move
	f.MoveTime = 0
	f.HitMask = 0
	f.EventCursor = 0
	f.Connected = false
	f.Meter = max(0, f.Meter-move.Cost)
	f.Cooldowns[move.ID] = move.Cooldown
	state := fighter.Attacking
	if move.Grab {
		state = fighter.Grabbing
	}
	if move.ID == "special" || move.Defense {
		state = fighter.Special
	}
	if move.Ultimate {
		state = fighter.Ultimate
	}
	f.SetState(state)
	f.Animation = ""
	f.SetAnimation(move.Animation, false, move.Duration())
	f.Status.Stationary = 0
	dir := m.Fighters[1-f.Index].Position.Sub(f.Position)
	dir.Y = 0
	f.Facing = dir.Unit()
	if move.Ultimate {
		m.HitStop = max(m.HitStop, .1)
		m.Emit("ultimate", move.Name, f.Position, f.Index, 1)
	} else if state == fighter.Special {
		m.Emit("special", move.Name, f.Position, f.Index, .4)
	}
	m.moveEvents(f)
}
func (m *Match) moveEvents(f *fighter.Fighter) {
	// A cursor is not sufficient for intentionally unordered definition events.
	for i, e := range f.Move.Events {
		bit := 1 << i
		if f.EventCursor&bit != 0 || f.MoveTime < e.At {
			continue
		}
		f.EventCursor |= bit
		switch e.Type {
		case combat.EventBuff:
			m.buff(f, e.Value)
		case combat.EventProjectile:
			m.spawnProjectiles(f, *f.Move, e.Value, max(1, e.Count))
		case combat.EventVFX, combat.EventProp:
			m.Emit(e.Value, "", f.Position.Add(combat.Vec3{Y: 1}), f.Index, .7)
			if f.Move.Ultimate && e.Type == combat.EventVFX {
				kind := "sparks"
				if e.Value == "skewer" {
					kind = "smoke"
				}
				if e.Value == "wrench" {
					kind = "dust"
				}
				m.Emit(kind, "", f.Position.Add(f.Facing).Add(combat.Vec3{Y: .4}), f.Index, 1)
			}
		case combat.EventTeleport:
			opp := m.Fighters[1-f.Index]
			f.Position = opp.Position.Add(opp.Facing.Mul(-1.05))
			f.Position.Y = 0
			m.clamp(f)
			f.Facing = opp.Position.Sub(f.Position).Unit()
			f.Status.Invisible = 0
		case combat.EventSound:
			m.Emit("sound", e.Value, f.Position, f.Index, 1)
		}
	}
}
func (m *Match) buff(f *fighter.Fighter, kind string) {
	switch kind {
	case "hearing":
		f.Status.Resistance = 2
	case "invisible":
		f.Status.Invisible = 3
	case "omni":
		f.Status.Omni = 8
	case "quick":
		f.Status.QuickRecovery = 3
		f.Status.Armor = 1.2
	case "counter":
		f.Status.Counter = max(.7, f.Move.CounterWindow)
	case "autoblock":
		f.Status.AutoBlock = 2.5
	case "armor":
		f.Status.Armor = 2
	case "evade":
		f.Status.Invulnerable = .55
		f.Velocity = f.Facing.Mul(-5)
	case "vanish":
		f.Status.Invisible = .5
		f.Status.Invulnerable = .5
	case "gym":
		f.Status.Gym = 6
	case "focus":
		f.Status.Focus = 8
	}
}
func (m *Match) spawnProjectiles(f *fighter.Fighter, move combat.MoveDefinition, kind string, count int) {
	for i := 0; i < count && len(m.Projectiles) < 32; i++ {
		pos := f.Position.Add(f.Facing.Mul(.85)).Add(combat.Vec3{Y: 1.35})
		if kind == "water" {
			pos.Y = .45
		}
		m.Projectiles = append(m.Projectiles, Projectile{Position: pos, Direction: f.Facing, Owner: f.Index, Move: move, Kind: kind, Life: 3.5, Delay: float32(i) * .17, Speed: 5.2})
	}
}
func (m *Match) updateProjectiles(dt float32) {
	// Reserve newly spawned passive projectiles until next step; never invalidate an element pointer.
	count := len(m.Projectiles)
	for i := 0; i < count; i++ {
		p := &m.Projectiles[i]
		if p.Hit {
			continue
		}
		if p.Delay > 0 {
			p.Delay -= dt
			continue
		}
		p.Life -= dt
		if p.Life <= 0 {
			p.Hit = true
			continue
		}
		opp := m.Fighters[1-p.Owner]
		if p.Kind == "bot" {
			p.Direction = opp.Position.Add(combat.Vec3{Y: 1.15}).Sub(p.Position).Unit()
		}
		p.Position = p.Position.Add(p.Direction.Mul(p.Speed * dt))
		hit := combat.Hitbox{Center: p.Position, Radius: .38}
		if p.Kind == "water" {
			hit.Radius = .6
		}
		collided := false
		for _, hurt := range opp.Hurtboxes() {
			if combat.Overlap(hit, hurt) {
				collided = true
				break
			}
		}
		if collided {
			p.Hit = true
			owner, move := p.Owner, p.Move
			m.strike(owner, move, 0)
		}
	}
	kept := m.Projectiles[:0]
	for _, p := range m.Projectiles {
		if !p.Hit {
			kept = append(kept, p)
		}
	}
	m.Projectiles = kept
}
func (m *Match) strike(owner int, move combat.MoveDefinition, part int) {
	a, d := m.Fighters[owner], m.Fighters[1-owner]
	impact := d.Position.Add(combat.Vec3{X: -a.Facing.X * .1, Y: 1.25})
	if len(move.Hitboxes) > 0 {
		impact.Y = a.Position.Y + move.Hitboxes[0].Height
	}
	if d.HP <= 0 || d.State == fighter.Knockdown || d.State == fighter.GettingUp {
		return
	}
	if d.Status.Invulnerable > 0 {
		if d.Definition.ID == "kalachev" {
			d.Meter = min(100, d.Meter+7)
		}
		m.Emit("evade", "УКЛОНЕНИЕ", d.Position, 1-owner, .4)
		return
	}
	if move.Grab && d.Position.Y > .3 {
		return
	}
	if d.Status.Counter > 0 && !move.Grab {
		d.Status.Counter = 0
		d.Move = nil
		d.Stun = 0
		d.SetState(fighter.Idle)
		if d.Definition.ID == "fedoseev" {
			a.Meter = max(0, a.Meter-15)
			d.Position.X -= d.Facing.X * .32
			m.clamp(d)
		}
		counter := d.Definition.Moves["heavy"]
		counter.ID = "counter"
		counter.Name = "КОНТРАТАКА"
		counter.Damage = 90
		counter.HitStun = .7
		counter.KnockbackX = .6
		m.Emit("counter", "КОНТРАТАКА", d.Position, 1-owner, .8)
		m.strike(1-owner, counter, 0)
		d.Meter = min(100, d.Meter+10)
		return
	}
	autoblock := d.Status.AutoBlock > 0 && !move.Grab
	blocked := !move.Grab && (autoblock || (d.Input.Block && d.Move == nil && (d.Stun <= 0 || d.State == fighter.Blocking) && d.Position.Y <= .1))
	damage := move.Damage * a.DamageMultiplier() / d.Definition.Defense
	if d.Definition.ID == "elkhimov" && move.Heavy {
		damage *= .8
	}
	if d.Status.Resistance > 0 {
		damage *= .5
	}
	if d.Status.Armor > 0 {
		damage *= .7
	}
	if blocked {
		damage *= move.Chip
		d.Status.AutoBlock = 0
		d.Stun = max(d.Stun, move.BlockStun)
		d.SetState(fighter.Blocking)
		d.SetAnimation("block", true, 0)
		if autoblock {
			d.Move = nil
		}
		if d.BlockAge < .14 {
			d.Meter = min(100, d.Meter+4)
			d.Stun *= .7
		}
		m.Emit("block", "", impact, 1-owner, .25)
	} else {
		d.ReceivedHits++
		d.Status.Invisible = 0
		stun := move.HitStun
		if d.Status.Resistance > 0 {
			stun *= .2
		}
		if d.Definition.ID == "ivanov" && d.ReceivedHits%3 == 0 {
			stun *= .7
		}
		if d.Status.QuickRecovery > 0 {
			stun *= .7
		}
		last := part >= max(1, move.Hits)-1 || move.Projectile
		if d.Status.Armor <= 0 && d.Status.Resistance <= 0 {
			d.Move = nil
			d.Stun = stun
			d.SetState(fighter.HitStun)
			d.Animation = ""
			d.SetAnimation("hit_react", false, stun)
			if move.Knockdown && last {
				d.Stun = 0
				d.SetState(fighter.Knockdown)
				d.SetAnimation("knockdown", false, .82)
				d.Velocity.Y = move.KnockbackY
			}
		}
		scale := max(.4, 1-float32(a.Combo.Hits)*.075)
		damage *= scale
		a.Combo.Add(damage)
		if a.Status.Focus > 0 {
			a.Combo.Remaining = 1.5
		}
		a.Connected = true
		a.HitCount++
		if a.Definition.ID == "novatskiy" && a.Combo.Hits > 1 {
			a.Meter = min(100, a.Meter+4)
		}
		if a.Definition.ID == "zhirnov" && a.HitCount%3 == 0 && !move.Projectile {
			assist := a.Definition.Moves["special"]
			assist.Damage = 18
			m.spawnProjectiles(a, assist, "bot", 1)
		}
		if move.Effect == "water" {
			d.Status.Slippery = 2.5
		}
		strength := float32(.35)
		kind := "hit"
		if move.Heavy {
			strength = .7
			kind = "heavy"
		}
		if move.Ultimate {
			strength = 1.2
		}
		if move.ID == "kick" || move.ID == "low" {
			kind = "kick"
		}
		m.Emit(kind, "", impact, owner, strength)
		m.HitStop = max(m.HitStop, .025+strength*.055)
	}
	if d.Definition.ID == "kiselik" {
		d.Status.Sueta = min(5, d.Status.Sueta+1)
	}
	d.HP = max(0, d.HP-damage)
	a.Meter = min(100, a.Meter+move.MeterGain)
	d.Meter = min(100, d.Meter+damage*.055)
	push := move.KnockbackX
	if blocked {
		push = .2
	}
	if d.Status.QuickRecovery > 0 || d.Status.Armor > 0 {
		push *= .25
	}
	if move.Hits > 1 && part < move.Hits-1 {
		push *= .15
	}
	dir := d.Position.Sub(a.Position)
	dir.Y = 0
	d.Velocity = d.Velocity.Add(dir.Unit().Mul(push * 5))
}
