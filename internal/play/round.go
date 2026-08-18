package play

// 一回合的結算（docs/spec/22）。
//
// 規格 12 定了誰先動、規格 14 定了指令怎麼選、規格 21 把戰鬥接到地圖上。
// 這個檔補中間那一段：**指令問完之後發生什麼**。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// RoundResult 是一回合結算完的結果。
type RoundResult struct {
	Lines []string // 這一回合的訊息，照發生順序
	CJK   []string // 同一批訊息的中文（UTF-8）；查不到的那一句是空字串
	Over  bool
	Won   bool // Over 為 true 才有意義
}

// msgs 是一回合累積的訊息：英文與中文兩份，**長度一定相同**。
//
// 中文那一份查不到就放 nil——呈現層看到 nil 就顯示英文那一句，
// 而不是顯示半句中文（`zhJoin` 的同一條原則）。
type msgs struct {
	EN  []string
	CJK []string
}

func (m *msgs) add(en string, zh string) {
	m.EN = append(m.EN, en)
	m.CJK = append(m.CJK, zh)
}

func (m *msgs) join(o msgs) {
	m.EN = append(m.EN, o.EN...)
	m.CJK = append(m.CJK, o.CJK...)
}

// ResolveRound 把指令階段的結果跑完一回合。
//
// 指令是「這一回合」的，不是持續狀態——跑完就回到指令階段重問。
func (s *CombatScene) ResolveRound() RoundResult {
	var res RoundResult
	var out msgs
	b := s.Battle
	b.BeginRound(s.acting)

	for {
		actor, ok := b.NextActor()
		if !ok {
			break
		}
		if actor.IsParty {
			out.join(s.partyActs(actor))
		} else {
			out.join(s.enemyActs(actor))
		}
		if over, won := b.Over(); over {
			res.Over, res.Won = true, won
			break
		}
	}
	// 攻擊結算完才輪到「非攻擊」的那些指令（`docs/re/107` §1：
	// 原版四趟 `sub_12CFE` 的第 4 趟，排在行動順序表 `0x1AD70` 後面）。
	// **順序有後果**：這一回合裝填的子彈，這一回合的攻擊用不到。
	out.join(s.resolveCommands())

	if !res.Over {
		if over, won := b.Over(); over {
			res.Over, res.Won = true, won
		}
	}
	res.Lines, res.CJK = out.EN, out.CJK
	s.Log = append(s.Log, res.Lines...)
	return res
}

// acting 回答第 i 個隊伍成員這回合下的是不是攻擊令。
//
// 原版只把下攻擊令的人排進行動表（`0x1AE78` 的 `cmp al, 2`，docs/re/90 §2）——
// 迴避、換武器、使用物品的人**這一回合根本不會被叫到**。
func (s *CombatScene) acting(i int) bool {
	return i >= 0 && i < len(s.Phase.Cmd) && s.Phase.Cmd[i] == game.CmdAttack
}

// partyActs 是隊伍成員的一次行動。
func (s *CombatScene) partyActs(actor game.Combatant) msgs {
	b := s.Battle
	m := b.Member(actor)
	var out msgs
	if m == nil || m.Down() {
		return out
	}
	i := actor.Slot - game.EnemySlots
	if i < 0 || i >= len(s.Phase.Cmd) || s.Phase.Cmd[i] != game.CmdAttack {
		// Evade／Use／Hire／Load 這一版只影響防禦值（規格 14 已解的部分），
		// 其餘效果未解——**不做事，不猜**。
		return out
	}

	target, slot := s.firstEnemy()
	if target == nil {
		return out
	}
	// 命中累加值只跟隊伍成員的 Brawling 與 Agility、以及目標的行動值有關
	// （docs/re/88）——**裝備不進命中**，只進傷害。
	w := s.weaponOf(m)
	// 命中基礎值看**目標這一回合的移動計畫**（`ds:711Dh`，docs/re/101 §5）：
	// 會移動的 60、留在原地的 50。remake 沒有敵人地圖移動，所以是 50。
	acc := game.HitChance(m, game.HitBase(b.MovePlan), target.Data)
	if !game.PartyHits(b.RNG, acc) {
		// 整句就是原版字串 35（`\x0B misses.`），名字由 `\x0B` 代入。
		out.add(m.Name+" misses.",
			s.zhStr(strPartyMisses, textlayout.Options{Name: nameOf(m)}))
		return out
	}
	dmg := int(game.PartyDamage(b.RNG, m, w, 0))
	applied, killed := target.TakeDamage(b.RNG, dmg, 0)
	out.add(fmt.Sprintf("%s hits %s for %d", m.Name, s.enemyLabel(target), applied),
		s.zhHit(m.Name, s.zhEnemy(target), applied))
	if killed {
		out.add(s.enemyLabel(target)+" died!",
			zhJoin(s.zhEnemy(target), s.zhStr(strDied, textlayout.Options{})))
		xp := target.Data.KillXP()
		m.AddXP(xp)
		out.add(fmt.Sprintf("%s gains %d experience.", m.Name, xp),
			s.zhXP(m.Name, xp))
		_ = slot // 槽留著：原版打死不清格，Over() 看的是 HP
	}
	return out
}

