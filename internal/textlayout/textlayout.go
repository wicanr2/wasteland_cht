// Package textlayout 把原版的字串排成「一格一個字元」的行。
//
// 這一層**不認識畫面**（docs/spec/03 §2.0）：輸入是 assets 解出來的字串，
// 輸出是格子與事件，所以控制碼的處理可以在無頭環境完整測試。
//
// 控制碼共 18 個（`0x00`–`0x11`，docs/re/14 §4.1），**全部已對上語意**。
// **沒對上的碼一律回報成事件，絕不當文字印出去**——那會讓整行位移全錯，
// 而且症狀是「畫面上多了奇怪的字」，很難回推是哪個碼。
package textlayout

import "fmt"

// 控制碼。名稱只給已確認與強證據的，未解的留原碼。
const (
	CodeEnd        = 0x00 // 收尾
	CodeInverseOn  = 0x01
	CodeInverseOff = 0x02
	CodeClearEOL   = 0x03 // 反白關 ＋ 清到行尾
	CodeClearRect  = 0x04
	CodeWaitKey    = 0x05
	CodeWaitKey2   = 0x06
	CodeNewFrame   = 0x07 // 開一個新的文字框（強證據）
	CodeFlushLine  = 0x08 // 結束這一行但**不捲動**（docs/re/58）
	CodeMoveTo     = 0x09 // ⚠ 帶一個 byte 參數
	CodePlural     = 0x0A // 單複數二選一
	CodeInsertName = 0x0B
	CodeGender     = 0x0C // 性別二選一
	CodeNewline    = 0x0D // 組行版：換行
	CodePronoun    = 0x0E // him／her／it 三選一
	CodeCount      = 0x0F // 印出數量
	CodeCapture    = 0x10 // 捕捉下一個字元
	CodeClearLine  = 0x11
	CodeMax        = 0x11
)

// Cell 是一個字元格。
type Cell struct {
	Char    byte // ASCII
	Inverse bool
}

// Line 是排好的一行。
type Line struct {
	Cells []Cell
}

// String 回傳這一行的文字（測試與比對用）。
func (l Line) String() string {
	b := make([]byte, len(l.Cells))
	for i, c := range l.Cells {
		b[i] = c.Char
	}
	return string(b)
}

// EventKind 是排版過程中遇到的、不屬於「把字放進格子」的事情。
type EventKind int

const (
	EventWaitKey     EventKind = iota // 分頁暫停（0x05／0x06）
	EventClearRect                    // 0x04
	EventClearEOL                     // 0x03
	EventClearLine                    // 0x11
	EventMoveTo                       // 0x09，Param 是位移量
	EventCapture                      // 0x10
	EventNewFrame                     // 0x07
	EventEnd                          // 0x00
	EventFlushLine                    // 0x08，結束這一行但不捲動
	EventUnknownCode                  // 還沒解出語意的控制碼
)

// Event 帶著它發生在第幾行、第幾欄，讓呼叫端能重建原版的時序。
type Event struct {
	Kind  EventKind
	Code  byte
	Param byte
	Line  int
	Col   int
}

// Options 是排版設定。
//
// 三個選擇子對應原版的三個全域（docs/re/28 §3）：文字裡的分段只有一段會輸出，
// **段數必須與原版一致**，少一個分隔碼整句就錯位。
type Options struct {
	// Width 是行寬上限（字元數）。原版訊息視窗是 38（ds:4675h ＝ 0x26）。
	Width int

	// Name 提供 0x0B 要插入的角色名字。沒給就插入空字串。
	Name func() string

	// Count 是數量（原版 ds:4687h）：1 ＝ 單數。同時是 0x0A 的選擇子與
	// 0x0F 印出來的數字。預設 1。
	Count int

	// Gender 是 0x0C 的選擇子（原版 ds:470Bh ← 角色記錄 +0x18）：
	// 0 ＝ 第 0 段（男）、非 0 ＝ 第 1 段。
	Gender int

	// Pronoun 是 0x0E 的選擇子（原版 ds:471Ah）：0／1／2 → him／her／it。
	Pronoun int
}

// Result 是排版結果。
type Result struct {
	Lines  []Line
	Events []Event
}

// variant 是「分段只輸出其中一段」的狀態（原版是逐字元過濾器，docs/re/28 §2）。
//
// 原版**每個碼各有自己的段計數器**（`0x0A` → `ds:CDCBh`、`0x0C` → `ds:CDCAh`、
// `0x0E` → `ds:CDC9h`），所以不同的碼可以互相巢狀——實際的戰鬥訊息就是這樣用的：
// 外層 `0x0A` 的第 0 段裡面包著一整組 `0x0E`。
// **外層沒選中時，內層連分隔碼都不會被看到**（原版的過濾器是串起來的，
// 前面吃掉就不會傳到後面）。
type variant struct {
	active bool
	seg    int // 目前在第幾段
	last   int // 最後一段的編號（段數 − 1）
	pick   int // 要輸出哪一段
}

func (v *variant) match() bool { return !v.active || v.seg == v.pick }

