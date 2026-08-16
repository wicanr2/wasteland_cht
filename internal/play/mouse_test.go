package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// cellPix 把字元格換成高解畫布的像素（格子中央）。
func cellPix(col, row int) (int, int) {
	return col*render.HiCellWidth + render.HiCellWidth/2,
		row*render.HiCellHeight + render.HiCellHeight/2
}

// 驗收 1（`docs/spec/29`）：點地圖會走，點自己身上不動。
func TestMouseWalksOnMap(t *testing.T) {
	s := newScene(t)
	w := s.World()
	x0, y0 := w.Party.X, w.Party.Y

	// 隊伍固定在視窗的第 (9,4) 個圖磚；點它右邊那一格。
	px := render.ViewX + (render.PartyCol+2)*render.TileSize
	py := render.ViewY + render.PartyRow*render.TileSize
	step(t, s, input.Input{Dir: input.DirNone, Mouse: input.Mouse{
		X: px * render.HiScale, Y: py * render.HiScale, Left: true}})
	if w.Party.X == x0 && w.Party.Y == y0 {
		t.Fatalf("點地圖右邊沒有走（還在 %d,%d）", x0, y0)
	}
	if w.Party.X != x0+1 {
		t.Errorf("點右邊走到 X=%d，預期 %d", w.Party.X, x0+1)
	}

	// 點在自己身上不動。
	x1, y1 := w.Party.X, w.Party.Y
	px = render.ViewX + render.PartyCol*render.TileSize
	py = render.ViewY + render.PartyRow*render.TileSize
	step(t, s, input.Input{Dir: input.DirNone, Mouse: input.Mouse{
		X: px * render.HiScale, Y: py * render.HiScale, Left: true}})
	if w.Party.X != x1 || w.Party.Y != y1 {
		t.Errorf("點在自己身上卻動了：(%d,%d) → (%d,%d)", x1, y1, w.Party.X, w.Party.Y)
	}
}

// 驗收 2＋5：點指令列等同按那個字母——**同一個動作用鍵盤與滑鼠結果相同**。
func TestMouseCommandBarMatchesKeyboard(t *testing.T) {
	byKey := sceneWithCatalogue(t)
	step(t, byKey, input.Input{Dir: input.DirNone, Char: 'U'})
	want := byKey.Mode()
	if want == "map" {
		t.Fatal("按 U 沒有進 USE，這一條沒驗到東西")
	}

	byMouse := sceneWithCatalogue(t)
	// 找出指令列上 `U` 在第幾格——**問的是實際畫出來的那一行**。
	col := -1
	for c := 0; c < render.MsgWidth; c++ {
		if byMouse.charAt(c, render.CmdRow) == 'U' {
			col = c
			break
		}
	}
	if col < 0 {
		t.Fatal("指令列上找不到 U")
	}
	x, y := cellPix(col, render.CmdRow)
	step(t, byMouse, input.Input{Dir: input.DirNone,
		Mouse: input.Mouse{X: x, Y: y, Left: true}})
	if got := byMouse.Mode(); got != want {
		t.Errorf("點指令列的 U 之後在 %s，鍵盤按 U 之後在 %s", got, want)
	}
}

// 驗收 4：右鍵是取消，**任何一層都不會結束遊戲**。
func TestMouseRightClickCancelsNeverQuits(t *testing.T) {
	s := sceneWithCatalogue(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'U'})
	if s.Mode() == "map" {
		t.Fatal("按 U 沒有進 USE")
	}
	ok, err := s.Update(input.Input{Dir: input.DirNone,
		Mouse: input.Mouse{X: 100, Y: 100, Right: true}})
	if err != nil || !ok {
		t.Fatalf("右鍵結束了遊戲：ok=%v err=%v", ok, err)
	}
	if s.Mode() != "map" {
		t.Errorf("右鍵之後停在 %s，預期回到 map", s.Mode())
	}
}

// 驗收 3：清單開著時點那一項的數字 ＝ 按那個數字。
func TestMousePicksListItem(t *testing.T) {
	byKey := sceneWithCatalogue(t)
	step(t, byKey, input.Input{Dir: input.DirNone, Char: 'U'})
	step(t, byKey, input.Input{Dir: input.DirNone, Char: '1'})
	want := byKey.Mode()

	byMouse := sceneWithCatalogue(t)
	step(t, byMouse, input.Input{Dir: input.DirNone, Char: 'U'})
	col, row := -1, -1
	for r := render.MsgRow; r <= render.MsgRowEnd && col < 0; r++ {
		for c := 0; c < render.MsgWidth+render.MsgCol; c++ {
			if byMouse.charAt(c, r) == '1' {
				col, row = c, r
				break
			}
		}
	}
	if col < 0 {
		t.Skip("這一層清單裡沒有數字 1（挑人那層可能只有一個人）")
	}
	x, y := cellPix(col, row)
	step(t, byMouse, input.Input{Dir: input.DirNone,
		Mouse: input.Mouse{X: x, Y: y, Left: true}})
	if got := byMouse.Mode(); got != want {
		t.Errorf("點清單的 1 之後在 %s，鍵盤按 1 之後在 %s", got, want)
	}
}
