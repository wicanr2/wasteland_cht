package assets

import "fmt"

// 兩套字型（docs/re/14）。
//
//	主文字   內嵌在執行檔 seg003:0xCA60，8 bytes 一個字、8×8 單色，索引 ＝ ASCII − 0x20
//	彩色選單 COLORF.FNT，32 bytes 一個字、8×8、EGA 4 平面，平面連續存放
//
// 主文字字型**不是獨立檔案**——這是踩過的坑：把它當成外部檔去找會找不到。

const (
	fontRows      = 8
	monoLinear    = seg003 + 0xCA60
	monoCount     = 128 // sub_1060C 的 and ax, 7Fh
	monoGlyphSize = 8
	colorGlyphSz  = 32
	// ColorWarm／ColorCool 是同形不同色的兩組字模的起點差（docs/re/14 §3）。
	ColorWarm = 0x18
	ColorCool = 0x34
)

// Glyph 是一個 8×8 字模，值是顏色索引（單色字型只有 0 與 1）。
type Glyph struct {
	Pix [fontRows * 8]byte
}

// Font 是一套字型。
type Font struct {
	Glyphs []Glyph
	// FirstASCII 是第 0 個字模對應的 ASCII 碼；單色字型是 0x20，
	// 彩色字型沒有這個概念（是選單詞專用的字模集），值為 0。
	FirstASCII byte
}

// Glyph 取一個字模；超出範圍回錯而不是靜靜回空白。
func (f *Font) Glyph(i int) (*Glyph, error) {
	if i < 0 || i >= len(f.Glyphs) {
		return nil, fmt.Errorf("字模編號 %d 超出範圍（共 %d 個）", i, len(f.Glyphs))
	}
	return &f.Glyphs[i], nil
}

// GlyphForASCII 取 ASCII 對應的字模（只有主文字字型有意義）。
func (f *Font) GlyphForASCII(c byte) (*Glyph, error) {
	if f.FirstASCII == 0 {
		return nil, fmt.Errorf("這套字型不是以 ASCII 索引的")
	}
	return f.Glyph(int(c) - int(f.FirstASCII))
}

// FontMain 從分析映像取出主文字字型（8×8 單色，128 個字模）。
func (r *Rom) FontMain() (*Font, error) {
	at, err := r.fileOffset(monoLinear)
	if err != nil {
		return nil, fmt.Errorf("主文字字型：%w", err)
	}
	need := monoCount * monoGlyphSize
	if at+need > len(r.image) {
		return nil, fmt.Errorf("主文字字型超出映像（要 %d bytes）", need)
	}
	font := &Font{Glyphs: make([]Glyph, monoCount), FirstASCII: 0x20}
	for i := 0; i < monoCount; i++ {
		raw := r.image[at+i*monoGlyphSize : at+(i+1)*monoGlyphSize]
		for y := 0; y < fontRows; y++ {
			for x := 0; x < 8; x++ {
				font.Glyphs[i].Pix[y*8+x] = (raw[y] >> (7 - uint(x))) & 1
			}
		}
	}
	return font, nil
}

// FontColor 讀 COLORF.FNT（8×8、4 平面，172 個字模）。
func (r *Rom) FontColor() (*Font, error) {
	data, err := r.File("colorf.fnt")
	if err != nil {
		return nil, err
	}
	if len(data)%colorGlyphSz != 0 {
		return nil, fmt.Errorf("colorf.fnt 長度 %d 不是 %d 的倍數", len(data), colorGlyphSz)
	}
	count := len(data) / colorGlyphSz
	font := &Font{Glyphs: make([]Glyph, count)}
	for i := 0; i < count; i++ {
		raw := data[i*colorGlyphSz : (i+1)*colorGlyphSz]
		im := planarToIndexed(raw, 8, fontRows)
		copy(font.Glyphs[i].Pix[:], im.Pix)
	}
	return font, nil
}
