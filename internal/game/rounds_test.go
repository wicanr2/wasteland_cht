package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

func mkBattle(enemies, members int) *Battle {
	p := &Party{}
	for i := 0; i < members; i++ {
		p.Members = append(p.Members, &Character{CON: 20, MaxCON: 20, AC: 0})
	}
	b := NewBattle(p, rng.New())
	for i := 0; i < enemies; i++ {
		b.AddEnemy(i/EnemiesPerGroup, i%EnemiesPerGroup, &Enemy{HP: 10})
	}
	return b
}

// 驗收 1／2：活著的才排進去、死的不排，而且排完不留 Pending。
func TestBeginRoundSkipsDead(t *testing.T) {
	b := mkBattle(5, 3)
	b.Enemies[2].HP = 0        // 死了一個敵人
	b.Party.Members[1].CON = 0 // 死了一個隊員
	order := b.BeginRound(func(Combatant) int { return 1 })

	if len(order) != 4+2 {
		t.Fatalf("應該排進 4 個敵人 ＋ 2 個隊員，得到 %d", len(order))
	}
	seen := map[int]bool{}
	for _, c := range order {
		if c.Pending {
			t.Fatalf("排完之後不該還是 Pending：%+v", c)
		}
		if seen[c.Slot] {
			t.Fatalf("格子 %d 排了兩次——同一個單位一回合不得行動兩次", c.Slot)
		}
		seen[c.Slot] = true
	}
	if seen[2] {
		t.Error("死掉的敵人不該排進行動順序")
	}
	if seen[EnemySlots+1] {
		t.Error("死掉的隊員不該排進行動順序")
	}
}

// 驗收 3：格子位置 ＝ g*10 + n。
func TestEnemySlotLayout(t *testing.T) {
	b := mkBattle(0, 1)
	if !b.AddEnemy(2, 7, &Enemy{HP: 5}) {
		t.Fatal("第 2 組第 7 格應該放得進去")
	}
	if b.Enemy(2*EnemiesPerGroup+7) == nil {
		t.Fatal("格子位置不對")
	}
	if b.AddEnemy(3, 0, &Enemy{HP: 1}) {
		t.Error("只有三組，第 3 組不該放得進去")
	}
	if b.AddEnemy(0, EnemiesPerGroup, &Enemy{HP: 1}) {
		t.Error("每組只有 10 格")
	}
}

// 行動值 ＝ 2d6 ＋ 欄位 × 8，所以欄位大的先動。
func TestInitiativeUsesSpeedField(t *testing.T) {
	b := mkBattle(2, 0)
	const fast = 0
	if n := len(b.BeginRound(func(c Combatant) int {
		if c.Slot == fast {
			return 5
		}
		return 0
	})); n != 2 {
		t.Fatalf("應該有兩個單位，得到 %d", n)
	}
	// 2d6 最大 12（續擲會更大，但 5×8 ＝ 40 的差距吃得掉一般情況）。
	c, ok := b.NextActor()
	if !ok || c.Slot != fast {
		t.Fatalf("欄位 5（＝ ＋40）的應該先動，得到 %+v", c)
	}
}

// 每個單位一回合只動一次，動完表就空了。
func TestNextActorConsumesEachUnitOnce(t *testing.T) {
	b := mkBattle(4, 2)
	b.BeginRound(func(Combatant) int { return 1 })
	seen := map[int]int{}
	for {
		c, ok := b.NextActor()
		if !ok {
			break
		}
		seen[c.Slot]++
	}
	if len(seen) != 6 {
		t.Fatalf("應該有 6 個單位行動，得到 %d", len(seen))
	}
	for slot, n := range seen {
		if n != 1 {
			t.Fatalf("格子 %d 動了 %d 次", slot, n)
		}
	}
	if _, ok := b.NextActor(); ok {
		t.Fatal("都動完了還挑得出人")
	}
}

