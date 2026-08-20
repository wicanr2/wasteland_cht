package render

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 計量表由上到下的字模（`docs/re/120` §3）。索引 0 ＝ 字元列 9。
func TestMeterGlyphs(t *testing.T) {
	// 沒有蓋氏計數器：底座換一組，讀數一律當成沒有 → 整條空的。
	got := MeterGlyphs(0, false)
	want := [MeterRows]int{
		meterCapPlain + 2, meterCapPlain, // 列 9、10：頂蓋
		meterEmpty, meterEmpty, meterEmpty, meterEmpty, // 列 11–14：空管
		meterNoCounter + 2, meterNoCounter, // 列 15、16：底座
	}
	if got != want {
		t.Errorf("沒有計數器時應該是空管：\n得到 %#x\n期望 %#x", got, want)
	}

	// 有計數器但視野裡沒有輻射格（保留值）→ 一樣是空管，只有底座不同。
	got = MeterGlyphs(MeterNoReading, true)
	want[MeterRows-1], want[MeterRows-2] = meterHaveCounter, meterHaveCounter+2
	if got != want {
		t.Errorf("沒有讀數時應該是空管：\n得到 %#x\n期望 %#x", got, want)
	}

	// 讀數 0（踩在輻射格上）→ 9 階全滿：四格 ＋ 半格，
	// 而半格頂到頂之後改成把頂蓋換成亮的那一組（`0x17F25` 的 `jz`）。
	got = MeterGlyphs(0, true)
	want = [MeterRows]int{
		meterCapLit + 2, meterCapLit,
		meterFullLow, meterFullLow, meterFullHigh, meterFullHigh,
		meterHaveCounter + 2, meterHaveCounter,
	}
	if got != want {
		t.Errorf("讀數 0 應該滿格：\n得到 %#x\n期望 %#x", got, want)
	}

	// 讀數 8（退一階）→ n ＝ 8：四格滿、沒有半格，頂蓋是暗的。
	got = MeterGlyphs(8, true)
	want[0], want[1] = meterCapPlain+2, meterCapPlain
	if got != want {
		t.Errorf("讀數 8 應該四格滿、頂蓋是暗的：\n得到 %#x\n期望 %#x", got, want)
	}

	// 讀數 24 → step 3、n ＝ 6：三格滿 ＋ 沒有半格。列 > 12 的那兩格換色。
	got = MeterGlyphs(24, true)
	want = [MeterRows]int{
		meterCapPlain + 2, meterCapPlain,
		meterEmpty, meterFullLow, meterFullHigh, meterFullHigh,
		meterHaveCounter + 2, meterHaveCounter,
	}
	if got != want {
		t.Errorf("讀數 24 應該三格滿：\n得到 %#x\n期望 %#x", got, want)
	}

	// 讀數 ≥ 72（step ≥ 9）→ 空管，與沒有讀數一樣。
	if MeterGlyphs(72, true) != MeterGlyphs(MeterNoReading, true) {
		t.Error("讀數 72 以上與沒有讀數應該畫成一樣")
	}
}

// 名單框：欄 0 與欄 39 是兩條邊，列 23 是下緣（`sub_16F70`，docs/re/125）。
//
// ⚠ 這一條擋的是「框畫了但被別的東西蓋掉」——畫面上很難分辨
// 「沒畫」與「畫了但很暗」，所以直接量像素。
func TestRosterBoxCorners(t *testing.T) {
	lit := func(f *Frame, col, row int) bool {
		for y := 0; y < CharHeight; y++ {
			for x := 0; x < CharWidth; x++ {
				if f.Pix[(row*CharHeight+y)*ScreenWidth+col*CharWidth+x] != 0 {
					return true
				}
			}
		}
		return false
	}
	f := NewFrame()
	font := testColourFont()
	if err := f.DrawRosterBox(font, true); err != nil {
		t.Fatalf("畫名單框失敗：%v", err)
	}
	for _, c := range []struct {
		col, row int
		what     string
	}{
		{0, RosterBoxTopRow, "上緣左角"},
		{MsgBorderRight, RosterBoxTopRow, "上緣右角"},
		{0, RosterMemberRow, "左邊框"},
		{MsgBorderRight, RosterMemberRow, "右邊框"},
		{0, RosterBoxBottomRow, "下緣左角"},
		{20, RosterBoxBottomRow, "下緣"},
		{MsgBorderRight, RosterBoxBottomRow, "下緣右角"},
	} {
		if !lit(f, c.col, c.row) {
			t.Errorf("%s（欄 %d 列 %d）一個亮點都沒有", c.what, c.col, c.row)
		}
	}
}

// testColourFont 造一份「每一格都是實心」的假彩色字型——這裡驗的是**位置**，
// 不是字模長什麼樣。
func testColourFont() *assets.Font {
	f := &assets.Font{Glyphs: make([]assets.Glyph, 256)}
	for i := range f.Glyphs {
		for j := range f.Glyphs[i].Pix {
			f.Glyphs[i].Pix[j] = 15
		}
	}
	return f
}
