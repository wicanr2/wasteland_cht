package textlayout

import "testing"

func TestControlCodesNeverPrint(t *testing.T) {
	text := []byte{
		CodeInverseOn, 'A', 'B', CodeInverseOff,
		CodeMoveTo, 0x41, // ⚠ 0x41 是參數不是文字
		CodeUnknown07,
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

	var moveTo, unknown int
	for _, e := range res.Events {
		switch e.Kind {
		case EventMoveTo:
			moveTo++
			if e.Param != 0x41 {
				t.Fatalf("0x09 的參數是 %#x，應該是 0x41", e.Param)
			}
		case EventUnknownCode:
			unknown++
			if e.Code != CodeUnknown07 {
				t.Fatalf("未解控制碼回報成 %#x", e.Code)
			}
		}
	}
	if moveTo != 1 || unknown != 1 {
		t.Fatalf("事件數不對：moveTo=%d unknown=%d", moveTo, unknown)
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
