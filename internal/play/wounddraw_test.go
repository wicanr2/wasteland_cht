package play

import (
	"os"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// 死亡的骷髏字模在中文畫面上要畫得出**原版那一格**，不是倚天的第 `0x7F` 格。
//
// `0x7F` 在原版主文字字型是骷髏（`docs/re/17` §4.4），在倚天半形字型是別的圖形
// （實測 40 個亮點）。兩邊都畫得出東西，所以「有像素」證明不了走對了路——
// 這一條直接比對兩份字模的亮點數。
func TestDeathGlyphComesFromOriginalFont(t *testing.T) {
	dir := os.Getenv("WL_ETEN")
	if dir == "" {
		dir = "../../workplace/eten"
	}
	eten, err := assets.LoadETen(dir)
	if err != nil {
		t.Skipf("沒有倚天字型（%v），這一條驗不到", err)
	}
	rom := openRom(t)
	font, err := rom.FontMain()
	if err != nil {
		t.Fatalf("讀原版主文字字型：%v", err)
	}

	s := newScene(t)
	if err := s.LoadFont(dir); err != nil {
		t.Skipf("場景載入倚天字型失敗：%v", err)
	}

	// 兩份字模的亮點數：不一樣才分得出畫的是哪一份。
	origOn := 0
	if g, gerr := font.GlyphForASCII(0x7F); gerr == nil {
		for _, p := range g.Pix {
			if p != 0 {
				origOn++
			}
		}
	}
	origOn *= render.HiScale * render.HiScale // 原版字模在高解畫面上放大 3 倍

	etenOn := 0
	for _, r := range eten.ASCIIRows(0x7F) {
		for _, b := range r {
			if b {
				etenOn++
			}
		}
	}
	if origOn == 0 {
		t.Fatal("原版字型第 0x7F 格是空的——那不是骷髏字模")
	}
	if origOn == etenOn {
		t.Skipf("兩份字模的亮點數剛好相同（%d），這一條分不出來", origOn)
	}

	h := render.NewHiFrame()
	s.drawASCII(h, 0x7F, 0, 0)
	on := 0
	for y := 0; y < render.HiCellHeight; y++ {
		for x := 0; x < render.HiCellWidth; x++ {
			if h.At(x, y) != 0 {
				on++
			}
		}
	}
	if on != origOn {
		t.Errorf("畫出來 %d 個亮點：原版字模是 %d、倚天那一格是 %d——畫錯字型了",
			on, origOn, etenOn)
	}
}
