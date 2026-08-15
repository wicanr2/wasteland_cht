package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestSpawnEncounterPlacesEnemyCell 驗遭遇生成器（docs/re/78）。
//
// 出貨地圖上一格 nibble 15 都沒有——敵人格是每走一步擲 1／標頭 `+0x2F`
// 生出來的。沒有這一步，`ScanEncounters` 永遠掃不到東西，
// 隨機戰鬥整個不會發生（`docs/re/77` 量到 105 步 0 次）。
func TestSpawnEncounterPlacesEnemyCell(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	tbl, ok := s.spawnTables()
	if !ok {
		t.Fatal("讀不到遭遇生成的三張表")
	}
	if tbl.Dist[1] == 0 {
		t.Fatalf("距離表看起來不對：%v", tbl.Dist)
	}
	w := s.World()
	// 出生點四周一格空地都沒有，生成器永遠放不下——換到空曠處。
	w.Party.X, w.Party.Y = 12, 2

	placed := 0
	for i := 0; i < 3000 && placed == 0; i++ {
		if r := w.SpawnEncounter(tbl); r.Placed {
			placed++
			if terrain, idx, _, err := w.Block.At(r.X, r.Y); err != nil ||
				terrain != 15 || int(idx) != r.Slot {
				t.Fatalf("生成後 (%d,%d) 是 nibble %d 記錄 %d，預期 15／%d",
					r.X, r.Y, terrain, idx, r.Slot)
			}
			rec, err := w.Block.SectionRecord(15, r.Slot)
			if err != nil {
				t.Fatal(err)
			}
			if rec[3] != r.Kind || int(rec[4]) != r.Count {
				t.Errorf("槽 %d 的記錄是 %02x %02x，預期種類 %d 數量 %d",
					r.Slot, rec[3], rec[4], r.Kind, r.Count)
			}
		}
	}
	if placed == 0 {
		t.Fatal("3000 次嘗試一次都沒生成——生成器沒接上")
	}
}

// TestRandomEncounterActuallyStarts 是端到端驗收：走在空曠處會打起來。
func TestRandomEncounterActuallyStarts(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(0, 12, 2); err != nil {
		t.Fatalf("載入地圖 0 失敗：%v", err)
	}
	dir := input.DirRight
	for i := 0; i < 400; i++ {
		if _, err := s.Update(input.Input{Dir: dir}); err != nil {
			t.Fatalf("第 %d 步：%v", i, err)
		}
		if s.InCombat() {
			t.Logf("第 %d 步打起來了", i+1)
			return
		}
		if dir == input.DirRight {
			dir = input.DirLeft
		} else {
			dir = input.DirRight
		}
	}
	t.Error("走了 400 步一次遭遇都沒有")
}
