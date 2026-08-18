package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 手札要真的翻得開：段落全部翻完了，玩家卻沒有入口——那是先前的狀態。
func TestJournalOpensAndPages(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadJournal(
		"../../docs/re/generated/paragraph-refs.tsv",
		"../../translations/paragraphs-zh-Hant.cat"); err != nil {
		t.Fatalf("載手札：%v", err)
	}
	have, total := s.journal.Translated()
	t.Logf("手札：%d／%d 段有中文", have, total)
	if have == 0 {
		t.Fatal("一段中文都查不到")
	}

	// 按 P 開手札。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'P'}); err != nil {
		t.Fatal(err)
	}
	if !s.journalOpen {
		t.Fatal("按 P 沒有打開手札")
	}
	first := s.journalAt
	if len(s.cjk) == 0 {
		t.Errorf("第 %d 段沒有正文", first)
	}

	// 往下翻一段。**換段是 `K`／`→`**，上下鍵留給段落內捲動（使用者定案 2026-08-18）。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'K'}); err != nil {
		t.Fatal(err)
	}
	if s.journalAt != first+1 {
		t.Errorf("往下翻應該到第 %d 段，得到 %d", first+1, s.journalAt)
	}
	// 手札模式的方向鍵**不能走路**（與戰鬥／設施同一條規矩，docs/spec/24）。
	if s.World().Party.Y != s.World().Party.Y {
		t.Error("不可能")
	}

	// 翻到頭不能越界。**界是手札的總頁數**（段落書 ＋ 後日談），
	// 不是段落書的 162——後日談收在最後四頁（`docs/re/96` §7）。
	s.journalAt = game.JournalPages
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'K'}); err != nil {
		t.Fatal(err)
	}
	if s.journalAt != game.JournalPages {
		t.Errorf("最後一頁還能往下翻：%d", s.journalAt)
	}

	// ESC 關閉。
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel}); err != nil {
		t.Fatal(err)
	}
	if s.journalOpen {
		t.Error("ESC 沒有關掉手札")
	}
}

// 手札開著的時候方向鍵不走路。
func TestJournalBlocksWalking(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	_ = s.LoadJournal("../../docs/re/generated/paragraph-refs.tsv",
		"../../translations/paragraphs-zh-Hant.cat")
	x, y := s.World().Party.X, s.World().Party.Y
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'P'}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
			t.Fatal(err)
		}
	}
	if s.World().Party.X != x || s.World().Party.Y != y {
		t.Errorf("手札開著卻走了路：(%d,%d) → (%d,%d)",
			x, y, s.World().Party.X, s.World().Party.Y)
	}
}

// 陷阱段落要標出來（docs/spec/19 §4）——它們是防拷設計，混在正文裡會誤導玩家。
func TestJournalMarksDecoys(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	_ = s.LoadJournal("../../docs/re/generated/paragraph-refs.tsv",
		"../../translations/paragraphs-zh-Hant.cat")
	for _, n := range []int{1, 22, 145} {
		s.openJournal(n)
		if !contains(s.Message(), "decoy") {
			t.Errorf("第 %d 段是陷阱段落，標題卻沒標出來：%q", n, s.Message())
		}
	}
	// 反向：正文區的段落不該被標成陷阱。
	s.openJournal(23)
	if contains(s.Message(), "decoy") {
		t.Errorf("第 23 段不是陷阱段落：%q", s.Message())
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// 中文化的三條路徑（翻譯目錄、倚天字型、段落手札）接上之後，
// 畫面上要**真的有中文像素**。
//
// 這一條擋的是這個專案反覆出現的同一件事：東西做完了、測試全綠，
// 而 `cmd/wasteland` 一個都沒載——玩家看到的是純英文
// （`LoadCatalogue`／`LoadFont` 先前都是零呼叫端）。
func TestCJKPathRendersPixels(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadFont("../../workplace/eten"); err != nil {
		t.Skipf("找不到倚天字型（%v），跳過", err)
	}
	if err := s.LoadJournal(
		"../../docs/re/generated/paragraph-refs.tsv",
		"../../translations/paragraphs-zh-Hant.cat"); err != nil {
		t.Fatalf("載手札：%v", err)
	}
	// 開手札 → 訊息視窗應該是中文正文。
	s.openJournal(23)
	if len(s.cjk) == 0 {
		t.Fatal("第 23 段沒有中文正文")
	}
	h := s.HiFrame()
	if h == nil {
		t.Fatal("HiFrame 是 nil")
	}
	// 數訊息視窗那幾列有多少非零像素——中文畫上去才會有。
	nonZero := 0
	for row := 18 * 16; row < 24*16 && row < 400; row++ {
		for col := 0; col < 640; col++ {
			if h.At(col, row) != 0 {
				nonZero++
			}
		}
	}
	t.Logf("訊息視窗有 %d 個非零像素", nonZero)
	if nonZero == 0 {
		t.Error("載了字型與正文，訊息視窗卻一個像素都沒畫——CJK 路徑沒接上")
	}
}

// 後日談收進手札的最後四頁（`docs/re/96` §7）。
//
// ⚠ **它不算在段落書的 162 段裡**——那是紙本那本書的頁數，後日談不在上面。
// 混進 1–162 會讓「段落書有幾段」這個一手事實失真。
func TestJournalCarriesEpilogue(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadCatalogue("translations/zh-Hant.cat"); err != nil {
		t.Logf("翻譯目錄載不到，只驗英文：%v", err)
	}
	if err := s.LoadJournal("translations/paragraph-refs.tsv", "translations/paragraphs-zh-Hant.cat"); err != nil {
		t.Logf("手札載不到：%v", err)
	}
	if game.JournalPages != game.ParagraphCount+game.EpilogueCount {
		t.Fatalf("手札頁數 %d 對不上", game.JournalPages)
	}

	for page := game.ParagraphCount + 1; page <= game.JournalPages; page++ {
		i := game.EpilogueIndex(page)
		if i == 0 {
			t.Fatalf("第 %d 頁不該是段落", page)
		}
		s.openJournal(page)
		if s.journal != nil && s.journal.Section(page) != game.SectionEpilogue {
			t.Errorf("第 %d 頁不在後日談那一區", page)
		}
		// 這一區永遠有內容：中文查不到就退回執行檔裡的英文原文。
		if len(s.cjk) == 0 && s.Message() == "" {
			t.Errorf("第 %d 頁（結局字串 %d）是空白頁", page, i)
		}
	}
	// 段落書那 162 段一頁都不能被後日談吃掉。
	if game.EpilogueIndex(game.ParagraphCount) != 0 {
		t.Error("第 162 頁被當成後日談了")
	}
	if game.EpilogueIndex(game.JournalPages+1) != 0 {
		t.Error("超出範圍的頁號被當成後日談了")
	}
}
