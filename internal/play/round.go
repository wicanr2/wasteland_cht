package play

// 一回合的結算（docs/spec/22）。
//
// 規格 12 定了誰先動、規格 14 定了指令怎麼選、規格 21 把戰鬥接到地圖上。
// 這個檔補中間那一段：**指令問完之後發生什麼**。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// 命中判定的基礎值（docs/re/20 §1.2、§1.3）。
//
// 兩條路徑各有 60／50（隊伍）與 60／50／40（敵方）兩三種，
// 挑哪一種由目標的一個欄位決定，**那個欄位的語意未解**（docs/spec/06 §6）。
// 這一版一律用「否則」那一支的 60，並且標明是暫代——
// ⚠ **不要用 0**：累加值 0 等於永遠打不中，而症狀是「戰鬥打不完」，
// 看起來像回合迴圈壞了。
const baseHitDefault = 60

// RoundResult 是一回合結算完的結果。
type RoundResult struct {
	Lines []string // 這一回合的訊息，照發生順序
	Over  bool
	Won   bool // Over 為 true 才有意義
}

// speedOf 是行動值那個「單位的一個欄位 × 8」。
//
// ⚠ **原版取的是攻擊資料的哪個位移還沒對上**（docs/re/36 §2），
// 所以這裡用敵人資料 `+0x02`（行動值，已確認）與隊伍的 0，
// 並且標明隊伍那一邊是暫代——不要當成解出來的。
func (s *CombatScene) speedOf(c game.Combatant) int {
	if c.IsParty {
		return 0 // 暫代
	}
	if e := s.Battle.Enemy(c.Slot); e != nil {
		return int(e.Data.Speed)
	}
	return 0
}

// ResolveRound 把指令階段的結果跑完一回合。
//
// 指令是「這一回合」的，不是持續狀態——跑完就回到指令階段重問。
func (s *CombatScene) ResolveRound() RoundResult {
	var res RoundResult
	b := s.Battle
	b.BeginRound(s.speedOf)

	for {
		actor, ok := b.NextActor()
		if !ok {
			break
		}
		if actor.IsParty {
			res.Lines = append(res.Lines, s.partyActs(actor)...)
		} else {
			res.Lines = append(res.Lines, s.enemyActs(actor)...)
		}
		if over, won := b.Over(); over {
			res.Over, res.Won = true, won
			break
		}
	}
	if !res.Over {
		if over, won := b.Over(); over {
			res.Over, res.Won = true, won
		}
	}
	s.Log = append(s.Log, res.Lines...)
	return res
}

// partyActs 是隊伍成員的一次行動。
func (s *CombatScene) partyActs(actor game.Combatant) []string {
	b := s.Battle
	m := b.Member(actor)
	if m == nil || m.Dead() {
		return nil
	}
	i := actor.Slot - game.EnemySlots
	if i < 0 || i >= len(s.Phase.Cmd) || s.Phase.Cmd[i] != game.CmdAttack {
		// Evade／Use／Hire／Load 這一版只影響防禦值（規格 14 已解的部分），
		// 其餘效果未解——**不做事，不猜**。
		return nil
	}

	target, slot := s.firstEnemy()
	if target == nil {
		return nil
	}
	// ⚠ 命中累加值要武器技能與射程；裝備欄還沒解到能自動判斷（規格 22 §5），
	// 所以這一版用「已裝備槽的物品資料」＋ 距離 0。**這是暫代。**
	w := s.weaponOf(m)
	acc := game.HitChance(m, baseHitDefault, w.Skill, 0, 0)
	if !game.PartyHits(b.RNG, acc) {
		return []string{m.Name + " misses."}
	}
	dmg := int(game.PartyDamage(b.RNG, m, w, 0))
	applied, killed := target.TakeDamage(b.RNG, dmg, 0)
	out := []string{fmt.Sprintf("%s hits for %d", m.Name, applied)}
	if killed {
		out = append(out, "died!")
		xp := target.Data.KillXP()
		m.AddXP(xp)
		out = append(out, fmt.Sprintf("%s gains %d experience.", m.Name, xp))
		_ = slot // 槽留著：原版打死不清格，Over() 看的是 HP
	}
	return out
}

// enemyActs 是敵人的一次行動。
//
// ⚠ **目標選擇是暫代**：原版走 `sub_15036` 的目標表（docs/re/51 §8 只讀到
// 頂層形狀），這裡用「還能行動的人裡第一個」。不假裝已經解了。
func (s *CombatScene) enemyActs(actor game.Combatant) []string {
	b := s.Battle
	e := b.Enemy(actor.Slot)
	if e == nil || e.HP == 0 {
		return nil
	}
	var target *game.Character
	for _, m := range b.Party.Members {
		if m != nil && !m.Dead() {
			target = m
			break
		}
	}
	if target == nil {
		return nil
	}
	// 敵方命中要 ≥（docs/re/20 §2）——方向與隊伍相反，別寫反了。
	if !game.EnemyHits(b.RNG, game.HitChance(target, baseHitDefault, 0, 0, 0)) {
		return []string{"misses."}
	}
	dmg := game.EnemyDamage(b.RNG, e.Data)
	applied := target.TakeDamage(b.RNG, dmg)
	out := []string{fmt.Sprintf("hits %s for %d", target.Name, applied)}
	if target.Dead() {
		out = append(out, target.Name+" died!")
	}
	return out
}

// weaponOf 取這個人裝備的武器資料。
//
// ⚠ **暫代**：物品表在存檔區、每個存檔槽一份（docs/re/45 §2），
// 這一層還沒接到那張表，所以沒有裝備時回零值——零值的 Dice ＝ 0，
// 傷害會是 0 而不是崩掉。接上物品表之後這一支要換掉。
func (s *CombatScene) weaponOf(c *game.Character) game.ItemData {
	w, ok := s.Items.Get(c.EquipIndex)
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
