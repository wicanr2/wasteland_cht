package play

// 遭遇迴圈：地圖與戰鬥的接合（docs/spec/21、docs/re/51）。
//
// 規格 12–16 各自解決戰鬥裡的一塊，規格 04 解決地圖上的一步。
// 這個檔只做中間那一層——**誰決定要打、什麼時候切畫面、打完怎麼回地圖**。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
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
	groups := s.scanGroups()
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
	// 遭遇時載入那種敵人的肖像圖（敵人資料 `+0x07`，`docs/re/37` §3.2）。
	// 走的是與設施圖同一支載入器 `sub_184E8`（`docs/re/23` §4），
	// 所以顯示位置也一樣：視窗原點 (8, 8)、96 × 84。
	s.portrait = -1
	for slot := 0; slot < game.EnemySlots; slot++ {
		if e := b.Enemy(slot); e != nil && e.HP > 0 {
			s.portrait = int(e.Data.Portrait)
			break
		}
	}
	// 中文那三條路（原版字串表 1、重製版介面文字、敵人名稱）與 `World`
	// 都在 `wireCombat` 裡一起接——三個都可以是 nil／空，那時戰鬥訊息就是英文。
	c := s.wireCombat(NewCombatScene(b))
	c.Log = append(c.Log, "Encounter begins...")
	c.LastCJK = c.zhStr(strEncounterBegins, textlayout.Options{})
	s.combat = c
	return c, nil
}

// scanGroups 建遭遇掃描要的四組狀態（`docs/re/39` §4）。
//
// **只有在同一張地圖上的組才參與**，而且距離要用各組自己的座標算。
// 目前這一組拿記憶體裡的隊伍（座標可能還沒寫回槽表），其餘組讀槽表。
func (s *Scene) scanGroups() [game.QueueGroups]game.PartyGroupState {
	var out [game.QueueGroups]game.PartyGroupState
	for i, g := range s.save.SlotGroups() {
		if i >= len(out) {
			break
		}
		if i == s.groupID {
			out[i] = game.PartyGroupState{
				Present: true,
				X:       int(s.world.Party.X),
				Y:       int(s.world.Party.Y),
				Engage: game.Engagement(s.world.Party.Members,
					func(int) bool { return false }),
			}
			continue
		}
		if groupSize(g) == 0 || int(g.MapID) != s.blockID {
			continue // 空的、或不在這張地圖上
		}
		members := s.groupMembers(g)
		out[i] = game.PartyGroupState{
			Present: true,
			X:       int(g.X),
			Y:       int(g.Y),
			Engage:  game.Engagement(members, func(int) bool { return false }),
		}
	}
	return out
}

// groupMembers 把一組槽表的角色記錄讀出來（只給接戰值用，不建完整隊伍）。
func (s *Scene) groupMembers(g assets.SlotGroup) []*game.Character {
	var out []*game.Character
	for _, id := range g.Members {
		if id == 0 {
			continue
		}
		raw, err := s.save.Record(int(id))
		if err != nil {
			continue
		}
		out = append(out, game.LoadCharacter(raw))
	}
	return out
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

// enemyNamesCJK 是六個敵人種類的中文名（Big5）。
//
// 與 `enemyNames` 同一組編號（`0x52 + Kind`，`docs/re/85`），差別只在
// 這裡走翻譯目錄、而且單複數由 `RenderBytes` 依 Count ＝ 1 取單數那一段——
// **不能自己找 `0x0A` 切**：譯文的分段位置與原文不一定一樣。
func (s *Scene) enemyNamesCJK() [6]string {
	var out [6]string
	for k := range out {
		b := s.cjkExe(exeTable1, 0x52+k, textlayout.Options{Count: 1})
		out[k] = strings.TrimSpace(b)
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
	s.portrait = -1
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