// nameOf 是 `\x0B` 的名字來源。
func nameOf(c *game.Character) func() string {
	return func() string {
		if c == nil {
			return ""
		}
		return c.Name
	}
}

// zhHit 是命中那一句的中文。
//
// ⚠ **這一句的組法是重製版的**（`docs/re/86` §2 定的「主詞 動詞 受詞 傷害」），
// 原版把它拆成 47（`hits \x0F `）與 33（` for `）兩段、印的順序還沒逐指令讀完
// （`docs/re/40` §3 自己標著命中訊息那條沒確認）。所以走 `ui:` 而不是
// 假裝它是某一條原版字串。
func (s *CombatScene) zhHit(attacker, target string, dmg int) string {
	if s.UI == nil || target == "" {
		return ""
	}
	f := s.UI("combat.hit")
	if len(f) == 0 {
		return ""
	}
	return fmt.Sprintf(f, attacker, string(target), dmg)
}

// zhXP 是「某某獲得 N 點經驗值」——原版字串 39 ＋ 數字 ＋ 40。
func (s *CombatScene) zhXP(name string, xp uint32) string {
	return zhJoin(
		s.zhStr(strGainsXP, textlayout.Options{Name: func() string { return name }}),
		fmt.Sprintf("%d", xp),
		s.zhStr(strExperience, textlayout.Options{}))
}

// enemyActs 是敵人的一次行動。
//
// 目標是隨機挑的，挑到倒下的人就重抽（`pickEnemyTarget`，docs/re/89 §1）。
//
// `sub_15036` **不是**目標表——那一支是敵人在地圖上移動
//（`move to a better position.`／`run away.`／`run at you.`，docs/re/87）。
func (s *CombatScene) enemyActs(actor game.Combatant) msgs {
	var out msgs
	b := s.Battle
	e := b.Enemy(actor.Slot)
	if e == nil || e.HP == 0 {
		return out
	}
	target, targetIdx := s.pickEnemyTarget()
	if target == nil {
		return out
	}
	// 敵方命中要 ≥（docs/re/20 §2）——方向與隊伍相反，別寫反了。
	// 累加的仍是**被打的那個隊伍成員**的本事，敵人資料是攻擊者這一邊。
	//
	// 基礎值取的是**被打的那個人這回合下了什麼令**（`0x1B06D`，docs/re/20 §1.1）：
	// 迴避 60、攻擊 50、其餘 40。迴避的處理程式是空的，效果全在這一個數字上。
	base := s.Phase.Defence(targetIdx)
	if !game.EnemyHits(b.RNG, game.HitChance(target, base, e.Data)) {
		// 敵方那一句的主詞在字串外面：名稱 ＋ 字串 31（` miss\x0Aes\x0A\x0A.`）。
		// 一隻敵人所以取單數那一段（Count ＝ 1）。
		out.add(s.enemyLabel(e)+" misses.",
			zhJoin(s.zhEnemy(e), s.zhStr(strEnemyMisses, textlayout.Options{Count: 1})))
		return out
	}
	dmg := game.EnemyDamage(b.RNG, e.Data)
	applied := target.TakeDamage(b.RNG, dmg)
	out.add(fmt.Sprintf("%s hits %s for %d", s.enemyLabel(e), target.Name, applied),
		s.zhHitBy(s.zhEnemy(e), target.Name, applied))
	if target.Down() {
		// 原版這裡印的是傷勢等級（`sub_157D6`，docs/re/19 §4）；
		// remake 目前只報一句，條件照原版用 CON ≤ 0。
		out.add(target.Name+" died!",
			zhJoin(target.Name, s.zhStr(strDied, textlayout.Options{})))
	}
	return out
}

