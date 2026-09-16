package config

import (
	"encoding/json"
	"fightogm/internal/input"
	"log"
	"os"
	"path/filepath"
)

type Settings struct {
	Width      int                   `json:"width"`
	Height     int                   `json:"height"`
	Fullscreen bool                  `json:"fullscreen"`
	Master     float32               `json:"master_volume"`
	Music      float32               `json:"music_volume"`
	SFX        float32               `json:"sfx_volume"`
	Mappings   [2]input.InputMapping `json:"key_bindings"`
}

func Defaults() Settings {
	return Settings{Width: 1600, Height: 900, Master: .7, Music: .35, SFX: .8, Mappings: input.DefaultMappings()}
}
func Load(path string) Settings {
	s := Defaults()
	data, e := os.ReadFile(path)
	if e != nil {
		if !os.IsNotExist(e) {
			log.Printf("[ПРЕДУПРЕЖДЕНИЕ] настройки: %v", e)
		}
		return s
	}
	if e = json.Unmarshal(data, &s); e != nil {
		log.Printf("[ПРЕДУПРЕЖДЕНИЕ] настройки: %v", e)
		return Defaults()
	}
	if s.Width < 960 || s.Width > 7680 || s.Height < 540 || s.Height > 4320 {
		s.Width = 1600
		s.Height = 900
	}
	s.Master = max(0, min(1, s.Master))
	s.Music = max(0, min(1, s.Music))
	s.SFX = max(0, min(1, s.SFX))
	for i, b := range s.Mappings {
		if b.Light == 0 || b.Left == 0 {
			s.Mappings[i] = input.DefaultMappings()[i]
		}
	}
	return s
}
func (s Settings) Save(path string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	tmp := path + ".tmp"
	if e = os.WriteFile(tmp, b, 0644); e != nil {
		return e
	}
	return os.Rename(tmp, path)
}
