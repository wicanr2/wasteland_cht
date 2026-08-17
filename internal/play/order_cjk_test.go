package play

import (
	"bytes"
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
	if !bytes.HasPrefix(s.cjk, want) {
		t.Errorf("面板上不是字串 15 開頭：%q", s.cjk)
	}
	// 名字不翻譯，所以一定看得到。
	if !bytes.Contains(s.cjk, []byte("Hell Razor")) {
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
	if !bytes.HasPrefix(s.cjk, head) {
		t.Errorf("收尾那句不是走 ui:order.done：%q", s.cjk)
	}
}

// **接進 Big5 位元組串的東西不能是 UTF-8。**
//
// ⚠ `、` 的 UTF-8 是三個 byte，直接 append 會被當成 Big5 高位元組解讀，
// 把後面那個字吃掉——畫面上是「Vargas粻 hrasher」這種讀得出筆畫卻不成字的東西。
// 這條掃整段輸出：**Big5 兩兩配對之後不該剩下落單的高位元組**。
func TestCJKMessagesAreValidBig5(t *testing.T) {
	s := orderScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'O'})
	for _, k := range []byte{'4', '3', '2', '1'} {
		step(t, s, input.Input{Dir: input.DirNone, Char: k})
	}
	checkBig5(t, "ORDER 收尾", s.cjk)

	// 順帶掃幾條走過的路。
	for _, k := range []byte{'U', 'D', 'V', 'S'} {
		s2 := orderScene(t)
		step(t, s2, input.Input{Dir: input.DirNone, Char: k})
		checkBig5(t, string(k), s2.cjk)
	}
}

func checkBig5(t *testing.T, what string, b []byte) {
	t.Helper()
	for i := 0; i < len(b); {
		c := b[i]
		if c < 0x80 {
			i++
			continue
		}
		// Big5 首位元組 0xA1–0xF9，尾位元組 0x40–0x7E 或 0xA1–0xFE。
		if c < 0xA1 || c > 0xF9 {
			t.Errorf("%s：位移 %d 的 %#x 不是合法的 Big5 首位元組（UTF-8 漏進來了？）：%q",
				what, i, c, b)
			return
		}
		if i+1 >= len(b) {
			t.Errorf("%s：結尾有落單的高位元組 %#x：%q", what, c, b)
			return
		}
		lo := b[i+1]
		if !((lo >= 0x40 && lo <= 0x7E) || (lo >= 0xA1 && lo <= 0xFE)) {
			t.Errorf("%s：位移 %d 的尾位元組 %#x 不合法：%q", what, i+1, lo, b)
			return
		}
		i += 2
	}
	_ = strings.TrimSpace
}
