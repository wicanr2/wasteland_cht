package play

import (
	"unicode/utf8"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

func orderScene(t *testing.T) *Scene {
	t.Helper()
	s := newScene(t)
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("沒有翻譯目錄：%v", err)
	}
	return s
}

// `ORDER` 印的是**原版字串 15**（`Pick a player:`），不是重製版自己寫的話。
//
// ⚠ 自己寫一句的話那條原版字串永遠不會出現在畫面上，中文化覆蓋率也量不到——
// 與 `docs/re/105` §5 的 `Nothing to fight here.` 是同一條規則。
func TestOrderUsesTheOriginalPrompt(t *testing.T) {
	s := orderScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'O'})

	want := s.cjkExe(exeTable1, orderPromptStr, textlayout.Options{})
	if len(want) == 0 {
		t.Fatal("字串 15 查不到中文")
	}
	if !strings.HasPrefix(s.cjk, want) {
		t.Errorf("面板上不是字串 15 開頭：%q", s.cjk)
	}
	// 名字不翻譯，所以一定看得到。
	if !strings.Contains(s.cjk, "Hell Razor") {
		t.Error("名單不見了")
	}
}

// 排完之後那一句是**重製版加的**（原版排完不印任何一句），所以走 `ui:`。
func TestOrderDoneLineIsMarkedAsRemakeText(t *testing.T) {
	s := orderScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'O'})
	for _, k := range []byte{'4', '3', '2', '1'} {
		step(t, s, input.Input{Dir: input.DirNone, Char: k})
	}
	head := s.uiText("order.done")
	if len(head) == 0 {
		t.Skip("沒有 ui:order.done")
	}
	if !strings.HasPrefix(s.cjk, head) {
		t.Errorf("收尾那句不是走 ui:order.done：%q", s.cjk)
	}
}


// 訊息全部是合法的 UTF-8。
//
// ⚠ **這條在 Big5 那一版是必要的**：任何拼接只要漏了一次編碼，
// 就會出現「讀得出筆畫卻不成字」的東西（`、` 的三個 UTF-8 byte 被當成
// Big5 高位元組）。改成 UTF-8 之後那個 bug 類別消失了——Go 的字串串接
// 不可能產生半個字。留著這一條是為了擋**另一種**錯：從中間截斷。
func TestCJKMessagesAreValidUTF8(t *testing.T) {
	s := orderScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'O'})
	for _, k := range []byte{'4', '3', '2', '1'} {
		step(t, s, input.Input{Dir: input.DirNone, Char: k})
	}
	check := func(what, v string) {
		t.Helper()
		if !utf8.ValidString(v) {
			t.Errorf("%s 不是合法的 UTF-8：% x", what, v)
		}
	}
	check("ORDER 收尾", s.cjk)
	for _, k := range []byte{'U', 'D', 'V', 'S'} {
		s2 := orderScene(t)
		step(t, s2, input.Input{Dir: input.DirNone, Char: k})
		check(string(k), s2.cjk)
	}
}
