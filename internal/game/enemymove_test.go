package game

// 敵人在地圖上移動的計畫（`docs/re/116`）。

import "testing"

// 九個步向的位移要與原版的跳表一致（`ds:AAB1h`）。
//
// ⚠ **0–3 與隊伍走路的方向編號是同一套**（上／下／左／右）。
// 憑直覺照「繞一圈」重排會讓敵人往反方向走，而畫面上看起來只是「它亂走」。
func TestStepDeltaMatchesJumpTable(t *testing.T) {
	want := [9][2]int{
		{0, -1}, {0, 1}, {-1, 0}, {1, 0}, {0, 0},
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1},
	}
	for i, w := range want {
		dx, dy, ok := StepDelta(i)
		if !ok || dx != w[0] || dy != w[1] {
			t.Errorf("步向 %d ＝ (%d, %d)，預期 (%d, %d)", i, dx, dy, w[0], w[1])
		}
	}
	if _, _, ok := StepDelta(9); ok {
		t.Error("步向 9 不存在，卻回了 ok")
	}
}

// `ds:AA05h` 的 3 × 3 與位移表要互為逆運算。
//
// 兩張表各自照抄，**互相驗證**：任一張抄錯就對不回來。
func TestStepTowardRoundTrips(t *testing.T) {
	for _, sy := range []int{-1, 0, 1} {
		for _, sx := range []int{-1, 0, 1} {
			step := StepToward(sx, sy)
			dx, dy, ok := StepDelta(step)
			if !ok || dx != sx || dy != sy {
				t.Errorf("(%d, %d) → 步向 %d → (%d, %d)", sx, sy, step, dx, dy)
			}
		}
	}
	if StepToward(0, 0) != StepStay {
		t.Error("兩軸都是 0 應該對到「不動」")
	}
	// 正負號要先取號：座標差不是 ±1 也要對得上。
	if StepToward(7, -3) != StepToward(1, -1) {
		t.Error("StepToward 沒有先取正負號")
	}
}

// 三條分數線的值（`ds:A5B8h` ＝ `fb 08 13`、`ds:A5BBh` ＝ `05 02 01`）。
func TestPositionScoreBands(t *testing.T) {
	// 帶 0：5k；帶 1：2k ＋ 10；帶 2：k ＋ 20（k ＝ ⌈距離／3⌉）。
	for _, tc := range []struct{ dist, band, want int }{
		{0, 0, 0}, {1, 0, 5}, {3, 0, 5}, {4, 0, 10}, {30, 0, 50},
		{0, 1, 10}, {1, 1, 12}, {30, 1, 30},
		{0, 2, 20}, {1, 2, 21}, {30, 2, 30},
	} {
		if got := PositionScore(tc.dist, tc.band); got != tc.want {
			t.Errorf("距離 %d、帶 %d ＝ %d，預期 %d", tc.dist, tc.band, got, tc.want)
		}
	}
	// 三條都遞增——「挑分數最小的那一格」才等於「靠近隊伍」。
	for band := 0; band < 3; band++ {
		for d := 1; d < 99; d++ {
			if PositionScore(d, band) < PositionScore(d-1, band) {
				t.Fatalf("帶 %d 在距離 %d 反轉了", band, d)
			}
		}
	}
}

// 射程帶的分類（`sub_156DE`）。
func TestRangeBand(t *testing.T) {
	for _, c := range []ItemClass{2, 5, 10, 13} {
		if RangeBand(c) != 0 {
			t.Errorf("類別 %d 應該是帶 0", c)
		}
	}
	for _, c := range []ItemClass{3, 6, 8, 11} {
		if RangeBand(c) != 1 {
			t.Errorf("類別 %d 應該是帶 1", c)
		}
	}
	for _, c := range []ItemClass{0, 1, 4, 7, 9, 12} {
		if RangeBand(c) != 2 {
			t.Errorf("類別 %d 應該是帶 2", c)
		}
	}
}

// 空曠地形上的環境：哪一格都站得上去。
func openEnv() MoveEnv {
	return MoveEnv{
		X: 20, Y: 20, PartyX: 20, PartyY: 20,
		Passable: func(x, y int) bool { return true },
		InView:   func(x, y int) bool { return true },
	}
}

// `+0x09` 的 bit2 設著 ＝ 這一筆遭遇整個不移動（`docs/re/101` §6.4）。
func TestStaysPutNeverMoves(t *testing.T) {
	e := openEnv()
	e.NoMove = true
	e.PartyY = 24
	e.Distance = 40
	e.Data = EnemyData{Kind: KindAnimal, Weapon: 0}
	if got := PlanEnemyMove(e); got != NoMovePlan {
		t.Errorf("計畫 ＝ %#x，預期不動", got)
	}
}

