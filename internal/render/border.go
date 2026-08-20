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

// DrawBox 是通用畫框（`sub_19814`，`docs/re/124` §1）：左上角 (col0, row0)、
// 右下角 (col1, row1)，六個字模拼出來。
func (f *Frame) DrawBox(font *assets.Font, col0, row0, col1, row1 int) error {
	if font == nil {
		return nil
	}
	put := func(index, col, row int) error {
		return f.DrawGlyph(font, index, col, row, 0, false)
	}
	if err := put(boxTopLeft, col0, row0); err != nil {
		return err
	}
	for col := col0 + 1; col < col1; col++ {
		if err := put(boxHorizontal, col, row0); err != nil {
			return err
		}
	}
	if err := put(boxTopRight, col1, row0); err != nil {
		return err
	}
	for row := row0 + 1; row < row1; row++ {
		if err := put(boxLeft, col0, row); err != nil {
			return err
		}
		if err := put(boxRight, col1, row); err != nil {
			return err
		}
	}
	if err := put(boxBottomLeft, col0, row1); err != nil {
		return err
	}
	for col := col0 + 1; col < col1; col++ {
		if err := put(boxHorizontal, col, row1); err != nil {
			return err
		}
	}
	return put(boxBottomRight, col1, row1)
}

// 名單模式的兩個框（`sub_19770` 與 `sub_19727`，`docs/re/127`）。
//
// 肖像框 (0,0)–(13,13)、選單框 (14,0)–(39,13)；選單框裡的文字區是
// 欄 15–38、**列 1–12**（列 13 是框的下緣，`POOL MONEY` 就印在上面）。
const (
	PortraitBoxCol1, PortraitBoxRow1 = 13, 13
	MenuBoxCol0, MenuBoxCol1         = 14, 39
	MenuBoxRow1                      = 13
)

// DrawPortraitBox 畫左上角那一圈（`sub_19770`）。
func (f *Frame) DrawPortraitBox(font *assets.Font) error {
	return f.DrawBox(font, 0, 0, PortraitBoxCol1, PortraitBoxRow1)
}

// DrawMenuBox 畫右上角那一圈（`sub_19727`）。
func (f *Frame) DrawMenuBox(font *assets.Font) error {
	return f.DrawBox(font, MenuBoxCol0, 0, MenuBoxCol1, MenuBoxRow1)
}

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

// 框邊上的標籤（`ds:CBBDh` 的版面表 ＋ `ds:CA70h` 的字模串，`docs/re/126`）。
//
// ⚠ 這一族字串在原版映像裡是**彩色字型的字模編號**，不是 ASCII——
// 所以這裡也照字模編號存，不要換成 Go 字串再轉。
type BoxLabel struct {
	Col, Row int
	Glyphs   []int
	// Key 是這個標籤代表的按鍵（版面表的 `+0x03`）。
	//
	// 原版的標籤同時是滑鼠按鈕：熱區表 `ds:CAEBh` 的第 4–20 筆與版面表
	// **一筆對一筆、順序相同**，共用同一支處理常式（`0x8BFF`），
	// 那支做的事就是「送出這個按鍵」（`docs/re/126` §3）。
	Key byte

	// Vertical ＝ 字模往**下**疊而不是往右排。
	//
	// 版面表前四筆（四個箭頭）就是這一種：它們都在欄 39，而 39 已經是
	// 最後一欄——橫著排會排到畫面外。原版的畫法是每畫一個字模就
	// `inc 列 / dec 欄`（`0x18BD0`），後者正好抵銷繪製常式的欄前進
	// （`docs/re/129` §3）。
	Vertical bool
}

// Hit 回答「這一格點得到這個標籤嗎」。
//
// **重製決策**：範圍取標籤自己占的格子。原版的熱區是像素矩形，
// 與標籤差一兩格（`POOL MONEY` 的標籤是欄 21–32、熱區是欄 20–30），
// 而重製版的滑鼠一律「功能等價就好」（`docs/spec/29`）。
func (l BoxLabel) Hit(col, row int) bool {
	if l.Vertical {
		return col == l.Col && row >= l.Row && row < l.Row+len(l.Glyphs)
	}
	return row == l.Row && col >= l.Col && col < l.Col+len(l.Glyphs)
}

// 字模編碼：`索引 ＝ (字元 & 0xDF) − 0x29`，空白 `0x33`，頭尾各一個蓋子
// （`sub_17451`，`docs/re/14` §3.1）。
const (
	labelCapLeft  = 0x17
	labelCapRight = 0x28
	labelSpace    = 0x33

	// colourBankLo／colourBankHi 是**會換色的那一段**字模（`0x18`–`0x33`，
	// A–Z 加六個符號）。overlay slot 19 在 `ds:722Fh` ＝ 0 時把索引 **＋0x1C**
	// 換成冷色那一組（`docs/re/14` §3），而 `ds:722Fh` 的預設值就是 0。
	//
	// ⚠ 不補這 `0x1C` 的話標籤會畫成暖色（紅／黃），實機是藍的——
	// 而畫面上只是「顏色不太一樣」，不像少了一步。
	colourBankLo   = 0x18
	colourBankHi   = 0x33
	colourBankCool = 0x1C
)

