package play

import (
	"bytes"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 場景模式與按鍵路由（docs/spec/24 §6 的驗收條件）。

// 驗收 1：戰鬥中送方向鍵，隊伍不得移動。
func TestCombatSwallowsDirectionKeys(t *testing.T) {
	s := openScene(t)
	w := s.World()
	s.combat = NewCombatScene(game.NewBattle(w.Party, w.RNG))
	x, y := w.Party.X, w.Party.Y

	for _, d := range []input.Direction{input.DirUp, input.DirDown, input.DirLeft, input.DirRight} {
		if _, err := s.Update(input.Input{Dir: d}); err != nil {
			t.Fatalf("Update：%v", err)
		}
	}
	if w.Party.X != x || w.Party.Y != y {
		t.Errorf("戰鬥中按方向鍵不該走路：(%d,%d) → (%d,%d)", x, y, w.Party.X, w.Party.Y)
	}
}

// 驗收 2：設施中送方向鍵不動，ESC 之後回地圖。
func TestFacilitySwallowsDirectionKeys(t *testing.T) {
	s := openScene(t)
	w := s.World()
	rec := make([]byte, 32)
	rec[0] = 0x80 | byte(game.FacilityShop)
	copy(rec[0x07:], "Shop\x00")
	s.EnterFacility(rec)
	x, y := w.Party.X, w.Party.Y

	if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
		t.Fatal(err)
	}
	if w.Party.X != x || w.Party.Y != y {
		t.Error("設施中按方向鍵不該走路")
	}
	if !s.InFacility() {
		t.Fatal("方向鍵不該把人踢出設施")
	}
	if _, err := s.Update(input.Input{Action: input.ActionCancel}); err != nil {
		t.Fatal(err)
	}
	if s.InFacility() {
		t.Error("ESC 之後應該回地圖")
	}
}

// 驗收 3：F10 任何模式都能離開。
func TestQuitWorksInEveryMode(t *testing.T) {
	s := openScene(t)
	for _, name := range []string{"地圖", "戰鬥", "設施"} {
		s.combat, s.facility = nil, nil
		switch name {
		case "戰鬥":
			s.combat = NewCombatScene(game.NewBattle(s.World().Party, s.World().RNG))
		case "設施":
			rec := make([]byte, 32)
			rec[0] = 0x80 | byte(game.FacilityShop)
			s.EnterFacility(rec)
		}
		keep, err := s.Update(input.Input{Action: input.ActionQuit})
		if err != nil {
			t.Fatalf("%s：%v", name, err)
		}
		if keep {
			t.Errorf("%s模式下 F10 應該離開", name)
		}
	}
}

// 驗收 4：三種模式各畫一幀，而且**畫出來不一樣**。
//
// 只驗「不回錯誤」是不夠的——三種都畫成同一張地圖也會通過。
func TestEveryModeDrawsSomethingDifferent(t *testing.T) {
	s := openScene(t)
	s.combat, s.facility = nil, nil
	s.dirty = true
	mapPix := append([]byte(nil), s.Frame().RGBA()...)

	s.combat = NewCombatScene(game.NewBattle(s.World().Party, s.World().RNG))
	s.dirty = true
	combatPix := append([]byte(nil), s.Frame().RGBA()...)

	s.combat = nil
	rec := make([]byte, 32)
	rec[0] = 0x80 | byte(game.FacilityShop)
	copy(rec[0x07:], "Shop\x00")
	s.EnterFacility(rec)
	s.dirty = true
	facPix := append([]byte(nil), s.Frame().RGBA()...)

	if bytes.Equal(mapPix, combatPix) {
		t.Error("戰鬥畫面與地圖畫面一模一樣")
	}
	if bytes.Equal(mapPix, facPix) {
		t.Error("設施畫面與地圖畫面一模一樣")
	}
	if bytes.Equal(combatPix, facPix) {
		t.Error("設施畫面與戰鬥畫面一模一樣")
	}
}
