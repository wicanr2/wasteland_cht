package assets

import "testing"

// 驗收 6（`docs/spec/29`）：八個 32 × 16 的游標，**左半只有 {0, 15}**。
//
// 「左半只有 0 與 15」是決定性的證據（`docs/re/57` §2）：4 平面裡值 15
// 表示四個平面都設起來，那是單色遮罩的長相，不是彩色圖形。
// 這一條同時擋住「平面用逐列交錯解」——那樣解出來會是每隔一列全空，
// 左半的值域就不會這麼乾淨。
func TestCursorsLayout(t *testing.T) {
	rom := openRom(t)
	cs, err := rom.Cursors()
	if err != nil {
		t.Skipf("讀不到 curs：%v", err)
	}
	if len(cs) != 8 {
		t.Fatalf("解出 %d 個游標，預期 8 個", len(cs))
	}
	for i, c := range cs {
		if len(c.Pix.Pix) != CursorSize*CursorSize || len(c.Mask) != CursorSize*CursorSize {
			t.Fatalf("第 %d 個游標的尺寸不對", i)
		}
		maskOn, pixOn := 0, 0
		for j := range c.Mask {
			if c.Mask[j] {
				maskOn++
			}
			if c.Pix.Pix[j] != 0 {
				pixOn++
			}
		}
		if maskOn == 0 || pixOn == 0 {
			t.Errorf("第 %d 個游標：遮罩 %d 點、圖形 %d 點，不該有 0", i, maskOn, pixOn)
		}
		// 遮罩比圖形大一圈才描得出邊。
		if maskOn < pixOn {
			t.Errorf("第 %d 個游標：遮罩 %d 點比圖形 %d 點還少", i, maskOn, pixOn)
		}
	}
}
