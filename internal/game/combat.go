package game

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// 單次攻擊的結算（docs/spec/06、docs/re/19、20、21、32）。
//
// 這一層只做「一次攻擊」：回合結構、逃跑、隊形都還沒逆向出來。

// 傷勢門檻（ds:CCCEh，docs/re/15 §3）。
var woundThresholds = [4]int16{-11, -20, -30, -40}

// WoundNames 對應原版的五個狀態字（docs/re/17 §4.4）。
var WoundNames = [6]string{"", "SER", "CRT", "MRT", "COM", "UNC"}

// AttackData 是攻擊資料表的一筆（記錄區標頭 +0x04，8 bytes，docs/re/32 §8.1）。
type AttackData struct {
	XPBase   uint16 // +0x00/+0x01
	DiceN    byte   // +0x03，傷害骰數
	XPMul    byte   // +0x04 的低 4 位（實際倍數要 +1）
	DamBase byte // +0x05 的高 4 位，傷害基底
	Raw     [8]byte

	// 敵方護甲的骰數來自別的路徑（loc_12A92），不在這 8 bytes 裡，
	// 所以由呼叫者傳給 Enemy.TakeDamage，不放進這個結構假裝解過了。
}

// ParseAttackData 拆一筆 8 bytes 的攻擊資料。
func ParseAttackData(b []byte) AttackData {
	var d AttackData
	copy(d.Raw[:], b)
	d.XPBase = uint16(b[0]) | uint16(b[1])<<8
	d.DiceN = b[3]
	d.XPMul = b[4] & 0x0F
	d.DamBase = b[5] >> 4
	return d
}

// KillXP 是擊殺這個敵人給的經驗值：基值 × (倍數 + 1)。
func (d AttackData) KillXP() uint32 {
	return uint32(d.XPBase) * uint32(d.XPMul+1)
}

// SatAdd 是原版那支 16-bit 飽和加（sub_19C2C）：借位夾 0、進位夾 0xFFFF。
func SatAdd(acc uint16, delta int) uint16 {
	v := int(acc) + delta
	if v < 0 {
		return 0
	}
	if v > 0xFFFF {
		return 0xFFFF
	}
	return uint16(v)
}

// AttrModifier 是屬性 → 修正值的階梯（loc_15716，docs/re/21 §2）。
//
//	v ≥ 13 → (v − 12) >> 1
//	v ≤ 8  → floor((v − 9) / 2)      ← 原版用 stc/rcr 做帶號右移，是 floor 不是截斷
//	9–12   → 0
//
// 死區其實是 9–13：13 走正的分支，但 (13−12)>>1 剛好是 0。
func AttrModifier(v byte) int {
	switch {
	case v >= 13:
		return (int(v) - 12) >> 1
	case v <= 8:
		// Go 對負數的 >> 是算術右移，正好等於 floor 除法。
		return (int(v) - 9) >> 1
	default:
		return 0
	}
}

// HitChance 累加出命中判定要比的那個值（sub_1B108 的形狀）。
//
// base 由目標的一個欄位決定（40／50／60），那個欄位的語意未解，
// 所以當參數收進來（docs/spec/06 §6）。
func HitChance(c *Character, base int, skillID byte, fieldValue int, distancePenalty int) uint16 {
	acc := SatAdd(0, base)
	acc = SatAdd(acc, int(c.SkillLevel(skillID))*3)
	acc = SatAdd(acc, fieldValue)
	acc = SatAdd(acc, -distancePenalty)
	if acc > 100 {
		acc = 100 // 原版夾在 100（sub_19C72）
	}
	return acc
}

// PartyHits 是隊伍攻擊的命中判定：roll(1..100) < 累加值。
func PartyHits(r *rng.State, acc uint16) bool { return r.Roll(100) < int(acc) }

// EnemyHits 是敵方攻擊的命中判定：roll(1..100) ≥ 累加值。
// **方向與 PartyHits 相反**，這不是筆誤（docs/re/20 §2）。
func EnemyHits(r *rng.State, acc uint16) bool { return r.Roll(100) >= int(acc) }

// EnemyDamage 是敵方打隊伍的傷害：基底 ＋ N 顆 d6。
func EnemyDamage(r *rng.State, d AttackData) int {
	return r.SumD6(int(d.DamBase), int(d.DiceN))
}

// PartyDamage 是隊伍打敵方的傷害——**五項相加，沒有骰**（docs/re/20 §4.1）。
//
// ⚠ 第一項（距離／射程，sub_15755）還沒解出來，這裡**暫代成 0**。
// 解出來之前，隊伍的傷害會比原版低。
func PartyDamage(c *Character, weaponSkill byte) uint16 {
	const rangeTermUnsolved = 0 // 暫代：sub_15755 未解（docs/spec/06 §6）

	acc := SatAdd(0, rangeTermUnsolved)
	acc = SatAdd(acc, int(c.SkillLevel(weaponSkill))*3)
	acc = SatAdd(acc, AttrModifier(c.Attributes[AttrDexterity]))
	acc = SatAdd(acc, AttrModifier(c.Attributes[AttrStrength]))
	acc = SatAdd(acc, AttrModifier(c.Attributes[AttrLuck]))
	return acc
}

// Absorb 是護甲吸收：N 顆 d6 的和（兩邊機制相同，docs/re/19 §4）。
func Absorb(r *rng.State, n byte) int { return r.SumD6(0, int(n)) }

// TakeDamage 對角色套用傷害：先扣掉護甲，再扣 CON。
// CON **可以為負**；扣血前的值會存進 PreHurt（自然恢復要用）。
func (c *Character) TakeDamage(r *rng.State, damage int) (applied int) {
	applied = damage - Absorb(r, c.AC)
	if applied <= 0 {
		return 0
	}
	c.PreHurt = c.CON
	c.CON -= int16(applied)
	return applied
}

// WoundLevel 把 CON 換成 0–5 的傷勢等級（sub_19A1D）。
// 0 ＝ 沒事，5 ＝ 死亡（CON 剛好是 0）。
func (c *Character) WoundLevel() int {
	if c.CON == 0 {
		return 5
	}
	if c.CON > 0 {
		return 0
	}
	level := 0
	for i, th := range woundThresholds {
		if c.CON <= th {
			level = i + 1
		}
	}
	return level
}

// Enemy 是敵方的一個目標。HP 減到 ≤0 就夾成 0，**不會變負**。
type Enemy struct {
	HP   uint16
	Data AttackData
}

// TakeDamage 對敵人套用傷害，回傳是否被擊殺。
func (e *Enemy) TakeDamage(r *rng.State, damage int, armorDice byte) (applied int, killed bool) {
	applied = damage - Absorb(r, armorDice)
	if applied < 0 {
		applied = 0
	}
	if int(e.HP) <= applied {
		e.HP = 0
		return applied, true
	}
	e.HP -= uint16(applied)
	return applied, false
}
