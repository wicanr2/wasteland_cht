package game

// 設施：商店、醫生、技能訓練師（docs/spec/09、docs/re/22、29 §5.4、35 §4）。
//
// 這一層只有規則：價格怎麼算、錢夠不夠、成功之後改哪個欄位。
// 選單長怎樣、字怎麼印是呈現層的事。

// FacilityKind 是設施編號（記錄 +0x00 的 bit7 設起來時，低 7 位就是它）。
type FacilityKind int

const (
	FacilityDoctor  FacilityKind = 0
	FacilityShop    FacilityKind = 1
	FacilityTrainer FacilityKind = 2
	// FacilityRoster 是 Ranger Center 的角色管理：`CREATE DELETE PLAY`
	// （`ds:CE12h` 的選單字串，`docs/re/72` §3）。
	// ⚠ **不是存檔處**——存檔走指令列的 `Save`（`docs/re/91`）。
	FacilityRoster FacilityKind = 3
	// FacilityEnding 是**結局**：`ds:A4E0h` 第 4 格指向 `0x1B4F0`，
	// 那一支載 `END.CPA` 並播完四段敘述（`docs/re/96`）。
	// 它沒有店面也沒有圖——走進去遊戲就結束。
	FacilityEnding FacilityKind = 4
	FacilityCount                = 5
)

// 地點名稱在記錄裡的起點，每個設施不一樣（docs/re/29 §5.4）。
var facilityNameAt = [FacilityCount]int{
	FacilityDoctor:  0x07,
	FacilityShop:    0x07,
	FacilityTrainer: 0x04,
	FacilityRoster:  0x03,
	FacilityEnding:  0x03,
}

const facilityNameLen = 13

// 醫生的三種價格在地圖記錄的哪裡（docs/re/35 §4.1）。
const (
	docHealPer = 0x04 // 每點 CON
	docExam    = 0x05
	docCure    = 0x06 // 每種疾病
)

// 商店的兩個價格指數在地圖記錄裡的位置（docs/re/22 §3.1）。
//
// ⚠ **買與賣是同一個公式，只差指數**（`sub_1C1CC` 與 `sub_1C1C2` 共用
// `loc_1C1D3`）。不要為賣另寫一套。
const (
	shopDiscount = 0x03 // 買價指數 → ds:46C3h
	shopSellExp  = 0x04 // 賣價指數 → ds:46C2h
)

// SellExponent 是這家店的賣價指數（地圖記錄 +0x04）。
func (f Facility) SellExponent() byte {
	if len(f.Record) <= shopSellExp {
		return 0
	}
	return f.Record[shopSellExp]
}

// SellPrice 是賣掉一件東西拿到的錢：與買價同一個公式，指數換成 +0x04。
func (f Facility) SellPrice(base uint16) uint32 {
	return uint32(ShopPrice(base, f.SellExponent()))
}

// Facility 是一個設施畫面的資料。
type Facility struct {
	Kind   FacilityKind
	Name   string
	Record []byte
}

// ParseFacility 從 nibble 6 的記錄拆出設施。
// 回傳 ok ＝ false 表示這筆不是設施（bit7 沒設，那是腳本指令）。
func ParseFacility(record []byte) (Facility, bool) {
	if len(record) < 1 || record[0]&0x80 == 0 {
		return Facility{}, false
	}
	kind := FacilityKind(record[0] & 0x7F)
	if kind < 0 || kind >= FacilityCount {
		return Facility{}, false
	}
	f := Facility{Kind: kind, Record: record}
	at := facilityNameAt[kind]
	if at+facilityNameLen <= len(record) {
		f.Name = cstring(record[at : at+facilityNameLen])
	}
	return f, true
}

// ShopPrice 是商店價格：基礎價 − (基礎價 >> n)。
//
// n 來自這筆地圖記錄的 +0x03：0 ＝ 原價、1 ＝ 半價、2 ＝ 75%…
//
// ⚠ **n ＝ 0 要提早回原價**，不能讓公式自己算：`base >> 0` 就是 `base`，
// 相減得 0——指數 0 的店會變成全館免費。原版 `sub_1C1CC` 是
// 「`dl ＝ 0` → 直接 return」，右移迴圈根本不跑（`docs/re/22` §3）。
func ShopPrice(base uint16, discount byte) uint16 {
	if discount == 0 {
		return base
	}
	return base - (base >> discount)
}

// Price 讀設施記錄裡的一個價格欄位。
func (f Facility) Price(at int) uint32 {
	if at >= len(f.Record) {
		return 0
	}
	return uint32(f.Record[at])
}

// HealCost 是把這個角色治滿要多少錢：(MAXCON − CON) × 每點單價。
func (f Facility) HealCost(c *Character) (points int, cost uint32) {
	points = int(c.MaxCON) - int(c.CON)
	if points <= 0 {
		return 0, 0
	}
	return points, uint32(points) * f.Price(docHealPer)
}

// Heal 是「一次治滿」的整筆版本：付不起就整筆失敗。
//
// ⚠ **原版沒有這一支。** 原版是逐點付（HealSession.HealOne，docs/re/42 §5），
// 錢不夠就停在中途。這一支留著給不需要互動的呼叫端（測試、腳本），
// **玩家路徑一律走 HealSession**。
func (f Facility) Heal(c *Character) (ok bool, reason string) {
	points, cost := f.HealCost(c)
	if points <= 0 {
		return false, ReasonNoHealNeeded
	}
	if c.Money < cost {
		return false, ReasonNoMoney
	}
	c.Money -= cost
	c.CON = c.MaxCON
	return true, ""
}

// Cure 付錢治一種病（bit 0–7）。
func (f Facility) Cure(c *Character, bit int) (ok bool, reason string) {
	if bit < 0 || bit > 7 {
		return false, ReasonNoSuchDisease
	}
	mask := uint8(1) << bit
	if c.Status&mask == 0 {
		return false, ReasonNoSuchDisease
	}
	cost := f.Price(docCure)
	if c.Money < cost {
		return false, ReasonNoMoney
	}
	c.Money -= cost
	c.Status &^= mask
	return true, ""
}

// Exam 付檢查費。原版檢查之後才會顯示身上有哪些病，
// 規則層只負責扣款——要顯示什麼由呈現層決定。
func (f Facility) Exam(c *Character) (ok bool, reason string) {
	cost := f.Price(docExam)
	if c.Money < cost {
		return false, ReasonNoMoney
	}
	c.Money -= cost
	return true, ""
}

// Buy 在商店買一件物品。base 是物品資料的基礎價。
func (f Facility) Buy(c *Character, itemID byte, base uint16) (ok bool, reason string) {
	price := uint32(ShopPrice(base, f.Record[shopDiscount]))
	if c.Money < price {
		return false, ReasonNoMoney
	}
	slot, ok := FirstEmptyItemSlot(c.Items)
	if !ok {
		return false, ReasonInventoryFull
	}
	c.Money -= price
	c.Items = putSlot(c.Items, slot, Slot{ID: itemID, Value: 1})
	return true, ""
}
