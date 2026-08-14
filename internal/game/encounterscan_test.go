package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 驗收 1：只有 nibble 3 與 15 是遭遇格。
func TestIsEncounterCell(t *testing.T) {
	for n := 0; n < 16; n++ {
		want := n == 3 || n == 15
		if got := IsEncounterCell(byte(n)); got != want {
			t.Errorf("nibble %d：%v，預期 %v", n, got, want)
		}
	}
}

// 驗收 5：佇列每次清空重建，不累積。
func TestQueueClears(t *testing.T) {
	var q EncounterQueue
	q.Insert(0, QueueEntry{X: 1, Y: 2, Resource: 3, Distance: 15})
	if _, ok := q.Find(3, 1, 2); !ok {
		t.Fatal("放進去之後應該找得到")
	}
	q.Clear()
	if _, ok := q.Find(3, 1, 2); ok {
		t.Fatal("清空之後不該還找得到")
	}
	for _, e := range q.Slots {
		if e.Used {
			t.Fatal("清空之後不該有任何 Used")
		}
	}
}

// 佇列是 4 組 × 4 槽，第 g 組佔 [4g, 4g+4)。
func TestQueueGroupsAreSeparate(t *testing.T) {
	var q EncounterQueue
	for g := 0; g < QueueGroups; g++ {
		if !q.Insert(g, QueueEntry{X: byte(g + 1), Y: 1, Distance: 15}) {
			t.Fatalf("第 %d 組應該放得進去", g)
		}
	}
	for g := 0; g < QueueGroups; g++ {
		base := g * QueueSlotsPerGroup
		if !q.Slots[base].Used || q.Slots[base].X != byte(g+1) {
			t.Errorf("第 %d 組的第一格不對：%+v", g, q.Slots[base])
		}
		for i := base + 1; i < base+QueueSlotsPerGroup; i++ {
			if q.Slots[i].Used {
				t.Errorf("第 %d 組不該佔到槽 %d", g, i)
			}
		}
	}
	if q.Insert(QueueGroups, QueueEntry{}) || q.Insert(-1, QueueEntry{}) {
		t.Error("組編號越界應該回 false")
	}
}

// 驗收 6：四槽滿了之後，**距離最遠**的被擠掉；槽內由近到遠排。
func TestQueueEvictsFarthest(t *testing.T) {
	var q EncounterQueue
	for i, d := range []byte{40, 10, 30, 20} {
		q.Insert(0, QueueEntry{X: byte(i + 1), Distance: d})
	}
	for i := 1; i < QueueSlotsPerGroup; i++ {
		if q.Slots[i-1].Distance > q.Slots[i].Distance {
			t.Fatalf("槽內沒有由近到遠排：%v", q.Slots[:QueueSlotsPerGroup])
		}
	}
	if q.Insert(0, QueueEntry{X: 9, Distance: 50}) {
		t.Error("比最遠的還遠，不該擠得進去")
	}
	if !q.Insert(0, QueueEntry{X: 9, Distance: 25}) {
		t.Fatal("距離 25 應該擠掉 40")
	}
	if _, ok := q.Find(0, 2, 0); !ok {
		t.Error("距離 10 的那筆不該被擠掉")
	}
	for _, e := range q.Slots[:QueueSlotsPerGroup] {
		if e.Distance == 40 {
			t.Error("距離 40 應該已經被擠掉")
		}
	}
}

// 接戰值取全組最大；不能行動的人不算。
func TestEngagementTakesMax(t *testing.T) {
	members := []*Character{
		{CON: 0},  // 不能行動
		{CON: 10}, // 能行動
		{CON: 10}, // 能行動＋遠程
		nil,
	}
	if got := Engagement(members, nil); got != EngageClose {
		t.Errorf("沒有遠程時應該是 %d，得到 %d", EngageClose, got)
	}
	if got := Engagement(members, func(i int) bool { return i == 2 }); got != EngageFar {
		t.Errorf("有一個遠程就該是 %d，得到 %d", EngageFar, got)
	}
	// 全員倒下 → 0。
	if got := Engagement([]*Character{{CON: 0}, {CON: -3}}, nil); got != EngageNone {
		t.Errorf("全員倒下應該是 %d，得到 %d", EngageNone, got)
	}
	// 遠程旗標對不能行動的人無效。
	if got := Engagement([]*Character{{CON: 0}}, func(int) bool { return true }); got != EngageNone {
		t.Errorf("不能行動的人不該貢獻接戰值，得到 %d", got)
	}
}

