package game

// 角色記錄的解讀與寫回（docs/spec/05 §4、docs/re/15、21、31、32、35）。
//
// ⚠ 存檔策略是**改寫不是重建**：解出來的欄位寫回時就地蓋掉那幾個 byte，
// 其餘一律不動。所以這裡沒有「把 Character 序列化成 256 bytes」的函式，
// 只有 Load（讀）與 StoreTo（就地蓋）。

// 角色記錄的欄位位移。
const (
	recName       = 0x00
	recAttributes = 0x0E // 七個屬性
	recMoney      = 0x15 // 24-bit
	recGender     = 0x18
	recNation     = 0x19
	recAC         = 0x1A
	recMaxCON     = 0x1B // 16-bit
	recCON        = 0x1D // 16-bit 有號
	recEquip      = 0x1F
	recSkillPts   = 0x20
	recXP         = 0x21 // 24-bit
	recLevel      = 0x24
	recPreHurt    = 0x26 // 16-bit
	recStatus     = 0x28
	recRank       = 0x32
	// recRankEnd 是階級字串的邊界。**原版的寫入迴圈沒有長度檢查**
	// （`0x1BB6C`：逐字元抄到 NUL 為止，位移只有繞回 0 才停，`docs/re/109` §4），
	// 所以欄位大小是被**下一個已知欄位**擋出來的，不是宣告出來的。
	//
	// ⚠ 別拿出廠值量這一欄：出廠是 `PRIVATE`（7 字元），量出來會是 14——
	// 而實際最長的階級 `Lieutenant Commander` 有 20 個字元，寫到 `+0x46`。
	recRankEnd = 0x4B
	// recMission 是「參與過摧毀 Base Cochise」，recPraised 是「總部已經表揚過」。
	// 兩個 byte 都只用 bit0（`docs/re/96` §5）。
	recMission = 0x4B
	recPraised = 0x4C
	recSkills     = 0x80 // 30 × 2
	recItems      = 0xBD // 30 × 2
	slotCount     = 30
)

// 七個屬性的順序（docs/re/21）。
const (
	AttrStrength = iota
	AttrIQ
	AttrLuck
	AttrSpeed
	AttrAgility
	AttrDexterity
	AttrCharisma
	AttrCount
)

// AttributeNames 照原版選單的順序，中文化走翻譯目錄、不寫死在這裡。
var AttributeNames = [AttrCount]string{
	"Strength", "IQ", "Luck", "Speed", "Agility", "Dexterity", "Charisma",
}

// Slot 是技能或物品陣列裡的一格。
type Slot struct {
	ID    byte // 技能編號或物品編號
	Value byte // 技能是等級；物品的低 6 位是剩餘次數
}

// LoadCharacter 從 256 bytes 的記錄解出規則層用得到的欄位。
// raw 必須是存檔裡那一筆的切片——寫回時要用同一份。
func LoadCharacter(raw []byte) *Character {
	c := &Character{
		Name:    cstring(raw[recName:recAttributes]),
		CON:     int16(le16(raw, recCON)),
		MaxCON:  int16(le16(raw, recMaxCON)),
		PreHurt: int16(le16(raw, recPreHurt)),
		Status:  raw[recStatus],

		Money:      le24(raw, recMoney),
		XP:         le24(raw, recXP),
		Level:      raw[recLevel],
		SkillPts:   raw[recSkillPts],
		AC:         raw[recAC],
		Gender:     raw[recGender],
		Nation:     raw[recNation],
		EquipIndex: raw[recEquip],
		Rank:       cstring(raw[recRank:recRankEnd]),
		Mission:    raw[recMission]&1 != 0,
		Praised:    raw[recPraised]&1 != 0,
	}
	for i := 0; i < AttrCount; i++ {
		c.Attributes[i] = raw[recAttributes+i]
	}
	c.Skills = readSlots(raw, recSkills)
	c.Items = readSlots(raw, recItems)
	return c
}

// StoreTo 把已解欄位就地蓋回記錄，未解區域一個 byte 都不碰。
func (c *Character) StoreTo(raw []byte) {
	putCString(raw[recName:recAttributes], c.Name)
	for i := 0; i < AttrCount; i++ {
		raw[recAttributes+i] = c.Attributes[i]
	}
	put24(raw, recMoney, c.Money)
	raw[recGender] = c.Gender
	raw[recNation] = c.Nation
	raw[recAC] = c.AC
	put16(raw, recMaxCON, uint16(c.MaxCON))
	put16(raw, recCON, uint16(c.CON))
	raw[recEquip] = c.EquipIndex
	raw[recSkillPts] = c.SkillPts
	put24(raw, recXP, c.XP)
	raw[recLevel] = c.Level
	put16(raw, recPreHurt, uint16(c.PreHurt))
	raw[recStatus] = c.Status
	putRank(raw, c.Rank)
	// **只動 bit0**：這兩個 byte 的其餘七位未解，一個都不能碰。
	raw[recMission] = raw[recMission]&^1 | boolBit(c.Mission)
	raw[recPraised] = raw[recPraised]&^1 | boolBit(c.Praised)
	writeSlots(raw, recSkills, c.Skills)
	writeSlots(raw, recItems, c.Items)
}

