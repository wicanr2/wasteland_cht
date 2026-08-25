package artpack

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// FaithfulMap is one validated 48px map-graphics set. It is a vertical-slice
// type until every tileset and scene is represented by a complete Bundle.
type FaithfulMap struct {
	Tiles     []image.Image
	Icons     []image.Image
	PartyWalk []image.Image // up, right, down, left；每方向 3 格
}

// FaithfulComplete is the decoded presentation subset needed by one live
// play.Scene. Bundle validation has already checked every file in the pack;
// this type keeps only the current tileset plus globally shared pictures.
type FaithfulComplete struct {
	Map    *FaithfulMap
	Scenes []image.Image
	Title  image.Image
	Ending image.Image
}

// LoadFaithfulComplete prepares every decoded image used by a live scene.
// Callers must commit the returned pointer only after this function succeeds.
func LoadFaithfulComplete(bundle *Bundle, tileset, tileCount int) (*FaithfulComplete, error) {
	if bundle == nil || bundle.Manifest.Mode != ModeFaithfulHD {
		return nil, fmt.Errorf("需要已驗證的 faithful-hd bundle")
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
		im, err := png.Decode(f)
		if err != nil {
			return nil, err
		}
		return im, nil
	}

	set, err := loadFaithfulMap(bundle, load, tileset, tileCount)
	if err != nil {
		return nil, err
	}
	complete := &FaithfulComplete{Map: set}
	for i := 0; i < 82; i++ {
		im, err := load(fmt.Sprintf("scene.%03d", i))
		if err != nil {
			return nil, fmt.Errorf("faithful scenes：%w", err)
		}
		complete.Scenes = append(complete.Scenes, im)
	}
	if complete.Title, err = load("fullscreen.title"); err != nil {
		return nil, fmt.Errorf("faithful title：%w", err)
	}
	if complete.Ending, err = load("fullscreen.ending"); err != nil {
		return nil, fmt.Errorf("faithful ending：%w", err)
	}
	return complete, nil
}

// LoadFaithfulBundleMap 解碼換圖後所需的 tileset；Bundle 已在初次切換時
// 完整驗過雜湊與尺寸，因此這裡只準備下一張地圖並由 Scene 原子提交。
func LoadFaithfulBundleMap(bundle *Bundle, tileset, tileCount int) (*FaithfulMap, error) {
	if bundle == nil || bundle.Manifest.Mode != ModeFaithfulHD {
		return nil, fmt.Errorf("需要已驗證的 faithful-hd bundle")
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
	return loadFaithfulMap(bundle, load, tileset, tileCount)
}

func loadFaithfulMap(_ *Bundle, load func(string) (image.Image, error), tileset, tileCount int) (*FaithfulMap, error) {
	set := &FaithfulMap{}
	for i := 0; i < tileCount; i++ {
		im, err := load(fmt.Sprintf("tile.%d.%03d", tileset, i))
		if err != nil {
			return nil, fmt.Errorf("faithful tileset %d：%w", tileset, err)
		}
		set.Tiles = append(set.Tiles, im)
	}
	for i := 0; i < 10; i++ {
		im, err := load(fmt.Sprintf("icon.%03d", i))
		if err != nil {
			return nil, fmt.Errorf("faithful icons：%w", err)
		}
		set.Icons = append(set.Icons, im)
	}
	for i := 0; i < 12; i++ {
		im, err := load(fmt.Sprintf("party.walk.%03d", i))
		if err != nil {
			return nil, fmt.Errorf("faithful party walk：%w", err)
		}
		set.PartyWalk = append(set.PartyWalk, im)
	}
	return set, nil
}

func LoadFaithfulMap(root string, tileset, tileCount int) (*FaithfulMap, error) {
	tiles, err := loadIndexedPNGs(filepath.Join(root, fmt.Sprintf("tileset-%d", tileset)), tileCount)
	if err != nil {
		return nil, fmt.Errorf("faithful tileset %d：%w", tileset, err)
	}
	icons, err := loadIndexedPNGs(filepath.Join(root, "icons"), 10)
	if err != nil {
		return nil, fmt.Errorf("faithful icons：%w", err)
	}
	party, err := loadIndexedPNGs(filepath.Join(root, "party-walk"), 12)
	if err != nil {
		return nil, fmt.Errorf("faithful party walk：%w", err)
	}
	return &FaithfulMap{Tiles: tiles, Icons: icons, PartyWalk: party}, nil
}

func loadIndexedPNGs(dir string, count int) ([]image.Image, error) {
	out := make([]image.Image, 0, count)
	for i := 0; i < count; i++ {
		p := filepath.Join(dir, fmt.Sprintf("%03d.png", i))
		f, err := os.Open(p)
		if err != nil {
			return nil, fmt.Errorf("%03d.png：%w", i, err)
		}
		im, err := png.Decode(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%03d.png：%w", i, err)
		}
		if im.Bounds().Dx() != 48 || im.Bounds().Dy() != 48 {
			return nil, fmt.Errorf("%03d.png 是 %d×%d，不是 48×48", i, im.Bounds().Dx(), im.Bounds().Dy())
		}
		out = append(out, im)
	}
	return out, nil
}
