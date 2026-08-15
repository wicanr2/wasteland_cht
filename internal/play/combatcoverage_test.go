package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// TestEncounterTableCoverage 盤點 42 張地圖的遭遇資料。
//
// `Battle.Spawn` 在 `table.Lookup(sg.Type)` 查不到時**靜靜跳過**那一組——
// 症狀是「打起來少了一批敵人」，不會有任何錯誤訊息。這個測試把它變成數字。
func TestEncounterTableCoverage(t *testing.T) {
	rom := openRom(t)
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	var maps, noTable, records, groups, missing, emptyRec int
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		maps++
		raw, err := blk.EnemyData()
		if err != nil {
			noTable++
			t.Logf("資源 %2d：沒有敵人資料表（%v）", res.ID, err)
			continue
		}
		table := game.ParseEnemyTable(raw)
		n := blk.SectionCount(15)
		for i := 0; i < n; i++ {
			rec, err := blk.SectionRecord(15, i)
			if err != nil || len(rec) < 8 {
				continue
			}
			records++
			any := false
			for _, sg := range game.ReadSpawnGroups(rec) {
				if sg.Count == 0 {
					continue
				}
				groups++
				any = true
				if _, ok := table.Lookup(sg.Type); !ok {
					missing++
					t.Logf("資源 %2d section15 記錄 %2d：型別 %d 查不到（表 %d 筆）",
						res.ID, i, sg.Type, len(table))
				}
			}
			if !any {
				emptyRec++
			}
		}
	}
	t.Logf("%d 張地圖（%d 張沒有敵人表）：%d 筆遭遇記錄、%d 組敵人，查不到型別的 %d 組、整筆空的 %d 筆",
		maps, noTable, records, groups, missing, emptyRec)

	// 敵人型別**一個都不能查不到**——查不到會靜靜少生一組敵人。
	if missing != 0 {
		t.Errorf("有 %d 組敵人的型別在資料表裡查不到", missing)
	}
	// 159 筆遭遇記錄裡 150 筆是空的：section 15 是**遭遇生成器的槽**，
	// 出貨資料只填了少數固定遭遇，其餘由 `sub_16890` 執行期填
	// （docs/re/77）。這個數字是現況記錄，實作生成器之後不會變
	// （生成器改的是區塊的複本）。
	if emptyRec < 100 {
		t.Errorf("空槽只有 %d 筆，遠少於預期——section 15 的解讀可能變了", emptyRec)
	}
}

// TestNoEnemyCellsInShippedMaps 記錄一個**已知缺口**：
// 出貨的 42 張地圖上一格 nibble 15（敵人）都沒有。
//
// 敵人格是 `sub_16890` 每走一步擲骰生出來的（docs/re/77），
// remake 目前只有 `rollEncounter` 擲那一下，沒有生成器——
// 所以隨機遭遇永遠打不起來。實作生成器之後這個測試要跟著改。
func TestNoEnemyCellsInShippedMaps(t *testing.T) {
	rom := openRom(t)
	resources, _ := rom.Resources()
	total := 0
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for y := 0; y < blk.Dim; y++ {
			for x := 0; x < blk.Dim; x++ {
				if terrain, _, _, err := blk.At(x, y); err == nil && terrain == 15 {
					total++
				}
			}
		}
	}
	if total != 0 {
		t.Errorf("出貨地圖上有 %d 格 nibble 15，與「敵人格是生成的」不符", total)
	}
	t.Logf("42 張地圖的 nibble 15 格數 ＝ %d（敵人格全部由 sub_16890 生成）", total)
}
