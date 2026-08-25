package game

// 角色畫面的規則層（docs/re/131）：交給隊友、卸卡彈、指定彈匣裝填、均分現金。
// 裝備／卸下已在 Character.Equip（sub_1949E），丟棄 ＝ 清槽。

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// checkThreshold 是原版檢定門檻（sub_19CAC）：5n ＋ 15（docs/re/32 §3）。
func checkThreshold(n int) int { return 5*n + 15 }

// grudgeMax 是交易難度計數的上限（0x182F1 的 `cmp al, 0Ah`）。
const grudgeMax = 10

// TradeRefused 跑一次「肯不肯交出物品」檢定（sub_18288，docs/re/131 §7）。
//
//	給的人 +0x2E ＝ 0xFF        → 一律拒絕
//	2d6 逢同點續擲 < 5          → 拒絕（與技能檢定同一個 fumble）
//	收的人魅力 ＋ 擲值 ≥ 5×計數＋15 → 成功，1/20 機率計數 −1（下限 0）
//	否則                        → 拒絕，計數 ＋1（上限 10）
//
// **會改給的人的 Grudge**——拒絕越多次越難成功，這是原版的形狀。
// 呼叫端要先套 §7 的閘：+0x29 ＝ 0 或給的人倒下就**不檢定直接給**。
func TradeRefused(r *rng.State, giver, receiver *Character) bool {
	if giver.Grudge == 0xFF {
		return true
	}
	roll := r.PairD6()
	if roll < 5 {
		if giver.Grudge < grudgeMax {
			giver.Grudge++
		}
		return true
	}
	if int(receiver.Attributes[AttrCharisma])+roll >= checkThreshold(int(giver.Grudge)) {
		if r.Roll(0x14) == 1 && giver.Grudge > 0 {
			giver.Grudge--
		}
		return false
	}
	if giver.Grudge < grudgeMax {
		giver.Grudge++
	}
	return true
}

// TradeNeedsCheck 回報這一次交給要不要跑拒絕檢定（0x192E7 與 0x192EF）：
// 給的人 +0x29 非 0 而且還能行動才會拒絕——倒下的人擋不了你拿他的東西。
func TradeNeedsCheck(giver *Character) bool {
	return giver.RecordUsed != 0 && !giver.Down()
}

// TradeItem 把給的人第 slot 格搬給收的人（0x19313 之後那一段）。
// 回傳 false ＝ 收的人物品槽滿（原版印表 2 第 150 條）。
// 搬完原槽清 0；**裝備索引與 AC 照原版不動**（丟棄那條路也一樣）。
func TradeItem(giver, receiver *Character, slot int) bool {
	if slot < 0 || slot >= len(giver.Items) || giver.Items[slot].ID == 0 {
		return true
	}
	free, ok := FirstEmptyItemSlot(receiver.Items)
	if !ok {
		return false
	}
	receiver.Items = putSlot(receiver.Items, free, giver.Items[slot])
	giver.Items[slot] = Slot{}
	return true
}

// UnjamResult 是卸卡彈的結果（sub_19ACD，docs/re/131 §5）。
type UnjamResult int

const (
	UnjamNotJammed UnjamResult = iota // 那一格沒卡彈
	UnjamOK                           // 排除了：附屬 byte 整個清 0（彈藥也沒了）
	UnjamFail                         // 排不掉：卡彈計數 ＋1
)

// Unjam 對第 slot 格的卡彈武器跑排除檢定：
// IQ ＋ 武器技能等級×3 ＋ 2d6 續擲 ≥ 5×卡彈計數 ＋ 15；
// 計數 ≥ 6 或擲值 < 5 直接失敗。
func (c *Character) Unjam(r *rng.State, slot int, tbl ItemTable) UnjamResult {
	if slot < 0 || slot >= len(c.Items) || c.Items[slot].Value&jammedFlag == 0 {
		return UnjamNotJammed
	}
	n := int(c.Items[slot].Value & 0x3F)
	fail := func() UnjamResult {
		// 失敗計數 ＋1（0x19B45：值 < 0x85 才加，卡彈位元原樣帶著）。
		if c.Items[slot].Value < 0x85 {
			c.Items[slot].Value++
		}
		return UnjamFail
	}
	if n >= 6 {
		return fail()
	}
	roll := r.PairD6()
	if roll < 5 {
		return fail()
	}
	acc := int(c.Attributes[AttrIQ]) + roll
	if d, ok := tbl.Get(c.Items[slot].ID); ok {
		acc += int(c.SkillLevel(d.Skill)) * 3
	}
	if acc < checkThreshold(n) {
		return fail()
	}
	c.Items[slot].Value = 0 // 整個清 0：卡彈排除，彈藥也沒了（要重新裝填）
	return UnjamOK
}

// ReloadFrom 用**指定的那一格彈匣**裝填裝備中的武器（sub_196DB）。
// 與戰鬥的 ResolveLoad 差在彈匣是玩家選的，不是掃到的第一件。
// 回傳要印的字串（表別照 ResolveResult 的慣例）。
func (c *Character) ReloadFrom(clipSlot int, tbl ItemTable) ResolveResult {
	w := slotOf(c.EquipIndex)
	if w < 0 || w >= len(c.Items) || c.Items[w].ID == 0 {
		return ResolveResult{}
	}
	if c.Items[w].Value&jammedFlag != 0 {
		return ResolveResult{Message: MsgWeaponJammed, Table2: true}
	}
	d, ok := tbl.Get(c.Items[w].ID)
	if !ok || d.Ammo == 0 {
		return ResolveResult{Message: MsgCantBeReloaded}
	}
	if clipSlot < 0 || clipSlot >= len(c.Items) || c.Items[clipSlot].ID != d.Ammo {
		return ResolveResult{Message: MsgNoMoreClips}
	}
	c.Items[clipSlot] = Slot{} // 彈匣整件吃掉
	c.Items[w].Value = c.Items[w].Value&ammoFlags | d.Capacity&ammoMask
	return ResolveResult{Message: MsgReloads, Redraw: true}
}

// DivideCash 把第 who 個人的現金均分給全隊（0x19165，docs/re/131 §2.1）：
// 每人一份商，**除不盡的零頭歸自己**。
func DivideCash(p *Party, who int) {
	if who < 0 || who >= len(p.Members) || p.Members[who] == nil {
		return
	}
	n := 0
	for _, m := range p.Members {
		if m != nil {
			n++
		}
	}
	if n <= 1 {
		return
	}
	total := p.Members[who].Money
	share := total / uint32(n)
	rem := total % uint32(n)
	p.Members[who].Money = 0
	for _, m := range p.Members {
		if m != nil {
			m.Money += share
		}
	}
	p.Members[who].Money += rem
}