// Layout 把一段原版字串排版。
//
// ⚠ **自動換行目前是硬斷**：超過 Width 就換行，不退到前一個空白。
// 原版的斷字規則還沒從 `sub_19E53` 讀出來（docs/spec/03 §3），
// 所以這是暫代行為，不是逆向結論。
func Layout(text []byte, opt Options) (Result, error) {
	if opt.Width <= 0 {
		return Result{}, fmt.Errorf("行寬必須大於 0，收到 %d", opt.Width)
	}
	count := opt.Count
	if count <= 0 {
		count = 1
	}
	// 選擇子 → 要輸出第幾段。
	pluralPick := 1 // 複數取第 1 段
	if count == 1 {
		pluralPick = 0
	}
	genderPick := 0
	if opt.Gender != 0 {
		genderPick = 1
	}
	pronounPick := opt.Pronoun
	if pronounPick < 0 || pronounPick > 2 {
		pronounPick = 0
	}
	// 每個變形碼各一份狀態，對應原版三個獨立的段計數器。
	vars := map[byte]*variant{
		CodePlural:  {last: 1, pick: pluralPick},
		CodeGender:  {last: 1, pick: genderPick},
		CodePronoun: {last: 2, pick: pronounPick},
	}
	allMatch := func() bool {
		for _, v := range vars {
			if !v.match() {
				return false
			}
		}
		return true
	}
	var res Result
	line := Line{}
	inverse := false

	flush := func() {
		res.Lines = append(res.Lines, line)
		line = Line{}
	}
	emit := func(kind EventKind, code, param byte) {
		res.Events = append(res.Events, Event{
			Kind: kind, Code: code, Param: param,
			Line: len(res.Lines), Col: len(line.Cells),
		})
	}
	put := func(c byte) {
		if len(line.Cells) >= opt.Width {
			flush()
		}
		line.Cells = append(line.Cells, Cell{Char: c, Inverse: inverse})
	}
	putString := func(s string) {
		for i := 0; i < len(s); i++ {
			put(s[i])
		}
	}

	for i := 0; i < len(text); i++ {
		c := text[i]

		// 分段進行中：分隔碼推進段號，其餘字元只有全部選中時才輸出。
		if v, ok := vars[c]; ok && v.active {
			v.seg++
			if v.seg > v.last {
				v.active = false
				v.seg = 0
			}
			continue
		}
		if !allMatch() {
			continue // 被某一層吃掉：連內層的分隔碼都看不到
		}
		if v, ok := vars[c]; ok {
			v.active, v.seg = true, 0
			continue
		}

		if c > CodeMax {
			// 一般字元。原版的 \r（0x0D）落在控制碼範圍內，在下面處理。
			put(c)
			continue
		}
		switch c {
		case CodeInverseOn:
			inverse = true
		case CodeInverseOff:
			inverse = false
		case CodeNewline:
			flush()
		case CodeFlushLine:
			// 0x08 與 0x0D 都結束這一行，差別是**不捲動也不延遲**
			// （0x0D 走 sub_19EFC，0x08 直接進 sub_19F12，docs/re/58 §3）。
			// 三條用例都在字串結尾——用 0x0D 收尾畫面會多捲一行。
			emit(EventFlushLine, c, 0)
			flush()
		case CodeInsertName:
			if opt.Name != nil {
				putString(opt.Name())
			}
		case CodeCount:
			putString(itoa(count))
		case CodeNewFrame:
			emit(EventNewFrame, c, 0)
		case CodeMoveTo:
			// ⚠ 帶一個 byte 參數：那個 byte 不是文字，切字串時也不能當文字。
			var param byte
			if i+1 < len(text) {
				i++
				param = text[i]
			}
			emit(EventMoveTo, c, param)
		case CodeWaitKey, CodeWaitKey2:
			emit(EventWaitKey, c, 0)
		case CodeClearRect:
			emit(EventClearRect, c, 0)
		case CodeClearEOL:
			inverse = false
			emit(EventClearEOL, c, 0)
		case CodeClearLine:
			emit(EventClearLine, c, 0)
			line = Line{}
		case CodeCapture:
			emit(EventCapture, c, 0)
		case CodeEnd:
			emit(EventEnd, c, 0)
		default:
			// 18 個碼都有語意了；真的遇到別的碼要**回報，不印**——
			// 印出來就會變成畫面上的怪字。
			emit(EventUnknownCode, c, 0)
		}
	}
	if len(line.Cells) > 0 || len(res.Lines) == 0 {
		flush()
	}
	return res, nil
}

// itoa 只處理非負小數字，避免為了一個數字拉進 strconv。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [8]byte
	i := len(buf)
	for n > 0 && i > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Paginate 把排好的行切成每頁 height 行——訊息視窗只有 6 行，
// 原版靠 0x05／0x06 分頁，remake 要自己決定何時停。
func Paginate(lines []Line, height int) [][]Line {
	if height <= 0 {
		return nil
	}
	var pages [][]Line
	for i := 0; i < len(lines); i += height {
		end := i + height
		if end > len(lines) {
			end = len(lines)
		}
		pages = append(pages, lines[i:end])
	}
	return pages
}
