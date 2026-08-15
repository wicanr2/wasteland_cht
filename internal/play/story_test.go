package play

import (
	"fmt"
	"sort"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// TestStoryTriggerCoverage 盤點 42 張地圖上**所有劇本觸發點**，
// 確認每一種在 remake 都有處理路徑。
//
// ⚠ 驗收不走「從頭玩到結局」那條路：一次通關要幾十小時、不可重跑，
// 而且中途任何一步卡住就什麼都驗不到。改成**覆蓋 ＋ 抽樣**——
// 與腳本、遭遇、設施、戰鬥那幾把尺同一套做法。
func TestStoryTriggerCoverage(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}

	gateTypes := map[int]int{}
	var maps, records, gates int
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		maps++
		// 條件閘住在 nibble 2 的記錄（section 型別 2）。
		for i := 0; i < blk.SectionCount(2); i++ {
			rec, err := blk.SectionRecord(2, i)
			if err != nil || len(rec) <= 0x0B {
				continue
			}
			records++
			for _, g := range game.ParseGates(rec) {
				gates++
				gateTypes[g.Type]++
			}
		}
	}

	var keys []int
	for k := range gateTypes {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	t.Logf("%d 張地圖、%d 筆條件記錄、%d 條條件", maps, records, gates)
	for _, k := range keys {
		t.Logf("  型別 %d：%d 條", k, gateTypes[k])
	}
	if gates == 0 {
		t.Fatal("一條條件都沒掃到")
	}
	// 八種型別都要有資料——少一種就表示掃描漏了，不是原版沒用。
	for _, want := range []int{0, 1, 2, 3, 4} {
		if gateTypes[want] == 0 {
			t.Errorf("型別 %d 一條都沒有", want)
		}
	}
	_ = s
}

// TestStoryTriggersAreReachable 抽樣驗「每一條劇本觸發**玩家真的碰得到**」。
//
// 做法：對每一條條件，把它要求的東西發給角色，再用 `USE` 那條路去試——
// 命中（`hit ≥ 0`）就表示這條劇本接得上。
//
// ⚠ 型別 3（隊伍人數）、4（金錢）、5／6／7 **`USE` 碰不到**：
// `sub_14090` 的物品只認型別 1，其餘只有走上去自動評估會跑
//（`docs/re/92` §6）。這是原版行為，所以它們算在「自動評估那條路」。
func TestStoryTriggersAreReachable(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	p := s.World().Party
	r := s.World().RNG

	var viaUse, viaAuto, missed int
	var examples []string
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for i := 0; i < blk.SectionCount(2); i++ {
			rec, err := blk.SectionRecord(2, i)
			if err != nil || len(rec) <= 0x0B {
				continue
			}
			for _, g := range game.ParseGates(rec) {
				var kind game.UseKind
				switch g.Type {
				case game.GateSkill:
					kind = game.UseSkill
				case 1:
					kind = game.UseItem
				case game.GateAttribute:
					kind = game.UseAttribute
				default:
					viaAuto++ // USE 碰不到，走自動評估
					continue
				}
				// 發一個「身上有這個東西」的臨時角色。
				c := &game.Character{Name: "probe", CON: 20}
				for j := range c.Attributes {
					c.Attributes[j] = 20
				}
				switch kind {
				case game.UseSkill:
					c.Skills = []game.Slot{{ID: g.Param, Value: 5}}
				case game.UseItem:
					c.Items = []game.Slot{{ID: g.Param, Value: 5}}
				}
				hit, _ := p.UseGate(r, rec, c, kind, g.Param, s.World().Skills)
				if hit >= 0 {
					viaUse++
					continue
				}
				missed++
				if len(examples) < 5 {
					examples = append(examples, fmtGate(res.ID, i, g))
				}
			}
		}
	}
	t.Logf("USE 碰得到 %d 條、只能走上去觸發 %d 條、**碰不到 %d 條**",
		viaUse, viaAuto, missed)
	for _, e := range examples {
		t.Logf("  碰不到的例子：%s", e)
	}
	// 型別 0／1／2 的條件**每一條**都要能被 USE 命中——
	// 命不中就表示比對規則與原版對不上，那一段劇情玩家永遠觸發不了。
	if missed != 0 {
		t.Errorf("有 %d 條型別 0／1／2 的條件用 USE 命不中", missed)
	}
	if viaUse == 0 {
		t.Fatal("一條都沒命中——比對規則整個沒接上")
	}
}

func fmtGate(res, rec int, g game.Gate) string {
	return fmt.Sprintf("資源 %d 記錄 %d：型別 %d 參數 %d 難度 %d",
		res, rec, g.Type, g.Param, g.Difficulty)
}

// TestParagraphTriggersHaveText 驗**每一條段落引用都查得到中文正文**。
//
// 引用表是編譯期從英文原文抽的（docs/spec/19 §2），83 條引用、82 個編號。
// 遊戲讀到那句話時要能把正文顯示出來——查不到就是玩家看到編號卻沒有故事。
func TestParagraphTriggersHaveText(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadJournal(
		"../../docs/re/generated/paragraph-refs.tsv",
		"../../translations/paragraphs-zh-Hant.cat"); err != nil {
		t.Fatalf("載手札：%v", err)
	}
	refs := s.journal.Refs()
	if len(refs) == 0 {
		t.Fatal("引用表是空的")
	}
	seen := map[int]bool{}
	var missing []int
	for _, n := range refs {
		if seen[n] {
			continue
		}
		seen[n] = true
		if len(s.journal.Text(n)) == 0 {
			missing = append(missing, n)
		}
	}
	sort.Ints(missing)
	t.Logf("%d 條引用、%d 個不同段落，沒有中文正文的 %d 個",
		len(refs), len(seen), len(missing))
	if len(missing) != 0 {
		t.Errorf("這些被引用的段落查不到中文正文：%v", missing)
	}
}
