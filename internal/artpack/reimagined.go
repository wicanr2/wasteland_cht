package artpack

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

var ReimaginedDirections = [...]string{"n", "ne", "e", "se", "s", "sw", "w", "nw"}

var reimaginedStateFrames = map[string]int{
	"idle": 1, "walk": 4, "attack": 3, "hurt": 2, "down": 1, "interact": 2,
}

type ReimaginedComplete struct {
	Map        *FaithfulMap
	Scenes     []image.Image
	Title      image.Image
	Ending     image.Image
	Characters map[string]image.Image
	Weapons    map[string]image.Image
}

func LoadReimaginedComplete(bundle *Bundle, tileset, tileCount int) (*ReimaginedComplete, error) {
	if bundle == nil || bundle.Manifest.Mode != ModeReimagined {
		return nil, fmt.Errorf("需要已驗證的 reimagined bundle")
	}
	load := func(id string) (image.Image, error) {
		a, ok := bundle.Asset(id)
		if !ok {
			return nil, fmt.Errorf("bundle 缺少資產 %s", id)
		}
		f, err := os.Open(filepath.Join(bundle.Dir, a.Path))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return png.Decode(f)
	}

	set := &FaithfulMap{}
	for i := 0; i < tileCount; i++ {
		im, err := load(fmt.Sprintf("tile.%d.%03d", tileset, i))
		if err != nil {
			return nil, fmt.Errorf("reimagined tileset %d：%w", tileset, err)
		}
		set.Tiles = append(set.Tiles, im)
	}
	for i := 0; i < 10; i++ {
		im, err := load(fmt.Sprintf("icon.%03d", i))
		if err != nil {
			return nil, fmt.Errorf("reimagined icons：%w", err)
		}
		set.Icons = append(set.Icons, im)
	}

	out := &ReimaginedComplete{Map: set, Characters: map[string]image.Image{}, Weapons: map[string]image.Image{}}
	for i := 0; i < 82; i++ {
		im, err := load(fmt.Sprintf("scene.%03d", i))
		if err != nil {
			return nil, fmt.Errorf("reimagined scenes：%w", err)
		}
		out.Scenes = append(out.Scenes, im)
	}
	var err error
	if out.Title, err = load("fullscreen.title"); err != nil {
		return nil, fmt.Errorf("reimagined title：%w", err)
	}
	if out.Ending, err = load("fullscreen.ending"); err != nil {
		return nil, fmt.Errorf("reimagined ending：%w", err)
	}
	for character := 0; character < 7; character++ {
		for _, direction := range ReimaginedDirections {
			for state, frames := range reimaginedStateFrames {
				for frame := 0; frame < frames; frame++ {
					id := fmt.Sprintf("character.%d.%s.%s.%02d", character, direction, state, frame)
					im, err := load(id)
					if err != nil {
						return nil, fmt.Errorf("reimagined character：%w", err)
					}
					out.Characters[id] = im
				}
			}
		}
	}
	for _, kind := range []string{"rifle", "pistol", "blade", "launcher"} {
		for _, direction := range ReimaginedDirections {
			id := fmt.Sprintf("weapon.%s.%s", kind, direction)
			im, err := load(id)
			if err != nil {
				return nil, fmt.Errorf("reimagined weapon：%w", err)
			}
			out.Weapons[id] = im
		}
	}
	return out, nil
}

func LoadReimaginedBundleMap(bundle *Bundle, tileset, tileCount int) (*FaithfulMap, error) {
	if bundle == nil || bundle.Manifest.Mode != ModeReimagined {
		return nil, fmt.Errorf("需要已驗證的 reimagined bundle")
	}
	load := func(id string) (image.Image, error) {
		a, ok := bundle.Asset(id)
		if !ok {
			return nil, fmt.Errorf("bundle 缺少資產 %s", id)
		}
		f, err := os.Open(filepath.Join(bundle.Dir, a.Path))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return png.Decode(f)
	}
	set := &FaithfulMap{}
	for i := 0; i < tileCount; i++ {
		im, err := load(fmt.Sprintf("tile.%d.%03d", tileset, i))
		if err != nil {
			return nil, fmt.Errorf("reimagined tileset %d：%w", tileset, err)
		}
		set.Tiles = append(set.Tiles, im)
	}
	for i := 0; i < 10; i++ {
		im, err := load(fmt.Sprintf("icon.%03d", i))
		if err != nil {
			return nil, fmt.Errorf("reimagined icons：%w", err)
		}
		set.Icons = append(set.Icons, im)
	}
	return set, nil
}

func (r *ReimaginedComplete) Character(character, direction int, state string, frame int) image.Image {
	if r == nil || character < 0 || character >= 7 || direction < 0 || direction >= len(ReimaginedDirections) {
		return nil
	}
	return r.Characters[fmt.Sprintf("character.%d.%s.%s.%02d", character, ReimaginedDirections[direction], state, frame)]
}

func (r *ReimaginedComplete) Weapon(kind string, direction int) image.Image {
	if r == nil || direction < 0 || direction >= len(ReimaginedDirections) {
		return nil
	}
	return r.Weapons[fmt.Sprintf("weapon.%s.%s", kind, ReimaginedDirections[direction])]
}
