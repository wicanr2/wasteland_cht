package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/lang"
)

// runes 送一串中文（模擬 IME 提交）。
func runeIn(s string) input.Input {
	return input.Input{Dir: input.DirNone, Runes: []rune(s)}
}

// 角色建立要能打**中文名字**（重製版的擴充；原版只收 ASCII）。
//
// 名字欄是 13 bytes，中文一個字 2 bytes ＝ 放得下 6 個字。
func TestCreateCharacterWithChineseName(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	before := len(s.World().Party.Members)
	s.beginRoster()
	if _, err := s.Update(key('C')); err != nil {
		t.Fatal(err)
	}
	if !s.roster.naming {
		t.Fatal("按 C 沒有進入名字輸入")
	}
	// 打「沙漠遊俠」四個中文字。
	if _, err := s.Update(runeIn("沙漠遊俠")); err != nil {
		t.Fatal(err)
	}
	want, ok := lang.ToBig5("沙漠遊俠")
	if !ok {
		t.Fatal("這四個字編不出 Big5")
	}
	if got := string(s.roster.entry.Text()); got != string(want) {
		t.Fatalf("緩衝區應該是 %d bytes 的 Big5，得到 %d bytes", len(want), len(got))
	}
	// Enter 建立。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionConfirm}); err != nil {
		t.Fatal(err)
	}
	if s.roster.naming {
		t.Fatal("Enter 之後還在輸入")
	}
	if n := len(s.World().Party.Members); n != before+1 {
		t.Fatalf("隊伍應該多一個人：%d → %d", before, n)
	}
	c := s.World().Party.Members[before]
	if c.Name != string(want) {
		t.Errorf("名字應該是那四個字的 Big5，得到 %q", c.Name)
	}
	if c.Level != 1 || c.CON <= 0 {
		t.Errorf("新角色的等級／CON 不對：Lv%d CON%d", c.Level, c.CON)
	}
	// 起始裝備要發下去。
	items := 0
	for _, it := range c.Items {
		if it.ID != 0 {
			items++
		}
	}
	if items == 0 {
		t.Error("新角色身上沒有起始裝備")
	}
	t.Logf("建了 %d bytes 的中文名字、Lv%d CON%d、%d 件裝備",
		len(c.Name), c.Level, c.CON, items)
}

// 13 bytes 放不下第七個中文字。
func TestChineseNameRespectsLimit(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	s.beginRoster()
	if _, err := s.Update(key('C')); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(runeIn("一二三四五六七八")); err != nil {
		t.Fatal(err)
	}
	if n := len(s.roster.entry.Text()); n > input.MaxName {
		t.Errorf("名字超過 %d bytes：%d", input.MaxName, n)
	}
	if n := len(s.roster.entry.Text()); n != 12 {
		t.Errorf("13 bytes 應該只放得下 6 個中文字（12 bytes），得到 %d", n)
	}
}

// 建完之後要寫進存檔的角色記錄槽，而且那一格會被標成有人。
func TestCreateWritesRecordSlot(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	slot, ok := s.freeRecord()
	if !ok {
		t.Skip("沒有空的記錄槽")
	}
	s.beginRoster()
	if _, err := s.Update(key('C')); err != nil {
		t.Fatal(err)
	}
	for _, ch := range "ABC" {
		if _, err := s.Update(key(byte(ch))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionConfirm}); err != nil {
		t.Fatal(err)
	}
	raw, err := s.Save().Record(slot)
	if err != nil {
		t.Fatal(err)
	}
	if raw[recRosterUsed] == 0 {
		t.Error("記錄槽沒有被標成有人")
	}
	if !strings.HasPrefix(string(raw[:3]), "ABC") {
		t.Errorf("記錄裡的名字不對：%q", raw[:8])
	}
	// 同一個槽不該再被當成空的。
	if next, _ := s.freeRecord(); next == slot {
		t.Error("建完之後那一格還是空的")
	}
}
