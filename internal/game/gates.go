package game

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// 條件串列（nibble 2，docs/spec/07 §4、docs/re/32 §6）。
//
// 地圖記錄 +0x0A 起每 2 bytes 一條，逐條試到成功或碰到 0xFF：
//
//	byte 0：高 3 位 ＝ 型別、低 5 位 ＝ 難度
//	byte 1：型別相關的參數

const (
	gateListAt = 0x0A
	gateEnd    = 0xFF
)

// 條件型別。1／5／6／7 的判定路徑相同（都是找物品），
// 為什麼要分四個編號還沒解——照原版保留編號，不合併。
const (
	GateSkill     = 0 // 參數 ＝ 技能編號
	GateAttribute = 2 // 參數 ＝ 角色記錄位移
	GatePartySize = 3 // 參數 ＝ 要幾個人
	GateMoney     = 4 // 參數 ＝ 金額（只比不扣）
)

// Gate 是條件串列裡的一條。
type Gate struct {
	Type       int
	Difficulty int
	Param      byte
}

// ParseGates 解析一筆地圖記錄的條件串列。
func ParseGates(record []byte) []Gate {
	var out []Gate
	for at := gateListAt; at+1 < len(record); at += 2 {
		b0 := record[at]
		if b0 == gateEnd {
			break
		}
		out = append(out, Gate{
			Type:       int(b0 >> 5),
			Difficulty: int(b0 & 0x1F),
			Param:      record[at+1],
		})
	}
	return out
}

// SkillTable 讓規則層查技能資料（ds:BA20h），由呼叫者提供，
// 規則層不自己去讀執行檔。
type SkillTable interface {
	Skill(id byte) (SkillData, bool)
}

// Eval 逐條試這串條件，任何一條成立就通過（原版的迴圈就是這個形狀）。
// 回傳通過的是第幾條（−1 表示全部失敗）。
func (p *Party) Eval(r *rng.State, gates []Gate, tbl SkillTable) int {
	c := p.Current()
	if c == nil {
		return -1
	}
	for i, g := range gates {
		if p.evalOne(r, c, g, tbl) {
			return i
		}
	}
	return -1
}

func (p *Party) evalOne(r *rng.State, c *Character, g Gate, tbl SkillTable) bool {
	switch g.Type {
	case GateSkill:
		if c.SkillLevel(g.Param) == 0 {
			return false // 沒學過就不試
		}
		data, ok := SkillData{}, false
		if tbl != nil {
			data, ok = tbl.Skill(g.Param)
		}
		if !ok {
			return false
		}
		return c.SkillCheck(r, g.Param, data, g.Difficulty).OK

	case GateAttribute:
		return c.AttributeCheck(r, g.Param, g.Difficulty).OK

	case GatePartySize:
		return int(g.Param) == len(p.Members)

	case GateMoney:
		// 只比大小，不扣（sub_17B3E）。
		return c.Money >= uint32(g.Param)

	default:
		// 1／5／6／7：身上有沒有這件物品；有的話消耗一次。
		return c.UseItem(g.Param)
	}
}

// Current 是目前操作的角色。
func (p *Party) Current() *Character {
	if p.Selected < 0 || p.Selected >= len(p.Members) {
		if len(p.Members) == 0 {
			return nil
		}
		return p.Members[0]
	}
	return p.Members[p.Selected]
}

// UseItem 找一件物品並消耗一次（sub_1968C ＋ sub_19A58）。
//
// 附屬 byte 的**低 6 位是剩餘次數**，高 2 位原樣保留；
// 次數只剩 1 的時候再用就把整個槽移除。
func (c *Character) UseItem(id byte) bool {
	for i := range c.Items {
		if c.Items[i].ID != id {
			continue
		}
		uses := c.Items[i].Value & 0x3F
		switch {
		case uses == 0:
			return true // 次數為 0 的不遞減（原版直接離開）
		case uses == 1:
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
		default:
			c.Items[i].Value = (c.Items[i].Value & 0xC0) | (uses - 1)
		}
		return true
	}
	return false
}
