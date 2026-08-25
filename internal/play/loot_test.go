package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 撿拾流程（docs/re/130）：誰要撿 → 逐件拿 → 拿完收尾改寫。
func TestLootFlow(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	// 合成一筆寶箱記錄：一把刀（已決定）＋ 現金 12 元。
	data := []byte{0x80, 0x00, 0x84, 1, 0x5E | 0x80, 12, 0, 0xFF}
	x, y := int(s.World().Party.X), int(s.World().Party.Y)
	s.beginLoot(x, y, data)
	if !s.loot.active {
		t.Fatal("beginLoot 之後撿拾流程沒開")
	}
	// 選 1 號隊員。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: '1'}); err != nil {
		t.Fatal(err)
	}
	if s.loot.who != 0 {
		t.Fatalf("應該選中 1 號，got %d", s.loot.who)
	}
	c := s.World().Party.Members[0]
	knives := func() int {
		n := 0
		for _, sl := range c.Items {
			if sl.ID == 4 {
				n++
			}
		}
		return n
	}
	before, money := knives(), c.Money
	// 拿第 1 件（刀）。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: '1'}); err != nil {
		t.Fatal(err)
	}
	if knives() != before+1 {
		t.Fatalf("刀應該多一把：%d → %d", before, knives())
	}
	// 拿剩下的現金（清單重排後它是第 1 項）。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: '1'}); err != nil {
		t.Fatal(err)
	}
	if c.Money != money+12 {
		t.Fatalf("金錢應該 +12：%d → %d", money, c.Money)
	}
	if s.loot.active {
		t.Fatal("全部拿完應該回地圖")
	}
	if n := len(game.ChestEntries(data)); n != 0 {
		t.Fatalf("記錄裡不該再有東西：%d 項", n)
	}
}

// ESC 兩層語意：清單 → 回「誰要撿？」；誰要撿 → 收手回地圖，東西留著。
func TestLootEscLeavesItems(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte{0x80, 0x00, 0x84, 1, 0xFF}
	s.beginLoot(3, 4, data)
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: '1'}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel}); err != nil {
		t.Fatal(err)
	}
	if !s.loot.active || s.loot.who != -1 {
		t.Fatal("清單層的 ESC 應該回「誰要撿？」")
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel}); err != nil {
		t.Fatal(err)
	}
	if s.loot.active {
		t.Fatal("第二個 ESC 應該收手回地圖")
	}
	if len(game.ChestEntries(data)) != 1 {
		t.Fatal("收手之後東西應該留在原地")
	}
}
