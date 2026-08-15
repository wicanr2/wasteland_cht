package assets

import "testing"

// 結局動畫解得出來，而且**每一個元素都落在這張圖裡**。
//
// 這一條守著位移的列寬：原版的位移是**螢幕的 40 bytes 一列**，不是這張圖的 36。
// 用 36 算出來的列號一樣「看起來合理」，但會有元素跑到 288 像素之外——
// 值域檢查擋不住，邊界檢查才擋得住。
func TestEndAnimation(t *testing.T) {
	rom := openWithImage(t)
	anim, err := rom.EndAnim()
	if err != nil {
		t.Fatalf("EndAnim：%v", err)
	}
	im, err := rom.End()
	if err != nil {
		t.Fatalf("End：%v", err)
	}
	total := 0
	for i, f := range anim.Frames {
		for _, e := range f.Elements {
			total++
			if e.Y < 0 || e.Y >= im.Height || e.X < 0 || e.X+8 > im.Width {
				t.Fatalf("第 %d 格有元素落在圖外：(%d, %d)", i, e.X, e.Y)
			}
		}
	}
	t.Logf("%d 格、%d 個元素，迴圈起點第 %d 格", len(anim.Frames), total, anim.LoopFrom)
	if len(anim.Frames) < 12 {
		t.Fatalf("只解出 %d 格——迴圈起點（第 12 格）都還沒到", len(anim.Frames))
	}
	if total == 0 {
		t.Fatal("一個元素都沒有")
	}

	// 疊完整份之後畫面要真的變過（不是解出一堆空格）。
	before := append([]byte(nil), im.Pix...)
	for _, f := range anim.Frames {
		f.Apply(im)
	}
	diff := 0
	for i := range im.Pix {
		if im.Pix[i] != before[i] {
			diff++
		}
	}
	t.Logf("疊完之後 %d／%d 個像素變了", diff, len(im.Pix))
	if diff == 0 {
		t.Fatal("整份動畫疊完畫面一個像素都沒變")
	}
}
