package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestRangerCenterEntry 驗「走開再走回來才進得去」的完整路徑（docs/re/72）。
//
// 出廠存檔就站在地圖 0 的 (55,62)＝ nibble 6 記錄 0（設施 Ranger Ctr.），
// 但站著不觸發。走到 (55,61) 的 nibble 12，它的批次表把 (55,62) 改成
// nibble 10 記錄 39（傳送 ＋ 要確認），走回去才會問 Enter new location?。
//
// 實機送同一串鍵（i、k）得到同一個結果——見 docs/re/72 §2 的截圖指令。
func TestRangerCenterEntry(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	w := s.World()
	if x, y := w.Party.X, w.Party.Y; x != 55 || y != 62 {
		t.Fatalf("開場在 (%d,%d)，預期 (55,62)", x, y)
	}
	if terrain, idx, _, _ := w.Block.At(55, 62); terrain != 6 || idx != 0 {
		t.Fatalf("開場時 (55,62) 是 nibble %d 記錄 %d，預期 6／0", terrain, idx)
	}

	if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
		t.Fatalf("走上 (55,61) 失敗：%v", err)
	}
	terrain, idx, _, _ := w.Block.At(55, 62)
	if terrain != 10 || idx != 39 {
		t.Fatalf("批次改寫後 (55,62) 是 nibble %d 記錄 %d，預期 10／39", terrain, idx)
	}

	if _, err := s.Update(input.Input{Dir: input.DirDown}); err != nil {
		t.Fatalf("走回 (55,62) 失敗：%v", err)
	}
	if !s.Asking() {
		t.Error("走回去沒有停下來問，預期 Enter new location?")
	}
	if got := s.Message(); !strings.Contains(got, "Enter new location") {
		t.Errorf("訊息 = %q，預期含 Enter new location", got)
	}
	if x, y := w.Party.X, w.Party.Y; x != 55 || y != 61 {
		t.Errorf("問的時候應該停在 (55,61)，實際 (%d,%d)", x, y)
	}
}
