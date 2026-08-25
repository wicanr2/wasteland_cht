package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

func pumpRNG(r *rng.State, n int) {
	for i := 0; i < n; i++ {
		r.Next()
	}
}

// 交易拒絕檢定（docs/re/131 §7）：0xFF 一律拒、計數只在 0–10 之間動。
func TestTradeRefused(t *testing.T) {
	r := rng.New()
	pumpRNG(r, 77)
	g := &Character{Grudge: 0xFF}
	if !TradeRefused(r, g, &Character{}) {
		t.Fatal("Grudge 0xFF 應該一律拒絕")
	}
	// 高魅力、計數 0：多試幾次一定看得到成功；計數不出界。
	ok := false
	for i := 0; i < 50 && !ok; i++ {
		giver := &Character{}
		recv := &Character{}
		recv.Attributes[AttrCharisma] = 25
		if !TradeRefused(r, giver, recv) {
			ok = true
		}
		if giver.Grudge > grudgeMax {
			t.Fatalf("計數超過上限：%d", giver.Grudge)
		}
	}
	if !ok {
		t.Fatal("魅力 25 對計數 0，50 次一次都沒成功——公式接錯了")
	}
	// 拒絕要把計數往上推：低魅力狂試，計數必須離開 0。
	giver := &Character{}
	weak := &Character{} // 魅力 0：0 ＋ 2d6續擲 ≥ 15 很難
	moved := false
	for i := 0; i < 50; i++ {
		TradeRefused(r, giver, weak)
		if giver.Grudge > 0 {
			moved = true
			break
		}
	}
	if !moved {
		t.Fatal("被拒絕 50 次計數還是 0——+1 那條沒接")
	}
}

// 倒下的人不會拒絕；+0x29 ＝ 0 也不檢定。
func TestTradeNeedsCheck(t *testing.T) {
	up := &Character{RecordUsed: 1, CON: 10}
	if !TradeNeedsCheck(up) {
		t.Error("站著的隊員應該要檢定")
	}
	down := &Character{RecordUsed: 1, CON: 0}
	if TradeNeedsCheck(down) {
		t.Error("倒下的人不該拒絕")
	}
	npc := &Character{RecordUsed: 0, CON: 10}
	if TradeNeedsCheck(npc) {
		t.Error("+0x29 ＝ 0 不檢定")
	}
}

// 卸卡彈（docs/re/131 §5）：計數 ≥ 6 直接失敗且 +1；成功把附屬 byte 清 0。
func TestUnjam(t *testing.T) {
	tbl := loadItemTable(t)
	r := rng.New()
	pumpRNG(r, 31)

	c := &Character{Items: make([]Slot, 30)}
	c.Items[0] = Slot{ID: 13, Value: 5} // 沒卡彈
	if got := c.Unjam(r, 0, tbl); got != UnjamNotJammed {
		t.Fatalf("沒卡彈應該回 NotJammed，得到 %d", got)
	}

	c.Items[0] = Slot{ID: 13, Value: 0x86} // 卡彈計數 6 ＝ 太死
	if got := c.Unjam(r, 0, tbl); got != UnjamFail {
		t.Fatalf("計數 6 應該直接失敗，得到 %d", got)
	}
	if c.Items[0].Value != 0x86 {
		t.Fatalf("計數 ≥ 5 飽和（0x19B45 的 cmp 0x85），不該再加：%#x", c.Items[0].Value)
	}
	c.Items[0] = Slot{ID: 13, Value: 0x84} // 計數 4：失敗 +1
	if got := c.Unjam(r, 0, tbl); got == UnjamFail && c.Items[0].Value != 0x85 {
		t.Fatalf("計數 4 失敗要 +1：%#x", c.Items[0].Value)
	}

	// IQ 拉滿、計數 0：除了 fumble 一定過；多試到成功為止。
	c.Attributes[AttrIQ] = 30
	okSeen := false
	for i := 0; i < 30 && !okSeen; i++ {
		c.Items[0] = Slot{ID: 13, Value: 0x80}
		if c.Unjam(r, 0, tbl) == UnjamOK {
			okSeen = true
			if c.Items[0].Value != 0 {
				t.Fatalf("成功要把附屬 byte 整個清 0：%#x", c.Items[0].Value)
			}
		}
	}
	if !okSeen {
		t.Fatal("IQ 30 對門檻 15，30 次一次都沒成功")
	}
}

// 指定彈匣裝填（sub_196DB）：彈匣整件吃掉、武器填到容量；卡彈不能裝。
func TestReloadFrom(t *testing.T) {
	tbl := loadItemTable(t)
	c := &Character{Items: make([]Slot, 30)}
	d, _ := tbl.Get(13) // .45 手槍，彈藥 ＝ 30
	c.Items[0] = Slot{ID: 13, Value: 0}
	c.Items[1] = Slot{ID: d.Ammo, Value: d.Capacity}
	c.EquipIndex = 1 // 槽 0（原版索引 ＝ 槽號 ＋1）

	if r := c.ReloadFrom(1, tbl); r.Message != MsgReloads {
		t.Fatalf("應該裝填成功，訊息 %d", r.Message)
	}
	if c.Items[1].ID != 0 {
		t.Error("彈匣要整件吃掉")
	}
	if c.Items[0].Value&0x3F != d.Capacity&0x3F {
		t.Errorf("武器要填到容量 %d，得到 %d", d.Capacity, c.Items[0].Value&0x3F)
	}

	c.Items[0].Value |= 0x80 // 卡彈
	c.Items[2] = Slot{ID: d.Ammo, Value: d.Capacity}
	if r := c.ReloadFrom(2, tbl); r.Message != MsgWeaponJammed {
		t.Errorf("卡彈應該印 152，得到 %d", r.Message)
	}
}

// 均分現金（0x19165）：每人一份商，零頭歸自己。
func TestDivideCash(t *testing.T) {
	p := &Party{Members: []*Character{
		{Money: 100}, {Money: 0}, {Money: 0},
	}}
	DivideCash(p, 0)
	if p.Members[0].Money != 34 || p.Members[1].Money != 33 || p.Members[2].Money != 33 {
		t.Fatalf("100 ÷ 3 應該是 34/33/33，得到 %d/%d/%d",
			p.Members[0].Money, p.Members[1].Money, p.Members[2].Money)
	}
}

// sub_17CFF 的 CF（docs/re/132 §4）。
func TestCellPatchRewrote(t *testing.T) {
	rec := []byte{0, 0, 0x04, 1, 0x84, 1, 0xFD, 0, 0xFE, 0}
	cases := []struct {
		at   int
		want bool
	}{{2, true}, {4, false}, {6, false}, {8, true}, {99, false}}
	for _, tc := range cases {
		if got := CellPatchRewrote(rec, tc.at); got != tc.want {
			t.Errorf("位移 %d：want %v got %v", tc.at, tc.want, got)
		}
	}
}
