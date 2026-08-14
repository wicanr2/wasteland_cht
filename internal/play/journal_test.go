package play

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

func paragraphCat() string {
	return filepath.Join("..", "..", "translations", "paragraphs-zh-Hant.cat")
}

func TestJournalWithoutCatalogueIsSafe(t *testing.T) {
	j := NewJournal(game.ParagraphRefs{}, nil)
	if b := j.Text(7); b != nil {
		t.Fatalf("沒有目錄時應該回 nil，得到 %q", b)
	}
	have, total := j.Translated()
	if have != 0 || total != game.ParagraphCount {
		t.Fatalf("沒有目錄時是 %d/%d，應該是 0/%d", have, total, game.ParagraphCount)
	}
}

func TestJournalLoadsParagraphs(t *testing.T) {
	j := NewJournal(game.ParagraphRefs{}, nil)
	if err := j.LoadParagraphs(paragraphCat()); err != nil {
		t.Skipf("段落目錄還沒編出來：%v", err)
	}

	// 已翻的段落查得到正文；還沒翻的回 nil，不能變成空白頁。
	if b := j.Text(7); len(b) == 0 {
		t.Fatal("段落 7 已經翻好了，應該查得到正文")
	}
	have, total := j.Translated()
	if have == 0 || have > total {
		t.Fatalf("已翻 %d 段、共 %d 段，數字不合理", have, total)
	}
	if b := j.Text(total + 1); b != nil {
		t.Fatal("超出 162 的編號不該查得到東西")
	}

	// 陷阱段落照樣有正文——手札一段都不刪，只是標記出來（docs/spec/19 §4）。
	if b := j.Text(22); len(b) == 0 {
		t.Fatal("陷阱段落 22 也要收進手札，不能因為是防拷設計就拿掉")
	}
	if !j.IsTrap(22) {
		t.Fatal("段落 22 應該標成陷阱")
	}
}
