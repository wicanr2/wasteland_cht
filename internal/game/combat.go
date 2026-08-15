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
	Base    uint16    // +0x00/+0x01
	Speed   byte      // +0x02，行動值欄位（× 8 進行動值，docs/spec/12）
	DiceN   byte      // +0x03，傷害骰數
	XPMul   byte      // +0x04 的低 4 位（實際倍數要 +1）
	Weapon  ItemClass // +0x05 的低 4 位，武器類別（與物品表同一套編碼）
	DamBase byte      // +0x05 的高 4 位，傷害基底
	Kind    EnemyKind // +0x06，敵人種類
	// Portrait 是這種敵人的肖像圖編號（ALLPICS，docs/re/37 §3.2）。
	// 遭遇時用它載圖，**同一個編號也決定文字裡用 him／her／it**
	// （查 ds:A920h → 文字碼 0x0E 的選擇子）。
	Portrait byte // +0x07
	Raw      [8]byte

	// 敵方護甲的骰數來自別的路徑（loc_12A92），不在這 8 bytes 裡，
	// 所以由呼叫者傳給 Enemy.TakeDamage，不放進這個結構假裝解過了。
	//
	// +0x04 的**高 4 位沒有讀者**：資料裡有值（實測 0–10），但全檔只有
	// loc_12A92（& 0x0F）與 0x19A14（物品表那條路）碰 +0x04。
	// 原樣留在 Raw 裡，不給語意（docs/re/37 §3.2）。
}

// EnemyKind 是敵人種類（資料 +0x06）。42 個區塊 397 筆全部落在 1–5，
// 而且執行檔字串表的第 0x53–0x57 條正好是這五個名字（docs/re/37 §3.2）。
type EnemyKind byte

const (
	KindAnimal   EnemyKind = 1
	KindMutant   EnemyKind = 2
	KindHumanoid EnemyKind = 3
	KindCyborg   EnemyKind = 4
	KindRobot    EnemyKind = 5
)

// MessageID 回這個種類對應的訊息編號（原版 0x129FF 的 add al, 52h）。
func (k EnemyKind) MessageID() byte { return 0x52 + byte(k) }

// ParseEnemyData 拆一筆 8 bytes 的敵人資料。
func ParseEnemyData(b []byte) EnemyData {
	var d EnemyData
	copy(d.Raw[:], b)
	d.Base = uint16(b[0]) | uint16(b[1])<<8
	d.Speed = b[2]
	d.DiceN = b[3]
	d.XPMul = b[4] & 0x0F
	d.Weapon = ItemClass(b[5] & 0x0F)
	d.DamBase = b[5] >> 4
	d.Kind = EnemyKind(b[6])
	d.Portrait = b[7]
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

// SkillBrawling 是命中累加值固定用的技能編號（`sub_1B0F1` 的 `mov al, 1`）。
//
// ⚠ **不是「拿什麼武器就用什麼技能」**：`sub_1B108` 把 1 寫死在指令裡，
// 所以拿步槍的人命中也是加 Brawling 等級（docs/re/88 §2）。
const SkillBrawling byte = 1

// HitChance 累加出命中判定要比的那個值（`sub_1B108`，docs/re/88）。
//
//	base ＋ Brawling×3 ＋ Agility − 對手行動值（近戰類別 ×4，否則另加 5），夾在 100
//
// base 由目標的一個欄位決定（40／50／60），那個欄位的語意未解，
// 所以當參數收進來（docs/spec/06 §6）。
//
// c 永遠是**隊伍成員**：兩條攻擊路徑共用這一支，累加的都是隊伍那一邊的本事，
// 方向靠比較符號翻轉（docs/re/20 §2）。foe 是那一組敵人——隊伍攻擊時它是目標，
// 敵方攻擊時它是攻擊者，但原版讀的都是 `ds:CF80h` 指的同一組。
func HitChance(c *Character, base int, foe EnemyData) uint16 {
	acc := SatAdd(0, base)
	acc = SatAdd(acc, int(c.SkillLevel(SkillBrawling))*3)
	acc = SatAdd(acc, int(c.Attributes[AttrAgility]))

	// ⚠ `shl al, 1` 兩次是 **8-bit** 位移（0x1B139），高位直接丟掉：
	// 行動值 64 以上乘 4 會繞回小數字。照抄，不要改成 int 乘法。
	evade := int(foe.Speed)
	if foe.Weapon == ClassMelee {
		evade = int(foe.Speed << 2)
	} else {
		acc = SatAdd(acc, 5)
	}
	acc = SatAdd(acc, -evade)

	if acc > 100 {
		acc = 100 // 原版夾在 100（sub_19C72 → sub_19BF8 歸零再加 100）
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

// WeaponDice 是 sub_15755 算出來的骰數（docs/re/45 §4）。
//
// 顆數就是武器的 Dice；反戰車武器（類別 8／9）在 `Dice ≥ x` 時變成 `2·Dice − x`，
// 溢位飽和在 255。x 由呼叫端給：主攻擊路徑（0x1AFA3）是 0，
// sub_15738 那條路是敵人資料 +0x04 的低 4 位。
func WeaponDice(w ItemData, x byte) byte {
	n := w.Dice
	if w.Class != ClassATLight && w.Class != ClassATHeavy {
		return n
	}
	if n < x {
		return n
	}
	if v := int(n) + int(n-x); v <= 0xFF {
		return byte(v)
	}
	return 0xFF
}

// PartyWeaponDamage 是隊伍傷害的第一項：**0 ＋ N 顆 d6**（sub_15755）。
//
// ⚠ 這一項會擲骰，另外四項不會。不要為了「好算」換成期望值——
// 26 顆 d6 的分布和 91 點固定傷害是兩回事。
func PartyWeaponDamage(r *rng.State, w ItemData, x byte) int {
	return r.SumD6(0, int(WeaponDice(w, x)))
}

// PartyDamage 是隊伍打敵方的傷害：**五項相加**（docs/re/20 §4.1）。
//
// 第一項是武器的傷害骰（會擲骰），另外四項是不擲骰的加值。
// 技能編號來自武器資料的 Skill 欄位，不由呼叫端自己猜（docs/re/45 §3.3）。
func PartyDamage(r *rng.State, c *Character, w ItemData, x byte) uint16 {
	acc := SatAdd(0, PartyWeaponDamage(r, w, x))
	acc = SatAdd(acc, int(c.SkillLevel(w.Skill))*3)
	acc = SatAdd(acc, AttrModifier(c.Attributes[AttrDexterity]))
	acc = SatAdd(acc, AttrModifier(c.Attributes[AttrStrength]))
	acc = SatAdd(acc, AttrModifier(c.Attributes[AttrLuck]))
	return acc
}

// Equip 把一件物品裝上去，照 sub_1949E：類別 15（護甲）寫護甲槽並把
// 骰數搬進 AC，其餘寫武器槽。再選一次同一個槽 ＝ 卸下。
func (c *Character) Equip(slot int, w ItemData) {
	s := byte(slot)
	if c.EquipIndex == s || c.ArmorIndex == s {
		c.unequip(s)
		return
	}
	if w.Class == ClassArmor {
		c.AC = w.Dice
		c.ArmorIndex = s
		return
	}
	c.EquipIndex = s
}

func (c *Character) unequip(s byte) {
	if c.ArmorIndex == s {
		c.ArmorIndex = 0
		c.AC = 0
	}
	if c.EquipIndex == s {
		c.EquipIndex = 0
	}
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
