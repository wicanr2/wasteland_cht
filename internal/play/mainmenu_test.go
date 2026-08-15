package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 主選單只有一個選項，按 `S` 進遊戲（`docs/re/95`）。
//
// ⚠ **其餘的鍵什麼都不做**——原版索引 1 的處理程式是一支 `retn`，
// 包括 ESC 在內。這一條同時擋住「隨便一個鍵就開始」那種便宜實作。
func TestTitleStartsOnlyOnStartKey(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	s.BeginTitle()
	if !s.showTitle() {
		t.Fatal("BeginTitle 之後不在標題畫面")
	}
	if s.titlePic == nil {
		t.Fatal("標題圖沒載到")
	}
	// 畫得出來（標題模式不畫地圖，也不畫指令列）。
	if s.Frame() == nil {
		t.Fatal("標題畫面合成不出一幀")
	}

	for _, ch := range []byte{'X', 'Q', '1'} {
		if _, err := s.Update(input.Input{Dir: input.DirNone, Char: ch}); err != nil {
			t.Fatalf("按 %c：%v", ch, err)
		}
		if !s.showTitle() {
			t.Fatalf("按 %c 就離開標題畫面了", ch)
		}
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel}); err != nil {
		t.Fatalf("按 ESC：%v", err)
	}
	if !s.showTitle() {
		t.Fatal("ESC 離開了標題畫面——原版那一支只是 retn")
	}

	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 's'}); err != nil {
		t.Fatalf("按 S：%v", err)
	}
	if s.showTitle() {
		t.Fatal("按 S 之後還停在標題畫面")
	}
	// 進遊戲之後畫的是地圖與指令列。
	if s.titlePic != nil {
		t.Error("進遊戲之後標題圖還留著")
	}
}
