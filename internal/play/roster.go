package play

// Ranger Center 的角色管理：`CREATE DELETE PLAY`（設施 3，`docs/re/72` §3）。
//
// 原版的選單字串在 `ds:CE12h`，三個選項各自的流程：
//
//	CREATE  sub_1C6C9：找空的角色記錄槽 → 擲屬性 → 問名字 → 發起始裝備
//	DELETE  把那一筆記錄清掉、從隊伍槽表移除
//	PLAY    離開這個畫面，開始玩
//
// ⚠ **這裡不是存檔處**——存檔走指令列的 `Save`（`docs/re/91`）。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// exeTableRoster 是角色建立／刪除那一組訊息的字串表（`docs/re/17` §3 的第 3 張）。
const exeTableRoster = 3

// rosterState 是角色管理畫面的狀態。
type rosterState struct {
	active bool
	naming bool // 正在輸入名字
	entry  input.TextEntry
	del    bool // 正在選要刪誰
}

// recRosterUsed 是角色記錄「這一格有沒有人」的旗標位移（`docs/re/21` §5）。
const recRosterUsed = 0x29

// 角色管理用到的原版字串編號（**字串表 3**，不是表 1）。
const (
	strNoMoreChars = 1 // `You cannot create any more characters.`
	strDeleteWho   = 5 // `Which player do you want to delete?`
)

// rosterMenu 是三個選項那一行。原版的選單字串在 `ds:CE12h`（`docs/re/72` §3），
// 還沒抽進字串表，所以顯示文字走 `ui:`——**熱鍵字母不跟著翻譯走**。
const rosterMenu = "CREATE  DELETE  PLAY   (C/D/P)"

// beginRoster 進角色管理畫面。
func (s *Scene) beginRoster() {
	s.roster = rosterState{active: true}
	s.sayEN(rosterMenu, "roster.menu")
	s.dirty = true
}

// freeRecord 找一個空的角色記錄槽。
//
// 記錄 0 是全域狀態，角色從 1 起（`docs/spec/05` §2）。
// 空的判準照原版：`+0x29` ＝ 0。
func (s *Scene) freeRecord() (int, bool) {
	for n := 1; n < 8; n++ {
		raw, err := s.save.Record(n)
		if err != nil {
			break
		}
		if len(raw) > recRosterUsed && raw[recRosterUsed] == 0 {
			return n, true
		}
	}
	return 0, false
}

// createCharacter 建一個角色並放進目前這一組。
func (s *Scene) createCharacter(name string) error {
	slot, ok := s.freeRecord()
	if !ok {
		return fmt.Errorf("you cannot create any more characters")
	}
	kits, err := s.rom.StartingKits()
	if err != nil {
		return fmt.Errorf("起始裝備清單：%w", err)
	}
	c := game.CreateCharacter(s.world.RNG, name, game.StartingKits{
		Pistol45: kits[0], Pistol9: kits[1], Common: kits[2],
	}, s.items)

	raw, err := s.save.Record(slot)
	if err != nil {
		return err
	}
	// **整筆清零再寫**（原版 `rec[0..255] ← 0`）——舊角色的殘留欄位
	// 留著會變成新角色身上莫名其妙的狀態。
	for i := range raw {
		raw[i] = 0
	}
	c.StoreTo(raw)
	raw[recRosterUsed] = 1 // 這一格有人了

	// 放進目前這一組的隊伍槽表的第一個空格。
	groups := s.save.SlotGroups()
	g := groups[s.groupID]
	slotTab := s.save.Plain[g.RawIndex : g.RawIndex+14]
	placed := false
	for i := 0; i < 8; i++ {
		if slotTab[i] == 0 {
			slotTab[i] = byte(slot)
			placed = true
			break
		}
	}
	if !placed {
		return fmt.Errorf("這一組隊伍滿了")
	}
	s.world.Party.Members = append(s.world.Party.Members, c)
	return nil
}

// deleteMember 刪掉隊伍裡的第 i 個人（記錄與槽表都清掉）。
func (s *Scene) deleteMember(i int) error {
	members := s.world.Party.Members
	if i < 0 || i >= len(members) || members[i] == nil {
		return fmt.Errorf("第 %d 格沒有人", i)
	}
	groups := s.save.SlotGroups()
	g := groups[s.groupID]
	slotTab := s.save.Plain[g.RawIndex : g.RawIndex+14]

	// 槽表存的是記錄編號，要先找出第 i 個非空槽對應的那個編號。
	seen, recordID := 0, byte(0)
	for j := 0; j < 8; j++ {
		if slotTab[j] == 0 {
			continue
		}
		if seen == i {
			recordID = slotTab[j]
			slotTab[j] = 0
			break
		}
		seen++
	}
	if recordID == 0 {
		return fmt.Errorf("找不到第 %d 個人的記錄編號", i)
	}
	if raw, err := s.save.Record(int(recordID)); err == nil {
		for k := range raw {
			raw[k] = 0 // 整筆清掉，`+0x29` 跟著歸零＝這一格空出來
		}
	}
	s.world.Party.Members = append(members[:i:i], members[i+1:]...)
	return nil
}

