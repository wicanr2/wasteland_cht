package game

// 雇用（`Hire`）的結算（`docs/re/110`）。
//
// 一句話：**NPC 本來就是一筆完整的 256 byte 角色記錄**，放在地圖區塊的
// section 17；雇用成功就是把它整筆抄進隊伍。這裡做的是「抄之前要過的那幾關」
// 與「抄完要改的那兩格」。
//
// ⚠ 不要在這一層決定「畫面上要說什麼」——訊息編號在原版是固定的，
// 由呈現層去查字串表（成功印 NPC 記錄 `+0x30` 那一條、失敗印字串 44 ＋ 81）。

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

const (
	// recHireField 是**遭遇記錄**（`ds:46C6h` 指到的那一筆）裡放雇用資訊的 byte。
	//
	// ⚠ 它的擁有者是「這一格的記錄」，而那一格的 section 型別由 nibble 決定
	// （`docs/re/16` §3.1）——所以不要寫成「section 15 的 +0x09」。
	// remake 這一側拿到的就是 `Block.CellRecord` 回的那一筆，與原版同一條路。
	recHireField = 0x09
	// friendlyBit 是遭遇記錄 `+0x09` 的 **bit1 ＝ 這一組不敵對**
	// （`sub_12AC5` 的第二個 `shr` 把它送進 CF，`docs/re/114`）。
	//
	// 六個消費端全部同一個方向：友善的那一組**不排移動計畫**（`0x14C50`）、
	// **不攻擊隊伍**（`0x1ADBC`）、不算進「附近有敵人」（`0x149A2`），
	// 而且**可以雇用**（`0x132D8`）。
	//
	// ⚠ 雇用還要編號非 0，所以 `hireOfferBit` 與它同一個位元但**不是同義詞**：
	// 友善 ≠ 可雇用（出貨資料裡有友善但沒有 NPC 記錄的遭遇）。
	friendlyBit  = 0x02
	hireOfferBit = friendlyBit

	// recNPCGreeting／recNPCPrice 是 NPC 記錄自己的兩格（`docs/re/110` §4）。
	recNPCGreeting = 0x30
	recNPCPrice    = 0x31

	// HireCap 是隊伍人數上限（`0x132CE` 的 `cmp al, 7`）。
	// 滿了印字串 95「名冊沒有空位了」，**在擲骰之前就擋掉**。
	HireCap = 7

	// hireSectionNPC 是 NPC 記錄住的 section 型別（`0x132E1` 的 `mov bl, 11h`）。
	hireSectionNPC = 17

	// hireRollFloor 是骰子的門檻（`0x1333B` 的 `cmp al, 5`）：**擲出 5 以下直接失敗**。
	hireRollFloor = 5
)

// HireOffer 是遭遇記錄 `+0x09` 拆出來的兩件事。
type HireOffer struct {
	NPC   int  // section 17 裡的記錄編號（高 4 位）
	Valid bool // bit1 有沒有設，以及編號是不是非零
}

// ReadHireOffer 拆遭遇記錄的 `+0x09`。
//
// ⚠ **兩個條件都要成立**：`bit1` 設（`0x132D8`）且編號非 0（`0x132DD`）。
// 只看其中一個會把「有敵人但不能雇用」的格子當成可以雇用。
func ReadHireOffer(rec []byte) HireOffer {
	if len(rec) <= recHireField {
		return HireOffer{}
	}
	v := rec[recHireField]
	n := int(v >> 4)
	return HireOffer{NPC: n, Valid: v&hireOfferBit != 0 && n != 0}
}

// HireSection 是 NPC 記錄住的 section 型別，給呈現層取記錄用。
func HireSection() int { return hireSectionNPC }

// Friendly 回報這一筆遭遇是不是不敵對的（`docs/re/114` §2）。
//
// **與「可以雇用」是兩件事**：友善的那一組不一定有 NPC 記錄。
func Friendly(rec []byte) bool {
	return len(rec) > recHireField && rec[recHireField]&friendlyBit != 0
}

// TurnHostile 把友善位元清掉，回報這一下有沒有真的改變什麼。
//
// 原版在**隊伍攻擊的結算裡**做這件事（`0x1AF7B` → `sub_15C19`），
// 而且是在命中判定**之前**——所以**開槍就算數，打不打得中都一樣**。
// 同一支還會把 bit0 設起來（名字改用地圖的明文名字表，`docs/re/114` §3）。
//
// ⚠ `rec` 是地圖區塊 `Raw` 的切片，所以這一下**寫進區塊**，
// 這一場戰鬥結束之後仍然成立——那正是原版的行為：翻臉是永久的。
func TurnHostile(rec []byte) bool {
	if len(rec) <= recHireField || rec[recHireField]&friendlyBit == 0 {
		return false
	}
	rec[recHireField] = rec[recHireField]&^byte(friendlyBit) | 1
	return true
}

// HireRoll 是 `sub_19C84` 那顆**長度可變**的骰。
//
// 擲一顆 d6 當基準，再擲一顆累加；**兩顆同點就換一個新基準繼續擲**，
// 不同點才停。所以它多數時候是 2d6（且兩顆不同點），偶爾更多顆。
//
// ⚠ 原版是 byte 運算會繞回；這裡用 int 累加而不夾——擲到繞回要六十幾顆同點，
// 而**夾在 255 與繞回是不同的行為**，沒有證據就不要挑一個。
func HireRoll(r *rng.State) int {
	acc, base := 0, r.D6()
	for {
		x := r.D6()
		acc += x + base
		if x != base {
			return acc
		}
		base = r.D6()
	}
}

