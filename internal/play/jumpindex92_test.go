package play

import (
	"fmt"
	"testing"
)

// 跳表索引 92 的那兩筆記錄（`docs/re/98` §5.1）。
//
// `ds:A4E0h` 只有 49 格，而出貨資料裡有兩筆 nibble 6 的記錄寫著 92。
// 它們走不到——三道各自獨立，任何一道變了都代表「這兩筆記錄有路可走了」，
// 那時要做的是回去讀 `docs/re/98` §5.1 再決定，不是改期望值。
const (
	unreachableKind  = 92
	unreachableCount = 2
	patchStride      = 5    // nibble 12 的批次表一筆 5 bytes（`docs/re/71` §2）
	patchLast        = 0x80 // 旗標 bit7 ＝ 最後一筆
	patchNewLayer1   = 3    // 那一筆的 +3 是新的第 1 層
)

func TestUnreachableJumpIndex(t *testing.T) {
	rom := openRom(t)
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}

	found := 0
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		// 這個區塊有沒有 kind 92 的 nibble 6 記錄。
		hits := map[int]bool{}
		for i := 0; i < blk.SectionCount(6); i++ {
			rec, err := blk.SectionRecord(6, i)
			if err != nil || len(rec) == 0 || rec[0]&0x80 == 0 {
				continue
			}
			if rec[0]&0x7F == unreachableKind {
				hits[i] = true
				found++
			}
		}
		if len(hits) == 0 {
			continue
		}
		where := fmt.Sprintf("%s#%d", res.File, res.ID)
		t.Logf("%s 的 section 6 有 %d 筆 kind=%d", where, len(hits), unreachableKind)

		// 第一道：地圖上有沒有格子指過來。
		// 正對照在同一個迴圈裡：**這張地圖至少要解得出一格**，
		// 否則「沒有 nibble 6」只代表解碼失敗。
		cells, nibble6 := 0, 0
		for y := 0; y < 64; y++ {
			for x := 0; x < 64; x++ {
				terrain, record, _, err := blk.At(x, y)
				if err != nil {
					continue
				}
				cells++
				if terrain == 6 && hits[int(record)] {
					nibble6++
				}
			}
		}
		if cells == 0 {
			t.Fatalf("%s 一格都解不出來——這時候「沒有 nibble 6」證明不了任何事", where)
		}
		if nibble6 != 0 {
			t.Errorf("%s 有 %d 格指到 kind=%d 的記錄（%d 格解得出來）",
				where, nibble6, unreachableKind, cells)
		}

		// 第二道：nibble 12 的批次改寫有沒有把哪一格的第 1 層寫成 6。
		patches, toNibble6 := 0, 0
		for i := 0; i < blk.SectionCount(12); i++ {
			rec, err := blk.SectionRecord(12, i)
			if err != nil || len(rec) < 1+patchStride {
				continue
			}
			for p := 1; p+patchStride-1 < len(rec); p += patchStride {
				patches++
				if rec[p+patchNewLayer1] == 6 {
					toNibble6++
				}
				if rec[p]&patchLast != 0 {
					break
				}
			}
		}
		if patches == 0 {
			t.Fatalf("%s 一筆批次改寫都沒解出來——第二道證明不了任何事", where)
		}
		if toNibble6 != 0 {
			t.Errorf("%s 有 %d 筆批次改寫把第 1 層寫成 6（共 %d 筆）",
				where, toNibble6, patches)
		}
		t.Logf("  %d 格解得出來、%d 筆批次改寫，都沒有指過來", cells, patches)
	}

	if found != unreachableCount {
		t.Errorf("全部資料裡有 %d 筆 kind=%d，`docs/re/98` §5.1 記的是 %d 筆",
			found, unreachableKind, unreachableCount)
	}
}
