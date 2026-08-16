package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 隊伍攻擊的命中基礎值看目標這一回合的移動計畫（`ds:711Dh`，`docs/re/101` §5）。
func TestHitBaseFollowsMovePlan(t *testing.T) {
	if got := HitBase(NoMovePlan); got != 50 {
		t.Errorf("沒有移動計畫要用 50，得到 %d", got)
	}
	// 值的版面是「高 2 位訊息 ＋ 低 6 位步向」，三種訊息都算「有計畫」。
	for _, plan := range []int{0x00, 0x08, 0x40, 0x48, 0x80, 0x88} {
		if got := HitBase(plan); got != 60 {
			t.Errorf("計畫 %#02x 要用 60，得到 %d", plan, got)
		}
	}
}

// ⚠ `MovePlan` 的零值 0 是一個**合法的步向**，不是「沒有計畫」。
// `NewBattle` 沒有明寫的話命中基礎值會安靜地變成 60——
// 沒有任何斷言會紅，只有命中率會差 10 個百分點。
func TestNewBattleHasNoMovePlan(t *testing.T) {
	b := NewBattle(&Party{}, rng.New())
	if b.MovePlan != NoMovePlan {
		t.Fatalf("新戰鬥的 MovePlan 是 %d，預期 %d（NoMovePlan）", b.MovePlan, NoMovePlan)
	}
	if got := HitBase(b.MovePlan); got != 50 {
		t.Errorf("新戰鬥的命中基礎值是 %d，預期 50", got)
	}
}

// 基礎值 50 與 60 在判定上真的差一截——常數對不代表接上去的方向對。
//
// 隊伍那條是 `roll(1..100) < 累加值` 才命中，所以**值越大越容易打中**。
// 寫反的話這一條會紅，而單看常數的測試不會。
func TestHitBaseDirectionOnPartyAttack(t *testing.T) {
	c := &Character{Skills: []Slot{{ID: SkillBrawling, Value: 0}}}
	foe := EnemyData{Speed: 5, Weapon: ClassRifle}

	const iter = 20000
	rate := func(plan int) float64 {
		r := rng.New()
		acc := HitChance(c, HitBase(plan), foe)
		hits := 0
		for i := 0; i < iter; i++ {
			if PartyHits(r, acc) {
				hits++
			}
		}
		return float64(hits) / iter
	}
	still, moving := rate(NoMovePlan), rate(0x80)
	if !(moving > still) {
		t.Fatalf("會移動的敵人應該比較好打：不動 %.3f、移動 %.3f", still, moving)
	}
	if d := moving - still; d < 0.05 || d > 0.15 {
		t.Errorf("兩者差距應該接近 0.10（基礎值差 10），得到 %.3f", d)
	}
}
