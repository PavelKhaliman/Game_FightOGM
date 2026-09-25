package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyManifestAndGLBClips(t *testing.T) {
	root := filepath.Join("..", "..")
	manifest, e := ReadManifest(root)
	if e != nil {
		t.Fatal(e)
	}
	if len(manifest.Characters) != 10 {
		t.Fatal("archived 3D manifest must retain its ten original fighters")
	}
	for id, entry := range manifest.Characters {
		t.Run(id, func(t *testing.T) {
			meta, e := ReadGLB(filepath.Join(root, "assets", "models", id+".glb"))
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
