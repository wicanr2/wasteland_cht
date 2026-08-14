package game

import "testing"

func slots(ids ...byte) []Slot {
	out := make([]Slot, ItemSlots)
	for i, id := range ids {
		out[i] = Slot{ID: id}
	}
	return out
}

// 驗收 1：庫存 0 不上架；0xFF 賣多少次都還是 0xFF。
func TestStock(t *testing.T) {
	if (StockEntry{Item: 3, Stock: StockNone}).InStock() {
		t.Error("庫存 0 不該上架")
	}
	if !(StockEntry{Item: 3, Stock: 1}).InStock() {
		t.Error("庫存 1 該上架")
	}
	if got := AddStock(4); got != 5 {
		t.Errorf("庫存 4 賣一件應該變 5，得到 %d", got)
	}
	// ⚠ 0xFF 是「無限」不是邊界——加下去會變成 0 ＝ 缺貨，整項就消失了。
	s := byte(StockUnlimited)
	for i := 0; i < 10; i++ {
		s = AddStock(s)
	}
	if s != StockUnlimited {
		t.Errorf("無限供貨賣十次之後還該是 %#x，得到 %#x", StockUnlimited, s)
	}
}

// 驗收 2／3：賣掉會清槽、加錢、庫存 +1；裝備中的標記但賣得掉。
func TestSell(t *testing.T) {
	c := &Character{Items: slots(7, 0, 9), Money: 100}
	list := SellList(c.Items, nil, func(slot int) bool { return slot == 2 })
	if len(list) != 2 {
		t.Fatalf("應該列出 2 件，得到 %d", len(list))
	}
	if list[0].Slot != 0 || list[1].Slot != 2 {
		t.Errorf("槽編號不對：%+v", list)
	}
	if list[0].Equipped || !list[1].Equipped {
		t.Errorf("裝備標記不對：%+v", list)
	}

	// 裝備中的照樣賣得掉。
	stock, ok := Sell(c, 2, 30, 4)
	if !ok || stock != 5 {
		t.Fatalf("賣掉裝備中的應該成功、庫存變 5，得到 stock=%d ok=%v", stock, ok)
	}
	if c.Items[2].ID != 0 {
		t.Error("賣掉之後那個槽該是空的")
	}
	if c.Money != 130 {
		t.Errorf("錢應該變 130，得到 %d", c.Money)
	}
	if _, ok := Sell(c, 2, 30, 4); ok {
		t.Error("空槽不該賣得出去")
	}
	if _, ok := Sell(c, 99, 30, 4); ok {
		t.Error("越界的槽不該賣得出去")
	}

	// 過濾器只讓某些東西上架。
	if n := len(SellList(c.Items, func(id byte) bool { return id == 9 }, nil)); n != 0 {
		t.Errorf("9 已經賣掉了，不該還列得出來：%d", n)
	}
}

// 驗收 4：物品陣列滿了就沒有空槽。
func TestFirstEmptyItemSlot(t *testing.T) {
	if slot, ok := FirstEmptyItemSlot(slots(1, 2)); !ok || slot != 2 {
		t.Errorf("第一個空槽應該是 2，得到 %d（ok=%v）", slot, ok)
	}
	full := make([]Slot, ItemSlots)
	for i := range full {
		full[i] = Slot{ID: 1}
	}
	if _, ok := FirstEmptyItemSlot(full); ok {
		t.Error("滿了不該找得到空槽")
	}
}

func healFacility(perPoint byte) Facility {
	rec := make([]byte, 16)
	rec[docHealPer] = perPoint
	return Facility{Record: rec}
}

// 驗收 5／6：逐點治療每輪重算；錢不夠停在中途；不會超過上限。
func TestHealSession(t *testing.T) {
	c := &Character{CON: 5, MaxCON: 20, Money: 25}
	h := HealSession{Facility: healFacility(10), Char: c}

	if h.Remaining() != 15 {
		t.Fatalf("一開始應該差 15 點，得到 %d", h.Remaining())
	}
	healed := 0
	for {
		ok, reason := h.HealOne()
		if !ok {
			if reason != msg8(MsgNotEnoughMoney) {
				t.Fatalf("停下來的理由應該是錢不夠，得到 %q", reason)
			}
			break
		}
		healed++
	}
	// 25 塊、每點 10 塊 → 只治得起 2 點，剩 5 塊。
	if healed != 2 || c.CON != 7 || c.Money != 5 {
		t.Fatalf("應該治 2 點到 CON 7、剩 5 塊；得到 治了 %d 點、CON %d、錢 %d",
			healed, c.CON, c.Money)
	}
	// ⚠ 已治的點數要留著——差額算在迴圈外會變成「錢不夠就一點都不治」。
	if h.Remaining() != 13 {
		t.Errorf("治完之後應該還差 13 點，得到 %d", h.Remaining())
	}

	// 治滿就停，不會超過上限。
	rich := &Character{CON: 19, MaxCON: 20, Money: 1000}
	hr := HealSession{Facility: healFacility(1), Char: rich}
	for i := 0; i < 10; i++ {
		hr.HealOne()
	}
	if rich.CON != 20 {
		t.Errorf("不該治超過上限，得到 CON %d", rich.CON)
	}
	if rich.Money != 999 {
		t.Errorf("治滿之後不該繼續扣錢，得到 %d", rich.Money)
	}
	if ok, reason := hr.HealOne(); ok || reason != "" {
		t.Errorf("治滿之後應該無聲結束，得到 ok=%v reason=%q", ok, reason)
	}
}

// 驗收 7：沒有病時列不出東西。
func TestDiseases(t *testing.T) {
	if got := Diseases(&Character{Status: 0}); got != nil {
		t.Errorf("沒有病應該回 nil，得到 %v", got)
	}
	got := Diseases(&Character{Status: 0b1000_0101})
	want := []int{0, 2, 7}
	if len(got) != len(want) {
		t.Fatalf("應該有 %v，得到 %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("應該有 %v，得到 %v", want, got)
		}
	}
}
