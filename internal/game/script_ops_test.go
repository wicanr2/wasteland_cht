package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 這一批 opcode 出貨資料裡**一格都沒有指到**（`docs/re/76`），所以測試要
// 自己造記錄——那正是「走不到的路徑」唯一到得了的方式。
//
// ⚠ 造出來的是**輸入**不是規則：每一條的期望值都來自 `docs/re/102` 讀出來的
// 那一段組語，不是從實作反推的。

// scriptBlock 造一個帶 section 3 與 section 5 的區塊。
//
// 版面照 `docs/re/16`：標頭裡放指標陣列的位址、陣列裡放每一筆記錄的位址。
// 每一筆固定 recLen bytes，夠放到 `+0x09`。
func scriptBlock(t *testing.T, counts map[int]int) *assets.Block {
	t.Helper()
	const (
		headerAt = 0x400
		recLen   = 0x10
		dim      = 32
	)
	raw := make([]byte, 0x2000)
	offsets := make([]uint16, 24)
	at := 0x500
	for typ, n := range counts {
		hdrSlot := uint16(typ * 2)
		offsets[typ] = hdrSlot
		ptrs := at
		raw[headerAt+int(hdrSlot)] = byte(ptrs & 0xFF)
		raw[headerAt+int(hdrSlot)+1] = byte(ptrs >> 8)
		recs := ptrs + n*2
		for i := 0; i < n; i++ {
			r := recs + i*recLen
			raw[ptrs+i*2] = byte(r & 0xFF)
			raw[ptrs+i*2+1] = byte(r >> 8)
		}
		at = recs + n*recLen
	}
	cells := dim * dim
	return &assets.Block{
		Resource:       assets.Resource{ID: 7},
		Header:         make([]byte, 0x5C),
		Dim:            dim,
		Terrain:        make([]byte, cells),
		Record:         make([]byte, cells),
		Graphic:        make([]byte, cells),
		Raw:            raw,
		SectionOffsets: offsets,
		HeaderAt:       headerAt,
	}
}

// runOp 直接指定 opcode 跑一步（繞過 section 0x10 的查表）。
func runOp(w *World, op int, rec []byte) ScriptResult {
	s := &Script{World: w, Record: rec, Op: op}
	return s.Step()
}

// opcode 0：**有沒有另一支隊伍站在記錄 `+0x07` 起那串座標上**（`0x1A470`）。
func TestOpMatchPlace(t *testing.T) {
	newWorld := func() *World {
		w := NewWorld(scriptBlock(t, map[int]int{3: 4}), &Party{}, nil)
		w.MapID, w.GroupIndex = 5, 0
		w.Groups[0] = GroupPos{X: 3, Y: 4, MapID: 5, Active: true} // 自己這一組
		return w
	}
	// 記錄：+0x01/+0x02 是「下一步」，+0x03/+0x04 與 +0x05/+0x06 是兩個分支，
	// +0x07 起是座標對，0xFF 結束。
	rec := func() []byte {
		return []byte{0, 0, 0, 0xAA, 0xAB, 0xBB, 0xBC, 10, 20, 30, 40, 0xFF}
	}

	t.Run("別組站在清單上的第二對座標", func(t *testing.T) {
		w := newWorld()
		w.Groups[1] = GroupPos{X: 30, Y: 40, MapID: 5, Active: true}
		r := rec()
		if res := runOp(w, OpMatchPlace, r); !res.Handled {
			t.Fatal("opcode 0 沒有實作")
		}
		if r[1] != 0xAA || r[2] != 0xAB {
			t.Errorf("命中應該走 +0x03，得到 %#02x %#02x", r[1], r[2])
		}
	})
	t.Run("別組在同一格但在別張地圖", func(t *testing.T) {
		w := newWorld()
		w.Groups[1] = GroupPos{X: 10, Y: 20, MapID: 6, Active: true}
		r := rec()
		runOp(w, OpMatchPlace, r)
		if r[1] != 0xBB || r[2] != 0xBC {
			t.Errorf("不同地圖不算，應該走 +0x05，得到 %#02x %#02x", r[1], r[2])
		}
	})
	t.Run("站在上面的是自己這一組", func(t *testing.T) {
		// ⚠ 原版明確跳過 `ds:4654h`（目前這組）。少了這個判斷，
		// 隊伍站在自己觸發的那一格上會**永遠**走命中那一條。
		w := newWorld()
		w.Groups[0] = GroupPos{X: 10, Y: 20, MapID: 5, Active: true}
		r := rec()
		runOp(w, OpMatchPlace, r)
		if r[1] != 0xBB || r[2] != 0xBC {
			t.Errorf("自己這一組不算，應該走 +0x05，得到 %#02x %#02x", r[1], r[2])
		}
	})
}

