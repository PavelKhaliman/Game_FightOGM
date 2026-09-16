package input

import rl "github.com/gen2brain/raylib-go/raylib"

type InputMapping struct{ Left, Right, Jump, Crouch, StepBack, StepFront, Light, Heavy, Kick, Block, Special, Ultimate int32 }

func DefaultMappings() [2]InputMapping {
	return [2]InputMapping{
		{rl.KeyA, rl.KeyD, rl.KeyW, rl.KeyS, rl.KeyQ, rl.KeyE, rl.KeyF, rl.KeyG, rl.KeyH, rl.KeyR, rl.KeyT, rl.KeyY},
		{rl.KeyLeft, rl.KeyRight, rl.KeyUp, rl.KeyDown, rl.KeyU, rl.KeyM, rl.KeyJ, rl.KeyK, rl.KeyL, rl.KeyO, rl.KeyP, rl.KeyI},
	}
}

type Frame struct {
	Players                                                            [2]InputState
	SelectionX, SelectionY                                             [2]int
	Confirm, Cancel                                                    [2]bool
	Up, Down, Left, Right, Accept, Back, Fullscreen, Debug, MouseClick bool
	MouseX, MouseY                                                     float32
}

func Poll(mappings [2]InputMapping) Frame {
	down := rl.IsKeyDown
	pressed := rl.IsKeyPressed
	var f Frame
	for i, m := range mappings {
		s := &f.Players[i]
		if down(m.Left) {
			s.X--
		}
		if down(m.Right) {
			s.X++
		}
		s.Jump = pressed(m.Jump)
		s.Crouch = down(m.Crouch)
		s.Block = down(m.Block)
		if pressed(m.Light) && down(m.Heavy) || pressed(m.Heavy) && down(m.Light) {
			s.Pressed = append(s.Pressed, Grab)
		} else {
			if pressed(m.Light) {
				s.Pressed = append(s.Pressed, Light)
			}
			if pressed(m.Heavy) {
				s.Pressed = append(s.Pressed, Heavy)
			}
		}
		if pressed(m.Kick) {
			s.Pressed = append(s.Pressed, Kick)
		}
		if pressed(m.Special) {
			a := Special
			if s.Block {
				a = Secondary
			}
			s.Pressed = append(s.Pressed, a)
		}
		if pressed(m.Ultimate) {
			s.Pressed = append(s.Pressed, Ultimate)
		}
		if pressed(m.Left) {
			f.SelectionX[i]--
		}
		if pressed(m.Right) {
			f.SelectionX[i]++
		}
		if pressed(m.Jump) {
			f.SelectionY[i]--
		}
		if pressed(m.Crouch) {
			f.SelectionY[i]++
		}
		f.Confirm[i] = pressed(m.Light)
		f.Cancel[i] = pressed(m.Block)
	}
	f.Up = pressed(rl.KeyUp) || pressed(rl.KeyW)
	f.Down = pressed(rl.KeyDown) || pressed(rl.KeyS)
	f.Left = pressed(rl.KeyLeft) || pressed(rl.KeyA)
	f.Right = pressed(rl.KeyRight) || pressed(rl.KeyD)
	f.Accept = pressed(rl.KeyEnter) || pressed(rl.KeySpace)
	f.Back = pressed(rl.KeyEscape)
	f.Debug = pressed(rl.KeyF3)
	f.Fullscreen = pressed(rl.KeyEnter) && (down(rl.KeyLeftAlt) || down(rl.KeyRightAlt))
	if f.Fullscreen {
		f.Accept = false
	}
	pos := rl.GetMousePosition()
	f.MouseX = pos.X
	f.MouseY = pos.Y
	f.MouseClick = rl.IsMouseButtonPressed(rl.MouseLeftButton)
	return f
}
