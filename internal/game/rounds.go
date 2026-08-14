package game

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// 戰鬥的回合結構（docs/spec/12、docs/re/36）。
//
// 規格 06 做的是「單次攻擊的結算」，這一層把它包成回合：
// 誰還沒行動、行動值怎麼算、一輪跑完換下一輪。

const (
	// EnemyGroups／EnemiesPerGroup 是戰場的固定形狀（docs/re/36 §1）。
	EnemyGroups     = 3
	EnemiesPerGroup = 10
	EnemySlots      = EnemyGroups * EnemiesPerGroup
)

// Combatant 是行動順序表裡的一格。
//
// 表是**固定格子**不是壓縮清單：第 g 組第 n 個永遠在 g*10+n，
// 所以索引直接對得回單位編號。
type Combatant struct {
	Slot      int   // 敵人是 g*10+n；隊伍成員是 EnemySlots + 隊伍索引
	IsParty   bool
	Initiative int
	Pending   bool // 「這回合還沒行動」——排進表就清掉
}

// Battle 是一場戰鬥。
type Battle struct {
	Enemies [EnemySlots]*Enemy // nil ＝ 這一格沒有敵人
	Party   *Party
	RNG     *rng.State

	order []Combatant
	Round int
}

// MaxRounds 是回合上限。原版沒有這個限制，但沒有上限的迴圈在
// 「行動順序未解」的情況下有機會轉不停——寧可有一個明顯的天花板。
const MaxRounds = 200

// NewBattle 建一場戰鬥。
func NewBattle(p *Party, r *rng.State) *Battle {
	return &Battle{Party: p, RNG: r}
}

// AddEnemy 把敵人放進第 g 組的第 n 格。
func (b *Battle) AddEnemy(group, index int, e *Enemy) bool {
	if group < 0 || group >= EnemyGroups || index < 0 || index >= EnemiesPerGroup {
		return false
	}
	b.Enemies[group*EnemiesPerGroup+index] = e
	return true
}

// BeginRound 把活著的單位排進行動表（docs/re/36 §2）。
//
// **回傳的順序是表的格子順序，不是行動順序**——誰先動要用 NextActor
// 一個一個挑（原版沒有排序演算法，見 §3）。
//
// 行動值 ＝ 2d6（逢同點續擲）＋ 該單位的一個欄位 × 8（docs/re/36 §2）。
// speedOf 讓呼叫者提供那個欄位——**原版取的是攻擊資料的哪個位移還沒對上**，
// 所以不猜，收進來。
func (b *Battle) BeginRound(speedOf func(c Combatant) int) []Combatant {
	b.Round++
	b.order = b.order[:0]

	add := func(c Combatant) {
		c.Pending = true
		roll := b.RNG.PairD6()
		c.Initiative = roll + speedOf(c)*8
		// 排進表就把旗標清掉——同一個單位一回合不得行動兩次
		// （原版 0x1AE06 就是這個動作，docs/re/36 §1）。
		c.Pending = false
		b.order = append(b.order, c)
	}

	for slot, e := range b.Enemies {
		if e == nil || e.HP == 0 {
			continue
		}
		add(Combatant{Slot: slot})
	}
	for i, m := range b.Party.Members {
		if m == nil || m.Dead() {
			continue
		}
		add(Combatant{Slot: EnemySlots + i, IsParty: true})
	}

	return b.order
}

// NextActor 挑出這一回合下一個要行動的單位，並把它從表裡清掉。
//
// 原版**沒有排序**：每次線性掃一遍挑最大的，做完把那一格清成 0，再掃一遍
// （`0x1AEAE`–`0x1AEE0`）。比較是 `A ≥ B`，所以**同值時取後面的格子**；
// 隊伍排在敵人後面，**行動值相同時隊伍先動**。
//
// 回傳 false 表示這一回合的人都動完了。
func (b *Battle) NextActor() (Combatant, bool) {
	best, found := -1, false
	for i, c := range b.order {
		if c.Initiative <= 0 {
			continue // 0 ＝ 已經動過（原版就是把那一格清成 0）
		}
		// ≥ 而不是 >：同值覆蓋，取後面的那一格。
		if !found || c.Initiative >= b.order[best].Initiative {
			best, found = i, true
		}
	}
	if !found {
		return Combatant{}, false
	}
	c := b.order[best]
	b.order[best].Initiative = 0
	return c, true
}

// Order 是這一回合排好的行動順序。
func (b *Battle) Order() []Combatant { return b.order }

// EnemiesLeft 是還活著的敵人數。
func (b *Battle) EnemiesLeft() int {
	n := 0
	for _, e := range b.Enemies {
		if e != nil && e.HP > 0 {
			n++
		}
	}
	return n
}

// PartyLeft 是還活著的隊員數。
func (b *Battle) PartyLeft() int {
	n := 0
	for _, m := range b.Party.Members {
		if m != nil && !m.Dead() {
			n++
		}
	}
	return n
}

// Over 回傳戰鬥有沒有結束，以及是不是隊伍贏了。
func (b *Battle) Over() (over, partyWon bool) {
	switch {
	case b.EnemiesLeft() == 0:
		return true, true
	case b.PartyLeft() == 0:
		return true, false
	case b.Round >= MaxRounds:
		return true, false
	default:
		return false, false
	}
}

// Enemy 取出某一格的敵人（沒有就回 nil）。
func (b *Battle) Enemy(slot int) *Enemy {
	if slot < 0 || slot >= EnemySlots {
		return nil
	}
	return b.Enemies[slot]
}

// Member 取出行動順序裡的隊伍成員。
func (b *Battle) Member(c Combatant) *Character {
	if !c.IsParty {
		return nil
	}
	i := c.Slot - EnemySlots
	if i < 0 || i >= len(b.Party.Members) {
		return nil
	}
	return b.Party.Members[i]
}