// opcode 1：第一個隊員站著 → +0x03，倒下 → +0x05（`0x1A4F4`）。
func TestOpBranchOnFirstMember(t *testing.T) {
	for _, tc := range []struct {
		name     string
		con      int16
		wantLow  byte
		wantHigh byte
	}{
		{"站著", 20, 0xAA, 0xAB},
		{"倒下", -5, 0xBB, 0xBC},
		{"死了（CON ＝ 0）", 0, 0xBB, 0xBC},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(scriptBlock(t, map[int]int{3: 1}),
				&Party{Members: []*Character{{CON: tc.con}}}, nil)
			r := []byte{0, 0, 0, 0xAA, 0xAB, 0xBB, 0xBC}
			runOp(w, OpBranch, r)
			if r[1] != tc.wantLow || r[2] != tc.wantHigh {
				t.Errorf("得到 %#02x %#02x，預期 %#02x %#02x",
					r[1], r[2], tc.wantLow, tc.wantHigh)
			}
		})
	}
}

// opcode 7：沒帶物品 0x2F 的人 CON ← −5；只要有一個人帶著就走 +0x05（`0x1A699`）。
func TestOpNeedItem(t *testing.T) {
	t.Run("有人帶著", func(t *testing.T) {
		with := &Character{Name: "有", CON: 20, Items: []Slot{{ID: needItemID}}}
		without := &Character{Name: "沒有", CON: 18}
		w := NewWorld(scriptBlock(t, map[int]int{3: 1}),
			&Party{Members: []*Character{with, without}}, nil)
		r := []byte{0, 0, 0, 0xAA, 0xAB, 0xBB, 0xBC}
		runOp(w, OpNeedItem, r)
		if with.CON != 20 {
			t.Errorf("帶著的人不該被扣：CON %d", with.CON)
		}
		if without.CON != needItemCON {
			t.Errorf("沒帶的人 CON 應該是 %d，得到 %d", needItemCON, without.CON)
		}
		if without.PreHurt != 18 {
			t.Errorf("扣血前的值要備份到 +0x26，得到 %d", without.PreHurt)
		}
		if r[1] != 0xBB || r[2] != 0xBC {
			t.Errorf("有人帶著要走 +0x05，得到 %#02x %#02x", r[1], r[2])
		}
	})
	t.Run("一個都沒有", func(t *testing.T) {
		w := NewWorld(scriptBlock(t, map[int]int{3: 1}),
			&Party{Members: []*Character{{CON: 20}, {CON: 20}}}, nil)
		r := []byte{0, 0, 0, 0xAA, 0xAB, 0xBB, 0xBC}
		runOp(w, OpNeedItem, r)
		if r[1] != 0xAA || r[2] != 0xAB {
			t.Errorf("一個都沒有才走 +0x03，得到 %#02x %#02x", r[1], r[2])
		}
	})
}

// opcode 8／9／34／37：往 section 5 的某一筆放東西（`0x1A6F5` 等四處）。
//
// ⚠ 四支寫的是**不同的筆數**，而且 op 9 的第二個參數是常數 0 不是 `+0x04`。
// 抄錯一個數字的症狀是「東西出現在別的地方」，不會有任何斷言變紅。
func TestOpPlaceInSection5(t *testing.T) {
	for _, tc := range []struct {
		name  string
		op    int
		index int
		wantB byte
	}{
		{"op 8 → 第 9 筆", OpPlace9, 9, 0x77},
		{"op 34 → 第 9 筆", OpPlace9Param, 9, 0x77},
		{"op 9 → 第 25 筆，第二個參數是常數 0", OpPlace25, 25, 0},
		{"op 37 → 第 15 筆", OpPlace15, 15, 0x77},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := scriptBlock(t, map[int]int{5: 30})
			w := NewWorld(b, &Party{}, nil)
			r := []byte{0, 0, 0, 0x66, 0x77}
			if res := runOp(w, tc.op, r); !res.Handled {
				t.Fatalf("opcode %d 沒有實作", tc.op)
			}
			rec, err := b.SectionRecord(5, tc.index)
			if err != nil {
				t.Fatal(err)
			}
			if rec[2] != 0x5E || rec[3] != 0x66 || rec[4] != tc.wantB {
				t.Errorf("第 %d 筆得到 %#02x %#02x %#02x，預期 0x5e 0x66 %#02x",
					tc.index, rec[2], rec[3], rec[4], tc.wantB)
			}
		})
	}
}

