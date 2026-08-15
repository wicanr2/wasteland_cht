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

	// 條件閘收尾改寫地圖格時用的兩個位移（docs/re/68 §1）。
	gatePatchPass = 0x04 // 全部人都過
	gatePatchFail = 0x06 // 有人沒過
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

// Eval 用目前這個角色逐條試（給只關心單人的呼叫端用）。
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

// GateResult 是一格條件閘跑完的結果。
type GateResult struct {
	Blocked bool       // 有人沒通過 → 擋住（原版回傳 ds:A5D1h ≠ 0）
	Failed  []GateHurt // 沒通過的人，各自已經受罰
	// PatchAt 是收尾要拿哪個位移去改寫這一格（原版 `sub_17CFF` 的 al）：
	// 全部人都過 ＝ 4（`0x1406D`）、有人沒過 ＝ 6（`0x14045`）。
	PatchAt int
}

// GateHurt 是一個人沒通過條件時受的罰。
type GateHurt struct {
	Member int  // 隊伍裡的序號
	Field  byte // 改到的欄位（0x1D ＝ CON、0x15 ＝ 金錢）
	Amount int  // 已經套用的量（負數是扣）
}

// EvalGate 照原版的形狀跑一格條件閘（`sub_13EC9`）。
//
// 迴圈是 `ds:A5D3h` 從 1 數到 `ds:4653h`：**每個能行動的人都跑一次整串條件**，
//
//	通過   → 換下一個人（不計數）
//	沒通過 → 用記錄 +0x08／+0x09 罰那個人，並記一筆
//
// 全部跑完之後「沒過的人數」非 0 就擋住——**所以是每個人都要過才放行**，
// 不是「有人過就好」。對開門這條路很嚴，對沙漠高溫（每個沒水壺的人都扣血）
// 才是正確的形狀（docs/re/67 §3）。
//
// ⚠ **有副作用**：技能檢定會擲骰、物品條件會消耗一次（鑰匙、水壺），
// 失敗還會改角色欄位。只能在真的要通過時呼叫一次。
func (p *Party) EvalGate(r *rng.State, record []byte, tbl SkillTable) GateResult {
	gates := ParseGates(record)
	var out GateResult
	for i, c := range p.Members {
		if !CanCommand(c) {
			continue // sub_172BB：不能行動的人跳過，不算沒過
		}
		passed := false
		for _, g := range gates {
			if p.evalOne(r, c, g, tbl) {
				passed = true
				break
			}
		}
		if passed {
			continue
		}
		field, amount := applyGatePenalty(r, c, record)
		out.Failed = append(out.Failed, GateHurt{Member: i, Field: field, Amount: amount})
	}
	out.Blocked = len(out.Failed) > 0
	// 收尾一定會改寫這一格，只是拿哪一對 byte 不同（docs/re/68 §1）。
	out.PatchAt = gatePatchPass
	if out.Blocked {
		out.PatchAt = gatePatchFail
	}
	return out
}

// applyGatePenalty 套用記錄 +0x08／+0x09 的獎懲（原版 sub_14193）。
//
//	+0x08 低 7 位 ＝ 角色記錄的欄位位移（0x1D ＝ CON、0x15 ＝ 金錢）
//	+0x08 的 bit7 ＝ 1 固定值／0 擲 (+0x09 & 0x7F) 顆 d6
//	+0x09 低 7 位 ＝ 量或骰數；**它的 bit7 ＝ 1 減、0 加**
//
// 欄位是 0 就什麼都不做（原版 `0x141A4` 的 `jz`）。
func applyGatePenalty(r *rng.State, c *Character, record []byte) (field byte, amount int) {
	if len(record) < 10 {
		return 0, 0
	}
	f, q := record[0x08], record[0x09]
	if f == 0 {
		return 0, 0
	}
	amount = int(q & 0x7F)
	if f&0x80 == 0 {
		amount = rollD6(r, int(q&0x7F)) // 擲骰
	}
	if q&0x80 != 0 {
		amount = -amount
	}
	switch f & 0x7F {
	case 0x1D: // CON，可以為負（docs/spec/09）
		c.PreHurt = c.CON
		c.CON += int16(amount)
	case 0x15: // 金錢，24-bit，不會變負
		switch {
		case amount >= 0:
			c.Money += uint32(amount)
		case uint32(-amount) > c.Money:
			c.Money = 0
		default:
			c.Money -= uint32(-amount)
		}
	}
	// 其餘欄位還沒對上語意（docs/re/67 §4），照原版**不動**比亂改安全。
	return f & 0x7F, amount
}

// SkillBytes 是技能資料表的原始 bytes（36 筆 × 2），實作 SkillTable。
type SkillBytes []byte

// Skill 取第 id 筆技能資料。
func (s SkillBytes) Skill(id byte) (SkillData, bool) {
	if int(id)*2+1 >= len(s) {
		return SkillData{}, false
	}
	return ParseSkillData(s[int(id)*2], s[int(id)*2+1]), true
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
		return c.SkillCheck(r, g.Param, data, g.Difficulty, p.PlayerStepped).OK

	case GateAttribute:
		return c.AttributeCheck(r, g.Param, g.Difficulty, p.PlayerStepped).OK

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
