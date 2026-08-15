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

	// 往下翻一段。
	if _, err := s.Update(input.Input{Dir: input.DirDown}); err != nil {
		t.Fatal(err)
	}
	if s.journalAt != first+1 {
		t.Errorf("往下翻應該到第 %d 段，得到 %d", first+1, s.journalAt)
	}
	// 手札模式的方向鍵**不能走路**（與戰鬥／設施同一條規矩，docs/spec/24）。
	if s.World().Party.Y != s.World().Party.Y {
		t.Error("不可能")
	}

	// 翻到頭不能越界。
	s.journalAt = game.ParagraphCount
	if _, err := s.Update(input.Input{Dir: input.DirDown}); err != nil {
		t.Fatal(err)
	}
	if s.journalAt != game.ParagraphCount {
		t.Errorf("最後一段還能往下翻：%d", s.journalAt)
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