// hireBase 是 `sub_13427`：**(魅力 ＋ 智力) ÷ 2**。
//
// ⚠ 原版是 `add` 之後 `rcr al,1`——進位會被轉回最高位，所以那是
// **9 位元的和再除以 2**，不是「先截成 8 位元再除」。屬性上限 18 × 2 遠不到 255，
// 兩種算法在遊戲裡不會分岔，但公式要寫對的那一個。
func hireBase(c *Character) int {
	if c == nil {
		return 0
	}
	return (int(c.Attributes[AttrCharisma]) + int(c.Attributes[AttrIQ])) / 2
}

// HireOutcome 是一次雇用嘗試的結果。
type HireOutcome struct {
	Joined bool
	// Free ＝ NPC 記錄 `+0x31` 是 0，**不必檢定直接加入**（`0x13304`）。
	Free bool
	// Roll／Ours／Theirs 是檢定的三個數字；Free 為 true 時都是 0。
	Roll, Ours, Theirs int
}

// TryHire 跑一次雇用檢定（`0x13304`–`0x13348`）。
//
//	NPC   ＝ (魅力＋智力)/2 ＋ price ＋ 等級
//	招募者 ＝ (魅力＋智力)/2 ＋ 魅力 ＋ 等級 ＋ 骰
//	骰 < 5 → 直接失敗；否則 招募者 ≥ NPC → 成功
//
// **魅力在招募者那一側算兩次**（平均裡一次、單獨再加一次）——官方手冊那句
// 「你想雇用 NPC 時，他的反應也看這一項」講的就是它。
func TryHire(hirer, npc *Character, price byte, r *rng.State) HireOutcome {
	if price == 0 {
		return HireOutcome{Joined: true, Free: true}
	}
	out := HireOutcome{
		Theirs: hireBase(npc) + int(price) + int(npc.Level),
		Ours:   hireBase(hirer) + int(hirer.Attributes[AttrCharisma]) + int(hirer.Level),
		Roll:   HireRoll(r),
	}
	if out.Roll < hireRollFloor {
		return out
	}
	out.Ours += out.Roll
	out.Joined = out.Ours >= out.Theirs
	return out
}

// HireNPC 從 section 17 的記錄做出一個隊員。
//
// **整筆記錄留著**（`Character.Source`）：原版就是逐 byte 抄 256 bytes 進隊伍的
// 記錄槽，未解的欄位也一起。存檔時把這一份原樣寫下去，已解欄位再蓋上去，
// 就與原版一致。
//
// charisma 是招募者的魅力，照原版寫進新隊員的 `+0x31`（`0x13372`）。
func HireNPC(rec []byte, charisma byte) *Character {
	if len(rec) < 0x100 {
		return nil
	}
	raw := make([]byte, 0x100)
	copy(raw, rec[:0x100])
	raw[recNPCPrice] = charisma
	c := LoadCharacter(raw)
	c.Source = raw
	return c
}

// HirePrice 是 NPC 記錄的 `+0x31`（雇用門檻）；0 ＝ 不必檢定。
func HirePrice(rec []byte) byte {
	if len(rec) <= recNPCPrice {
		return 0
	}
	return rec[recNPCPrice]
}

// HireGreeting 是 NPC 記錄的 `+0x30`：加入時要印的那條字串編號。
func HireGreeting(rec []byte) byte {
	if len(rec) <= recNPCGreeting {
		return 0
	}
	return rec[recNPCGreeting]
}

// HireCandidates 是 `loc_1382B`：**遭遇裡「數量欄非 0」的組數**。
//
// ⚠ 它數的不是「可以雇用的組」——原版在這一層不篩，
// 篩在結算時的 `+0x09`。把篩選提前會讓「按 H 之後才發現不能雇用」
// 這個原版行為消失。
func HireCandidates(groups [EnemyGroups]SpawnGroup) int {
	n := 0
	for _, g := range groups {
		if g.Count != 0 {
			n++
		}
	}
	return n
}

// RemoveGroup 把第 g 組整組從戰場上移除（`0x133A9`：原版把那一組敵方記錄的
// `+0x00`–`+0x13` 清成零）。
//
// ⚠ **是移除不是打死**：那一組沒有被殺，是加入了隊伍或走掉了，
// 所以不給經驗值也不留屍體。用 `Enemies[...] = nil` 而不是把 HP 歸零。
func (b *Battle) RemoveGroup(g int) bool {
	if g < 0 || g >= EnemyGroups {
		return false
	}
	for i := 0; i < EnemiesPerGroup; i++ {
		b.Enemies[g*EnemiesPerGroup+i] = nil
	}
	return true
}

// GroupAlive 回報第 g 組還有沒有活著的敵人。
func (b *Battle) GroupAlive(g int) bool {
	if g < 0 || g >= EnemyGroups {
		return false
	}
	for i := 0; i < EnemiesPerGroup; i++ {
		if e := b.Enemies[g*EnemiesPerGroup+i]; e != nil && e.HP > 0 {
			return true
		}
	}
	return false
}
