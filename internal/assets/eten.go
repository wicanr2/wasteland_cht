package assets

import (
	"fmt"
	"os"
	"path/filepath"
)

// 倚天中文系統 3.53 的點陣字（docs/spec/10 §4）。
//
// **字型檔玩家自備、不隨專案散布**（CLAUDE.md §7）。
// 載入失敗時遊戲要能只用英文跑，不得直接崩。
//
// 支援兩種字級，優先用 24 點：
//
//	24 點  STDFONT.24（漢字 24×24）＋ SPCFONT.24（全形符號）＋ ASCFONT.24（半形 16×24）
//	15 點  STDFONT.15（漢字 16×15）＋ SPCFONT.15（全形符號）＋ ASCFONT.15（半形 8×15）
//
// ⚠ `STDFONT.24` **不是光碟上的原始檔**：光碟給的是 ETUNPACK 壓縮的
// `STD.24M`（明體）等六種字體，要先用 `tools/etunpack.py` 解開。
//
// ⚠ **英數不要拿遊戲原版的 8×8 字模放大。** 倚天自己就附了同高度的半形
// ASCII（`ASCFONT.*`），用它英數才會與中文同一套設計、同一個粗細；
// 拿 8×8 放大出來的字筆劃是兩倍粗，擺在中文旁邊像另一種字型。

const (
	etenCommonN = 5401 // 常用字區的字數（兩種字級一樣）
)

// ETenFont 是解開的倚天字型。
type ETenFont struct {
	std []byte // 漢字區
	spc []byte // 全形符號與標點
	asc []byte // 半形 ASCII（可能沒有）

	// Width／Height 是漢字的字模尺寸（16×15 或 24×24）。
	Width, Height int
	stride        int

	// ASCIIWidth／ASCIIHeight 是半形字模的尺寸（8×15 或 16×24）。
	ASCIIWidth, ASCIIHeight int
	ascStride               int
}

// etenSet 是一個字級的檔名與尺寸。
type etenSet struct {
	std, spc, asc          string
	w, h, stride           int
	ascW, ascH, ascStride  int
}

// 24 點排在前面：有 24 點就用 24 點（`docs/spec/10` §4 的字級取捨）。
var etenSets = []etenSet{
	{
		std: "STDFONT.24", spc: "SPCFONT.24", asc: "ASCFONT.24",
		w: 24, h: 24, stride: 72,
		ascW: 16, ascH: 24, ascStride: 48,
	},
	{
		std: "STDFONT.15", spc: "SPCFONT.15", asc: "ASCFONT.15",
		w: 16, h: 15, stride: 30,
		ascW: 8, ascH: 15, ascStride: 15,
	},
}

// LoadETen 從一個目錄載入倚天字型（檔名大小寫都試）。
//
// 24 點與 15 點都找不到才算失敗。半形 ASCII 找不到**不算失敗**——
// 那時英數退回遊戲原版的 8×8 字模，只是粗細對不上中文。
func LoadETen(dir string) (*ETenFont, error) {
	var firstErr error
	for _, s := range etenSets {
		std, err := readAnyCase(dir, s.std)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		// ⚠ 只帶 STDFONT 會讓所有全形標點掉 fallback——它從「一」開始，
		// 不含 A140–A3BF 的標點（docs/spec/10 §4）。
		spc, err := readAnyCase(dir, s.spc)
		if err != nil {
			return nil, fmt.Errorf("有 %s 卻沒有 %s：沒有它全形標點會全部缺字：%w",
				s.std, s.spc, err)
		}
		f := &ETenFont{
			std: std, spc: spc,
			Width: s.w, Height: s.h, stride: s.stride,
		}
		if asc, err := readAnyCase(dir, s.asc); err == nil {
			f.asc = asc
			f.ASCIIWidth, f.ASCIIHeight, f.ascStride = s.ascW, s.ascH, s.ascStride
		}
		return f, nil
	}
	return nil, fmt.Errorf("%s 裡沒有倚天字型（找 STDFONT.24 或 STDFONT.15）：%w",
		dir, firstErr)
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

// Glyph 回傳一個 Big5 字的點陣（每列 (W+7)/8 bytes、MSB-first）。
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
	at := idx * f.stride
	if idx < 0 || at+f.stride > len(src) {
		return nil
	}
	return src[at : at+f.stride]
}

// rowsOf 把點陣攤成 h 列、每列 w 個 bool。
//
// ⚠ **這一支每叫一次配置 h+1 個物件**（外層一個、每列一個），
// 所以**畫圖的熱路徑不要走它**——一個 24 點的字就是 25 次配置，
// 一幀滿版中文的畫面量到 900 次（`internal/play/framebench_test.go`）。
// 直接讀點陣的 `Glyph`／`ASCIIGlyph` 不配置任何東西，
// 攤成 `[][]bool` 只是為了讓測試好讀。
func rowsOf(g []byte, w, h int) [][]bool {
	rowBytes := (w + 7) / 8
	rows := make([][]bool, h)
	for y := 0; y < h; y++ {
		row := make([]bool, w)
		for x := 0; x < w; x++ {
			b := g[y*rowBytes+x/8]
			row[x] = b&(0x80>>(x%8)) != 0
		}
		rows[y] = row
	}
	return rows
}

// GlyphRows 把漢字的點陣攤成 Height 列、每列 Width 個 bool。
func (f *ETenFont) GlyphRows(hi, lo byte) [][]bool {
	g := f.Glyph(hi, lo)
	if g == nil {
		return nil
	}
	return rowsOf(g, f.Width, f.Height)
}

// HasASCII 回報這份字型帶不帶半形 ASCII。
func (f *ETenFont) HasASCII() bool { return f != nil && f.asc != nil }

// ASCIIGlyph 回傳一個半形字元的點陣（每列 (ASCIIWidth+7)/8 bytes、MSB-first）。
// 這是 `ASCIIRows` 的免配置版，畫圖走這一支。
func (f *ETenFont) ASCIIGlyph(c byte) []byte {
	if !f.HasASCII() {
		return nil
	}
	at := int(c) * f.ascStride
	if at+f.ascStride > len(f.asc) {
		return nil
	}
	return f.asc[at : at+f.ascStride]
}

// ASCIIRows 把一個 ASCII 字元的半形點陣攤成 ASCIIHeight 列。
//
// 沒有 `ASCFONT.*` 時回 nil，呼叫端要退回遊戲原版的 8×8 字模。
func (f *ETenFont) ASCIIRows(c byte) [][]bool {
	g := f.ASCIIGlyph(c)
	if g == nil {
		return nil
	}
	return rowsOf(g, f.ASCIIWidth, f.ASCIIHeight)
}
