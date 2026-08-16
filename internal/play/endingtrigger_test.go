package play

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// 設施／腳本跳表的兩個事實，寫成機器守得住的形狀。
//
// `docs/re/96` 把結局定位在跳表的第 4 格，而**資料裡沒有任何一筆記錄用得到
// 那一格**——玩家因此走不到結局。這兩道把「表長什麼樣」與「資料用了哪幾格」
// 各釘一次，之後補上觸發點時哪一邊變了都看得出來。

// facilityJumpTable 是 `ds:A4E0h`：0–4 五種設施、5–48 四十四個腳本 opcode。
const (
	facilityJumpTableAt   = 0xA4E0
	facilityJumpTableSize = 49
	endingHandler         = 0x0B4F0 // 第 4 格（`docs/re/96` §1）
)

func TestFacilityJumpTableShape(t *testing.T) {
	rom := openRom(t)
	raw, err := rom.DsBytes(facilityJumpTableAt, (facilityJumpTableSize+4)*2)
	if err != nil {
		t.Skipf("讀不到 ds:A4E0h：%v", err)
	}
	word := func(i int) uint16 { return uint16(raw[i*2]) | uint16(raw[i*2+1])<<8 }

	if got := word(4); int(got) != endingHandler {
		t.Errorf("第 4 格應該是結局 %#06x，得到 %#06x", endingHandler, got)
	}
	for i := 0; i < facilityJumpTableSize; i++ {
		if word(i) == 0 {
			t.Errorf("第 %d 格是 0，表應該有 %d 格", i, facilityJumpTableSize)
		}
	}
	// 表尾要是 0——**索引空間有多大決定了「kind 92 指到哪裡」**。
	if got := word(facilityJumpTableSize); got != 0 {
		t.Errorf("第 %d 格應該是表尾 0，得到 %#06x", facilityJumpTableSize, got)
	}
}

// TestEndingHasNoTriggerInData 記的是一個**負面事實**，而且它是刻意留紅不了的：
// 42 張地圖的 nibble 6 記錄裡沒有一筆是 kind 4，所以玩家走不到結局。
//
// 補上觸發點（或發現它其實走別條路）之後，這一道會紅——**那時要改的是這裡的
// 期望值，不是把測試刪掉**。負面事實沒有守著就會在下一輪被當成「已經解決」。
func TestEndingHasNoTriggerInData(t *testing.T) {
	rom := openRom(t)
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[byte]int{}
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
			seen[rec[0]&0x7F]++
		}
	}
	kinds := make([]int, 0, len(seen))
	for k := range seen {
		kinds = append(kinds, int(k))
	}
	sort.Ints(kinds)
	var used []string
	for _, k := range kinds {
		used = append(used, fmt.Sprintf("%d×%d", k, seen[byte(k)]))
	}
	t.Logf("資料裡用到的跳表索引：%s", strings.Join(used, " "))

	if n := seen[4]; n != 0 {
		t.Errorf("結局（第 4 格）在資料裡出現 %d 次——觸發點找到了，"+
			"把 WORKLIST 的 A0 那一列與這一道的期望值一起更新", n)
	}
	// 正對照：四種設施都要在，否則是掃描壞了而不是「真的沒有」。
	for _, k := range []byte{0, 1, 2, 3} {
		if seen[k] == 0 {
			t.Errorf("kind %d 一筆都沒掃到——掃描本身有問題，"+
				"這時候「kind 4 是 0」證明不了任何事", k)
		}
	}
}
