package game

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// 物品資料表（docs/re/45、docs/spec/06 §3.1）。
//
// ⚠ **這張表不在執行檔裡，在存檔區**，而且每個存檔槽各一份——
// 它是可變的遊戲狀態，不是唯讀資料。賣東西會把 `Stock` 加一，那個改動要存回去。
// 讀取與解密在 `internal/assets`；這一層只認欄位。

// ItemClass 是物品類別（資料 +0x03 的高 5 位）。
type ItemClass byte

// 表裡實際出現的 18 個類別。名稱是本專案依表的內容取的，原版沒有這些字串。
const (
	ClassMelee      ItemClass = 1  // 近戰／徒手
	ClassPistol     ItemClass = 2  // 手槍與投擲
	ClassFlameCarb  ItemClass = 3  // 火焰噴射器、卡賓槍
	ClassRifle      ItemClass = 4  // 步槍
	ClassSMG        ItemClass = 6  // 衝鋒槍
	ClassAssault    ItemClass = 7  // 突擊步槍
	ClassATLight    ItemClass = 8  // 反戰車（輕）
	ClassATHeavy    ItemClass = 9  // 反戰車（重）
	ClassLaserPist  ItemClass = 10 // 雷射手槍
	ClassEnergyMid  ItemClass = 11 // 能量（中）
	ClassEnergyHigh ItemClass = 12 // 能量（重）
	ClassExplosive  ItemClass = 13 // 爆裂物
	ClassAmmo       ItemClass = 14 // 彈藥
	ClassArmor      ItemClass = 15 // 護甲
	ClassGeneral    ItemClass = 16 // 一般物品
	ClassTrinket    ItemClass = 17 // 可賣雜物
	ClassPlot       ItemClass = 18 // 劇情物品
)

// rangedClasses 是 ds:CD00h 那張清單：**有射程的武器類別**。
// 原版逐 byte 比對到負值為止，順序無關，所以這裡用集合。
// 近戰（1）不在裡面，彈藥／護甲／雜物（≥14）也不在。
var rangedClasses = map[ItemClass]bool{
	2: true, 3: true, 4: true, 5: true, 6: true, 7: true,
	8: true, 9: true, 10: true, 11: true, 12: true, 13: true,
}

// Ranged 回報這個類別算不算「有射程的武器」（sub_19D2F）。
func (c ItemClass) Ranged() bool { return rangedClasses[c] }

// ItemData 是物品資料表的一筆（8 bytes）。
type ItemData struct {
	Price uint16 // +0x00/+0x01，基礎價
	Stock byte   // +0x02，這家店的庫存（0 缺貨、0xFF 無限）
	Class ItemClass
	// Capacity 是「裝滿時能用幾次」：槍是彈匣容量、Match 是 40、Rope 是 1。
	// 發裝備／換彈／買入時會寫進物品槽的附屬 byte（docs/re/21 §5.1）。
	Capacity byte // +0x04
	Skill    byte // +0x05，使用技能的編號
	// Dice 是「這件裝備值幾顆 d6」：武器是傷害骰數、護甲是 AC。
	// 同一個欄位兩種讀者，語意是同一個（docs/re/45 §3.4）。
	Dice byte // +0x06
	Ammo byte // +0x07，要用的彈藥物品編號，0 ＝ 不用
	Raw  [8]byte

	// +0x03 的低 3 位**沒有讀取端**：全檔讀這個欄位的只有 sub_199F1（>> 3）
	// 與呼叫它的 sub_15453（寶箱生成）。資料裡 bit0 只出現在類別 8／9／13，
	// 看起來像「用一次就沒了」，但程式沒讀它——原樣保留，不給語意
	// （docs/re/45 §6，與敵人資料 +0x04 高 4 位同一種情形）。
}

// ParseItemData 拆一筆 8 bytes 的物品資料。
func ParseItemData(b []byte) ItemData {
	var d ItemData
	copy(d.Raw[:], b)
	d.Price = uint16(b[0]) | uint16(b[1])<<8
	d.Stock = b[2]
	d.Class = ItemClass(b[3] >> 3) // ⚠ 右移三次，不是四次（sub_199F1 跳的是 loc_17C6B）
	d.Capacity = b[4]
	d.Skill = b[5]
	d.Dice = b[6]
	d.Ammo = b[7]
	return d
}

// ItemTable 是一整張物品資料表。索引 0 對到表的第 1 筆
// （`sub_17AE0` 的基址是表首 ＋ 8），所以第 0 筆沒有人定址得到。
type ItemTable []ItemData

// ParseItemTable 拆整張表。傳進來的是解密後的 760 bytes。
func ParseItemTable(b []byte) ItemTable {
	n := len(b) / 8
	if n <= 1 {
		return nil
	}
	out := make(ItemTable, 0, n-1)
	for i := 1; i < n; i++ {
		out = append(out, ParseItemData(b[i*8:(i+1)*8]))
	}
	return out
}

// Get 取索引 id 的那一筆；超出範圍回零值與 false。
func (t ItemTable) Get(id byte) (ItemData, bool) {
	if int(id) >= len(t) {
		return ItemData{}, false
	}
	return t[id], true
}

// StartingKit 是建角色時發的三張物品清單（ds:DECFh／DED9h／DEE3h，
// docs/re/21 §5.1）。清單裡是物品編號，`0xFF` 結束。
//
// roll(1..2) 從前兩張挑一張手槍組，第三張一定發。
var (
	KitPistol45 = []byte{13, 30, 30, 30, 30, 30, 30, 30, 30} // .45 手槍 ＋ 八個彈匣
	KitPistol9  = []byte{16, 32, 32, 32, 32, 32, 32, 32, 32} // 9mm 手槍 ＋ 八個彈匣
	KitCommon   = []byte{54, 44, 45, 4, 49, 52}              // 繩、水壺、撬棍、刀、鏡子、火柴
)

// GiveStartingKit 把一張清單發給角色（sub_1C9DE）。
//
// 每一格佔物品陣列的一個槽：編號 ＋ **附屬 byte ← 物品表的 Capacity**，
// 也就是「發滿」。表裡查不到的編號照樣發，附屬 byte 給 0——
// 原版沒有這道檢查，這裡只是不讓它越界。
func (c *Character) GiveStartingKit(list []byte, tbl ItemTable) {
	for _, id := range list {
		if id == 0xFF {
			return
		}
		slot, ok := FirstEmptyItemSlot(c.Items)
		if !ok {
			return
		}
		var full byte
		if d, ok := tbl.Get(id); ok {
			full = d.Capacity
		}
		c.Items[slot] = Slot{ID: id, Value: full}
	}
}

// RollStartingPistol 照原版擲 1d2 決定拿哪一把起始手槍。
func RollStartingPistol(r *rng.State) []byte {
	if r.Roll(2) == 1 {
		return KitPistol45
	}
	return KitPistol9
}
