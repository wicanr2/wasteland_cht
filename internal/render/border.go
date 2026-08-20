package render

// 外框與右邊那一欄（`docs/re/124`、`docs/re/120`）。
//
// 原版的外框、直排的 `WASTELAND` 與輻射計量表都是**彩色字型的字模**
// （`colorf.fnt`，`docs/re/14` §3），由 overlay slot 19 一格一格印出來。
// 這一層照抄字模編號與座標，顏色走字模自己的（`DrawGlyph` 的 color ＝ 0）。

import "github.com/wicanr2/wasteland_cht/internal/assets"

// 畫框用的六個字模（`sub_19814`，`docs/re/124` §1）。
const (
	boxTopLeft     = 0x0E
	boxHorizontal  = 0x12
	boxTopRight    = 0x0F
	boxLeft        = 0x0D
	boxRight       = 0x13
	boxBottomLeft  = 0x10
	boxBottomRight = 0x11

	// 列 17 那兩個接頭：主框的下緣與訊息視窗的上緣在這一列相接
	// （`sub_197BB` 把左下角與右下角蓋掉，`docs/re/124` §2）。
	boxTeeLeft  = 0x5E
	boxTeeRight = 0x5F
)

// 外框的邊界（`docs/re/25` §2.3、`docs/re/124` §2）。
const (
	BorderRight  = 37 // 主框右界
	BorderBottom = 17 // 主框下界，同時是訊息視窗的上緣
	// 訊息視窗那一段寬兩欄：欄 0 與欄 39。
	MsgBorderRight  = 39
	MsgBorderBottom = 23
	// 右邊那一欄（計量表與 RADIATION 標籤）。
	MeterCol = 38
)

// DrawBorder 畫地圖畫面的外框（`sub_197BB`）。
//
// ⚠ **列 24 沒有下緣**：原版的迴圈停在列 24，畫面最底下那一列是指令列本身
// （`docs/re/124` §2）。
func (f *Frame) DrawBorder(font *assets.Font) error {
	if font == nil {
		return nil
	}
	put := func(index, col, row int) error {
		return f.DrawGlyph(font, index, col, row, 0, false)
	}
	// 主框 (0,0)–(37,17)。
	if err := put(boxTopLeft, 0, 0); err != nil {
		return err
	}
	for col := 1; col < BorderRight; col++ {
		if err := put(boxHorizontal, col, 0); err != nil {
			return err
		}
	}
	if err := put(boxTopRight, BorderRight, 0); err != nil {
		return err
	}
	for row := 1; row < BorderBottom; row++ {
		if err := put(boxLeft, 0, row); err != nil {
			return err
		}
		if err := put(boxRight, BorderRight, row); err != nil {
			return err
		}
	}
	if err := put(boxBottomLeft, 0, BorderBottom); err != nil {
		return err
	}
	for col := 1; col < BorderRight; col++ {
		if err := put(boxHorizontal, col, BorderBottom); err != nil {
			return err
		}
	}
	if err := put(boxBottomRight, BorderRight, BorderBottom); err != nil {
		return err
	}
	// 列 17 的兩個接頭 ＋ 右邊多出來的兩欄。
	if err := put(boxTeeLeft, 0, BorderBottom); err != nil {
		return err
	}
	if err := put(boxTeeRight, BorderRight, BorderBottom); err != nil {
		return err
	}
	if err := put(boxHorizontal, MeterCol, BorderBottom); err != nil {
		return err
	}
	if err := put(boxTopRight, MsgBorderRight, BorderBottom); err != nil {
		return err
	}
	// 訊息視窗那一段：欄 0 與欄 39。
	for row := BorderBottom + 1; row <= MsgBorderBottom; row++ {
		if err := put(boxLeft, 0, row); err != nil {
			return err
		}
		if err := put(boxRight, MsgBorderRight, row); err != nil {
			return err
		}
	}
	return nil
}

// titleLabel 是右邊那一欄最上面的直排 `WASTELAND`（`ds:AA4Dh`，`docs/re/124` §4）。
//
// 九列、每列兩個字模，一列一個字母。第 2 列與第 7 列是同一對字模——
// 那正是 `WASTELAND` 的兩個 `A`。
var titleLabel = [9][2]int{
	{0x62, 0x63}, {0x64, 0x65}, {0x66, 0x67},
	{0x68, 0x69}, {0x6A, 0x6B}, {0x6C, 0x6D},
	{0x64, 0x65}, {0x6E, 0x6F}, {0x70, 0x71},
}

