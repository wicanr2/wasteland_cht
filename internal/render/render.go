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
	// DefaultTextColor 是單色字型畫出來的顏色。
	//
	// **重製決策**（使用者定案 2026-08-15，不追原版）：EGA 15 ＝ 亮白，
	// 在 mode 0Dh 的 16 色裡與訊息視窗底色對比最高。
	// 原版取哪個顏色沒有查證，也不打算查——這是重製版自己決定的事
	// （`docs/re/00-wiring-status.md`「重製版自訂的地方」）。
	DefaultTextColor = 15
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

// MapClip 是**畫地圖**用的裁剪：比 ViewClip 少最上面那一列。
//
// ⚠ 這一列是**實機對拍抓出來的**（docs/re/47 §5）：原版地圖視窗的
// `y = 8` 那一列留黑，內容從 `y = 9` 開始；第 1 列起與我們逐像素相同。
// 圖片視窗**不是**這樣——`TITLE.PIC` 滿滿 128 列在 (8, 8) 100% 吻合，
// 所以這是地圖繪製專屬的差別，不要順手套到 DrawPicture 上。
func MapClip() Clip { return Clip{ViewX, ViewY + 1, ViewWidth, ViewHeight - 1} }

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
	Masks [][]bool // MASKS.WLF：與 Icons 一一對應，1 ＝ 保留背景
	Tiles []*assets.Indexed

	// maskScratch 是原版 `seg003:0xDF60` 那 32 bytes（＝ 遮罩表的第 40 格）。
	// 腳本 opcode 2 把某一張遮罩與它對調，**映像裡那一格是全 0**，
	// 所以換上去等於「不透明」——正好是變裝要的效果（`docs/re/104` §2）。
	maskScratch []bool
}

// at 取得圖形編號在**連續那張表**上的位置（`seg003:0x420`，`docs/re/24` §2.3）。
func (g *Graphics) at(n byte) (**assets.Indexed, error) {
	if int(n) < len(g.Icons) {
		return &g.Icons[n], nil
	}
	i := int(n) - len(g.Icons)
	if i < len(g.Tiles) {
		return &g.Tiles[i], nil
	}
	return nil, fmt.Errorf("圖形編號 %d 超出範圍（%d 疊圖 ＋ %d 圖磚）",
		n, len(g.Icons), len(g.Tiles))
}

