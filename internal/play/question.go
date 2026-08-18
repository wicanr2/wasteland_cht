package play

// nibble 8 ＝ 問答（`0x15160`，`docs/re/46` §4）。
//
// 規則層（`internal/game/answers.go`）早就寫好了：怎麼數答案、怎麼比、
// 答對第 N 個要拿哪個位移去改寫這一格。缺的是**呈現層**——踩上去只印
// 一句 `CHOOSE.` 就沒有下文，玩家答不了任何一題。
//
// 少了這一條，密語、暗號、控制面板全部不能用；科奇斯基地的自毀序列
// （四站啟動 → 按鈕 → 紅黃綠藍）整段走不動，也就到不了結局
// （`docs/re/100` §3）。

import (
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// questionState 是一題進行中的狀態。
type questionState struct {
	active bool
	q      game.Question
	rec    []byte
	entry  input.TextEntry
}

// beginQuestion 開一題（`0x15160`–`0x1518A`）。
//
// 記錄壞掉（答案清單讀不到結束標記）就不開——原版會一路讀下去，
// 但那是資料異常，安靜地問一題讀不完的題目更難查。
func (s *Scene) beginQuestion(rec []byte) {
	q, ok := game.ParseQuestion(rec)
	if !ok {
		return
	}
	s.question = questionState{
		active: true,
		q:      q,
		rec:    rec,
		entry:  input.TextEntry{Max: input.MaxAnswer},
	}
	s.showQuestion()
}

// showQuestion 把題目（與打字模式下已經打了什麼）放到訊息列。
func (s *Scene) showQuestion() {
	slot := int(s.question.q.Prompt)
	s.message = ""
	s.cjk = ""
	if s.world != nil && s.world.Block != nil &&
		slot >= 0 && slot < len(s.world.Block.Strings) {
		s.message = s.world.Block.Strings[slot]
	}
	if zh := s.cjkLookup(lang.BlockKey(s.blockFile, s.blockID, slot),
		textlayout.Options{}); zh != "" {
		s.cjk = zh
		s.message = "" // 有中文就不要再畫英文（訊息視窗只有六行）
	}
	if !s.question.q.SingleKey {
		// 打字模式原版會先印一個 `>`（`0x151B6` 的 `sub_19DC3`）。
		typed := s.question.entry.Text()
		if s.cjk != "" {
			s.cjk += "\r>" + string(typed)
		} else {
			s.message += "\r>" + string(typed)
		}
	}
	s.dirty = true
}

// Asking 之外的第二種「畫面停在等一個鍵」：問答。
func (s *Scene) questionActive() bool { return s.question.active }

// updateQuestion 收一題的按鍵。
//
// 兩種模式（記錄 `+0x00` 的 bit7）：
//
//	單鍵：按什麼就是答案；Enter 與 ESC 是「不回答」，直接離開
//	打字：走 `input.TextEntry`（16 bytes 上限、Enter 收尾吃掉尾端空白）
func (s *Scene) updateQuestion(in input.Input) (bool, error) {
	key := byte(0)
	switch {
	case in.Action == input.ActionCancel:
		key = 0x1B
	case in.Action == input.ActionConfirm:
		key = 0x0D
	case in.Char != 0:
		key = input.Upper(in.Char)
	}
	if key == 0 {
		return true, nil
	}

	if s.question.q.SingleKey {
		// `0x1519D`：Enter 或 ESC 直接離開，不改這一格。
		if key == 0x0D || key == 0x1B {
			s.closeQuestion()
			return true, nil
		}
		s.answerQuestion([]byte{key & 0x7F})
		return true, nil
	}

	switch s.question.entry.Key(key) {
	case input.EntryCancel:
		s.closeQuestion()
	case input.EntryDone:
		// `0x151D0`：緩衝區是空的就離開，不比也不改。
		text := s.question.entry.Text()
		if len(text) == 0 {
			s.closeQuestion()
			return true, nil
		}
		s.answerQuestion(text)
	default:
		s.showQuestion()
	}
	return true, nil
}

// answerQuestion 比答案並改寫腳下那一格。
func (s *Scene) answerQuestion(text []byte) {
	q, rec := s.question.q, s.question.rec
	answers := make([][]byte, 0, len(q.Answers))
	for _, slot := range q.Answers {
		// ⚠ 密語比對是**逐 byte 全等**（`sub_18D8E`，`docs/re/46`），
		// 而輸入層是 ASCII——這條路刻意不走 UTF-8 文字層。
		answers = append(answers, []byte(s.blockAnswer(int(slot))))
	}
	// **照順序試，第一個相等的贏**；全部不中就落到「答案數」那一格，
	// 也就是答錯那一支（`0x15219` 的 `ds:0A651h`）。
	n := game.MatchAnswer(text, answers)
	s.closeQuestion()
	s.world.PatchHere(rec, q.AnswerBranch(n))
	// 改寫過的那一格要重畫；訊息由改寫後那一格自己的事件印
	// （踩上去才觸發，與原版一致）。
	s.dirty = true
}

// blockAnswer 取一條答案的原文位元組。
//
// ⚠ **答案不走翻譯目錄**：比對是逐 byte 全等而輸入層是 ASCII
// （`docs/re/46` §3），翻成中文玩家就永遠打不出來。
// `translations/must-not-translate.tsv` 擋的就是這一批。
func (s *Scene) blockAnswer(slot int) string {
	if s.world == nil || s.world.Block == nil ||
		slot < 0 || slot >= len(s.world.Block.Strings) {
		return ""
	}
	return strings.TrimRight(s.world.Block.Strings[slot], "\r\n ")
}

func (s *Scene) closeQuestion() {
	s.question = questionState{}
	s.message, s.cjk = "", ""
	s.dirty = true
}
