package game

// 經驗值、升級與技能學習的公式（docs/spec/05 §4.1、docs/re/31、32）。

// 24-bit 欄位的上限；經驗值加到這裡就飽和，不繞回。
const maxUint24 = 0xFFFFFF

// XPForLevel 是升到等級 L 需要的**累計**經驗值：(L² − L) × 512。
//
// 原版沒有 mul：L² 是「重複 L 次加 L」、× 512 是九次移位。
// 這裡直接寫成算式，結果相同（0x1B93D 起）。
func XPForLevel(level int) uint32 {
	if level <= 1 {
		return 0
	}
	l := uint32(level)
	return (l*l - l) * 512
}

// AddXP 累加經驗值，溢位飽和在 0xFFFFFF（sub_19BC0 的進位鏈）。
func (c *Character) AddXP(n uint32) {
	sum := uint64(c.XP) + uint64(n)
	if sum > maxUint24 {
		sum = maxUint24
	}
	c.XP = uint32(sum)
}

// LevelUp 依經驗值把等級一路升上去，回傳升了幾級。
//
// 原版每升一級：等級寫回 +0x24、技能點 +1（飽和 255）、播音效、查階級表，
// 然後回頭再檢查一次——所以經驗值夠的話會連升（0x1BA08）。
// **升級不扣經驗值。**
func (c *Character) LevelUp() int {
	gained := 0
	for c.XP >= XPForLevel(int(c.Level)+1) {
		if c.Level == 0xFF {
			break
		}
		c.Level++
		if c.SkillPts < 0xFF {
			c.SkillPts++
		}
		gained++
	}
	return gained
}

// SkillCost 是把某個技能升到等級 L 要花的技能點：基礎 × 2^(L−1)，飽和 0xFF。
func SkillCost(base byte, level int) byte {
	if level <= 0 {
		return 0
	}
	cost := uint32(base)
	for i := 1; i < level; i++ {
		cost *= 2
		if cost > 0xFF {
			return 0xFF
		}
	}
	if cost > 0xFF {
		return 0xFF
	}
	return byte(cost)
}

// SkillData 是技能資料表的一筆（ds:BA20h，2 bytes，docs/re/32 §2）。
type SkillData struct {
	IQ        byte // +0x00 >> 3
	BaseCost  byte // +0x00 & 7
	Attribute byte // +0x01，檢定用的屬性——值就是角色記錄的位移
}

// ParseSkillData 拆一筆技能資料。
func ParseSkillData(b0, b1 byte) SkillData {
	return SkillData{IQ: b0 >> 3, BaseCost: b0 & 7, Attribute: b1}
}

// SkillLevel 回傳角色某個技能的等級，沒學過回 0（sub_198CD）。
func (c *Character) SkillLevel(id byte) byte {
	for _, s := range c.Skills {
		if s.ID == id {
			return s.Value
		}
	}
	return 0
}

// LearnSkill 學一個技能或把它升一級，回傳是否成功與失敗原因。
//
// 兩道檢查照原版的順序：先看 IQ 夠不夠（不夠印 YOU ARE NOT SMART ENOUGH!），
// 再看技能點夠不夠（不夠印 NOT ENOUGH SKILL POINTS!）。
func (c *Character) LearnSkill(id byte, data SkillData) (ok bool, reason string) {
	if c.Attributes[AttrIQ] < data.IQ {
		return false, "IQ 不足"
	}
	cur := c.SkillLevel(id)
	cost := SkillCost(data.BaseCost, int(cur)+1)
	if cost == 0 || c.SkillPts < cost {
		return false, "技能點不足"
	}
	c.SkillPts -= cost
	for i := range c.Skills {
		if c.Skills[i].ID == id {
			if c.Skills[i].Value < 0xFF {
				c.Skills[i].Value++
			}
			return true, ""
		}
	}
	c.Skills = append(c.Skills, Slot{ID: id, Value: 1})
	return true, ""
}