// DrawTitleLabel 畫右邊那一欄最上面的直排 `WASTELAND`（`sub_166D3`）。
func (f *Frame) DrawTitleLabel(font *assets.Font) error {
	if font == nil {
		return nil
	}
	for row, pair := range titleLabel {
		for i, g := range pair {
			if err := f.DrawGlyph(font, g, MeterCol+i, row, 0, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// 計量表的字模與列（`sub_17E42`，`docs/re/120` §3）。
const (
	meterBaseRow = 16 // 底座畫在列 16 與 15
	meterTopRow  = 10 // 頂蓋畫在列 10 與 9
	meterCapRow  = 9

	meterHaveCounter = 0x7E // 底座：有蓋氏計數器（接著 7F 80 81）
	meterNoCounter   = 0x7A // 底座：沒有（接著 7B 7C 7D）
	meterEmpty       = 0x82 // 空管（接著 83）
	meterHalfLow     = 0x84 // 半格，列 ≤ 12（接著 85）
	meterFullLow     = 0x86 // 滿格，列 ≤ 12（接著 87）
	meterHalfHigh    = 0x88 // 半格，列 > 12（接著 89）
	meterFullHigh    = 0x8A // 滿格，列 > 12（接著 8B）
	meterCapPlain    = 0x72 // 頂蓋（接著 73 74 75）
	meterCapLit      = 0x76 // 頂蓋：條頂到頂（接著 77 78 79）

	// meterHighRow 是換色的門檻：**列 > 12 用另一組字模**（`0x17EEE` 的 `ja`）。
	meterHighRow = 12
	// MeterNoReading 是「沒有讀數」的保留值：視野裡沒有輻射格，
	// 或全隊沒有人帶蓋氏計數器（`docs/re/120` §2）。
	MeterNoReading = 0xFF
)

// MeterRows 是計量表占的字元列（由上到下：頂蓋兩列、條四列、底座兩列）。
const MeterRows = 8

// MeterGlyphs 算出計量表這一欄由上到下每一列的**左邊那個字模**
// （右邊那個永遠是它 ＋1），`docs/re/120` §3 的表逐條照抄。
//
// reading 是「視野內最近的輻射格有多遠」（距離表的值），
// `MeterNoReading` ＝ 沒有讀數；hasCounter ＝ 隊上有沒有人帶著蓋氏計數器，
// 沒有的話底座換一組字模，而且**讀數一律當成沒有**。
//
// 回傳的索引 0 對到字元列 9（`meterCapRow`），依序往下到列 16。
func MeterGlyphs(reading int, hasCounter bool) [MeterRows]int {
	var out [MeterRows]int
	base := meterHaveCounter
	if !hasCounter {
		base = meterNoCounter
		reading = MeterNoReading
	}
	// 底座固定在最底下兩列（列 15、16），字模連號。
	out[MeterRows-1] = base
	out[MeterRows-2] = base + 2

	// 長度：距離每 8 個表值退一階，共 9 階、畫成 4 格半（`0x17EC3`）。
	full, half := 0, 0
	if step := reading >> 3; step < 9 {
		n := 9 - step
		half = n & 1
		full = n >> 1
	}
	row := meterBaseRow - 2 // 列 14，往上畫
	put := func(g int) {
		out[row-meterCapRow] = g
		row--
	}
	for ; full > 0; full-- {
		g := meterFullLow
		if row > meterHighRow {
			g = meterFullHigh
		}
		put(g)
	}
	// ⚠ **條頂到頂（列 10）之後就不畫半格了**（`0x17F25` 的 `jz`），
	// 那時候半格改成把頂蓋換成亮的那一組。
	litCap := false
	if half != 0 {
		if row > meterTopRow {
			g := meterHalfLow
			if row > meterHighRow {
				g = meterHalfHigh
			}
			put(g)
		} else {
			litCap = true
		}
	}
	for ; row > meterTopRow; row-- {
		out[row-meterCapRow] = meterEmpty
	}
	cap := meterCapPlain
	if litCap {
		cap = meterCapLit
	}
	out[meterTopRow-meterCapRow] = cap
	out[0] = cap + 2
	return out
}

// DrawGeigerMeter 畫右邊那一欄的輻射計量表（`sub_17E42`，`docs/re/120` §3）。
func (f *Frame) DrawGeigerMeter(font *assets.Font, reading int, hasCounter bool) error {
	if font == nil {
		return nil
	}
	for i, g := range MeterGlyphs(reading, hasCounter) {
		row := meterCapRow + i
		if err := f.DrawGlyph(font, g, MeterCol, row, 0, false); err != nil {
			return err
		}
		if err := f.DrawGlyph(font, g+1, MeterCol+1, row, 0, false); err != nil {
			return err
		}
	}
	return nil
}

// 名單框（`sub_16F70`，`docs/re/125`）。
//
// 原版的名單占**字元列 14–23、欄 0–39**：列 14 與列 15 合起來是一條
// 雙倍高的橫幅（`ds:B1E1h` 與 `ds:B20Ah` 兩張字模表，上下半各一列），
// 成員從列 16 起（`sub_1738A` 的「列 ＝ 序號 ＋ 0x0F」），列 23 是下緣。
const (
	RosterBoxTopRow    = 14
	RosterBoxBottomRow = 23
	// RosterMemberRow 是第 1 個成員的字元列（`sub_1738A`）。
	RosterMemberRow = 16
)

// rosterBanner 是名單橫幅的兩張字模表（`ds:B1E1h`／`ds:B20Ah`）。
//
// 索引 1 與索引 3 是**執行期填的組別指示**（`0x16F7E` 起那五行）：
// 上半用 `0x57 + n`、下半用 `0x95 + n`，畫出來是「目前組 `<` 總組數」。
var rosterBanner = [2][]int{
	{
		0x9B, 0x57, 0x5B, 0x57, 0x9E, 0x9C, 0x9C, 0x9C, 0x50, 0x53, 0x52, 0x51,
		0x9C, 0x9C, 0x9C, 0x9C, 0x9E, 0x53, 0x55, 0x9E, 0x53, 0x52, 0x52, 0x9E,
		0x52, 0x53, 0x54, 0x9E, 0x55, 0x56, 0x50, 0x9E, 0x60, 0x51, 0x53, 0x61,
		0x56, 0x50, 0x9F, 0x9A,
	},
	{
		0xA1, 0x95, 0x99, 0x95, 0xA4, 0xA2, 0xA2, 0xA2, 0x8C, 0x8F, 0x8E, 0x8D,
		0xA2, 0xA2, 0xA2, 0xA2, 0xA4, 0x8F, 0x91, 0xA4, 0x8F, 0x8E, 0x8E, 0xA4,
		0x8E, 0x8F, 0x90, 0xA4, 0x91, 0x94, 0x8C, 0xA4, 0x92, 0x8D, 0x8F, 0x93,
		0x94, 0x8C, 0xA5, 0xA0,
	},
}

// 橫幅裡那兩格組別指示的位置與字模起點。
const (
	bannerGroupAt  = 1
	bannerTotalAt  = 3
	bannerTopBase  = 0x57
	bannerBotBase  = 0x95
)

// DrawRosterBanner 畫名單框頂上那條雙倍高的橫幅（列 14–15）。
//
// group 與 total 都是**原版存的那個值**（0 起算），畫出來是 `n+1`。
func (f *Frame) DrawRosterBanner(font *assets.Font, group, total int) error {
	if font == nil {
		return nil
	}
	for half, row := range [2]int{RosterBoxTopRow, RosterBoxTopRow + 1} {
		base := bannerTopBase
		if half == 1 {
			base = bannerBotBase
		}
		for col, g := range rosterBanner[half] {
			switch col {
			case bannerGroupAt:
				g = base + group
			case bannerTotalAt:
				g = base + total
			}
			if err := f.DrawGlyph(font, g, col, row, 0, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// DrawRosterBox 畫名單框的左右與下緣（列 16–23）。
//
// **重製決策**：中文表頭是單倍高的，畫不出原版那條雙倍高的橫幅，
// 這時列 14 改畫一條普通的上緣（與外框同一組字模）。
// 英文路徑走 `DrawRosterBanner`，與原版一樣。
func (f *Frame) DrawRosterBox(font *assets.Font, plainTop bool) error {
	if font == nil {
		return nil
	}
	put := func(index, col, row int) error {
		return f.DrawGlyph(font, index, col, row, 0, false)
	}
	if plainTop {
		if err := put(boxTopLeft, 0, RosterBoxTopRow); err != nil {
			return err
		}
		for col := 1; col < MsgBorderRight; col++ {
			if err := put(boxHorizontal, col, RosterBoxTopRow); err != nil {
				return err
			}
		}
		if err := put(boxTopRight, MsgBorderRight, RosterBoxTopRow); err != nil {
			return err
		}
	}
	for row := RosterMemberRow; row < RosterBoxBottomRow; row++ {
		if err := put(boxLeft, 0, row); err != nil {
			return err
		}
		if err := put(boxRight, MsgBorderRight, row); err != nil {
			return err
		}
	}
	if err := put(boxBottomLeft, 0, RosterBoxBottomRow); err != nil {
		return err
	}
	for col := 1; col < MsgBorderRight; col++ {
		if err := put(boxHorizontal, col, RosterBoxBottomRow); err != nil {
			return err
		}
	}
	return put(boxBottomRight, MsgBorderRight, RosterBoxBottomRow)
}
