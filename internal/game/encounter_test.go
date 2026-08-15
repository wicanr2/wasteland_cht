package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 一筆敵人資料：基礎值 base，其餘欄位隨便填但不全零。
func mkData(base uint16) []byte {
	return []byte{byte(base), byte(base >> 8), 4, 3, 0x02, 0x50, 0, 0}
}

// 驗收 1：型別 0 不生怪、數量 0 的組不佔格子。
func TestSpawnSkipsEmptyTypeAndCount(t *testing.T) {
	table := ParseEnemyTable(append(make([]byte, 8), mkData(20)...)) // 第 0 筆全零、第 1 筆有料
	record := make([]byte, 16)
	record[0x03], record[0x04] = 1, 3 // 第 0 組：型別 1、3 隻
	record[0x05], record[0x06] = 0, 5 // 第 1 組：型別 0（＝ 沒有敵人）
	record[0x07], record[0x08] = 1, 0 // 第 2 組：數量 0

	b := NewBattle(&Party{}, rng.New())
	if n := b.Spawn(record, table, rng.New()); n != 3 {
		t.Fatalf("應該只生出 3 隻，得到 %d", n)
	}
	for slot, e := range b.Enemies {
		want := slot < 3
		if (e != nil) != want {
			t.Fatalf("格子 %d：有敵人 ＝ %v，預期 %v", slot, e != nil, want)
		}
	}
}

// 驗收 2：血量落在 ⌊基礎/4⌋+1 … ⌊基礎/4⌋+基礎（基礎 < 256）。
// 驗收 3：同一組不會全部一樣。
func TestRollHPRangeAndVariety(t *testing.T) {
	r := rng.New()
	for _, base := range []uint16{1, 2, 20, 100, 255} {
		d := ParseEnemyData(mkData(base))
		lo, hi := int(base/4)+1, int(base/4)+int(base)
		seen := map[uint16]bool{}
		for i := 0; i < 500; i++ {
			hp := int(d.RollHP(r))
			if hp < lo || hp > hi {
				t.Fatalf("基礎 %d：血量 %d 不在 %d–%d", base, hp, lo, hi)
			}
			seen[uint16(hp)] = true
		}
		if base >= 20 && len(seen) < 5 {
			t.Errorf("基礎 %d：500 次只擲出 %d 種血量，看起來沒有逐隻擲", base, len(seen))
		}
	}
}

// 基礎值超過 255 時高位那一項要真的加進去（原版資料裡有 8 筆這種）。
func TestRollHPHighByteIsNotDeadCode(t *testing.T) {
	r := rng.New()
	const base = 390 // 0x0186：高位 1、低位 0x86
	d := ParseEnemyData(mkData(base))
	quarter := base / 4
	for i := 0; i < 200; i++ {
		hp := int(d.RollHP(r))
		// 高位固定貢獻 256 × 1d(1) ＝ 256。
		if hp < quarter+1+256 || hp > quarter+0x86+256 {
			t.Fatalf("基礎 %d：血量 %d 不在 %d–%d", base, hp, quarter+1+256, quarter+0x86+256)
		}
	}
}

// 驗收 4：數量超過 10 要被夾住並回報，不會蓋到下一組。
func TestSpawnClampsOversizedGroup(t *testing.T) {
	table := ParseEnemyTable(append(make([]byte, 8), mkData(20)...))
	record := make([]byte, 16)
	record[0x03], record[0x04] = 1, 25 // 第 0 組故意寫 25 隻

	groups := ReadSpawnGroups(record)
	if !groups[0].Clamped || groups[0].Count != EnemiesPerGroup {
		t.Fatalf("第 0 組應該被夾在 %d 並回報，得到 %+v", EnemiesPerGroup, groups[0])
	}

	b := NewBattle(&Party{}, rng.New())
	if n := b.Spawn(record, table, rng.New()); n != EnemiesPerGroup {
		t.Fatalf("應該只生出 %d 隻，得到 %d", EnemiesPerGroup, n)
	}
	for slot := EnemiesPerGroup; slot < EnemySlots; slot++ {
		if b.Enemies[slot] != nil {
			t.Fatalf("第 0 組溢出到格子 %d 了", slot)
		}
	}
}

