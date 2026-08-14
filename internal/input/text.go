package input

// 文字輸入（`loc_17750`，docs/re/46 §2）。
//
// 原版只有一支輸入常式，參數全是全域變數：緩衝區、上限、目前長度。
// 用它的地方有三個——角色名字（13 bytes）、打字回答（16 bytes）、
// 數字欄位（只收 `'0'`–`'9'`）。

// TextEntry 是一次文字輸入的狀態。
//
// ⚠ **上限是 byte 數不是字數。** 原版量的是 byte，中文一個字兩個 byte，
// 所以直接照抄會把中文從中間切斷。重製版要按字數限制（見 docs/re/46 §6），
// 這裡保留原版的 byte 語意，由呼叫端決定要不要換。
type TextEntry struct {
	Buf    []byte
	Max    int  // 對應 ds:4684h
	Digits bool // 對應 ds:467Bh ≠ 0：只收 '0'–'9'
}

// EntryResult 是一次按鍵的結果。
type EntryResult int

const (
	EntryContinue EntryResult = iota // 還在輸入
	EntryDone                        // Enter：成功
	EntryCancel                      // ESC：取消
)

// 原版的三個上限（docs/re/46 §2、§4、§5）。
const (
	MaxAnswer = 0x10 // 打字回答
	MaxName   = 0x0D // 角色名字
)

// Upper 是 `sub_18EFE` 出口那一段：`'a'`–`'z'` 減 0x20。
//
// ⚠ **整個遊戲拿到的字母永遠是大寫**，所以密語比對不需要（也沒有）
// 另外做大小寫折疊。
func Upper(c byte) byte {
	if c >= 'a' && c < '{' {
		return c - 0x20
	}
	return c
}

// Key 餵一個按鍵進去。key 是已經轉過大寫的碼（原版在鍵盤層就轉好了）。
func (t *TextEntry) Key(key byte) EntryResult {
	switch {
	case key == 0x1B: // ESC
		return EntryCancel

	case key == 0x08 || key == 0xFF: // Backspace
		if len(t.Buf) > 0 {
			t.Buf = t.Buf[:len(t.Buf)-1]
		}
		return EntryContinue

	case key == 0x0D: // Enter
		// ⚠ 原版會從尾巴往回吃掉空白（0x17809）。前導空白不吃。
		for len(t.Buf) > 0 && t.Buf[len(t.Buf)-1] == ' ' {
			t.Buf = t.Buf[:len(t.Buf)-1]
		}
		return EntryDone

	case key < 0x20: // 控制字元，丟掉
		return EntryContinue
	}

	if t.Digits && (key < '0' || key > '9') {
		return EntryContinue
	}
	if len(t.Buf) >= t.Max-1 { // 留一格給結尾的 NUL
		return EntryContinue
	}
	// ⚠ 原版的 `and al, 0x7F` 會把 Big5 的高位 byte 打壞（0xA4 → 0x24）。
	// 中文輸入不能走這條路，所以這裡只對 ASCII 範圍照抄。
	t.Buf = append(t.Buf, key&0x7F)
	return EntryContinue
}

// Text 回目前輸入的內容。
func (t *TextEntry) Text() []byte { return t.Buf }

// Reset 清空緩衝區。
func (t *TextEntry) Reset() { t.Buf = t.Buf[:0] }
