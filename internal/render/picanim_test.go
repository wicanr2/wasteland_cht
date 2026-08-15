package render

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 規格 26 §6 驗收 1：把動畫播到第 3 格，與 Ranger Center 的實機截圖
// 在 (8, 8) 起的 96 × 84 區域**逐像素相同**。
//
// 驗收要對原版，不是對自己的狀態機（`CLAUDE.md` §4）——所以這裡比的是
// 截圖的像素，不是「PicPlayer 有沒有吐出格 3」。
func TestPicAnimMatchesHardwareShot(t *testing.T) {
	const shot = "../../workplace/dosbox/shots/db05.ppm"
	if _, err := os.Stat(shot); err != nil {
		t.Skipf("找不到實機截圖 %s，跳過。產生方式見 docs/re/54 §4", shot)
	}
	rom := openRom(t)
	pics, err := rom.Pictures("allpics1")
	if err != nil {
		t.Fatalf("Pictures：%v", err)
	}
	anims, err := rom.PictureAnims("allpics1")
	if err != nil {
		t.Fatalf("PictureAnims：%v", err)
	}
	const facilityPic = 3 // ds:A4E0h 表裡的存檔／管理設施（docs/re/54 §1）

	want, err := readPPM(shot)
	if err != nil {
		t.Fatalf("讀截圖：%v", err)
	}

	f := NewFrame()
	f.DrawIndexed(pics[facilityPic], FacilityPicX, FacilityPicY, ViewClip())
	if diff := comparePic(f, want); diff == 0 {
		t.Fatal("底圖就已經吻合了——那截圖裡沒有動畫，這個測試證明不了東西")
	}

	p := NewPicPlayer(anims[facilityPic])
	best, bestTick := -1, -1
	for tick := 0; tick < 64; tick++ {
		f.ApplyAnim(p.Tick(), FacilityPicX, FacilityPicY)
		d := comparePic(f, want)
		if best < 0 || d < best {
			best, bestTick = d, tick
		}
		if d == 0 {
			t.Logf("第 %d 拍與實機截圖逐像素相同", tick)
			return
		}
	}
	t.Fatalf("64 拍之內沒有一拍對上實機截圖，最好的是第 %d 拍差 %d 個像素",
		bestTick, best)
}

// comparePic 回報畫面上設施圖那一塊與截圖差幾個像素。
func comparePic(f *Frame, want []byte) int {
	diff := 0
	for y := 0; y < FacilityPicHeight; y++ {
		for x := 0; x < FacilityPicWidth; x++ {
			sx, sy := FacilityPicX+x, FacilityPicY+y
			o := (sy*ScreenWidth + sx) * 3
			if toIndex(want[o], want[o+1], want[o+2]) != f.At(sx, sy) {
				diff++
			}
		}
	}
	return diff
}

// toIndex 把 EGA 的 RGB 換回索引；認不得就回 0xFF，讓它一定不吻合。
func toIndex(r, g, b byte) byte {
	for i, c := range assets.EGAPalette {
		if c.R == r && c.G == g && c.B == b {
			return byte(i)
		}
	}
	return 0xFF
}

// readPPM 讀 P6 的 320 × 200 截圖，回傳 RGB bytes。
func readPPM(path string) ([]byte, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	r := bufio.NewReader(fh)

	var magic string
	var w, h, max int
	if _, err := fmt.Fscan(r, &magic, &w, &h, &max); err != nil {
		return nil, fmt.Errorf("讀標頭：%w", err)
	}
	if magic != "P6" || max != 255 {
		return nil, fmt.Errorf("只支援 8-bit 的 P6，這份是 %s／max %d", magic, max)
	}
	if w != ScreenWidth || h != ScreenHeight {
		return nil, fmt.Errorf("截圖是 %d × %d，畫面是 %d × %d", w, h, ScreenWidth, ScreenHeight)
	}
	if _, err := r.ReadByte(); err != nil { // 標頭後那一個空白
		return nil, err
	}
	buf := make([]byte, w*h*3)
	for i := 0; i < len(buf); {
		n, err := r.Read(buf[i:])
		if err != nil {
			return nil, fmt.Errorf("讀像素（讀到第 %d byte）：%w", i, err)
		}
		i += n
	}
	return buf, nil
}
