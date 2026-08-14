package game

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// 單次攻擊的結算（docs/spec/06、docs/re/19、20、21、32）。
//
// 這一層只做「一次攻擊」：回合結構、逃跑、隊形都還沒逆向出來。

// 傷勢門檻（ds:CCCEh，docs/re/15 §3）。
var woundThresholds = [4]int16{-11, -20, -30, -40}

// WoundNames 對應原版的五個狀態字（docs/re/17 §4.4）。
var WoundNames = [6]string{"", "SER", "CRT", "MRT", "COM", "UNC"}

// EnemyData 是敵人資料表的一筆（記錄區標頭 +0x04，8 bytes，docs/re/37 §3.1）。
//
// **一張表，每張地圖一份**，索引就是地圖記錄裡的敵人型別。
// 第 0 筆恆為全零 ＝ 沒有敵人。
type EnemyData struct {
	// Base 同一組 byte 有兩個用途：生怪時是基礎血量、擊殺時是經驗值基值
	// （sub_12AAB 與 sub_15A18 讀的是同樣的 +0x00/+0x01）。
	Base    uint16 // +0x00/+0x01
	Speed   byte   // +0x02，行動值欄位（× 8 進行動值，docs/spec/12）
	DiceN   byte   // +0x03，傷害骰數
	XPMul   byte   // +0x04 的低 4 位（實際倍數要 +1）
	DamBase byte   // +0x05 的高 4 位，傷害基底
	Raw     [8]byte

	// 敵方護甲的骰數來自別的路徑（loc_12A92），不在這 8 bytes 裡，
	// 所以由呼叫者傳給 Enemy.TakeDamage，不放進這個結構假裝解過了。
	//
	// +0x04 的高 4 位、+0x06、+0x07 未解——只留在 Raw 裡原樣保存。
}

// ParseEnemyData 拆一筆 8 bytes 的敵人資料。
func ParseEnemyData(b []byte) EnemyData {
	var d EnemyData
	copy(d.Raw[:], b)
	d.Base = uint16(b[0]) | uint16(b[1])<<8
	d.Speed = b[2]
	d.DiceN = b[3]
	d.XPMul = b[4] & 0x0F
	d.DamBase = b[5] >> 4
	return d
}

// Empty 回報這一筆是不是「沒有敵人」（第 0 筆恆為全零）。
func (d EnemyData) Empty() bool {
	for _, b := range d.Raw {
		if b != 0 {
			return false
		}
	}
	return true
}

// KillXP 是擊殺這個敵人給的經驗值：基值 × (倍數 + 1)。
func (d EnemyData) KillXP() uint32 {
	return uint32(d.Base) * uint32(d.XPMul+1)
}

// RollHP 擲一隻敵人的血量（0x145FD–0x14639，docs/re/37 §3）。
//
//	血量 = ⌊基礎 / 4⌋ + 1d(基礎的低位) + 256 × 1d(基礎的高位)
//
// ⚠ **高低位是分開各擲一次再合起來的**，不是擲一次 16-bit。
// 基礎 < 256 時高位是 0、1d(0) ＝ 0，退化成 ⌊基礎/4⌋ + 1d(基礎)；
// 原版資料裡確實有基礎 > 255 的敵人（42 個區塊裡 8 筆），所以高位那一項不是死碼。
func (d EnemyData) RollHP(r *rng.State) uint16 {
	quarter := d.Base >> 2
	lo := int(quarter&0xFF) + r.Roll(int(d.Base&0xFF))
	hi := int(quarter>>8) + r.Roll(int(d.Base>>8)) + lo>>8
	return uint16(lo&0xFF) | uint16(hi&0xFF)<<8
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
func EnemyDamage(r *rng.State, d EnemyData) int {
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
	Data EnemyData
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