// updateRoster 是角色管理畫面的按鍵。
func (s *Scene) updateRoster(in input.Input) (bool, error) {
	// 名字輸入優先——它會吃掉所有字元。
	if s.roster.naming {
		return s.updateNaming(in)
	}
	if s.roster.del {
		return s.updateRosterDelete(in)
	}
	if in.Action == input.ActionCancel {
		s.roster = rosterState{}
		s.message = ""
		s.dirty = true
		return true, nil
	}
	switch input.Upper(in.Char) {
	case 'C':
		if _, ok := s.freeRecord(); !ok {
			s.sayT(exeTableRoster, strNoMoreChars, textlayout.Options{})
			return true, nil
		}
		s.roster.naming = true
		s.roster.entry = input.TextEntry{Max: input.MaxName}
		s.sayEN("Name: ", "roster.name")
		s.dirty = true
	case 'D':
		if len(s.world.Party.Members) == 0 {
			s.sayEN("Nobody to delete.", "roster.nobody")
			return true, nil
		}
		s.roster.del = true
		s.message = "Delete who? " + s.memberMenu()
		if zh := s.cjkExe(exeTableRoster, strDeleteWho, textlayout.Options{}); zh != nil {
			s.cjk = append(append([]byte{}, zh...), []byte(" "+s.memberMenu())...)
			s.message = ""
		}
		s.dirty = true
	case 'P':
		s.roster = rosterState{}
		s.message, s.cjk = "", nil
		s.dirty = true
		s.LeaveFacility()
	}
	return true, nil
}

// updateNaming 收名字（原版上限 13 bytes，`docs/re/46` §5）。
func (s *Scene) updateNaming(in input.Input) (bool, error) {
	k := in.Char
	switch {
	case in.Action == input.ActionCancel:
		k = 0x1B
	case in.Action == input.ActionConfirm:
		k = 0x0D
	}
	// 先收中文（`Runes` 帶完整字元），再走原版的 ASCII 路徑。
	res := input.EntryContinue
	for _, r := range in.Runes {
		if r >= 0x80 {
			s.roster.entry.KeyRune(r, lang.RuneToBig5)
			k = 0 // 這一幀已經吃掉了，不要再當 ASCII 收一次
		}
	}
	if k == 0 {
		s.message = "Name: " + s.nameForDisplay()
		s.dirty = true
		return true, nil
	}
	res = s.roster.entry.Key(k)
	switch res {
	case input.EntryCancel:
		s.roster.naming = false
		s.sayEN(rosterMenu, "roster.menu")
	case input.EntryDone:
		// 名字存的是 Big5 bytes（中文名字就是這樣進記錄的）。
		name := strings.TrimSpace(string(s.roster.entry.Text()))
		s.roster.naming = false
		if name == "" {
			s.sayEN(rosterMenu, "roster.menu")
			break
		}
		if err := s.createCharacter(name); err != nil {
			s.message = "ERROR: " + err.Error()
		} else {
			s.message = name + " joins the Rangers."
			s.cjkFmt("roster.joined", name)
		}
	default:
		s.message = "Name: " + s.nameForDisplay()
	}
	s.dirty = true
	return true, nil
}

// nameForDisplay 是輸入中的名字。中文那幾個 byte 是 Big5，
// 直接當 ASCII 印會是亂碼——所以只在訊息列顯示長度，
// 真正的中文要走 CJK 那條路（`Scene.cjk`）。
func (s *Scene) nameForDisplay() string {
	buf := s.roster.entry.Text()
	if isASCII(buf) {
		return string(buf)
	}
	s.cjk = append(s.cjk[:0], buf...)
	return ""
}

// isASCII 回報這串 bytes 是不是全部可列印 ASCII。
func isASCII(b []byte) bool {
	for _, c := range b {
		if c >= 0x80 {
			return false
		}
	}
	return true
}

// updateRosterDelete 選要刪誰。
func (s *Scene) updateRosterDelete(in input.Input) (bool, error) {
	if in.Action == input.ActionCancel {
		s.roster.del = false
		s.message = "CREATE  DELETE  PLAY   (C/D/P)"
		s.dirty = true
		return true, nil
	}
	ch := input.Upper(in.Char)
	if ch < '1' || ch > '9' {
		return true, nil
	}
	i := int(ch - '1')
	if i >= len(s.world.Party.Members) || s.world.Party.Members[i] == nil {
		return true, nil
	}
	name := s.world.Party.Members[i].Name
	s.roster.del = false
	if err := s.deleteMember(i); err != nil {
		s.message = "ERROR: " + err.Error()
	} else {
		s.message = name + " is gone."
		s.cjkFmt("roster.gone", name)
	}
	s.dirty = true
	return true, nil
}
