package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

func combatScene(t *testing.T) *Scene {
	t.Helper()
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	return s
}

// 戰鬥裡按 `U` 要開 S／I／A 那一層——**不用挑人**（輪到誰就是誰，`docs/re/108` §1）。
func TestCombatUseSkipsMemberPick(t *testing.T) {
	s := combatScene(t)
	turn := s.combat.Turn
	step(t, s, input.Input{Dir: input.DirNone, Char: 'U'})
	if s.use.stage != useStageKind {
		t.Fatalf("按 U 應該直接到 S／I／A 那一層，stage ＝ %d", s.use.stage)
	}
	if s.use.member != turn {
		t.Errorf("挑到的人是 %d，應該是輪到的那個 %d", s.use.member, turn)
	}
	if !s.use.combat {
		t.Error("沒有標成戰鬥那條路——選完會當場施用而不是記成指令")
	}
}

// 選完之後**要再問一個方向**，而且不當場施用。
func TestCombatUseAsksForDirection(t *testing.T) {
	s := combatScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'U'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'S'})
	if s.use.stage != useStagePick {
		t.Skip("這個人沒有技能可用")
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: '1'})
	if s.use.stage != useStageDir {
		t.Fatalf("選完技能應該問方向，stage ＝ %d", s.use.stage)
	}
	// 這時候還沒下令。
	if s.combat.Phase.Cmd[0] == game.CmdUse {
		t.Error("還沒選方向就記成指令了")
	}
}

// 方向選完 → 參數 ＝ `(選項 << 4) | 方向`，編號另存一格（`docs/re/108` §1）。
func TestCombatUsePacksTheParameter(t *testing.T) {
	s := combatScene(t)
	turn := s.combat.Turn
	step(t, s, input.Input{Dir: input.DirNone, Char: 'U'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'S'})
	if s.use.stage != useStagePick {
		t.Skip("這個人沒有技能可用")
	}
	wantID := s.use.options[0].id
	step(t, s, input.Input{Dir: input.DirNone, Char: '1'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'I'}) // 上

	c := s.combat
	if c.Phase.Cmd[turn] != game.CmdUse {
		t.Fatalf("指令碼應該是 USE，得到 %d", c.Phase.Cmd[turn])
	}
	kind, id, dir := c.UseParts(turn)
	if kind != game.UseSkill {
		t.Errorf("選項 ＝ %d，預期 UseSkill", kind)
	}
	if id != wantID {
		t.Errorf("編號 ＝ %d，預期 %d", id, wantID)
	}
	if dir != useDirUp {
		t.Errorf("方向 ＝ %d，預期上（%d）", dir, useDirUp)
	}
	// ⚠ 參數只有一個 byte，兩個 nibble 都要塞得下。
	if got := c.Phase.Arg[turn]; got != byte(game.UseSkill)<<4|useDirUp {
		t.Errorf("參數 ＝ %#x，預期 %#x", got, byte(game.UseSkill)<<4|useDirUp)
	}
}

// 方向鍵與 `IJKL` 是同一件事（原版 `0x1260E` 的 `al −= 5`，`docs/re/108` §2）。
func TestUseDirectionFoldsArrowKeys(t *testing.T) {
	for _, tc := range []struct {
		ch   byte
		dir  input.Direction
		want byte
	}{
		{'I', input.DirNone, useDirUp},
		{'K', input.DirNone, useDirDown},
		{'J', input.DirNone, useDirLeft},
		{'L', input.DirNone, useDirRight},
		{' ', input.DirNone, useDirStay},
		{0, input.DirUp, useDirUp},
		{0, input.DirDown, useDirDown},
		{0, input.DirLeft, useDirLeft},
		{0, input.DirRight, useDirRight},
	} {
		got, ok := useDirection(tc.ch, tc.dir)
		if !ok || got != tc.want {
			t.Errorf("鍵 %q／方向 %v → %d（ok=%v），預期 %d", tc.ch, tc.dir, got, ok, tc.want)
		}
	}
	if _, ok := useDirection('Z', input.DirNone); ok {
		t.Error("不相干的鍵不該被當成方向")
	}
}

// 五個方向的位移照 `ds:AAB1h` 的前五格。
func TestUseDirDelta(t *testing.T) {
	for _, tc := range []struct{ d, dx, dy int }{
		{int(useDirUp), 0, -1}, {int(useDirDown), 0, 1},
		{int(useDirLeft), -1, 0}, {int(useDirRight), 1, 0},
		{int(useDirStay), 0, 0},
	} {
		dx, dy := useDirDelta(byte(tc.d))
		if dx != tc.dx || dy != tc.dy {
			t.Errorf("方向 %d → (%d,%d)，預期 (%d,%d)", tc.d, dx, dy, tc.dx, tc.dy)
		}
	}
}

// ESC 收掉選單、回到指令選單，而且**這個人要重問**。
func TestCombatUseCancel(t *testing.T) {
	s := combatScene(t)
	turn := s.combat.Turn
	step(t, s, input.Input{Dir: input.DirNone, Char: 'U'})
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	if s.use.stage != useStageOff {
		t.Error("ESC 沒有收掉選單")
	}
	if s.combat.Turn != turn {
		t.Error("取消之後應該重問同一個人")
	}
}

// 結算打的是**往那個方向一格**，不是隊伍腳下那一格（`docs/re/108` §2）。
//
// ⚠ 拿腳下那一格會安靜地做錯事：多數格子的記錄都存在，所以看起來一樣會動。
func TestCombatUseTargetsTheCellInThatDirection(t *testing.T) {
	s := combatScene(t)
	c := s.combat
	c.Phase.Cmd[0] = game.CmdUse
	c.Phase.Arg[0] = byte(game.UseSkill)<<4 | useDirRight
	kind, _, dir := c.UseParts(0)
	if kind != game.UseSkill || dir != useDirRight {
		t.Fatalf("拆回來不對：%d／%d", kind, dir)
	}
	dx, dy := useDirDelta(dir)
	if dx != 1 || dy != 0 {
		t.Fatalf("右邊那一格的位移算錯：(%d,%d)", dx, dy)
	}
	// 目標格 ＝ 隊伍座標 ＋ 位移。
	w := c.World
	if w == nil {
		t.Fatal("CombatScene 沒有接上 World——USE 會什麼都不做")
	}
	if int(w.Party.X)+dx == int(w.Party.X) {
		t.Error("目標格與腳下那一格一樣")
	}
}