// opcode 32：section 3 的第 0x1D 筆**往回到第 0 筆**，每筆寫 `+0x03` 與 `+0x09`
//（`0x1AA3F`）。
func TestOpFillPair(t *testing.T) {
	b := scriptBlock(t, map[int]int{3: 0x24})
	w := NewWorld(b, &Party{}, nil)
	r := []byte{0, 0, 0, 0x12, 0x34}
	if res := runOp(w, OpFillPair, r); !res.Handled {
		t.Fatal("opcode 32 沒有實作")
	}
	for i := 0; i <= 0x1D; i++ {
		rec, err := b.SectionRecord(3, i)
		if err != nil {
			t.Fatal(err)
		}
		if rec[3] != 0x12 || rec[9] != 0x34 {
			t.Fatalf("第 %#x 筆得到 +0x03=%#02x +0x09=%#02x", i, rec[3], rec[9])
		}
	}
	// 範圍外的不能被動到——上界是 0x1D 不是「全部」。
	rec, err := b.SectionRecord(3, 0x1E)
	if err != nil {
		t.Fatal(err)
	}
	if rec[3] != 0 || rec[9] != 0 {
		t.Errorf("第 0x1E 筆被動到了：+0x03=%#02x +0x09=%#02x", rec[3], rec[9])
	}
}

// opcode 36：三格都是 (nibble 4, 記錄 2) 才把下一步設成 (0x0C, 3)（`0x1AB35`）。
func TestOpMatchNeighbours(t *testing.T) {
	set := func(b *assets.Block, n int) {
		for i := 0; i < n; i++ {
			c := neighbourCells[i]
			at := c[1]*b.Dim + c[0]
			b.Terrain[at], b.Record[at] = neighbourTerrain, neighbourRecord
		}
	}
	t.Run("三格都符合", func(t *testing.T) {
		b := scriptBlock(t, map[int]int{3: 1})
		set(b, 3)
		w := NewWorld(b, &Party{}, nil)
		r := []byte{0, 0, 0}
		runOp(w, OpMatchNeigh, r)
		if r[1] != 0x0C || r[2] != 3 {
			t.Errorf("得到 %#02x %#02x，預期 0x0c 0x03", r[1], r[2])
		}
	})
	t.Run("只有兩格符合", func(t *testing.T) {
		// ⚠ 原版是三個連著的 `jnz` ——**任何一格不符就整個不做**。
		b := scriptBlock(t, map[int]int{3: 1})
		set(b, 2)
		w := NewWorld(b, &Party{}, nil)
		r := []byte{0, 0, 0}
		runOp(w, OpMatchNeigh, r)
		if r[1] != 0 || r[2] != 0 {
			t.Errorf("不該動到下一步，得到 %#02x %#02x", r[1], r[2])
		}
	})
}

// opcode 4 與 39 是一對：寄放一個角色，過了門檻時間再領回來
//（`0x1A54B`／`0x1AC03`）。
func TestOpStashAndElapsed(t *testing.T) {
	newWorld := func() *World {
		w := NewWorld(scriptBlock(t, map[int]int{3: 1}),
			&Party{Members: []*Character{{
				Name: "Angela", CON: 20, MaxCON: 20, Money: 500, AC: 3,
				EquipIndex: 2, ArmorIndex: 4,
				Items:      []Slot{{ID: 9, Value: 1}},
				Skills:     []Slot{{ID: 1, Value: 3}},
			}}}, nil)
		return w
	}

	t.Run("寄放的是副本，原本那個人不動", func(t *testing.T) {
		w := newWorld()
		before := *w.Party.Members[0]
		runOp(w, OpStash, []byte{0, 0, 0, 2})
		got := w.Party.Members[0]
		if got.Money != before.Money || got.AC != before.AC ||
			got.EquipIndex != before.EquipIndex || len(got.Items) != len(before.Items) {
			t.Error("原本那個隊員被改到了——清除只發生在副本上")
		}
		st, ok := w.Stash[2]
		if !ok {
			t.Fatal("第 2 格沒有寄放到東西")
		}
		if st.Char.Name != "Angela" {
			t.Errorf("副本的名字是 %q", st.Char.Name)
		}
		if len(st.Char.Items) != 0 || st.Char.Money != 0 || st.Char.AC != 0 ||
			st.Char.EquipIndex != 0 || st.Char.ArmorIndex != 0 {
			t.Error("副本的物品／金錢／護甲／裝備要清掉")
		}
		if len(st.Char.Skills) == 0 {
			t.Error("技能陣列不在清除範圍裡（原版只清 +0xBD–+0xF8）")
		}
	})

	t.Run("時候未到走 +0x06", func(t *testing.T) {
		w := newWorld()
		runOp(w, OpStash, []byte{0, 0, 0, 2})
		w.Clock.Total = uint32(stashReturnTicks) - 1
		r := []byte{0, 0, 0, 2, 0x44, 0x45, 0x66, 0x67, 0x88, 0x89}
		runOp(w, OpElapsed, r)
		if r[1] != 0x66 || r[2] != 0x67 {
			t.Errorf("得到 %#02x %#02x，預期 +0x06 的 0x66 0x67", r[1], r[2])
		}
		if len(w.Party.Members) != 1 {
			t.Error("時候沒到不該把人放回隊伍")
		}
	})

	t.Run("時候到了就領回來，走 +0x04", func(t *testing.T) {
		w := newWorld()
		runOp(w, OpStash, []byte{0, 0, 0, 2})
		w.Clock.Total = uint32(stashReturnTicks)
		r := []byte{0, 0, 0, 2, 0x44, 0x45, 0x66, 0x67, 0x88, 0x89}
		runOp(w, OpElapsed, r)
		if r[1] != 0x44 || r[2] != 0x45 {
			t.Errorf("得到 %#02x %#02x，預期 +0x04 的 0x44 0x45", r[1], r[2])
		}
		if len(w.Party.Members) != 2 {
			t.Fatalf("隊伍應該多一個人，得到 %d", len(w.Party.Members))
		}
		if w.Party.Members[1].Name != "Angela" {
			t.Errorf("領回來的是 %q", w.Party.Members[1].Name)
		}
		if _, still := w.Stash[2]; still {
			t.Error("領回來之後那一格還留著")
		}
	})

	t.Run("隊伍滿了走 +0x08", func(t *testing.T) {
		w := newWorld()
		runOp(w, OpStash, []byte{0, 0, 0, 2})
		for len(w.Party.Members) < stashMaxParty {
			w.Party.Members = append(w.Party.Members, &Character{CON: 10})
		}
		w.Clock.Total = uint32(stashReturnTicks)
		r := []byte{0, 0, 0, 2, 0x44, 0x45, 0x66, 0x67, 0x88, 0x89}
		runOp(w, OpElapsed, r)
		if r[1] != 0x88 || r[2] != 0x89 {
			t.Errorf("得到 %#02x %#02x，預期 +0x08 的 0x88 0x89", r[1], r[2])
		}
		if len(w.Party.Members) != stashMaxParty {
			t.Error("隊伍滿了不該再塞人進來")
		}
	})
}

