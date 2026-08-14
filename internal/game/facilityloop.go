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
// price 是賣價，由呼叫端提供——**賣價公式（sub_1C1C2）還沒逆向**，
// 不要拿買價公式硬套（docs/spec/18 §3）。
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
func FirstEmptyItemSlot(items []Slot) (int, bool) {
	for i := 0; i < len(items) && i < ItemSlots; i++ {
		if items[i].ID == 0 {
			return i, true
		}
	}
	return 0, false
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
