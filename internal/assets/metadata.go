package assets

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ManifestEntry struct {
	File    string   `json:"file"`
	Base    []string `json:"base_animations"`
	Special []string `json:"special_animations"`
}
type Manifest struct {
	Characters map[string]ManifestEntry `json:"characters"`
}

func ReadManifest(root string) (Manifest, error) {
	var m Manifest
	b, e := os.ReadFile(filepath.Join(root, "assets/models/special_animation_manifest.json"))
	if e != nil {
		return m, e
	}
	e = json.Unmarshal(b, &m)
	return m, e
}

type Clip struct {
	Name     string
	Duration float32
}
type Metadata struct {
	Clips []Clip
	ZUp   bool
}

func ReadGLB(path string) (Metadata, error) {
	var out Metadata
	data, e := os.ReadFile(path)
	if e != nil {
		return out, e
	}
	if len(data) < 20 || string(data[:4]) != "glTF" || binary.LittleEndian.Uint32(data[4:]) != 2 || int(binary.LittleEndian.Uint32(data[8:])) != len(data) {
		return out, fmt.Errorf("некорректный GLB: %s", path)
	}
	size := int(binary.LittleEndian.Uint32(data[12:]))
	if size > len(data)-20 || binary.LittleEndian.Uint32(data[16:]) != 0x4E4F534A {
		return out, fmt.Errorf("нет JSON-секции: %s", path)
	}
	var doc struct {
		Animations []struct {
			Name     string
			Samplers []struct{ Input int }
		}
		Accessors []struct{ Max, Min []float32 }
		Skins     []struct{ Joints []int }
	}
	if e = json.Unmarshal(data[20:20+size], &doc); e != nil {
		return out, e
	}
	if len(doc.Skins) == 0 || len(doc.Skins[0].Joints) == 0 {
		return out, fmt.Errorf("нет скелета: %s", path)
	}
	for _, a := range doc.Animations {
		c := Clip{Name: a.Name}
		for _, s := range a.Samplers {
			if s.Input < 0 || s.Input >= len(doc.Accessors) {
				return out, fmt.Errorf("неверный индекс анимации")
			}
			v := doc.Accessors[s.Input].Max
			if len(v) > 0 {
				c.Duration = max(c.Duration, v[0])
			}
		}
		out.Clips = append(out.Clips, c)
	}
	if len(doc.Accessors) > 0 {
		a := doc.Accessors[0]
		if len(a.Max) == 3 && len(a.Min) == 3 {
			out.ZUp = a.Max[2]-a.Min[2] > a.Max[1]-a.Min[1]
		}
	}
	return out, nil
}