// 近戰敵人：太近就不動，夠遠就朝隊伍衝。
func TestMeleeChargesWhenFar(t *testing.T) {
	e := openEnv()
	e.Data = EnemyData{Kind: KindAnimal, Weapon: 0, Base: 20}
	e.GroupHP = 20

	// 距離 < 16 → 不動（`0x14CB7`）。
	e.PartyX, e.PartyY = 21, 20
	e.Distance = 10
	if got := PlanEnemyMove(e); got != NoMovePlan {
		t.Errorf("距離 10 應該不動，得到 %#x", got)
	}

	// 夠遠 → 朝隊伍衝，而且要真的更近。
	e.PartyX, e.PartyY = 24, 20
	e.Distance = 40
	got := PlanEnemyMove(e)
	if got&0xC0 != PlanCharge {
		t.Fatalf("計畫 %#x 不是「朝你衝」", got)
	}
	dx, dy, _ := StepDelta(PlanStep(got))
	if dx != 1 || dy != 0 {
		t.Errorf("往 (%d, %d) 衝，預期往右一格", dx, dy)
	}
}

// 士氣：整組血量低於基礎血量的一半就逃（種類 1 不做這個檢查）。
func TestMoraleMakesThemFlee(t *testing.T) {
	e := openEnv()
	e.Data = EnemyData{Kind: KindMutant, Weapon: 0, Base: 100}
	e.PartyX, e.PartyY = 24, 20
	e.Distance = 40
	e.GroupHP = 10 // < 100 / 2

	got := PlanEnemyMove(e)
	if got&0xC0 != PlanFlee {
		t.Fatalf("計畫 %#x 不是「逃走」", got)
	}
	dx, _, _ := StepDelta(PlanStep(got))
	if dx != -1 {
		t.Errorf("往 x %+d 跑，預期背離隊伍（−1）", dx)
	}

	// 同樣的血量給動物 → 不逃（動物不做士氣檢查）。
	e.Data.Kind = KindAnimal
	if got := PlanEnemyMove(e); got&0xC0 != PlanCharge {
		t.Errorf("動物不該逃，得到 %#x", got)
	}
}

// 遠程武器走「換位置」，而那道閘是「行動值 ＋ 種類常數 ≥ 現在這一格的分數」。
func TestRangedRepositionGate(t *testing.T) {
	e := openEnv()
	e.Data = EnemyData{Kind: KindMutant, Weapon: 2, Base: 20, Speed: 1} // 類別 2 ＝ 帶 0
	e.GroupHP = 20
	e.PartyX, e.PartyY = 26, 20
	e.Distance = 60 // 帶 0 的分數 ＝ 5 × ⌈60／3⌉ ＝ 100

	// 行動值 1 ＋ 常數 7 ＜ 100 → 這一格太不划算，動。
	got := PlanEnemyMove(e)
	if got&0xC0 != PlanReposition {
		t.Fatalf("計畫 %#x 不是「換位置」", got)
	}
	if dx, dy, _ := StepDelta(PlanStep(got)); dx != 1 || dy != 0 {
		t.Errorf("換到 (%+d, %+d)，預期靠近隊伍那一格", dx, dy)
	}

	// 行動值高到忍得下去 → 留在原地。
	e.Data.Speed = 200
	if got := PlanEnemyMove(e); got != NoMovePlan {
		t.Errorf("忍得下去應該不動，得到 %#x", got)
	}

	// 貼著隊伍時分數低（帶 0、距離 10 → 20），一般的行動值就忍得住。
	e.Data.Speed = 30
	e.PartyX = 21
	e.Distance = 10
	got = PlanEnemyMove(e)
	if got != NoMovePlan {
		t.Errorf("分數 20 ＜ 30 ＋ 7，應該不動，得到 %#x", got)
	}

	// 同樣的位置、行動值低 → 還是會動。
	e.Data.Speed = 1
	got = PlanEnemyMove(e)
	if got&0xC0 != PlanReposition {
		t.Fatalf("計畫 %#x 不是「換位置」", got)
	}
	dx, dy, _ := StepDelta(PlanStep(got))
	if dx != 1 || dy != 0 {
		t.Errorf("換到 (%+d, %+d)，預期靠近隊伍那一格", dx, dy)
	}
}

// 四周都走不動時就是不動——不要挑一個「看起來合理」的格子。
func TestBlockedEnemyStaysPut(t *testing.T) {
	e := openEnv()
	e.Passable = func(x, y int) bool { return false }
	e.Data = EnemyData{Kind: KindAnimal, Weapon: 0, Base: 20}
	e.GroupHP = 20
	e.PartyX, e.PartyY = 24, 20
	e.Distance = 40
	if got := PlanEnemyMove(e); got != NoMovePlan {
		t.Errorf("四周都不能站，卻回了 %#x", got)
	}
}

// 座標 0 是邊界不是格子（`0x14D69`、`0x14D6F`、`sub_14F26`）。
//
// ⚠ 少了這一條敵人會走到地圖外而**不會崩**——它只是消失。
func TestZeroCoordinateIsNotACell(t *testing.T) {
	e := openEnv()
	e.X, e.Y = 1, 1
	e.PartyX, e.PartyY = 0, 0 // 隊伍在角落（原版不會發生，這裡只逼出邊界）
	e.Data = EnemyData{Kind: KindAnimal, Weapon: 0, Base: 20}
	e.GroupHP = 20
	e.Distance = 40
	got := PlanEnemyMove(e)
	if got == NoMovePlan {
		return // 不動也是對的
	}
	dx, dy, _ := StepDelta(PlanStep(got))
	if e.X+dx == 0 || e.Y+dy == 0 {
		t.Errorf("走到座標 0（%d, %d）", e.X+dx, e.Y+dy)
	}
}
