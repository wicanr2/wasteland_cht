package assets

import "fmt"

// `CURS` ＝ 八個滑鼠游標（`docs/re/57`）。
//
// 版面：一個圖形 32 × 16 的 EGA 4 平面、**平面連續**（不是逐列交錯），
// 平面大小 4 bytes × 16 列 ＝ 64，一個圖形 4 × 64 ＝ 256；2,048 ÷ 256 ＝ 8 個。
//
// 一個 32 寬的圖形其實是**兩張 16 × 16 並排**：左半是遮罩（值域只有 0 與 15）、
// 右半才是圖形。合成照疊圖那一套：`螢幕 ← (背景 AND 遮罩) OR 圖形`。
//
// ⚠ **哪個圖形對應哪個狀態沒有解**（`docs/re/57` §4：全檔找不到讀 `CURS` 的
// 程式碼）。重製版固定用第 0 個，不猜（`docs/spec/29` §4）。

const (
	cursWidth  = 32
	cursHeight = 16
	cursBytes  = (cursWidth / 8) * cursHeight * 4 // 256
	// CursorSize 是一個游標的邊長（32 寬的左右各一半）。
	CursorSize = 16
)

// Cursor 是一個滑鼠游標：圖形與遮罩各 16 × 16。
type Cursor struct {
	// Pix 是圖形（索引色，0 ＝ 透明）。
	Pix *Indexed
	// Mask 為 true 的像素表示「這裡要先把背景清掉」。
	//
	// 遮罩比圖形大一圈，才能在任何背景上描出邊——所以**不能拿圖形的非 0
	// 當遮罩用**（8 個裡有 5 個兩者範圍不同，`docs/re/57` §2）。
	Mask []bool
}

// Cursors 解 `CURS` 的八個游標。
func (r *Rom) Cursors() ([]Cursor, error) {
	data, err := r.File("curs")
	if err != nil {
		return nil, err
	}
	if len(data)%cursBytes != 0 {
		return nil, fmt.Errorf("curs 有 %d bytes，不是 %d 的倍數", len(data), cursBytes)
	}
	out := make([]Cursor, 0, len(data)/cursBytes)
	for at := 0; at+cursBytes <= len(data); at += cursBytes {
		full := planarToIndexed(data[at:at+cursBytes], cursWidth, cursHeight)
		c := Cursor{
			Pix:  &Indexed{Width: CursorSize, Height: CursorSize, Pix: make([]byte, CursorSize*CursorSize)},
			Mask: make([]bool, CursorSize*CursorSize),
		}
		for y := 0; y < CursorSize; y++ {
			for x := 0; x < CursorSize; x++ {
				// 左半 ＝ 遮罩、右半 ＝ 圖形。
				c.Mask[y*CursorSize+x] = full.Pix[y*cursWidth+x] != 0
				c.Pix.Pix[y*CursorSize+x] = full.Pix[y*cursWidth+CursorSize+x]
			}
		}
		out = append(out, c)
	}
	return out, nil
}
