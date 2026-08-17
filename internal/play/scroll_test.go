package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// 文字比區域高的時候要**捲掉最前面的行**，不是切掉後面的（`docs/re/106` §1）。
//
// ⚠ 兩種寫法在「只印一則訊息」時看起來完全一樣，
// 要到第二則接上去才分得出來。
func TestMessageCellsScrollInsteadOfClipping(t *testing.T) {
	rect := textRect{Col: 15, Row: 1, Width: 24, Height: 3}
	var lines []string
	for _, s := range []string{"one", "two", "three", "four", "five"} {
		lines = append(lines, s)
	}
	text := strings.Join(lines, "\r")

	seen := map[int]string{}
	eachMessageCell(text, rect, rect.Row, func(col, row int, r rune) {
		seen[row] += string(r)
	})
	// 三列的區域要看到最後三行。
	want := map[int]string{1: "three", 2: "four", 3: "five"}
	for row, w := range want {
		if seen[row] != w {
			t.Errorf("列 %d 得到 %q，預期 %q（全部：%v）", row, seen[row], w, seen)
		}
	}
	if len(seen) != len(want) {
		t.Errorf("畫了 %d 列，預期 %d：%v", len(seen), len(want), seen)
	}
}

// 放得下的時候一列都不准動——捲動只在滿了才發生。
func TestMessageCellsDoNotScrollWhenItFits(t *testing.T) {
	rect := textRect{Col: 15, Row: 1, Width: 24, Height: 5}
	seen := map[int]string{}
	eachMessageCell("aa\rbb", rect, rect.Row, func(col, row int, r rune) {
		seen[row] += string(r)
	})
	if seen[1] != "aa" || seen[2] != "bb" {
		t.Errorf("放得下就不該捲，得到 %v", seen)
	}
}

// 點擊命中判定與繪製走同一支走訪，所以**捲動之後點到的還是看到的那一格**
// （`docs/re/106` §5：原版要另外搬每列熱鍵表，remake 因為共用走訪不必）。
func TestScrolledPanelStaysClickable(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	// 問完第一個人，第二個人的提示接上去 → 面板滿了 → 捲動。
	step(t, s, input.Input{Dir: input.DirNone, Char: 'A'})

	rect := s.msgRect()
	// 面板裡每一格「看得到的字」都要能被 charAt 找回來。
	drawn := map[[2]int]byte{}
	start := rect.Row
	if s.message != "" || len(s.journalHead) > 0 {
		start++
	}
	if len(s.cjk) > 0 {
		eachMessageCell(s.cjk, rect, start, func(col, row int, r rune) {
			if r > ' ' && r < 0x80 {
				drawn[[2]int{col, row}] = byte(r)
			}
		})
	} else {
		eachMessageCell(s.message, rect, rect.Row, func(col, row int, r rune) {
			if r > ' ' && r < 0x80 {
				drawn[[2]int{col, row}] = byte(r)
			}
		})
	}
	if len(drawn) == 0 {
		t.Fatal("面板上一個字都沒有")
	}
	for at, want := range drawn {
		if got := s.charAt(at[0], at[1]); got != want {
			t.Fatalf("欄 %d 列 %d：畫的是 %q，點到的是 %q", at[0], at[1], want, got)
		}
	}
}

// 第二個人的選單要**整份**看得到——這是 `docs/re/106` 那一輪的起因。
func TestSecondMemberSeesTheWholeMenu(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'A'})

	msg := s.Message()
	if msg == "" {
		t.Skip("這一輪走的是中文那條路")
	}
	rect := s.msgRect()
	rows := map[int]string{}
	eachMessageCell(msg, rect, rect.Row, func(col, row int, r rune) {
		rows[row] += string(r)
	})
	var body []string
	for r := rect.Row; r <= rect.LastRow(); r++ {
		body = append(body, rows[r])
	}
	joined := strings.Join(body, "\n")
	// ⚠ **要認「輪到的那個人」，不能只找 `, choose:`**——
	// 面板上還留著前一個人的提示，找第一個就會拿他的選單去驗，
	// 而那一份本來就是完整的（假綠）。
	c := s.combat
	if c.Turn < 0 || c.Turn >= len(c.Battle.Party.Members) {
		t.Skip("指令階段已經結束")
	}
	m := c.Battle.Party.Members[c.Turn]
	if m == nil {
		t.Skip("輪到的格子是空的")
	}
	i := strings.Index(joined, m.Name+", choose:")
	if i < 0 {
		t.Fatalf("面板上沒有 %s 的提示：\n%s", m.Name, joined)
	}
	tail := joined[i:]
	for _, opt := range []string{"Run", "Use", "Hire", "Evade", "Attack", "Weapon", "Load"} {
		if !strings.Contains(tail, opt) {
			t.Errorf("選項 %q 被擠出面板了：\n%s", opt, joined)
		}
	}
	_ = render.PanelHeight
}
