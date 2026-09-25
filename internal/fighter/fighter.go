package fighter

import (
	"fightogm/internal/combat"
	"fightogm/internal/input"
)

type State uint8

const (
	Idle State = iota
	Walking
	Crouching
	Jumping
	Attacking
	Blocking
	HitStun
	Knockdown
	GettingUp
	Grabbing
	Special
	Ultimate
	Invisible
	KO
	Victory
)

func (s State) String() string {
	return [...]string{"Стойка", "Ходьба", "Присед", "Прыжок", "Атака", "Блок", "Оглушение", "Падение", "Подъём", "Захват", "Приём", "Суперприём", "Невидимость", "Нокаут", "Победа"}[s]
}

type Status struct {
	Armor, Resistance, Invisible, Invulnerable, Counter, AutoBlock float32
	Omni, Gym, Focus, Slippery, QuickRecovery                      float32
	Capacitor                                                      float32
	Sueta                                                          int
	Stationary                                                     float32
}

func (s *Status) Tick(dt float32) {
	for _, p := range []*float32{&s.Armor, &s.Resistance, &s.Invisible, &s.Invulnerable, &s.Counter, &s.AutoBlock, &s.Omni, &s.Gym, &s.Focus, &s.Slippery, &s.QuickRecovery, &s.Capacitor} {
		*p = max(0, *p-dt)
	}
}

type Fighter struct {
	Definition                     Definition
	Index                          int
	Position, Velocity, Facing     combat.Vec3
	HP, DisplayHP, DamageHP, Meter float32
	State                          State
	StateTime, Stun, BlockAge      float32
	Move                           *combat.MoveDefinition
	MoveTime                       float32
	MoveSerial, AnimationSerial    uint64
	HitMask                        uint32
	EventCursor                    int
	Connected                      bool
	Cooldowns                      map[string]float32
	Status                         Status
	Buffer                         input.InputBuffer
	Input                          input.InputState
	Combo                          combat.Combo
	Animation                      string
	AnimationLoop                  bool
	AnimationDuration              float32
	PreviousMove                   string
	FinisherPending                bool
	HitCount, ReceivedHits         int
}

func New(def Definition, index int) *Fighter {
	f := &Fighter{Definition: def, Index: index}
	f.Reset()
	return f
}
func (f *Fighter) Reset() {
	def, index := f.Definition, f.Index
	*f = Fighter{Definition: def, Index: index, HP: def.MaxHP, DisplayHP: def.MaxHP, DamageHP: def.MaxHP, Cooldowns: make(map[string]float32)}
	f.Position = combat.Vec3{X: -1.35 + 2.7*float32(index)}
	f.Facing = combat.Vec3{X: 1 - 2*float32(index)}
	f.SetAnimation("idle", true, 0)
}
func (f *Fighter) SetState(s State) {
	if f.State != s {
		f.State = s
		f.StateTime = 0
	}
}
func (f *Fighter) SetAnimation(name string, loop bool, duration float32) {
	if f.Animation != name {
		f.AnimationSerial++
		f.Animation = name
		f.AnimationLoop = loop
		f.AnimationDuration = duration
	}
}
func (f *Fighter) Free() bool {
	return f.Move == nil && f.Stun <= 0 && f.State != Knockdown && f.State != GettingUp && f.State != KO && f.State != Victory
}
func (f *Fighter) Hurtboxes() [4]combat.Hurtbox {
	scale := float32(1)
	if f.State == Crouching {
		scale = .65
	}
	p := f.Position
	return [4]combat.Hurtbox{
		{Center: p.Add(combat.Vec3{Y: 1.72 * scale}), Radius: .23, Name: "Голова"},
		{Center: p.Add(combat.Vec3{Y: 1.32 * scale}), Radius: .34, Name: "Грудь"},
		{Center: p.Add(combat.Vec3{Y: .91 * scale}), Radius: .29, Name: "Корпус"},
		{Center: p.Add(combat.Vec3{Y: .4 * scale}), Radius: .27, Name: "Ноги"},
	}
}
func (f *Fighter) Hitboxes() []combat.Hitbox {
	if f.Move == nil || f.Move.Projectile || f.Move.HitIndex(f.MoveTime) < 0 {
		return nil
	}
	boxes := make([]combat.Hitbox, 0, len(f.Move.Hitboxes))
	for _, b := range f.Move.Hitboxes {
		boxes = append(boxes, combat.Hitbox{Center: f.Position.Add(f.Facing.Mul(b.Forward)).Add(combat.Vec3{Y: b.Height}), Radius: b.Radius, Instance: f.MoveSerial})
	}
	return boxes
}
func (f *Fighter) DamageMultiplier() float32 {
	p := f.Definition.Power
	if f.Status.Omni > 0 {
		p *= 1.45
	}
	if f.Status.Gym > 0 {
		p *= 1.3
	}
	return p
}
func (f *Fighter) Speed() float32 {
	s := f.Definition.Speed * (1 + float32(f.Status.Sueta)*.035)
	if f.Status.Omni > 0 {
		s *= 1.35
	}
	if f.Status.Focus > 0 {
		s *= 1.3
	}
	return s
}
