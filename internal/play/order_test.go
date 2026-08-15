package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// ORDER 要真的換得動順序（docs/re/93 §1）。
func TestOrderReordersParty(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	before := make([]string, 0, 4)
	for _, m := range s.World().Party.Members {
		if m != nil {
			before = append(before, m.Name)
		}
	}
	if len(before) < 2 {
		t.Skip("出廠隊伍不到兩個人，沒得排")
	}
	if _, err := s.Update(key('O')); err != nil {
		t.Fatal(err)
	}
	if !s.order.active {
		t.Fatal("按 O 沒有開始重排")
	}
	// 倒著選：最後一個先放。
	for i := len(before); i >= 1; i-- {
		if _, err := s.Update(key(byte('0' + i))); err != nil {
			t.Fatal(err)
		}
	}
	if s.order.active {
		t.Fatalf("排完了狀態還在，訊息 %q", s.Message())
	}
	var after []string
	for _, m := range s.World().Party.Members {
		if m != nil {
			after = append(after, m.Name)
		}
	}
	t.Logf("%v → %v", before, after)
	if len(after) != len(before) {
		t.Fatalf("人數變了：%d → %d", len(before), len(after))
	}
	// 順序要真的反過來。
	for i := range before {
		if after[i] != before[len(before)-1-i] {
			t.Errorf("第 %d 個應該是 %s，得到 %s", i, before[len(before)-1-i], after[i])
		}
	}
}

// 已經放回去的人不能再選一次（原版 `jz` 回去重問）。
func TestOrderRejectsUsedSlot(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(key('O')); err != nil {
		t.Fatal(err)
	}
	if !s.order.active {
		t.Skip("這個隊伍沒得排")
	}
	if _, err := s.Update(key('1')); err != nil {
		t.Fatal(err)
	}
	placed := len(s.order.placed)
	// 再按一次 1 —— 那一格已經空了，不該有動作。
	if _, err := s.Update(key('1')); err != nil {
		t.Fatal(err)
	}
	if len(s.order.placed) != placed {
		t.Error("同一個人被排了兩次")
	}
}

// ESC 取消要把原本的順序整份放回去，不留半排好的狀態。
func TestOrderCancelRestores(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	var before []string
	for _, m := range s.World().Party.Members {
		if m != nil {
			before = append(before, m.Name)
		}
	}
	if _, err := s.Update(key('O')); err != nil {
		t.Fatal(err)
	}
	if !s.order.active {
		t.Skip("這個隊伍沒得排")
	}
	if _, err := s.Update(key('2')); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel}); err != nil {
		t.Fatal(err)
	}
	if s.order.active {
		t.Error("ESC 沒有取消")
	}
	var after []string
	for _, m := range s.World().Party.Members {
		if m != nil {
			after = append(after, m.Name)
		}
	}
	if len(after) != len(before) {
		t.Fatalf("取消之後人數變了：%d → %d", len(before), len(after))
	}
	for i := range before {
		if after[i] != before[i] {
			t.Errorf("取消之後順序變了：第 %d 個 %s → %s", i, before[i], after[i])
		}
	}
}