// readSlots 讀 30 格的陣列。技能是「偶數位移 ＝ 編號」，
// 物品是「奇數位移 ＝ 編號」——原版兩個陣列的起點差一個 byte，
// 所以用同一支就好：base 指到編號那一格。
//
// ⚠ **一定要讀滿 30 格，遇到 0 不能停。** 原版的陣列是**固定 30 槽、
// 0 ＝ 空**，而且中間會有洞：賣掉一件是「把那兩個 byte 清成 0」、
// 不把後面的往前搬（`docs/re/42` §3，已確認），連清單的繪製回呼都寫著
// 「`al ＝ 0` → 這一列跳過」（§3.1）。角色記錄 `+0x1F`／`+0x25` 存的
// 又是**槽號**，往前搬會讓裝備指到別件東西。
//
// 提早停會製造兩個安靜的錯：買東西時找不到空槽（回報「背包滿了」），
// 以及賣掉中間一件之後重讀存檔，洞後面的物品全部消失。
func readSlots(raw []byte, base int) []Slot {
	out := make([]Slot, slotCount)
	for i := 0; i < slotCount; i++ {
		at := base + i*2
		if at+1 >= len(raw) {
			break
		}
		out[i] = Slot{ID: raw[at], Value: raw[at+1]}
	}
	return out
}

// putSlot 把一格寫進 30 槽陣列，切片比 slot 短就先補空槽。
//
// 存檔讀出來的一定是滿 30 格（readSlots），這一支是給
// 手工建出來的角色（測試、`CreateCharacter` 之外的路徑）用的保險。
func putSlot(slots []Slot, slot int, v Slot) []Slot {
	if slot < 0 || slot >= slotCount {
		return slots
	}
	for len(slots) <= slot {
		slots = append(slots, Slot{})
	}
	slots[slot] = v
	return slots
}

func writeSlots(raw []byte, base int, slots []Slot) {
	for i := 0; i < slotCount; i++ {
		at := base + i*2
		if at+1 >= len(raw) {
			return
		}
		if i < len(slots) {
			raw[at], raw[at+1] = slots[i].ID, slots[i].Value
		} else {
			raw[at], raw[at+1] = 0, 0
		}
	}
}

func cstring(b []byte) string {
	for i, ch := range b {
		if ch == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// putRank 寫階級字串，照原版那支迴圈的行為（`0x1BB6C`，`docs/re/109` §4）：
// **抄到 NUL 為止，NUL 之後的 byte 一個都不碰**。
//
// ⚠ 不要改成「整欄清空再寫」。原版沒有清尾巴，所以存檔裡 NUL 之後可能留著
// 上一個字串的殘骸；清掉它就不是 round-trip 了（`CLAUDE.md` §4）。
//
// 太長就截到 NUL 還放得進 `recRankEnd` 之內為止——寫過頭會蓋掉 `+0x4B`
// 的任務旗標（原版沒有這道防護，是它自己的字串表最長只有 20 個字元）。
func putRank(raw []byte, rank string) {
	max := recRankEnd - recRank - 1 // 留一個 byte 給 NUL
	if len(rank) > max {
		rank = rank[:max]
	}
	n := copy(raw[recRank:recRankEnd], rank)
	raw[recRank+n] = 0
}

// putCString 寫入名字並補 NUL，超長就截斷（原版欄位是固定長度）。
func putCString(dst []byte, s string) {
	n := copy(dst, s)
	for i := n; i < len(dst); i++ {
		dst[i] = 0
	}
}

func le16(b []byte, at int) uint16 { return uint16(b[at]) | uint16(b[at+1])<<8 }
func le24(b []byte, at int) uint32 {
	return uint32(b[at]) | uint32(b[at+1])<<8 | uint32(b[at+2])<<16
}

func put16(b []byte, at int, v uint16) {
	b[at], b[at+1] = byte(v), byte(v>>8)
}

func put24(b []byte, at int, v uint32) {
	b[at], b[at+1], b[at+2] = byte(v), byte(v>>8), byte(v>>16)
}

func boolBit(b bool) byte {
	if b {
		return 1
	}
	return 0
}
