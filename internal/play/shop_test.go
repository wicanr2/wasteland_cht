package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// 設施選單（docs/spec/25 §6 的驗收條件）。

func mkShop(t *testing.T, kind game.FacilityKind) (*FacilityScene, *game.Party, game.ItemTable) {
	t.Helper()
	rec := make([]byte, 32)
	rec[0] = 0x80 | byte(kind)
	rec[0x03] = 1 // 商店折價指數／醫生每點價格都落在這附近，測試只用得到形狀
	rec[0x04] = 5 // 醫生：每點 CON 的價格
	copy(rec[0x07:], "Store\x00")
	s := &Scene{}
	f := s.EnterFacility(rec)
	if f == nil {
		t.Fatal("設施解不開")
	}
	p := mkParty(20, 20)
	for _, m := range p.Members {
		m.Money = 100
		m.MaxCON = 30
		m.Items = make([]game.Slot, game.ItemSlots)
	}
	items := game.ItemTable{{}, {Price: 40, Stock: 3}, {Price: 10, Stock: game.StockUnlimited}}
	return f, p, items
}

// 驗收 1：背包滿了按 B 回主選單，不離開商店。
func TestBuyWithFullInventoryStaysInShop(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityShop)
	for i := range p.Members[0].Items {
		p.Members[0].Items[i] = game.Slot{ID: 1}
	}
	if !f.Key('B', p, items) {
		t.Fatal("背包滿了不該離開商店")
	}
	if f.state.Step != StepMain {
		t.Errorf("應該留在主選單，得到 %v", f.state.Step)
	}
	if !hasLine(f.Lines, "inventory is full") {
		t.Errorf("應該印背包滿了，得到 %v", f.Lines)
	}
}

// 驗收 2：身上沒有賣得掉的東西按 S 也回主選單。
func TestSellWithNothingStaysInShop(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityShop)
	if !f.Key('S', p, items) {
		t.Fatal("沒東西賣不該離開商店")
	}
	if f.state.Step != StepMain || !hasLine(f.Lines, "anything they want") {
		t.Errorf("應該留在主選單並印訊息，得到 %v / %v", f.state.Step, f.Lines)
	}
}

// 驗收 3：賣一件——槽清空、錢增加、店家庫存 +1（0xFF 不變）。
func TestSellOneItem(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityShop)
	c := p.Members[0]
	c.Items[0] = game.Slot{ID: 1}
	c.Items[1] = game.Slot{ID: 2} // 庫存 0xFF 的那一種
	before := c.Money

	f.Key('S', p, items)
	if f.state.Step != StepSell {
		t.Fatalf("應該進賣的清單，得到 %v", f.state.Step)
	}
	f.Key('1', p, items)

	if c.Items[0].ID != 0 {
		t.Error("賣掉的槽應該清成 0")
	}
	if c.Money <= before {
		t.Errorf("錢應該增加：%d → %d", before, c.Money)
	}
	if got := f.state.Stock[1]; got != 4 {
		t.Errorf("店家庫存應該從 3 變 4，得到 %d", got)
	}

	// 0xFF 那一種賣多少次都不變。
	f.Key('1', p, items)
	if got := f.state.Stock[2]; got != game.StockUnlimited {
		t.Errorf("0xFF 的庫存不該變，得到 %#02x", got)
	}
}

// 驗收 4：錢只夠治 2 點就只治 2 點，而且已治的留著。
func TestHealStopsWhenMoneyRunsOut(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityDoctor)
	c := p.Members[0]
	c.CON, c.MaxCON = 10, 30
	c.Money = 10 // 每點 5 元 → 只夠兩點
	for i := 0; i < 5; i++ {
		f.Key('H', p, items)
	}
	if c.CON != 12 {
		t.Errorf("錢只夠治 2 點，CON 應該是 12，得到 %d", c.CON)
	}
	if c.Money != 0 {
		t.Errorf("錢應該花光，得到 %d", c.Money)
	}
}

// 驗收 5：CON ≤ 0 的人不會被 P 選到。
func TestNextCharSkipsDownedMembers(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityShop)
	p.Members[1].CON = 0
	f.Key('P', p, items)
	if f.state.Who != 0 {
		t.Errorf("第 1 個人倒下了，P 應該留在第 0 個，得到 %d", f.state.Who)
	}
}

// 驗收 6：ESC 一層一層退。
func TestEscapeUnwindsOneLayerAtATime(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityShop)
	p.Members[0].Items[0] = game.Slot{ID: 1}
	f.Key('S', p, items)
	if !f.Key(0x1B, p, items) {
		t.Fatal("清單裡按 ESC 應該回主選單而不是離開")
	}
	if f.state.Step != StepMain {
		t.Errorf("應該回到主選單，得到 %v", f.state.Step)
	}
	if f.Key(0x1B, p, items) {
		t.Error("主選單按 ESC 應該離開設施")
	}
}

func hasLine(lines []string, want string) bool {
	for _, l := range lines {
		if strings.Contains(l, want) {
			return true
		}
	}
	return false
}
