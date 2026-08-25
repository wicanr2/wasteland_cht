// Command wl-art-preview renders modern candidate tiles back through the
// original map's graphic indices. Output is WIP evidence, not a parity oracle.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

const (
	cell = 48
	cols = 19
	rows = 9
	outW = 288 * 3
	outH = 128 * 3
)

func main() {
	romDir := flag.String("rom", "workplace/orig/wastland", "玩家自備原版資料目錄")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	tileRoot := flag.String("tiles", "artpacks/faithful-hd/assets", "候選 tile 根目錄")
	mapID := flag.Int("map", 0, "地圖資源 ID")
	x := flag.Int("x", 55, "隊伍 X")
	y := flag.Int("y", 62, "隊伍 Y")
	out := flag.String("out", "workplace/art-preview/faithful-map-0.png", "輸出 PNG")
	flag.Parse()
	if err := run(*romDir, *imagePath, *tileRoot, *mapID, *x, *y, *out); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}

func run(romDir, imagePath, tileRoot string, mapID, partyX, partyY int, out string) error {
	rom, err := assets.Open(romDir)
	if err != nil {
		return err
	}
	if err := rom.LoadImage(imagePath); err != nil {
		return err
	}
	b, err := rom.BlockByID(mapID)
	if err != nil {
		return err
	}
	tiles, err := loadTiles(filepath.Join(tileRoot, fmt.Sprintf("tileset-%d", b.Tileset)))
	if err != nil {
		return err
	}
	originalIcons, err := rom.Icons()
	if err != nil {
		return err
	}
	icons := make([]image.Image, len(originalIcons))
	for i, src := range originalIcons {
		icons[i] = nearest3(src.RGBA())
	}

	dst := image.NewRGBA(image.Rect(0, 0, outW, outH))
	originX, originY := partyX-9, partyY-4
	clip := dst.Bounds()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			mx, my := originX+col, originY+row
			code := b.OutsideGraphic()
			if mx >= 0 && my >= 0 && mx < b.Dim && my < b.Dim {
				code = b.Graphic[my*b.Dim+mx]
			}
			var src image.Image
			if int(code) < len(icons) {
				src = icons[code] // WIP：sprite 尚未換新，刻意沿用原版作醒目缺口
			} else {
				idx := int(code) - 10
				if idx < 0 || idx >= len(tiles) {
					return fmt.Errorf("圖形 %d 對不到候選 tile", code)
				}
				src = tiles[idx]
			}
			x0, y0 := -cell/2+col*cell, -cell/2+row*cell
			r := image.Rect(x0, y0, x0+cell, y0+cell).Intersect(clip)
			if !r.Empty() {
				draw.Draw(dst, r, src, image.Pt(r.Min.X-x0, r.Min.Y-y0), draw.Src)
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, dst); err != nil {
		return err
	}
	fmt.Printf("tileset %d → %s（%d×%d；原版 icon 為 WIP fallback）\n", b.Tileset, out, outW, outH)
	return nil
}

func loadTiles(dir string) ([]image.Image, error) {
	var out []image.Image
	for n := 0; ; n++ {
		p := filepath.Join(dir, fmt.Sprintf("%03d.png", n))
		f, err := os.Open(p)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return nil, err
		}
		im, err := png.Decode(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s：%w", p, err)
		}
		if im.Bounds().Dx() != cell || im.Bounds().Dy() != cell {
			return nil, fmt.Errorf("%s 是 %dx%d，不是 %dx%d", p, im.Bounds().Dx(), im.Bounds().Dy(), cell, cell)
		}
		out = append(out, im)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s 沒有候選 tile", dir)
	}
	return out, nil
}

func nearest3(src image.Image) *image.RGBA {
	d := image.NewRGBA(image.Rect(0, 0, src.Bounds().Dx()*3, src.Bounds().Dy()*3))
	for y := 0; y < d.Bounds().Dy(); y++ {
		for x := 0; x < d.Bounds().Dx(); x++ {
			d.Set(x, y, src.At(x/3, y/3))
		}
	}
	return d
}
