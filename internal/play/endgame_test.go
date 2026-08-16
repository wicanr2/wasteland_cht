package play

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 四把鑰匙的物品編號（字串表 `exe:2:36+n`）。
const (
	itemBlackstarKey = 61
	itemNovaKey      = 73
	itemPulsarKey    = 78
	itemQuasarKey    = 79
)

// 科奇斯基地反應爐層（資源 20）的四個角，也就是四根圓柱與後來的四個控制面板。
var cochiseCorners = [4][2]int{{1, 1}, {30, 1}, {30, 30}, {1, 30}}

// TestCochiseEndgame 從四把鑰匙走到結局，全程走 `Scene.Update`。
//
// 這是「這個遊戲玩得完」的驗收。整條鏈子是：
//
//	黑星／新星／脈衝星／類星體鑰匙 → 四根圓柱依序解鎖（USE，收尾改寫地圖格）
//	→ 四個角變成「按下按鈕以啟動安全程序 #1342-666」
//	→ 紅、黃、綠、藍四站依序啟動（nibble 8 問答）
//	→ 自毀警報，那一格變成腳本 opcode 35 → 倒數 240 刻
//	→ 主迴圈的 `sub_1CB30` 合成一個 kind 4 的分派 → 結局
//
// 中間有三段以前是斷的：USE 不改寫地圖格、nibble 8 沒有呈現層、
// opcode 35 沒實作（`docs/re/100` §3）。任何一段斷掉這條就走不到底。
//
// ⚠ **站與站之間用 `Teleport` 直接移過去**：四個角在地圖的四個角落，
// 中間怎麼走不是這一條要驗的東西。每一站的互動本身全部走真的按鍵。
func TestCochiseEndgame(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(selfDestructMap, 1, 1); err != nil {
		t.Fatalf("換到地圖 %d 失敗：%v", selfDestructMap, err)
	}
	w := s.World()
	hero := w.Party.Members[0]
	if hero == nil {
		t.Fatal("隊伍第一個人是 nil")
	}
	// 只留四把鑰匙，USE 的清單才數得準（選項用數字鍵）。
	for i := range hero.Items {
		hero.Items[i] = game.Slot{}
	}
	keys := []byte{itemBlackstarKey, itemNovaKey, itemPulsarKey, itemQuasarKey}
	for i, id := range keys {
		hero.Items[i] = game.Slot{ID: id}
	}

	// ── 四根圓柱：站上去，USE 對應的鑰匙 ──
	for i, id := range keys {
		x, y := cochiseCorners[i][0], cochiseCorners[i][1]
		before, _, _, _ := w.Block.At(x, y)
		if before != 2 {
			t.Fatalf("第 %d 根圓柱 (%d,%d) 是 nibble %d，預期 2（條件閘）", i+1, x, y, before)
		}
		w.Teleport(uint8(x), uint8(y))
		useItemOn(t, s, id)
		if got := s.Message(); !strings.Contains(got, "works") {
			t.Fatalf("第 %d 根圓柱插鑰匙 %d：訊息 = %q，預期 It works!", i+1, id, got)
		}
		// 收尾改寫把這一格換成下一步；再踩一次才會跑它。
		revisit(t, s, x, y)
	}
	// 第四站跑完會再接一段（`第 4 站已啟動。` → nibble 12 記錄 6）。
	revisit(t, s, cochiseCorners[3][0], cochiseCorners[3][1])

	// ── 四個角現在是按鈕 ──
	for i, c := range cochiseCorners {
		terrain, _, _, _ := w.Block.At(c[0], c[1])
		if terrain != 8 {
			t.Fatalf("第 %d 個角 (%d,%d) 是 nibble %d，預期 8（按鈕）", i+1, c[0], c[1], terrain)
		}
	}

	// 按下按鈕：單鍵題，答案是 R。
	answerAt(t, s, cochiseCorners[0][0], cochiseCorners[0][1], 'R')
	revisit(t, s, cochiseCorners[0][0], cochiseCorners[0][1]) // 跑 nibble 12 記錄 5

	// ── 紅、黃、綠、藍四站，順序照資料 ──
	sequence := []struct {
		corner int
		key    byte
	}{
		{0, 'R'}, {2, 'Y'}, {3, 'G'}, {1, 'B'},
	}
	for i, stage := range sequence {
		x, y := cochiseCorners[stage.corner][0], cochiseCorners[stage.corner][1]
		if terrain, _, _, _ := w.Block.At(x, y); terrain != 8 {
			t.Fatalf("第 %d 階段的面板 (%d,%d) 是 nibble %d，預期 8", i+1, x, y, terrain)
		}
		answerAt(t, s, x, y, stage.key)
		revisit(t, s, x, y) // 印「階段 N 站已啟動」並改寫成下一步
		revisit(t, s, x, y) // 跑那一步（nibble 12）
	}

	// 最後一站跑完，那一格變成啟動自毀的腳本。
	last := cochiseCorners[sequence[len(sequence)-1].corner]
	if terrain, rec, _, _ := w.Block.At(last[0], last[1]); terrain != 6 || rec != selfDestructRecord {
		t.Fatalf("最後一站現在是 nibble %d 記錄 %d，預期 6／%d",
			terrain, rec, selfDestructRecord)
	}
	revisit(t, s, last[0], last[1]) // 踩上去 → opcode 35
	if !w.SelfDestruct.Armed {
		t.Fatal("走完整段啟動序列，自毀倒數卻沒有開始")
	}

	// ── 倒數 240 刻 ──
	armedAt := w.Clock.Total
	for i := 0; i < game.SelfDestructTicks*3 && !s.ending.active; i++ {
		d := input.DirRight
		if i%2 == 1 {
			d = input.DirLeft
		}
		step(t, s, input.Input{Dir: d})
	}
	if !s.ending.active {
		t.Fatalf("倒數走了 %d 刻，結局還沒開始", w.Clock.Total-armedAt)
	}
	t.Logf("啟動自毀之後第 %d 刻進結局（原版是 %d 刻）",
		w.Clock.Total-armedAt, game.SelfDestructTicks)
}

