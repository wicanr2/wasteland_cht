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
	"github.com/wicanr2/wasteland_cht/internal/lang"
)

// JournalKey 是打開手札的按鍵。
//
// **重製決策**：`P` ＝ Paragraph。不能用 `J`——那是原版的方向鍵（左），
// 而指令列的七個首字母（U E O D V S R）也全部避開了。
const JournalKey = 'P'

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
	if len(s.journal.Text(n)) == 0 {
		return false
	}
	// **直接開手札那一頁**，不是把正文倒進訊息視窗。
	//
	// 訊息視窗滿了會自動捲到最後一列（`eachMessageCell`），所以長段落
	// 只看得到結尾——27 段那種一開始就看不到開頭。手札那條路有手動捲動
	// （`↑`／`↓`），而地圖上那兩個鍵是走路，**同一塊視窗沒辦法兩種都要**。
	//
	// 關掉手札會把剛才那句地圖訊息放回去（`openJournal` 存、ESC 還原），
	// 所以原版那句 `Read paragraph 23.` 沒有消失，只是被蓋著。
	s.openJournal(n)
	return true
}

// OpenJournal 讓外部（截圖工具、測試）直接翻到某一頁。
func (s *Scene) OpenJournal(n int) { s.openJournal(n) }

// openJournal 打開手札，停在指定段落（超出範圍就夾回來）。
//
// 打開之前先把訊息視窗上的東西收起來，關掉時放回去——**手札是蓋在畫面上的
// 一層，不是一個新場景**。少了這一步，讀完一段回到地圖是一片空白，
// 而剛剛踩到那一格說了什麼就再也看不到了。
func (s *Scene) openJournal(n int) {
	if n < 1 {
		n = 1
	}
	if n > game.JournalPages {
		n = game.JournalPages
	}
	if !s.journalOpen {
		s.journalReturn = journalReturn{message: s.message, cjk: s.cjk}
	}
	s.journalAt = n
	s.journalOpen = true
	s.showJournalPage()
}

// journalReturn 是打開手札之前訊息視窗上的東西。
type journalReturn struct {
	message string
	cjk     string
}

// showJournalPage 把目前這一頁畫進訊息視窗。
func (s *Scene) showJournalPage() {
	n := s.journalAt
	// 換一段就從頭讀起。**要在後日談那條路之前歸零**——它 return 得比正文早，
	// 寫在下面的話從長段落翻到後日談會停在半路。
	s.journalScroll = 0
	// 後日談自成一區：正文不在段落書裡，在執行檔的結局字串表
	// （`docs/re/96` §7）。原版玩不到，重製版收進手札保存。
	if i := game.EpilogueIndex(n); i != 0 {
		s.showEpiloguePage(n, i)
		return
	}
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
	s.setJournalHeader(n, sec)
	s.cjk = ""
	if s.journal != nil {
		s.cjk = s.journal.Text(n)
	}
	if len(s.cjk) == 0 {
		// ⚠ 標記要留著：陷阱段落沒翻也還是陷阱段落。
		s.message = fmt.Sprintf("Journal %d/%d%s  (not translated)",
			n, game.JournalPages, sec)
	}
	s.dirty = true
}

// setJournalHeader 把標題列放進中文視窗；沒有中文就退回英文。
//
// 標題列本身是**重製版的介面**（原版沒有手札畫面），走 `ui:` 那組 key。
func (s *Scene) setJournalHeader(n int, sec string) {
	head := s.uiText("journal.header")
	if len(head) == 0 {
		s.message = fmt.Sprintf("Journal %d/%d%s  (I/K page, ESC close)",
			n, game.JournalPages, sec)
		return
	}
	secCJK := ""
	switch sec {
	case " (decoy)":
		secCJK = string(s.uiText("journal.decoy"))
	case " (appendix)":
		secCJK = string(s.uiText("journal.appendix"))
	case " (epilogue)":
		secCJK = string(s.uiText("journal.epilogue"))
	}
	s.message = ""
	s.journalHead = fmt.Sprintf(string(head), n, game.JournalPages, secCJK +
		"  " + string(s.uiText("journal.hint")))
}

// updateJournal 是手札模式的按鍵。回傳 false 表示要離開遊戲。
//
// 兩組手勢分工（使用者定案 2026-08-18）：
//
//	I／K（與 ←／→）  換段落
//	↑／↓（與 W／S）  在同一段落裡捲動
//
// ⚠ **`I`／`K` 只能從 `Char` 認**：移動鍵是方向鍵與 `WASD`（`internal/input`
// 的 `Bindings`），`IKJL` 沒有綁定，只會以字元進來。以前標題寫著「I／K 翻頁」
// 而程式只看 `Dir`——**按 I 什麼都不會發生，而畫面上看不出來**。
func (s *Scene) updateJournal(in input.Input) (bool, error) {
	switch {
	case in.Action == input.ActionCancel,
		input.Upper(in.Char) == JournalKey:
		s.journalOpen = false
		s.message, s.cjk = s.journalReturn.message, s.journalReturn.cjk
		s.journalReturn = journalReturn{}
		s.journalHead = ""
		s.journalScroll = 0
		s.dirty = true
	case input.Upper(in.Char) == 'I', in.Dir == input.DirLeft:
		if s.journalAt > 1 {
			s.journalAt--
			s.showJournalPage()
		}
	case input.Upper(in.Char) == 'K', in.Dir == input.DirRight:
		if s.journalAt < game.JournalPages {
			s.journalAt++
			s.showJournalPage()
		}
	case in.Dir == input.DirUp:
		s.scrollJournal(-1)
	case in.Dir == input.DirDown:
		s.scrollJournal(1)
	}
	return true, nil
}

// scrollJournal 在同一段落裡上下捲，夾在 0 與「多出來的列數」之間。
//
// ⚠ 上限要**每次重算**：段落長度不同，而且視窗高度在戰鬥時會變（`msgRect`）。
// 記一個算好的上限遲早會與畫面不一致，而症狀是「捲到底還有字看不到」。
func (s *Scene) scrollJournal(d int) {
	rect := s.msgRect()
	max := overflowRows(s.cjk, rect, s.bodyStart(rect))
	at := s.journalScroll + d
	if at < 0 {
		at = 0
	}
	if at > max {
		at = max
	}
	if at == s.journalScroll {
		return
	}
	s.journalScroll = at
	s.dirty = true
}

// showEpiloguePage 畫後日談的一頁。
//
// 正文是結局字串表（`ExeStrings()` 第 4 張）的第 i 條：
// 中文走一般的執行檔字串翻譯（key `exe:4:<i>`），沒有翻譯就顯示英文原文。
func (s *Scene) showEpiloguePage(page, i int) {
	s.setJournalHeader(page, " (epilogue)")
	s.cjk = ""
	if s.cat != nil {
		if b, ok := s.cat.Lookup(lang.ExeKey(EndingTable, i)); ok {
			s.cjk = b
		}
	}
	if len(s.cjk) == 0 {
		// 沒有中文就把英文原文放進訊息視窗——**這一區不會有「未翻譯」的空白頁**，
		// 因為原文一定在執行檔裡拿得到。
		if en := s.endingText(i); en != "" {
			s.message = en
		} else {
			s.message = fmt.Sprintf("Journal %d/%d (epilogue)  (not translated)",
				page, game.JournalPages)
		}
	}
	if s.journal != nil {
		s.journal.MarkRead(page)
	}
	s.dirty = true
}
