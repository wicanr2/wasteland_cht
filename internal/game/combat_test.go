package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 驗收 1：屬性修正的階梯，死區 9–13 必須是 0。
func TestAttrModifierLadder(t *testing.T) {
	// 逐值抄自 docs/re/21 §2 的表（負的那一半是 floor 除法）。
	want := map[byte]int{
		3: -3, 4: -3, 5: -2, 6: -2, 7: -1, 8: -1,
		9: 0, 10: 0, 11: 0, 12: 0, 13: 0,
		14: 1, 15: 1, 16: 2, 17: 2, 18: 3,
	}
	for v := byte(3); v <= 18; v++ {
		if got := AttrModifier(v); got != want[v] {
			t.Errorf("屬性 %d 的修正應該是 %d，得到 %d", v, want[v], got)
		}
	}
}

// 驗收 2：飽和加法。
func TestSatAdd(t *testing.T) {
	if got := SatAdd(10, -100); got != 0 {
		t.Errorf("減到負的應該夾 0，得到 %d", got)
	}
	if got := SatAdd(0xFFF0, 100); got != 0xFFFF {
		t.Errorf("加爆應該夾 0xFFFF，得到 %#x", got)
	}
	if got := SatAdd(100, 23); got != 123 {
		t.Errorf("一般加法錯了：%d", got)
	}
}

// 驗收 3：護甲吸收 ＝ N 顆 d6，平均落在 3.5N 附近。
func TestAbsorbAverage(t *testing.T) {
	r := rng.New()
	const n, iter = 4, 20000
	total := 0
	for i := 0; i < iter; i++ {
		v := Absorb(r, n)
		if v < n || v > 6*n {
			t.Fatalf("%d 顆 d6 的和 %d 超出 %d–%d", n, v, n, 6*n)
		}
		total += v
	}
	avg := float64(total) / iter
	if avg < 3.5*n-0.3 || avg > 3.5*n+0.3 {
		t.Fatalf("%d 顆 d6 的平均應該接近 %.1f，得到 %.2f", n, 3.5*n, avg)
	}
}

// 驗收 4：角色扣血、CON 可為負、PreHurt 有記到、傷勢門檻。
func TestCharacterTakeDamage(t *testing.T) {
	r := rng.New()
	c := &Character{CON: 20, MaxCON: 20, AC: 0}
	applied := c.TakeDamage(r, 30)
	if applied != 30 {
		t.Fatalf("AC 0 應該全額吃下 30 點，得到 %d", applied)
	}
	if c.CON != -10 {
		t.Fatalf("CON 應該變成 −10，得到 %d", c.CON)
	}
	if c.PreHurt != 20 {
		t.Fatalf("PreHurt 應該記到扣血前的 20，得到 %d", c.PreHurt)
	}

	for _, tc := range []struct {
		con  int16
		want int
	}{{5, 0}, {-1, 0}, {-11, 1}, {-19, 1}, {-20, 2}, {-30, 3}, {-40, 4}, {0, 5}} {
		c := &Character{CON: tc.con}
		if got := c.WoundLevel(); got != tc.want {
			t.Errorf("CON %d 的傷勢等級應該是 %d，得到 %d", tc.con, tc.want, got)
		}
	}

	// 護甲擋掉全部時不扣血。
	tough := &Character{CON: 10, MaxCON: 10, AC: 10}
	if got := tough.TakeDamage(r, 1); got != 0 || tough.CON != 10 {
		t.Fatalf("擋下來就不該扣血：applied %d、CON %d", got, tough.CON)
	}
}

// 驗收 5：敵人 HP 夾在 0，不得為負。
func TestEnemyTakeDamage(t *testing.T) {
	r := rng.New()
	e := &Enemy{HP: 10}
	applied, killed := e.TakeDamage(r, 4, 0)
	if applied != 4 || killed || e.HP != 6 {
		t.Fatalf("扣 4 點：applied %d、killed %v、HP %d", applied, killed, e.HP)
	}
	_, killed = e.TakeDamage(r, 100, 0)
	if !killed || e.HP != 0 {
		t.Fatalf("打死之後 HP 應該是 0，得到 %d（killed %v）", e.HP, killed)
	}
}

