package assets

// `END.CPA` 的第二段：結局動畫的腳本（`docs/re/96`）。
//
// `sub_1B7FE` 從 `END.CPA` 解出兩段 Huffman：第一段是 288 × 128 的畫面
// （`Rom.End`），第二段載到 `seg003:0x5120`，由 `sub_1B735` 逐 BIOS tick 驅動。
//
// 版面（`ds:D160h` 從 `0x5122` 起，也就是**跳過開頭那個長度 word**）：
//
// ```
// word  總長度                    ← ds:D160h 指到它後面
// 一格：
//   word 延遲                     ; 計數器每 tick ＋1，等到相等才播這一格
//   元素 × N                      ; word ＞ 0 就是元素
//   word ≤ 0                      ; 這一格結束
// 最後：word ＜ 0                  ; 整份結束 → 跳回記憶點
// ```
//
// 一個元素 6 bytes：`word 螢幕位移` ＋ **4 bytes ＝ 8 個像素**
// （`sub_10FA7` → `sub_10FD3`：一個 byte 兩個像素、高 nibble 在左，
// 再位切片成四個 EGA 平面）。螢幕是 40 bytes 一列，畫到 `0x141 + 位移`——
// `0x141` ＝ 8 × 40 ＋ 1 ＝ 圖片視窗左上角 (8, 8)。
//
// ⚠ **位移的列寬是螢幕的 40，不是這張圖的 36。** 兩個都除得出「看起來合理」
// 的列號，差別要到畫面右緣才看得出來——用 40 算，全部 2,433 個元素都落在
// 這張圖的 288 像素內；用 36 算會有元素跑到圖外（`EndAnimation` 的測試守著）。

import (
	"encoding/binary"
	"fmt"
)

// EndScreenStride 是螢幕一列的 bytes 數（mode 0Dh，320 像素）。
const EndScreenStride = 40

// endLoopAfter 是原版 `ds:D165h` 的初值：**前 11 格只播一次**，
// 第 12 格開始無限循環（`sub_1B735` 的 `0x1B777`）。
const endLoopAfter = 0x0C

// EndElement 是動畫的一筆改寫：從 (X, Y) 起的 8 個像素。
type EndElement struct {
	X, Y   int
	Pixels [8]byte // 調色盤索引 0–15
}

// EndFrame 是動畫的一格。
type EndFrame struct {
	Delay    int // 播這一格之前要等幾個 tick
	Elements []EndElement
}

// EndAnimation 是整份腳本。
type EndAnimation struct {
	Frames []EndFrame
	// LoopFrom 是跑完之後要跳回去的格號（原版的 `ds:D166h`）。
	LoopFrom int
}

// EndAnim 解 `END.CPA` 的第二段。
//
// 第一段用掉多少 bytes 由 Huffman 解碼器自己回報（`HuffInfo.Consumed`），
// **不要用搜尋 magic 的方式找第二段的起點**——第二段的那 4 bytes 不是 `msq`
// （`sub_11AE8(0)` 就是「不驗 magic」的那個呼叫）。
func (r *Rom) EndAnim() (*EndAnimation, error) {
	data, err := r.File("end.cpa")
	if err != nil {
		return nil, err
	}
	_, info, err := Decompress(data, false)
	if err != nil {
		return nil, fmt.Errorf("end.cpa 第一段：%w", err)
	}
	if info.Consumed >= len(data) {
		return nil, fmt.Errorf("end.cpa 第一段就吃完整個檔案（%d bytes）", len(data))
	}
	raw, _, err := Decompress(data[info.Consumed:], false)
	if err != nil {
		return nil, fmt.Errorf("end.cpa 第二段：%w", err)
	}
	return parseEndAnim(raw)
}

func parseEndAnim(raw []byte) (*EndAnimation, error) {
	if len(raw) < 4 {
		return nil, fmt.Errorf("第二段只有 %d bytes", len(raw))
	}
	word := func(p int) int { return int(int16(binary.LittleEndian.Uint16(raw[p:]))) }

	out := &EndAnimation{LoopFrom: -1}
	p := 2 // 跳過開頭的長度 word（原版 ds:D160h ← 0x5122）
	for p+2 <= len(raw) {
		if word(p) < 0 {
			break // 結束標記
		}
		f := EndFrame{Delay: word(p)}
		p += 2
		for p+2 <= len(raw) && word(p) > 0 {
			if p+6 > len(raw) {
				return nil, fmt.Errorf("第 %d 格的元素在位移 %d 被截斷", len(out.Frames), p)
			}
			off := word(p)
			e := EndElement{X: (off % EndScreenStride) * 8, Y: off / EndScreenStride}
			for i, b := range raw[p+2 : p+6] {
				e.Pixels[i*2] = b >> 4
				e.Pixels[i*2+1] = b & 0x0F
			}
			f.Elements = append(f.Elements, e)
			p += 6
		}
		p += 2 // 那個 ≤ 0 的終止字
		out.Frames = append(out.Frames, f)
		if len(out.Frames) == endLoopAfter {
			out.LoopFrom = len(out.Frames) - 1
		}
	}
	if len(out.Frames) == 0 {
		return nil, fmt.Errorf("一格都沒解出來")
	}
	if out.LoopFrom < 0 {
		out.LoopFrom = 0
	}
	return out, nil
}

// Apply 把一格疊到畫面上。畫面是 `Rom.End()` 回傳的那張。
func (f EndFrame) Apply(im *Indexed) {
	if im == nil {
		return
	}
	for _, e := range f.Elements {
		if e.Y < 0 || e.Y >= im.Height {
			continue
		}
		for i, v := range e.Pixels {
			x := e.X + i
			if x < 0 || x >= im.Width {
				continue
			}
			im.Pix[e.Y*im.Width+x] = v
		}
	}
}
