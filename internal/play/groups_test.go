package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 出廠存檔只有第 0 組有人，另外三組是空的（docs/spec/05 §3.1）。
func TestSaveHasFourGroups(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	groups := s.Save().SlotGroups()
	if len(groups) != PartyGroupCount {
		t.Fatalf("應該有 %d 組槽表，得到 %d", PartyGroupCount, len(groups))
	}
	if n := groupSize(groups[0]); n < 2 {
		t.Fatalf("第 0 組應該有人，得到 %d", n)
	}
	if s.groupCount() != 1 {
		t.Errorf("出廠應該只有一組有人，得到 %d", s.groupCount())
	}
	if _, ok := s.nextGroup(); ok {
		t.Error("只有一組時不該找得到下一組")
	}
}

// DISBAND 把人分出去自成一組，VIEW 就切得過去（docs/re/93 §2、§3）。
//
// 兩支指令是同一個資料結構的兩個介面——**分完才切得過去**，
// 所以這一條把它們串起來驗。
func TestDisbandThenView(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	before := len(s.World().Party.Members)
	if before < 2 {
		t.Skip("隊伍不到兩個人")
	}
	// 分之前切不過去。
	if _, err := s.Update(key('V')); err != nil {
		t.Fatal(err)
	}
	if s.Message() != "No other party." {
		t.Errorf("只有一組時 View 的訊息不對：%q", s.Message())
	}

	// 分出最後一個人。
	if _, err := s.Update(key('D')); err != nil {
		t.Fatal(err)
	}
	if !s.disband {
		t.Fatal("按 D 沒有進入分隊")
	}
	leaving := s.World().Party.Members[before-1].Name
	if _, err := s.Update(key(byte('0' + before))); err != nil {
		t.Fatal(err)
	}
	if s.disband {
		t.Fatal("選完之後狀態還在")
	}
	if n := len(s.World().Party.Members); n != before-1 {
		t.Errorf("分完之後應該剩 %d 人，得到 %d", before-1, n)
	}
	if s.groupCount() != 2 {
		t.Errorf("應該有兩組了，得到 %d", s.groupCount())
	}
	t.Logf("%s 離隊；現在 %d 組", leaving, s.groupCount())

	// 現在切得過去了。
	if _, err := s.Update(key('V')); err != nil {
		t.Fatal(err)
	}
	if s.groupID == 0 {
		t.Fatalf("View 沒有切組，訊息 %q", s.Message())
	}
	if n := len(s.World().Party.Members); n != 1 {
		t.Errorf("新隊伍應該只有一個人，得到 %d", n)
	}
	if got := s.World().Party.Members[0].Name; got != leaving {
		t.Errorf("新隊伍裡應該是 %s，得到 %s", leaving, got)
	}
	// 切回去，原隊伍要完整。
	if _, err := s.Update(key('V')); err != nil {
		t.Fatal(err)
	}
	if s.groupID != 0 {
		t.Errorf("應該切回第 0 組，得到 %d", s.groupID)
	}
	if n := len(s.World().Party.Members); n != before-1 {
		t.Errorf("切回來應該還是 %d 人，得到 %d", before-1, n)
	}
}

// 一個人不能分隊（0x15E77 的 cmp al, 1）。
func TestDisbandRejectsSingle(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	s.World().Party.Members = s.World().Party.Members[:1]
	if _, err := s.Update(key('D')); err != nil {
		t.Fatal(err)
	}
	if s.disband {
		t.Error("一個人時不該能分隊")
	}
	if s.Message() != "Can't disband a single ranger." {
		t.Errorf("訊息不對：%q", s.Message())
	}
}

// 切組進行中方向鍵不走路——分隊選擇也一樣。
func TestDisbandBlocksWalking(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(key('D')); err != nil {
		t.Fatal(err)
	}
	x, y := s.World().Party.X, s.World().Party.Y
	for i := 0; i < 3; i++ {
		if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
			t.Fatal(err)
		}
	}
	if s.World().Party.X != x || s.World().Party.Y != y {
		t.Error("分隊進行中卻走了路")
	}
}
