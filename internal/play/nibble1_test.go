package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestNibble1MessageList 驗 nibble 1 的氛圍敘述（sub_16CD0，docs/re/70）。
//
// 地圖 2 的 (24,9)–(28,9) 連續五格都是 nibble 1 記錄 6，那筆記錄有**兩條**訊息
// （bit7 設的那一條是最後一條）：
//
//	第一格 → 兩條都印
//	後面四格 → 一條都不印（原版比對 ds:4716h／4717h 的 (nibble, 記錄)）
func TestNibble1MessageList(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(2, 23, 9); err != nil {
		t.Fatalf("載入地圖 2 失敗：%v", err)
	}

	if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
		t.Fatalf("走進 (24,9) 失敗：%v", err)
	}
	first := s.Message()
	for _, want := range []string{"beside a low stage", "dancers prance out"} {
		if !strings.Contains(first, want) {
			t.Errorf("第一格的訊息 = %q，預期含 %q", first, want)
		}
	}

	for i := 0; i < 4; i++ {
		if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
			t.Fatalf("第 %d 步失敗：%v", i+2, err)
		}
		if got := s.Message(); got != "" {
			t.Errorf("第 %d 格是同一筆記錄，不該再印，卻印了 %q", i+2, got)
		}
	}
	if x, y := s.World().Party.X, s.World().Party.Y; x != 28 || y != 9 {
		t.Fatalf("走完在 (%d,%d)，預期 (28,9)", x, y)
	}
}
