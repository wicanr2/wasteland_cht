package render

import (
	"image"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 640 × 400 的中文畫布（docs/spec/10 §2）。
//
// 決策：**拉畫布不縮字**。原版 320 × 200 的每個像素放大成 2 × 2（nearest），
// 倚天 16 × 15 的中文字直繪 1:1——一個中文字剛好佔原版一個 8×8 字元格，
// 所以 docs/re/25 的所有座標都不用重算。

const (
	HiScale        = 2
	HiScreenWidth  = ScreenWidth * HiScale  // 640
	HiScreenHeight = ScreenHeight * HiScale // 400

	// HiCellWidth／HiCellHeight 是放大後的字元格（原版 8×8）。
	HiCellWidth  = 8 * HiScale
	HiCellHeight = 8 * HiScale
)

// HiFrame 是 640 × 400 的索引畫面。
type HiFrame struct {
	Pix [HiScreenWidth * HiScreenHeight]byte
}

// NewHiFrame 建一張空的高解畫布。
func NewHiFrame() *HiFrame { return &HiFrame{} }

// Upscale 把 320 × 200 的一幀逐像素放大成 2 × 2。
// **nearest，不做插值**——插值會把 pixel art 糊掉（rulebook/81）。
func (h *HiFrame) Upscale(f *Frame) {
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			v := f.At(x, y)
			bx, by := x*HiScale, y*HiScale
			for dy := 0; dy < HiScale; dy++ {
				row := (by + dy) * HiScreenWidth
				for dx := 0; dx < HiScale; dx++ {
					h.Pix[row+bx+dx] = v
				}
			}
		}
	}
}

// At 讀一個像素。
func (h *HiFrame) At(x, y int) byte {
	if x < 0 || y < 0 || x >= HiScreenWidth || y >= HiScreenHeight {
		return 0
	}
	return h.Pix[y*HiScreenWidth+x]
}

// Set 寫一個像素。
func (h *HiFrame) Set(x, y int, v byte) {
	if x < 0 || y < 0 || x >= HiScreenWidth || y >= HiScreenHeight {
		return
	}
	h.Pix[y*HiScreenWidth+x] = v
}

// DrawCJK 把一個 Big5 字畫在第 (col, row) 個原版字元格上。
//
// 字高 15、格高 16，所以**垂直置中**（上下各留半列，取上緣 0）。
// 回傳 false 表示字型裡沒有這個字——呼叫者要自己決定 fallback，
// 不要靜靜畫成空白（那會變成看不見的缺字）。
func (h *HiFrame) DrawCJK(font *assets.ETenFont, hi, lo byte, col, row int, fg byte) bool {
	rows := font.GlyphRows(hi, lo)
	if rows == nil {
		return false
	}
	x0 := col * HiCellWidth
	y0 := row*HiCellHeight + (HiCellHeight-assets.ETenHeight)/2
	for y, line := range rows {
		for x, on := range line {
			if on {
				h.Set(x0+x, y0+y, fg)
			}
		}
	}
	return true
}

// ToImage 把高解畫面轉成 RGBA（調色盤與 320 × 200 那張共用）。
// DrawASCIIAt 在高解析畫面上畫一個 ASCII 字元（原版 8 × 8 字模放大兩倍）。
//
// 中文譯文裡會夾英文與標點（人名、數字、`%s` 之後接的東西），
// **不能整串當成 Big5 兩兩配對**——錯一個 byte 之後整行都會變亂碼。
func (h *HiFrame) DrawASCIIAt(font *assets.Font, c byte, col, row int, fg byte) bool {
	if font == nil || c < font.FirstASCII {
		return false
	}
	g, err := font.Glyph(int(c) - int(font.FirstASCII))
	if err != nil {
		return false
	}
	x0, y0 := col*HiCellWidth, row*HiCellHeight
	for y := 0; y < CharHeight; y++ {
		for x := 0; x < CharWidth; x++ {
			if g.Pix[y*CharWidth+x] == 0 {
				continue
			}
			for dy := 0; dy < HiScale; dy++ {
				for dx := 0; dx < HiScale; dx++ {
					h.Set(x0+x*HiScale+dx, y0+y*HiScale+dy, fg)
				}
			}
		}
	}
	return true
}

func (h *HiFrame) ToImage() *image.RGBA {
	pix := make([]byte, len(h.Pix)*4)
	for i, v := range h.Pix {
		c := assets.EGAPalette[v&0x0F]
		pix[i*4+0], pix[i*4+1], pix[i*4+2], pix[i*4+3] = c.R, c.G, c.B, c.A
	}
	return &image.RGBA{
		Pix:    pix,
		Stride: HiScreenWidth * 4,
		Rect:   image.Rect(0, 0, HiScreenWidth, HiScreenHeight),
	}
}