// useItemOn 走完 `USE` 的三層選單，對腳下那一格用指定的物品。
func useItemOn(t *testing.T, s *Scene, id byte) {
	t.Helper()
	step(t, s, input.Input{Dir: input.DirNone, Char: 'U'})
	if s.use.stage == useStageMember {
		step(t, s, input.Input{Dir: input.DirNone, Char: '1'})
	}
	if s.use.stage != useStageKind {
		t.Fatalf("USE 沒走到選種類那一層，現在是 %d", int(s.use.stage))
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'I'})
	idx := -1
	for i, o := range s.use.options {
		if o.id == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("物品清單裡沒有編號 %d", id)
	}
	if idx > 8 {
		t.Fatalf("物品 %d 排在第 %d 項，超出數字鍵能選的範圍", id, idx+1)
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: byte('1' + idx)})
}

// answerAt 走到 (x, y) 打開那一題，然後按一個鍵回答。
func answerAt(t *testing.T, s *Scene, x, y int, key byte) {
	t.Helper()
	revisit(t, s, x, y)
	if !s.question.active {
		t.Fatalf("走上 (%d,%d) 沒有開問答，現在是 %s", x, y, s.Mode())
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: key})
	if s.question.active {
		t.Fatalf("(%d,%d) 回答 %q 之後問答還開著", x, y, key)
	}
}

// revisit 從一個走得過去的鄰格踩上 (x, y)。
//
// 地圖格就是這台直譯器的程式計數器：每踩一次跑一格指令，
// 所以同一格常常要來回踩好幾次才走得完一段（`docs/re/71` §5.1）。
func revisit(t *testing.T, s *Scene, x, y int) {
	t.Helper()
	w := s.World()
	dirs := []struct {
		dx, dy int
		dir    input.Direction
	}{
		{0, -1, input.DirDown}, {0, 1, input.DirUp},
		{-1, 0, input.DirRight}, {1, 0, input.DirLeft},
	}
	for _, d := range dirs {
		nx, ny := x+d.dx, y+d.dy
		if nx < 0 || ny < 0 || nx >= w.Block.Dim || ny >= w.Block.Dim {
			continue
		}
		if !w.Passable(nx, ny) {
			continue
		}
		w.Teleport(uint8(nx), uint8(ny))
		step(t, s, input.Input{Dir: d.dir})
		if int(w.Party.X) == x && int(w.Party.Y) == y {
			return
		}
		// 被擋住（條件閘）也算踩到了——訊息已經印出來，格子也改寫過。
		return
	}
	t.Fatal(fmt.Sprintf("(%d,%d) 四周沒有走得過去的格子", x, y))
}
