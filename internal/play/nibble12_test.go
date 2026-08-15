package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestNibble12BatchPatch 驗 nibble 12 的批次改寫（docs/re/71）。
//
// 地圖 3 的 (28,16) 是 nibble 12 記錄 14，它的批次表第一筆要把 (28,17)
// 從 nibble 1 記錄 0 改成 nibble 3 記錄 16——**改的是別的格子，不是腳下**。
func TestNibble12BatchPatch(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(3, 27, 16); err != nil {
		t.Fatalf("載入地圖 3 失敗：%v", err)
	}
	w := s.World()
	if terrain, idx, _, _ := w.Block.At(28, 17); terrain != 1 || idx != 0 {
		t.Fatalf("走進去之前 (28,17) 是 nibble %d 記錄 %d，預期 1／0", terrain, idx)
	}

	if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
		t.Fatalf("走進 (28,16) 失敗：%v", err)
	}
	if x, y := w.Party.X, w.Party.Y; x != 28 || y != 16 {
		t.Fatalf("走完在 (%d,%d)，預期 (28,16)", x, y)
	}
	terrain, idx, _, _ := w.Block.At(28, 17)
	if terrain != 3 || idx != 16 {
		t.Errorf("批次改寫後 (28,17) 是 nibble %d 記錄 %d，預期 3／16", terrain, idx)
	}
}