// 驗收 6：擊殺經驗值 ＝ 基值 × (倍數 + 1)，加到角色身上會飽和。
func TestKillXP(t *testing.T) {
	d := ParseEnemyData([]byte{0x10, 0x00, 0, 3, 0x02, 0x50, 0, 0})
	if d.Base != 0x10 || d.XPMul != 2 || d.DiceN != 3 || d.DamBase != 5 {
		t.Fatalf("拆錯了：%+v", d)
	}
	if got := d.KillXP(); got != 0x10*3 {
		t.Fatalf("擊殺經驗值應該是 %d，得到 %d", 0x10*3, got)
	}

	c := &Character{XP: maxUint24 - 10}
	c.AddXP(d.KillXP())
	if c.XP != maxUint24 {
		t.Fatalf("經驗值應該飽和，得到 %#x", c.XP)
	}
}

// 驗收 7：兩邊的命中判定方向相反。
func TestHitDirectionsAreOpposite(t *testing.T) {
	r := rng.New()
	const acc = 60
	party, enemy := 0, 0
	const iter = 20000
	for i := 0; i < iter; i++ {
		s := r.Snapshot()
		if PartyHits(r, acc) {
			party++
		}
		r.Restore(s)
		if EnemyHits(r, acc) {
			enemy++
		}
	}
	if party+enemy != iter {
		t.Fatalf("同一組骰下兩邊應該剛好互補：隊伍 %d ＋ 敵方 %d ≠ %d",
			party, enemy, iter)
	}
	// 累加值 60 → 隊伍命中率大約 59%（roll 1..100 < 60）。
	rate := float64(party) / iter
	if rate < 0.55 || rate > 0.63 {
		t.Fatalf("累加值 60 的隊伍命中率應該接近 0.59，得到 %.3f", rate)
	}
}

// 命中累加值夾在 100。
func TestHitChanceClamps(t *testing.T) {
	c := &Character{Level: 9, Skills: []Slot{{ID: 1, Value: 30}}}
	if got := HitChance(c, 60, 1, 50, 0); got != 100 {
		t.Fatalf("累加值應該夾在 100，得到 %d", got)
	}
	if got := HitChance(c, 60, 1, 0, 1000); got != 0 {
		t.Fatalf("扣過頭應該夾在 0，得到 %d", got)
	}
}

// 隊伍傷害是五項相加：第一項是武器的傷害骰（會擲骰），另外四項固定。
// 用 0 顆骰的武器把隨機那一項拿掉，就驗得到剩下四項。
func TestPartyDamageFixedTerms(t *testing.T) {
	r := rng.New()
	c := &Character{Skills: []Slot{{ID: 6, Value: 4}}}
	c.Attributes[AttrDexterity] = 18 // ＋3
	c.Attributes[AttrStrength] = 16  // ＋2
	c.Attributes[AttrLuck] = 3       // −3
	w := ParseItemData([]byte{0, 0, 0, 4 << 3, 0, 6, 0, 0}) // 類別 4、技能 6、0 顆骰
	want := uint16(4*3 + 3 + 2 - 3)
	for i := 0; i < 100; i++ {
		if got := PartyDamage(r, c, w, 0); got != want {
			t.Fatalf("固定四項應該是 %d，得到 %d", want, got)
		}
	}
}

