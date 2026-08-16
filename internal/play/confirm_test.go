package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// confirmed 下一個要確認的指令並答 Y。
//
// `Save` 與 `Radio` 現在都會先問一句，所以**測那兩個指令的效果時要走完整條路**——
// 直接呼叫 `doSave`／`doRadio` 測得到效果，卻測不到「按鍵真的接到那個效果」。
func confirmed(t *testing.T, s *Scene, key byte) {
	t.Helper()
	step(t, s, input.Input{Dir: input.DirNone, Char: key})
	if s.Mode() != "confirm" {
		t.Fatalf("按 %c 之後沒有停在確認，停在 %s", key, s.Mode())
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
}

// `Save` 與 `Radio` 都要先問一句（`sub_19B4F`，`docs/re/91` §2）。
//
// ⚠ **兩個要一起驗**：只接一邊的話另一邊仍然是「按下去就發生」，
// 而那正是玩家最容易誤觸的兩個。
func TestSaveAndRadioAskFirst(t *testing.T) {
	for _, tc := range []struct{ name string; key byte }{
		{"Save", 'S'}, {"Radio", 'R'},
	} {
		s := newScene(t)
		step(t, s, input.Input{Dir: input.DirNone, Char: tc.key})
		if s.Mode() != "confirm" {
			t.Errorf("%s 沒有先問就做了（停在 %s）", tc.name, s.Mode())
			continue
		}
		if s.Message() == "" && len(s.CJK()) == 0 {
			t.Errorf("%s 的確認沒有印任何話", tc.name)
		}
		// N 取消，回到地圖。
		step(t, s, input.Input{Dir: input.DirNone, Char: 'N'})
		if s.Mode() != "map" {
			t.Errorf("%s 答 N 之後停在 %s，預期回地圖", tc.name, s.Mode())
		}
	}
}

// 答 N 就真的不做：`Radio` 的表揚旗標不可以被用掉。
func TestRadioCancelDoesNotConsumePraise(t *testing.T) {
	s := newScene(t)
	for _, m := range s.World().Party.Members {
		if m != nil {
			m.Mission, m.Praised = true, false
		}
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'R'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'N'})
	for i, m := range s.World().Party.Members {
		if m != nil && m.Praised {
			t.Fatalf("答 N 卻用掉了第 %d 個人的表揚", i)
		}
	}
	// 答 Y 才會用掉。
	step(t, s, input.Input{Dir: input.DirNone, Char: 'R'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	used := false
	for _, m := range s.World().Party.Members {
		if m != nil && m.Praised {
			used = true
		}
	}
	if !used {
		t.Error("答 Y 卻沒有進到升級流程")
	}
}

// ESC 在確認畫面上是取消，**不會結束遊戲**（規格 27 §1）。
func TestConfirmEscapeCancels(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'S'})
	ok, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	if err != nil || !ok {
		t.Fatalf("確認畫面的 ESC 結束了遊戲：ok=%v err=%v", ok, err)
	}
	if s.Mode() != "map" {
		t.Errorf("ESC 之後停在 %s，預期回地圖", s.Mode())
	}
}

// 確認畫面吃掉方向鍵：等 Y／N 的時候不該走路。
func TestConfirmSwallowsMovement(t *testing.T) {
	s := newScene(t)
	w := s.World()
	step(t, s, input.Input{Dir: input.DirNone, Char: 'S'})
	x, y := w.Party.X, w.Party.Y
	step(t, s, input.Input{Dir: input.DirUp})
	if w.Party.X != x || w.Party.Y != y {
		t.Errorf("確認畫面上按方向鍵走了路：(%d,%d) → (%d,%d)", x, y, w.Party.X, w.Party.Y)
	}
}
