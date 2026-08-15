package play

// 結局（`0x1B4F0`，`docs/re/96`）。
//
// 它掛在**設施跳表 `ds:A4E0h` 的第 4 格**——所以觸發方式與商店、醫院一樣，
// 是走進一格「設施」，只是那一種設施沒有店面也沒有圖，進去就結束遊戲。
//
// 原版的順序：
//
// ```
// 逐人把角色記錄 +0x4B 的 bit0 設起來
// sub_1B7FE          ; END.CPA：畫面 ＋ 動畫腳本
// ds:D168h ← 0；sub_1B735(dx ＝ 3Ch)   ; 播 60 tick，**按鍵不中斷**
// sub_1142B；sub_1B735(dx ＝ 3Ch)
// ds:D168h ← 1                          ; 之後按鍵可以跳過
// 四次：sub_162C7；sub_1B7B7(al ＝ 1…4)；sub_1B735(dx ＝ 96h)
// ```
//
// `sub_1B7B7` 把字串表基址換成 `ds:D173h`（＝ `ds:D18Eh` 那張表，
// `ExeStrings()` 的第 **4** 張「結局敘述」）再印第 `al` 條。

import (
	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// EndingTable 是結局敘述在 `ExeStrings()` 裡的表號（`ds:D18Eh`）。
const EndingTable = 4

// EndingPages 是原版依序印的四條（`sub_1B7B7(1…4)`）。
var EndingPages = []int{1, 2, 3, 4}

// endingIntroTicks 是進場那兩段**不能用按鍵跳過**的長度（原版兩次 `dx ＝ 3Ch`）。
const endingIntroTicks = 60 * 2

// endingPageTicks 是每一段文字停留的 tick 數（原版 `dx ＝ 96h`）。
const endingPageTicks = 0x96

// endingState 是結局播放中的狀態。
type endingState struct {
	active bool
	pic    *assets.Indexed
	anim   *assets.EndAnimation

	frame int // 動畫播到第幾格
	wait  int // 這一格還要等幾個 tick
	tick  int // 從進場算起的 tick 數
	page  int // 已經印到第幾段（0 ＝ 還沒開始印）
	sub   int // 幀數餘數：Ebiten 60 fps → BIOS 的 18.2 Hz
}

// Ending 讓外部（測試、截圖）知道現在是不是在結局。
func (s *Scene) Ending() bool { return s.ending.active }

// BeginEnding 進結局。
//
// 載不到 `END.CPA` 就**不中斷遊戲**——結局照樣結束，只是沒有畫面。
// 這一段是全遊戲最後一個畫面，讓它變成「開不起來」是最糟的失敗方式。
func (s *Scene) BeginEnding() {
	s.ending = endingState{active: true}
	if s.rom == nil {
		return // 測試用的空場景：結局照樣「進去了」，只是沒有素材
	}
	if im, err := s.rom.End(); err == nil {
		s.ending.pic = im
	}
	if a, err := s.rom.EndAnim(); err == nil {
		s.ending.anim = a
		if len(a.Frames) > 0 {
			s.ending.wait = a.Frames[0].Delay
		}
	}
	// ⚠ 原版還會逐人把角色記錄 `+0x4B` 的 bit0 設起來（`0x1B4FB`）。
	// **那個 bit 的語意未解**（`docs/re/91` §2.1 的同一個欄位，
	// Radio 第一輪讀的也是它），角色結構還沒有對應欄位——
	// 照 `CLAUDE.md` §0 不猜，這一步先不做，記在 `docs/re/96` §5。
	s.message = ""
	s.dirty = true
}

// TickEnding 推進一個 BIOS tick 的結局播放。
//
// 回傳 false 表示播完了（原版播完最後一段就回主流程）。
func (s *Scene) TickEnding() bool {
	e := &s.ending
	if !e.active {
		return false
	}
	e.tick++

	// 動畫：延遲到了就疊下一格，跑完從 LoopFrom 再來（`sub_1B735`）。
	if e.anim != nil && len(e.anim.Frames) > 0 {
		if e.wait > 0 {
			e.wait--
		} else {
			f := e.anim.Frames[e.frame]
			f.Apply(e.pic)
			e.frame++
			if e.frame >= len(e.anim.Frames) {
				e.frame = e.anim.LoopFrom
			}
			e.wait = e.anim.Frames[e.frame].Delay
			s.dirty = true
		}
	}

	// 進場的兩段跑完才開始印字。
	if e.tick == endingIntroTicks || (e.page > 0 && (e.tick-endingIntroTicks)%endingPageTicks == 0) {
		if e.page >= len(EndingPages) {
			e.active = false
			return false
		}
		s.message = s.endingText(EndingPages[e.page])
		e.page++
		s.dirty = true
	}
	return true
}

// endingText 取第 n 條結局敘述。查不到就回空字串。
func (s *Scene) endingText(n int) string {
	tables, err := s.rom.ExeStrings()
	if err != nil || EndingTable >= len(tables) || n >= len(tables[EndingTable]) {
		return ""
	}
	return tables[EndingTable][n]
}

// updateEnding 是結局播放中的按鍵。
//
// **進場的兩段按鍵無效**（原版 `ds:D168h` 那時是 0）；之後按任意鍵跳到下一段。
func (s *Scene) updateEnding(in input.Input) (bool, error) {
	// 動畫與計時走畫面更新，不等按鍵。
	// **BIOS 一個 tick 是 1/18.2 秒**，Ebiten 一幀是 1/60——每 3 幀推一個 tick，
	// 不然結局會用三倍速播完（`docs/re/96` §4）。
	s.ending.sub++
	if s.ending.sub >= 3 {
		s.ending.sub = 0
		s.TickEnding()
	}
	if s.ending.tick < endingIntroTicks {
		return true, nil
	}
	if in.Char == 0 && in.Action == 0 {
		return true, nil
	}
	if s.ending.page >= len(EndingPages) {
		s.ending.active = false
		s.dirty = true
		return true, nil
	}
	s.message = s.endingText(EndingPages[s.ending.page])
	s.ending.page++
	s.dirty = true
	return true, nil
}

// drawEnding 畫結局：整張圖佔圖片視窗，敘述走訊息視窗。
func (s *Scene) drawEnding(f *render.Frame) {
	if s.ending.pic != nil {
		_ = f.DrawPicture(s.ending.pic)
	}
}
