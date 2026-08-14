package game

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// 檢定：技能、屬性、以及成功之後的練等（docs/spec/07 §3、docs/re/32）。
//
// 全遊戲共用同一個骨架，技能與屬性只是不同的填法：
//
//	門檻 = 難度 × 5 + 15
//	本事 = 2d6 續擲（< 5 直接失敗）＋ 屬性 ＋ 技能等級 × 3
//	成功 ⟺ 本事 ≥ 門檻

// fumble 以下的骰點直接失敗，不看任何加成（sub_18146 的 cmp al, 5）。
const fumble = 5

// Threshold 是難度換算出來的門檻（sub_19CAC）。
func Threshold(difficulty int) int { return difficulty*5 + 15 }

// CheckResult 是一次檢定的結果。
type CheckResult struct {
	OK      bool
	Roll    int // 2d6 續擲的點數；0 表示沒擲（性別／等級那種直接比的）
	Total   int // 累加出來的本事
	Trained bool
}

// SkillCheck 做一次技能檢定。data 是技能資料表那一筆（ds:BA20h）。
//
// awardXP 對應原版的 ds:916Bh（docs/re/32 §7.1）：**只有玩家自己走出來的那一步
// 之後的檢定才給經驗值**，自動步（時間流逝觸發的那種）不給。
// 它記的是「最後一次移動」而不是每次檢定各自設，所以由呼叫端把
// Party.PlayerStepped 傳進來，不要改成 per-check。
func (c *Character) SkillCheck(r *rng.State, id byte, data SkillData, difficulty int, awardXP bool) CheckResult {
	var res CheckResult
	roll := r.PairD6()
	res.Roll = roll
	if roll < fumble {
		return res
	}
	if awardXP {
		c.AddXP(uint32(roll))
	}

	total := roll
	if off := int(data.Attribute); off >= recAttributes && off < recAttributes+AttrCount {
		total += int(c.Attributes[off-recAttributes])
	}
	total += int(c.SkillLevel(id)) * 3
	res.Total = total
	res.OK = total >= Threshold(difficulty)

	if res.OK {
		res.Trained = c.TrainSkill(r, id, difficulty)
	}
	return res
}

// AttributeCheck 做一次屬性檢定。offset 是角色記錄的位移。
//
// 兩個特例不擲骰：0x18（性別）比相等、0x24（等級）比大小。
// awardXP 的語意同 SkillCheck。
func (c *Character) AttributeCheck(r *rng.State, offset byte, difficulty int, awardXP bool) CheckResult {
	switch offset {
	case recGender:
		return CheckResult{OK: int(c.Gender) == difficulty}
	case recLevel:
		return CheckResult{OK: int(c.Level) >= difficulty}
	}

	var res CheckResult
	roll := r.PairD6()
	res.Roll = roll
	if roll < fumble {
		return res
	}
	if awardXP {
		c.AddXP(uint32(roll))
	}

	total := roll
	if int(offset) >= recAttributes && int(offset) < recAttributes+AttrCount {
		total += int(c.Attributes[int(offset)-recAttributes])
	}
	res.Total = total
	res.OK = total >= Threshold(difficulty)
	return res
}

// TrainSkill 是檢定成功之後的自動練等（sub_1818E）。
//
//	沒學過                  → 不升
//	技能等級 ≥ 角色等級      → 不升（技能等級的上限就是角色等級）
//	技能等級 ≥ 難度          → 不升
//	門檻 = (難度 − 等級) ÷ 2 + 1；roll(1..10) < 門檻 → +1
func (c *Character) TrainSkill(r *rng.State, id byte, difficulty int) bool {
	lvl := c.SkillLevel(id)
	if lvl == 0 || int(lvl) >= int(c.Level) || int(lvl) >= difficulty {
		return false
	}
	gate := (difficulty-int(lvl))/2 + 1
	if r.Roll(10) >= gate {
		return false
	}
	for i := range c.Skills {
		if c.Skills[i].ID == id {
			if c.Skills[i].Value < 0xFF {
				c.Skills[i].Value++
			}
			return true
		}
	}
	return false
}
