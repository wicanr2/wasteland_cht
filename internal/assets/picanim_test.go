package assets

import "testing"

// 規格 26 §6 驗收 2：一輪播完 XOR 互相抵消，畫面回到底圖。
//
// 這是**全檔級的自洽檢查**——相位、位置、長度、播放順序任何一項讀錯，
// 82 張圖不可能都對。比單張截圖強得多，而且不需要實機。
func TestPicAnimCycleIsIdentity(t *testing.T) {
	rom := openRom(t)
	total, animated := 0, 0
	for _, name := range []string{"allpics1", "allpics2"} {
		pics, err := rom.Pictures(name)
		if err != nil {
			t.Fatalf("%s 的圖：%v", name, err)
		}
		anims, err := rom.PictureAnims(name)
		if err != nil {
			t.Fatalf("%s 的動畫：%v", name, err)
		}
		if len(pics) != len(anims) {
			t.Fatalf("%s：%d 張圖但 %d 份動畫，索引對不起來", name, len(pics), len(anims))
		}
		for i, a := range anims {
			total++
			if a.Empty() {
				continue
			}
			animated++
			buf := make([]byte, len(pics[i].Pix))
			copy(buf, pics[i].Pix)
			for _, ch := range a.Channels {
				for _, fi := range ch.Frame {
					if int(fi) >= len(a.Frames) {
						t.Fatalf("%s 第 %d 張：腳本指到不存在的格 %d", name, i, fi)
					}
					xorInto(buf, pics[i].Width, a.Frames[fi])
				}
			}
			for k := range buf {
				if buf[k] != pics[i].Pix[k] {
					t.Fatalf("%s 第 %d 張：一輪播完沒有回到底圖（第 %d 個像素 %d ≠ %d）",
						name, i, k, buf[k], pics[i].Pix[k])
				}
			}
		}
	}
	if animated == 0 {
		t.Fatal("一張有動畫的圖都沒有——解析大概整個失敗了")
	}
	t.Logf("%d 張圖，其中 %d 張有動畫，全部通過循環恆等式", total, animated)
}

func xorInto(buf []byte, width int, frame []AnimElem) {
	for _, e := range frame {
		for i, v := range e.Pixels {
			if x := e.X + i; x >= 0 && x < width {
				buf[e.Y*width+x] ^= v
			}
		}
	}
}

// 規格 26 §6 驗收 3：相位非 0 的元素起點是「欄 × 8 ＋ 2 × 相位」。
//
// 相位少乘 2 會讓整段左移，循環恆等式**照樣成立**（XOR 兩次還是抵消），
// 所以這一條要單獨驗——恆等式抓不到平移。
func TestPicAnimPhaseShiftsByPairs(t *testing.T) {
	rom := openRom(t)
	anims, err := rom.PictureAnims("allpics1")
	if err != nil {
		t.Fatalf("PictureAnims：%v", err)
	}
	odd := 0
	for _, a := range anims {
		for _, frame := range a.Frames {
			for _, e := range frame {
				if e.X%8 != 0 {
					odd++
					if e.X%2 != 0 {
						t.Fatalf("起點 %d 不是偶數——相位是「幾對像素」不是「幾個像素」", e.X)
					}
				}
			}
		}
	}
	if odd == 0 {
		t.Fatal("沒有任何相位非 0 的元素——相位大概沒被算進去")
	}
	t.Logf("%d 個元素的起點帶相位偏移", odd)
}

// 規格 26 §6 驗收 4：沒有參數區不是錯誤。
func TestDecodePicAnimEmpty(t *testing.T) {
	for _, raw := range [][]byte{nil, {}, {0x00}, {0x00, 0x00}} {
		a, err := DecodePicAnim(raw)
		if err != nil {
			t.Fatalf("%v：%v", raw, err)
		}
		if !a.Empty() {
			t.Fatalf("%v 應該是零值", raw)
		}
	}
}
