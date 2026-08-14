package render

import (
	"os"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

const (
	romDir    = "../../workplace/orig/wastland"
	imagePath = "../../workplace/analysis/unpacked/wl.merged.exe"
)

func openRom(t *testing.T) *assets.Rom {
	t.Helper()
	if _, err := os.Stat(romDir); err != nil {
		t.Skipf("找不到原版資料目錄 %s，跳過（玩家自備）", romDir)
	}
	rom, err := assets.Open(romDir)
	if err != nil {
		t.Fatalf("Open：%v", err)
	}
	if _, err := os.Stat(imagePath); err != nil {
		t.Skipf("找不到分析映像 %s，跳過", imagePath)
	}
	if err := rom.LoadImage(imagePath); err != nil {
		t.Fatalf("LoadImage：%v", err)
	}
	return rom
}

// 地圖視窗必須剛好蓋滿 288 × 128 @ (8, 8)，一個像素都不能溢出。
// 這是四邊半格裁切的直接後果，也是最容易寫錯的地方。
func TestMapViewportGeometry(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(0)
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := rom.Tileset(block.Tileset)
	if err != nil {
		t.Fatal(err)
	}
	icons, err := rom.Icons()
	if err != nil {
		t.Fatal(err)
	}

	f := NewFrame()
	// 先把整張畫面塗成 7，畫完之後「還是 7」的地方就是沒被碰過的。
	for i := range f.Pix {
		f.Pix[i] = 7
	}
	if err := f.DrawMap(block, &Graphics{Icons: icons, Tiles: tiles}, 20, 20); err != nil {
		t.Fatalf("DrawMap：%v", err)
	}

	minX, minY, maxX, maxY := ScreenWidth, ScreenHeight, -1, -1
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			inside := x >= ViewX && x < ViewX+ViewWidth && y >= ViewY && y < ViewY+ViewHeight
			if !inside && f.At(x, y) != 7 {
				t.Fatalf("(%d, %d) 在視窗外卻被畫到了", x, y)
			}
			if inside && f.At(x, y) != 7 {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if minX != ViewX || minY != ViewY || maxX != ViewX+ViewWidth-1 || maxY != ViewY+ViewHeight-1 {
		t.Fatalf("畫到的範圍是 x %d–%d、y %d–%d，應該是 x %d–%d、y %d–%d",
			minX, maxX, minY, maxY, ViewX, ViewX+ViewWidth-1, ViewY, ViewY+ViewHeight-1)
	}
}

// 捲動一格之後，畫面應該整體位移一個圖磚寬——半格裁切沒做對的話，
// 邊緣會出現整格的跳動（docs/re/26 §4 的補畫兩排就是為了這個）。
func TestScrollShiftsByOneTile(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(1)
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := rom.Tileset(block.Tileset)
	if err != nil {
		t.Fatal(err)
	}
	icons, err := rom.Icons()
	if err != nil {
		t.Fatal(err)
	}
	g := &Graphics{Icons: icons, Tiles: tiles}

	a, b := NewFrame(), NewFrame()
	if err := a.DrawMap(block, g, 10, 10); err != nil {
		t.Fatal(err)
	}
	if err := b.DrawMap(block, g, 11, 10); err != nil { // 往右一格
		t.Fatal(err)
	}
	// b 的某個像素 ＝ a 往右 16 像素的那個像素（重疊區內）
	for y := ViewY; y < ViewY+ViewHeight; y++ {
		for x := ViewX; x < ViewX+ViewWidth-TileSize; x++ {
			if b.At(x, y) != a.At(x+TileSize, y) {
				t.Fatalf("捲動一格後 (%d, %d) 對不上：%d vs %d",
					x, y, b.At(x, y), a.At(x+TileSize, y))
			}
		}
	}
}

// 規格 03 §4 第 3 項：全部字串跑過排版器，控制碼不得被當成文字印出。
func TestAllStringsLayout(t *testing.T) {
	rom := openRom(t)

	var corpus [][]byte
	tables, err := rom.ExeStrings()
	if err != nil {
		t.Fatal(err)
	}
	for _, tb := range tables {
		for _, s := range tb {
			corpus = append(corpus, []byte(s))
		}
	}
	res, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range b.Strings {
			corpus = append(corpus, []byte(s))
		}
	}
	if len(corpus) != 4889 {
		t.Fatalf("語料共 %d 條，應該是 4,889 個字串槽", len(corpus))
	}

	unknown := map[byte]int{}
	for n, text := range corpus {
		out, err := textlayout.Layout(text, textlayout.Options{
			Width: MsgWidth,
			Name:  func() string { return "SNAKE" },
		})
		if err != nil {
			t.Fatalf("第 %d 條排版失敗：%v", n, err)
		}
		for _, line := range out.Lines {
			for _, cell := range line.Cells {
				if cell.Char < 0x20 {
					t.Fatalf("第 %d 條把控制碼 %#x 當成文字印出來了", n, cell.Char)
				}
			}
		}
		for _, e := range out.Events {
			if e.Kind == textlayout.EventUnknownCode {
				unknown[e.Code]++
			}
		}
	}
	// 未解的控制碼要看得到統計，不能靜靜吞掉。
	t.Logf("語料 %d 條，未解控制碼出現次數：%v", len(corpus), unknown)
}

func TestDrawTextAndClock(t *testing.T) {
	rom := openRom(t)
	font, err := rom.FontMain()
	if err != nil {
		t.Fatal(err)
	}
	out, err := textlayout.Layout([]byte("RANGER CENTER"), textlayout.Options{Width: MsgWidth})
	if err != nil {
		t.Fatal(err)
	}
	f := NewFrame()
	if err := f.DrawText(font, out.Lines); err != nil {
		t.Fatalf("DrawText：%v", err)
	}
	// 文字必須落在訊息視窗裡：字元列 18 起、欄 1 起。
	if !anyLit(f, MsgCol*CharWidth, MsgRow*CharHeight, MsgWidth*CharWidth, CharHeight) {
		t.Fatal("訊息視窗第一行是空的")
	}
	if anyLit(f, 0, 0, ScreenWidth, MsgRow*CharHeight) {
		t.Fatal("文字畫到訊息視窗以外了")
	}

	f2 := NewFrame()
	if err := f2.DrawClock(font, 7, 5); err != nil {
		t.Fatalf("DrawClock：%v", err)
	}
	if !anyLit(f2, ClockCol*CharWidth, ClockRow*CharHeight, 5*CharWidth, CharHeight) {
		t.Fatal("時鐘沒畫在外框上緣（欄 28、列 0）")
	}
	if err := f2.DrawClock(font, 24, 0); err == nil {
		t.Fatal("24 時應該被拒絕（24 小時制是 0–23）")
	}
}

func TestGraphicsOutOfRangeIsAnError(t *testing.T) {
	g := &Graphics{Icons: make([]*assets.Indexed, 10), Tiles: make([]*assets.Indexed, 66)}
	if _, err := g.Get(75); err != nil { // 10 疊圖 ＋ 圖磚 65 ＝ 最後一張，還在範圍內
		t.Fatalf("圖形編號 75 應該是最後一張圖磚：%v", err)
	}
	if _, err := g.Get(76); err == nil {
		t.Fatal("圖形編號 76 超出 10+66 卻沒回錯——資料問題會被藏起來")
	}
}

func anyLit(f *Frame, x, y, w, h int) bool {
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			if f.At(col, row) != 0 {
				return true
			}
		}
	}
	return false
}
