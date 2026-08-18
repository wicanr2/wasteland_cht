package play

// 敵人在地圖上移動：計畫（`internal/game`）與執行（這裡）。
//
// 執行要動地圖與遭遇那一格，所以它在 `Scene` 這一層——
// `game.PlanEnemyMove` 只算「往哪一格走」，一個 byte 都不改。
//
// 原版一個戰鬥回合的順序（`sub_11CD0`，`docs/re/101` §5）：
//
//	sub_14BF0  算計畫      ← 隊伍下令之前，所以這一回合的命中基礎值用得到
//	sub_11F76  下令與結算
//	sub_15036  執行移動    ← 結算之後
//
// 這裡照同一個順序：開場與每回合結算完各算一次計畫，結算完執行。

import (
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/render"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// 三句移動訊息在執行檔字串表 1 的編號（`ds:A643h`，`docs/re/87` §1）。
//
// ⚠ 順序就是計畫 byte 的高 2 位：0 換位置、1 逃跑、2 朝你衝。
var moveStrings = [3]int{72, 74, 75}

// planEnemyMove 算這一回合的移動計畫，結果放進 `Battle.MovePlan`。
//
// **算不出來就是 `NoMovePlan`**，命中基礎值跟著回 50（`game.HitBase`）。
func (s *Scene) planEnemyMove() {
	c := s.combat
	if c == nil || c.Battle == nil {
		return
	}
	c.Battle.MovePlan = game.NoMovePlan
	// 空地 `ENC` 那種回合沒有遭遇格，沒有東西可以搬。
	if c.EncX <= 0 || c.EncY <= 0 || s.world == nil {
		return
	}
	e, slot := c.firstEnemy()
	if e == nil {
		return
	}
	px, py := int(s.world.Party.X), int(s.world.Party.Y)
	dist, ok := game.Distance(c.EncX-px, c.EncY-py)
	if !ok {
		return // 已經在視窗外，原版那張距離表也查不到
	}
	c.Battle.MovePlan = game.PlanEnemyMove(game.MoveEnv{
		X: c.EncX, Y: c.EncY,
		PartyX: px, PartyY: py,
		Data:     e.Data,
		Distance: dist,
		GroupHP:  c.groupHP(slot / game.EnemiesPerGroup),
		NoMove:   game.StaysPut(c.EncRecord),
		Passable: s.enemyCanStand,
		InView:   s.inMapView,
	})
}

// groupHP 是這一組還剩多少血（`sub_14FDE` 的十筆 16-bit 血量總和）。
func (s *CombatScene) groupHP(group int) int {
	total := 0
	for i := 0; i < game.EnemiesPerGroup; i++ {
		if e := s.Battle.Enemy(group*game.EnemiesPerGroup + i); e != nil {
			total += int(e.HP)
		}
	}
	return total
}

// enemyCanStand 回答「這一格站得上去嗎」（`sub_17C20` ＋ `sub_16840`）。
//
// ⚠ **隊伍腳下那一格也不行**：原版那道 `sub_16840` 擋的正是「走進別人身上」。
func (s *Scene) enemyCanStand(x, y int) bool {
	if s.world == nil || s.world.Block == nil {
		return false
	}
	if x == int(s.world.Party.X) && y == int(s.world.Party.Y) {
		return false
	}
	if _, _, _, err := s.world.Block.At(x, y); err != nil {
		return false
	}
	return s.world.Passable(x, y)
}

// inMapView 回答「這一格還在地圖視窗裡嗎」（`sub_169CF`）。
//
// **只有「換位置」那一支查它**——追出畫面的敵人在原版是走得出去的。
func (s *Scene) inMapView(x, y int) bool {
	if s.world == nil {
		return false
	}
	dx, dy := x-s.world.ViewX, y-s.world.ViewY
	return dx >= 0 && dx < render.ViewCols && dy >= 0 && dy < render.ViewRows
}

// moveEnemies 執行這一回合的移動計畫（`sub_15036`）。
//
// 回傳那一句話（英文、中文）；沒動就是兩個空字串。
func (s *Scene) moveEnemies() (en, cjk string) {
	c := s.combat
	if c == nil || c.Battle == nil || s.world == nil || s.world.Block == nil {
		return "", ""
	}
	plan := c.Battle.MovePlan
	if plan == game.NoMovePlan {
		return "", ""
	}
	dx, dy, ok := game.StepDelta(game.PlanStep(plan))
	if !ok || (dx == 0 && dy == 0) {
		return "", ""
	}
	nx, ny := c.EncX+dx, c.EncY+dy
	// ⚠ **執行時再檢查一次**（原版 `0x15097`／`0x1509C` 也是）：
	// 計畫是下令之前算的，這一回合裡地圖可能被腳本改過。
	if !s.enemyCanStand(nx, ny) {
		return "", ""
	}
	terrain, record, _, err := s.world.Block.At(c.EncX, c.EncY)
	if err != nil {
		return "", ""
	}
	// 舊格清成 nibble 0、記錄 0，新格寫回原本的 (nibble, 記錄)。
	// **順序不能反**：先寫新格再清舊格的話，兩格相鄰時會把剛寫好的清掉。
	if err := s.world.Block.SetCell(c.EncX, c.EncY, 0, 0); err != nil {
		return "", ""
	}
	if err := s.world.Block.SetCell(nx, ny, terrain, record); err != nil {
		return "", ""
	}
	c.EncX, c.EncY = nx, ny
	s.dirty = true

	// 那一句話：主詞是這一組的名字，單複數由 `\n` 那一套控制碼決定
	// （`docs/re/17` §4.1）。
	e, _ := c.firstEnemy()
	if e == nil {
		return "", ""
	}
	n := moveStrings[game.PlanMessage(plan)]
	count := c.groupCount(e)
	opt := textlayout.Options{Count: count}
	en = c.enemyLabel(e) + s.exeString(n)
	if zh := c.zhEnemy(e); zh != "" {
		if body := s.cjkExe(exeTable1, n, opt); body != "" {
			cjk = zh + body
		}
	}
	return en, cjk
}

// groupCount 數這一組還活著幾隻（單複數要用）。
func (s *CombatScene) groupCount(e *game.Enemy) int {
	n := 0
	for i := 0; i < game.EnemySlots; i++ {
		if x := s.Battle.Enemy(i); x != nil && x.HP > 0 && x.Data.Kind == e.Data.Kind {
			n++
		}
	}
	return n
}
