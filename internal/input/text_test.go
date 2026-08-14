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

// 按鍵巨集（docs/re/43 §6）：Alt+F1 開始錄、再按一次收工、F1 播回來。
func TestMacroRecordAndPlay(t *testing.T) {
	var m Macros
	const altF1, f1 = 0xDE, 0xBB

	if _, ok := m.Next(altF1); ok {
		t.Fatal("Alt+F1 本身不該被送進遊戲")
	}
	if m.Recording() != 1 {
		t.Fatalf("應該在錄第 1 組，得到 %d", m.Recording())
	}
	for _, c := range []byte("NORTH") {
		if got, ok := m.Next(c); !ok || got != c {
			t.Fatalf("錄製中按鍵還是要照送：%q ok=%v", got, ok)
		}
	}
	if _, ok := m.Next(altF1); ok || m.Recording() != 0 {
		t.Fatalf("再按一次 Alt+F1 應該收工，Recording=%d", m.Recording())
	}

	// 播回來。
	if _, ok := m.Next(f1); ok {
		t.Fatal("F1 本身不該被送進遊戲")
	}
	var got []byte
	for i := 0; i < 20; i++ {
		k, ok := m.Next(0) // 沒有真實按鍵
		if !ok {
			break
		}
		got = append(got, k)
	}
	if string(got) != "NORTH" {
		t.Fatalf("播回來是 %q，應該是 \"NORTH\"", got)
	}
	if m.Playing() != 0 {
		t.Fatal("播完應該歸零")
	}
}

// 錄製中按 F 鍵要被丟掉（不會變成巢狀播放）。
func TestMacroIgnoresPlayKeyWhileRecording(t *testing.T) {
	var m Macros
	m.Next(0xDE) // Alt+F1 開始錄
	if _, ok := m.Next(0xBB); ok {
		t.Fatal("F1 不該被送進遊戲")
	}
	if m.Playing() != 0 {
		t.Fatal("錄製中按 F1 不該開始播放")
	}
}

// 錄滿 255 個之後不再收，但按鍵照樣送進遊戲。
func TestMacroFullStopsRecordingNotInput(t *testing.T) {
	var m Macros
	m.Next(0xDE)
	for i := 0; i < MacroLen+50; i++ {
		if _, ok := m.Next('A'); !ok {
			t.Fatalf("第 %d 個按鍵沒被送出去", i)
		}
	}
	m.Next(0xDE) // 收工
	m.Next(0xBB) // 播
	n := 0
	for {
		if _, ok := m.Next(0); !ok {
			break
		}
		n++
		if n > MacroLen+10 {
			t.Fatal("播不完——長度沒有被限住")
		}
	}
	if n != MacroLen {
		t.Fatalf("錄了 %d 個，上限應該是 %d", n, MacroLen)
	}
}
