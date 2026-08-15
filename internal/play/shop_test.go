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

// 買（docs/spec/25 §2、docs/re/42 §2、§4）。

func TestBuyListSkipsOutOfStock(t *testing.T) {
	f, _, _ := mkShop(t, game.FacilityShop)
	items := game.ItemTable{
		{Price: 10, Stock: 0},                    // 缺貨：不列
		{Price: 20, Stock: 2},                    // 列
		{Price: 30, Stock: game.StockUnlimited},  // 列
	}
	list := f.buyList(items)
	if len(list) != 2 {
		t.Fatalf("缺貨的不該列出來，得到 %d 筆：%v", len(list), list)
	}
	if list[0].ID != 1 || list[1].ID != 2 {
		t.Errorf("編號對不上：%v", list)
	}
	// 折價指數 1 → 半價（基礎價 − 基礎價>>1）。
	if list[0].Price != 10 {
		t.Errorf("折價後應該是 10，得到 %d", list[0].Price)
	}
}

func TestBuyOneDeductsMoneyAndStock(t *testing.T) {
	f, p, _ := mkShop(t, game.FacilityShop)
	items := game.ItemTable{{}, {Price: 40, Stock: 2}}
	c := p.Members[0]
	c.Money = 100

	f.Key('B', p, items)
	if f.state.Step != StepBuy {
		t.Fatalf("應該進買的清單，得到 %v", f.state.Step)
	}
	f.Key('1', p, items)

	if c.Money != 80 { // 40 − 40>>1 ＝ 20
		t.Errorf("錢應該剩 80，得到 %d", c.Money)
	}
	if c.Items[0].ID != 1 {
		t.Errorf("第一個空槽應該放進物品 1，得到 %d", c.Items[0].ID)
	}
	if got := f.state.Stock[1]; got != 1 {
		t.Errorf("店家庫存應該從 2 變 1，得到 %d", got)
	}
}

func TestBuyWithoutMoneyChangesNothing(t *testing.T) {
	f, p, _ := mkShop(t, game.FacilityShop)
	items := game.ItemTable{{}, {Price: 400, Stock: 2}}
	c := p.Members[0]
	c.Money = 10

	f.Key('B', p, items)
	f.Key('1', p, items)

	if c.Money != 10 || c.Items[0].ID != 0 {
		t.Error("錢不夠時不該扣錢也不該給東西")
	}
	if !hasLine(f.Lines, "enough money") {
		t.Errorf("應該印錢不夠，得到 %v", f.Lines)
	}
}

// 賣價與買價是同一個公式、指數不同（docs/re/22 §3.1）。
func TestSellUsesItsOwnExponent(t *testing.T) {
	f, p, _ := mkShop(t, game.FacilityShop)
	// mkShop 的記錄：+0x03 ＝ 1（買，半價）、+0x04 ＝ 5（賣）。
	f.Facility.Record[0x04] = 5
	items := game.ItemTable{{}, {Price: 64, Stock: 1}}
	c := p.Members[0]
	c.Items[0] = game.Slot{ID: 1}
	before := c.Money

	f.Key('S', p, items)
	f.Key('1', p, items)

	// 64 − 64>>5 ＝ 64 − 2 ＝ 62
	if got := c.Money - before; got != 62 {
		t.Errorf("賣價應該用 +0x04 的指數（62），得到 %d", got)
	}
}

// 訓練師（docs/spec/25 §3.1、docs/re/52）。

func TestTrainerRefusesWithoutSkillPoints(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityTrainer)
	f.Skills = []TrainableSkill{{ID: 3, Data: game.SkillData{BaseCost: 2, IQ: 3}}}
	c := p.Members[0]
	c.SkillPts = 0
	c.Attributes[game.AttrIQ] = 20

	if !f.Key('1', p, items) {
		t.Fatal("點數不夠不該離開設施")
	}
	if c.SkillLevel(3) != 0 {
		t.Error("點數 0 不該學得起來")
	}
	if !hasLine(f.Lines, "no skill points") {
		t.Errorf("應該印沒有技能點數，得到 %v", f.Lines)
	}
}

func TestTrainerLearnsAndDeductsPoints(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityTrainer)
	f.Skills = []TrainableSkill{{ID: 3, Data: game.SkillData{BaseCost: 2, IQ: 3}}}
	c := p.Members[0]
	c.SkillPts = 5
	c.Attributes[game.AttrIQ] = 20
	c.Skills = make([]game.Slot, game.ItemSlots)

	f.Key('1', p, items)
	if c.SkillLevel(3) != 1 {
		t.Errorf("應該學到等級 1，得到 %d", c.SkillLevel(3))
	}
	if c.SkillPts != 3 { // 基礎費用 2
		t.Errorf("技能點應該剩 3，得到 %d", c.SkillPts)
	}
}

// 費用大於點數時不扣點也不升級，而且**留在選單裡**。
func TestTrainerKeepsMenuWhenTooExpensive(t *testing.T) {
	f, p, items := mkShop(t, game.FacilityTrainer)
	f.Skills = []TrainableSkill{{ID: 3, Data: game.SkillData{BaseCost: 7, IQ: 3}}}
	c := p.Members[0]
	c.SkillPts = 1
	c.Attributes[game.AttrIQ] = 20
	c.Skills = make([]game.Slot, game.ItemSlots)

	if !f.Key('1', p, items) {
		t.Fatal("費用不夠不該離開設施")
	}
	if c.SkillPts != 1 || c.SkillLevel(3) != 0 {
		t.Error("費用不夠時不該扣點也不該升級")
	}
}
