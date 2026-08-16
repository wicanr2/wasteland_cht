package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 結局播得完：畫面 ＋ 動畫 ＋ 四段敘述（`docs/re/96`）。
func TestEndingPlaysThrough(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	s.BeginEnding()
	if !s.Ending() {
		t.Fatal("BeginEnding 之後不在結局")
	}
	if s.ending.pic == nil {
		t.Fatal("結局畫面沒載到")
	}
	if s.ending.anim == nil || len(s.ending.anim.Frames) == 0 {
		t.Fatal("結局動畫沒載到")
	}
	if s.Frame() == nil {
		t.Fatal("結局畫面合成不出一幀")
	}

	// 進場那兩段按鍵無效（原版 ds:D168h ＝ 0）。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: ' '}); err != nil {
		t.Fatalf("進場按鍵：%v", err)
	}
	if s.ending.page != 0 {
		t.Fatalf("進場時按鍵就跳頁了（page ＝ %d）", s.ending.page)
	}

	// 逐 tick 跑到結束。上界比原版整段長得多，跑不完就是有迴圈沒收。
	seen := []string{}
	for i := 0; i < 20000 && s.Ending(); i++ {
		before := s.ending.page
		s.TickEnding()
		if s.ending.page != before && s.Message() != "" {
			seen = append(seen, s.Message())
		}
	}
	if s.Ending() {
		t.Fatal("跑了 20000 tick 還沒播完")
	}
	if len(seen) != len(EndingPages) {
		t.Fatalf("印了 %d 段，應該是 %d 段：%q", len(seen), len(EndingPages), seen)
	}
	// 第一段是爆炸，最後一段提到機器人——這四條是 ds:D18Eh 表的 1–4。
	if !strings.Contains(seen[0], "explosions") {
		t.Errorf("第一段不像結局敘述：%q", seen[0])
	}
	if !strings.Contains(seen[3], "robots") {
		t.Errorf("第四段不像結局敘述：%q", seen[3])
	}
	t.Logf("四段：%q", seen)
}

// 清算：只有在 Base Cochise 裡（地圖 0x10–0x14）的隊伍會死（`docs/re/96` §2.1）。
//
// ⚠ 這一條同時擋住「全隊一律陣亡」那種簡化——原版是**依地圖判定**的，
// 留在外面的巡守員活著回去領表揚，而那正是 `+0x4B` 那個 bit 存在的理由。
func TestEndingTollOnlyInsideBaseCochise(t *testing.T) {
	for _, tc := range []struct {
		mapID int
		dies  bool
	}{
		{BaseCochiseFirst - 1, false},
		{BaseCochiseFirst, true},
		{BaseCochiseEnd - 1, true},
		{BaseCochiseEnd, false},
	} {
		rom := openRom(t)
		s, err := New(rom)
		if err != nil {
			t.Fatalf("開場失敗：%v", err)
		}
		groups := s.save.SlotGroups()
		slot := s.save.Plain[groups[0].RawIndex : groups[0].RawIndex+14]
		slot[10] = byte(tc.mapID)

		s.BeginEnding()
		s.collectToll()
		got := len(s.Killed()) > 0
		if got != tc.dies {
			t.Errorf("地圖 %#x：死了 %v，應該是 %v（名單 %q）", tc.mapID, got, tc.dies, s.Killed())
		}
		if tc.dies {
			if !strings.Contains(s.Killed()[0], "killed in the blast") {
				t.Errorf("地圖 %#x 的死訊不對：%q", tc.mapID, s.Killed()[0])
			}
			if strings.Contains(s.Killed()[0], "\v") {
				t.Errorf("名字佔位符 \\v 沒有換掉：%q", s.Killed()[0])
			}
		}
	}
}

// 結局逐人設「參與過摧毀 Base Cochise」，Radio 靠它發表揚（`docs/re/96` §5）。
func TestEndingSetsMissionFlagAndRadioPraises(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	m := s.World().Party.Members[0]
	if m.Mission {
		t.Fatal("一開始就有 Mission 旗標")
	}
	// 還沒去過結局：Radio 不該有賀詞。
	//
	// 這裡直接叫 `doRadio`（跳過 Y／N 確認）：`BeginEnding` 之後場景不在地圖上，
	// 按鍵那條路進不來。**確認流程本身由 `confirm_test.go` 守著**，
	// 這一份只管賀詞的旗標。
	if _, err := s.doRadio(); err != nil {
		t.Fatalf("doRadio：%v", err)
	}
	if strings.Contains(s.Message(), "Congratulations") {
		t.Fatalf("沒去過結局就被表揚了：%q", s.Message())
	}

	s.BeginEnding()
	if !m.Mission {
		t.Fatal("結局沒有設 Mission 旗標")
	}
	if _, err := s.doRadio(); err != nil {
		t.Fatalf("doRadio：%v", err)
	}
	if !strings.Contains(s.Message(), "Congratulations") {
		t.Fatalf("Radio 沒念賀詞：%q", s.Message())
	}
	if !m.Praised {
		t.Fatal("表揚完沒有設 Praised 旗標")
	}
	// 第二次不該再念（原版 +0x4C 就是為了這個）。
	if _, err := s.doRadio(); err != nil {
		t.Fatalf("doRadio 第二次：%v", err)
	}
	if strings.Contains(s.Message(), "Congratulations") {
		t.Fatalf("賀詞念了第二次：%q", s.Message())
	}
}
