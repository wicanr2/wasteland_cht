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

// 記錄 +0x00 低位的四個旗標（`sub_13EC9` 的四條 `sub_142E2 & n` 分支，
// docs/re/69）。四條都有資料會走到，不是死路。
const (
	// GateAnyPass（`& 4`，`0x1401D`，76 筆）：收尾時只要**有人通過**
	// （`ds:A5D0h` ≠ 0）就算過，其他人失敗不影響。
	GateAnyPass = 0x04
	// GateCondPatch（`& 8`，`0x14052`，61 筆）：有人通過就**立刻收尾**，
	// 而且改寫這一格用的位移由 `sub_142B1` 依**通過的是哪一條條件**算出來，
	// 不是固定的 4——條件串列的 0xFF 之後接著一張逐條件的改寫表。
	GateCondPatch = 0x08
	// GateWholeParty（`& 0x10`，`0x13FB5`，15 筆）：有人失敗就**立刻收尾**，
	// 懲罰改成 `sub_14296`（**全隊每個人都套一次**），
	// 訊息走 `sub_142ED`（在暫時換掉的時鐘值底下顯示記錄 +0x03）。
	GateWholeParty = 0x10
	// GateEachMember（`& 0x20`，`0x13FE6`／`0x14063`，139 筆）：**逐個角色跑**；
	// 沒設就只處理第一個能行動的人，他的結果決定一切。
	GateEachMember = 0x20
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
	// 全部人都過 ＝ 4（`0x1406D`）、有人沒過 ＝ 6（`0x14045`），
	// GateCondPatch 那條路則由 `condPatchOffset` 算（`0x1405D`）。
	PatchAt int
	// Message 是收尾要印的字串編號（0 ＝ 不印）：
	// 通過印記錄 +0x02（`sub_14175`）、沒過且沒人受罰印 +0x03（`sub_1417A`）。
	// 有人受罰時原版不印這句——受罰本身已經有 " gets hurt for "。
	Message int
}

// GateHurt 是一個人沒通過條件時受的罰。
type GateHurt struct {
	Member int  // 隊伍裡的序號
	Field  byte // 改到的欄位（0x1D ＝ CON、0x15 ＝ 金錢）
	Amount int  // 已經套用的量（負數是扣）
}

// EvalGate 照原版的形狀跑一格條件閘（`sub_13EC9`，docs/re/69）。
//
// 迴圈是 `ds:A5D3h` 從 1 數到 `ds:4653h`，每個能行動的人跑一次整串條件；
// 走完之後的收尾由記錄 +0x00 低位的四個旗標決定（見上面那組常數）。
// 預設（只有 GateEachMember）是**每個人都要過才放行**——對開門很嚴，
// 對沙漠高溫（每個沒水壺的人都扣血）才是正確的形狀。
//
// ⚠ **有副作用**：技能檢定會擲骰、物品條件會消耗一次（鑰匙、水壺），
// 失敗還會改角色欄位。只能在真的要通過時呼叫一次。
func (p *Party) EvalGate(r *rng.State, record []byte, tbl SkillTable) GateResult {
	gates := ParseGates(record)
	var flags byte
	if len(record) > 0 {
		flags = record[0]
	}
	var out GateResult
	passed, failed := 0, 0
	out.PatchAt = -1

loop:
	for i, c := range p.Members {
		if !CanCommand(c) {
			continue // sub_172BB：不能行動的人跳過，不算沒過
		}
		hit := p.evalGateFor(r, c, gates, tbl)

		if hit >= 0 {
			passed++
			if flags&GateCondPatch != 0 {
				// 0x14056：印 +0x02，改寫位移由通過的那一條條件算出來。
				out.PatchAt = condPatchOffset(record, hit)
				out.Message = recordString(record, 2)
				break loop
			}
			if flags&GateEachMember != 0 {
				continue // 0x14063 → 0x13FED：換下一個人
			}
			out.PatchAt = gatePatchPass // 0x1406A
			out.Message = recordString(record, 2)
			break loop
		}

		failed++
		if flags&GateWholeParty != 0 {
			// 0x14010：**全隊**每個人各套一次懲罰（sub_14296），然後直接收尾。
			for j, m := range p.Members {
				if field, amount := applyGatePenalty(r, m, record); field != 0 {
					out.Failed = append(out.Failed,
						GateHurt{Member: j, Field: field, Amount: amount})
				}
			}
			break loop
		}
		// 0x13FBC：+0x08 是 0 就完全不罰——**連 ds:A5D2h（受罰人數）都不加**，
		// 所以 Failed 也不記，收尾才印得出 +0x03（0x1402D 只看 A5D2h）。
		if field, amount := applyGatePenalty(r, c, record); field != 0 {
			out.Failed = append(out.Failed,
				GateHurt{Member: i, Field: field, Amount: amount})
		}
		if flags&GateEachMember == 0 {
			break loop // 0x13FEB → 0x14028
		}
	}

	if out.PatchAt >= 0 {
		return out // 已經走到 0x1406A／0x14056 那兩條放行路
	}
	// 收尾（0x1401A）：只要旗標允許「有人過就算過」就放行，否則擋住。
	if flags&GateAnyPass != 0 && passed > 0 {
		out.PatchAt = gatePatchPass
		out.Message = recordString(record, 2)
		return out
	}
	if failed == 0 {
		out.PatchAt = gatePatchPass // 迴圈跑完沒人失敗（0x14000 的 jz 1406A）
		out.Message = recordString(record, 2)
		return out
	}
	out.Blocked = true // 0x14041：ds:A5D1h ← 1
	out.PatchAt = gatePatchFail
	if len(out.Failed) == 0 {
		// 0x1402D：一個人都沒受罰才印 +0x03。
		out.Message = recordString(record, 3)
	}
	return out
}

