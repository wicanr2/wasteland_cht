package textlayout

import "testing"

func TestControlCodesNeverPrint(t *testing.T) {
	text := []byte{
		CodeInverseOn, 'A', 'B', CodeInverseOff,
		CodeMoveTo, 0x41, // ⚠ 0x41 是參數不是文字
		'C',
	}
	res, err := Layout(text, Options{Width: 38})
	if err != nil {
		t.Fatal(err)
	}
	got := res.Lines[0].String()
	if got != "ABC" {
		t.Fatalf("排出來是 %q，應該是 \"ABC\"——控制碼或它的參數被當成文字了", got)
	}
	if !res.Lines[0].Cells[0].Inverse || res.Lines[0].Cells[2].Inverse {
		t.Fatalf("反白範圍不對：%+v", res.Lines[0].Cells)
	}

	moveTo := 0
	for _, e := range res.Events {
		if e.Kind == EventMoveTo {
			moveTo++
			if e.Param != 0x41 {
				t.Fatalf("0x09 的參數是 %#x，應該是 0x41", e.Param)
			}
		}
		if e.Kind == EventUnknownCode {
			t.Fatalf("18 個碼都有語意了，不該回報未解碼 %#x", e.Code)
		}
	}
	if moveTo != 1 {
		t.Fatalf("EventMoveTo 出現 %d 次", moveTo)
	}
}

// 驗收（docs/re/58 §3）：0x08 結束這一行但不捲動，而且不印任何字。
//
// 三條語料用例都在字串結尾——`Which way?` 的十字圖用 0x0D 收尾會多捲一行。
func TestFlushLineEndsLineWithoutPrinting(t *testing.T) {
	res, err := Layout([]byte{'K', CodeFlushLine, 'C'}, Options{Width: 38})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 2 {
		t.Fatalf("0x08 應該結束這一行，得到 %d 行", len(res.Lines))
	}
	if res.Lines[0].String() != "K" || res.Lines[1].String() != "C" {
		t.Fatalf("排出來是 %q／%q", res.Lines[0].String(), res.Lines[1].String())
	}
	flush := 0
	for _, e := range res.Events {
		if e.Kind == EventFlushLine {
			flush++
			if e.Code != CodeFlushLine {
				t.Fatalf("事件帶的碼是 %#x", e.Code)
			}
		}
	}
	if flush != 1 {
		t.Fatalf("EventFlushLine 出現 %d 次", flush)
	}
}

func TestNewlineAndWrap(t *testing.T) {
	res, err := Layout([]byte("abc\rdef"), Options{Width: 38})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 2 || res.Lines[0].String() != "abc" || res.Lines[1].String() != "def" {
		t.Fatalf("換行結果：%q", linesOf(res))
	}

	// 硬斷是暫代行為（docs/spec/03 §3）——這個測試釘住它，換成原版規則時會亮。
	res, err = Layout([]byte("abcdef"), Options{Width: 4})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 2 || res.Lines[0].String() != "abcd" || res.Lines[1].String() != "ef" {
		t.Fatalf("硬斷結果：%q", linesOf(res))
	}
}

// docs/re/28：三個變形碼是同一套機制，段數必須與原版一致。
func TestVariants(t *testing.T) {
	// 單複數：hammer\ns\n\n → 單數 hammers、複數 hammer
	text := []byte("hammer\ns\n\n!")
	for _, c := range []struct {
		count int
		want  string
	}{{1, "hammers!"}, {3, "hammer!"}} {
		res, err := Layout(text, Options{Width: 38, Count: c.count})
		if err != nil {
			t.Fatal(err)
		}
		if got := res.Lines[0].String(); got != c.want {
			t.Fatalf("數量 %d 時排出 %q，應該是 %q", c.count, got, c.want)
		}
	}

	// 三選一：him／her／it
	pron := []byte("hits \x0ehim\x0eher\x0eit\x0e now")
	pron = []byte{'h', 'i', 't', 's', ' ',
		CodePronoun, 'h', 'i', 'm', CodePronoun, 'h', 'e', 'r', CodePronoun, 'i', 't', CodePronoun,
		' ', 'n', 'o', 'w'}
	for i, want := range []string{"hits him now", "hits her now", "hits it now"} {
		res, err := Layout(pron, Options{Width: 38, Pronoun: i})
		if err != nil {
			t.Fatal(err)
		}
		if got := res.Lines[0].String(); got != want {
			t.Fatalf("代名詞 %d 排出 %q，應該是 %q", i, got, want)
		}
	}

	// 性別二選一
	gender := []byte{CodeGender, 'h', 'i', 's', CodeGender, 'h', 'e', 'r', CodeGender, ' ', 'g', 'u', 'n'}
	for g, want := range map[int]string{0: "his gun", 1: "her gun"} {
		res, err := Layout(gender, Options{Width: 38, Gender: g})
		if err != nil {
			t.Fatal(err)
		}
		if got := res.Lines[0].String(); got != want {
			t.Fatalf("性別 %d 排出 %q，應該是 %q", g, got, want)
		}
	}

	// 數量：0x0F 印出數字，與 0x0A 共用同一個選擇子
	res, err := Layout([]byte{'h', 'i', 't', 's', ' ', CodeCount, ' ', 'o', 'f'}, Options{Width: 38, Count: 12})
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Lines[0].String(); got != "hits 12 of" {
		t.Fatalf("數量插入排出 %q", got)
	}
}

// 戰鬥訊息的完整例子（docs/re/28 §1）。
func TestCombatMessageVariants(t *testing.T) {
	var text []byte
	text = append(text, []byte(" sending ")...)
	text = append(text, CodePlural)
	text = append(text, CodePronoun, 'h', 'i', 'm', CodePronoun, 'h', 'e', 'r', CodePronoun, 'i', 't', CodePronoun)
	text = append(text, CodePlural)
	text = append(text, CodeCount)
	text = append(text, []byte(" of them")...)
	text = append(text, CodePlural)
	text = append(text, []byte(" to meet ")...)
	text = append(text, CodePlural)
	text = append(text, CodePronoun, 'h', 'i', 's', CodePronoun, 'h', 'e', 'r', CodePronoun, 'i', 't', 's', CodePronoun)
	text = append(text, CodePlural)
	text = append(text, []byte("their")...)
	text = append(text, CodePlural)
	text = append(text, []byte(" maker")...)

	single, err := Layout(text, Options{Width: 80, Count: 1, Pronoun: 0})
	if err != nil {
		t.Fatal(err)
	}
	if got := single.Lines[0].String(); got != " sending him to meet his maker" {
		t.Fatalf("單數版排出 %q", got)
	}
	plural, err := Layout(text, Options{Width: 80, Count: 3, Pronoun: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got := plural.Lines[0].String(); got != " sending 3 of them to meet their maker" {
		t.Fatalf("複數版排出 %q", got)
	}
}

func TestInsertName(t *testing.T) {
	res, err := Layout([]byte{CodeInsertName, ' ', 'i', 's'}, Options{
		Width: 38,
		Name:  func() string { return "SNAKE" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Lines[0].String(); got != "SNAKE is" {
		t.Fatalf("插入名字後是 %q", got)
	}
}

func TestPaginate(t *testing.T) {
	lines := make([]Line, 14)
	pages := Paginate(lines, 6)
	if len(pages) != 3 || len(pages[2]) != 2 {
		t.Fatalf("分頁結果：%d 頁，最後一頁 %d 行", len(pages), len(pages[len(pages)-1]))
	}
}

func linesOf(r Result) []string {
	out := make([]string, len(r.Lines))
	for i, l := range r.Lines {
		out[i] = l.String()
	}
	return out
}
