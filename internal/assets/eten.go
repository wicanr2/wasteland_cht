package assets

import (
	"fmt"
	"os"
	"path/filepath"
)

// 倚天中文系統 3.53 的 16 × 15 點陣字（docs/spec/10 §4）。
//
// **字型檔玩家自備、不隨專案散布**（CLAUDE.md §7）。
// 載入失敗時遊戲要能只用英文跑，不得直接崩。

const (
	ETenWidth   = 16
	ETenHeight  = 15
	etenStride  = 30 // 每列 2 bytes × 15 列
	etenCommonN = 5401
)

// ETenFont 是解開的倚天字型。
type ETenFont struct {
	std []byte // STDFONT.15，漢字
	spc []byte // SPCFONT.15，全形標點與符號
}

// LoadETen 從一個目錄載入 STDFONT.15 與 SPCFONT.15（檔名大小寫都試）。
func LoadETen(dir string) (*ETenFont, error) {
	std, err := readAnyCase(dir, "STDFONT.15")
	if err != nil {
		return nil, err
	}
	// ⚠ 只帶 STDFONT 會讓所有全形標點掉 fallback——它從「一」開始，
	// 不含 A140–A3BF 的標點（docs/spec/10 §4）。
	spc, err := readAnyCase(dir, "SPCFONT.15")
	if err != nil {
		return nil, fmt.Errorf("找不到 SPCFONT.15：沒有它的話全形標點會全部缺字：%w", err)
	}
	return &ETenFont{std: std, spc: spc}, nil
}

func readAnyCase(dir, name string) ([]byte, error) {
	for _, n := range []string{name, lower(name)} {
		if b, err := os.ReadFile(filepath.Join(dir, n)); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("%s 裡沒有 %s", dir, name)
}

func lower(s string) string {
	out := []byte(s)
	for i, c := range out {
		if c >= 'A' && c <= 'Z' {
			out[i] = c + 32
		}
	}
	return string(out)
}

// etenRaw 是 Big5 的線性序號（kb `retro-cht/eten-bitmap-font` 的公式）。
func etenRaw(hi, lo byte) int {
	off := int(lo) - 0x40
	if lo >= 0x7F {
		off = int(lo) - 0x62
	}
	return (int(hi)-0xA1)*157 + off
}

var (
	etenLastSPC    = etenRaw(0xA3, 0xBF) // 符號區尾
	etenBaseA440   = etenRaw(0xA4, 0x40) // 漢字區起點
	etenLastCommon = etenRaw(0xC6, 0x7E) // 常用字尾
	etenBaseC940   = etenRaw(0xC9, 0x40) // 次常用起點
)

// Glyph 回傳一個 Big5 字的點陣（30 bytes，每列 2 bytes、MSB-first）。
// 查不到就回 nil——呼叫者要自己決定 fallback，不要靜靜畫成空白。
func (f *ETenFont) Glyph(hi, lo byte) []byte {
	if f == nil || hi < 0xA1 {
		return nil
	}
	r := etenRaw(hi, lo)
	var src []byte
	var idx int
	switch {
	case r < 0:
		return nil
	case r <= etenLastSPC:
		src, idx = f.spc, r
	case r <= etenLastCommon:
		src, idx = f.std, r-etenBaseA440
	default:
		src, idx = f.std, etenCommonN+(r-etenBaseC940)
	}
	at := idx * etenStride
	if idx < 0 || at+etenStride > len(src) {
		return nil
	}
	return src[at : at+etenStride]
}

// GlyphRows 把點陣攤成 15 列、每列 16 個 bool（畫起來比較直觀）。
func (f *ETenFont) GlyphRows(hi, lo byte) [][]bool {
	g := f.Glyph(hi, lo)
	if g == nil {
		return nil
	}
	rows := make([][]bool, ETenHeight)
	for y := 0; y < ETenHeight; y++ {
		row := make([]bool, ETenWidth)
		hiByte, loByte := g[y*2], g[y*2+1]
		for x := 0; x < 8; x++ {
			row[x] = hiByte&(0x80>>x) != 0
			row[8+x] = loByte&(0x80>>x) != 0
		}
		rows[y] = row
	}
	return rows
}
