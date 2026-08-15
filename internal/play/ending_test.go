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
