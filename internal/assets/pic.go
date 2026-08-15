package assets

import (
	"fmt"
	"image"
	"image/color"
)

// 圖片與圖磚（docs/re/23、docs/re/24）。
//
//	packed 4bpp：一個 byte 兩個像素，高 4 位在左
//	列間 XOR delta：out[n+stride] ^= out[n]，以 word 為單位、n 由 0 每次 +2
//	**回看距離 stride 就是一列的 byte 數**
//
// 圖片與圖磚是同一套格式，只有尺寸不同；圖磚每張各自帶 8 bytes 種子列。

// 三種尺寸（docs/spec/01 §2.8）。
const (
	picStride, picWidth, picHeight       = 48, 96, 84
	titleStride, titleWidth, titleHeight = 144, 288, 128
	tileStride, tileWidth, tileHeight    = 8, 16, 16
	tileBytes                            = tileStride * tileHeight // 128
	picBytes                             = picStride * picHeight   // 4032
)

// EGAPalette 是 mode 0Dh 的標準 16 色。
//
// ⚠ **暫代**：原版設定調色盤的程式碼還沒找到（盤點 A14），只知道圖片用滿 16 色。
// 對原版驗收時比索引值，不要比 RGB。
var EGAPalette = [16]color.RGBA{
	{0x00, 0x00, 0x00, 0xFF}, {0x00, 0x00, 0xAA, 0xFF},
	{0x00, 0xAA, 0x00, 0xFF}, {0x00, 0xAA, 0xAA, 0xFF},
	{0xAA, 0x00, 0x00, 0xFF}, {0xAA, 0x00, 0xAA, 0xFF},
	{0xAA, 0x55, 0x00, 0xFF}, {0xAA, 0xAA, 0xAA, 0xFF},
	{0x55, 0x55, 0x55, 0xFF}, {0x55, 0x55, 0xFF, 0xFF},
	{0x55, 0xFF, 0x55, 0xFF}, {0x55, 0xFF, 0xFF, 0xFF},
	{0xFF, 0x55, 0x55, 0xFF}, {0xFF, 0x55, 0xFF, 0xFF},
	{0xFF, 0xFF, 0x55, 0xFF}, {0xFF, 0xFF, 0xFF, 0xFF},
}

// undelta 就地解列間 delta。順序不能顛倒：n ≥ stride 之後讀到的是已經解過的內容。
func undelta(buf []byte, stride int) []byte {
	out := make([]byte, len(buf))
	copy(out, buf)
	for n := 0; n+stride+1 < len(out); n += 2 {
		out[n+stride] ^= out[n]
		out[n+stride+1] ^= out[n+1]
	}
	return out
}

// Indexed 是解碼後的圖，一格一個 0–15 的顏色索引。
//
// 調色盤未定案，所以解碼層回傳索引而不是 RGBA——驗收要比的是索引
// （docs/spec/01 §4）。要畫出來時再呼叫 RGBA。
type Indexed struct {
	Width, Height int
	Pix           []byte // len == Width*Height，值 0–15
}

// RGBA 用 EGAPalette 上色。**顏色是暫代的**，見 EGAPalette 的說明。
func (im *Indexed) RGBA() *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, im.Width, im.Height))
	for y := 0; y < im.Height; y++ {
		for x := 0; x < im.Width; x++ {
			out.SetRGBA(x, y, EGAPalette[im.Pix[y*im.Width+x]&0x0F])
		}
	}
	return out
}

// unpack4bpp 把 packed 4bpp 展開成一格一個索引。
func unpack4bpp(buf []byte, width, height, stride int) (*Indexed, error) {
	if len(buf) < stride*height {
		return nil, fmt.Errorf("資料 %d bytes 不足 %d × %d（stride %d）", len(buf), width, height, stride)
	}
	pix := make([]byte, width*height)
	for y := 0; y < height; y++ {
		row := buf[y*stride:]
		for x := 0; x < width; x++ {
			b := row[x>>1]
			if x&1 == 0 {
				pix[y*width+x] = b >> 4
			} else {
				pix[y*width+x] = b & 0x0F
			}
		}
	}
	return &Indexed{Width: width, Height: height, Pix: pix}, nil
}

// Title 解 TITLE.PIC（288 × 128）。
func (r *Rom) Title() (*Indexed, error) {
	data, err := r.File("title.pic")
	if err != nil {
		return nil, err
	}
	if len(data) != titleStride*titleHeight {
		return nil, fmt.Errorf("title.pic 長度 %d 與 %d × %d 不符",
			len(data), titleWidth, titleHeight)
	}
	return unpack4bpp(undelta(data, titleStride), titleWidth, titleHeight, titleStride)
}

