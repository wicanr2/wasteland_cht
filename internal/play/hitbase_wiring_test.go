package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 回合結算真的把**目標的移動計畫**當成命中基礎值（`ds:711Dh`，`docs/re/101` §5）。
//
// ⚠ 只驗 `HitBase` 的常數對不夠——那支在 `internal/game` 有自己的測試。
// 這一條驗的是「接上去了」：`round.go` 從 2026-08-16 之前一直傳寫死的 60，
// 而那時候 `HitBase` 就算存在也不會有任何斷言變紅。
func TestRoundUsesMovePlanAsHitBase(t *testing.T) {
	misses := func(plan int) int {
		s := mkBattle(t, 60000, 0) // 血夠厚，打不死才數得完
		s.Battle.MovePlan = plan
		n := 0
		for r := 0; r < 150; r++ {
			// 每回合把隊伍補滿：150 回合不補的話隊伍會先倒，
			// 而樣本不夠多的話兩邊的失手次數分不開。
			for _, m := range s.Battle.Party.Members {
				m.CON = m.MaxCON
			}
			s.BeginCommands()
			for !s.Done() {
				if !s.Choose('A', true) { // A ＝ Attack
					s.Choose(' ', true)
				}
			}
			res := s.ResolveRound()
			for _, l := range res.Lines {
				if strings.HasSuffix(l, " misses.") {
					n++
				}
			}
			if res.Over {
				t.Fatalf("計畫 %#02x：第 %d 回合就結束了，數不到命中率", plan, r)
			}
		}
		return n
	}
	still, moving := misses(game.NoMovePlan), misses(0x80)
	t.Logf("不動 %d 次失手、移動 %d 次失手", still, moving)
	// 會移動的敵人比較好打 → 失手次數要比較少。
	if moving >= still {
		t.Errorf("接反了：不動 %d 次失手、會移動 %d 次失手，"+
			"預期會移動的那一邊比較少", still, moving)
	}
}

// 從地圖開打的戰鬥，`MovePlan` 要是 `NoMovePlan`。
//
// remake 沒有實作敵人在地圖上移動（`docs/re/87` §2），所以沒有人會設它——
// 但零值 0 是一個**合法的步向**，忘了走 `NewBattle` 就會安靜地變成 60。
func TestEncounterBattleHasNoMovePlan(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(0, 12, 2); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 400 && !s.InCombat(); i++ {
		d := input.DirRight
		if i%2 == 1 {
			d = input.DirLeft
		}
		if _, err := s.Update(input.Input{Dir: d}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.InCombat() {
		t.Skip("400 步沒遇到敵人，這一輪測不到")
	}
	if got := s.Combat().Battle.MovePlan; got != game.NoMovePlan {
		t.Errorf("地圖遭遇開出來的戰鬥 MovePlan ＝ %d，預期 %d", got, game.NoMovePlan)
	}
}
