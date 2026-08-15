package play

import (
	"sort"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// TestFacilityCoverage 盤點 42 張地圖的設施：每一家能不能跑出**非空的清單**。
//
// 進得去（`docs/re/73`）只是第一步；進去之後如果商品清單是空的、
// 疾病清單是空的、技能清單是空的，那家店就是個空殼——
// 畫面畫得出來、測試全綠，玩家按什麼都沒反應。
func TestFacilityCoverage(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	kindName := map[game.FacilityKind]string{
		game.FacilityDoctor: "醫生", game.FacilityShop: "商店",
		game.FacilityTrainer: "訓練", game.FacilityRoster: "角色管理",
		game.FacilityEnding: "結局",
	}
	byKind := map[game.FacilityKind]int{}
	empty := map[game.FacilityKind]int{}
	var total, noName int

	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for i := 0; i < blk.SectionCount(6); i++ {
			rec, err := blk.SectionRecord(6, i)
			if err != nil || len(rec) == 0 || rec[0]&0x80 == 0 {
				continue
			}
			f, ok := game.ParseFacility(rec)
			if !ok {
				// bit7 設但 kind ≥ 5 ＝ **直接指定 opcode** 的腳本記錄
				// （`docs/re/79` §1），不是設施——跳過不算。
				if int(rec[0]&0x7F) >= game.FacilityKindCount {
					continue
				}
				t.Errorf("資源 %d 記錄 %d 的 bit7 設著、kind < 5，卻解不出設施",
					res.ID, i)
				continue
			}
			total++
			byKind[f.Kind]++
			if f.Name == "" {
				noName++
			}
			fs := s.EnterFacility(rec)
			if fs == nil {
				t.Errorf("資源 %d 記錄 %d：EnterFacility 回 nil", res.ID, i)
				continue
			}
			rows := -1
			switch f.Kind {
			case game.FacilityShop:
				rows = len(fs.buyList(s.items))
			case game.FacilityDoctor:
				// 醫生的清單是「角色身上的病」，與這家店無關；
				// 這家店該有的是三個價格（docs/re/35 §4.1）。
				if len(rec) > 0x06 && rec[0x04]|rec[0x05]|rec[0x06] != 0 {
					rows = 3
				} else {
					rows = 0
				}
			case game.FacilityTrainer:
				rows = len(fs.Skills)
			}
			if rows == 0 {
				empty[f.Kind]++
			}
		}
	}
	s.LeaveFacility()

	kinds := make([]int, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, int(k))
	}
	sort.Ints(kinds)
	for _, k := range kinds {
		kk := game.FacilityKind(k)
		t.Logf("%s：%d 家，清單空的 %d 家", kindName[kk], byKind[kk], empty[kk])
	}
	t.Logf("設施共 %d 家，沒有名字的 %d 家", total, noName)

	// 商店與醫生必須家家有內容——空的代表那家店進去了什麼都不能做。
	if empty[game.FacilityShop] != 0 {
		t.Errorf("%d 家商店的商品清單是空的", empty[game.FacilityShop])
	}
	if empty[game.FacilityDoctor] != 0 {
		t.Errorf("%d 家醫生沒有價格欄位", empty[game.FacilityDoctor])
	}
	// 訓練師列的是整張技能資料表（`docs/re/80`）——每家店都一樣，
	// 篩選發生在選完之後。空的代表技能表讀不到。
	if empty[game.FacilityTrainer] != 0 {
		t.Errorf("%d 家訓練師的技能清單是空的", empty[game.FacilityTrainer])
	}
}
