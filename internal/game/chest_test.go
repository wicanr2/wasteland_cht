package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 寶箱記錄（docs/re/130）：+0x00/+0x01 是拿完後的改寫對，+0x02 起每兩個
// byte 一組。合成一筆：一件已決定的刀（4）×2、一格空、一件待擲的類別、現金。
func chestRecord() []byte {
	return []byte{
		0x80, 0x00, // 改寫對（bit7 ＝ 不改，只驗流程）
		0x84, 2, // 物品 4，2 件
		0x00, 0, // 空格——**續掃不是結束**
		0x01, 0x83, // 類別 1 待擲（出貨資料就只用 1）；數量 bit7 設 ＝ roll(3)
		0x5E, 20, 0, // 現金：上限 20
		0xFF,
	}
}

func TestChestEntriesSkipsEmptySlots(t *testing.T) {
	got := ChestEntries(chestRecord())
	if len(got) != 3 {
		t.Fatalf("應該有 3 項（空格要跳過、0xFF 收尾），得到 %d：%+v", len(got), got)
	}
	if got[0].ID != 4 || got[0].Count != 2 {
		t.Errorf("第 1 項應該是物品 4 ×2，得到 %+v", got[0])
	}
	if got[2].ID != chestCash {
		t.Errorf("第 3 項應該是現金，得到 %+v", got[2])
	}
}

func TestRollChestDecidesOnce(t *testing.T) {
	tbl := loadItemTable(t)
	w := &World{RNG: rng.New()}
	for i := 0; i < 123; i++ {
		w.RNG.Next() // 推熵：映像初值全零，不推的話 roll 全是退化序列
	}
	data := chestRecord()
	w.RollChest(tbl, data)

	if data[6]&0x80 == 0 {
		t.Fatalf("類別 1 沒被擲成具體物品：%#x", data[6])
	}
	if data[7]&0x80 != 0 || data[7] == 0 || data[7] > 3 {
		t.Fatalf("數量應該是 roll(3) 的結果：%#x", data[7])
	}
	if data[8] != chestCash|0x80 {
		t.Fatalf("現金應該被標成已決定（0xDE）：%#x", data[8])
	}
	if data[9] < 1 || data[9] > 20 {
		t.Fatalf("金額應該是 roll(20)：%d", data[9])
	}
	// 擲出來的物品，類別要真的吻合（sub_15453 的比對是物品資料 +0x03）。
	if d, ok := tbl.Get(data[6] & 0x7F); !ok || d.Class != 1 {
		t.Errorf("擲出的物品 %d 的類別不是 1", data[6]&0x7F)
	}
	// 再跑一次不可以變（lazy 生成，同一個寶箱重看不會變）。
	snapshot := append([]byte(nil), data...)
	w.RollChest(tbl, data)
	for i := range data {
		if data[i] != snapshot[i] {
			t.Fatalf("第二次 RollChest 改了位移 %d：%#x → %#x", i, snapshot[i], data[i])
		}
	}
}

func TestTakeChestEntry(t *testing.T) {
	tbl := loadItemTable(t)
	w := &World{RNG: rng.New()}
	c := &Character{Items: make([]Slot, 30)}
	data := chestRecord()

	// 拿刀（2 件）：第一次拿完剩 1，第二次拿完那一項清成 0。
	for take := 1; take <= 2; take++ {
		if !w.TakeChestEntry(tbl, data, 2, c) {
			t.Fatalf("第 %d 次拿失敗", take)
		}
	}
	if data[2] != 0 {
		t.Errorf("兩件拿完那一項應該清成 0：%#x", data[2])
	}
	n := 0
	for _, s := range c.Items {
		if s.ID == 4 {
			n++
			// 附屬 byte ＝ 容量（發滿），與起始裝備同一條規則。
			if d, ok := tbl.Get(4); ok && s.Value != d.Capacity {
				t.Errorf("附屬 byte 應該是容量 %d，得到 %d", d.Capacity, s.Value)
			}
		}
	}
	if n != 2 {
		t.Fatalf("應該拿到 2 把刀，得到 %d", n)
	}

	// 拿現金：整筆一次拿走，第三個 byte 蓋 0xFF（原版 0x153B9）。
	before := c.Money
	if !w.TakeChestEntry(tbl, data, 8, c) {
		t.Fatal("拿現金失敗")
	}
	if c.Money != before+20 {
		t.Errorf("金錢應該 +20：%d → %d", before, c.Money)
	}
	if data[8] != 0 || data[10] != 0xFF {
		t.Errorf("現金項收尾不對：%#x %#x", data[8], data[10])
	}

	// 物品槽全滿要回 false（原版印 `You can't carry any more.`）。
	full := &Character{Items: make([]Slot, 30)}
	for i := range full.Items {
		full.Items[i] = Slot{ID: 1}
	}
	if w.TakeChestEntry(tbl, data, 6, full) {
		t.Error("槽滿應該拿不了")
	}
}