// opcode 14：印倒數然後**停住**（`0x1A7E8` 收尾是 `stc`）。
func TestOpCountdown(t *testing.T) {
	w := NewWorld(scriptBlock(t, map[int]int{3: 1}), &Party{}, nil)
	w.SelfDestruct = SelfDestruct{Armed: true, At: 0}

	for _, tc := range []struct {
		elapsed uint32
		m, s, h int
	}{
		{0, 1, 0, 0},    // 240 刻 ÷ 4 ＝ 60 秒 ＝ 1:00.00
		{1, 0, 59, 45},  // 剩 239：239 & 3 ＝ 3 → 45；239 >> 2 ＝ 59
		{120, 0, 30, 0}, // 剩 120 → 30 秒
		{240, 0, 0, 0},  // 到了
		{300, 0, 0, 0},  // 過頭也不能變成負的
	} {
		w.Clock.Total = tc.elapsed
		res := runOp(w, OpCountdown, []byte{0, 0, 0})
		if !res.Handled {
			t.Fatal("opcode 14 沒有實作")
		}
		if res.Continue {
			t.Error("印完倒數要停住腳本（原版是 stc）")
		}
		if res.Countdown == nil {
			t.Fatal("沒有回報倒數")
		}
		got := *res.Countdown
		if got.Minutes != tc.m || got.Seconds != tc.s || got.Hundredths != tc.h {
			t.Errorf("經過 %d 刻：得到 %d:%02d.%02d，預期 %d:%02d.%02d",
				tc.elapsed, got.Minutes, got.Seconds, got.Hundredths, tc.m, tc.s, tc.h)
		}
		if got.Head != countdownHeadString || got.Tail != countdownTailString {
			t.Errorf("前後字串編號是 %#x／%#x", got.Head, got.Tail)
		}
	}
}

// opcode 2 是**唯一還沒實作的**：它把兩個參數交給 overlay 的 `sub_10036`，
// 而那支的參數語意還沒讀（`docs/re/34` 標 `?`）。
//
// ⚠ 這一條是**負面斷言**：它守著「不要為了讓數字好看而猜一個行為填進去」。
// 真的解出來之後把這條改掉，不要讓它擋著。
func TestOpOverlayStaysUnhandled(t *testing.T) {
	w := NewWorld(scriptBlock(t, map[int]int{3: 1}), &Party{}, nil)
	if res := runOp(w, OpOverlay, []byte{0, 0, 0, 1, 2}); res.Handled {
		t.Error("opcode 2 的語意還沒讀出來，不該回報 Handled")
	}
}
