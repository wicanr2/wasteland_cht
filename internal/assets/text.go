package assets

import "fmt"

// 5-bit 打包文字（docs/re/17、docs/re/18）。
//
//	表基址 +0x00 … +0x3B   60 bytes 字元對照表（符號 → ASCII，0 ＝ 字串結束）
//	表基址 +0x3C …         16-bit 位移表，每 4 個字串一項，位移相對於 +0x3C
//	位移指到的地方          5-bit 符號流，每個 byte 由低位往高位取
//
// 符號 0x1E ＝ 下一個字元轉大寫；0x1F ＝ escape，再取一個符號後 +0x1E。
// **每張表有自己的字元對照表，不能共用。**

const (
	alphabetSize = 0x3C
	symShift     = 0x1E
	symEscape    = 0x1F
	groupSize    = 4 // 位移表每一項涵蓋四個字串
)

// ExeStringTables 是執行檔內九張字串表的基址（ds: 位移，docs/re/17）。
// 來源是「誰寫 ds:4692h」——那個變數就是目前的字串表基址。
var ExeStringTables = []struct {
	Base int
	Note string
}{
	{0xA703, "開場字幕與製作名單"},
	{0xAB3E, "無線電、隊伍、戰鬥"},
	{0xB270, "技能、物品、介面"},
	{0xCE4B, "角色建立"},
	{0xD18E, "結局敘述"},
	{0xD622, "階級"},
	{0xDACC, "技能學習"},
	{0xDBF8, "商店"},
	{0xDCED, "疾病與狀態"},
}

// symbolReader 每個 byte 由最低位往最高位取，5 個位元一個符號。
type symbolReader struct {
	buf []byte
	pos int
	cur byte
	n   int
}

func (s *symbolReader) symbol() (int, error) {
	v := 0
	for bit := 0; bit < 5; bit++ {
		if s.n == 0 {
			if s.pos >= len(s.buf) {
				return 0, fmt.Errorf("符號流在第 %d byte 用完", s.pos)
			}
			s.cur = s.buf[s.pos]
			s.pos++
			s.n = 8
		}
		v |= int(s.cur&1) << bit
		s.cur >>= 1
		s.n--
	}
	return v, nil
}

// decodeRun 從 pos 解 count 個字串。alphaBase 是字元對照表的絕對位置。
//
// 原版沒有邊界檢查——escape 之後的符號最大到 0x3D，會讀到 60 bytes 之外。
// 這裡照原版讀（要與原版逐字相同），只在超出緩衝區時回錯。
func decodeRun(buf []byte, alphaBase, pos, count int) ([]string, error) {
	r := &symbolReader{buf: buf, pos: pos}
	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		var chars []byte
		upper := false
		for {
			sym, err := r.symbol()
			if err != nil {
				return out, err
			}
			if sym == symShift {
				upper = true
				continue
			}
			if sym == symEscape {
				next, err := r.symbol()
				if err != nil {
					return out, err
				}
				sym = next + symShift
			}
			at := alphaBase + sym
			if at >= len(buf) {
				return out, fmt.Errorf("字元對照表索引 %d 超出緩衝區", sym)
			}
			code := buf[at]
			if code == 0 {
				break
			}
			c := code
			if upper && c >= 'a' && c <= 'z' {
				c -= 'a' - 'A'
			}
			upper = false
			chars = append(chars, c)
		}
		out = append(out, string(chars))
	}
	return out, nil
}

// decodeTable 解一張完整的字串表。
//
// 位移表沒有長度欄位，但它自己的第一項就是第一個字串的位置，所以
// 「第一項 ÷ 2」就是項數。最後一項通常是沒用到的填充、值落在區塊外——
// 遇到不遞增或跨距離譜就停，**不要靜靜截掉**（呼叫端會拿到解到哪為止）。
func decodeTable(buf []byte, base int) ([]string, error) {
	if base+alphabetSize+2 > len(buf) {
		return nil, fmt.Errorf("字串表基址 %#x 超出緩衝區（%d bytes）", base, len(buf))
	}
	data := base + alphabetSize
	declared := int(le16(buf, data)) / 2

	var offsets []int
	for i := 0; i < declared; i++ {
		at := data + i*2
		if at+2 > len(buf) {
			break
		}
		off := int(le16(buf, at))
		// 一組只有四個字串，跨距不可能到 1 KB；差太多就是讀到表尾的填充了。
		if len(offsets) > 0 && (off <= offsets[len(offsets)-1] || off-offsets[len(offsets)-1] > 0x400) {
			break
		}
		offsets = append(offsets, off)
	}

	var out []string
	for _, off := range offsets {
		part, err := decodeRun(buf, base, data+off, groupSize)
		if err != nil {
			out = append(out, part...) // 解到哪算哪，讓呼叫端看得到
			break
		}
		out = append(out, part...)
	}

	return out, nil
}

// ExeStrings 解出執行檔內九張表。回傳的順序與 ExeStringTables 相同。
func (r *Rom) ExeStrings() ([][]string, error) {
	if r.image == nil {
		return nil, fmt.Errorf("還沒載入分析映像（先呼叫 LoadImage）")
	}
	out := make([][]string, 0, len(ExeStringTables))
	for _, t := range ExeStringTables {
		at, err := r.dsOffset(t.Base)
		if err != nil {
			return nil, fmt.Errorf("字串表 ds:%04Xh（%s）：%w", t.Base, t.Note, err)
		}
		s, err := decodeTable(r.image, at)
		if err != nil {
			return nil, fmt.Errorf("字串表 ds:%04Xh（%s）：%w", t.Base, t.Note, err)
		}
		out = append(out, s)
	}
	return out, nil
}
