package input

import "testing"

func TestBufferExpiresConsumesAndBounds(t *testing.T) {
	var b InputBuffer
	b.Push(Light, 1)
	if b.Peek(1.1) != Light {
		t.Fatal("valid command expired")
	}
	b.Pop()
	if b.Peek(1.1) != None {
		t.Fatal("command consumed twice")
	}
	b.Push(Heavy, 1)
	if b.Peek(1.2) != None {
		t.Fatal("stale command retained")
	}
	for i := 0; i < 30; i++ {
		b.Push(Kick, 2)
	}
	if b.Count != 12 {
		t.Fatal("buffer overflow")
	}
	b.Clear()
	if b.Peek(2) != None {
		t.Fatal("clear failed")
	}
}
