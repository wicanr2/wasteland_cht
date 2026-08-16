package render

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 倚天字型檔玩家自備、不入版控，沒有就跳過（docs/spec/10 §4）。
func openETen(t *testing.T) *assets.ETenFont {
	t.Helper()
	dir := os.Getenv("WL_ETEN")
	if dir == "" {
		dir = "../../workplace/eten"
	}
	f, err := assets.LoadETen(dir)
	if err != nil {
		t.Skipf("找不到倚天字型（%v），跳過", err)
	}
	return f
}

// 把一個字畫成 ASCII art，肉眼可以核對。
func art(rows [][]bool) string {
	var b strings.Builder
	for _, r := range rows {
		for _, on := range r {
			if on {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// 驗收 1：索引 oracle。這關沒過就不要往下做——整批字會整體偏移，
// 看起來像「有字但都不對」（docs/spec/10 §5）。
func TestETenIndexOracle(t *testing.T) {
	f := openETen(t)

	// 「一」＝ A440，STDFONT 的第 0 個字：只有一條橫線。
	rows := f.GlyphRows(0xA4, 0x40)
	if rows == nil {
		t.Fatal("取不到「一」")
	}
	filled := 0
	for _, r := range rows {
		on := 0
		for _, v := range r {
			if v {
				on++
			}
		}
		if on > 0 {
			filled++
		}
	}
	if filled == 0 || filled > 4 {
		t.Fatalf("「一」應該只有一到數列有像素，得到 %d 列：\n%s", filled, art(rows))
	}
	t.Logf("「一」：\n%s", art(rows))

	// 「中」＝ A4A4、「猴」＝ B555：要有內容而且不是同一個字。
	zhong := f.GlyphRows(0xA4, 0xA4)
	hou := f.GlyphRows(0xB5, 0x55)
	if zhong == nil || hou == nil {
		t.Fatal("取不到「中」或「猴」")
	}
	if art(zhong) == art(hou) {
		t.Fatal("「中」與「猴」畫出來一樣，索引一定錯了")
	}
	countOn := func(rows [][]bool) int {
		n := 0
		for _, r := range rows {
			for _, v := range r {
				if v {
					n++
				}
			}
		}
		return n
	}
	// 「猴」筆畫比「中」多很多，這是一個便宜但有效的健全性檢查。
	if countOn(hou) <= countOn(zhong) {
		t.Errorf("「猴」的像素數 %d 應該比「中」的 %d 多", countOn(hou), countOn(zhong))
	}
	t.Logf("「中」：\n%s", art(zhong))
}

// 驗收 2：全形標點要從 SPCFONT 取到，不得落 fallback。
func TestETenPunctuation(t *testing.T) {
	f := openETen(t)
	for _, tc := range []struct {
		name   string
		hi, lo byte
	}{
		{"，", 0xA1, 0x41},
		{"。", 0xA1, 0x43},
		{"！", 0xA1, 0x49},
		{"？", 0xA1, 0x48},
		{"「", 0xA1, 0x62},
		{"」", 0xA1, 0x63},
	} {
		rows := f.GlyphRows(tc.hi, tc.lo)
		if rows == nil {
			t.Errorf("%s（%02X%02X）取不到——SPCFONT 沒帶或索引錯了", tc.name, tc.hi, tc.lo)
			continue
		}
		on := 0
		for _, r := range rows {
			for _, v := range r {
				if v {
					on++
				}
			}
		}
		if on == 0 {
			t.Errorf("%s 是全空的字模", tc.name)
		}
	}
}

// 驗收 3：640 × 400 是乾淨的 2×，每個原像素變成 2 × 2，不得插值。
func TestUpscaleIsNearest(t *testing.T) {
	f := NewFrame()
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			f.Set(x, y, byte((x*7+y*13)&0x0F))
		}
	}
	h := NewHiFrame()
	h.Upscale(f)
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			want := f.At(x, y)
			for dy := 0; dy < HiScale; dy++ {
				for dx := 0; dx < HiScale; dx++ {
					if got := h.At(x*HiScale+dx, y*HiScale+dy); got != want {
						t.Fatalf("(%d,%d) 的 %d×%d 區塊應該全是 %d，(%d,%d) 是 %d",
							x, y, HiScale, HiScale, want,
							x*HiScale+dx, y*HiScale+dy, got)
					}
				}
			}
		}
	}
}

// 驗收 4：中文字畫在原版字元格的位置上，字模比格子小時垂直置中。
func TestDrawCJKPosition(t *testing.T) {
	f := openETen(t)
	h := NewHiFrame()
	const col, row = 5, 3
	if !h.DrawCJK(f, 0xA4, 0xA4, col, row, 15) { // 「中」
		t.Fatal("畫不出「中」")
	}
	x0, y0 := col*HiCellWidth, row*HiCellHeight
	// 格子外面不得有像素。
	for y := 0; y < HiScreenHeight; y++ {
		for x := 0; x < HiScreenWidth; x++ {
			if h.At(x, y) == 0 {
				continue
			}
			if x < x0 || x >= x0+HiCellWidth ||
				y < y0 || y >= y0+HiCellHeight {
				t.Fatalf("(%d,%d) 有像素，超出第 (%d,%d) 格", x, y, col, row)
			}
		}
	}
	// 字模比格子小的時候要**置中**：24 點剛好填滿（沒有留白），
	// 15 點放進 24 的格子則上下都要留白。這裡只驗「不超出格子」，
	// 上面那個迴圈已經做完了；置中的算式在 HiFrame.blit。
	blank := true
	if f.Height < HiCellHeight {
		for x := x0; x < x0+HiCellWidth; x++ {
			if h.At(x, y0+HiCellHeight-1) != 0 {
				blank = false
			}
		}
	}
	if !blank {
		t.Errorf("字高 %d 放進 %d 的格子，最後一列應該留白", f.Height, HiCellHeight)
	}

	// 查不到的字要回 false，不能靜靜畫成空白。
	if h.DrawCJK(f, 0x41, 0x41, 0, 0, 15) {
		t.Error("非 Big5 的碼不該回 true")
	}
}