// coolGlyph 把會換色的那一段字模換成冷色那一組。
func coolGlyph(g int) int {
	if g >= colourBankLo && g <= colourBankHi {
		return g + colourBankCool
	}
	return g
}

func labelGlyphs(text string) []int {
	out := make([]int, 0, len(text)+2)
	out = append(out, coolGlyph(labelCapLeft))
	for _, r := range text {
		if r == ' ' {
			out = append(out, coolGlyph(labelSpace))
			continue
		}
		out = append(out, coolGlyph(int(byte(r)&0xDF)-0x29))
	}
	return append(out, coolGlyph(labelCapRight))
}

// 三個實機確認過位置的標籤（`docs/re/126` §2）。
//
// ⚠ **`ROSTER ON` 與 `ROSTER OFF` 是兩筆記錄，位置不同**：名單收起來時
// 按鈕在地圖框的下緣（列 17），名單開著時在名單框的下緣（列 23）。
// 當成「同一個按鈕換字」會畫錯位置。
var (
	LabelRosterOn  = BoxLabel{Col: 15, Row: 17, Glyphs: labelGlyphs("ROSTER ON"), Key: ' '}
	LabelRosterOff = BoxLabel{Col: 15, Row: 23, Glyphs: labelGlyphs("ROSTER OFF"), Key: ' '}
	LabelEsc       = BoxLabel{Col: 17, Row: 0, Glyphs: labelGlyphs("ESC"), Key: 0x1B}
	LabelPoolMoney = BoxLabel{Col: 21, Row: 13, Glyphs: labelGlyphs("POOL MONEY"), Key: 'P'}
	LabelMap       = BoxLabel{Col: 17, Row: 13, Glyphs: labelGlyphs("MAP"), Key: ' '}
	// `NEXT` 在選單框上緣的右半（訓練師的技能清單，`docs/re/129` §4 的 `0x1C8C2`）。
	LabelNext = BoxLabel{Col: 31, Row: 0, Glyphs: labelGlyphs("NEXT"), Key: ' '}
)

// 四個捲動箭頭（版面表第 0–3 筆，`docs/re/129` §4）。
//
// 兩組：清單那一組在選單框右邊（列 3／10），訊息視窗那一組在畫面右下
// （列 18／22）。**兩格高、沒有頭尾的蓋子**，欄一律 39。
//
// ⚠ 按鍵是原版的擴充碼（方向鍵與 PageUp／PageDown 的掃描碼 ＋ `0x80`），
// 不是 ASCII——轉成重製版的動作在 `internal/play/mouse.go`。
var (
	LabelListUp   = BoxLabel{Col: 39, Row: 3, Glyphs: arrowUp, Key: KeyArrowUp, Vertical: true}
	LabelListDown = BoxLabel{Col: 39, Row: 10, Glyphs: arrowDown, Key: KeyArrowDown, Vertical: true}
	LabelMsgUp    = BoxLabel{Col: 39, Row: 18, Glyphs: arrowUp, Key: KeyPageUp, Vertical: true}
	LabelMsgDown  = BoxLabel{Col: 39, Row: 22, Glyphs: arrowDown, Key: KeyPageDown, Vertical: true}
)

// 原版版面表 `+0x03` 給這四筆的按鍵碼（`generated/ida94/box-labels.md`）。
const (
	KeyArrowUp   = 0xC8
	KeyArrowDown = 0xD0
	KeyPageUp    = 0xC9
	KeyPageDown  = 0xD1
)

// 箭頭的字模：字模串 `ds:CA70h`／`ds:CA73h` 各兩個，換色規則與其他標籤相同。
var (
	arrowUp   = []int{coolGlyph(0x1E), coolGlyph(0x1F)}
	arrowDown = []int{coolGlyph(0x21), coolGlyph(0x22)}
)

// DrawBoxLabel 把一個標籤畫在框線上。
func (f *Frame) DrawBoxLabel(font *assets.Font, l BoxLabel) error {
	if font == nil {
		return nil
	}
	for i, g := range l.Glyphs {
		col, row := l.Col+i, l.Row
		if l.Vertical {
			col, row = l.Col, l.Row+i
		}
		if err := f.DrawGlyph(font, g, col, row, 0, false); err != nil {
			return err
		}
	}
	return nil
}
