package input

import "testing"

// 按鍵在讀進來的當下就轉大寫（sub_18EFE 出口，docs/re/46 §2.1）。
func TestUpperMatchesOriginal(t *testing.T) {
	for c := 0; c < 256; c++ {
		got := Upper(byte(c))
		want := byte(c)
		if c >= 'a' && c < '{' {
			want = byte(c) - 0x20
		}
		if got != want {
			t.Fatalf("Upper(%#02x) ＝ %#02x，應該是 %#02x", c, got, want)
		}
	}
	// '{' 不在範圍內（原版比的是 < 0x7B）。
	if Upper('{') != '{' {
		t.Fatal("'{' 不該被動到")
	}
}

// 打字回答：上限 16 bytes，其中一格留給結尾的 NUL。
func TestAnswerEntryLength(t *testing.T) {
	e := &TextEntry{Max: MaxAnswer}
	for i := 0; i < 40; i++ {
		e.Key('A')
	}
	if len(e.Text()) != MaxAnswer-1 {
		t.Fatalf("最多收 %d bytes，實際 %d", MaxAnswer-1, len(e.Text()))
	}
}

// Enter 會吃掉尾端空白，前導空白不吃（0x17809）。
func TestEnterTrimsTrailingSpaces(t *testing.T) {
	e := &TextEntry{Max: MaxAnswer}
	for _, c := range []byte("  BIRD   ") {
		e.Key(c)
	}
	if res := e.Key(0x0D); res != EntryDone {
		t.Fatalf("Enter 應該回 EntryDone，得到 %v", res)
	}
	if got := string(e.Text()); got != "  BIRD" {
		t.Fatalf("trim 之後是 %q，應該是 %q", got, "  BIRD")
	}
}

// ESC 取消、Backspace 刪一個、控制字元丟掉。
func TestEntryControlKeys(t *testing.T) {
	e := &TextEntry{Max: MaxAnswer}
	e.Key('A')
	e.Key('B')
	e.Key(0x08)
	if got := string(e.Text()); got != "A" {
		t.Fatalf("Backspace 之後是 %q", got)
	}
	e.Key(0x0A) // 控制字元
	if got := string(e.Text()); got != "A" {
		t.Fatalf("控制字元不該進緩衝區，得到 %q", got)
	}
	if res := e.Key(0x1B); res != EntryCancel {
		t.Fatalf("ESC 應該回 EntryCancel，得到 %v", res)
	}
	// 空緩衝區按 Backspace 不該 panic。
	empty := &TextEntry{Max: MaxAnswer}
	empty.Key(0x08)
}

// 數字模式只收 '0'–'9'（ds:467Bh ≠ 0）。
func TestDigitsOnlyMode(t *testing.T) {
	e := &TextEntry{Max: MaxName, Digits: true}
	for _, c := range []byte("1A2B3") {
		e.Key(c)
	}
	if got := string(e.Text()); got != "123" {
		t.Fatalf("數字模式收到 %q，應該是 \"123\"", got)
	}
}
