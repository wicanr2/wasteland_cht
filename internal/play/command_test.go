package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 指令列的七項與原版字串一一對應（docs/re/91 §1）。
func TestCommandKeysMatchOriginalStrings(t *testing.T) {
	want := map[byte]Command{
		'U': CmdUse, 'E': CmdEnc, 'O': CmdOrder, 'D': CmdDisband,
		'V': CmdView, 'S': CmdSave, 'R': CmdRadio,
	}
	for ch, cmd := range want {
		if got := CommandFor(ch); got != cmd {
			t.Errorf("按 %c 應該對到 %s，得到 %d", ch, CommandNames[cmd], got)
		}
		// 小寫也要通（原版把按鍵大寫化之後才比對，docs/re/46 §4）。
		if got := CommandFor(ch | 0x20); got != cmd {
			t.Errorf("按小寫 %c 應該對到 %s，得到 %d", ch|0x20, CommandNames[cmd], got)
		}
	}
	// 方向鍵不能被當成指令——IKJL 與七個首字母不重疊，這是原版的安排。
	for _, ch := range []byte{'I', 'K', 'J', 'L'} {
		if c := CommandFor(ch); c >= 0 {
			t.Errorf("方向鍵 %c 被當成指令 %s", ch, CommandNames[c])
		}
	}
	if len(commandBar()) == 0 {
		t.Error("指令列是空字串")
	}
}

// RADIO 是升級的唯一入口（docs/re/91）：經驗值夠的人按下去才會升。
//
// 這一條同時擋住「LevelUp() 又變成零呼叫端」——那正是它先前的狀態。
func TestRadioLevelsUp(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	m := s.World().Party.Members[0]
	if m == nil {
		t.Fatal("出廠隊伍第一個人是 nil")
	}
	before := m.Level

	// 給到剛好夠升兩級的經驗值。**不扣經驗值**，所以門檻是累計值。
	need := game.XPForLevel(int(before) + 2)
	m.XP = need
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'R'}); err != nil {
		t.Fatal(err)
	}
	if m.Level != before+2 {
		t.Errorf("經驗值 %d 應該升到等級 %d，得到 %d", need, before+2, m.Level)
	}
	if s.TakeSound() != 4 {
		t.Error("升級要播音效 4（sub_1B8A0，docs/re/44 §6）")
	}
	t.Logf("等級 %d → %d，訊息：%s", before, m.Level, s.Message())

	// 再按一次不該再升——經驗值沒變。
	lvl := m.Level
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'R'}); err != nil {
		t.Fatal(err)
	}
	if m.Level != lvl {
		t.Errorf("經驗值沒增加卻又升了：%d → %d", lvl, m.Level)
	}
	if s.TakeSound() == 4 {
		t.Error("沒升級不該播號角")
	}
}

// 倒下的人不升級（loc_172BE ＝ CON ≤ 0，docs/re/89 §2）。
func TestRadioSkipsDownMembers(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	m := s.World().Party.Members[0]
	m.CON = -5
	before := m.Level
	m.XP = game.XPForLevel(int(before) + 3)
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'R'}); err != nil {
		t.Fatal(err)
	}
	if m.Level != before {
		t.Errorf("CON −5 的人不該升級，%d → %d", before, m.Level)
	}
}

// SAVE 走的是既有的存檔路徑，按下去要真的寫進存檔物件。
func TestSaveCommandWritesSave(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadMap(4, 18, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'S'}); err != nil {
		t.Fatal(err)
	}
	if s.Message() != "Game saved." {
		t.Errorf("存檔訊息不對：%q", s.Message())
	}
	// 存檔槽的地圖編號要跟著換過去（隊伍槽表 +0x0A，docs/re/60 §3）。
	sv := s.Save()
	g := sv.SlotGroups()[0]
	if got := sv.Plain[g.RawIndex+10]; got != 4 {
		t.Errorf("存檔裡的地圖編號應該是 4，得到 %d", got)
	}
}