// zhHitBy 是敵方命中的中文，與 zhHit 同一句、主詞受詞對調。
func (s *CombatScene) zhHitBy(attacker, target string, dmg int) string {
	if attacker == "" {
		return ""
	}
	return s.zhHit(string(attacker), target, dmg)
}

// pickEnemyTarget 挑敵人這一下打誰（`0x1B054`，docs/re/89）。
//
//	al ← 隊伍人數；sub_18E41 ＝ roll(1..人數)
//	ds:CF84h ← al；sub_172BB → CF 設就 jb 回去**重抽**
//
// `sub_172BB` 是「CON 兩個 byte 不全為 0 且高位不為負」——**倒下的人打不到**。
// 重抽沒有次數上限：原版靠進這條路之前的 `sub_19D0E`（還有沒有可打的目標）
// 擋住死迴圈，這裡先數一次活人，數到 0 就直接回。
func (s *CombatScene) pickEnemyTarget() (*game.Character, int) {
	b := s.Battle
	alive := 0
	for _, m := range b.Party.Members {
		if m != nil && !m.Down() {
			alive++
		}
	}
	if alive == 0 {
		return nil, -1
	}
	for {
		// roll(1..人數) 是 1-based（原版的成員編號從 1 數）。
		i := b.RNG.Roll(len(b.Party.Members)) - 1
		if m := b.Party.Members[i]; m != nil && !m.Down() {
			return m, i
		}
	}
}

// enemyLabel 是敵人在訊息裡的稱呼（種類名稱，`docs/re/85`）。
// 查不到名稱時回 "It"——**不留空白**，不然訊息會變成 " misses."。
func (s *CombatScene) enemyLabel(e *game.Enemy) string {
	if n := s.Names.Name(e.Data.Kind); n != "" {
		return n
	}
	return "It"
}

// weaponOf 取這個人裝備的武器資料。
//
// 物品表在存檔區、每個存檔槽一份（docs/re/45 §2），開場時載進 `s.Items`。
// 沒有裝備或查不到就回零值——零值的 Dice ＝ 0，傷害會是 0 而不是崩掉。
func (s *CombatScene) weaponOf(c *game.Character) game.ItemData {
	// ⚠ `EquipIndex` 是**背包的槽號**（`Equip(slot)` 存進去的），
	// 不是物品 ID——要先取那一格的 ID 再查表。
	// 直接拿槽號當 ID 查會取到完全不相干的物品：出廠存檔的
	// Hell Razor 因此打出 112 點傷害（正確的武器只有 3 顆 d6）。
	//
	// ⚠ **而且槽號是 1 起算的**（`sub_19AC8`：`0xBB ＋ 2n`，陣列在 `+0xBD`，
	// `docs/re/103` §3）。用 `equippedSlot` 這一支，**不要自己再算一次索引**——
	// 名單那一欄與傷害計算算出不同的武器時，畫面上完全看不出來：
	// 出廠的 Angela Deth 名單印 `VP91Z 9mm`，而攻擊拿的是她的**彈匣**。
	slot, ok := equippedSlot(c)
	if !ok {
		return game.ItemData{}
	}
	w, ok := s.Items.Get(slot.ID)
	if !ok {
		return game.ItemData{}
	}
	return w
}

// firstEnemy 回第一個還活著的敵人與它的槽號。
func (s *CombatScene) firstEnemy() (*game.Enemy, int) {
	for slot := 0; slot < game.EnemySlots; slot++ {
		if e := s.Battle.Enemy(slot); e != nil && e.HP > 0 {
			return e, slot
		}
	}
	return nil, -1
}

// resolveCommands 跑「非攻擊」那些指令的結算（`docs/re/107` §1 的第 4 趟）。
//
// ⚠ **指令有兩張跳表。** 下令階段（`internal/game/handlers.go`）只檢查前提、
// 選參數；**動作在這裡才發生**。以前這一段是空的，所以玩家在戰鬥裡按
// `W` 換武器、`L` 裝填**什麼都不會發生**，而畫面上看不出任何異狀。
//
// 順序照原版：這一趟排在攻擊結算後面，所以**這一回合裝填的子彈這一回合用不到**。
func (s *CombatScene) resolveCommands() msgs {
	var out msgs
	b := s.Battle
	for i, m := range b.Party.Members {
		if m == nil || m.Down() {
			continue // 原版的 `sub_172BB`：CON ≤ 0 的人不結算
		}
		if i >= len(s.Phase.Cmd) {
			break
		}
		var r game.ResolveResult
		switch s.Phase.Cmd[i] {
		case game.CmdWeapon:
			r = game.ResolveWeapon(m, s.Phase.Arg[i], s.Items)
		case game.CmdLoad:
			r = game.ResolveLoad(m, s.Items)
		case game.CmdEvade:
			r = game.ResolveEvade()
		case game.CmdUse:
			out.join(s.resolveUse(i, m))
			continue
		case game.CmdHire:
			out.join(s.resolveHire(i, m))
			continue
		default:
			// 攻擊與逃跑在別的迴圈結算。
			continue
		}
		if en := resolveText(m.Name, r); en != "" {
			out.add(en, s.zhResolve(m, r))
		}
	}
	return out
}

