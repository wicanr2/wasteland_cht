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
// 給的人是 NPC（+0x29 非 0，docs/re/133 §1）而且還能行動才會拒絕——
// 自己建的角色不會拒絕你，倒下的 NPC 也擋不了你拿他的東西。
func TradeNeedsCheck(giver *Character) bool {
	return giver.NPCFlag != 0 && !giver.Down()
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

// 技能 25 ＝ Medic、32 ＝ Doctor（docs/re/133 §2）。
const (
	SkillMedic  byte = 25
	SkillDoctor byte = 32
)

// FirstAidResult 是急救的結果。
type FirstAidResult int

const (
	FirstAidNotDown FirstAidResult = iota // 目標沒倒（CON ≥ 0）或已死，救不了
	FirstAidOK                            // 成功：負血折半靠向 0，印字串 10
	FirstAidFail                          // 失敗（fumble 或不夠）：印字串 9
)

// FirstAid 對倒下的隊員急救（loc_13D82，docs/re/133 §2）：
//
//	目標 CON ＝ 0（死亡）或 ≥ 0（沒倒）→ 救不了
//	難度 ＝ |CON 低 byte| >> 2（Medic）或 >> 3（Doctor）
//	檢定 ＝ 2d6 續擲（< 5 fumble）＋ 檢定屬性 ＋ 技能×3 ≥ 5×難度＋15
//	成功 ＝ CON（有號）右移一位；施用者照一般規則有機率技能 +1，**不給經驗值**
func (healer *Character) FirstAid(r *rng.State, target *Character,
	skill byte, tbl SkillTable) FirstAidResult {
	if target.CON >= 0 || target.Dead() {
		return FirstAidNotDown
	}
	shift := 2
	if skill == SkillDoctor {
		shift = 3
	}
	// 難度取 CON 低 byte 的絕對值再右移（0x13DBF–0x13DD4）。
	low := int(int8(byte(target.CON)))
	if low < 0 {
		low = -low
	}
	difficulty := low >> shift
	data, ok := SkillData{}, false
	if tbl != nil {
		data, ok = tbl.Skill(skill)
	}
	if !ok {
		return FirstAidFail
	}
	// awardXP ＝ false：sub_18146 不加經驗值（docs/re/133 §2）。
	if !healer.SkillCheck(r, skill, data, difficulty, false).OK {
		return FirstAidFail
	}
	target.CON >>= 1 // 負數的有號右移 ＝ 折半靠向 0（0x13DE4 的 rcr）
	return FirstAidOK
}

// ReorderItems 照 order 給的槽號順序重排物品陣列（0x193AE，docs/re/134）：
// 挑走的順位就是新位置，裝備索引（+0x1F／+0x25 存的是槽號＋1）跟著搬。
// order 必須涵蓋所有非空槽——原版就是「一件一件點到點完」。
func (c *Character) ReorderItems(order []int) {
	out := make([]Slot, slotCount)
	var newEquip, newArmor byte
	n := 0
	for _, slot := range order {
		if slot < 0 || slot >= len(c.Items) || c.Items[slot].ID == 0 {
			continue
		}
		out[n] = c.Items[slot]
		if c.EquipIndex == byte(slot)+1 {
			newEquip = byte(n) + 1
		}
		if c.ArmorIndex == byte(slot)+1 {
			newArmor = byte(n) + 1
		}
		n++
	}
	c.Items = out
	c.EquipIndex, c.ArmorIndex = newEquip, newArmor
}
