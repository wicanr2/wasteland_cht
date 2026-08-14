package render

import (
	"image/png"
	"os"
	"testing"
)

// 樣張用的三行中文（已經是 Big5 bytes，免得在 Go 裡再實作一次編碼）。
var demoLines = [][]byte{
	{0xA7, 0x41, 0xAD, 0xCC, 0xBE, 0x44, 0xA8, 0xEC, 0xA7, 0xF0, 0xC0, 0xBB, 0xA1, 0x49},
	{0xB3, 0x6F, 0xB8, 0xCC, 0xA6, 0xB3, 0xA4, 0x40, 0xAD, 0xD3, 0xA4, 0x57, 0xC2, 0xEA, 0xAA, 0xBA, 0xBD, 0x63, 0xA4, 0x6C, 0xA1, 0x43},
	{0xA1, 0x75, 0xAE, 0xB3, 0xA5, 0x68, 0xA7, 0x61, 0xA1, 0x41, 0xA1, 0x76, 0xA5, 0x4C, 0xBB, 0xA1, 0xA1, 0x43},
}

// 每一個字都要畫得出來——缺字就是索引或字型檔的問題，不是「正常的 fallback」。
func TestCJKSampleHasNoMissingGlyphs(t *testing.T) {
	f := openETen(t)
	h := NewHiFrame()
	for i, b := range demoLines {
		col := 1
		for j := 0; j+1 < len(b); j += 2 {
			if !h.DrawCJK(f, b[j], b[j+1], col, 18+i, 15) {
				t.Errorf("第 %d 行第 %d 個字（%02X%02X）缺字", i+1, col, b[j], b[j+1])
			}
			col++
		}
		// 訊息視窗是欄 1–38（docs/re/25），超過就是排版爆了。
		if col > 39 {
			t.Errorf("第 %d 行用了 %d 格，超出訊息視窗的 38 格", i+1, col-1)
		}
	}

	if out := os.Getenv("WL_CJK_SAMPLE"); out != "" {
		fh, err := os.Create(out)
		if err != nil {
			t.Fatal(err)
		}
		defer fh.Close()
		if err := png.Encode(fh, h.ToImage()); err != nil {
			t.Fatal(err)
		}
		t.Logf("已寫出樣張 %s", out)
	}
}
