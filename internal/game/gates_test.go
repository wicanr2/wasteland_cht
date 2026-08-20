package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)


// USE 只試「型別與編號都吻合」的那一條，不是逐條試到成功（docs/re/92 §4）。
func TestUseGateMatchesOnlyTheRightEntry(t *testing.T) {
	r := rng.New()
	// 條件串列：技能 9（難度 3）、物品 12（型別 1）、屬性位移 0x11（型別 2）。
	rec := make([]byte, 0x20)
	rec[0x0A], rec[0x0B] = 0<<5|3, 9    // 技能 9
	rec[0x0C], rec[0x0D] = 1<<5|3, 12   // 物品 12
	rec[0x0E], rec[0x0F] = 2<<5|3, 0x11 // 屬性（Speed 的記錄位移）
	rec[0x10] = 0xFF

	c := &Character{Name: "T", CON: 20}
	c.Attributes[AttrSpeed] = 20
	c.Skills = []Slot{{ID: 9, Value: 10}}
	c.Items = []Slot{{ID: 12, Value: 5}}
	p := &Party{Members: []*Character{c}}

	// 拿不在串列裡的技能去用 → 沒有吻合的條件。
	if hit, _ := p.UseGate(r, rec, c, UseSkill, 7, nil); hit != -1 {
		t.Errorf("技能 7 不在串列裡，卻命中第 %d 條", hit)
	}
	// 技能 9 → 命中第 0 條。
	if hit, _ := p.UseGate(r, rec, c, UseSkill, 9, nil); hit != 0 {
		t.Errorf("技能 9 應該命中第 0 條，得到 %d", hit)
	}
	// 物品 12 → 命中第 1 條（**不是**第 0 條——型別要一起比）。
	if hit, _ := p.UseGate(r, rec, c, UseItem, 12, nil); hit != 1 {
		t.Errorf("物品 12 應該命中第 1 條，得到 %d", hit)
	}
	// 屬性位移 0x11 → 命中第 2 條。
	if hit, _ := p.UseGate(r, rec, c, UseAttribute, 0x11, nil); hit != 2 {
		t.Errorf("屬性 0x11 應該命中第 2 條，得到 %d", hit)
	}
	// 編號對但型別不對 → 不算命中。
	if hit, _ := p.UseGate(r, rec, c, UseItem, 9, nil); hit != -1 {
		t.Errorf("物品 9 不該命中技能那一條，得到 %d", hit)
	}
	// 倒下的人不能用。
	c.CON = -1
	if hit, _ := p.UseGate(r, rec, c, UseSkill, 9, nil); hit != -1 {
		t.Error("CON ≤ 0 的人不該能用 USE")
	}
}

// 物品只認型別 1，自動評估把 1／5／6／7 都當找物品——這個差別要保留。
func TestUseItemOnlyAcceptsTypeOne(t *testing.T) {
	r := rng.New()
	c := &Character{Name: "T", CON: 20, Items: []Slot{{ID: 20, Value: 3}}}
	p := &Party{Members: []*Character{c}}
	for _, typ := range []byte{5, 6, 7} {
		rec := make([]byte, 0x20)
		rec[0x0A], rec[0x0B] = typ<<5|3, 20
		rec[0x0C] = 0xFF
		if hit, _ := p.UseGate(r, rec, c, UseItem, 20, nil); hit != -1 {
			t.Errorf("型別 %d 不該被 USE 的物品那條吃掉（sub_14090 只比 1）", typ)
		}
	}
	// 型別 1 才吃。
	rec := make([]byte, 0x20)
	rec[0x0A], rec[0x0B] = 1<<5|3, 20
	rec[0x0C] = 0xFF
	if hit, _ := p.UseGate(r, rec, c, UseItem, 20, nil); hit != 0 {
		t.Errorf("型別 1 應該命中，得到 %d", hit)
	}
}

// UseKind 的順序照字母表，不照字串 4 的顯示文字（docs/re/92 §2）。
//
// 字串 4 印的是「Item / Skill / Attribute」，字母表卻是 `SIA`——
// 照顯示文字編號會把三條路全部對錯人。
func TestUseKindFollowsKeyTable(t *testing.T) {
	if UseSkill != 0 || UseItem != 1 || UseAttribute != 2 {
		t.Errorf("順序應該是 S=0 I=1 A=2（字母表 ds:A5E8h ＝ 53 49 41），"+
			"得到 S=%d I=%d A=%d", UseSkill, UseItem, UseAttribute)
	}
}

// 條件閘扣 CON 走的是傷害結算，所以**護甲會吸收**——除非記錄 `+0x00` 的
// bit0 設著（`docs/re/122` §2、`docs/re/55` §1）。
//
// ⚠ 這一條擋的不是「有沒有扣血」而是「扣多少」。少了護甲那一段照樣編得過、
// 測得過、玩得動，只是穿不穿護甲在全檔 105 筆閘上完全沒差別。
func TestGatePenaltyGoesThroughArmour(t *testing.T) {
	// 固定值 10 點、減；條件是難度 99 的技能檢定 —— 一定失敗。
	rec := func(msg byte) []byte {
		b := make([]byte, 0x20)
		b[0x00] = msg
		b[0x08] = 0x80 | 0x1D // CON、固定值
		b[0x09] = 0x80 | 10   // 減 10
		b[0x0A], b[0x0B] = 0<<5|31, 9
		b[0x0C] = 0xFF
		return b
	}
	// 護甲 20 顆 d6 —— 期望值 70，吸收一定蓋過 10 點。
	newGuy := func() *Character {
		return &Character{Name: "T", CON: 30, AC: 20, Skills: []Slot{{ID: 9, Value: 1}}}
	}

	// bit0 ＝ 0：照扣護甲 → 10 點被吸光，CON 不動。
	c := newGuy()
	p := &Party{Members: []*Character{c}}
	out := p.EvalGate(rng.New(), rec(0x02), nil)
	if len(out.Failed) != 1 {
		t.Fatalf("難度 31 的檢定應該失敗一次，得到 %d 筆", len(out.Failed))
	}
	if c.CON != 30 {
		t.Errorf("bit0 ＝ 0 時 20 點護甲該吸光 10 點傷害，CON 變成 %d", c.CON)
	}
	if out.Failed[0].Amount != 0 {
		t.Errorf("全部被吸收時回報的量應該是 0，得到 %d", out.Failed[0].Amount)
	}

	// bit0 ＝ 1：跳過護甲 → 整整 10 點。
	c = newGuy()
	p = &Party{Members: []*Character{c}}
	out = p.EvalGate(rng.New(), rec(0x03), nil)
	if c.CON != 20 {
		t.Errorf("bit0 ＝ 1 時護甲不吸收，CON 應該是 20，得到 %d", c.CON)
	}
	if out.Failed[0].Amount != -10 {
		t.Errorf("跳過護甲時該扣滿 10 點，回報 %d", out.Failed[0].Amount)
	}
	if c.PreHurt != 30 {
		t.Errorf("扣血前的值要留在 PreHurt，得到 %d", c.PreHurt)
	}
}
