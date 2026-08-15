package play

// 多隊伍：原版最多四支，各自帶座標與地圖（`docs/re/93` §2、§3）。
//
// 槽表在存檔的全域區，四組間隔 14 bytes（`docs/spec/05` §3.1）：
// `Members[8]` 是角色記錄編號（0 ＝ 空槽）、`X`／`Y`／`MapID` 是那一組在哪。
//
// `VIEW` 切到下一組、`DISBAND` 把人分出去，兩支指令都是這個資料結構的介面。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// PartyGroupCount 是隊伍組數上限（原版 `ds:4657h` 比的是 3，所以是四組）。
const PartyGroupCount = 4

// groupCount 數目前有幾組有人。
func (s *Scene) groupCount() int {
	n := 0
	for _, g := range s.save.SlotGroups() {
		if groupSize(g) > 0 {
			n++
		}
	}
	return n
}

// groupSize 數一組裡有幾個人。
func groupSize(g assets.SlotGroup) int {
	n := 0
	for _, id := range g.Members {
		if id != 0 {
			n++
		}
	}
	return n
}

// storeGroup 把目前這一組的座標與地圖寫回槽表。
//
// **切組之前一定要做**——不然回來的時候會站在別組的位置上。
func (s *Scene) storeGroup() {
	groups := s.save.SlotGroups()
	if s.groupID < 0 || s.groupID >= len(groups) {
		return
	}
	slot := s.save.Plain[groups[s.groupID].RawIndex : groups[s.groupID].RawIndex+14]
	slot[8] = s.world.Party.X
	slot[9] = s.world.Party.Y
	slot[10] = byte(s.blockID)
}

// SwitchGroup 切到第 n 組。
//
// 那一組是空的就回錯誤（原版 `sub_160E1` 回 CF ＝ 1，呼叫端接著試下一組）。
func (s *Scene) SwitchGroup(n int) error {
	if n == s.groupID {
		return nil
	}
	party, mapID, err := loadPartyGroup(s.save, n)
	if err != nil {
		return err
	}
	s.storeGroup()
	groups := s.save.SlotGroups()
	g := groups[n]
	s.groupID = n
	s.world.Party = party
	if err := s.LoadMap(int(mapID), g.X, g.Y); err != nil {
		return fmt.Errorf("切到第 %d 組的地圖 %d：%w", n, mapID, err)
	}
	return nil
}

// nextGroup 找下一組有人的（原版從目前這組往後掃，繞回起點就是沒有別組）。
func (s *Scene) nextGroup() (int, bool) {
	groups := s.save.SlotGroups()
	for i := 1; i < len(groups); i++ {
		n := (s.groupID + i) % len(groups)
		if groupSize(groups[n]) > 0 {
			return n, true
		}
	}
	return s.groupID, false
}

// freeGroup 找第一個空的組（`DISBAND` 要用）。
func (s *Scene) freeGroup() (int, bool) {
	for i, g := range s.save.SlotGroups() {
		if groupSize(g) == 0 {
			return i, true
		}
	}
	return 0, false
}

// disbandMember 把第 i 個隊員分出去自成一組。
//
// 原版的規則（`docs/re/93` §2）：**一個人不能分**、已經有四組就不能再分。
// 新的一組站在**原地**——分隊不是傳送。
func (s *Scene) disbandMember(i int) error {
	members := s.world.Party.Members
	if i < 0 || i >= len(members) || members[i] == nil {
		return fmt.Errorf("第 %d 格沒有人", i)
	}
	if len(s.world.Party.Members) <= 1 {
		return fmt.Errorf("一個人不能分隊")
	}
	dst, ok := s.freeGroup()
	if !ok {
		return fmt.Errorf("已經有 %d 支隊伍了", PartyGroupCount)
	}
	groups := s.save.SlotGroups()
	src := groups[s.groupID]

	// 槽表存的是**角色記錄編號**，不是索引——要找到這個人在來源槽表的哪一格。
	var recordID byte
	seen := 0
	for _, id := range src.Members {
		if id == 0 {
			continue
		}
		if seen == i {
			recordID = id
			break
		}
		seen++
	}
	if recordID == 0 {
		return fmt.Errorf("找不到第 %d 個人的記錄編號", i)
	}

	// 從來源移除、放進目的組的第一格。
	srcSlot := s.save.Plain[src.RawIndex : src.RawIndex+14]
	for j := 0; j < 8; j++ {
		if srcSlot[j] == recordID {
			srcSlot[j] = 0
			break
		}
	}
	dstSlot := s.save.Plain[groups[dst].RawIndex : groups[dst].RawIndex+14]
	dstSlot[0] = recordID
	// 新隊伍站在原地（`docs/re/93` §2：分隊不動座標）。
	dstSlot[8] = s.world.Party.X
	dstSlot[9] = s.world.Party.Y
	dstSlot[10] = byte(s.blockID)

	// 目前這一組少一個人。
	s.world.Party.Members = append(members[:i:i], members[i+1:]...)
	return nil
}
