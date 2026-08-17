package render

// 畫字改成**直接讀點陣**（`blit` 收 []byte）之後，畫出來的像素要與
// 攤成 `[][]bool` 的舊路一模一樣。
//
// ⚠ 「有字、看起來沒壞」證明不了這件事：位元序讀錯是左右鏡像、
// 每列 byte 數算錯是逐列漸進偏移，**兩種都畫得出像字的東西**。
// 所以這裡逐像素比對兩條路，不看樣子。

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// blitRows 是攤平版的畫法（`assets.rowsOf` ＋ 逐格 Set），留在測試裡當 oracle。
func blitRows(h *HiFrame, rows [][]bool, col, row, w, ht int, fg byte) {
	x0 := col*HiCellWidth + (HiCellWidth-w)/2
	y0 := row*HiCellHeight + (HiCellHeight-ht)/2
	for y, line := range rows {
		for x, on := range line {
			if on {
				h.Set(x0+x, y0+y, fg)
			}
		}
	}
}

func TestBlitMatchesRows(t *testing.T) {
	f := openETen(t)

	// 挑筆劃差很多的幾個字：一（一橫）、中（對稱）、鬱（最密）、
	// 全形逗號（在符號區，走另一條索引）。
	cases := []struct {
		name   string
		hi, lo byte
	}{
		{"一", 0xA4, 0x40},
		{"中", 0xA4, 0xA4},
		{"鬱", 0xC6, 0x74},
		{"，", 0xA1, 0x41},
	}
	for _, tc := range cases {
		rows := f.GlyphRows(tc.hi, tc.lo)
		if rows == nil {
			t.Fatalf("取不到 %s", tc.name)
		}
		// 位置也要一起驗：置中的位移算錯不會影響形狀，只會整個字挪一格。
		for _, at := range [][2]int{{0, 0}, {7, 3}, {37, 22}} {
			var a, b HiFrame
			blitRows(&a, rows, at[0], at[1], f.Width, f.Height, 15)
			if !b.DrawCJK(f, tc.hi, tc.lo, at[0], at[1], 15) {
				t.Fatalf("%s 畫不出來", tc.name)
			}
			if a.Pix != b.Pix {
				t.Fatalf("%s 在 (%d, %d) 兩條路畫出來不一樣", tc.name, at[0], at[1])
			}
		}
	}
}

func TestBlitMatchesRowsASCII(t *testing.T) {
	f := openETen(t)
	if !f.HasASCII() {
		t.Skip("這份字型沒有半形 ASCII")
	}
	for _, c := range []byte{'A', 'g', '0', '%', ' ', 0x7F} {
		rows := f.ASCIIRows(c)
		if rows == nil {
			t.Fatalf("取不到 %q", c)
		}
		var a, b HiFrame
		blitRows(&a, rows, 5, 9, f.ASCIIWidth, f.ASCIIHeight, 15)
		if !b.DrawETenASCII(f, c, 5, 9, 15) {
			t.Fatalf("%q 畫不出來", c)
		}
		if a.Pix != b.Pix {
			t.Fatalf("%q 兩條路畫出來不一樣", c)
		}
	}
}

// 畫一個字**一次配置都不能有**——這是每幀每字都會跑的路。
func TestDrawCJKDoesNotAllocate(t *testing.T) {
	f := openETen(t)
	var h HiFrame
	got := testing.AllocsPerRun(50, func() {
		h.DrawCJK(f, 0xA4, 0xA4, 3, 3, 15)
		h.DrawETenASCII(f, 'A', 4, 3, 15)
	})
	if got != 0 {
		t.Errorf("畫兩個字配置了 %v 次，預期 0", got)
	}
}

// 字型檔被截斷時**不准畫半個字**——寧可整格空白。
func TestBlitRejectsShortGlyph(t *testing.T) {
	var h HiFrame
	if h.blit(make([]byte, 3), 0, 0, 24, 24, 15) {
		t.Error("點陣不足一個字，blit 卻說畫了")
	}
	for _, v := range h.Pix {
		if v != 0 {
			t.Fatal("畫布被動到了")
		}
	}
	_ = assets.ETenFont{}
}
