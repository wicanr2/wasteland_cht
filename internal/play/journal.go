package play

import (
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/lang"
)

// 段落手札的呈現層接線（docs/spec/19、`20-mouse-input` 之外的另一份 READY 規格）。
//
// 規則層（`internal/game` 的 Journal）只認編號；正文放在自己的目錄檔
// `translations/paragraphs-zh-Hant.cat`，key 是 `para:<編號>`。
// 段落書不在原版的字串語料裡（它是紙本），所以另開一個檔，
// 但格式與 key 的形狀沿用同一套，Go 這邊不必多一份讀檔器。

// Journal 是手札：規則層的分區／已讀狀態，加上正文的來源。
type Journal struct {
	*game.Journal
	paragraphs *lang.Catalogue
}

// NewJournal 建一份手札。paragraphs 可以是 nil——那時所有段落都查不到正文，
// 呼叫端顯示英文原文或提示尚未翻譯，**不要顯示編號充數**（`docs/spec/19` §6）。
func NewJournal(refs game.ParagraphRefs, paragraphs *lang.Catalogue) *Journal {
	return &Journal{Journal: game.NewJournal(refs), paragraphs: paragraphs}
}

// LoadParagraphs 載入段落正文目錄；載不到就維持沒有中文，不當成錯誤
// （與翻譯目錄一致，`docs/spec/11` §7）。
func (j *Journal) LoadParagraphs(path string) error {
	c, err := lang.Load(path)
	if err != nil {
		return err
	}
	j.paragraphs = c
	return nil
}

// Text 拿編號換中文正文（Big5）。查不到回 nil——**沒翻的段落不能變成空白頁**，
// 呼叫端要自己決定顯示什麼。
func (j *Journal) Text(n int) []byte {
	if j == nil || j.paragraphs == nil {
		return nil
	}
	if b, ok := j.paragraphs.Lookup(lang.ParagraphKey(n)); ok {
		return b
	}
	return nil
}

// Translated 回報已經有中文正文的段落數，以及總段數。
// 中文化是分批進行的，這個數字是進度也是驗收依據。
func (j *Journal) Translated() (have, total int) {
	total = game.ParagraphCount
	if j == nil || j.paragraphs == nil {
		return 0, total
	}
	for n := 1; n <= total; n++ {
		if _, ok := j.paragraphs.Lookup(lang.ParagraphKey(n)); ok {
			have++
		}
	}
	return have, total
}
