// Command wl-art-inventory records the exact original-art slots consumed by
// the remake. It emits metadata safe for version control and optional contact
// sheets under workplace/ for AI-generation reference only.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

type Inventory struct {
	Schema   int           `json:"schema"`
	Evidence string        `json:"evidence"`
	Tilesets []TilesetInfo `json:"tilesets"`
	Icons    AssetGroup    `json:"icons"`
	Scenes   AssetGroup    `json:"scenes"`
	Title    AssetGroup    `json:"title"`
	Ending   AssetGroup    `json:"ending"`
}

type AssetGroup struct {
	Count  int `json:"count"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type TilesetInfo struct {
	ID          int   `json:"id"`
	Count       int   `json:"count"`
	Used        []int `json:"used_indices"`
	MapIDs      []int `json:"map_resource_ids"`
	OutsideUsed []int `json:"outside_indices"`
}

func main() {
	romDir := flag.String("rom", "workplace/orig/wastland", "玩家自備原版資料目錄")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	out := flag.String("out", "artwork/manifests/original-art-inventory.json", "metadata JSON")
	contacts := flag.String("contacts", "", "原版 contact sheet 輸出目錄（應放 workplace/；空＝不輸出）")
	flag.Parse()
	if err := run(*romDir, *imagePath, *out, *contacts); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}

func run(romDir, imagePath, out, contacts string) error {
	rom, err := assets.Open(romDir)
	if err != nil {
		return err
	}
	if err := rom.LoadImage(imagePath); err != nil {
		return err
	}

	used := make([]map[int]bool, 9)
	maps := make([]map[int]bool, 9)
	outside := make([]map[int]bool, 9)
	for i := range used {
		used[i], maps[i], outside[i] = map[int]bool{}, map[int]bool{}, map[int]bool{}
	}
	res, err := rom.Resources()
	if err != nil {
		return err
	}
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			return fmt.Errorf("地圖區塊 %d：%w", i, err)
		}
		if b.Tileset < 0 || b.Tileset >= len(used) {
			return fmt.Errorf("地圖資源 %d 的 tileset %d 超出 0–8", b.Resource.ID, b.Tileset)
		}
		maps[b.Tileset][b.Resource.ID] = true
		for _, code := range b.Graphic {
			if code >= 10 {
				used[b.Tileset][int(code)-10] = true
			}
		}
		if code := int(b.OutsideGraphic()); code >= 10 {
			idx := code - 10
			used[b.Tileset][idx], outside[b.Tileset][idx] = true, true
		}
	}

	inv := Inventory{Schema: 1, Evidence: "player-supplied DOS data validated by internal/assets KnownFiles SHA-256"}
	for id := 0; id < 9; id++ {
		tiles, err := rom.Tileset(id)
		if err != nil {
			return err
		}
		info := TilesetInfo{ID: id, Count: len(tiles), Used: keys(used[id]), MapIDs: keys(maps[id]), OutsideUsed: keys(outside[id])}
		for _, n := range info.Used {
			if n < 0 || n >= info.Count {
				return fmt.Errorf("tileset %d 使用索引 %d，但只有 %d 張", id, n, info.Count)
			}
		}
		inv.Tilesets = append(inv.Tilesets, info)
		if contacts != "" {
			if err := writeContact(filepath.Join(contacts, fmt.Sprintf("tileset-%d.png", id)), tiles, 16); err != nil {
				return err
			}
			for index, tile := range tiles {
				name := filepath.Join(contacts, "tiles", fmt.Sprintf("tileset-%d", id), fmt.Sprintf("%03d.png", index))
				if err := writePNG(name, tile.RGBA()); err != nil {
					return err
				}
			}
			for start := 0; start < len(tiles); start += 16 {
				end := start + 16
				if end > len(tiles) {
					end = len(tiles)
				}
				name := fmt.Sprintf("tileset-%d-batch-%03d-%03d.png", id, start, end-1)
				if err := writeContact(filepath.Join(contacts, name), tiles[start:end], 4); err != nil {
					return err
				}
			}
		}
	}
	icons, err := rom.Icons()
	if err != nil {
		return err
	}
	inv.Icons = AssetGroup{Count: len(icons), Width: 16, Height: 16}
	p1, err := rom.Pictures("allpics1")
	if err != nil {
		return err
	}
	p2, err := rom.Pictures("allpics2")
	if err != nil {
		return err
	}
	pics := append(p1, p2...)
	inv.Scenes = AssetGroup{Count: len(pics), Width: 96, Height: 84}
	inv.Title = AssetGroup{Count: 1, Width: 288, Height: 128}
	inv.Ending = AssetGroup{Count: 1, Width: 288, Height: 128}
	if contacts != "" {
		title, err := rom.Title()
		if err != nil {
			return err
		}
		ending, err := rom.End()
		if err != nil {
			return err
		}
		if err := writePNG(filepath.Join(contacts, "title.png"), title.RGBA()); err != nil {
			return err
		}
		if err := writePNG(filepath.Join(contacts, "ending.png"), ending.RGBA()); err != nil {
			return err
		}
		if err := writeContact(filepath.Join(contacts, "icons.png"), icons, 10); err != nil {
			return err
		}
		if err := writeContact(filepath.Join(contacts, "icons-5x2.png"), icons, 5); err != nil {
			return err
		}
		if err := writeContact(filepath.Join(contacts, "scenes.png"), pics, 10); err != nil {
			return err
		}
		for index, pic := range pics {
			name := filepath.Join(contacts, "scenes", fmt.Sprintf("%03d.png", index))
			if err := writePNG(name, pic.RGBA()); err != nil {
				return err
			}
		}
	}

	buf, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		return err
	}
	buf = append(buf, '\n')
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, buf, 0o644); err != nil {
		return err
	}
	fmt.Printf("9 組 tileset、%d 張場景 → %s\n", len(pics), out)
	return nil
}

func keys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for n := range m {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

func writeContact(path string, images []*assets.Indexed, cols int) error {
	if len(images) == 0 {
		return fmt.Errorf("%s：沒有圖片", path)
	}
	if cols < 1 {
		cols = 1
	}
	w, h := images[0].Width, images[0].Height
	rows := (len(images) + cols - 1) / cols
	dst := image.NewRGBA(image.Rect(0, 0, cols*w, rows*h))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)
	for i, src := range images {
		if src.Width != w || src.Height != h {
			return fmt.Errorf("%s：第 %d 張尺寸 %d×%d，不是 %d×%d", path, i, src.Width, src.Height, w, h)
		}
		r := image.Rect((i%cols)*w, (i/cols)*h, (i%cols+1)*w, (i/cols+1)*h)
		draw.Draw(dst, r, src.RGBA(), image.Point{}, draw.Src)
	}
	return writePNG(path, dst)
}

func writePNG(path string, im image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, im)
}
