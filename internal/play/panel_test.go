package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// 戰鬥的文字走**面板**（欄 15–38、列 1–13），不是訊息視窗（`docs/re/105` §2）。
//
// ⚠ 兩塊的判斷條件只有一個地方（`msgRect`）——繪製與滑鼠命中判定共用它。
// 分成兩份的話戰鬥中點選單會全部落空，而畫面看起來完全正常。
func TestCombatTextGoesToThePanel(t *testing.T) {
	s := newScene(t)
	if got := s.msgRect(); got.Row != render.MsgRow {
		t.Fatalf("地圖模式應該用訊息視窗，得到 %+v", got)
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Fatal("沒進戰鬥")
	}
	r := s.msgRect()
	if r.Col != render.PanelCol || r.Row != render.PanelRow ||
		r.Width != render.PanelWidth || r.Height != render.PanelHeight {
		t.Errorf("戰鬥應該用面板，得到 %+v", r)
	}
	// 13 列放得下「名字, choose:」＋ 七個選項。
	if r.Height < 8 {
		t.Errorf("面板只有 %d 列，八行放不下", r.Height)
	}
}

// 指令選單**一行一個選項**（原版的結構，`docs/re/40` §4）。
//
// ⚠ 以前擠成一行是因為訊息視窗只有 6 列；搬到面板之後那個理由不成立了。
func TestCombatMenuIsOnePerLine(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	msg := s.Message()
	if msg == "" {
		t.Skip("這一輪走的是中文那條路，英文訊息是空的")
	}
	lines := strings.Split(strings.ReplaceAll(msg, "\n", "\r"), "\r")
	if len(lines) < 8 {
		t.Fatalf("標題 ＋ 七個選項應該是八行，得到 %d：%q", len(lines), msg)
	}
	if !strings.HasSuffix(lines[0], ", choose:") {
		t.Errorf("第一行應該是「名字, choose:」，得到 %q", lines[0])
	}
	// 每一行的第一個字元就是熱鍵（`\x10` 捕捉的那一個，`docs/re/40` §4.1）。
	for i, want := range []byte{'R', 'U', 'H', 'E', 'A', 'W', 'L'} {
		if got := lines[i+1]; len(got) == 0 || got[0] != want {
			t.Errorf("第 %d 個選項應該以 %q 開頭，得到 %q", i+1, want, got)
		}
	}
	// 每一行都要放得進面板的寬度。
	for _, l := range lines {
		if len(l) > render.PanelWidth {
			t.Errorf("這一行 %d 格，超過面板的 %d：%q", len(l), render.PanelWidth, l)
		}
	}
}

// 戰鬥時名單排得下**整隊**。
//
// ⚠ 照地圖那條 `MsgRow-1` 算的話只放得下三個人，**第四個隊員直接看不到**——
// 而畫面上看起來只像「隊伍只有三個人」，不像 bug。
func TestCombatRosterShowsEveryone(t *testing.T) {
	s := newScene(t)
	n := 0
	for _, m := range s.World().Party.Members {
		if m != nil {
			n++
		}
	}
	if n < 4 {
		t.Skipf("出廠隊伍只有 %d 個人，測不出差別", n)
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})

	last := s.rosterLastRow()
	rows := render.RosterHeaderRow + 1 + n - 1
	if rows > last {
		t.Errorf("%d 個人要排到列 %d，而最後一列是 %d", n, rows, last)
	}
	// 地圖模式仍然停在訊息視窗上緣——那一塊要留給訊息。
	s.combat = nil
	if got := s.rosterLastRow(); got != render.MsgRow-1 {
		t.Errorf("地圖模式的名單最後一列是 %d，預期 %d", got, render.MsgRow-1)
	}
}

// 名單一行是 39 欄，行首是序號 ＋ `>`（`docs/re/103` §4 的實機截圖）。
func TestRosterLineWidthAndIndex(t *testing.T) {
	r := RosterRow{Index: 1, Name: "Hell Razor", AC: "0", Ammo: "0",
		MaxCON: "28", CON: "28", Weapon: "Crowbar"}
	line := r.Text()
	if len(line) != 39 {
		t.Fatalf("一行應該是 39 欄，得到 %d", len(line))
	}
	if !strings.HasPrefix(line, "1>") {
		t.Errorf("行首應該是序號 ＋ `>`，得到 %q", line[:4])
	}
	// ⚠ 武器欄從 0x20 起，39 − 32 ＝ 7 格：`Crowbar` 剛好整個放得下。
	// 少一格會切成 `Crowba`，而那看起來只像「名字比較長」。
	if got := strings.TrimSpace(line[colWeapon:]); got != "Crowbar" {
		t.Errorf("武器欄得到 %q，預期 Crowbar", got)
	}
	_ = game.WoundDead
}
