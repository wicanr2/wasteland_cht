package play

// 遭遇迴圈：地圖與戰鬥的接合（docs/spec/21、docs/re/51）。
//
// 規格 12–16 各自解決戰鬥裡的一塊，規格 04 解決地圖上的一步。
// 這個檔只做中間那一層——**誰決定要打、什麼時候切畫面、打完怎麼回地圖**。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// EncounterResult 是一場遭遇跑完的結果。
type EncounterResult struct {
	Fought   bool              // 有沒有真的打起來
	XPGained map[string]uint32 // 角色名字 → 這一場拿到的經驗值
	Wiped    bool              // 隊伍全滅（原版走另一條路，本版先回報）
}

// xpSnapshot 是打之前每個角色的經驗值。
//
// ⚠ **原版沒有「這一場拿了多少」這個欄位**，它是打之前抄一份、打完相減
// （`sub_11CD0` 的 0x11CF1 與 0x11ED7，docs/re/51 §3、§6）。
// 照做就不必多一個會不同步的累加器。
type xpSnapshot map[string]uint32

// StartEncounter 從地圖進戰鬥。
//
// 回 nil 表示掃描說沒有可打的——地圖照舊，畫面不切（docs/re/51 §2）。
func (s *Scene) StartEncounter() (*CombatScene, error) {
	var groups [game.QueueGroups]game.PartyGroupState
	// 只有第 0 組：DISBAND 的分隊管理還沒做（docs/spec/21 §4）。
	groups[0] = game.PartyGroupState{
		Present: true,
		Engage:  game.Engagement(s.world.Party.Members, func(int) bool { return false }),
	}
	scan := s.world.ScanEncounters(groups)
	entry, ok := scan.Queue.Nearest(0)
	if !ok {
		return nil, nil
	}

	rec, _, err := s.world.Block.CellRecord(int(entry.X), int(entry.Y))
	if err != nil {
		return nil, fmt.Errorf("遭遇格 (%d, %d) 的記錄：%w", entry.X, entry.Y, err)
	}
	raw, err := s.world.Block.EnemyData()
	if err != nil {
		return nil, fmt.Errorf("這張地圖的敵人資料表：%w", err)
	}

	b := game.NewBattle(s.world.Party, s.world.RNG)
	if n := b.Spawn(rec, game.ParseEnemyTable(raw), s.world.RNG); n == 0 {
		return nil, nil // 記錄說有敵人但一個都生不出來 → 不打
	}
	s.snapshot = s.takeXP()
	c := NewCombatScene(b)
	c.Items = s.items
	c.Names = s.enemyNames()
	c.Log = append(c.Log, "Encounter begins...")
	s.combat = c
	return c, nil
}

// enemyNames 查六個敵人種類的名稱（`0x52 + Kind`，取單數那一段）。
func (s *Scene) enemyNames() EnemyNames {
	var out EnemyNames
	for k := range out {
		raw := s.exeString(0x52 + k)
		if i := strings.IndexByte(raw, 0x0A); i >= 0 {
			raw = raw[:i]
		}
		out[k] = strings.TrimSpace(raw)
	}
	return out
}

// takeXP 抄一份每個角色的經驗值。
func (s *Scene) takeXP() xpSnapshot {
	out := make(xpSnapshot, len(s.world.Party.Members))
	for _, m := range s.world.Party.Members {
		out[m.Name] = m.XP
	}
	return out
}

// FinishEncounter 收尾：算經驗值差、切回地圖。
//
// 原版最後是 `sub_18350`（重新載入地圖）＋ `jmp sub_163C4`（重畫），
// 座標、時鐘、地圖都不變——所以這裡只切模式，不動世界狀態。
func (s *Scene) FinishEncounter() EncounterResult {
	res := EncounterResult{Fought: s.combat != nil, XPGained: map[string]uint32{}}
	for _, m := range s.world.Party.Members {
		before := s.snapshot[m.Name]
		if m.XP > before {
			res.XPGained[m.Name] = m.XP - before
		}
	}
	res.Wiped = true
	for _, m := range s.world.Party.Members {
		if game.CanCommand(m) {
			res.Wiped = false
			break
		}
	}
	s.combat = nil
	s.snapshot = nil
	s.dirty = true
	return res
}

// InCombat 回報現在是不是在戰鬥畫面。
func (s *Scene) InCombat() bool { return s.combat != nil }

// Combat 回傳目前的戰鬥畫面（沒有就 nil）。
func (s *Scene) Combat() *CombatScene { return s.combat }
