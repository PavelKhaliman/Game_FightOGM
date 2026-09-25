package assets

import (
	"fightogm/internal/characters"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestPlayableRosterHasAllMoveAnimations(t *testing.T) {
	root := filepath.Join("..", "..")
	spec, err := ReadSpriteRoster(root)
	if err != nil {
		t.Fatal(err)
	}
	definitions := characters.Roster()
	if len(spec.Characters) != len(definitions) {
		t.Fatalf("playable roster has %d fighters, definitions have %d", len(spec.Characters), len(definitions))
	}
	seen := map[string]bool{}
	for _, id := range spec.Characters {
		if seen[id] {
			t.Fatalf("duplicate fighter %s", id)
		}
		seen[id] = true
		atlas, err := ReadSpriteAtlas(filepath.Join(root, "assets", "sprites", id, "animations.json"))
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range definitions {
			if d.ID == id {
				for _, move := range d.Moves {
					if move.Duration() <= 0 {
						t.Errorf("%s: invalid timing for %s", id, move.ID)
					}
					if _, ok := atlas.Clips[move.Animation]; !ok {
						t.Errorf("%s: missing %s for %s", id, move.Animation, move.ID)
					}
				}
			}
		}
		if id == "khaliman" {
			if _, ok := atlas.Clips["standby"]; !ok {
				t.Fatal("Khaliman standby is missing")
			}
		}
	}
	for _, d := range definitions {
		if !seen[d.ID] {
			t.Errorf("fighter %s is not selectable", d.ID)
		}
	}
}

func TestIvanovAtlasImagesAndAttackTiming(t *testing.T) {
	dir := filepath.Join("..", "..", "assets", "sprites", "ivanov")
	atlas, err := ReadSpriteAtlas(filepath.Join(dir, "animations.json"))
	if err != nil {
		t.Fatal(err)
	}
	sizes := map[string]image.Config{}
	for id, f := range atlas.Frames {
		size, ok := sizes[f.Sheet]
		if !ok {
			file, e := os.Open(filepath.Join(dir, f.Sheet))
			if e != nil {
				t.Fatal(e)
			}
			size, _, e = image.DecodeConfig(file)
			file.Close()
			if e != nil {
				t.Fatal(e)
			}
			sizes[f.Sheet] = size
		}
		if f.Rect[0]+f.Rect[2] > size.Width || f.Rect[1]+f.Rect[3] > size.Height {
			t.Fatalf("clipped frame %s", id)
		}
	}
	// Changing the move speed must also move its contact pose, regardless of FPS.
	for _, name := range []string{"punch_light", "punch_heavy", "kick_high", "kick_low", "grab", "omni_rush"} {
		clip := atlas.Clips[name]
		for _, speed := range []float32{.6, 1, 1.7} {
			startup, active := .2/speed, .12/speed
			for _, sample := range []struct {
				time float32
				want int
			}{{0, 0}, {startup - .001, 0}, {startup, 1}, {startup + active*.5, 1}, {startup + active + .001, 2}} {
				if got := clip.FrameAt(sample.time, 1/speed, startup, active); got != sample.want {
					t.Fatalf("%s: speed %f time %f got %d want %d", name, speed, sample.time, got, sample.want)
				}
			}
		}
	}
}
