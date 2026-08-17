package play

// 全隊倒下的處置（`docs/spec/28`，逆向在 `docs/re/99`）。
//
// 原版主迴圈每一輪檢查一次（`0x16C2B`，位置就在自毀倒數那道檢查後面），
// 三種結果：什麼都不做、自動切到下一支隊伍、死亡畫面。
//
// 死亡畫面是**設施畫面那個形狀的第六張**（圖 ＋ 地點名 ＋ 一句話 ＋ 等按鍵），
// 只是它沒掛在設施跳表上，是 `0x1C570` 手寫的一段。

import (
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

const (
	// wipePicture 是死亡畫面那張圖（`mov al, 3Bh; call sub_190A6`）。
	wipePicture = 0x3B
	// dsWipePlace／dsWipeText 是那兩段**明文 ASCII**在映像裡的位置。
	// 它們不在打包字串表裡，抽字串的工具看不到（`docs/re/99` §2）。
	dsWipePlace = 0xDE60 // `Grim Reaper ` ＋ NUL，13 bytes
	dsWipeText  = 0xDE6D // `Your life has ended in The Wasteland.\r`
)

// wipeState 是死亡畫面的狀態。
type wipeState struct {
	active bool
	place  string // 從映像讀出來的 `Grim Reaper`
	text   string // 從映像讀出來的那一句
}

// checkWipe 是主迴圈那一輪的檢查。回傳 false 表示這一幀已經被處理掉了。
func (s *Scene) checkWipe() {
	if s.wipe.active || s.ending.active || s.combat != nil {
		return
	}
	switch s.world.Party.Wipe() {
	case game.WipeNone:
		return
	case game.WipeSwitch:
		// 全倒但救得回來 → 原版自動切到下一支隊伍（`0x16C69` 的 `View`）。
		// ⚠ 沒有別組可切時**不是**什麼都不做：那時這一組就是全部，
		// 走死亡畫面（原版的 `View` 繞回原組，下一輪同樣的檢查會再跑一次）。
		if n, ok := s.nextGroup(); ok {
			if err := s.SwitchGroup(n); err == nil {
				s.sayEN("Party "+string(rune('1'+n))+".", "view.switched")
				return
			}
		}
	}
	s.beginWipe()
}

// beginWipe 進死亡畫面。
func (s *Scene) beginWipe() {
	s.wipe = wipeState{active: true}
	// 兩段文字從映像直讀——它們是原版的字，不是我們編的。
	if s.rom != nil {
		if b, err := s.rom.DsBytes(dsWipePlace, 13); err == nil {
			s.wipe.place = strings.TrimRight(string(b), "\x00 ")
		}
		if b, err := s.rom.DsBytes(dsWipeText, 40); err == nil {
			s.wipe.text = strings.TrimRight(cutAtNUL(string(b)), "\r ")
		}
	}
	s.resetModes()
	s.wipe.active = true // resetModes 會清掉所有模式，這一個要留著
	s.message = s.wipe.text
	if zh := s.uiText("wipe.message"); len(zh) > 0 {
		s.message, s.cjk = "", zh
	}
	s.dirty = true
}

// cutAtNUL 砍掉第一個 NUL 之後的東西（映像是連續的，讀多了會拖到下一句）。
func cutAtNUL(s string) string {
	if i := strings.IndexByte(s, 0); i >= 0 {
		return s[:i]
	}
	return s
}

// WipePlaceLine 是死亡畫面那一行地點名（英文，畫在設施招牌的位置）。
func (s *Scene) WipePlaceLine() string { return s.wipe.place }

// wipePlaceCJK 是同一行的中文。
//
// ⚠ 它走 `ui:` 不走地點招牌那張表：招牌表的 key 是**資料裡真實存在的招牌**
// （`TestNoStalePlaceKeys` 對著地圖記錄驗），而 `Grim Reaper` 是寫死在
// 執行檔裡的一段，不是任何一張地圖上的招牌。
func (s *Scene) wipePlaceCJK() string { return s.uiText("wipe.place") }

// updateWipe 是死亡畫面的按鍵：**任何鍵都回標題畫面**。
//
// ⚠ 原版按鍵之後回到哪裡還沒讀出來（`sub_18E90` 之後那一段沒追），
// 回標題是重製版的決定。原版沒有遊戲內的重新開始——重來要離開遊戲跑
// `SETUP.EXE`（`docs/re/99` §1），重製版不假裝有那個選項，
// 玩家自己決定要不要按 F9 讀快速存檔。
func (s *Scene) updateWipe(in input.Input) (bool, error) {
	if in.Dir == input.DirNone && in.Action == input.ActionNone &&
		in.Char == 0 && in.Fn == input.FnNone {
		return true, nil
	}
	s.wipe = wipeState{}
	s.message, s.cjk = "", ""
	s.BeginTitle()
	return true, nil
}

// drawWipe 畫死亡畫面：圖 ＋ 地點名（與設施畫面同一個形狀、同一組座標）。
func (s *Scene) drawWipe(f *render.Frame) {
	if s.pics != nil && wipePicture < len(s.pics) {
		f.DrawIndexed(s.pics[wipePicture], render.FacilityPicX, render.FacilityPicY,
			render.ViewClip())
	}
	// 有中文時這一行交給 HiFrame（8 × 8 的字模畫不出中文，
	// 先畫英文再蓋會留殘影——與設施招牌同一條）。
	if len(s.wipePlaceCJK()) == 0 && !s.hiText() {
		_ = f.DrawLineAt(s.font, s.wipe.place, render.FacilityNameCol, render.FacilityNameRow)
	}
}
