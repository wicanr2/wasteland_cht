package assets

import "fmt"

// 地圖編號的兩張表（`ds:BF1Ch`，256 bytes，docs/re/61）。
//
// 同一張表被 `sub_1841F` 讀兩次，語意隨編號的 bit7 而不同：
//
//	bit7 沒設（0–127）：值 ＝ 0x40 → 記錄區標頭在 0x1800，否則 0x600
//	bit7 設（128–255）：值 ＝ **真正的資源編號**
//
// 建築內部就是靠後者：Quartz 的每一棟建築都用 128–191 的編號進去，
// 而它們全部映射到同一張內部地圖（`docs/re/61` §2）。

const tblMapID = 0xBF1C // ds: 位移；sub_1841F 的 `[bx-40E4h]`

// MapHeaderAt 回報這個地圖編號的記錄區標頭位置（0x600 或 0x1800）。
//
// 只對 bit7 沒設的編號有意義——bit7 設的編號要先用 ResolveMapID 換過。
func (r *Rom) MapHeaderAt(id byte) (int, error) {
	v, err := r.mapIDEntry(id)
	if err != nil {
		return 0, err
	}
	if v == 0x40 {
		return 0x1800, nil
	}
	return 0x600, nil
}

// ResolveMapID 把地圖編號換成真正的資源編號。
//
// bit7 沒設就是它自己；設起來的要查表——**忘了查會拿 130 這種值去索引
// 42 個區塊**，症狀是「進建築就爆掉」而不是「進錯地方」。
func (r *Rom) ResolveMapID(id byte) (byte, error) {
	if id&0x80 == 0 {
		return id, nil
	}
	v, err := r.mapIDEntry(id)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (r *Rom) mapIDEntry(id byte) (byte, error) {
	off, err := r.dsOffset(tblMapID + int(id))
	if err != nil {
		return 0, fmt.Errorf("地圖編號表 %d：%w", id, err)
	}
	return r.image[off], nil
}
