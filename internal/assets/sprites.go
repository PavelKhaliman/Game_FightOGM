package assets

import (
	"encoding/json"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"path/filepath"
)

type SpriteRoster struct {
	Characters []string `json:"characters"`
	Stage      string   `json:"stage"`
}

type SpriteFrame struct {
	Sheet       string     `json:"sheet"`
	Rect        [4]int     `json:"rect"`
	Pivot       [2]float32 `json:"pivot"`
	WorldHeight float32    `json:"world_height"`
	Regions     [][4]int   `json:"regions,omitempty"`
}

type SpriteClip struct {
	Frames []string `json:"frames"`
	FPS    float32  `json:"fps"`
	Loop   bool     `json:"loop"`
	Attack bool     `json:"attack"`
}

type SpriteAtlas struct {
	Frames   map[string]SpriteFrame  `json:"frames"`
	Clips    map[string]SpriteClip   `json:"clips"`
	Textures map[string]rl.Texture2D `json:"-"`
}

func ReadSpriteRoster(root string) (SpriteRoster, error) {
	var roster SpriteRoster
	data, err := os.ReadFile(filepath.Join(root, "assets/sprites/roster.json"))
	if err != nil {
		return roster, err
	}
	err = json.Unmarshal(data, &roster)
	if err == nil && (len(roster.Characters) == 0 || roster.Stage == "") {
		err = fmt.Errorf("пустой список 2D-бойцов или фон арены")
	}
	return roster, err
}

func ReadSpriteAtlas(path string) (*SpriteAtlas, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var atlas SpriteAtlas
	if err = json.Unmarshal(data, &atlas); err != nil {
		return nil, err
	}
	if len(atlas.Frames) == 0 {
		return nil, fmt.Errorf("нет кадров спрайта: %s", path)
	}
	for name, frame := range atlas.Frames {
		if frame.Rect[0] < 0 || frame.Rect[1] < 0 || frame.Rect[2] <= 0 || frame.Rect[3] <= 0 || frame.WorldHeight <= 0 || frame.Sheet == "" {
			return nil, fmt.Errorf("неверный кадр %s", name)
		}
		if frame.Pivot[0] < 0 || frame.Pivot[0] > float32(frame.Rect[2]) || frame.Pivot[1] < 0 || frame.Pivot[1] > float32(frame.Rect[3]) {
			return nil, fmt.Errorf("неверная опора кадра %s", name)
		}
		for _, region := range frame.Regions {
			if region[0] < 0 || region[1] < 0 || region[2] <= 0 || region[3] <= 0 || region[0]+region[2] > frame.Rect[2] || region[1]+region[3] > frame.Rect[3] {
				return nil, fmt.Errorf("неверная область отрисовки кадра %s", name)
			}
		}
	}
	for name, clip := range atlas.Clips {
		if len(clip.Frames) == 0 || clip.FPS <= 0 || (clip.Attack && len(clip.Frames) != 3) {
			return nil, fmt.Errorf("неверная анимация %s", name)
		}
		for _, id := range clip.Frames {
			if _, ok := atlas.Frames[id]; !ok {
				return nil, fmt.Errorf("анимация %s: нет кадра %s", name, id)
			}
		}
	}
	for _, name := range []string{"idle", "walk", "crouch", "jump", "punch_light", "punch_heavy", "kick_high", "kick_low", "block", "hit_react", "knockdown", "get_up", "grab", "victory"} {
		if _, ok := atlas.Clips[name]; !ok {
			return nil, fmt.Errorf("нет 2D-анимации %s", name)
		}
	}
	return &atlas, nil
}

func LoadSpriteAtlas(path string) (*SpriteAtlas, error) {
	atlas, err := ReadSpriteAtlas(path)
	if err != nil {
		return nil, err
	}
	atlas.Textures = make(map[string]rl.Texture2D)
	for _, frame := range atlas.Frames {
		tex, ok := atlas.Textures[frame.Sheet]
		if !ok {
			tex = rl.LoadTexture(filepath.Join(filepath.Dir(path), frame.Sheet))
			if tex.ID == 0 {
				atlas.Close()
				return nil, fmt.Errorf("не удалось загрузить спрайт %s", frame.Sheet)
			}
			rl.SetTextureFilter(tex, rl.FilterBilinear)
			atlas.Textures[frame.Sheet] = tex
		}
		if frame.Rect[0]+frame.Rect[2] > int(tex.Width) || frame.Rect[1]+frame.Rect[3] > int(tex.Height) {
			atlas.Close()
			return nil, fmt.Errorf("кадр выходит за границы %s", frame.Sheet)
		}
	}
	return atlas, nil
}

func (a *SpriteAtlas) Close() {
	for _, tex := range a.Textures {
		rl.UnloadTexture(tex)
	}
	a.Textures = nil
}

// FrameAt uses game time. Impact poses coincide with the move's active window.
func (c SpriteClip) FrameAt(elapsed, duration, startup, active float32) int {
	if c.Attack && startup >= 0 {
		if elapsed < startup {
			return 0
		}
		if elapsed < startup+active {
			return 1
		}
		return 2
	}
	if c.Loop {
		return int(max(0, elapsed)*c.FPS) % len(c.Frames)
	}
	if duration <= 0 {
		duration = float32(len(c.Frames)) / c.FPS
	}
	return min(len(c.Frames)-1, int(max(0, elapsed)/duration*float32(len(c.Frames))))
}
