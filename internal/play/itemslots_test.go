package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 物品陣列是**固定 30 槽、0 ＝ 空**（`docs/re/15`），中間可以有洞：
// 賣掉一件是「把那兩個 byte 清成 0」，不把後面的往前搬（`docs/re/42` §3）。
//
// 這一組門檻守著兩個症狀——它們都不會讓任何既有測試變紅，
// 只會讓玩家在商店裡買不到東西、或是重讀存檔之後少一批物品。

// TestBuyListOpensWithRoomLeft：帶著 15 件東西的出廠 Ranger 走進商店，
// 按 `B` 必須開得出清單。
//
// 抓到的錯是「空槽掃到切片長度為止」——出廠的四個人物品欄剛好連續 15 格，
// 掃到第 15 格就結束，於是回報「背包滿了」，**商店整個不能買**。
func TestBuyListOpensWithRoomLeft(t *testing.T) {
	s := enterVia(t, 10, 30, 25, input.DirUp)
	if !s.InFacility() {
		t.Fatalf("沒有進商店，訊息是 %q", s.Message())
	}
	c := s.World().Party.Members[0]
	used := 0
	for _, it := range c.Items {
		if it.ID != 0 {
			used++
		}
	}
	if used >= game.ItemSlots {
		t.Fatalf("這個人身上 %d 件，本來就滿了，換一個測資", used)
	}
	if _, ok := game.FirstEmptyItemSlot(c.Items); !ok {
		t.Fatalf("身上 %d／%d 件卻找不到空槽", used, game.ItemSlots)
	}

	s.Facility().Key('B', s.World().Party, s.items)
	if got := s.Facility().state.Step; got != StepBuy {
		t.Errorf("按 B 之後停在第 %d 層，預期 StepBuy（%d）；註記是 %q",
			got, StepBuy, s.Facility().note)
	}
}

// TestSoldSlotSurvivesReload：賣掉中間一件之後寫回記錄、再讀出來，
// 洞後面的物品必須還在。
//
// 抓到的錯是「讀到 0 就停」——洞後面那 12 件在重讀時整批消失，
// 而存檔本身是好的，症狀只有玩家看得到。
func TestSoldSlotSurvivesReload(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	raw, err := s.Save().Record(1)
	if err != nil {
		t.Fatal(err)
	}
	c := game.LoadCharacter(raw)
	if len(c.Items) != game.ItemSlots {
		t.Fatalf("讀出來 %d 槽，預期 %d 槽", len(c.Items), game.ItemSlots)
	}
	last := -1
	for i, it := range c.Items {
		if it.ID != 0 {
			last = i
		}
	}
	if last < 2 {
		t.Fatalf("這個人身上只有 %d 槽有東西，測不到「洞後面」", last+1)
	}
	tail := c.Items[last]

	// 挖一個洞：把第 1 槽清掉（＝賣掉那一件）。
	c.Items[1] = game.Slot{}

	// 寫回同一筆記錄再讀出來——這就是存檔重開走的那條路。
	buf := make([]byte, len(raw))
	copy(buf, raw)
	c.StoreTo(buf)
	back := game.LoadCharacter(buf)

	if back.Items[1].ID != 0 {
		t.Errorf("挖掉的那一槽又有東西了：%v", back.Items[1])
	}
	if back.Items[last] != tail {
		t.Errorf("洞後面第 %d 槽變成 %v，原本是 %v", last, back.Items[last], tail)
	}
}