// resolveText 把結算結果換成英文那一句（原版字串的內容，`docs/re/107`）。
func resolveText(name string, r game.ResolveResult) string {
	switch r.Message {
	case 0:
		return ""
	case game.MsgSwapsEquipment:
		return name + " swaps equipment."
	case game.MsgEvades:
		return name + " evades."
	case game.MsgReloads:
		return name + " reloads."
	case game.MsgNoMoreClips:
		return name + " has no more clips."
	case game.MsgCantBeReloaded:
		return name + "'s weapon can't be reloaded."
	case game.MsgWeaponJammed:
		return "Your weapon is jammed."
	}
	return ""
}

// zhResolve 取那一句的中文。
//
// ⚠ 卡彈那句原版走**字串表 2 第 152 條**，而這一層只接得到表 1；
// 表 1 第 56 條是**同一個句子**（攻擊時用的那份），所以指過去。
// 兩份的譯文要一致——不一致的話同一件事會有兩種說法。
func (s *CombatScene) zhResolve(m *game.Character, r game.ResolveResult) string {
	n := int(r.Message)
	if r.Message == game.MsgWeaponJammed {
		n = strJammed
	}
	return s.zhStr(n, textlayout.Options{Name: nameOf(m)})
}

// resolveUse 是戰鬥版 `USE` 的結算（`docs/re/107` §4、`docs/re/108` §1）。
//
// 參數拆回「選項 ＋ 方向」，編號另存一格；目標是**往那個方向一格**的地圖記錄
// （原版 `sub_164E0` 用九向位移表走一步，`docs/re/108` §2）。
// 施用本身走與地圖 `USE` 同一支 `Party.UseOn`——判定那一層原版就是同一份資料
// （`docs/re/92` §6）。
func (s *CombatScene) resolveUse(i int, m *game.Character) msgs {
	var out msgs
	kind, id, dir := s.UseParts(i)
	w := s.World
	if w == nil {
		return out
	}
	dx, dy := useDirDelta(dir)
	x, y := int(w.Party.X)+dx, int(w.Party.Y)+dy
	rec, _, err := w.Block.CellRecord(x, y)
	if err != nil || len(rec) == 0 {
		out.add(m.Name+" uses... Nothing happens.", s.zhUse(m, useNothing))
		return out
	}
	res := w.Party.UseOn(w.RNG, rec, m, kind, id, w.Skills)
	switch {
	case res.Hit < 0:
		out.add(m.Name+" uses... Nothing happens.", s.zhUse(m, useNothing))
	case res.Passed:
		out.add(m.Name+" uses... It works!", s.zhUse(m, useWorks))
	default:
		out.add(m.Name+" uses... It fails.", s.zhUse(m, useFails))
	}
	// 收尾照原版改寫那一格（`0x13D23`／`0x13D7C` 的 `sub_17CFF`）。
	w.PatchCell(x, y, rec, res.PatchAt)
	return out
}

// USE 三種結果的 `ui:` key——**這三句是重製版自己的話**，
// 原版把字串 105（`\x0B uses...`）與判定結果分開印，順序還沒逐指令讀完。
const (
	useNothing = "use.nothinghappens"
	useWorks   = "use.works"
	useFails   = "use.fails"
)

// zhUse 是那三句的中文。格式與地圖那條共用（`%s uses %s.` 少一個參數，
// 這裡把名字帶進去、道具名留空——戰鬥訊息區窄，原版也只印 `uses...`）。
func (s *CombatScene) zhUse(m *game.Character, key string) string {
	if s.UI == nil {
		return ""
	}
	f := s.UI(key)
	if len(f) == 0 {
		return ""
	}
	return fmt.Sprintf(f, m.Name, "")
}