// recordString 取記錄裡某個位移的字串編號（原版 sub_16D1A(bl)，0 ＝ 不印）。
func recordString(record []byte, at int) int {
	if at >= len(record) {
		return 0
	}
	return int(record[at])
}

// evalGateFor 逐條試，回傳通過的是第幾條（−1 ＝ 這個人沒過）。
//
// 型別 4（比金錢）失敗時原版 `0x13F50` 是 `jb loc_13FB1`——
// **直接判這個人沒過**，不再試後面的條件；其餘型別失敗才往下一條。
func (p *Party) evalGateFor(r *rng.State, c *Character, gates []Gate, tbl SkillTable) int {
	for i, g := range gates {
		if p.evalOne(r, c, g, tbl) {
			return i
		}
		if g.Type == GateMoney {
			return -1
		}
	}
	return -1
}

// condPatchOffset 算 GateCondPatch 那條路要拿記錄的哪個位移（原版 sub_142B1）。
//
// 條件串列以 0xFF 結束，**0xFF 之後接著一張逐條件的改寫表**：
// 通過第 n 條就用「0xFF 的位置 + 1 + 2n」那一對 byte 去改寫地圖格。
func condPatchOffset(record []byte, hit int) int { return CondPatchOffsetFor(record, hit) }

// CondPatchOffsetFor 是 condPatchOffset 的對外版，給驗證工具檢查資料用。
func CondPatchOffsetFor(record []byte, hit int) int {
	end := gateListAt
	for end+1 < len(record) && record[end] != gateEnd {
		end += 2
	}
	return end + 1 + hit*2
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
		// ⚠ **扣 CON 走的是傷害結算，不是直接減**（`0x1422F` → `sub_157D6`）。
		// 也就是說**護甲會吸收**，除非這一格記錄 `+0x00` 的 bit0 設著。
		// 全檔 168 筆扣 CON 的條件閘裡 105 筆是照扣護甲的（`docs/re/122` §3）——
		// 少了這一段，穿好裝備在那 105 格完全沒有用，而畫面上只是「扣得比較多」。
		//
		// 加的那一路沒有這一段（`loc_1426A` 直接加），所以只有負值走這裡。
		if amount < 0 {
			hurt := -amount
			if !BypassesArmour(record) {
				hurt -= absorbWith(r, c.AC)
			}
			if hurt <= 0 {
				return f & 0x7F, 0 // 全部被吸收：不扣血也不留 PreHurt
			}
			c.PreHurt = c.CON
			c.CON -= int16(hurt)
			return f & 0x7F, -hurt
		}
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

// absorbWith 是護甲吸收，沒有亂數源時回 0（無頭測試會這樣）。
func absorbWith(r *rng.State, ac byte) int {
	if r == nil {
		return 0
	}
	return Absorb(r, ac)
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
// 次數只剩 1 的時候再用就把那一槽清成 0。
//
// ⚠ **清空不是移除**：物品陣列是固定 30 槽（`docs/re/15`），賣掉一件
// 也是「把那兩個 byte 清成 0」而不搬動後面的（`docs/re/42` §3）。
// 把後面的往前搬會讓角色記錄 `+0x1F`／`+0x25` 存的槽號指到別件東西——
// 症狀是用掉一件消耗品之後，裝備欄安靜地換成另一把武器。
func (c *Character) UseItem(id byte) bool {
	if id == 0 {
		return false // 0 是空槽，不是一件物品
	}
	for i := range c.Items {
		if c.Items[i].ID != id {
			continue
		}
		uses := c.Items[i].Value & 0x3F
		switch {
		case uses == 0:
			return true // 次數為 0 的不遞減（原版直接離開）
		case uses == 1:
			c.Items[i] = Slot{}
		default:
			c.Items[i].Value = (c.Items[i].Value & 0xC0) | (uses - 1)
		}
		return true
	}
	return false
}


// UseGate 是 `USE` 指令的規則層：拿指定的技能／物品／屬性去試這一格的條件串列。
//
// 與走上去自動評估（`Eval`）的差別是**只試吻合的那一條**——
// 原版 `sub_140DD`／`sub_14090`／`sub_14126` 三支各自逐筆掃條件串列，
// 找型別與編號都對上的那一筆，找到就跑它的判定，找不到就失敗
//（`docs/re/92` §4）。自動評估是逐條試到成功為止，兩者不一樣。
//
// ⚠ **物品只認型別 1**（`sub_14090` 的 `cmp al, 1`），
// 而自動評估把 1／5／6／7 都當成找物品（`docs/re/32` §6）。
// 照原版保留這個差別，不要統一。
//
// 回傳命中的條列索引（−1 ＝ 沒有吻合的條件）與判定結果。
func (p *Party) UseGate(r *rng.State, record []byte, c *Character,
	kind UseKind, id byte, tbl SkillTable) (hit int, passed bool) {
	if c == nil || c.Down() {
		return -1, false
	}
	for i, g := range ParseGates(record) {
		if !matchUse(g, kind, id) {
			continue
		}
		return i, p.evalOne(r, c, g, tbl)
	}
	return -1, false
}

// UseResult 是 `USE` 打在一格條件閘上的完整結果。
//
// `UseGate` 只回答「哪一條吻合、過了沒」，但原版在那之後還有收尾
// （`0x13D18` 與 `loc_13D3F`）：**兩條路都會改寫腳下那一格**。
// 少了這一步，科奇斯基地的四根圓柱插了鑰匙也不會有下一根出現——
// 整個結局的啟動序列走不動（`docs/re/100` §3）。
type UseResult struct {
	Hit     int  // 命中的條件序號，−1 ＝ 沒有吻合的條件
	Passed  bool // 命中那一條過了沒
	PatchAt int  // 收尾要拿記錄的哪個位移去改寫這一格
	Message int  // 要印的字串編號（0 ＝ 不印）
	// Failed 是失敗那條路套出去的懲罰（記錄 `+0x08`／`+0x09`）。
	Failed []GateHurt
}

// UseOn 是 `USE` 的完整規則層：判定 ＋ 收尾。
//
// 三條路照原版（`docs/re/92` §4）：
//
//	命中且通過   → 位移 4（`GateCondPatch` 設起來時改由通過的那一條算），印 +0x02
//	命中但沒通過 → 位移 6，印 +0x03，套懲罰
//	沒有吻合的   → `loc_13D3F` 同一條：位移 6，套懲罰
//
// ⚠ **「沒有吻合的條件」不是什麼都不做**——原版照樣改寫這一格。
// 多數記錄的 `+0x06`／`+0x07` 是 `0xFF 0xFF`（不改），所以看起來像沒事。
func (p *Party) UseOn(r *rng.State, record []byte, c *Character,
	kind UseKind, id byte, tbl SkillTable) UseResult {
	hit, passed := p.UseGate(r, record, c, kind, id, tbl)
	out := UseResult{Hit: hit, Passed: passed, PatchAt: gatePatchFail}
	if hit >= 0 && passed {
		out.PatchAt = gatePatchPass
		if len(record) > 0 && record[0]&GateCondPatch != 0 {
			out.PatchAt = CondPatchOffsetFor(record, hit)
		}
		out.Message = recordString(record, 2)
		return out
	}
	out.Message = recordString(record, 3)
	// 懲罰的形狀與自動評估同一支：`GateWholeParty` 設起來就全隊各套一次。
	if len(record) > 0 && record[0]&GateWholeParty != 0 {
		for j, m := range p.Members {
			field, amount := applyGatePenalty(r, m, record)
			out.Failed = append(out.Failed, GateHurt{Member: j, Field: field, Amount: amount})
		}
		return out
	}
	for j, m := range p.Members {
		if m != c {
			continue
		}
		field, amount := applyGatePenalty(r, m, record)
		out.Failed = append(out.Failed, GateHurt{Member: j, Field: field, Amount: amount})
	}
	return out
}

// matchUse 是「這一條條件吃不吃這個東西」。
func matchUse(g Gate, kind UseKind, id byte) bool {
	switch kind {
	case UseSkill:
		return g.Type == GateSkill && g.Param == id
	case UseItem:
		return g.Type == 1 && g.Param == id
	case UseAttribute:
		// 屬性那條的參數是**角色記錄位移**（0x0E–0x14），不是屬性索引
		// （`docs/re/32` §4 的 `sub_1820C(難度, 位移)`）。
		return g.Type == GateAttribute && g.Param == id
	}
	return false
}
