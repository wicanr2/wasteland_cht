package game

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
	Price    uint16 // +0x00/+0x01，基礎價
	Stock    byte   // +0x02，這家店的庫存（0 缺貨、0xFF 無限）
	Class    ItemClass
	ClipSize byte // +0x04，彈匣容量
	Skill    byte // +0x05，使用技能的編號
	// Dice 是「這件裝備值幾顆 d6」：武器是傷害骰數、護甲是 AC。
	// 同一個欄位兩種讀者，語意是同一個（docs/re/45 §3.4）。
	Dice byte // +0x06
	Ammo byte // +0x07，要用的彈藥物品編號，0 ＝ 不用
	Raw  [8]byte

	// +0x03 的低 3 位未解。只有類別 8／9／13 的 bit0 是 1，
	// 看起來像「用一次就沒了」，但沒讀到任何程式碼碰它，所以不給語意。
}

// ParseItemData 拆一筆 8 bytes 的物品資料。
func ParseItemData(b []byte) ItemData {
	var d ItemData
	copy(d.Raw[:], b)
	d.Price = uint16(b[0]) | uint16(b[1])<<8
	d.Stock = b[2]
	d.Class = ItemClass(b[3] >> 3) // ⚠ 右移三次，不是四次（sub_199F1 跳的是 loc_17C6B）
	d.ClipSize = b[4]
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
