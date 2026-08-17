package play

import (
	"sort"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// TestScriptOpcodeCoverage 盤點 42 張地圖實際用到的腳本 opcode，
// 以及 remake 的直譯器有幾個回報 Handled ＝ false。
//
// `Handled ＝ false` 是**啞掉的缺口**——遊戲照跑、測試全綠，
// 代價是那一格什麼都不會發生。這個測試把缺口變成看得見的數字。
func TestScriptOpcodeCoverage(t *testing.T) {
	rom := openRom(t)
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	type stat struct{ cells, records int }
	used := map[int]*stat{}
	unhandled := map[int]bool{}

	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		cells := map[byte]int{}
		for y := 0; y < blk.Dim; y++ {
			for x := 0; x < blk.Dim; x++ {
				if terrain, idx, _, err := blk.At(x, y); err == nil && terrain == 6 {
					cells[idx]++
				}
			}
		}
		w := game.NewWorld(blk, &game.Party{
			Members: []*game.Character{{CON: 20, MaxCON: 20}},
		}, rng.New())
		for i := 0; i < blk.SectionCount(6); i++ {
			rec, err := blk.SectionRecord(6, i)
			if err != nil || len(rec) == 0 || rec[0]&0x80 != 0 {
				continue // 設施不是腳本
			}
			op, err := blk.SectionEntry(0x10, int(rec[0]))
			if err != nil {
				continue
			}
			s := used[int(op)]
			if s == nil {
				s = &stat{}
				used[int(op)] = s
			}
			s.records++
			s.cells += cells[byte(i)]
			// 拿記錄的副本試跑，避免污染區塊
			probe := append([]byte(nil), rec...)
			if r := (&game.Script{World: w, Record: probe, Op: int(op)}).Step(); !r.Handled {
				unhandled[int(op)] = true
			}
		}
	}

	ops := make([]int, 0, len(used))
	for op := range used {
		ops = append(ops, op)
	}
	sort.Ints(ops)
	var missCells, missRecords int
	for _, op := range ops {
		mark := ""
		if unhandled[op] {
			mark = "  ← 未實作"
			missCells += used[op].cells
			missRecords += used[op].records
		}
		t.Logf("opcode %2d：%3d 筆記錄、%4d 格%s", op, used[op].records, used[op].cells, mark)
	}
	t.Logf("出貨資料用到 %d 種 opcode，未實作 %d 種（%d 筆記錄、%d 格）",
		len(ops), len(unhandled), missRecords, missCells)

	// **有格子指到的 opcode 一個都不能漏。** 那些是玩家走過去就會踩到的，
	// 沒實作等於那一格什麼都不發生——遊戲照跑、測試全綠，缺口完全沒有聲音。
	// 只有格子數 0 的（靠改寫才到得了）允許暫時未實作。
	if missCells != 0 {
		t.Errorf("有格子指到的 opcode 還有 %d 格沒實作", missCells)
	}
	// **44 種 opcode 全部實作完了**（`docs/re/102`、`docs/re/104`）。
	//
	// 剩下的 5 筆全部是 section `0x10` 的**索引越界**（查出來的「opcode」是
	// 1282／2271／26478／29813），也就是「這一筆記錄根本不是腳本」——
	// `Step()` 本來就該擋掉它們（`docs/re/76` §3）。
	//
	// ⚠ **這個門檻降到底了。** 再往下只有一種可能：那四個越界值不再出現，
	// 而那代表 section `0x10` 的解讀改了——那時候要回頭看 `docs/re/71` §5.2，
	// 不是調這個數字。
	if missRecords > 5 {
		t.Errorf("未實作的 opcode 記錄數 %d 超過門檻 5", missRecords)
	}
	if len(unhandled) > 4 {
		t.Errorf("未實作的 opcode 種類 %d 超過門檻 4（只該剩四個越界值）",
			len(unhandled))
	}
	// 越界值以外的一律不准出現：**44 種都實作了，正常的 opcode 不該落進來**。
	for op := range unhandled {
		if op >= 0 && op < game.OpCount {
			t.Errorf("opcode %d 是合法編號卻沒實作", op)
		}
	}
}
