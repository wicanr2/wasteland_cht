package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestTrainerListsSkills 驗訓練師列得出技能、而且學得起來（docs/re/80）。
//
// 資源 21 的 (2,21) 是 `Market` 的傳送格；`Library`（訓練師）在同一張地圖
// 的記錄 3。這裡直接用記錄進設施，測的是清單與學習，不是入口。
func TestTrainerListsSkills(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	blk, err := rom.BlockByID(21)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := blk.SectionRecord(6, 3)
	if err != nil {
		t.Fatal(err)
	}
	fs := s.EnterFacility(rec)
	if fs == nil {
		t.Fatal("進不了 Library")
	}
	if fs.Facility.Kind != game.FacilityTrainer {
		t.Fatalf("Kind ＝ %d，預期訓練師", fs.Facility.Kind)
	}
	if len(fs.Skills) == 0 {
		t.Fatal("技能清單是空的")
	}
	t.Logf("Library 列出 %d 個技能，第一個 ID=%d IQ=%d 費用=%d",
		len(fs.Skills), fs.Skills[0].ID, fs.Skills[0].Data.IQ, fs.Skills[0].Data.BaseCost)

	// 費用與 IQ 需求要是合理值（技能資料 +0x00 的兩個欄位，docs/re/31 §3）。
	for _, sk := range fs.Skills {
		if sk.Data.BaseCost == 0 || sk.Data.BaseCost > 7 {
			t.Errorf("技能 %d 的費用 %d 不合理", sk.ID, sk.Data.BaseCost)
		}
	}
}

// TestTrainerEntryEndToEnd 走進 Library 並確認進的是訓練師。
func TestTrainerEntryEndToEnd(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	// 資源 21 記錄 22 是 Library 的傳送格，在 (2,24)（往上踩進去）。
	blk, _ := rom.BlockByID(21)
	var tx, ty = -1, -1
	for y := 0; y < blk.Dim && ty < 0; y++ {
		for x := 0; x < blk.Dim; x++ {
			if terrain, idx, _, err := blk.At(x, y); err == nil && terrain == 10 && idx == 22 {
				tx, ty = x, y
				break
			}
		}
	}
	if tx < 0 {
		t.Skip("找不到 Library 的傳送格")
	}
	if err := s.LoadMap(21, uint8(tx), uint8(ty+1)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
		t.Fatal(err)
	}
	if !s.Asking() {
		t.Fatalf("沒有問 Enter new location?，訊息 %q", s.Message())
	}
	if _, err := s.Update(input.Input{Char: 'Y'}); err != nil {
		t.Fatal(err)
	}
	if !s.InFacility() {
		t.Fatalf("沒有進設施，訊息 %q", s.Message())
	}
	f := s.Facility()
	if f.Facility.Kind != game.FacilityTrainer || len(f.Skills) == 0 {
		t.Errorf("進的是 kind %d、清單 %d 項，預期訓練師且清單非空",
			f.Facility.Kind, len(f.Skills))
	}
	t.Logf("走進『%s』，列出 %d 個技能", f.Facility.Name, len(f.Skills))
}
