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

// 框邊標籤的字模編碼（`docs/re/126` §1）：`(字元 & 0xDF) − 0x29`，
// 空白 `0x33`，頭尾各一個蓋子，**整段再 ＋0x1C 換成冷色那一組**。
func TestLabelGlyphs(t *testing.T) {
	got := labelGlyphs("ESC")
	// ⚠ 換色只套在 `0x18`–`0x33` 那一段：**左蓋 `0x17` 不在範圍內，原樣畫**，
	// 右蓋 `0x28` 在範圍內，會被換掉。看起來不對稱，但原版存的就是這兩個值，
	// 而且走的是同一支繪製常式。
	want := []int{
		labelCapLeft,
		0x1C + colourBankCool, // E
		0x2A + colourBankCool, // S
		0x1A + colourBankCool, // C
		labelCapRight + colourBankCool,
	}
	if len(got) != len(want) {
		t.Fatalf("ESC 應該是 %d 格，得到 %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 格是 %#x，預期 %#x", i, got[i], want[i])
		}
	}
	// 空白走 0x33 那一格，不是 ASCII 的 0x20。
	if g := labelGlyphs("A B")[2]; g != labelSpace+colourBankCool {
		t.Errorf("空白應該是 %#x，得到 %#x", labelSpace+colourBankCool, g)
	}
	// 位置照實機：`ROSTER ON` 在列 17、`ROSTER OFF` 在列 23——
	// **兩筆不同的記錄**，不是同一個按鈕換字。
	if LabelRosterOn.Row == LabelRosterOff.Row {
		t.Error("ROSTER ON 與 OFF 的列不該一樣（docs/re/126 §2）")
	}
	if LabelPoolMoney.Col != 21 || LabelPoolMoney.Row != 13 {
		t.Errorf("POOL MONEY 應該在 (21, 13)，得到 (%d, %d)",
			LabelPoolMoney.Col, LabelPoolMoney.Row)
	}
}

// 四個捲動箭頭是**直的**：欄 39 固定，字模往下疊（`docs/re/129` §3）。
// 橫著排會排到畫面外——39 已經是最後一欄。
func TestArrowLabelsStackDownwards(t *testing.T) {
	up := LabelListUp
	if len(up.Glyphs) != 2 {
		t.Fatalf("上箭頭應該是兩個字模，得到 %d", len(up.Glyphs))
	}
	for _, c := range []struct {
		col, row int
		want     bool
	}{
		{39, 3, true}, {39, 4, true}, {39, 5, false}, {39, 2, false},
		{38, 3, false}, {40, 3, false},
	} {
		if got := up.Hit(c.col, c.row); got != c.want {
			t.Errorf("(%d,%d) 點得到 ＝ %v，預期 %v", c.col, c.row, got, c.want)
		}
	}
	// 四筆都在欄 39，列是版面表寫死的。
	for _, c := range []struct {
		l   BoxLabel
		row int
	}{
		{LabelListUp, 3}, {LabelListDown, 10},
		{LabelMsgUp, 18}, {LabelMsgDown, 22},
	} {
		if c.l.Col != 39 || c.l.Row != c.row || !c.l.Vertical {
			t.Errorf("箭頭位置錯了：%+v，預期欄 39 列 %d", c.l, c.row)
		}
	}
	// 字模要換成冷色那一組，否則畫出來是暖色（`docs/re/126` §1）。
	if LabelListUp.Glyphs[0] != 0x1E+colourBankCool {
		t.Errorf("上箭頭第一個字模 ＝ %#x，預期換色後的 %#x",
			LabelListUp.Glyphs[0], 0x1E+colourBankCool)
	}
}
