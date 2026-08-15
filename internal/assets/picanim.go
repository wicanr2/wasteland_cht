package assets

import (
	"encoding/binary"
	"fmt"
)

// 規格 26：ALLPICS 每張圖後面那段參數區是**局部動畫**。
// 表 A 是每個通道的播放腳本、表 B 是每一格的像素（XOR 疊上去）。

const animRowBytes = 12 // 一列 12 個螢幕 byte ＝ 96 像素

// AnimChannel 是一個通道的播放腳本：播 Frame[i] 之前要等 Delay[i] 拍。
type AnimChannel struct {
	Delay []byte
	Frame []byte
}

// AnimElem 是一格裡的一段橫向像素，XOR 到 (X, Y) 起的位置。
// Pixels 一個像素一個 byte（值 0–15），X 已經含了相位偏移。
type AnimElem struct {
	X, Y   int
	Pixels []byte
}

// PicAnim 是一張圖的局部動畫。沒有動畫時 Channels 與 Frames 都是空的。
type PicAnim struct {
	Channels []AnimChannel
	Frames   [][]AnimElem
}

// Empty 回報這張圖有沒有動畫。
func (a PicAnim) Empty() bool { return len(a.Channels) == 0 || len(a.Frames) == 0 }

// DecodePicAnim 解一段已經解壓的參數區。
//
// 版面（規格 26 §2）：
//
//	word n ＋ n bytes 表 A（0xFF 分隔）
//	word m ＋ m bytes 表 B（0xFFFF 分隔）
func DecodePicAnim(raw []byte) (PicAnim, error) {
	var a PicAnim
	if len(raw) < 2 {
		return a, nil
	}
	n := int(binary.LittleEndian.Uint16(raw))
	if 2+n > len(raw) {
		return a, fmt.Errorf("表 A 長度 %d 超出參數區 %d bytes", n, len(raw))
	}

	// 表 A：第一個 byte 是初始倒數，其後 (格, 延遲) 交錯。
	start := 0
	body := raw[2 : 2+n]
	for i, b := range body {
		if b != 0xFF {
			continue
		}
		rec := body[start:i]
		start = i + 1
		if len(rec) < 2 {
			continue
		}
		// 偶數位是延遲、奇數位是格編號——第一個 byte 是初始倒數，
		// 所以整串是「延遲, 格, 延遲, 格, …」而不是「格, 延遲, …」。
		var ch AnimChannel
		for k, v := range rec {
			if k%2 == 0 {
				ch.Delay = append(ch.Delay, v)
			} else {
				ch.Frame = append(ch.Frame, v)
			}
		}
		if len(ch.Frame) > 0 {
			a.Channels = append(a.Channels, ch)
		}
	}

	rest := raw[2+n:]
	if len(rest) < 2 {
		return a, nil
	}
	m := int(binary.LittleEndian.Uint16(rest))
	if 2+m > len(rest) {
		return a, fmt.Errorf("表 B 長度 %d 超出剩餘的 %d bytes", m, len(rest)-2)
	}
	body = rest[2 : 2+m]

	var cur []AnimElem
	for k := 0; k+2 <= len(body); {
		w := binary.LittleEndian.Uint16(body[k:])
		if w == 0xFFFF {
			a.Frames = append(a.Frames, cur)
			cur = nil
			k += 2
			continue
		}
		length := int(w>>12) + 1
		if k+2+length > len(body) {
			return a, fmt.Errorf("表 B 第 %d byte 的酬載 %d bytes 超出範圍", k, length)
		}
		row, col := int((w>>2)&0x3FF)/animRowBytes, int((w>>2)&0x3FF)%animRowBytes
		px := make([]byte, 0, length*2)
		for _, b := range body[k+2 : k+2+length] {
			px = append(px, b>>4, b&0x0F) // 高 nibble 在左
		}
		cur = append(cur, AnimElem{
			X: col*8 + 2*int(w&3), // 相位 ＝ 左邊缺幾對像素
			Y: row,
			Pixels: px,
		})
		k += 2 + length
	}
	if len(cur) > 0 {
		a.Frames = append(a.Frames, cur)
	}
	return a, nil
}

// PictureAnims 解一個 ALLPICS 檔裡每張圖的動畫，順序與 Pictures 相同。
//
// 一張圖後面沒有參數區時給零值，所以兩邊的索引永遠對得起來。
func (r *Rom) PictureAnims(name string) ([]PicAnim, error) {
	data, err := r.File(name)
	if err != nil {
		return nil, err
	}
	blocks, err := SplitAll(data)
	if err != nil {
		return nil, fmt.Errorf("%s：%w", name, err)
	}
	var out []PicAnim
	for i, b := range blocks {
		if b.Size != picBytes {
			continue
		}
		var anim PicAnim
		if i+1 < len(blocks) && blocks[i+1].Size != picBytes {
			raw, err := DecompressAt(data, blocks, i+1)
			if err != nil {
				return nil, fmt.Errorf("%s 第 %d 個子區塊：%w", name, i+1, err)
			}
			if anim, err = DecodePicAnim(raw); err != nil {
				return nil, fmt.Errorf("%s 第 %d 張圖的動畫：%w", name, len(out), err)
			}
		}
		out = append(out, anim)
	}
	return out, nil
}
