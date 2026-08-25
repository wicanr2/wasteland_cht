package play

import (
	"unicode/utf8"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
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
	// 緩衝區是 **UTF-8**（Big5 只出現在畫的那一刻，`render.DrawRune`）。
	if got := string(s.roster.entry.Text()); got != "沙漠遊俠" {
		t.Fatalf("緩衝區應該是「沙漠遊俠」，得到 %q", got)
	}
	// Enter 建立。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionConfirm}); err != nil {
		t.Fatal(err)
	}
	if s.roster.naming {
		t.Fatal("Enter 之後還在輸入")
	}
	// Enter 之後停在 `Keep this char?`（docs/re/21 §5），答 Y 才寫進記錄。
	if !s.roster.keep {
		t.Fatal("Enter 之後應該停在 Keep this char? 上")
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'Y'}); err != nil {
		t.Fatal(err)
	}
	if n := len(s.World().Party.Members); n != before+1 {
		t.Fatalf("隊伍應該多一個人：%d → %d", before, n)
	}
	c := s.World().Party.Members[before]
	if c.Name != "沙漠遊俠" {
		t.Errorf("名字應該是「沙漠遊俠」，得到 %q", c.Name)
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
	// 記憶體裡的上限放寬到 200 bytes，八個中文字（24 bytes）進得去。
	if _, err := s.Update(runeIn("一二三四五六七八")); err != nil {
		t.Fatal(err)
	}
	if got := string(s.roster.entry.Text()); got != "一二三四五六七八" {
		t.Errorf("八個字應該都收得下，得到 %q", got)
	}
	if n := len(s.roster.entry.Text()); n > input.MaxName {
		t.Errorf("名字超過 %d bytes：%d", input.MaxName, n)
	}
}

// ⚠ **存檔那一格還是 13 bytes**（`docs/re/15` 的角色記錄 `+0x00`–`+0x0C`），
// 放寬的只有記憶體裡的字串。截斷一定要落在 rune 邊界——
// 從中間切下去存檔裡會留半個字，而那個亂碼**寫進玩家的存檔**。
func TestLongNameTruncatesOnRuneBoundary(t *testing.T) {
	const long = "一二三四五六七八" // 24 bytes
	if game.NameFitsSave(long) {
		t.Fatal("24 bytes 不該寫得進 13 bytes 的欄位")
	}
	cut := game.NameForSave(long)
	if len(cut) > game.NameFieldBytes {
		t.Errorf("截出來 %d bytes，超過欄位的 %d", len(cut), game.NameFieldBytes)
	}
	if !utf8.ValidString(cut) {
		t.Errorf("截在字中間了：% x", cut)
	}
	if cut != "一二三四" {
		t.Errorf("13 bytes 放得下四個中文字（12 bytes），得到 %q", cut)
	}
	// 短名字一個 byte 都不動。
	if got := game.NameForSave("Hell Razor"); got != "Hell Razor" {
		t.Errorf("短名字被動到了：%q", got)
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
	if _, err := s.Update(key('Y')); err != nil { // Keep this char? → Yes
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