// 驗收 2：掃描不越出地圖邊界，也不掃視窗外。
func TestScanStaysInsideWindowAndMap(t *testing.T) {
	b := testBlock(t, 32, nil)
	p := &Party{X: 0, Y: 0} // 視窗原點會是 (−9, −4)，左上大半在地圖外
	w := NewWorld(b, p, nil)

	got := w.ScanEncounters([QueueGroups]PartyGroupState{{Present: true, Engage: EngageClose}})
	// 視窗 19 × 9，隊伍在 (0,0) → x ∈ [−9, 9]、y ∈ [−4, 4]，
	// 地圖內的只有 x ∈ [0, 9]（10 行）、y ∈ [0, 4]（5 列）。
	if want := 10 * 5; got.Scanned != want {
		t.Errorf("應該掃 %d 格，掃了 %d 格", want, got.Scanned)
	}
}

// 驗收 3／4：察覺距離與主動距離的兩道關。
func TestScanDistanceGates(t *testing.T) {
	// 在 (12, 10) 放一個遭遇格（nibble 3），隊伍站在 (10, 10)。
	// 距離 (dx=2, dy=0) → 距離表是 20。
	cells := map[[2]int]byte{{12, 10}: 3}

	for _, tc := range []struct {
		name           string
		notice, active byte
		engage         int
		present        bool
		want           bool
	}{
		{"察覺距離不夠 → 完全不進", 10, 10, EngageFar, true, false},
		{"主動距離夠 → 進", 30, 20, EngageNone, true, true},
		{"主動距離不夠但接戰值夠 → 進", 30, 5, EngageFar, true, true},
		{"兩邊都不夠 → 不進", 30, 5, EngageClose, true, false},
		{"沒有任何一組在場 → 不進", 30, 20, EngageFar, false, false},
	} {
		b := testBlock(t, 32, func(rec []byte) {
			rec[recNoticeRange], rec[recActiveRange] = tc.notice, tc.active
			rec[0x03], rec[0x04] = 1, 2 // 第 0 組：型別 1、2 隻
		})
		setCells(b, cells)
		w := NewWorld(b, &Party{X: 10, Y: 10}, nil)
		got := w.ScanEncounters([QueueGroups]PartyGroupState{{Present: tc.present, Engage: tc.engage}})
		_, ok := got.Queue.Find(byte(b.Resource.ID), 12, 10)
		if ok != tc.want {
			t.Errorf("%s：進佇列 ＝ %v，預期 %v", tc.name, ok, tc.want)
		}
	}
}

// 沒有敵人的遭遇格不進佇列。
func TestScanSkipsCellWithNoEnemies(t *testing.T) {
	b := testBlock(t, 32, func(rec []byte) {
		rec[recNoticeRange], rec[recActiveRange] = 90, 90
		// 三組的數量全 0
	})
	setCells(b, map[[2]int]byte{{12, 10}: 3})
	w := NewWorld(b, &Party{X: 10, Y: 10}, nil)
	got := w.ScanEncounters([QueueGroups]PartyGroupState{{Present: true, Engage: EngageFar}})
	if got.Hits != 1 {
		t.Fatalf("應該命中 1 格 nibble 3，得到 %d", got.Hits)
	}
	if _, ok := got.Queue.Find(byte(b.Resource.ID), 12, 10); ok {
		t.Error("三組數量都是 0，不該進佇列")
	}
}

// testBlock 造一個最小的合成區塊：32 × 32、只有型別 3 的 section 有一筆記錄。
//
// 用合成區塊而不是原版資料，是因為這幾條驗收要控制距離門檻與敵人數量；
// 拿真地圖來測會變成「這張圖剛好怎樣」。真地圖的驗收在 game_test.go。
func testBlock(t *testing.T, dim int, fill func(rec []byte)) *assets.Block {
	t.Helper()
	const (
		headerAt  = 0x600
		hdrSlot   = 0x40  // SectionOffsets[3] 指到標頭的這個位移
		sectionAt = 0x700 // 指標陣列
		recordAt  = 0x720
		rawLen    = 0x800
	)
	raw := make([]byte, rawLen)
	raw[headerAt+hdrSlot], raw[headerAt+hdrSlot+1] = sectionAt&0xFF, sectionAt>>8
	raw[sectionAt], raw[sectionAt+1] = recordAt&0xFF, recordAt>>8
	if fill != nil {
		fill(raw[recordAt:])
	}
	offsets := make([]uint16, 24)
	offsets[3] = hdrSlot
	n := dim * dim
	return &assets.Block{
		Resource:       assets.Resource{ID: 7},
		Header:         make([]byte, 0x5C),
		Dim:            dim,
		Terrain:        make([]byte, n),
		Record:         make([]byte, n),
		Graphic:        make([]byte, n),
		Raw:            raw,
		SectionOffsets: offsets,
		HeaderAt:       headerAt,
	}
}

// setCells 把指定座標的第 1 層設成某個 nibble。
func setCells(b *assets.Block, cells map[[2]int]byte) {
	for at, n := range cells {
		b.Terrain[at[1]*b.Dim+at[0]] = n
	}
}
