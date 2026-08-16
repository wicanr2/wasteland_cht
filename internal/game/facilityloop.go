package game

// 設施的互動迴圈（docs/spec/18、docs/re/42）。
//
// 規格 09 做的是「一次交易的價格與規則」；這一層是**流程**——
// 怎麼挑物品、怎麼逐點付錢治療。

// 庫存的兩個特殊值（物品表 +0x02，docs/re/42 §4）。
const (
	StockNone      = 0x00 // 缺貨：不出現在列表
	StockUnlimited = 0xFF // 無限：賣東西給它也不會變動
)

// StockEntry 是商店列表的一筆：物品編號 ＋ 這家店的庫存。
type StockEntry struct {
	Item  byte
	Stock byte
}

// InStock 回報這一筆會不會出現在商店列表（庫存非 0）。
func (e StockEntry) InStock() bool { return e.Stock != StockNone }

// AddStock 把一件物品賣回給店家：庫存 +1，**0xFF 不動**。
//
// ⚠ `0xFF` 被跳過是「無限供貨」不是邊界保護（docs/re/42 §9）——
// 讓它加下去會變成 0x00 ＝ 缺貨，整家店的那一項就消失了。
func AddStock(stock byte) byte {
	if stock == StockUnlimited {
		return stock
	}
	return stock + 1
}

// SellListEntry 是「賣」的清單裡的一列。
type SellListEntry struct {
	Slot     int // 物品陣列的槽編號
	Item     byte
	Equipped bool // 裝備中（記錄 +0x1F 或 +0x25）——**標記但不擋著不讓賣**
}

// SellList 列出身上賣得掉的東西（docs/re/42 §3.1）。
//
// sellable 由呼叫端提供（原版是 sub_17AE0 ＋ sub_17AF5 的物品類別過濾，
// 那道過濾還沒逆向）。equipped 同理。
func SellList(items []Slot, sellable func(byte) bool, equipped func(slot int) bool) []SellListEntry {
	var out []SellListEntry
	for i := 0; i < len(items) && i < ItemSlots; i++ {
		if items[i].ID == 0 {
			continue
		}
		if sellable != nil && !sellable(items[i].ID) {
			continue
		}
		out = append(out, SellListEntry{
			Slot: i, Item: items[i].ID,
			Equipped: equipped != nil && equipped(i),
		})
	}
	return out
}

// Sell 賣掉一件東西：從身上移除、加錢、店家庫存 +1。
//
// price 是賣價，由呼叫端提供（算法是 Facility.SellPrice：
// 與買價同一個公式、指數換成商店記錄 +0x04，docs/re/22 §3.1）。
func Sell(c *Character, slot int, price uint32, stock byte) (byte, bool) {
	if c == nil || slot < 0 || slot >= len(c.Items) || c.Items[slot].ID == 0 {
		return stock, false
	}
	c.Items[slot] = Slot{}
	// 金錢是 24-bit（角色記錄 +0x15），夾在上限不繞回去。
	if v := uint32(c.Money) + price; v > 0xFFFFFF {
		c.Money = 0xFFFFFF
	} else {
		c.Money = v
	}
	return AddStock(stock), true
}

// FirstEmptyItemSlot 找物品陣列的第一個空槽（sub_1968A）。
// 滿了就回 false——買的清單根本不會開（docs/re/42 §2）。
//
// ⚠ **上界是 30 槽，不是切片長度。** 拿 `len(items)` 當上界的話，
// 一個帶著 15 件東西的角色會被判成「滿了」——原版掃的是記錄裡固定的
// 30 格（`docs/re/15`），切片短只代表後面那幾格是空的。
func FirstEmptyItemSlot(items []Slot) (int, bool) {
	for i := 0; i < ItemSlots; i++ {
		if i >= len(items) || items[i].ID == 0 {
			return i, true
		}
	}
	return 0, false
}