// 驗收 5：距離表要與原版那 50 個 byte 逐格相同。
//
// 這裡把原版的值寫死當對照——Go 那一份是用公式生的，
// 兩邊獨立，才驗得出「公式 ＋ 兩個例外」有沒有寫對。
func TestDistanceTableMatchesOriginal(t *testing.T) {
	original := [50]byte{
		0x02, 0x0a, 0x14, 0x1e, 0x28, 0x32, 0x3c, 0x46, 0x50, 0x5a,
		0x0a, 0x0e, 0x16, 0x1f, 0x29, 0x33, 0x3d, 0x47, 0x51, 0x5c,
		0x14, 0x16, 0x1c, 0x24, 0x2c, 0x36, 0x3f, 0x49, 0x52, 0x5c,
		0x1e, 0x1f, 0x24, 0x2a, 0x32, 0x3a, 0x43, 0x4c, 0x55, 0x5f,
		0x28, 0x29, 0x2d, 0x32, 0x39, 0x40, 0x48, 0x51, 0x59, 0x62,
	}
	for dy := 0; dy <= DistanceMaxDY; dy++ {
		for dx := 0; dx <= DistanceMaxDX; dx++ {
			got, ok := Distance(dx, dy)
			if !ok {
				t.Fatalf("(%d, %d) 應該在範圍內", dx, dy)
			}
			if want := int(original[dy*10+dx]); got != want {
				t.Errorf("(%d, %d)：得到 %d，原版是 %d", dx, dy, got, want)
			}
		}
	}
	// 負的座標差要取絕對值。
	if a, _ := Distance(-3, -2); a != 36 {
		t.Errorf("(−3, −2) 應該與 (3, 2) 一樣是 36，得到 %d", a)
	}
	// 視窗外不是「距離 0」。
	if _, ok := Distance(10, 0); ok {
		t.Error("dx = 10 超出視窗，應該回報 false")
	}
	if _, ok := Distance(0, 5); ok {
		t.Error("dy = 5 超出視窗，應該回報 false")
	}
}

// 驗收 6：生成 → 打完，1,000 場不 panic、血量不溢位。
func TestSpawnedBattlesRunToCompletion(t *testing.T) {
	r := rng.New()
	table := ParseEnemyTable(append(append(make([]byte, 8), mkData(20)...), mkData(60)...))
	for game := 0; game < 1000; game++ {
		record := make([]byte, 16)
		record[0x03], record[0x04] = byte(1+game%2), byte(1+game%4)
		record[0x05], record[0x06] = byte(1+(game+1)%2), byte(game%3)

		p := &Party{}
		for i := 0; i < 1+game%4; i++ {
			p.Members = append(p.Members, &Character{CON: 20, MaxCON: 20})
		}
		b := NewBattle(p, r)
		if b.Spawn(record, table, r) == 0 {
			continue
		}
		for {
			if over, _ := b.Over(); over {
				break
			}
			b.BeginRound(allAttack)
			for {
				c, ok := b.NextActor()
				if !ok {
					break
				}
				if c.IsParty {
					for _, e := range b.Enemies {
						if e != nil && e.HP > 0 {
							e.TakeDamage(r, 5, 0)
							break
						}
					}
					continue
				}
				e := b.Enemy(c.Slot)
				if e == nil || e.HP == 0 {
					continue
				}
				for _, m := range b.Party.Members {
					if m != nil && !m.Dead() && m.CON > deathFloor {
						m.TakeDamage(r, EnemyDamage(r, e.Data))
						break
					}
				}
			}
		}
		for _, e := range b.Enemies {
			if e != nil && e.HP > 60000 {
				t.Fatalf("第 %d 場：敵人血量溢位成 %d", game, e.HP)
			}
		}
	}
}
