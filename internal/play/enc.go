package play

// `ENC` 指令（`sub_11CD0`，指令表入口是 `0x11CE7`）。
//
// 它是**戰鬥驅動器的手動入口**：走一步之後掃到遭遇會自動進戰鬥，
// 按 `E` 則是玩家自己叫它跑一輪。同一支程式碼兩個入口
// （`docs/re/51` §1）——所以這裡不另寫一套，直接叫 `StartEncounter`。
//
// 原版的完整形狀是「對四支隊伍各跑一個戰鬥回合」：
// 逐組切過去（`sub_16149`），**只有在同一張地圖的組直接開打**，
// 其餘的先印字串 `0x36` ＋ `0x4C` 問一句 Y／N（`docs/re/94` §2）。
// 打完回到 `0x11ED7` 逐人把經驗值前後相減，差值不為零就印
// 「`\x0b gains ` 數字 ` experience.`」——**原版沒有「這場拿多少」這個欄位**，
// 它是打之前抄一份、打完相減，`Scene.finishEncounter` 已經照做。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// encOffMapPrompt 是字串 `0x36` ＋ `0x4C` 接起來的那一句。
//
// 原文照抄不翻——**熱鍵比對走 `Y`／`N` 的靜態字母**，
// 顯示文字要中文化時走翻譯目錄（`docs/re/40` §4）。
const encOffMapPrompt = "This party isn't on this map and isn't in battle. " +
	"Do you want them to execute a battle round? (Y/N)"

// cmdEnc 是 `Enc`：先讓目前這一組打，沒得打就問別組要不要打。
func (s *Scene) cmdEnc() (bool, error) {
	if s.InFacility() {
		s.sayEN("Not in here.", "enc.inside")
		return true, nil
	}
	c, err := s.StartEncounter()
	if err != nil {
		s.message = "ERROR: " + err.Error()
		s.dirty = true
		return true, nil
	}
	if c != nil {
		s.sayAttacked()
		s.showCombatPrompt()
		return true, nil
	}

	// 這一組沒得打——**原版還是會問「要不要跑一個戰鬥回合」**
	// （`sub_11F76` 的 `0x11FD9`：字串 20 ＋ 76，`docs/re/105`）。
	// 答 Y 就照常進指令階段，玩家因此可以在空地上換武器、裝填、用道具。
	if s.partyCanAct() {
		return s.askConfirmText(encNoEnemyPrompt,
			zhJoin(s.cjkExe(exeTable1, strNotAttacked, textlayout.Options{}),
				s.cjkExe(exeTable1, strExecuteRound, textlayout.Options{})),
			s.beginEmptyRound)
	}

	// 這一組一個人都站不起來。原版接著往下一組走；不同地圖的要先問一句。
	n, ok := s.encOffMapGroup()
	if !ok {
		s.sayEN("Nothing to fight here.", "enc.none")
		return true, nil
	}
	s.encAsk = n + 1 // 0 保留給「沒在問」
	s.message = encOffMapPrompt
	// 中文是同兩條原版字串接起來（54 ＋ 76）。
	s.cjk = zhJoin(s.cjkExe(exeTable1, strNotOnMap, textlayout.Options{}),
		s.cjkExe(exeTable1, strExecuteRound, textlayout.Options{}))
	if s.cjk != "" {
		s.message = ""
	}
	s.dirty = true
	return true, nil
}

// 這一支用到的原版字串編號（字串表 1）。
const (
	strNotAttacked  = 0x14 // `This party is not being attacked. `
	strNotOnMap     = 0x36 // `This party isn't on this map and isn't in battle. `
	strExecuteRound = 0x4C // `Do you want them to execute a battle round?` ＋ Yes／No
)

