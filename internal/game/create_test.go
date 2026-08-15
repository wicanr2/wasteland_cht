package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 屬性是 5d6 取最高三顆，不是 3d6（docs/re/21 §5.1）。
//
// 兩者的值域都是 3–18，**期望值差了將近 3 點**——寫成 3d6 不會有任何
// 斷言紅，只是每個角色都弱一截。
func TestRollAttributeIsBestThreeOfFive(t *testing.T) {
	r := rng.New()
	const iter = 20000
	sum, lo, hi := 0, 99, 0
	hist := map[byte]int{}
	for i := 0; i < iter; i++ {
		v := RollAttribute(r)
		sum += int(v)
		hist[v]++
		if int(v) < lo {
			lo = int(v)
		}
		if int(v) > hi {
			hi = int(v)
		}
	}
	mean := float64(sum) / iter
	t.Logf("值域 %d–%d，平均 %.2f（原版 13.43）", lo, hi, mean)
	if lo < 3 || hi > 18 {
		t.Errorf("值域應該是 3–18，得到 %d–%d", lo, hi)
	}
	// 3d6 的期望值是 10.5，5d6 取三是 13.43——門檻放在中間偏上。
	if mean < 13.0 || mean > 13.9 {
		t.Errorf("平均應該接近 13.43，得到 %.2f（3d6 會是 10.5）", mean)
	}
	// 眾數是 14（14.85%）。
	best, bestN := byte(0), 0
	for v, n := range hist {
		if n > bestN {
			best, bestN = v, n
		}
	}
	if best != 14 {
		t.Errorf("眾數應該是 14，得到 %d", best)
	}
}

// 建出來的角色要符合 sub_1C6C9 寫進記錄的那幾格。
func TestCreateCharacterFields(t *testing.T) {
	r := rng.New()
	kits := StartingKits{
		Pistol45: []byte{13, 30, 30},
		Pistol9:  []byte{16, 32, 32},
		Common:   []byte{54, 44},
	}
	tbl := ItemTable{}
	for i := 0; i < 8; i++ {
		c := CreateCharacter(r, "Tester", kits, tbl)
		if c.Level != 1 {
			t.Errorf("等級應該是 1，得到 %d", c.Level)
		}
		if c.Rank != "PRIVATE" {
			t.Errorf("階級應該是 PRIVATE，得到 %q", c.Rank)
		}
		if c.CON != c.MaxCON {
			t.Errorf("CON 與 MAXCON 應該相同：%d ≠ %d", c.CON, c.MaxCON)
		}
		if c.CON < 3+18 || c.CON > 18+18 {
			t.Errorf("CON 應該是 3–18 加 18，得到 %d", c.CON)
		}
		if c.SkillPts != c.Attributes[AttrIQ] {
			t.Errorf("技能點應該等於 IQ：%d ≠ %d", c.SkillPts, c.Attributes[AttrIQ])
		}
		// 起始裝備：一把手槍 ＋ 兩個彈匣，加上第三張清單的兩件 ＝ 5 件。
		// 陣列本身是 30 格，數的是非空的格子。
		n := 0
		for _, it := range c.Items {
			if it.ID != 0 {
				n++
			}
		}
		if n != 5 {
			t.Errorf("應該拿到 5 件東西，得到 %d", n)
		}
		if len(c.Items) != 30 || len(c.Skills) != 30 {
			t.Errorf("陣列應該是 30 格：物品 %d、技能 %d", len(c.Items), len(c.Skills))
		}
		if id := c.Items[0].ID; id != 13 && id != 16 {
			t.Errorf("第一件應該是兩把起始手槍之一，得到 %d", id)
		}
	}
	// 名字超過 13 bytes 要截斷（原版的輸入上限）。
	long := CreateCharacter(r, "0123456789ABCDEFG", kits, tbl)
	if len(long.Name) > 14 {
		t.Errorf("名字沒截斷：%q", long.Name)
	}
}
