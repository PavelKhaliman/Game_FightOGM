package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadAndInvalidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "settings.json")
	s := Defaults()
	s.Master = .25
	s.Width = 1920
	s.Height = 1080
	s.Mappings[0].Light = 90
	if e := s.Save(path); e != nil {
		t.Fatal(e)
	}
	loaded := Load(path)
	if loaded != s {
		t.Fatalf("settings did not round trip: %+v", loaded)
	}
	if e := os.WriteFile(path, []byte("broken JSON"), 0600); e != nil {
		t.Fatal(e)
	}
	if Load(path) != Defaults() {
		t.Fatal("corrupt settings did not use defaults")
	}
}
