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

// 框邊上的標籤是按鈕：點下去要送出它代表的按鍵（`docs/re/126` §3）。
//
// ⚠ 這一條同時擋「畫了但點不到」：`boxLabels` 是繪製與滑鼠共用的那一支，
// 兩邊各列一份的話會漂成看得到卻沒反應，而那沒有任何症狀。
func TestBoxLabelsAreButtons(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	click := func(col, row int) (input.Input, bool) {
		return s.translateMouse(input.Mouse{
			Left: true,
			X:    (col*render.CharWidth + 1) * render.HiScale,
			Y:    (row*render.CharHeight + 1) * render.HiScale,
		})
	}
	// 地圖畫面：地圖框下緣的 `ROSTER ON`。
	in, ok := click(render.LabelRosterOn.Col+1, render.LabelRosterOn.Row)
	if !ok || in.Char != ' ' {
		t.Errorf("點 ROSTER ON 應該送空白，得到 %+v ok=%v", in, ok)
	}
	// 地圖畫面上沒有 `POOL MONEY`：那一格在地圖視窗裡，點下去是走路不是按 `P`。
	if in, _ := click(render.LabelPoolMoney.Col+1, render.LabelPoolMoney.Row); in.Char == 'P' {
		t.Error("地圖畫面不該點得到 POOL MONEY")
	}
	// 進商店之後才有 `POOL MONEY` 與 `ESC`。
	if err := s.LoadMap(10, 30, 25); err != nil {
		t.Fatalf("載入地圖 10 失敗：%v", err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
		t.Fatalf("走進商店失敗：%v", err)
	}
	if _, err := s.Update(input.Input{Char: 'Y'}); err != nil {
		t.Fatalf("回答 Y 失敗：%v", err)
	}
	if s.facility == nil {
		t.Fatal("沒有進到設施")
	}
	in, ok = click(render.LabelPoolMoney.Col+1, render.LabelPoolMoney.Row)
	if !ok || in.Char != 'P' {
		t.Errorf("點 POOL MONEY 應該送 P，得到 %+v ok=%v", in, ok)
	}
	in, ok = click(render.LabelEsc.Col+1, render.LabelEsc.Row)
	if !ok || in.Action != input.ActionCancel {
		t.Errorf("點 ESC 應該送取消，得到 %+v ok=%v", in, ok)
	}
}
