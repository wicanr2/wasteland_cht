package play

import (
	"sort"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 踩到引用段落的格子 → **直接開手札那一頁**，而且從第一列開始。
//
// 訊息視窗滿了會自動捲到最後一列，所以長段落倒進去只看得到結尾——
// 27 段那種一開始就看不到開頭。這一道守的是「開頭看得到」。
func TestParagraphOpensJournalAtTop(t *testing.T) {
	s := journalScene(t)
	// 找一條真的會引用段落的 key，模擬踩上去。
	var keys []string
	for k := range s.journal.Refs() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var hit string
	for _, k := range keys {
		n, _ := s.journal.Refs().Lookup(k)
		if len(s.journal.Text(n)) > 0 {
			hit = k
			break
		}
	}
	if hit == "" {
		t.Fatal("引用表裡一條有中文正文的都沒有")
	}
	n, _ := s.journal.Refs().Lookup(hit)

	s.message, s.cjk = "Read paragraph 23.", "地圖上那一句"
	if !s.maybeParagraph(hit) {
		t.Fatalf("key %q 引用第 %d 段，卻沒有顯示出來", hit, n)
	}
	if !s.journalOpen {
		t.Error("踩到引用段落的格子沒有開手札")
	}
	if s.journalAt != n {
		t.Errorf("手札停在第 %d 段，應該是第 %d 段", s.journalAt, n)
	}
	if s.journalScroll != 0 {
		t.Errorf("一開啟就捲到第 %d 列，應該從頭讀", s.journalScroll)
	}
	// 畫面上第一列正文要真的是段落的開頭。
	body := s.journal.Text(n)
	first := firstBodyRune(t, s)
	want := []rune(strings.TrimLeft(body, "\r\n "))[0]
	if first != want {
		t.Errorf("畫面第一個字是 %q，段落開頭是 %q", first, want)
	}

	// 關掉手札 → 地圖那一句要回來。
	if _, err := s.Update(input.Input{Dir: input.DirNone,
		Action: input.ActionCancel}); err != nil {
		t.Fatal(err)
	}
	if s.journalOpen {
		t.Error("ESC 沒有關掉手札")
	}
	if s.Message() != "Read paragraph 23." || s.CJK() != "地圖上那一句" {
		t.Errorf("關掉手札之後訊息是 %q／%q，應該是踩上去那一句",
			s.Message(), s.CJK())
	}
}

// firstBodyRune 是訊息視窗正文區第一個畫出來的字。
func firstBodyRune(t *testing.T, s *Scene) rune {
	t.Helper()
	rect := s.msgRect()
	best := struct {
		col, row int
		r        rune
	}{col: 1 << 30, row: 1 << 30}
	s.walkBody(s.cjk, rect, func(col, row int, r rune) {
		if row < best.row || (row == best.row && col < best.col) {
			best.col, best.row, best.r = col, row, r
		}
	})
	if best.r == 0 {
		t.Fatal("正文區一個字都沒畫")
	}
	return best.r
}
