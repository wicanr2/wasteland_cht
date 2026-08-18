package play

// 敵人在地圖上移動的執行（`docs/re/116` §5）。

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// 移動是**把那一格搬過去**：舊格清成 nibble 0、新格寫回原本的 (nibble, 記錄)。
//
// ⚠ 少了清舊格那一步，地圖上會留下一個永遠打不完的遭遇——
// 而畫面上看起來只是「這裡又有敵人」。
func TestEnemyMoveCarriesTheCell(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(4, 18, 2); err != nil {
		t.Fatal(err)
	}
	c, err := s.StartEncounter()
	if err != nil || c == nil {
		t.Fatalf("開不了戰：%v", err)
	}
	x0, y0 := c.EncX, c.EncY
	if x0 <= 0 || y0 <= 0 {
		t.Fatalf("遭遇格是 (%d, %d)", x0, y0)
	}
	terrain0, record0, _, err := s.world.Block.At(x0, y0)
	if err != nil {
		t.Fatal(err)
	}

	// 找一個站得上去的鄰格，直接指定計畫（挑哪一格是 `internal/game` 的事，
	// 這裡驗的是搬運本身）。
	step := -1
	for i := 0; i < 9; i++ {
		if i == game.StepStay {
			continue
		}
		dx, dy, _ := game.StepDelta(i)
		if s.enemyCanStand(x0+dx, y0+dy) {
			step = i
			break
		}
	}
	if step < 0 {
		t.Skip("這一格四周都站不上去")
	}
	dx, dy, _ := game.StepDelta(step)
	c.Battle.MovePlan = game.PlanCharge | step

	en, _ := s.moveEnemies()
	if en == "" {
		t.Fatal("移動了卻沒有訊息")
	}
	if !strings.Contains(en, "run") {
		t.Errorf("訊息 %q 不是「朝你衝」那一句", en)
	}
	if c.EncX != x0+dx || c.EncY != y0+dy {
		t.Errorf("遭遇格在 (%d, %d)，預期 (%d, %d)", c.EncX, c.EncY, x0+dx, y0+dy)
	}
	if tr, rec, _, _ := s.world.Block.At(x0, y0); tr != 0 || rec != 0 {
		t.Errorf("舊格沒清乾淨：nibble %d、記錄 %d", tr, rec)
	}
	tr, rec, _, _ := s.world.Block.At(c.EncX, c.EncY)
	if tr != terrain0 || rec != record0 {
		t.Errorf("新格是 (%d, %d)，預期原本的 (%d, %d)", tr, rec, terrain0, record0)
	}
}

// 計畫說要走到站不上去的格子時**什麼都不做**（原版執行時再檢查一次）。
func TestEnemyMoveRechecksTheDestination(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(4, 18, 2); err != nil {
		t.Fatal(err)
	}
	c, err := s.StartEncounter()
	if err != nil || c == nil {
		t.Fatalf("開不了戰：%v", err)
	}
	// 隊伍腳下那一格永遠站不上去（原版的 `sub_16840`）。
	c.EncX, c.EncY = int(s.world.Party.X)+1, int(s.world.Party.Y)
	c.Battle.MovePlan = game.PlanCharge | 2 // 往左 ＝ 隊伍那一格
	if en, _ := s.moveEnemies(); en != "" {
		t.Errorf("走進隊伍身上了：%q", en)
	}
}

// 空地 `ENC` 的那種回合沒有遭遇格，計畫一律是不動。
//
// ⚠ 少了這道判斷會拿 (0, 0) 去搬地圖那一格——**而那是合法座標**，
// 症狀是地圖左上角冒出一個遭遇。
func TestEmptyRoundHasNoMovePlan(t *testing.T) {
	s := newScene(t)
	if _, err := s.beginEmptyRound(); err != nil {
		t.Fatal(err)
	}
	s.planEnemyMove()
	if got := s.combat.Battle.MovePlan; got != game.NoMovePlan {
		t.Errorf("空回合的計畫 ＝ %#x，預期不動", got)
	}
	if en, _ := s.moveEnemies(); en != "" {
		t.Errorf("空回合不該有人移動：%q", en)
	}
}

// 命中基礎值跟著計畫走：這一條擋的是「移動接上了，60 那一支卻沒亮」。
func TestHitBaseFollowsTheMovePlan(t *testing.T) {
	if game.HitBase(game.NoMovePlan) != 50 {
		t.Error("不動的敵人基礎值應該是 50")
	}
	if game.HitBase(game.PlanCharge|3) != 60 {
		t.Error("會移動的敵人基礎值應該是 60")
	}
}

// 端到端：三格外的遠程敵人**真的會靠過來**。
//
// ⚠ 這一條擋的是「機制接上了，實際上永遠不動」。單元測試餵的是自己組的
// `MoveEnv`，餵錯了照樣全綠；這裡走的是出貨資料 ＋ 真的地圖 ＋ 真的距離。
func TestEnemyClosesInOverARound(t *testing.T) {
	s := newScene(t)
	// 地圖 4 的 (18, 2) 是一格遭遇；把隊伍放在三格外。
	if err := s.LoadMap(4, 21, 2); err != nil {
		t.Fatal(err)
	}
	c, err := s.StartEncounter()
	if err != nil || c == nil {
		t.Fatalf("開不了戰：%v", err)
	}
	px, py := int(s.world.Party.X), int(s.world.Party.Y)
	before, ok := game.Distance(c.EncX-px, c.EncY-py)
	if !ok {
		t.Fatalf("遭遇格 (%d, %d) 量不到距離", c.EncX, c.EncY)
	}
	if c.Battle.MovePlan == game.NoMovePlan {
		t.Fatalf("距離 %d 的遠程敵人卻不動（計畫沒算出來？）", before)
	}
	// 命中基礎值跟著亮起來（`docs/re/101` §7 的整個重點）。
	if game.HitBase(c.Battle.MovePlan) != 60 {
		t.Error("會移動的敵人基礎值應該是 60")
	}

	en, _ := s.moveEnemies()
	if en == "" {
		t.Fatal("有計畫卻沒有移動")
	}
	after, ok := game.Distance(c.EncX-px, c.EncY-py)
	if !ok {
		t.Fatalf("移動之後 (%d, %d) 量不到距離", c.EncX, c.EncY)
	}
	if after >= before {
		t.Errorf("距離從 %d 變成 %d，沒有靠近", before, after)
	}
}
