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

	// 框邊上的標籤是按鈕：點下去送它代表的按鍵（`docs/re/126` §3）。
	// **這一條要排在 `charAt` 前面**——標籤畫在框線上，那幾格底下
	// 沒有可點的字，落到 `charAt` 就什麼都不會發生。
	for _, l := range s.boxLabels() {
		if l.Hit(col, row) {
			if l.Key == 0x1B {
				none.Action = input.ActionCancel
				return none, true
			}
			none.Char = l.Key
			return none, true
		}
	}
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

// 游標圖形的編號（`CURS` 的 8 個，`docs/re/112` §5）。
//
// ⚠ **編號的語意是原版的，熱區是重製版的。** 原版拿螢幕座標比兩條 45° 線
// 與一個 16 × 16 的方框（`0x18C62`–`0x18CC3`）；重製版的畫面是 960 × 600、
// 訊息視窗也重排過，所以這裡套的是**同一份狀態表**配重製版自己的熱區
// （`docs/spec/29` §4）。第 7 個原版選不到，這裡也不用。
const (
	CursorDefault = 0 // 預設
	CursorPick    = 1 // 指到可點的字
	CursorUp      = 2 // 地圖視窗的上楔形（原版送 `I`）
	CursorDown    = 3 // 下（`K`）
	CursorLeft    = 4 // 左（`J`）
	CursorRight   = 5 // 右（`L`）
	CursorHere    = 6 // 地圖正中央那一格（原版送 ESC）
)

// cursorDir 把方向換成游標編號。
var cursorDir = map[input.Direction]int{
	input.DirUp:    CursorUp,
	input.DirDown:  CursorDown,
	input.DirLeft:  CursorLeft,
	input.DirRight: CursorRight,
}

// cursorGlyph 回報現在該畫哪一個游標圖形。
//
// 判斷順序與 `translateMouse` 一樣——**兩邊分岔的話畫面上的箭頭會指向
// 點下去不會發生的事**，而那種錯不會有任何測試以外的症狀。
func (s *Scene) cursorGlyph(m input.Mouse) int {
	ox, oy := m.X/render.HiScale, m.Y/render.HiScale
	if s.charAt(ox/render.CharWidth, oy/render.CharHeight) != 0 {
		return CursorPick
	}
	if s.title || s.combat != nil || s.facility != nil || s.ending.active || s.wipe.active {
		return CursorDefault
	}
	if d, ok := viewDirection(ox, oy); ok {
		return cursorDir[d]
	}
	if inPartyTile(ox, oy) {
		return CursorHere
	}
	return CursorDefault
}

// inPartyTile 回答「這個點是不是踩在隊伍自己那一格上」。
//
// 原版是地圖正中央 16 × 16 的方框（`0x18C62`），點下去送 ESC；
// 重製版的隊伍格由 `render.PartyCol`／`PartyRow` 決定，點下去不動
// （`viewDirection` 的最後一條）。
func inPartyTile(ox, oy int) bool {
	tx := (ox - render.ViewX + render.TileSize/2) / render.TileSize
	ty := (oy - render.ViewY + render.TileSize/2) / render.TileSize
	if ox < render.ViewX || oy < render.ViewY ||
		tx < 0 || tx >= render.ViewCols || ty < 0 || ty >= render.ViewRows {
		return false
	}
	return tx == render.PartyCol && ty == render.PartyRow
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
	// ⚠ 走訪改成逐 rune 之後，命中判定只認 ASCII——
	// 中文那幾格點下去沒有等價的按鍵可送（原版的熱鍵表也只放一個 byte）。
	pick := func(c, r int, ch rune) {
		if c == col && r == row && ch > ' ' && ch < 0x80 {
			got = byte(ch)
		}
	}
	switch {
	case row == render.CmdRow:
		if bar := s.uiText("cmd.bar"); len(bar) > 0 {
			eachCell(bar, 0, row, pick)
		} else {
			eachCell(commandBar(), 0, row, pick)
		}
	default:
		// 訊息那一塊**戰鬥時是面板、地圖時是訊息視窗**（`msgRect`）。
		// 這裡不能寫死列號——寫死的話戰鬥中點面板會全部落空。
		rect := s.msgRect()
		if row < rect.Row || row > rect.LastRow() {
			break
		}
		s.walkBody(s.cjk, rect, pick)
		if got == 0 && s.message != "" {
			// 沒有中文正文時那一塊是英文——同一條規則。
			eachMessageCell(s.message, rect, rect.Row, pick)
		}
	}
	return got
}
