package play

// 片頭（`sub_161C0` 的後半，`docs/re/113`）。
//
// 原版的樣子：主選單按過非 `S` 的鍵之後，六頁開場字幕無限循環，
// 每一頁先清畫面再印第 0 張字串表的幾條，停 255 個計時器刻換下一頁。
//
// 與原版差在三處，逐條寫在這裡而不是散在程式碼裡：
//
//	進入      原版要按**兩次**非 `S` 鍵（`0x16207` 與 `0x1620A` 各等一次），
//	          這裡按一次就進去——第二次等待在畫面上沒有任何提示，看起來像沒反應。
//	行間停頓  原版第 4 頁用 `sub_18DB4`，那是跟著 CPU 速度跑的空轉迴圈，
//	          在現代機器上等於 0（三行同時出現）。這裡改成固定 18 刻。
//	離開      原版沒有出口（`0x162C4` 無條件跳回頭），只有 `S` 會進遊戲。
//	          這裡一樣：`S` 進遊戲，其他鍵不中斷。
//
// ⚠ **每頁 255 刻是原版的值**，不是挑順眼的。`sub_162CF(3)` 的那個 `3` 沒有
// 進等待迴圈（它被存進 `ds:A6E3h`），真正傳下去的是寫死的 `0xFF`。

import (
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// AttractTable 是開場字幕在 `ExeStrings()` 裡的表號（`ds:A703h`）。
const AttractTable = 0

// AttractPages 是原版依序播的六頁，每一頁列出它印的字串槽
// （`0x1623F`–`0x162C4` 逐行讀出來的）。
//
// ⚠ 槽 8（` Computer defense initiative activated.`）**不在裡面**——
// 原版 15 個呼叫端沒有一個傳 8（`docs/re/113` §3.1）。它是寫了沒接上的一句，
// 這裡照原版不播；要播的話那是改設計，不是修 bug。
var AttractPages = [][]int{
	{1},
	{2},
	{3},
	{5, 6, 7},
	{9, 10, 11, 12},
	{13, 14, 15, 16, 17},
}

// attractTypeIn 是「一條一條浮出來」的那一頁（第 4 頁，索引 3）。
const attractTypeIn = 3

const (
	// attractPageTicks 是每一頁停留的計時器刻數（`sub_162E1` 收到的 `0xFF`）。
	attractPageTicks = 0xFF
	// attractLineTicks 是打字頁的行間停頓，重製版自己的值（約 1 秒）。
	attractLineTicks = 18
	// attractRow 是文字從第幾列起畫。原版清完畫面從最上面印。
	attractRow = 1
)

// attractState 是片頭播放中的狀態。
type attractState struct {
	active bool
	page   int // 播到第幾頁
	tick   int // 這一頁播了幾刻
	sub    int // 幀數餘數：Ebiten 60 fps → 計時器的 18.2 Hz
}

// Attract 讓外部（測試、截圖）知道現在是不是在播片頭。
func (s *Scene) Attract() bool { return s.attract.active }

// AttractPage 是現在播到第幾頁（0 起算）。
func (s *Scene) AttractPage() int { return s.attract.page }

// beginAttract 從第一頁開始播。
func (s *Scene) beginAttract() {
	s.attract = attractState{active: true}
	s.dirty = true
}

// tickAttract 推一個計時器刻，該換頁就換。
func (s *Scene) tickAttract() {
	a := &s.attract
	a.tick++
	if a.tick < attractPageTicks {
		// 打字頁每 attractLineTicks 多一行，畫面要跟著更新。
		if a.page == attractTypeIn && a.tick%attractLineTicks == 0 {
			s.dirty = true
		}
		return
	}
	a.tick = 0
	a.page = (a.page + 1) % len(AttractPages)
	s.dirty = true
}

// updateAttract 是片頭播放中的每一幀。回傳 false 表示這一幀沒被片頭吃掉。
//
// ⚠ **計時走畫面更新不等按鍵**：計時器一刻是 1/18.2 秒、Ebiten 一幀是 1/60，
// 每 3 幀推一刻（與結局同一條，`docs/re/96` §4）。
func (s *Scene) updateAttract() {
	s.attract.sub++
	if s.attract.sub >= 3 {
		s.attract.sub = 0
		s.tickAttract()
	}
}

// attractSlots 是這一頁**現在**該顯示的字串槽。
//
// 打字頁一開始只有第一條，每 attractLineTicks 多一條；其餘的頁一次全出。
func (s *Scene) attractSlots() []int {
	if s.attract.page >= len(AttractPages) {
		return nil
	}
	slots := AttractPages[s.attract.page]
	if s.attract.page != attractTypeIn {
		return slots
	}
	n := s.attract.tick/attractLineTicks + 1
	if n > len(slots) {
		n = len(slots)
	}
	return slots[:n]
}

// attractLines 把該顯示的字串攤成一行一行。
//
// `zh` 為 true 時走翻譯目錄；**任何一條查不到就整頁回 nil**——
// 半頁中文半頁英文比整頁英文更難讀，也讓「哪裡沒翻」看不出來
// （與 `hireFailCJK` 同一條規矩）。
func (s *Scene) attractLines(zh bool) []string {
	tables, err := s.rom.ExeStrings()
	if err != nil || AttractTable >= len(tables) {
		return nil
	}
	var out []string
	for _, slot := range s.attractSlots() {
		text := ""
		if zh {
			if s.cat == nil {
				return nil
			}
			t, ok := s.cat.Lookup(lang.ExeKey(AttractTable, slot))
			if !ok {
				return nil
			}
			text = t
		} else {
			if slot >= len(tables[AttractTable]) {
				continue
			}
			text = tables[AttractTable][slot]
		}
		// 這份執行檔的文字用 `\r` 換行，一條字串可以是好幾行。
		out = append(out, strings.Split(text, "\r")...)
	}
	// 尾端的空行不必畫，但**中間的要留**——那是原版排版的一部分。
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	return out
}

// drawAttract 畫片頭那一頁的英文。**中文由 `HiFrame` 那一層畫**：
// 8 × 8 的字模畫不出中文，兩層都畫會留殘影（`hiText` 的註解）。
func (s *Scene) drawAttract(f *render.Frame) {
	if len(s.attractLines(true)) > 0 && s.hiText() {
		return
	}
	for i, line := range s.attractLines(false) {
		if row := attractRow + i; row < render.ScreenHeight/render.CharHeight {
			_ = f.DrawLineAt(s.font, line, 0, row)
		}
	}
}

// drawAttractCJK 畫片頭那一頁的中文。
func (s *Scene) drawAttractCJK(h *render.HiFrame) {
	for i, line := range s.attractLines(true) {
		if row := attractRow + i; row < render.ScreenHeight/render.CharHeight {
			s.drawCJKLine(h, line, 0, row)
		}
	}
}

// attractKey 判斷標題畫面的這一下要不要開始播片頭。
//
// `S`（或確認鍵）進遊戲，其餘任何按鍵開始播——與原版一樣，
// 片頭開始之後除了 `S` 沒有出口。
func attractKey(in input.Input) bool {
	return in.Char != 0 || in.Dir != input.DirNone || in.Action != input.ActionNone
}