// encNoEnemyPrompt 是字串 `0x14` ＋ `0x4C` 接起來的那一句。
//
// ⚠ **與 `encOffMapPrompt` 是兩條不同的路**：這一條問的是「**這一組**沒有
// 敵人在打，要不要還是跑一個回合」，那一條問的是「**別組**不在這張地圖上」。
// 原版兩處都在（`0x11FD9` 與 `0x11D97`），只接一邊會少掉一整個玩法。
const encNoEnemyPrompt = "This party is not being attacked. " +
	"Do you want them to execute a battle round? (Y/N)"

// partyCanAct 回報這一組還有沒有人站得起來（原版 `sub_19D0E`，`0x11F8C`）。
func (s *Scene) partyCanAct() bool {
	for _, m := range s.world.Party.Members {
		if m != nil && !m.Down() {
			return true
		}
	}
	return false
}

// beginEmptyRound 進「沒有敵人的戰鬥回合」（`sub_11F76` 的 `0x11FEF`）。
//
// 沒有敵人時 `Battle.Over()` 立刻回 (true, true)，所以這一回合下完令就結束，
// 畫面回地圖——與原版一樣。**指令階段本身照跑**，換武器與裝填才有意義。
func (s *Scene) beginEmptyRound() (bool, error) {
	b := game.NewBattle(s.world.Party, s.world.RNG)
	s.snapshot = s.takeXP()
	s.portrait = -1
	s.combat = s.wireCombat(NewCombatScene(b))
	s.message, s.cjk = "", ""
	s.showCombatPrompt()
	s.dirty = true
	return true, nil
}

// sayAttacked 是「被攻擊了」那一句。原版沒有這一條字串
//（遭遇開始印的是 30 `Encounter begins...`），所以走 `ui:`。
func (s *Scene) sayAttacked() {
	s.sayEN("YOU ARE BEING ATTACKED!", "combat.attacked")
}

// encOffMapGroup 找第一支「有人、不是目前這組、而且不在這張地圖上」的隊伍。
//
// 原版的條件是「那張地圖不在交戰清單 `ds:A9B0h` 上，也不是主地圖
// `ds:46E0h`」（`docs/re/94` §2）。交戰清單是原版拿來記「哪幾張地圖還有
// 未結束的戰鬥」用的，重製版一次只驅動一支隊伍的戰鬥，
// **那張表沒有對應物**——所以這裡只用「不在同一張地圖」這個條件，
// 同一張地圖的組會走上面那條直接開打的路。
func (s *Scene) encOffMapGroup() (int, bool) {
	for n, g := range s.save.SlotGroups() {
		if n == s.groupID || groupSize(g) == 0 {
			continue
		}
		if int(g.MapID) != s.blockID {
			return n, true
		}
	}
	return 0, false
}

// updateEncAsk 是那一句 Y／N 的按鍵。
//
// 答 `N`／ESC ＝ 跳過這一組（原版 `0x11DA6` 跳到 `loc_11E8A`，
// 也就是繼續掃下一組）；答 `Y` 就切過去讓那一組打。
func (s *Scene) updateEncAsk(in input.Input) (bool, error) {
	n := s.encAsk - 1
	switch {
	case in.Action == input.ActionCancel, input.Upper(in.Char) == 'N':
		s.encAsk = 0
		s.message = ""
		s.dirty = true
		return true, nil
	case input.Upper(in.Char) == 'Y':
		s.encAsk = 0
		if err := s.SwitchGroup(n); err != nil {
			s.message = "ERROR: " + err.Error()
			s.dirty = true
			return true, nil
		}
		c, err := s.StartEncounter()
		switch {
		case err != nil:
			s.message = "ERROR: " + err.Error()
		case c != nil:
			s.sayAttacked()
			s.showCombatPrompt()
		default:
			s.message = fmt.Sprintf("Party %d: nothing to fight.", n+1)
			if f := s.uiText("enc.groupnone"); len(f) > 0 {
				s.cjk = fmt.Sprintf(f, n+1)
				s.message = ""
			}
		}
		s.dirty = true
		return true, nil
	}
	return true, nil
}
