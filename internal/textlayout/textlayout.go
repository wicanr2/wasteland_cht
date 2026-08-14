// Package textlayout 把原版的字串排成「一格一個字元」的行。
//
// 這一層**不認識畫面**（docs/spec/03 §2.0）：輸入是 assets 解出來的字串，
// 輸出是格子與事件，所以控制碼的處理可以在無頭環境完整測試。
//
// 控制碼共 18 個（`0x00`–`0x11`，docs/re/14 §4.1），其中 14 個已對上語意。
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
	CodeUnknown07  = 0x07
	CodeUnknown08  = 0x08
	CodeMoveTo     = 0x09 // ⚠ 帶一個 byte 參數
	CodeUnknown0A  = 0x0A
	CodeInsertName = 0x0B
	CodeGender     = 0x0C // 夾 his/her 做性別選字
	CodeNewline    = 0x0D // 組行版：換行
	CodeUnknown0E  = 0x0E
	CodeUnknown0F  = 0x0F
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
	EventGender                       // 0x0C
	EventEnd                          // 0x00
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
type Options struct {
	// Width 是行寬上限（字元數）。原版訊息視窗是 38（ds:4675h ＝ 0x26）。
	Width int

	// Name 提供 0x0B 要插入的角色名字。沒給就插入空字串。
	Name func() string

	// Gender 提供 0x0C 的性別選字結果。沒給就當事件回報、不插字。
	Gender func() string
}

// Result 是排版結果。
type Result struct {
	Lines  []Line
	Events []Event
}

// Layout 把一段原版字串排版。
//
// ⚠ **自動換行目前是硬斷**：超過 Width 就換行，不退到前一個空白。
// 原版的斷字規則還沒從 `sub_19E53` 讀出來（docs/spec/03 §3），
// 所以這是暫代行為，不是逆向結論。
func Layout(text []byte, opt Options) (Result, error) {
	if opt.Width <= 0 {
		return Result{}, fmt.Errorf("行寬必須大於 0，收到 %d", opt.Width)
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
		case CodeInsertName:
			if opt.Name != nil {
				putString(opt.Name())
			}
		case CodeGender:
			emit(EventGender, c, 0)
			if opt.Gender != nil {
				putString(opt.Gender())
			}
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
			// 0x07／0x08／0x0A／0x0E／0x0F：語意未解。
			// **回報，不印**——印出來就會變成畫面上的怪字。
			emit(EventUnknownCode, c, 0)
		}
	}
	if len(line.Cells) > 0 || len(res.Lines) == 0 {
		flush()
	}
	return res, nil
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
