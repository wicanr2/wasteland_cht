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
//
// 四段敘述之後是**清算**（`0x1B571`）：逐組看槽表 `+0x0A` 的地圖編號，
// **`0x10` ≤ 地圖 ＜ `0x15` 的那幾組人全部死在爆炸裡**——一個一個念
// 「`\v was killed in the blast.`」，記錄 `+0x1D`／`+0x1E`（CON）歸零，
// 然後整組從槽表刪掉。不在 Base Cochise 裡的隊伍活下來。

import (
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/lang"
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

	// killed 是清算階段點名的死者，依原版的順序。
	killed []string
	// toll 是清算跑過了沒（四段敘述播完才做，只做一次）。
	toll bool
}

// BaseCochiseMaps 是 Base Cochise 的地圖編號範圍（半開區間）。
//
// 原版的判斷就在 `0x1B58A`：槽表 `+0x0A` 的地圖編號 `< 0x10` 或 `≥ 0x15`
// 就跳過那一組——**只有這五張地圖上的人會死**。
const (
	BaseCochiseFirst = 0x10
	BaseCochiseEnd   = 0x15
)

// EndingKilledString 是「`\v was killed in the blast.`」在結局表裡的編號。
const EndingKilledString = 5

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
	// 逐人記上「參與過摧毀 Base Cochise」（`0x1B4FB` 設記錄 `+0x4B` 的 bit0）。
	// **活下來的人靠這個 bit 去 Radio 領表揚**（`docs/re/96` §5）。
	for _, m := range s.world.Party.Members {
		if m != nil {
			m.Mission = true
		}
	}
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
			s.collectToll()
			e.active = false
			return false
		}
		s.showEndingPage(EndingPages[e.page])
		e.page++
		s.dirty = true
	}
	return true
}

// endingCJK 取第 n 條結局敘述的中文（key `exe:4:<n>`）。沒翻就回 nil。
func (s *Scene) endingCJK(n int) string {
	if s.cat == nil {
		return ""
	}
	if b, ok := s.cat.Lookup(lang.ExeKey(EndingTable, n)); ok {
		return b
	}
	return ""
}

// endingText 取第 n 條結局敘述的英文原文。查不到就回空字串。
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
		s.collectToll()
		s.ending.active = false
		s.dirty = true
		return true, nil
	}
	s.showEndingPage(EndingPages[s.ending.page])
	s.ending.page++
	s.dirty = true
	return true, nil
}

// showEndingPage 把第 n 條敘述放進訊息視窗。
//
// ⚠ **有中文就不要留英文**（與 `showCombatPrompt`、`sayPlace` 同一條規矩）：
// 兩份一起印會佔掉訊息視窗六列裡的兩列，而且英文那一列與中文那一列
// **緊貼著**——24 點的中文字填滿整個字元格，看起來像兩行疊在一起。
// 查不到譯文時 `endingCJK` 回空字串，那時就只有英文，fallback 不變。
func (s *Scene) showEndingPage(n int) {
	s.message = s.endingText(n)
	s.cjk = s.endingCJK(n)
	if s.cjk != "" {
		s.message = ""
	}
}

// drawEnding 畫結局：整張圖佔圖片視窗，敘述走訊息視窗。
func (s *Scene) drawEnding(f *render.Frame) {
	if s.ending.pic != nil {
		_ = f.DrawPicture(s.ending.pic)
	}
}

// collectToll 是清算（`0x1B571`）：Base Cochise 裡的隊伍全部死在爆炸裡。
//
// 原版逐組做三件事——把每個人的 CON 歸零並念一句、把那個人移出隊伍、
// 整組跑完就從槽表刪掉並把後面的組往前搬（`0x1B5F2` 的 15-byte 搬移）。
// 這裡照做，**順序也照原版**：死者名單就是玩家看到的那一串。
func (s *Scene) collectToll() {
	if s.ending.toll || s.save == nil {
		return
	}
	s.ending.toll = true

	line := s.endingText(EndingKilledString)
	for n, g := range s.save.SlotGroups() {
		if groupSize(g) == 0 || int(g.MapID) < BaseCochiseFirst || int(g.MapID) >= BaseCochiseEnd {
			continue
		}
		for _, name := range s.groupMemberNames(n) {
			s.ending.killed = append(s.ending.killed, strings.Replace(line, "\v", name, 1))
		}
		// 整組清空（原版 `sub_1B7EC`：槽表 14 bytes 全歸零）。
		slot := s.save.Plain[g.RawIndex : g.RawIndex+14]
		for i := range slot {
			slot[i] = 0
		}
		if n == s.groupID {
			s.world.Party.Members = nil
		}
	}
}

// Killed 回傳清算的死者名單（依原版的點名順序）。
func (s *Scene) Killed() []string { return s.ending.killed }

// groupMemberNames 取第 n 組每個人的名字，並把他們的 CON 歸零。
func (s *Scene) groupMemberNames(n int) []string {
	var out []string
	if n == s.groupID {
		for _, m := range s.world.Party.Members {
			if m == nil {
				continue
			}
			m.CON = 0 // 原版把記錄 +0x1D／+0x1E 歸零
			out = append(out, m.Name)
		}
		return out
	}
	// 別組的人只在存檔裡——照角色記錄改，不建立一份規則層的隊伍。
	p, _, err := loadPartyGroup(s.save, n)
	if err != nil {
		return nil
	}
	for _, m := range p.Members {
		if m != nil {
			m.CON = 0
			out = append(out, m.Name)
		}
	}
	return out
}
