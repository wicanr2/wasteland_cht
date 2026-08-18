package play

// 戰鬥指令 `H` 雇用（`docs/re/110`、規格 17 §4）。
//
// 兩段，與原版一樣分在指令階段與結算階段：
//
//	指令階段  印字串 63「Which group?」→ 收一個組號 → 記成指令參數
//	結算階段  讀遭遇記錄的 +0x09 → 取 section 17 那筆 NPC 記錄 → 檢定 → 加入隊伍
//
// ⚠ **可雇用與否是在結算階段才篩的**（`loc_1382B` 只數「有敵人的組」）。
// 提前篩掉不能雇用的組會讓「按下去才發現不行」這個原版行為消失。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// 原版字串編號（表 1）。
const (
	strWhichGroup = 63  // `Which group?`
	strTriesHire  = 44  // `\r\x0B tries to hire `
	strButFails   = 81  // ` but fails.\r`
	strNoRoom     = 95  // `\rNo room in roster.\r`
)

// hirePick 是「哪一組？」開著的狀態。
type hirePick struct {
	open   bool
	groups []int // 還有敵人的組號（0-based）
}

// beginHirePick 開「哪一組？」。
//
// 候選是 `loc_1382B` 數的那一批：**這場遭遇裡還有敵人的組**，
// 不看能不能雇用。一組都沒有就照規格 17 §4 印「射程內沒有可雇用的對象」。
func (s *CombatScene) beginHirePick() bool {
	var groups []int
	for g := 0; g < game.EnemyGroups; g++ {
		if s.Battle.GroupAlive(g) {
			groups = append(groups, g)
		}
	}
	if len(groups) == 0 {
		s.Log = append(s.Log, "No one is within range.")
		s.LastCJK = s.zhStr(int(game.MsgNoOneInRange), textlayout.Options{})
		return false
	}
	s.hire = hirePick{open: true, groups: groups}
	s.Log = append(s.Log, "Which group?")
	s.LastCJK = s.zhStr(strWhichGroup, textlayout.Options{})
	return true
}

// HirePicking 回答「哪一組？」開著沒有。
func (s *CombatScene) HirePicking() bool { return s.hire.open }

// HirePickLines 是選單的英文顯示（一組一行，數字鍵選）。
func (s *CombatScene) HirePickLines(name func(group int) string) []string {
	if !s.hire.open {
		return nil
	}
	out := make([]string, 0, len(s.hire.groups))
	for i, g := range s.hire.groups {
		label := fmt.Sprintf("group %d", g+1)
		if name != nil {
			if n := name(g); n != "" {
				label = n
			}
		}
		out = append(out, fmt.Sprintf("%d) %s", i+1, label))
	}
	return out
}

// PickHire 收「哪一組？」的一個按鍵。回傳這個按鍵有沒有被吃掉。
func (s *CombatScene) PickHire(key byte) bool {
	if !s.hire.open {
		return false
	}
	if key >= '1' && key <= '9' {
		i := int(key - '1')
		if i >= len(s.hire.groups) {
			return true // 沒有那一組：吃掉按鍵，不要當成別的指令
		}
		g := s.hire.groups[i]
		s.hire = hirePick{}
		s.Phase.Set(s.Turn, game.CmdHire, byte(g))
		s.advance(s.Turn + 1)
		return true
	}
	return true // 選單開著就吃掉所有按鍵
}

// CancelHirePick 關掉選單，回到指令選單（原版回 0xFF ＝ 取消，重問）。
func (s *CombatScene) CancelHirePick() { s.hire = hirePick{} }

