package play

// 手札的段落內捲動（使用者定案 2026-08-18）。
//
// ⚠ 這一組守的是「**長段落讀得到開頭**」。166 條裡有 27 條放不下六列、
// 最長 38 列——訊息視窗那套自動捲到底會把開頭吃掉，而畫面上完全看不出來：
// 那一段就是從半句話開始，看起來像翻譯本來就是那樣。

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// journalScene 開一個載好手札的場景。
func journalScene(t *testing.T) *Scene {
	t.Helper()
	s := newScene(t)
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("載入翻譯目錄：%v", err)
	}
	if err := s.LoadJournal("../../docs/re/generated/paragraph-refs.tsv",
		"../../translations/paragraphs-zh-Hant.cat"); err != nil {
		t.Skipf("載入手札：%v", err)
	}
	return s
}

// longParagraph 找第一條放不下的段落。找不到就跳過——沒有長段落就沒有這個問題。
func longParagraph(t *testing.T, s *Scene) int {
	t.Helper()
	for n := 1; n <= game.JournalPages; n++ {
		s.openJournal(n)
		if len(s.cjk) == 0 {
			continue
		}
		rect := s.msgRect()
		if overflowRows(s.cjk, rect, s.bodyStart(rect)) > 0 {
			return n
		}
	}
	t.Skip("沒有放不下的段落")
	return 0
}

// visibleRows 回傳畫面上看得到的每一列。
func visibleRows(s *Scene) map[int]string {
	rows := map[int]string{}
	rect := s.msgRect()
	s.walkBody(s.cjk, rect, func(col, row int, r rune) { rows[row] += string(r) })
	return rows
}

// 打開一條長段落時，看到的要是**開頭**。
func TestJournalStartsAtTheTop(t *testing.T) {
	s := journalScene(t)
	n := longParagraph(t, s)
	s.openJournal(n)
	rows := visibleRows(s)
	rect := s.msgRect()
	first := rows[s.bodyStart(rect)]
	// 正文的第一個字元就該出現在第一列的第一格。
	want := []rune(strings.TrimLeft(s.cjk, "\r\n"))
	if len(want) == 0 || first == "" {
		t.Fatalf("第 %d 段沒有正文", n)
	}
	if []rune(first)[0] != want[0] {
		t.Errorf("第 %d 段的第一列是 %q，預期從 %q 開始", n, first, string(want[0]))
	}
}

// Page Up／Page Down 在段落內捲，夾在 0 與多出來的列數之間（使用者定案 2026-08-18）。
//
// ⚠ 自己組 `Input` 一定要寫 `Dir: input.DirNone`：`Direction` 的零值是 `DirUp`
// （`internal/input` 的說明），忘了就等於每一次都同時按了「上」。
func TestJournalScrollsWithinAParagraph(t *testing.T) {
	s := journalScene(t)
	n := longParagraph(t, s)
	s.openJournal(n)
	rect := s.msgRect()
	max := overflowRows(s.cjk, rect, s.bodyStart(rect))
	page := s.journalPageStep()
	top := visibleRows(s)[s.bodyStart(rect)]

	down := input.Input{Dir: input.DirNone, Scroll: input.ScrollDown}
	up := input.Input{Dir: input.DirNone, Scroll: input.ScrollUp}

	step(t, s, down)
	want := page
	if want > max {
		want = max
	}
	if s.journalScroll != want {
		t.Fatalf("Page Down 之後 scroll ＝ %d，預期 %d", s.journalScroll, want)
	}
	if got := visibleRows(s)[s.bodyStart(rect)]; got == top {
		t.Errorf("捲了一頁，第一列卻沒變：%q", got)
	}
	// 段落沒有換。
	if s.journalAt != n {
		t.Errorf("Page Down 把段落換掉了：%d → %d", n, s.journalAt)
	}

	step(t, s, up)
	if s.journalScroll != 0 {
		t.Errorf("Page Up 之後 scroll ＝ %d，預期回到 0", s.journalScroll)
	}
	// 上緣夾住：再按一次不會變成負的。
	step(t, s, up)
	if s.journalScroll != 0 {
		t.Errorf("捲過頭了：%d", s.journalScroll)
	}
	// 下緣夾住，而且**捲到底時最後一行看得到**。
	for i := 0; i < max/page+5; i++ {
		step(t, s, down)
	}
	if s.journalScroll != max {
		t.Errorf("捲到底 ＝ %d，預期 %d", s.journalScroll, max)
	}
	rows := visibleRows(s)
	if _, ok := rows[rect.LastRow()]; !ok {
		t.Error("捲到底之後最後一列是空的")
	}
	if _, ok := rows[rect.Row]; ok && s.journalHead != "" {
		t.Error("正文捲到標題那一列上了")
	}
}

// 手札裡的 ↑／↓ 不再捲動（使用者定案 2026-08-18：捲動改由 Page Up／Page Down 負責）。
//
// ⚠ 這一條要**同時**擋兩件事：↑／↓ 還在捲（改了一半），
// 以及它們被別的分支接走去換段落——後者在畫面上看起來像「捲得很快」。
func TestJournalArrowsDoNotScroll(t *testing.T) {
	s := journalScene(t)
	n := longParagraph(t, s)
	s.openJournal(n)

	step(t, s, input.Input{Dir: input.DirDown})
	step(t, s, input.Input{Dir: input.DirUp})
	if s.journalScroll != 0 {
		t.Errorf("↑／↓ 還在捲：scroll ＝ %d", s.journalScroll)
	}
	if s.journalAt != n {
		t.Errorf("↑／↓ 把段落換掉了：%d → %d", n, s.journalAt)
	}
}

// 換段落要從頭讀起。
func TestJournalPageChangeResetsScroll(t *testing.T) {
	s := journalScene(t)
	n := longParagraph(t, s)
	s.openJournal(n)
	step(t, s, input.Input{Dir: input.DirDown})
	step(t, s, input.Input{Dir: input.DirDown})
	if s.journalScroll == 0 {
		t.Skip("這一段捲不動")
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'K'})
	if s.journalScroll != 0 {
		t.Errorf("換段之後還停在第 %d 列", s.journalScroll)
	}
}

// ⚠ `I`／`K` 一定要真的能換段落——標題上寫著它們。
// 以前程式只看 `Dir`，而 `IKJL` 沒有綁進 `Bindings`，所以按下去什麼都不會發生。
func TestJournalIKChangeParagraph(t *testing.T) {
	s := journalScene(t)
	s.openJournal(5)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'K'})
	if s.journalAt != 6 {
		t.Errorf("按 K 應該到第 6 段，得到 %d", s.journalAt)
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'I'})
	if s.journalAt != 5 {
		t.Errorf("按 I 應該回到第 5 段，得到 %d", s.journalAt)
	}
}