// BuyListEntry 是「買」的清單裡的一列。
type BuyListEntry struct {
	ID    byte // 物品編號
	Price uint16
	Stock byte
}

// BuyList 列出這家店現在賣得出來的東西（`sub_1C140`，docs/re/42 §2）。
//
// 原版從索引 1 掃到 0x5E（94），每筆 8 bytes，**庫存 `+0x02` 為 0 就不列**
// ——那是「缺貨」不是「沒這個東西」（docs/re/42 §4）。
//
// ⚠ **物品編號的起點很容易弄錯**：原版的表第 0 筆沒有人定址得到，
// 所以 `ParseItemTable` 已經把它跳掉了——**這裡的索引 0 就是物品編號 0**
// （`Fists`），不要再跳一次（docs/re/45 §2、tools/dump_items.py 的說明）。
func BuyList(t ItemTable, price func(base uint16) uint16, stockOf func(id byte) byte) []BuyListEntry {
	var out []BuyListEntry
	for id := 0; id < len(t); id++ {
		stock := t[id].Stock
		if stockOf != nil {
			stock = stockOf(byte(id))
		}
		if stock == StockNone {
			continue
		}
		p := t[id].Price
		if price != nil {
			p = price(p)
		}
		out = append(out, BuyListEntry{ID: byte(id), Price: p, Stock: stock})
	}
	return out
}

// Buy 買一件東西：扣錢、放進第一個空槽、店家庫存 −1。
//
// 回傳 ok ＝ false 表示買不成（錢不夠或背包滿了），此時什麼都不改。
func Buy(c *Character, id byte, price uint32, stock byte) (byte, bool) {
	if c == nil || uint32(c.Money) < price {
		return stock, false
	}
	slot, ok := FirstEmptyItemSlot(c.Items)
	if !ok {
		return stock, false
	}
	c.Money -= price
	c.Items = putSlot(c.Items, slot, Slot{ID: id})
	// 0xFF 是無限，買再多也不會減（與賣的 +1 對稱，docs/re/42 §4）。
	if stock != StockUnlimited && stock > 0 {
		stock--
	}
	return stock, true
}

// HealSession 是逐點治療的一次會話（docs/re/42 §5）。
//
// ⚠ **每一輪重算差額。** 差額算在迴圈外會變成「錢不夠就一點都不治」，
// 而原版是治到沒錢就停在那裡，已治的點數留著。
type HealSession struct {
	Facility Facility
	Char     *Character
}

// Remaining 是還差幾點（MAXCON − CON）。負的（CON 超過上限）當成 0。
func (h HealSession) Remaining() int {
	n := int(h.Char.MaxCON) - int(h.Char.CON)
	if n < 0 {
		return 0
	}
	return n
}

// HealOne 治一點：扣一次「每點價格」、CON +1。
//
// 回傳 false 表示這一次沒治成（錢不夠或已經治滿），reason 是訊息的 key。
func (h HealSession) HealOne() (bool, string) {
	if h.Remaining() == 0 {
		return false, ""
	}
	cost := h.Facility.Price(docHealPer)
	if h.Char.Money < cost {
		return false, msg8(MsgNotEnoughMoney)
	}
	h.Char.Money -= cost
	h.Char.CON++
	return true, ""
}

// 醫生用到的字串編號（字串表 8，docs/re/42 §5–§6）。
const (
	MsgNoDiseases     = 17 // You have no diseases.
	MsgNotEnoughMoney = 18 // You don't have enough money.
)

func msg8(id int) string { return "exe:8:" + itoa(id) }

// Diseases 列出身上有哪幾種病（角色記錄 +0x28 的八個狀態位元）。
// 沒有病時回 nil——解毒選單根本不會開。
func Diseases(c *Character) []int {
	var out []int
	for bit := 0; bit < 8; bit++ {
		if c.Status&(1<<uint(bit)) != 0 {
			out = append(out, bit)
		}
	}
	return out
}