// 同值時取後面的格子——隊伍排在敵人後面，所以平手時隊伍先動
// （原版的比較是 A ≥ B，docs/re/36 §3）。
func TestTiesFavourLaterSlot(t *testing.T) {
	b := mkBattle(1, 1)
	// 讓兩邊的行動值一樣：BeginRound 之後直接覆寫。
	b.BeginRound(func(Combatant) int { return 0 })
	for i := range b.order {
		b.order[i].Initiative = 7
	}
	c, ok := b.NextActor()
	if !ok {
		t.Fatal("挑不出人")
	}
	if !c.IsParty {
		t.Fatalf("平手時應該是隊伍先動（格子編號比較大），得到 %+v", c)
	}
}

// 驗收 4：一方全滅就結束，而且有回合上限。
func TestBattleOver(t *testing.T) {
	b := mkBattle(1, 1)
	if over, _ := b.Over(); over {
		t.Fatal("兩邊都活著不該結束")
	}
	b.Enemies[0].HP = 0
	over, won := b.Over()
	if !over || !won {
		t.Fatalf("敵人全滅應該是隊伍贏：over=%v won=%v", over, won)
	}

	b2 := mkBattle(1, 1)
	b2.Party.Members[0].CON = 0
	over, won = b2.Over()
	if !over || won {
		t.Fatalf("隊伍全滅應該是輸：over=%v won=%v", over, won)
	}

	b3 := mkBattle(1, 1)
	b3.Round = MaxRounds
	if over, _ := b3.Over(); !over {
		t.Fatal("回合上限到了應該結束")
	}
}

// 驗收 5：跑 1,000 場，不得 panic、HP 不得為負、CON 破 −50 不得還活著。
func TestFullBattlesAreStable(t *testing.T) {
	r := rng.New()
	for game := 0; game < 1000; game++ {
		b := mkBattle(3+game%5, 1+game%4)
		speed := func(Combatant) int { return game % 4 }
		for {
			if over, _ := b.Over(); over {
				break
			}
			b.BeginRound(speed)
			for {
				c, ok := b.NextActor()
				if !ok {
					break
				}
				if c.IsParty {
					m := b.Member(c)
					if m == nil || m.Dead() {
						continue
					}
					// 打第一個活著的敵人。
					for slot, e := range b.Enemies {
						if e == nil || e.HP == 0 {
							continue
						}
						dmg := int(PartyDamage(m, 1))
						e.TakeDamage(r, dmg, 0)
						_ = slot
						break
					}
					continue
				}
				e := b.Enemy(c.Slot)
				if e == nil || e.HP == 0 {
					continue
				}
				for _, m := range b.Party.Members {
					// 已經倒下的不再是目標——原版由 sub_172BB 把關
					// （那支還沒逐行讀，這裡用傷勢等級當等價條件）。
					if m == nil || m.Dead() || m.CON <= deathFloor {
						continue
					}
					m.TakeDamage(r, 3)
					break
				}
			}
			// 每個單位一回合只行動一次，所以回合數不會爆。
			if b.Round > MaxRounds {
				t.Fatalf("第 %d 場超過回合上限", game)
			}
		}
		for _, e := range b.Enemies {
			if e != nil && e.HP > 60000 {
				t.Fatalf("第 %d 場的敵人 HP 溢位成 %d", game, e.HP)
			}
		}
		// ⚠ **CON 破 −50 不會在扣血當下歸零**——那是時鐘每 16 刻做的
		// （sub_1251E，docs/re/35 §2）。所以戰鬥中 CON 可以是 −60，
		// 這裡驗的是「沒有溢位翻正」而不是「不得低於 −50」。
		for _, m := range b.Party.Members {
			if m == nil {
				continue
			}
			if m.CON > m.MaxCON {
				t.Fatalf("第 %d 場有人 CON %d 超過上限 %d，八成是溢位", game, m.CON, m.MaxCON)
			}
		}
		// 時鐘跑一次之後，破 −50 的要被歸零。
		b.Party.Selected = -1
		for i := 0; i < 64; i++ {
			b.Party.Tick16(0)
		}
		for _, m := range b.Party.Members {
			if m != nil && m.CON < deathFloor {
				t.Fatalf("第 %d 場：時鐘跑過之後 CON %d 還在 −50 以下", game, m.CON)
			}
		}
	}
}
