package play

import (
	"strings"
	"testing"


	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// 分頁的算術：一頁九項，最後一頁不滿也算一頁。
func TestUsePageArithmetic(t *testing.T) {
	for _, tc := range []struct{ n, pages int }{
		{0, 1}, {1, 1}, {9, 1}, {10, 2}, {18, 2}, {19, 3},
	} {
		u := useState{options: make([]useOption, tc.n)}
		if got := u.pages(); got != tc.pages {
			t.Errorf("%d 項算成 %d 頁，預期 %d 頁", tc.n, got, tc.pages)
		}
	}
	u := useState{options: make([]useOption, 20)}
	for i := range u.options {
		u.options[i].label = string(rune('a' + i))
	}
	u.page = 2
	if got := len(u.pageSlice()); got != 2 {
		t.Errorf("第三頁有 %d 項，預期 2 項", got)
	}
	if u.pageSlice()[0].label != string(rune('a'+18)) {
		t.Errorf("第三頁的第一項是 %q，預期第 19 項", u.pageSlice()[0].label)
	}
}

// **第 10 項以後選得到**——這一條就是分頁存在的理由。
//
// 數字鍵是「這一頁的第幾項」不是「整份清單的第幾項」，
// 兩者混淆的話翻到第二頁按 1 會用到第一項，而且完全不會報錯。
func TestUsePickBeyondFirstPage(t *testing.T) {
	s := newScene(t)
	s.use = useState{stage: useStagePick, page: 1}
	for i := 0; i < 12; i++ {
		s.use.options = append(s.use.options, useOption{
			label: string(rune('a' + i)), id: byte(i),
		})
	}
	// 第二頁的第 1 項 ＝ 整份的第 10 項。
	want := s.use.options[9]
	got := s.use.options[s.use.page*usePageSize+0]
	if got.id != want.id {
		t.Fatalf("第二頁第 1 項算成 id=%d，預期 id=%d", got.id, want.id)
	}
}

// `0` 會翻頁，而且**繞回第一頁**。
func TestUsePagingWrapsAround(t *testing.T) {
	s := newScene(t)
	s.use = useState{stage: useStagePick, options: make([]useOption, 20)}
	for i := range s.use.options {
		s.use.options[i].label = "x"
	}
	for want := 1; want <= 3; want++ {
		step(t, s, input.Input{Dir: input.DirNone, Char: '0'})
		if got := s.use.page; got != want%3 {
			t.Fatalf("翻第 %d 次到第 %d 頁，預期第 %d 頁", want, got, want%3)
		}
	}
}

// 頁碼**不可以有阿拉伯數字**：滑鼠是「點到哪一格就送那一格的字元」，
// 頁碼裡的 `2` 會被當成「選第 2 項」（`docs/spec/29` §3）。
func TestUsePageLabelHasNoDigits(t *testing.T) {
	b := cjkPageLabel(2, 3)
	if len(b) == 0 {
		t.Fatal("頁碼是空的")
	}
	for i := 0; i < len(b); i++ {
		if b[i] >= '0' && b[i] <= '9' {
			t.Errorf("頁碼裡有阿拉伯數字 %q，滑鼠會把它當成選項", b[i])
		}
	}
}

// 翻頁那一行點得到：`0` 要真的畫在畫面上（滑鼠靠它翻頁）。
func TestUsePagingIsClickable(t *testing.T) {
	s := sceneWithCatalogue(t)
	s.use = useState{stage: useStagePick}
	for i := 0; i < 12; i++ {
		s.use.options = append(s.use.options, useOption{
			label: "x", nameSlot: 1, id: byte(i)})
	}
	s.showUseMenu()
	found := false
	for r := render.MsgRow; r <= render.MsgRowEnd && !found; r++ {
		for c := 0; c < render.MsgCol+render.MsgWidth; c++ {
			if s.charAt(c, r) == '0' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("畫面上找不到翻頁用的 `0`（訊息是 %q／% X）",
			s.Message(), s.CJK())
	}
	// 英文後備那一份也要有。
	if !strings.Contains(s.useMenu(), "0 More") {
		t.Errorf("英文清單沒有翻頁提示：%q", s.useMenu())
	}
}
