package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// idle 是一幀「什麼都沒按」。
//
// ⚠ `Direction` 的零值是 `DirUp` 不是 `DirNone`——寫 `input.Input{}` 的話
// 每一幀都會被當成按了上鍵。
func idle() input.Input { return input.Input{Dir: input.DirNone} }

// 畫面的格數（中文佔兩格）。
const (
	screenCols = render.ScreenWidth / render.CharWidth
	screenRows = render.ScreenHeight / render.CharHeight
)

// cells 是一行字佔幾格：ASCII 一格、中文兩格。
func cells(line string) int {
	n := 0
	for _, r := range line {
		if r < 0x80 {
			n++
		} else {
			n += 2
		}
	}
	return n
}

func titleScene(t *testing.T) *Scene {
	t.Helper()
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Logf("沒有翻譯目錄（%v），這一輪只驗英文", err)
	}
	s.BeginTitle()
	return s
}

// 標題畫面按非 `S` 的鍵就開始播片頭（`docs/re/113` §1）。
func TestAttractStartsOnAnyKey(t *testing.T) {
	s := titleScene(t)
	if s.Attract() {
		t.Fatal("還沒按鍵就在播片頭")
	}
	if _, err := s.Update(idle()); err != nil {
		t.Fatal(err)
	}
	if s.Attract() {
		t.Fatal("空的一幀不該觸發片頭——那會變成閒置逾時，原版沒有這個")
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'X'}); err != nil {
		t.Fatal(err)
	}
	if !s.Attract() {
		t.Fatal("按了鍵卻沒開始播片頭")
	}
	if s.AttractPage() != 0 {
		t.Errorf("片頭從第 %d 頁開始，應該是第 0 頁", s.AttractPage())
	}
}

// 每頁 255 刻、六頁循環（`docs/re/113` §3、§4）。
func TestAttractPagesLoop(t *testing.T) {
	s := titleScene(t)
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'X'}); err != nil {
		t.Fatal(err)
	}
	// 一刻 ＝ 3 幀。差一幀不該換頁，剛好一頁就要換。
	frames := func(n int) {
		for i := 0; i < n; i++ {
			if _, err := s.Update(idle()); err != nil {
				t.Fatal(err)
			}
		}
	}
	frames(attractPageTicks*3 - 1)
	if s.AttractPage() != 0 {
		t.Errorf("差一幀就換頁了（現在第 %d 頁）", s.AttractPage())
	}
	frames(1)
	if s.AttractPage() != 1 {
		t.Errorf("滿一頁沒換頁（現在第 %d 頁）", s.AttractPage())
	}
	// 再走五頁要繞回第 0 頁。
	frames(attractPageTicks * 3 * (len(AttractPages) - 1))
	if s.AttractPage() != 0 {
		t.Errorf("繞了一圈之後在第 %d 頁，應該回到第 0 頁", s.AttractPage())
	}
}

// 第 4 頁（Place／Year／Status）是一條一條浮出來的。
func TestAttractTypeInPage(t *testing.T) {
	s := titleScene(t)
	s.attract = attractState{active: true, page: attractTypeIn}
	if got := len(s.attractSlots()); got != 1 {
		t.Errorf("打字頁一開始有 %d 條，應該只有 1 條", got)
	}
	s.attract.tick = attractLineTicks
	if got := len(s.attractSlots()); got != 2 {
		t.Errorf("過了一個行間停頓有 %d 條，應該是 2 條", got)
	}
	s.attract.tick = attractPageTicks - 1
	if got := len(s.attractSlots()); got != len(AttractPages[attractTypeIn]) {
		t.Errorf("這一頁結束前有 %d 條，應該是 %d 條",
			got, len(AttractPages[attractTypeIn]))
	}
}

// 六頁的文字都拿得到，而且中文與英文的行數一致。
func TestAttractPagesHaveText(t *testing.T) {
	s := titleScene(t)
	s.attract = attractState{active: true}
	for page := range AttractPages {
		s.attract = attractState{active: true, page: page, tick: attractPageTicks - 1}
		en := s.attractLines(false)
		if len(en) == 0 {
			t.Errorf("第 %d 頁一行英文都沒有", page)
			continue
		}
		zh := s.attractLines(true)
		if s.cat == nil {
			continue
		}
		if len(zh) == 0 {
			t.Errorf("第 %d 頁沒有中文（有一條沒翻，整頁就退回英文）", page)
			continue
		}
		// 行數不必與英文一致——中文一行裝得下英文兩行。
		// 要守的是**畫得下**：每一行不超過 40 格，整頁不超過畫面高度。
		for i, line := range zh {
			if w := cells(line); w > screenCols {
				t.Errorf("第 %d 頁第 %d 行寬 %d 格，畫面只有 %d 格：%q",
					page, i, w, screenCols, line)
			}
		}
		if rows := attractRow + len(zh); rows > screenRows {
			t.Errorf("第 %d 頁要 %d 列，畫面只有 %d 列", page, rows, screenRows)
		}
	}
}

// 槽 8 是寫了沒接上的一句：原版 15 個呼叫端沒有一個傳 8（`docs/re/113` §3.1）。
//
// 正對照在同一道裡：槽 8 **確實有內容**，所以「不播」不是因為它是空的。
func TestAttractSkipsUnusedSlot(t *testing.T) {
	rom := openRom(t)
	tables, err := rom.ExeStrings()
	if err != nil {
		t.Fatal(err)
	}
	tbl := tables[AttractTable]
	if len(tbl) <= 8 || !strings.Contains(tbl[8], "Computer defense") {
		t.Fatalf("槽 8 不是那一句（%q）——表號或解碼變了，下面那一條證明不了任何事",
			tbl[min(8, len(tbl)-1)])
	}
	for page, slots := range AttractPages {
		for _, slot := range slots {
			if slot == 8 {
				t.Errorf("第 %d 頁播了槽 8，原版不播", page)
			}
		}
	}
}

// 片頭播到一半按 `S` 照樣開始遊戲（原版的兩支處理程式沒變）。
func TestAttractStartStillWorks(t *testing.T) {
	s := titleScene(t)
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'X'}); err != nil {
		t.Fatal(err)
	}
	if !s.Attract() {
		t.Fatal("沒進片頭")
	}
	if err := s.LoadMap(4, 18, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'S'}); err != nil {
		t.Fatal(err)
	}
	if s.showTitle() || s.Attract() {
		t.Error("按了 S 還停在標題／片頭")
	}
}
