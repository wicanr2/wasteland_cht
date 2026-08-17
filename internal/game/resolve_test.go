package game

import "testing"

// 一張夠用的假物品表：1 ＝ 手槍（吃彈匣 2、容量 8）、2 ＝ 彈匣、3 ＝ 護甲（AC 4）、
// 4 ＝ 撬棍（近戰，不吃彈匣）。
func testItems() ItemTable {
	mk := func(class ItemClass, cap, dice, ammo byte) ItemData {
		return ItemData{Class: class, Capacity: cap, Dice: dice, Ammo: ammo}
	}
	// ⚠ `Get(id)` 是**直接索引**（`t[id]`），不是 id−1——
	// 原版那張表的第 0 筆沒有人定址得到，`ParseItemTable` 已經把它丟掉了。
	return ItemTable{
		mk(ClassGeneral, 0, 0, 0), // id 0：定址不到，放個佔位
		mk(ClassPistol, 8, 3, 2),  // id 1 手槍，吃彈匣 2、容量 8
		mk(ClassAmmo, 1, 0, 0),    // id 2 彈匣
		mk(ClassArmor, 0, 4, 0),   // id 3 護甲，AC 4
		mk(ClassMelee, 0, 2, 0),   // id 4 撬棍，近戰
	}
}

func charWith(items ...Slot) *Character {
	c := &Character{Items: make([]Slot, ItemSlots)}
	copy(c.Items, items)
	return c
}

// 換裝備是**切換**：選到已經裝著的那一格就卸下（`docs/re/107` §2）。
func TestResolveWeaponEquipsAndUnequips(t *testing.T) {
	tbl := testItems()
	c := charWith(Slot{ID: 1, Value: 8}, Slot{ID: 4})

	if r := ResolveWeapon(c, 1, tbl); r.Message != MsgSwapsEquipment || !r.Redraw {
		t.Fatalf("裝上應該印 102 並重畫，得到 %+v", r)
	}
	if c.EquipIndex != 1 {
		t.Fatalf("EquipIndex 應該是 1（1 起算），得到 %d", c.EquipIndex)
	}
	// 再選同一格 → 卸下。
	ResolveWeapon(c, 1, tbl)
	if c.EquipIndex != 0 {
		t.Errorf("再選一次應該卸下，EquipIndex ＝ %d", c.EquipIndex)
	}
	// 換另一件。
	ResolveWeapon(c, 2, tbl)
	if c.EquipIndex != 2 {
		t.Errorf("換第二格失敗，EquipIndex ＝ %d", c.EquipIndex)
	}
}

// 護甲走 `+0x25` 並連帶設 AC，**與武器欄互不干擾**。
func TestResolveWeaponArmourUsesItsOwnSlot(t *testing.T) {
	tbl := testItems()
	c := charWith(Slot{ID: 1, Value: 8}, Slot{ID: 3})
	ResolveWeapon(c, 1, tbl) // 武器
	ResolveWeapon(c, 2, tbl) // 護甲
	if c.EquipIndex != 1 {
		t.Errorf("穿護甲不該動到武器欄，EquipIndex ＝ %d", c.EquipIndex)
	}
	if c.ArmorIndex != 2 || c.AC != 4 {
		t.Errorf("護甲欄 %d、AC %d，預期 2 與 4", c.ArmorIndex, c.AC)
	}
	// 卸下護甲要把 AC 一起歸零（`sub_194E8` 的 `bl ← 0x1A`）。
	ResolveWeapon(c, 2, tbl)
	if c.ArmorIndex != 0 || c.AC != 0 {
		t.Errorf("卸護甲之後 %d／AC %d，預期 0／0", c.ArmorIndex, c.AC)
	}
}

// 裝填的四條路（`docs/re/107` §3），順序不能換。
func TestResolveLoadGates(t *testing.T) {
	tbl := testItems()

	// 沒裝備武器：靜靜結束，不印訊息。
	if r := ResolveLoad(charWith(Slot{ID: 2}), tbl); r.Message != 0 {
		t.Errorf("沒裝備武器不該印訊息，得到 %d", r.Message)
	}

	// 卡彈（bit7）→ 表 2 第 152 條。
	c := charWith(Slot{ID: 1, Value: 0x80}, Slot{ID: 2})
	c.EquipIndex = 1
	r := ResolveLoad(c, tbl)
	if r.Message != MsgWeaponJammed || !r.Table2 {
		t.Errorf("卡彈應該是表 2 第 152 條，得到 %+v", r)
	}
	if c.Items[1].ID == 0 {
		t.Error("卡彈時不該吃掉彈匣")
	}

	// 近戰武器不吃彈匣。
	c = charWith(Slot{ID: 4}, Slot{ID: 2})
	c.EquipIndex = 1
	if r := ResolveLoad(c, tbl); r.Message != MsgCantBeReloaded {
		t.Errorf("近戰武器應該印 66，得到 %d", r.Message)
	}

	// 身上沒有那種彈匣。
	c = charWith(Slot{ID: 1, Value: 0})
	c.EquipIndex = 1
	if r := ResolveLoad(c, tbl); r.Message != MsgNoMoreClips {
		t.Errorf("沒彈匣應該印 65，得到 %d", r.Message)
	}
}

// 裝填成功：彈數填到容量，**彈匣整件消失**（不是扣 1）。
func TestResolveLoadConsumesTheWholeClip(t *testing.T) {
	tbl := testItems()
	// ⚠ 高 2 位裡的 **bit7 是「卡彈」**，拿它當「隨便一個旗標」會走進卡彈那條路。
	c := charWith(Slot{ID: 1, Value: 0x40 | 2}, Slot{ID: 2, Value: 1})
	c.EquipIndex = 1

	r := ResolveLoad(c, tbl)
	if r.Message != MsgReloads || !r.Redraw {
		t.Fatalf("成功應該印 104 並重畫，得到 %+v", r)
	}
	if got := c.Items[0].Value; got != 0x40|8 {
		t.Errorf("附屬 byte ＝ %#x，預期 %#x（高 2 位保留、低 6 位填容量）", got, 0x40|8)
	}
	if c.Items[1] != (Slot{}) {
		t.Errorf("彈匣該整件消失，得到 %+v", c.Items[1])
	}
}

// 迴避在結算階段**只印一句話**——不要在這裡補效果，
// 它的效果全在命中門檻的基礎值（`docs/re/38` §2）。
func TestResolveEvadeOnlyPrints(t *testing.T) {
	if r := ResolveEvade(); r.Message != MsgEvades || r.Redraw {
		t.Errorf("迴避應該只印 82，得到 %+v", r)
	}
}

// 槽號是 1 起算的（`sub_19AC8`）。0 ＝ 沒有裝備，不是第 0 格。
func TestSlotIsOneBased(t *testing.T) {
	if slotOf(0) != -1 {
		t.Error("0 應該是「沒有裝備」")
	}
	if slotOf(1) != 0 {
		t.Error("槽 1 對到 Items[0]")
	}
	if slotOf(ItemSlots) != ItemSlots-1 {
		t.Error("最後一格算錯")
	}
	if slotOf(ItemSlots+1) != -1 {
		t.Error("超界要擋掉")
	}
}
