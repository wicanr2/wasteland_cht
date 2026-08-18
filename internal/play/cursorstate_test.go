package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// 游標圖形對應哪個狀態（`docs/re/112` §5）。
//
// 這一道守的是「畫出來的箭頭」與「點下去會發生的事」同一套判斷：
// 每一格都拿 `cursorGlyph` 與 `translateMouse` 對一次，
// **兩邊分岔的話畫面會指向點下去不會發生的事**。
func TestCursorGlyphMatchesClick(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadMap(4, 18, 3); err != nil {
		t.Fatal(err)
	}

	// 地圖視窗裡一格的中心（高解像素）。
	//
	// ⚠ 圖磚的原點在 `ViewX − TileSize/2`（四邊各裁半格，`docs/re/25`），
	// 所以第 tx 格的中心是 `ViewX + tx × TileSize`，**不是再加半格**。
	tile := func(tx, ty int) input.Mouse {
		ox := render.ViewX + tx*render.TileSize
		oy := render.ViewY + ty*render.TileSize
		return input.Mouse{X: ox * render.HiScale, Y: oy * render.HiScale, Left: true}
	}

	cases := []struct {
		name string
		m    input.Mouse
		want int
	}{
		{"隊伍自己那一格", tile(render.PartyCol, render.PartyRow), CursorHere},
		{"隊伍左邊", tile(render.PartyCol-1, render.PartyRow), CursorLeft},
		{"隊伍右邊", tile(render.PartyCol+1, render.PartyRow), CursorRight},
		{"隊伍上面", tile(render.PartyCol, render.PartyRow-1), CursorUp},
		{"隊伍下面", tile(render.PartyCol, render.PartyRow+1), CursorDown},
		{"畫面左上角（什麼都不是）", input.Mouse{X: 0, Y: 0, Left: true}, CursorDefault},
	}
	for _, c := range cases {
		if got := s.cursorGlyph(c.m); got != c.want {
			t.Errorf("%s：游標 %d，應該是 %d", c.name, got, c.want)
		}
	}

	// 指令列上有字的那一格 → 可點的游標，而且點下去真的送得出那個字。
	found := false
	for col := 0; col < render.ScreenWidth/render.CharWidth && !found; col++ {
		ox := col*render.CharWidth + render.CharWidth/2
		oy := render.CmdRow*render.CharHeight + render.CharHeight/2
		m := input.Mouse{X: ox * render.HiScale, Y: oy * render.HiScale, Left: true}
		if s.charAt(col, render.CmdRow) == 0 {
			continue
		}
		found = true
		if got := s.cursorGlyph(m); got != CursorPick {
			t.Errorf("指令列第 %d 格：游標 %d，應該是 %d", col, got, CursorPick)
		}
		in, ok := s.translateMouse(m)
		if !ok || in.Char == 0 {
			t.Errorf("指令列第 %d 格畫成可點，點下去卻沒有字元", col)
		}
	}
	if !found {
		t.Fatal("指令列一格字都沒有——這時候上面那一條證明不了任何事")
	}

	// 標題畫面沒有地圖視窗：同一個點要退回預設，不能還畫方向箭頭。
	s.BeginTitle()
	if got := s.cursorGlyph(tile(render.PartyCol-1, render.PartyRow)); got != CursorDefault {
		t.Errorf("標題畫面：游標 %d，應該是 %d", got, CursorDefault)
	}
}

// 原版選不到第 8 個圖形（`docs/re/112` §3：十個寫入點沒有一處寫 7），
// 重製版也不該用它。
func TestCursorSevenUnused(t *testing.T) {
	for _, g := range cursorDir {
		if g > CursorHere {
			t.Errorf("方向對到游標 %d，超出原版用得到的 0–6", g)
		}
	}
}
