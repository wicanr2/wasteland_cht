package game

// 角色建立（`sub_1C6C9`，docs/re/21 §5）。

import (
	"sort"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// RollAttribute 是 `sub_1CAD1`：**擲 5 顆 d6，排序後把最大的三顆相加**。
//
// 值域 3–18、期望值 13.43、眾數 14（`docs/re/21` §5.1）。
// ⚠ 不是 3d6——那樣期望值只有 10.5，整個角色強度會偏掉。
func RollAttribute(r *rng.State) byte {
	var d [5]int
	for i := range d {
		d[i] = r.D6()
	}
	sort.Ints(d[:])
	return byte(d[2] + d[3] + d[4])
}

// StartingKits 是三張起始物品清單（`ds:DECFh`／`ds:DED9h`／`ds:DEE3h`，
// `docs/re/21` §5.1）。前兩張二選一、第三張一定發。
type StartingKits struct {
	Pistol45 []byte // 13 M1911A1 ＋ 30 45 clip × 8
	Pistol9  []byte // 16 VP91Z ＋ 32 9mm clip × 8
	Common   []byte // 繩、水壺、撬棍、刀、鏡子、火柴
}

// CreateCharacter 建一個新角色（`sub_1C6C9`）。
//
//	七個屬性各擲一次 RollAttribute
//	MAXCON ＝ CON ＝ RollAttribute() ＋ 18（兩者寫成同一個值）
//	等級 ＝ 1、技能點 ＝ IQ、階級 ＝ "PRIVATE"
//	性別 roll(1..2) 決定拿哪一把起始手槍，第三張清單一定發
//
// name 超過 13 個 byte 會截斷（原版的輸入上限，`docs/re/46`）。
func CreateCharacter(r *rng.State, name string, kits StartingKits, tbl ItemTable) *Character {
	if len(name) > recAttributes {
		name = name[:recAttributes]
	}
	c := &Character{Name: name, Level: 1, Rank: "PRIVATE"}
	// 技能與物品陣列是**固定 30 格**（記錄 +0x80／+0xBD，`docs/re/15`）——
	// 原版整筆清零之後那些格子就在那裡，不是動態長出來的。
	c.Skills = make([]Slot, slotCount)
	c.Items = make([]Slot, slotCount)
	for i := range c.Attributes {
		c.Attributes[i] = RollAttribute(r)
	}
	con := int16(RollAttribute(r)) + 18
	c.CON, c.MaxCON = con, con
	// 技能點 ＝ IQ（`rec[+0x20] ← rec[+0x0F]`）。
	c.SkillPts = c.Attributes[AttrIQ]

	// roll(1..2) 挑前兩張其中一張——**這是原版唯一用到「性別」的地方**。
	first := kits.Pistol45
	if r.Roll(2) == 2 {
		first = kits.Pistol9
	}
	c.GiveStartingKit(first, tbl)
	c.GiveStartingKit(kits.Common, tbl)
	return c
}