// resolveHire 是結算階段那一段（`0x132AC`–`0x133E0`）。
//
// 順序照原版，**不要重排**：先看名冊滿不滿（滿了連骰都不擲），
// 再看這一組能不能雇用，最後才檢定。
func (s *CombatScene) resolveHire(i int, m *game.Character) msgs {
	var out msgs
	w := s.World
	if w == nil || w.Block == nil {
		return out
	}
	g := int(s.Phase.Arg[i])
	fail := func() msgs {
		// 原版：「⟨名字⟩ tries to hire ⟨那一組⟩ but fails.」（字串 44 ＋ 81）
		out.add(m.Name+" tries to hire "+s.groupName(g)+" but fails.",
			s.hireFailCJK(m, g))
		return out
	}

	if len(w.Party.Members) >= game.HireCap {
		out.add("No room in roster.", s.zhStr(strNoRoom, textlayout.Options{}))
		return out
	}
	offer := game.ReadHireOffer(s.EncRecord)
	if !offer.Valid {
		return fail()
	}
	npcRec, err := w.Block.SectionRecord(game.HireSection(), offer.NPC)
	if err != nil || len(npcRec) < 0x100 {
		return fail()
	}
	npc := game.LoadCharacter(npcRec)
	res := game.TryHire(m, npc, game.HirePrice(npcRec), w.RNG)
	if !res.Joined {
		return fail()
	}

	joined := game.HireNPC(npcRec, m.Attributes[game.AttrCharisma])
	if joined == nil {
		return fail()
	}
	w.Party.Members = append(w.Party.Members, joined)
	s.Battle.RemoveGroup(g) // 那一組加入了隊伍，從戰場上移除（不是打死）

	// 加入時印 NPC 記錄 `+0x30` 那條字串（`0x13362`）。
	en := joined.Name + " joins the party."
	if greet := int(game.HireGreeting(npcRec)); greet != 0 {
		if zh := s.zhStr(greet, textlayout.Options{Name: func() string { return joined.Name }}); zh != "" {
			out.add(en, zh)
			return out
		}
	}
	out.add(en, s.uiText("hire.joins", joined.Name))
	return out
}

// GroupName 是那一組敵人的名稱，給呈現層畫清單用。
func (s *CombatScene) GroupName(g int) string { return s.groupName(g) }

// HirePickCJK 是「哪一組？」清單的中文：一行一組，數字鍵選。
//
// 組名查不到中文就用英文名（名稱不是句子，混排讀得通）；
// **框架字查不到就整份回空字串**走英文那一份。
func (s *CombatScene) HirePickCJK() string {
	if !s.hire.open {
		return ""
	}
	// ⚠ **標題不在這裡**：`beginHirePick` 已經把「哪一組？」放進訊息區了，
	// 這裡再加一次畫面上就會出現兩行一樣的字。
	// 它仍然是這一份能不能走中文的判準——查不到就整份回空字串走英文。
	if s.zhStr(strWhichGroup, textlayout.Options{}) == "" {
		return ""
	}
	out := ""
	for i, g := range s.hire.groups {
		if i > 0 {
			out += string('\n')
		}
		out += string(rune('1'+i)) + ") " + s.groupNameCJK(g)
	}
	return out
}

// groupName 是那一組敵人的名稱，訊息要用。查不到就回 `group N`。
func (s *CombatScene) groupName(g int) string {
	for i := 0; i < game.EnemiesPerGroup; i++ {
		if e := s.Battle.Enemy(g*game.EnemiesPerGroup + i); e != nil {
			if n := s.Names.Name(e.Data.Kind); n != "" {
				return n
			}
		}
	}
	return fmt.Sprintf("group %d", g+1)
}

// hireFailCJK 是失敗那一句的中文：字串 44（「⟨名字⟩試著雇用」）
// ＋ 那一組的名稱 ＋ 字串 81（「但失敗了。」）。
//
// ⚠ 三段**任一段查不到就整句放棄**（回空字串走英文）——
// 半句中文半句英文比整句英文更難讀，也讓「哪裡沒翻」看不出來。
func (s *CombatScene) hireFailCJK(m *game.Character, g int) string {
	head := s.zhStr(strTriesHire, textlayout.Options{Name: func() string { return m.Name }})
	tail := s.zhStr(strButFails, textlayout.Options{})
	if head == "" || tail == "" {
		return ""
	}
	return head + s.groupNameCJK(g) + tail
}

// groupNameCJK 是那一組的中文名；查不到回英文名（名稱不是句子，混排讀得通）。
func (s *CombatScene) groupNameCJK(g int) string {
	for i := 0; i < game.EnemiesPerGroup; i++ {
		if e := s.Battle.Enemy(g*game.EnemiesPerGroup + i); e != nil {
			if n := s.CJKNames[int(e.Data.Kind)%len(s.CJKNames)]; n != "" {
				return n
			}
		}
	}
	return s.groupName(g)
}

// uiText 查重製版自己的介面文字，帶一個參數。UI 沒接上就回空字串。
func (s *CombatScene) uiText(key string, args ...any) string {
	if s.UI == nil {
		return ""
	}
	f := s.UI(key)
	if f == "" {
		return ""
	}
	if len(args) == 0 {
		return f
	}
	return fmt.Sprintf(f, args...)
}
