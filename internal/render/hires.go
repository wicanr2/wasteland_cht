package render

import (
	"image"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 960 × 600 的中文畫布（docs/spec/10 §2）。
//
// 決策：**拉畫布不縮字**。原版 320 × 200 的每個像素放大成 3 × 3（nearest），
// 一個原版字元格（8 × 8）因此變成 24 × 24 —— 倚天 24 點的漢字剛好整格填滿，
// 所以 docs/re/25 的所有座標都不用重算。
//
// 為什麼是 3 倍而不是 2 倍：倚天 16 × 15 放在 2 倍畫布上，字只佔畫面的 2.5%，
// 比原版的中文還小，筆劃多的字（鬱／體／覺）在 240 格裡塞不下。
// 24 × 24 有 576 格、而且是**為該尺寸手工調過的**點陣，
// 同樣的畫面比例下細節多一倍（kb `retro-cht/eten-bitmap-font` 的字級取捨表）。
//
// ⚠ **不要把 16 × 15 放大成 24**：非整數倍放大點陣字一定醜。
// 只有 15 點字型時就照原尺寸畫在格子中央，字小但銳利。

const (
	HiScale        = 3
	HiScreenWidth  = ScreenWidth * HiScale  // 960
	HiScreenHeight = ScreenHeight * HiScale // 600

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

// Upscale 把 320 × 200 的一幀逐像素放大成 HiScale × HiScale。
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
	g := font.Glyph(hi, lo)
	if g == nil {
		return false
	}
	return h.blit(g, col, row, font.Width, font.Height, fg)
}

// blit 把一個字模畫在第 (col, row) 個原版字元格上，**在格內置中**。
//
// 字模尺寸與排版格解耦：24 × 24 剛好填滿格子（位移 0），
// 16 × 15 就在 24 × 24 的格子裡居中。換字級不必動任何座標。
//
// ⚠ 收的是**原始點陣**（每列 (w+7)/8 bytes、MSB-first），不是攤平的
// `[][]bool`：畫面上每個字每一幀都會走這裡，而攤平一次要 h+1 次配置
// （`assets.rowsOf`）。位元怎麼排與 `rowsOf` 完全相同——
// `TestBlitMatchesRows` 逐像素比對兩條路。
//
// 點陣不足一個字（字型檔被截斷）就不畫並回 false，**不要畫半個字**。
func (h *HiFrame) blit(g []byte, col, row, w, ht int, fg byte) bool {
	rowBytes := (w + 7) / 8
	if len(g) < rowBytes*ht {
		return false
	}
	x0 := col*HiCellWidth + (HiCellWidth-w)/2
	y0 := row*HiCellHeight + (HiCellHeight-ht)/2
	for y := 0; y < ht; y++ {
		base := y * rowBytes
		for x := 0; x < w; x++ {
			if g[base+x/8]&(0x80>>(x%8)) != 0 {
				h.Set(x0+x, y0+y, fg)
			}
		}
	}
	return true
}

// DrawETenASCII 用倚天自己的半形字模畫一個 ASCII 字元。
//
// **這是英數與中文對齊的關鍵**：倚天的 `ASCFONT.*` 與漢字同高、同一套設計，
// 而遊戲原版的 8 × 8 字模放大三倍之後筆劃有三像素寬，擺在中文旁邊
// 像另一種字型。沒有 `ASCFONT.*` 時回 false，呼叫端退回 DrawASCIIAt。
func (h *HiFrame) DrawETenASCII(font *assets.ETenFont, c byte, col, row int, fg byte) bool {
	g := font.ASCIIGlyph(c)
	if g == nil {
		return false
	}
	return h.blit(g, col, row, font.ASCIIWidth, font.ASCIIHeight, fg)
}

// DrawASCIIAt 在高解析畫面上畫一個 ASCII 字元（原版 8 × 8 字模放大 HiScale 倍）。
//
// 這是**沒有倚天半形字型時的後備**；有的話走 DrawETenASCII。
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

// RGBABytes 是一幀 RGBA 的長度（`WriteRGBA` 的緩衝區要這麼大）。
const RGBABytes = HiScreenWidth * HiScreenHeight * 4

// WriteRGBA 把畫面上色寫進**呼叫端自己的**緩衝區。
//
// 給遊戲迴圈用：`ToImage` 每叫一次配置 2.3 MB，60 fps ＝ 每秒 138 MB 的垃圾。
// 視窗那一層拿同一塊緩衝區重複用（`internal/ui`），配置次數是零。
// 一次性的用途（截圖工具、測試）走 `ToImage` 就好。
//
// 緩衝區不夠大就不寫並回 false——寧可畫面空白，也不要寫出界。
func (h *HiFrame) WriteRGBA(dst []byte) bool {
	if len(dst) < RGBABytes {
		return false
	}
	for i, v := range h.Pix {
		c := assets.EGAPalette[v&0x0F]
		dst[i*4+0], dst[i*4+1], dst[i*4+2], dst[i*4+3] = c.R, c.G, c.B, c.A
	}
	return true
}

// ToImage 把高解畫面轉成 RGBA（調色盤與 320 × 200 那張共用）。
func (h *HiFrame) ToImage() *image.RGBA {
	pix := make([]byte, RGBABytes)
	h.WriteRGBA(pix)
	return &image.RGBA{
		Pix:    pix,
		Stride: HiScreenWidth * 4,
		Rect:   image.Rect(0, 0, HiScreenWidth, HiScreenHeight),
	}
}
