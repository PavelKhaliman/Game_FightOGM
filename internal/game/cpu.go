package game

import (
	"fightogm/internal/input"
	"math"
)

// CPU makes decisions at a human-readable cadence and submits ordinary inputs.
// It uses the same movement, cooldown, meter and hit rules as the human player.
type CPU struct {
	UntilDecision float32
	Seed          uint32
	Held          input.InputState
}

func (c *CPU) Input(m *Match, dt float32) input.InputState {
	if m.Phase != Fighting {
		return input.InputState{}
	}
	c.UntilDecision -= dt
	if c.UntilDecision > 0 {
		s := c.Held
		s.Pressed = nil
		s.Jump = false
		return s
	}
	if c.Seed == 0 {
		c.Seed = 1947
	}
	c.Seed = c.Seed*1664525 + 1013904223
	roll := int(c.Seed>>16) % 100
	c.UntilDecision = .19 + float32(roll%12)*.014
	self, enemy := m.Fighters[1], m.Fighters[0]
	difference := enemy.Position.X - self.Position.X
	distance := float32(math.Abs(float64(difference)))
	direction := float32(1)
	if difference < 0 {
		direction = -1
	}
	s := input.InputState{}
	if distance > 1.12 {
		s.X = direction
	}
	if distance < .83 && roll < 18 {
		s.X = -direction
	}
	if enemy.Move != nil && distance < 1.9 && roll < 63 {
		s.Block = true
		s.X = 0
	}
	if self.Free() {
		switch {
		case self.Meter >= 100 && roll < 62:
			s.Pressed = []input.Action{input.Ultimate}
		case self.Status.Omni > 0 && distance < 3 && self.Cooldowns["rush"] <= 0:
			s.Pressed = []input.Action{input.Special}
		case distance < 1.15 && roll > 28:
			actions := []input.Action{input.Light, input.Light, input.Heavy, input.Kick, input.Grab}
			s.Pressed = []input.Action{actions[roll%len(actions)]}
			s.Block = false
			s.Crouch = roll > 88
		case distance < 1.8 && roll < 10:
			s.Jump = true
			s.X = direction
		case self.Cooldowns["special"] <= 0 && roll > 83:
			s.Pressed = []input.Action{input.Special}
		}
	}
	c.Held = s
	return s
}