// Pictures 解一個 ALLPICS 檔裡的所有圖片（96 × 84）。
//
// 子區塊嚴格交錯：一張圖 ＋ 一段變動長度的參數區。參數區未解，這裡跳過——
// 判斷方式是解壓後長度剛好 4,032，不是靠位置推。
func (r *Rom) Pictures(name string) ([]*Indexed, error) {
	data, err := r.File(name)
	if err != nil {
		return nil, err
	}
	blocks, err := SplitAll(data)
	if err != nil {
		return nil, fmt.Errorf("%s：%w", name, err)
	}
	var out []*Indexed
	for i, b := range blocks {
		if b.Size != picBytes {
			continue // 參數區
		}
		raw, err := DecompressAt(data, blocks, i)
		if err != nil {
			return nil, fmt.Errorf("%s 第 %d 個子區塊：%w", name, i, err)
		}
		im, err := unpack4bpp(undelta(raw, picStride), picWidth, picHeight, picStride)
		if err != nil {
			return nil, fmt.Errorf("%s 第 %d 張圖：%w", name, i, err)
		}
		out = append(out, im)
	}
	return out, nil
}

// Tileset 解一組圖磚（16 × 16，一張 128 bytes）。
//
// n 是跨兩個檔案的組編號 0–8（地圖記錄區標頭 +0x30 的值）：
// 0–3 在 ALLHTDS1、4–8 在 ALLHTDS2。
func (r *Rom) Tileset(n int) ([]*Indexed, error) {
	name := "allhtds1"
	index := n
	if n >= 4 {
		name, index = "allhtds2", n-4
	}
	data, err := r.File(name)
	if err != nil {
		return nil, err
	}
	blocks, err := SplitAll(data)
	if err != nil {
		return nil, fmt.Errorf("%s：%w", name, err)
	}
	if index < 0 || index >= len(blocks) {
		return nil, fmt.Errorf("圖磚組 %d 超出範圍（%s 有 %d 組）", n, name, len(blocks))
	}
	raw, err := DecompressAt(data, blocks, index)
	if err != nil {
		return nil, fmt.Errorf("圖磚組 %d：%w", n, err)
	}
	if len(raw)%tileBytes != 0 {
		return nil, fmt.Errorf("圖磚組 %d 解出 %d bytes，不是 %d 的倍數", n, len(raw), tileBytes)
	}
	out := make([]*Indexed, 0, len(raw)/tileBytes)
	for i := 0; i+tileBytes <= len(raw); i += tileBytes {
		// delta 不跨圖磚：每張各自有 8 bytes 種子列。
		im, err := unpack4bpp(undelta(raw[i:i+tileBytes], tileStride), tileWidth, tileHeight, tileStride)
		if err != nil {
			return nil, fmt.Errorf("圖磚組 %d 第 %d 張：%w", n, i/tileBytes, err)
		}
		out = append(out, im)
	}
	return out, nil
}

// Icons 解 IC0_9.WLF 的十個 16 × 16 疊圖。
//
// 檔案裡是 EGA 4 平面（128 bytes 一張 ＝ 4 × 32），與圖磚在原版記憶體裡的形式相同；
// 地圖第 3 層的圖形編號 0–9 指的就是這十張（docs/re/24 §2.3）。
func (r *Rom) Icons() ([]*Indexed, error) {
	data, err := r.File("ic0_9.wlf")
	if err != nil {
		return nil, err
	}
	out := make([]*Indexed, 0, len(data)/tileBytes)
	for i := 0; i+tileBytes <= len(data); i += tileBytes {
		out = append(out, planarToIndexed(data[i:i+tileBytes], tileWidth, tileHeight))
	}
	return out, nil
}

// planarToIndexed 把 EGA 4 平面（平面連續存放）展成索引圖。
// 一個像素的顏色是四個平面同位置的位元，plane0 是 bit0。
func planarToIndexed(raw []byte, width, height int) *Indexed {
	rowBytes := width / 8
	plane := rowBytes * height
	pix := make([]byte, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var v byte
			for p := 0; p < 4; p++ {
				b := raw[p*plane+y*rowBytes+x/8]
				v |= ((b >> (7 - uint(x%8))) & 1) << p
			}
			pix[y*width+x] = v
		}
	}
	return &Indexed{Width: width, Height: height, Pix: pix}
}

// Masks 解 MASKS.WLF 的十個 16 × 16 遮罩。
//
// 一張 32 bytes ＝ 16 列 × 2 bytes，**一個位元一個像素**（不是 4 平面）。
// 320 ÷ 32 ＝ 10，與 IC0_9.WLF 的十張疊圖一一對應（docs/re/24 §2.3）。
//
// 合成規則是 `螢幕 ← (背景 AND 遮罩) OR 疊圖`：位元 1 ＝ 保留背景、
// 0 ＝ 清掉再讓疊圖蓋上去。
func (r *Rom) Masks() ([][]bool, error) {
	data, err := r.File("masks.wlf")
	if err != nil {
		return nil, err
	}
	const maskBytes = tileHeight * (tileWidth / 8) // 16 × 2
	out := make([][]bool, 0, len(data)/maskBytes)
	for i := 0; i+maskBytes <= len(data); i += maskBytes {
		m := make([]bool, tileWidth*tileHeight)
		for y := 0; y < tileHeight; y++ {
			for x := 0; x < tileWidth; x++ {
				b := data[i+y*(tileWidth/8)+x/8]
				m[y*tileWidth+x] = (b>>(7-uint(x%8)))&1 == 1
			}
		}
		out = append(out, m)
	}
	return out, nil
}
