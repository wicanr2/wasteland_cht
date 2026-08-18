package game

// 戰鬥指令的**結算階段**（`docs/re/107`）。
//
// ⚠ **指令有兩張跳表。** `ds:A43Bh` 是下令階段（檢查前提、開清單、選參數，
// `internal/game/handlers.go`），`ds:A568h` 是結算階段——**動作在這裡才真的發生**。
// 只接前一張的話玩家按 `W`／`L` 會什麼都不發生，而畫面上看不出任何異狀。

// ResolveResult 是一支結算處理程式的結果。
//
// Message 是原版字串表 1 的編號（0 ＝ 不印）；Table2 為真時那個編號屬於表 2。
type ResolveResult struct {
	Message byte
	Table2  bool
	// Redraw 為真表示原版跑完會重畫那個人的名片行（`sub_17033`）——
	// 裝備與彈數變了，名單上的 `AMM`／`WEAPON` 兩欄要跟著換。
	Redraw bool
}

// 結算階段用到的字串編號（`docs/re/107` §2–§6）。
const (
	MsgSwapsEquipment byte = 102 // `\x0B swaps equipment.`
	MsgEvades         byte = 82  // `\x0B evades.`
	MsgReloads        byte = 104 // `\x0B reloads.`
	MsgUses           byte = 105 // `\x0B uses...`
	MsgNoRoomInRoster byte = 95  // `No room in roster.`
	// MsgWeaponJammed 在**表 2**（表 1 第 56 條是同一句話的另一份）。
	MsgWeaponJammed byte = 152
)

// ammoMask 是物品附屬 byte 的低 6 位 ＝ 剩餘次數；高 2 位是旗標。
//
// bit7 ＝ **這把武器卡彈**（`sub_196DB` 的 `0x196EE` 直讀，`docs/re/107` §3）。
const (
	ammoMask   byte = 0x3F
	ammoFlags  byte = 0xC0
	jammedFlag byte = 0x80
)

// Jammed 回答這一格的武器卡彈了沒（附屬 byte 的 bit7）。
//
// 呈現層要它：卡彈的武器名在名單上是**反白**的（`docs/re/111` §1）。
func Jammed(s Slot) bool { return s.Value&jammedFlag != 0 }

// slotOf 把 1-based 的槽號換成 Items 的索引；0 或超界回 −1。
//
// ⚠ **槽號是 1-based**（`sub_19AC8`：位移 ＝ 槽號 × 2 ＋ 0xBB，槽 1 → `+0xBD`）。
// 當成 0-based 會整批差一格，而每一格都還是合法的物品——**不會有任何錯誤訊息**。
func slotOf(n byte) int {
	i := int(n) - 1
	if n == 0 || i >= ItemSlots {
		return -1
	}
	return i
}

// ResolveWeapon 是換裝備的結算（`sub_1949E`，`docs/re/107` §2）。
//
// **同一支既裝也卸**：選到已經裝著的那一格就卸下來，否則裝上。
// 類別 15（護甲）走 `+0x25` 並連帶設 `+0x1A` 的 AC，其餘走 `+0x1F`
// ——所以武器與護甲可以同時裝著。
func ResolveWeapon(c *Character, slot byte, tbl ItemTable) ResolveResult {
	res := ResolveResult{Message: MsgSwapsEquipment, Redraw: true}
	i := slotOf(slot)
	if i < 0 || i >= len(c.Items) {
		return res
	}
	switch slot {
	case c.EquipIndex: // 卸下武器
		c.EquipIndex = 0
		return res
	case c.ArmorIndex: // 卸下護甲，AC 跟著歸零
		c.ArmorIndex = 0
		c.AC = 0
		return res
	}
	d, ok := tbl.Get(c.Items[i].ID)
	if !ok {
		return res
	}
	if d.Class == ClassArmor {
		c.ArmorIndex = slot
		c.AC = d.Dice
		return res
	}
	c.EquipIndex = slot
	return res
}

// ResolveLoad 是裝填的結算（`0x13228` ＋ `sub_196DB`，`docs/re/107` §3）。
//
// 四種結果各有各的訊息，順序不能換：
//
//	沒裝備武器          → 靜靜結束（原版先試著自動裝一件，裝不到就什麼都不印）
//	武器卡彈（bit7）    → 表 2 第 152 條
//	這把武器不吃彈匣    → 表 1 第 66 條
//	身上沒有那種彈匣    → 表 1 第 65 條
//	都過                → 填滿並印第 104 條
//
// ⚠ **裝填會把彈匣整件消耗掉**（`sub_19AB6` 把那一槽的兩個 byte 都清 0），
// 不是「扣掉幾發」——彈匣是一次性物品。
func ResolveLoad(c *Character, tbl ItemTable) ResolveResult {
	w := slotOf(c.EquipIndex)
	if w < 0 || w >= len(c.Items) || c.Items[w].ID == 0 {
		return ResolveResult{} // 沒裝備武器：靜靜結束
	}
	if c.Items[w].Value&jammedFlag != 0 {
		return ResolveResult{Message: MsgWeaponJammed, Table2: true}
	}
	d, ok := tbl.Get(c.Items[w].ID)
	if !ok || d.Ammo == 0 {
		return ResolveResult{Message: MsgCantBeReloaded}
	}
	clip := -1
	for i := range c.Items {
		if c.Items[i].ID == d.Ammo {
			clip = i
			break
		}
	}
	if clip < 0 {
		return ResolveResult{Message: MsgNoMoreClips}
	}
	c.Items[clip] = Slot{} // 整件吃掉
	c.Items[w].Value = c.Items[w].Value&ammoFlags | d.Capacity&ammoMask
	return ResolveResult{Message: MsgReloads, Redraw: true}
}

// ResolveEvade 是迴避的結算（`0x13223`）：**只印一句話**。
//
// 下令階段的處理程式也是空的（`docs/re/38` §2）——迴避的效果全在命中門檻的
// 基礎值 60，兩張跳表的這一格都不做事是對的，不要在這裡補一個效果。
func ResolveEvade() ResolveResult { return ResolveResult{Message: MsgEvades} }