// Swap 對調兩張圖形，並把其中編號 < 10 的那一張的遮罩換成暫存格
//（腳本 opcode 2 ＝ overlay slot 18，`docs/re/104`）。
//
// ⚠ **這是一個開關不是單向操作**：原版做的是 `xchg`，同一對編號再呼叫一次
// 就換回來。出貨資料的六筆正好是三對「換過去／換回來」。
//
// ⚠ 遮罩那一半只挑**一張**：`a < 10` 就用 a，否則 `b < 10` 才用 b，
// 兩個都 ≥ 10 就完全不動遮罩（`0x10C88`–`0x10C91`）。
func (g *Graphics) Swap(a, b byte) error {
	pa, err := g.at(a)
	if err != nil {
		return err
	}
	pb, err := g.at(b)
	if err != nil {
		return err
	}
	*pa, *pb = *pb, *pa

	n := -1
	switch {
	case int(a) < len(g.Icons):
		n = int(a)
	case int(b) < len(g.Icons):
		n = int(b)
	}
	if n < 0 || n >= len(g.Masks) {
		return nil
	}
	if g.maskScratch == nil {
		// 映像裡 `seg003:0xDF60` 那 32 bytes 全是 0 ＝ 一格背景都不留。
		g.maskScratch = make([]bool, len(g.Masks[n]))
	}
	g.Masks[n], g.maskScratch = g.maskScratch, g.Masks[n]
	return nil
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
	clip := MapClip()
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

// DrawOverlay 把一張疊圖用原版的規則合成上去：
//
//	螢幕 ← (背景 AND 遮罩) OR 疊圖
//
// ⚠ 不是「0 當透明」。遮罩是獨立的一張圖，位元 0 的地方**先把背景清成 0**
// 再 OR，所以疊圖的 0 也可能是有意義的黑。
func (f *Frame) DrawOverlay(im *assets.Indexed, mask []bool, x, y int, clip Clip) {
	for row := 0; row < im.Height; row++ {
		for col := 0; col < im.Width; col++ {
			px, py := x+col, y+row
			if !clip.contains(px, py) {
				continue
			}
			bg := f.At(px, py)
			if i := row*im.Width + col; i < len(mask) && !mask[i] {
				bg = 0
			}
			f.Set(px, py, bg|im.Pix[row*im.Width+col])
		}
	}
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
// color 是單色字型要用的顏色（預設 DefaultTextColor，重製決策）；
// 彩色字型的字模自帶顏色，這時 color 傳 0 表示照字模的值畫。
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

// 戰鬥的指令／訊息面板（`docs/re/105` §2）。
//
// 原版戰鬥時**不用訊息視窗**：`sub_19727` 把文字區設成
// 欄 15–38、列 1–13（左邊那一塊放肖像），所以「名字, choose:」加七個選項
// 一共八行綽綽有餘。訊息視窗那 6 列在戰鬥時是名單的一部分。
const (
	PanelCol   = 15
	PanelRow   = 1
	PanelWidth = 24 // 欄 15–38
	// PanelHeight ＝ **12 不是 13**：選單框的下緣在列 13（`sub_19727` 傳
	// `bl ＝ 0x0D`），`POOL MONEY` 就印在那條線上（`docs/re/126`、`docs/re/127`）。
	// 多算一列的話最後一行字會壓在框線與那個標籤上。
	PanelHeight = 12 // 列 1–12
)

// DrawText 把排好的行畫進訊息視窗（欄 1–38、字元列 18–23）。
// 超過 MsgHeight 行的部分不畫——分頁是呼叫端的事（textlayout.Paginate）。
func (f *Frame) DrawText(font *assets.Font, lines []textlayout.Line) error {
	return f.DrawTextIn(font, lines, MsgCol, MsgRow, MsgWidth, MsgHeight)
}

// DrawTextIn 是 DrawText 的任意矩形版本（戰鬥面板要用）。
func (f *Frame) DrawTextIn(font *assets.Font, lines []textlayout.Line,
	atCol, atRow, w, h int) error {
	// 行數超過區域高度時**丟掉最前面的行**，不是切掉後面的
	// （`docs/re/106` §1：原版是把整塊往上捲，最後一行一定看得到）。
	if len(lines) > h {
		lines = lines[len(lines)-h:]
	}
	for row, line := range lines {
		if row >= h {
			break
		}
		for col, cell := range line.Cells {
			if col >= w {
				break
			}
			if cell.Char < 0x20 {
				continue // 控制碼不該走到這裡；保險起見不畫
			}
			idx := int(cell.Char) - int(font.FirstASCII)
			if err := f.DrawGlyph(font, idx, atCol+col, atRow+row,
				DefaultTextColor, cell.Inverse); err != nil {
				return fmt.Errorf("欄 %d 列 %d：%w", col, row, err)
			}
		}
	}
	return nil
}

// CmdRow 是地圖指令列的字元列。
//
// **重製決策**：原版的指令列字串已確認（`docs/re/72` §4、`docs/re/91`），
// 但它印在哪一列還沒從程式碼讀出來。訊息視窗占字元列 18–23，
// 螢幕共 25 列，所以放在最後一列 24——不蓋到任何已解的區域。
const CmdRow = 24

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

// RGBA 用 assets 的 EGA 調色盤上色（mode 0Dh 預設 16 色，docs/re/23 §7），
// 給 internal/ui 送上螢幕。
func (f *Frame) RGBA() []byte {
	out := make([]byte, len(f.Pix)*4)
	for i, v := range f.Pix {
		c := assets.EGAPalette[v&0x0F]
		out[i*4+0], out[i*4+1], out[i*4+2], out[i*4+3] = c.R, c.G, c.B, c.A
	}
	return out
}

// RosterHeaderRow 是隊伍名單的表頭字元列（`docs/re/40` §1、`docs/re/125`）。
//
// ⚠ **是 15 不是 14。** 原版的名單框從列 14 起，而列 14 與列 15 合起來是
// 一條雙倍高的橫幅；成員從列 16 起（`sub_1738A` 的「列 ＝ 序號 ＋ 0x0F」）。
// 少算一列的話整份名單往上移一格，而畫面上只是「名單貼著上面那一塊」。
const RosterHeaderRow = 15

// 設施畫面的版面（實機對拍量出來的，docs/re/54）。
//
// 圖在視窗原點 (8, 8)、96 × 84，**地點名在圖的正下方**（字元列 12），
// 不是名單那一列。
const (
	FacilityPicX, FacilityPicY = ViewX, ViewY
	FacilityNameRow            = 12
	FacilityNameCol            = 1

	// ALLPICS 的圖是 96 × 84，**只佔視窗左邊**——右邊的地圖照常露出來
	// （docs/re/54 §2）。不要拿 ViewWidth／ViewHeight 當它的尺寸。
	FacilityPicWidth  = 96
	FacilityPicHeight = 84
)

// DrawLineAt 在指定的字元格畫一行純 ASCII。
//
// 與 DrawText 的分工：DrawText 走排版器、畫在訊息視窗；這一支不排版，
// 呼叫端說畫在哪就畫在哪——隊伍名單與設施的地點名都是「已經排好的一行」。
//
// ⚠ 超出畫面的字**直接不畫**，不要回錯誤：名單的欄座標是原版定死的，
// 中文化重排時字會變長，那時候需要的是截斷不是崩掉。
func (f *Frame) DrawLineAt(font *assets.Font, s string, col, row int) error {
	return f.DrawLineInverse(font, s, col, row, nil)
}

// DrawLineInverse 是 DrawLineAt 的反白版：`inv` 回答「第幾欄要反白」。
//
// 反白是原版的「這一格有問題」標記（`ds:4678h`，`docs/re/111`）——
// 卡彈的武器名、身上有狀態的隊員。`inv` 是 nil 就整行正常畫。
func (f *Frame) DrawLineInverse(font *assets.Font, s string, col, row int,
	inv func(col int) bool) error {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			continue
		}
		if col+i >= ScreenWidth/CharWidth {
			break
		}
		idx := int(s[i]) - int(font.FirstASCII)
		bad := inv != nil && inv(col+i)
		if err := f.DrawGlyph(font, idx, col+i, row, DefaultTextColor, bad); err != nil {
			return fmt.Errorf("欄 %d 列 %d：%w", col+i, row, err)
		}
	}
	return nil
}
