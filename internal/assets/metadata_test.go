package assets

import (
	"fightogm/internal/characters"
	"os"
	"path/filepath"
	"testing"
)

func TestRosterManifestAndGLBClips(t *testing.T) {
	root := filepath.Join("..", "..")
	manifest, e := ReadManifest(root)
	if e != nil {
		t.Fatal(e)
	}
	roster := characters.Roster()
	if len(roster) != 10 || len(manifest.Characters) != 10 {
		t.Fatal("roster must contain exactly ten fighters")
	}
	for _, d := range roster {
		t.Run(d.ID, func(t *testing.T) {
			entry, ok := manifest.Characters[d.ID]
			if !ok {
				t.Fatal("fighter missing from manifest")
			}
			meta, e := ReadGLB(filepath.Join(root, d.ModelPath))
			if e != nil {
				t.Fatal(e)
			}
			names := map[string]bool{}
			for _, c := range meta.Clips {
				if c.Duration <= 0 {
					t.Fatal("animation has no duration")
				}
				names[c.Name] = true
			}
			for _, name := range append(entry.Base, entry.Special...) {
				if !names[name] {
					t.Fatalf("missing clip %s", name)
				}
			}
			for _, move := range d.Moves {
				if !names[move.Animation] {
					t.Fatalf("move references unknown animation: %s", move.Animation)
				}
				if move.Duration() <= 0 {
					t.Fatal("invalid move timing")
				}
			}
		})
	}
}
func TestCorruptGLBReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.glb")
	if e := os.WriteFile(path, []byte("glTFbad"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := ReadGLB(path); e == nil {
		t.Fatal("corrupt file accepted")
	}
}
