// Package combat contains renderer-independent combat volumes and move data.
package combat

import "math"

type Vec3 struct{ X, Y, Z float32 }

func (a Vec3) Add(b Vec3) Vec3    { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3    { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Mul(s float32) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }
func (a Vec3) Len() float32       { return float32(math.Sqrt(float64(a.X*a.X + a.Y*a.Y + a.Z*a.Z))) }
func (a Vec3) Unit() Vec3 {
	if l := a.Len(); l > 0.0001 {
		return a.Mul(1 / l)
	}
	return Vec3{X: 1}
}
func Clamp(x, lo, hi float32) float32 { return max(lo, min(hi, x)) }
func Approach(x, target, step float32) float32 {
	if x < target {
		return min(x+step, target)
	}
	return max(x-step, target)
}

type Hitbox struct {
	Center   Vec3
	Radius   float32
	Instance uint64
}
type Hurtbox struct {
	Center Vec3
	Radius float32
	Name   string
}

func Overlap(hit Hitbox, hurt Hurtbox) bool {
	return hit.Center.Sub(hurt.Center).Len() <= hit.Radius+hurt.Radius
}

type EventType int

const (
	EventProp EventType = iota
	EventVFX
	EventProjectile
	EventBuff
	EventTeleport
	EventSound
)

type MoveEvent struct {
	At    float32
	Type  EventType
	Value string
	Count int
}
type HitboxDefinition struct{ Forward, Height, Radius float32 }
type MoveDefinition struct {
	ID, Name, Animation, Effect                                         string
	Startup, Active, Recovery                                           float32
	Damage, Chip, HitStun, BlockStun, KnockbackX, KnockbackY, MeterGain float32
	Cooldown, Cost, Dash, CounterWindow                                 float32
	Hits                                                                int
	Heavy, Grab, Knockdown, Projectile, Defense, Ultimate               bool
	Hitboxes                                                            []HitboxDefinition
	Events                                                              []MoveEvent
}

func (m MoveDefinition) Duration() float32 { return m.Startup + m.Active + m.Recovery }
func (m MoveDefinition) HitIndex(t float32) int {
	if t < m.Startup || t >= m.Startup+m.Active || m.Damage <= 0 {
		return -1
	}
	hits := max(1, m.Hits)
	return min(hits-1, int((t-m.Startup)/m.Active*float32(hits)))
}

type Combo struct {
	Hits              int
	Damage, Remaining float32
}

func (c *Combo) Add(damage float32) { c.Hits++; c.Damage += damage; c.Remaining = 1.1 }
func (c *Combo) Update(dt float32, opponentStunned bool) {
	if opponentStunned {
		c.Remaining = max(c.Remaining, .3)
	}
	c.Remaining -= dt
	if c.Remaining <= 0 {
		*c = Combo{}
	}
}
