package play

// 寶箱撿拾（nibble 5）。判定照原版（`docs/re/29` §4、`docs/re/130`）：
// 第一次踩到才擲出內容並寫回記錄 → 「誰要撿？」 → 逐件拿 →
// 全部拿完用位移 0 的改寫對把這一格改掉。
//
// 呈現是重製決策：原版切到名單畫面（`sub_1728C`）再開清單框架，
// 這裡照 `USE` 的慣例走訊息視窗＋數字鍵（`docs/re/92` §7 同一條）。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// 撿拾用到的原版字串編號（表 1，`docs/re/130` §5）。
const (
	strWhoWantsLoot = 16 // `Who wants loot?`
	strCantGetAny   = 17 // `\x0B can't get any.`
	strCantCarry    = 18 // `You can't carry any more.`
)

// lootState 是撿拾流程的狀態。
type lootState struct {
	active bool
	data   []byte // 這一格的 section 記錄（活切片，寫回就是進地圖資料）
	x, y   int
	who    int // 選中的隊員（−1 ＝ 還在問「誰要撿？」）
}

// beginLoot 踩上寶箱格：擲出內容（第一次才會真的擲）再問誰要撿。
func (s *Scene) beginLoot(x, y int, data []byte) {
	if len(data) == 0 {
		return
	}
	s.world.RollChest(s.items, data)
	if len(game.ChestEntries(data)) == 0 {
		// 空的（類別一件都擲不出來，或早就拿光）：直接收尾改寫。
		s.world.ChestEmptied(x, y, data)
		s.dirty = true
		return
	}
	s.loot = lootState{active: true, data: data, x: x, y: y, who: -1}
	s.showLootWho()
}

// showLootWho 印「誰要撿？」＋ 隊員清單。
func (s *Scene) showLootWho() {
	s.message = "Who wants loot? " + s.memberMenu()
	if zh := s.cjkExe(exeTable1, strWhoWantsLoot, textlayout.Options{}); zh != "" {
		s.cjk = zh + " " + s.memberMenu()
		s.message = ""
	}
	s.dirty = true
}

// showLootList 印選中隊員的拿取清單（表頭 ＝ 字串 19 的重製版）。
func (s *Scene) showLootList() {
	c := s.world.Party.Members[s.loot.who]
	entries := game.ChestEntries(s.loot.data)
	var en, zh strings.Builder
	fmt.Fprintf(&en, "%s, take an item\r", c.Name)
	if t := s.uiText("loot.take"); len(t) > 0 {
		fmt.Fprintf(&zh, t, c.Name)
		zh.WriteByte('\r')
	}
	for i, e := range entries {
		if i >= 9 {
			break // 與 USE 的清單同一個界線：超過九項還沒分頁
		}
		name := s.itemName(e.ID)
		zhName := s.itemNameCJK(e.ID)
		if e.ID == 0x5E {
			name = fmt.Sprintf("$%d", e.Count)
			zhName = ""
		}
		fmt.Fprintf(&en, "%d) %d %s\r", i+1, e.Count, name)
		if zh.Len() > 0 {
			if zhName == "" {
				zhName = name
			}
			fmt.Fprintf(&zh, "%d) %d %s\r", i+1, e.Count, zhName)
		}
	}
	s.message, s.cjk = en.String(), ""
	if zh.Len() > 0 {
		s.message, s.cjk = "", zh.String()
	}
	s.dirty = true
}

// updateLoot 是撿拾流程的按鍵。
func (s *Scene) updateLoot(in input.Input) (bool, error) {
	if in.Action == input.ActionCancel {
		if s.loot.who >= 0 {
			// 清單那一層的 ESC ＝ 回「誰要撿？」（原版 0xFF → 0x1534B）。
			s.loot.who = -1
			s.showLootWho()
			return true, nil
		}
		// 「誰要撿？」的 ESC ＝ 收手回地圖，東西留在原地。
		s.loot = lootState{}
		s.message, s.cjk = "", ""
		s.dirty = true
		return true, nil
	}
	ch := input.Upper(in.Char)
	if ch < '1' || ch > '9' {
		return true, nil
	}
	n := int(ch - '1')
	if s.loot.who < 0 {
		// 選人：倒下的人不能撿（原版印字串 17 再重問）。
		if n >= len(s.world.Party.Members) || s.world.Party.Members[n] == nil {
			return true, nil
		}
		if s.world.Party.Members[n].Down() {
			s.sayName(strCantGetAny, s.world.Party.Members[n].Name)
			return true, nil
		}
		s.loot.who = n
		s.showLootList()
		return true, nil
	}
	// 拿第 n 件。
	entries := game.ChestEntries(s.loot.data)
	if n >= len(entries) {
		return true, nil
	}
	c := s.world.Party.Members[s.loot.who]
	if !s.world.TakeChestEntry(s.items, s.loot.data, entries[n].At, c) {
		s.say(strCantCarry, textlayout.Options{})
		return true, nil
	}
	if len(game.ChestEntries(s.loot.data)) == 0 {
		// 全部拿完：收尾改寫（布袋消失），回地圖。
		s.world.ChestEmptied(s.loot.x, s.loot.y, s.loot.data)
		s.loot = lootState{}
		s.message, s.cjk = "", ""
		s.dirty = true
		return true, nil
	}
	s.showLootList()
	return true, nil
}

// sayName 印一條 `\x0B`（名字佔位）開頭的字串，名字用參數帶。
func (s *Scene) sayName(n int, name string) {
	s.message = name + strings.ReplaceAll(s.exeStringN(exeTable1, n), "\x0b", "")
	if zh := s.cjkExe(exeTable1, n, textlayout.Options{Name: func() string { return name }}); zh != "" {
		s.cjk, s.message = zh, ""
	}
	s.dirty = true
}
