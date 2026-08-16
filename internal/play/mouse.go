package play

// 滑鼠（重製版自己的路，`docs/spec/29`）。
//
// 原版的滑鼠不產生事件：它把游標座標比對 21 筆熱區表，命中就換成一個按鍵碼
// 再走鍵盤那條路（`docs/re/43`）。**重製版不照抄那一套**——它的每一格都綁在
// 原版的畫面座標與遮罩狀態上，而這裡的畫面已經是 960 × 600、訊息視窗也重排過。
// 功能等價就好（使用者定案 2026-08-16）。
//
// ⚠ **滑鼠不是新的一條輸入路徑**：點擊一律翻成與鍵盤等價的 `input.Input`
// 再送進既有的 `Update`。多一條平行的處置分支，就會有「鍵盤做得到、
// 滑鼠做不到」的漏洞，而那種漏洞每加一個模式就多一個。

import (
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// LoadCursors 載入 `CURS` 的八個游標。載不到就沒有游標圖，滑鼠照樣能點。
func (s *Scene) LoadCursors() error {
	cs, err := s.rom.Cursors()
	if err != nil {
		return err
	}
	s.cursors = cs
	return nil
}

// translateMouse 把一次點擊翻成與鍵盤等價的輸入。
//
// 回傳 false 表示這一下不對應任何動作（點在外框上那類），呼叫端就當它沒發生。
func (s *Scene) translateMouse(m input.Mouse) (input.Input, bool) {
	none := input.Input{Dir: input.DirNone}
	// 右鍵 ＝ 取消／退一層。**任何一層都不會結束遊戲**（規格 27 §1）。
	if m.Right {
		none.Action = input.ActionCancel
		return none, true
	}
	if !m.Left {
		return none, false
	}
	// 高解像素 → 原版座標 → 字元格。分區用格子講，換視窗大小不必重算。
	ox, oy := m.X/render.HiScale, m.Y/render.HiScale
	col, row := ox/render.CharWidth, oy/render.CharHeight

	// 點到哪一格就送那一格上的字元——指令列與清單共用這一條規則，
	// 所以中文版也對（熱鍵字母不跟著翻譯走，那一格本來就是 ASCII）。
	if c := s.charAt(col, row); c != 0 {
		none.Char = c
		return none, true
	}
	// 地圖／圖片視窗：只有真的在地圖模式才當成走路。
	if s.title || s.combat != nil || s.facility != nil || s.ending.active || s.wipe.active {
		return none, false
	}
	if d, ok := viewDirection(ox, oy); ok {
		return input.Input{Dir: d}, true
	}
	return none, false
}

// viewDirection 把地圖視窗裡的一個點換成方向。
//
// 以隊伍那一格為原點，**絕對值大的那個軸決定方向**，相等時取水平。
// 點在隊伍身上不動。這是重製版的決定——原版的地圖視窗根本沒有熱區。
func viewDirection(ox, oy int) (input.Direction, bool) {
	// 地圖是用圖磚畫的（16 px 一格，四邊各裁半格，`docs/re/25`），
	// 所以換算要走圖磚不是字元格。
	tx := (ox - render.ViewX + render.TileSize/2) / render.TileSize
	ty := (oy - render.ViewY + render.TileSize/2) / render.TileSize
	if ox < render.ViewX || oy < render.ViewY ||
		tx < 0 || tx >= render.ViewCols || ty < 0 || ty >= render.ViewRows {
		return input.DirNone, false
	}
	dx, dy := tx-render.PartyCol, ty-render.PartyRow
	if dx == 0 && dy == 0 {
		return input.DirNone, false // 點在自己身上
	}
	if abs(dx) >= abs(dy) {
		if dx < 0 {
			return input.DirLeft, true
		}
		return input.DirRight, true
	}
	if dy < 0 {
		return input.DirUp, true
	}
	return input.DirDown, true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// charAt 回報畫面第 (col, row) 格上是哪一個 ASCII 字元；沒有就回 0。
//
// **走的是與繪製同一支逐格走訪**（`eachCell`／`eachMessageCell`）——
// 兩邊各走一次的話遲早會漂成「看到的字」與「點到的字」不一致，
// 而那種錯不會有任何症狀。
func (s *Scene) charAt(col, row int) byte {
	var got byte
	pick := func(c, r int, ascii, _, _ byte) {
		if c == col && r == row && ascii > ' ' {
			got = ascii
		}
	}
	switch {
	case row == render.CmdRow:
		if bar := s.uiText("cmd.bar"); len(bar) > 0 {
			eachCell(bar, 0, row, pick)
		} else {
			eachCell([]byte(commandBar()), 0, row, pick)
		}
	case row >= render.MsgRow && row <= render.MsgRowEnd:
		start := render.MsgRow
		if s.message != "" || len(s.journalHead) > 0 {
			start++
		}
		eachMessageCell(s.cjk, start, pick)
		if got == 0 && s.message != "" {
			// 沒有中文正文時訊息視窗是英文——同一條規則。
			eachMessageCell([]byte(s.message), render.MsgRow, pick)
		}
	}
	return got
}
