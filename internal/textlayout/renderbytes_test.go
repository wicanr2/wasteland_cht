package textlayout

import "testing"

// 中文那條路的控制碼要解乾淨——`\x0A` 就是 `\n`，不解的話畫面上會把
// 單複數兩段都印出來、中間多一個換行，看起來像「這句沒翻好」。
func TestRenderBytesResolvesVariants(t *testing.T) {
	// 「一隻\x0A老鼠\x0A群老鼠\x0A出現」：單數取第 0 段、複數取第 1 段。
	raw := []byte("A\x0Aone\x0Amany\x0AB")
	if got := string(RenderBytes(raw, Options{Count: 1})); got != "AoneB" {
		t.Errorf("單數應該是 %q，得到 %q", "AoneB", got)
	}
	if got := string(RenderBytes(raw, Options{Count: 3})); got != "AmanyB" {
		t.Errorf("複數應該是 %q，得到 %q", "AmanyB", got)
	}
}

func TestRenderBytesInsertsNameAndCount(t *testing.T) {
	raw := []byte("\x0B gains \x0F experience.")
	opt := Options{Count: 14, Name: func() string { return "Hell Razor" }}
	want := "Hell Razor gains 14 experience."
	if got := string(RenderBytes(raw, opt)); got != want {
		t.Errorf("應該是 %q，得到 %q", want, got)
	}
}

// `\x0D` 是換行、`\x10` 是熱鍵標記（標記本身不印、後面那個字母要留）。
func TestRenderBytesNewlineAndHotkey(t *testing.T) {
	raw := []byte("Q?\x0D\x11\x10Y yes\x0D\x10N no")
	want := "Q?\nY yes\nN no"
	if got := string(RenderBytes(raw, Options{})); got != want {
		t.Errorf("應該是 %q，得到 %q", want, got)
	}
}

// Big5 的位元組都 ≥ 0x40，不能被當成控制碼吃掉。
func TestRenderBytesKeepsBig5(t *testing.T) {
	// 「中文」＝ A4 A4 A4 E5
	raw := []byte{0xA4, 0xA4, 0xA4, 0xE5, 0x0D, 0xA4, 0xA4}
	got := RenderBytes(raw, Options{})
	want := []byte{0xA4, 0xA4, 0xA4, 0xE5, '\n', 0xA4, 0xA4}
	if string(got) != string(want) {
		t.Errorf("Big5 位元組被改動了：% x → % x", want, got)
	}
}
