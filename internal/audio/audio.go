package audio

import (
	"encoding/binary"
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"math"
	"os"
	"path/filepath"
)

type Manager struct {
	Ready    bool
	Sounds   map[string]rl.Sound
	Music    rl.Music
	HasMusic bool
	SFX      float32
}

func New(root string) *Manager {
	a := &Manager{Sounds: make(map[string]rl.Sound), SFX: .8}
	rl.InitAudioDevice()
	a.Ready = rl.IsAudioDeviceReady()
	if !a.Ready {
		log.Print("[ПРЕДУПРЕЖДЕНИЕ] Аудиоустройство недоступно")
		return a
	}
	kinds := []string{"menu", "round", "start", "hit", "heavy", "kick", "block", "special", "ultimate", "ko", "victory", "counter"}
	for i, k := range kinds {
		path := filepath.Join(root, "assets/audio/"+k+".wav")
		if _, e := os.Stat(path); e == nil {
			a.Sounds[k] = rl.LoadSound(path)
			continue
		}
		data := synth(float64(180+i*47), .12+float64(i%4)*.06, k == "hit" || k == "heavy" || k == "kick")
		w := rl.LoadWaveFromMemory(".wav", data, int32(len(data)))
		a.Sounds[k] = rl.LoadSoundFromWave(w)
		rl.UnloadWave(w)
	}
	log.Print("[FightOGM] Аудио: отсутствующие WAV заменены встроенными синтезированными сигналами")
	path := filepath.Join(root, "assets/audio/music.ogg")
	if _, e := os.Stat(path); e == nil {
		a.Music = rl.LoadMusicStream(path)
		a.HasMusic = a.Music.FrameCount > 0
		if a.HasMusic {
			a.Music.Looping = true
			rl.PlayMusicStream(a.Music)
		}
	} else {
		log.Print("[ПРЕДУПРЕЖДЕНИЕ] assets/audio/music.ogg отсутствует; музыка отключена")
	}
	return a
}
func (a *Manager) Volumes(master, music, sfx float32) {
	if !a.Ready {
		return
	}
	rl.SetMasterVolume(master)
	a.SFX = sfx
	if a.HasMusic {
		rl.SetMusicVolume(a.Music, music)
	}
}
func (a *Manager) Play(kind string) {
	if !a.Ready {
		return
	}
	if s, ok := a.Sounds[kind]; ok && s.FrameCount > 0 {
		rl.SetSoundVolume(s, a.SFX)
		rl.PlaySound(s)
	}
}
func (a *Manager) Update() {
	if a.Ready && a.HasMusic {
		rl.UpdateMusicStream(a.Music)
	}
}
func (a *Manager) Close() {
	if !a.Ready {
		return
	}
	for _, s := range a.Sounds {
		rl.UnloadSound(s)
	}
	if a.HasMusic {
		rl.UnloadMusicStream(a.Music)
	}
	rl.CloseAudioDevice()
}
func synth(freq, seconds float64, noise bool) []byte {
	const rate = 22050
	n := int(seconds * rate)
	b := make([]byte, 44+n*2)
	copy(b, "RIFF")
	binary.LittleEndian.PutUint32(b[4:], uint32(len(b)-8))
	copy(b[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(b[16:], 16)
	binary.LittleEndian.PutUint16(b[20:], 1)
	binary.LittleEndian.PutUint16(b[22:], 1)
	binary.LittleEndian.PutUint32(b[24:], rate)
	binary.LittleEndian.PutUint32(b[28:], rate*2)
	binary.LittleEndian.PutUint16(b[32:], 2)
	binary.LittleEndian.PutUint16(b[34:], 16)
	copy(b[36:], "data")
	binary.LittleEndian.PutUint32(b[40:], uint32(n*2))
	var seed uint32 = 37
	for i := 0; i < n; i++ {
		t := float64(i) / rate
		envelope := math.Pow(1-float64(i)/float64(n), 2) * math.Min(1, t*500)
		v := math.Sin(2 * math.Pi * (freq*t - freq*t*t))
		if noise {
			seed = seed*1664525 + 1013904223
			v = .6*v + .4*(float64(seed>>8)/8388608-1)
		}
		sample := int16(v * envelope * 10000)
		binary.LittleEndian.PutUint16(b[44+i*2:], uint16(sample))
	}
	return b
}
