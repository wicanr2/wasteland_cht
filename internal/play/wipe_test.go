package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// 三分支的判準（`docs/spec/28` §2）。
//
// ⚠ 差別是「**救不救得回來**」不是「倒了幾個」：CON ≤ 0 但傷勢等級 0、
// 狀態位元也是 0 的人是會自己醒的昏迷，那一組還有救。
func TestWipeClassification(t *testing.T) {
	mk := func(con int16, status byte) *game.Character {
		return &game.Character{CON: con, Status: status}
	}
	for _, tc := range []struct {
		name string
		p    *game.Party
		want game.WipeState
	}{
		{"有人站著", &game.Party{Members: []*game.Character{mk(10, 0), mk(-30, 0)}}, game.WipeNone},
		{"全倒但都是昏迷", &game.Party{Members: []*game.Character{mk(-3, 0), mk(-5, 0)}}, game.WipeSwitch},
		// ⚠ CON 剛好 0 是傷勢等級 5（`sub_19A1D` 的第一條），那是死不是昏迷。
		{"CON 剛好 0 不是昏迷", &game.Party{Members: []*game.Character{mk(0, 0)}}, game.WipeDead},
		{"全倒且都有傷勢", &game.Party{Members: []*game.Character{mk(-30, 0), mk(-40, 0)}}, game.WipeDead},
		{"全倒且都有狀態", &game.Party{Members: []*game.Character{mk(-5, 0x01), mk(-5, 0x02)}}, game.WipeDead},
		{"混合：一個救得回來", &game.Party{Members: []*game.Character{mk(-30, 0), mk(-5, 0)}}, game.WipeSwitch},
		{"沒有人", &game.Party{}, game.WipeDead},
	} {
		if got := tc.p.Wipe(); got != tc.want {
			t.Errorf("%s：得到 %v，預期 %v", tc.name, got, tc.want)
		}
	}
}

// TestWipeShowsDeathScreen：整組打到底之後走一步 → 死亡畫面。
//
// 從**玩家的按鍵**進場（走一步觸發主迴圈那道檢查），不直接叫 beginWipe。
func TestWipeShowsDeathScreen(t *testing.T) {
	s := sceneWithCatalogue(t)
	for _, c := range s.World().Party.Members {
		c.CON = -40 // 每個人都倒下而且有傷勢等級
	}
	step(t, s, input.Input{Dir: input.DirUp})

	if s.Mode() != "wipe" {
		t.Fatalf("整組打到底之後停在 %s，預期 wipe", s.Mode())
	}
	// 兩段文字是**從映像讀出來的原版字**，不是我們編的。
	if s.WipePlaceLine() != "Grim Reaper" {
		t.Errorf("地點名是 %q，預期 Grim Reaper", s.WipePlaceLine())
	}
	if len(s.CJK()) == 0 {
		t.Errorf("死亡畫面那一句沒有中文（英文是 %q）", s.Message())
	}
	// 畫面上要有圖：ALLPICS 第 0x3B 張畫在設施圖的位置。
	f := s.Frame()
	on := 0
	for y := render.FacilityPicY; y < render.FacilityPicY+84; y++ {
		for x := render.FacilityPicX; x < render.FacilityPicX+96; x++ {
			if f.At(x, y) != 0 {
				on++
			}
		}
	}
	if on == 0 {
		t.Error("死亡畫面沒有畫圖")
	}

	// 任何鍵回標題，而且不會再掉回死亡畫面。
	step(t, s, input.Input{Dir: input.DirNone, Char: ' '})
	if s.Mode() != "title" {
		t.Errorf("按鍵之後停在 %s，預期回標題", s.Mode())
	}
}

// TestWipeSwitchesPartyWhenSalvageable：全倒但救得回來 → 自動換隊。
func TestWipeSwitchesPartyWhenSalvageable(t *testing.T) {
	s := newScene(t)
	n := len(s.World().Party.Members)
	if n < 2 {
		t.Skip("隊伍不到兩個人，分不出第二組")
	}
	// 先用 `D` 分最後一個人出去自成一組，這樣才有「下一支隊伍」可切。
	step(t, s, input.Input{Dir: input.DirNone, Char: 'D'})
	step(t, s, input.Input{Dir: input.DirNone, Char: byte('0' + n)})
	if s.groupCount() != 2 {
		t.Skipf("沒有分出第二組（共 %d 組）", s.groupCount())
	}
	before := s.groupID
	for _, c := range s.World().Party.Members {
		// ⚠ **不能用 CON = 0**：原版的 `sub_19A1D` 第一條就是
		// 「CON 剛好 0 → 傷勢等級 5」，那是死不是昏迷。
		// 要「倒下但救得回來」得落在 0 與第一個門檻 −11 之間。
		c.CON = -5
		c.Status = 0
	}
	step(t, s, input.Input{Dir: input.DirUp})

	if s.Mode() == "wipe" {
		t.Fatal("還救得回來卻走了死亡畫面")
	}
	if s.groupID == before {
		t.Errorf("全倒但救得回來時應該自動換隊，還停在第 %d 組", before)
	}
}