// 武器的傷害骰：值域要落在 [N, 6N]，而且技能欄要真的被用到。
func TestPartyDamageRollsWeaponDice(t *testing.T) {
	r := rng.New()
	c := &Character{}
	// 屬性放進死區 9–13，四個固定項全部是 0，剩下的就只有骰。
	c.Attributes[AttrDexterity], c.Attributes[AttrStrength], c.Attributes[AttrLuck] = 10, 10, 10
	w := ParseItemData([]byte{0, 0, 0, 6 << 3, 0, 6, 5, 0}) // 5 顆 d6
	lo, hi, sum := 999, 0, 0
	const n = 20000
	for i := 0; i < n; i++ {
		v := int(PartyDamage(r, c, w, 0))
		sum += v
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	if lo < 5 || hi > 30 {
		t.Fatalf("5 顆 d6 的值域應該在 [5, 30]，實測 [%d, %d]", lo, hi)
	}
	// ⚠ 不要拿「有沒有掃到兩端」當判準：五顆全 6 的機率是 1/7776，
	// 兩萬次也有機會漏掉，那會變成偶爾紅一次的測試。看平均。
	if mean := float64(sum) / n; mean < 16.5 || mean > 18.5 {
		t.Fatalf("5 顆 d6 的平均應該在 17.5 附近，實測 %.2f", mean)
	}
}

// 反戰車武器（類別 8／9）的骰數是 2N − x，其餘類別不受 x 影響。
func TestATWeaponDiceDoubles(t *testing.T) {
	at := ParseItemData([]byte{0, 0, 0, 9<<3 | 1, 1, 11, 13, 0}) // RPG-7：類別 9、13 顆
	if got := WeaponDice(at, 0); got != 26 {
		t.Fatalf("x ＝ 0 時應該加倍成 26，得到 %d", got)
	}
	if got := WeaponDice(at, 5); got != 21 {
		t.Fatalf("x ＝ 5 時應該是 2×13 − 5 ＝ 21，得到 %d", got)
	}
	// N < x 就用原值，不是負數也不是 0。
	if got := WeaponDice(at, 20); got != 13 {
		t.Fatalf("x 大於顆數時應該用原值 13，得到 %d", got)
	}
	// 飽和：顆數大到 2N 超過 255。
	big := ParseItemData([]byte{0, 0, 0, 8 << 3, 0, 11, 200, 0})
	if got := WeaponDice(big, 0); got != 0xFF {
		t.Fatalf("2×200 應該飽和在 255，得到 %d", got)
	}
	// 別的類別不吃這條規則。
	rifle := ParseItemData([]byte{0, 0, 0, 4 << 3, 0, 6, 13, 0})
	if got := WeaponDice(rifle, 0); got != 13 {
		t.Fatalf("非反戰車武器不該加倍，得到 %d", got)
	}
}

// 類別是 +0x03 **右移三次**。移四次的話 RPG-7（0x49）會變成 4，
// 看起來還「像個步槍類別」——所以要釘死這一位。
func TestItemClassShiftsThree(t *testing.T) {
	d := ParseItemData([]byte{0, 0, 0, 0x49, 1, 11, 13, 0})
	if d.Class != ClassATHeavy {
		t.Fatalf("0x49 >> 3 ＝ 9，得到 %d", d.Class)
	}
	if !d.Class.Ranged() || ClassMelee.Ranged() {
		t.Fatal("ds:CD00h 的清單是 2–13：近戰不該算有射程")
	}
	if ClassArmor.Ranged() || ClassAmmo.Ranged() {
		t.Fatal("護甲與彈藥不在清單裡")
	}
}

// 護甲把骰數搬進 AC，武器不會；再選一次同一個槽是卸下（sub_1949E）。
func TestEquipArmorSetsAC(t *testing.T) {
	c := &Character{}
	armor := ParseItemData([]byte{0, 0, 0, 15 << 3, 0, 0, 4, 0}) // Kevlar vest，AC 4
	c.Equip(3, armor)
	if c.AC != 4 || c.ArmorIndex != 3 {
		t.Fatalf("穿上護甲後 AC ＝ %d、槽 ＝ %d", c.AC, c.ArmorIndex)
	}
	rifle := ParseItemData([]byte{0, 0, 0, 4 << 3, 8, 6, 5, 31})
	c.Equip(7, rifle)
	if c.EquipIndex != 7 || c.AC != 4 {
		t.Fatalf("拿武器不該動到 AC：槽 ＝ %d、AC ＝ %d", c.EquipIndex, c.AC)
	}
	c.Equip(3, armor) // 再選一次 ＝ 脫掉
	if c.AC != 0 || c.ArmorIndex != 0 {
		t.Fatalf("脫掉之後 AC ＝ %d、槽 ＝ %d", c.AC, c.ArmorIndex)
	}
}

// 敵方傷害是基底 ＋ Nd6，值域要對。
func TestEnemyDamageRange(t *testing.T) {
	r := rng.New()
	d := ParseEnemyData([]byte{0, 0, 0, 3, 0, 0x50, 0, 0}) // 基底 5、3 顆 d6
	for i := 0; i < 5000; i++ {
		v := EnemyDamage(r, d)
		if v < 5+3 || v > 5+18 {
			t.Fatalf("傷害 %d 超出 8–23", v)
		}
	}
}
