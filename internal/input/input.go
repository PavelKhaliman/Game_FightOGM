package input

type Action uint8

const (
	None Action = iota
	Light
	Heavy
	Kick
	Grab
	Special
	Secondary
	Ultimate
)

type InputState struct {
	X, Z                float32
	Crouch, Block, Jump bool
	Pressed             []Action
}
type Entry struct {
	Action  Action
	Expires float32
}

// InputBuffer consumes each press once. Age is simulation time, so pause freezes it.
type InputBuffer struct {
	Entries [12]Entry
	Count   int
}

const BufferSeconds float32 = .135

func (b *InputBuffer) Push(a Action, now float32) {
	if a == None {
		return
	}
	if b.Count == len(b.Entries) {
		copy(b.Entries[:], b.Entries[1:])
		b.Count--
	}
	b.Entries[b.Count] = Entry{a, now + BufferSeconds}
	b.Count++
}
func (b *InputBuffer) Peek(now float32) Action {
	for b.Count > 0 && b.Entries[0].Expires < now {
		b.Pop()
	}
	if b.Count == 0 {
		return None
	}
	return b.Entries[0].Action
}
func (b *InputBuffer) Pop() {
	if b.Count > 0 {
		copy(b.Entries[:], b.Entries[1:b.Count])
		b.Count--
	}
}
func (b *InputBuffer) Clear() { b.Count = 0 }
