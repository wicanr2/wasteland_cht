package play

// 手札的呈現與按鍵（`docs/spec/19` §3、§4）。
//
// 原版沒有這個畫面——段落印在紙本上，遊戲只給編號。**重製版不移植防拷**
// （`CLAUDE.md` §3.2），所以讀到編號就直接把正文顯示出來，
// 另外給一個隨時翻得開的手札。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// JournalKey 是打開手札的按鍵。
//
// **重製決策**：`P` ＝ Paragraph。不能用 `J`——那是原版的方向鍵（左），
// 而指令列的七個首字母（U E O D V S R）也全部避開了。
const JournalKey = 'P'

// showParagraph 把某一段的正文放進訊息視窗，並記成已讀。
//
// 查不到中文就回 false，呼叫端維持原本的英文訊息——
// **沒翻的段落不能變成空白頁**（`docs/spec/19` §6）。
func (s *Scene) showParagraph(n int) bool {
	if s.journal == nil {
		return false
	}
	text := s.journal.Text(n)
	if len(text) == 0 {
		return false
	}
	s.journal.MarkRead(n)
	s.cjk = text
	s.message = fmt.Sprintf("Paragraph %d", n)
	s.dirty = true
	return true
}

// paragraphFor 查這一則訊息有沒有引用段落（引用表是編譯期產物，
// `docs/spec/19` §2——**不在執行期解析翻譯過的文字**）。
func (s *Scene) paragraphFor(key string) (int, bool) {
	if s.journal == nil {
		return 0, false
	}
	return s.journal.Refs().Lookup(key)
}

// maybeParagraph 檢查這一步的訊息有沒有引用段落，有就把正文顯示出來。
//
// 原版那句 `Read paragraph 23.` 照樣留著（它本身也是原版的文字），
// **接著**把第 23 段的正文顯示出來——玩家不必翻紙本、不做編號驗證
// （`docs/spec/19` §3）。查不到中文就什麼都不做，維持原本的訊息。
func (s *Scene) maybeParagraph(key string) bool {
	n, ok := s.paragraphFor(key)
	if !ok {
		return false
	}
	if s.journal != nil {
		s.journal.MarkRead(n)
	}
	text := s.journal.Text(n)
	if len(text) == 0 {
		return false
	}
	s.cjk = text
	s.journalAt = n // 之後按 P 開手札就停在這一段
	s.dirty = true
	return true
}

// openJournal 打開手札，停在指定段落（超出範圍就夾回來）。
func (s *Scene) openJournal(n int) {
	if n < 1 {
		n = 1
	}
	if n > game.ParagraphCount {
		n = game.ParagraphCount
	}
	s.journalAt = n
	s.journalOpen = true
	s.showJournalPage()
}

// showJournalPage 把目前這一頁畫進訊息視窗。
func (s *Scene) showJournalPage() {
	n := s.journalAt
	sec := ""
	if s.journal != nil {
		switch s.journal.Section(n) {
		case game.SectionAppendix:
			// 附錄要標明是防拷設計（`docs/spec/19` §4）——
			// 陷阱段落遊戲永遠不會叫玩家去讀，混在正文裡會誤導。
			sec = " (appendix)"
		}
		if s.journal.IsTrap(n) {
			sec = " (decoy)"
		}
	}
	s.message = fmt.Sprintf("Journal %d/%d%s  (I/K page, ESC close)",
		n, game.ParagraphCount, sec)
	s.cjk = nil
	if s.journal != nil {
		s.cjk = s.journal.Text(n)
	}
	if len(s.cjk) == 0 {
		// ⚠ 標記要留著：陷阱段落沒翻也還是陷阱段落。
		s.message = fmt.Sprintf("Journal %d/%d%s  (not translated)",
			n, game.ParagraphCount, sec)
	}
	s.dirty = true
}

// updateJournal 是手札模式的按鍵。回傳 false 表示要離開遊戲。
func (s *Scene) updateJournal(in input.Input) (bool, error) {
	switch {
	case in.Action == input.ActionCancel,
		input.Upper(in.Char) == JournalKey:
		s.journalOpen = false
		s.message = ""
		s.cjk = nil
		s.dirty = true
	case in.Dir == input.DirUp || in.Dir == input.DirLeft:
		if s.journalAt > 1 {
			s.journalAt--
			s.showJournalPage()
		}
	case in.Dir == input.DirDown || in.Dir == input.DirRight:
		if s.journalAt < game.ParagraphCount {
			s.journalAt++
			s.showJournalPage()
		}
	}
	return true, nil
}
