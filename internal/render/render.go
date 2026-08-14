// Package render 把資產合成成一張 320 × 200 的**索引畫面**。
//
// 這一層不認識 Ebiten（docs/spec/03 §2.0）：輸出是每格一個 0–15 的顏色索引，
// 所以視窗幾何、半格裁切、文字排版都能在無頭環境逐格比對。
// 上色與送上螢幕是 `internal/ui` 的事。
package render

import (
	"fmt"
	"image"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// 畫布與視窗（docs/re/25、docs/spec/03 §2.2）。
const (
	ScreenWidth  = 320
	ScreenHeight = 200

	// 地圖／圖片視窗：288 × 128，左上角 (8, 8)。
	ViewX, ViewY = 8, 8
	ViewWidth    = 288
	ViewHeight   = 128

	// 視窗裡是 19 × 9 格 16 × 16 的圖磚，**四邊各半格**。
	ViewCols, ViewRows = 19, 9
	TileSize           = 16

	// 隊伍固定在第 (9, 4) 格（sub_16149 的原點 ＝ 隊伍 − (9, 4)）。
	PartyCol, PartyRow = 9, 4

	// 訊息視窗：字元欄 1–38、字元列 18–23。
	MsgCol, MsgRow   = 1, 18
	MsgRowEnd        = 23 // 訊息視窗最後一列（字元列 18–23，共 6 行）
	MsgWidth         = 38
	MsgHeight        = 6
	CharWidth        = 8
	CharHeight       = 8
	ClockCol         = 28 // 時鐘在外框上緣
	ClockRow         = 0
	DefaultTextColor = 15 // ⚠ 暫代：原版單色字型用什麼顏色畫還沒讀出來
)

// Frame 是一張索引畫面。
type Frame struct {
	Pix []byte // len == ScreenWidth*ScreenHeight，值 0–15
}

// NewFrame 回傳一張全 0 的畫面。
func NewFrame() *Frame {
	return &Frame{Pix: make([]byte, ScreenWidth*ScreenHeight)}
}

// Set 畫一個像素；超出畫布就忽略（裁切是這一層的常態，不是錯誤）。
func (f *Frame) Set(x, y int, v byte) {
	if x < 0 || y < 0 || x >= ScreenWidth || y >= ScreenHeight {
		return
	}
	f.Pix[y*ScreenWidth+x] = v & 0x0F
}

// At 讀一個像素。
func (f *Frame) At(x, y int) byte {
	if x < 0 || y < 0 || x >= ScreenWidth || y >= ScreenHeight {
		return 0
	}
	return f.Pix[y*ScreenWidth+x]
}

// Clip 是一個矩形裁切範圍（像素）。
type Clip struct{ X, Y, W, H int }

// ViewClip 是地圖／圖片視窗的裁切範圍。
func ViewClip() Clip { return Clip{ViewX, ViewY, ViewWidth, ViewHeight} }

func (c Clip) contains(x, y int) bool {
	return x >= c.X && y >= c.Y && x < c.X+c.W && y < c.Y+c.H
}

// DrawIndexed 把一張解碼好的圖畫到 (x, y)，超出 clip 的部分裁掉。
func (f *Frame) DrawIndexed(im *assets.Indexed, x, y int, clip Clip) {
	for row := 0; row < im.Height; row++ {
		for col := 0; col < im.Width; col++ {
			px, py := x+col, y+row
			if !clip.contains(px, py) {
				continue
			}
			f.Set(px, py, im.Pix[row*im.Width+col])
		}
	}
}

// Graphics 是畫地圖要用的圖形集合：編號 0–9 是 IC0_9.WLF 的疊圖，
// ≥10 是該地圖圖磚組的第 (編號 − 10) 張（docs/re/24 §2.3）。
type Graphics struct {
	Icons []*assets.Indexed
	Tiles []*assets.Indexed
}

// Get 依圖形編號取圖。**超出範圍回錯**，不要靜靜畫成空白——
// 那會把資料問題藏起來（docs/spec/01 §3）。
func (g *Graphics) Get(n byte) (*assets.Indexed, error) {
	if int(n) < len(g.Icons) {
		return g.Icons[n], nil
	}
	i := int(n) - len(g.Icons)
	if i < len(g.Tiles) {
		return g.Tiles[i], nil
	}
	return nil, fmt.Errorf("圖形編號 %d 超出範圍（%d 疊圖 ＋ %d 圖磚）",
		n, len(g.Icons), len(g.Tiles))
}

// DrawMap 畫地圖視窗：以 (originX, originY) 為左上格，19 × 9 格。
//
// **四邊各半格**：第一格只畫右半／下半，最後一格只畫左半／上半，
// 所以可見範圍是 8 ＋ 17 × 16 ＋ 8 ＝ 288 寬、8 ＋ 7 × 16 ＋ 8 ＝ 128 高。
// 不照做的話捲動時邊緣會整格跳（docs/spec/03 §2.2）。
func (f *Frame) DrawMap(b *assets.Block, g *Graphics, originX, originY int) error {
	clip := ViewClip()
	for row := 0; row < ViewRows; row++ {
		for col := 0; col < ViewCols; col++ {
			mx, my := originX+col, originY+row
			var code byte
			if mx < 0 || my < 0 || mx >= b.Dim || my >= b.Dim {
				code = b.OutsideGraphic() // 標頭 +0x33
			} else {
				code = b.Graphic[my*b.Dim+mx]
			}
			im, err := g.Get(code)
			if err != nil {
				return fmt.Errorf("(%d, %d)：%w", mx, my, err)
			}
			// 格子的左上角：第 0 格從視窗左緣往左退半格，其餘依序排。
			x := ViewX - TileSize/2 + col*TileSize
			y := ViewY - TileSize/2 + row*TileSize
			f.DrawIndexed(im, x, y, clip)
		}
	}
	return nil
}

// DrawPicture 把一張 288 × 128 的圖畫進同一個視窗。
func (f *Frame) DrawPicture(im *assets.Indexed) error {
	if im.Width != ViewWidth || im.Height != ViewHeight {
		return fmt.Errorf("圖片是 %d × %d，視窗要 %d × %d",
			im.Width, im.Height, ViewWidth, ViewHeight)
	}
	f.DrawIndexed(im, ViewX, ViewY, ViewClip())
	return nil
}

// DrawGlyph 在字元格 (col, row) 畫一個字模。
//
// color 是單色字型要用的顏色（暫代 DefaultTextColor）；彩色字型的字模自帶顏色，
// 這時 color 傳 0 表示照字模的值畫。
func (f *Frame) DrawGlyph(font *assets.Font, index, col, row int, color byte, inverse bool) error {
	g, err := font.Glyph(index)
	if err != nil {
		return err
	}
	x0, y0 := col*CharWidth, row*CharHeight
	for y := 0; y < CharHeight; y++ {
		for x := 0; x < CharWidth; x++ {
			v := g.Pix[y*CharWidth+x]
			if color != 0 { // 單色字型：0/1 → 背景／指定顏色
				if v != 0 {
					v = color
				}
			}
			if inverse {
				if v == 0 {
					v = color
					if v == 0 {
						v = DefaultTextColor
					}
				} else {
					v = 0
				}
			}
			f.Set(x0+x, y0+y, v)
		}
	}
	return nil
}

// DrawText 把排好的行畫進訊息視窗（欄 1–38、字元列 18–23）。
// 超過 MsgHeight 行的部分不畫——分頁是呼叫端的事（textlayout.Paginate）。
func (f *Frame) DrawText(font *assets.Font, lines []textlayout.Line) error {
	for row, line := range lines {
		if row >= MsgHeight {
			break
		}
		for col, cell := range line.Cells {
			if col >= MsgWidth {
				break
			}
			if cell.Char < 0x20 {
				continue // 控制碼不該走到這裡；保險起見不畫
			}
			idx := int(cell.Char) - int(font.FirstASCII)
			if err := f.DrawGlyph(font, idx, MsgCol+col, MsgRow+row,
				DefaultTextColor, cell.Inverse); err != nil {
				return fmt.Errorf("欄 %d 列 %d：%w", col, row, err)
			}
		}
	}
	return nil
}

// DrawClock 畫時鐘（外框上緣，字元欄 28、列 0，兩位補零、24 小時制）。
func (f *Frame) DrawClock(font *assets.Font, hour, minute int) error {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return fmt.Errorf("時間 %d:%d 超出 24 小時制", hour, minute)
	}
	text := fmt.Sprintf("%02d:%02d", hour, minute)
	for i := 0; i < len(text); i++ {
		idx := int(text[i]) - int(font.FirstASCII)
		if err := f.DrawGlyph(font, idx, ClockCol+i, ClockRow, DefaultTextColor, false); err != nil {
			return err
		}
	}
	return nil
}

// ToImage 把一幀轉成標準函式庫的圖——截圖與對拍用，不需要 Ebiten。
func (f *Frame) ToImage() *image.RGBA {
	return &image.RGBA{
		Pix:    f.RGBA(),
		Stride: ScreenWidth * 4,
		Rect:   image.Rect(0, 0, ScreenWidth, ScreenHeight),
	}
}

// RGBA 用 assets 的（暫代）調色盤上色，給 internal/ui 送上螢幕。
func (f *Frame) RGBA() []byte {
	out := make([]byte, len(f.Pix)*4)
	for i, v := range f.Pix {
		c := assets.EGAPalette[v&0x0F]
		out[i*4+0], out[i*4+1], out[i*4+2], out[i*4+3] = c.R, c.G, c.B, c.A
	}
	return out
}
