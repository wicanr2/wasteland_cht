package play

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// TestHelpPanelOpensAndCloses：F1 叫得出說明，任何鍵關掉，關掉之後方向鍵照走。
//
// 面板**吃掉所有按鍵**是刻意的（與離開確認同一個道理）：說明開著的時候
// 按方向鍵應該是關面板，不是一邊看說明一邊走路。
func TestHelpPanelOpensAndCloses(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Fn: input.FnHelp})
	if s.Mode() != "help" {
		t.Fatalf("F1 之後在 %s，預期 help", s.Mode())
	}
	if s.Message() == "" && len(s.CJK()) == 0 {
		t.Error("說明面板沒有任何內容")
	}
	// 沒有按鍵的那一幀不該把面板關掉，否則開起來的下一幀就沒了。
	step(t, s, input.Input{Dir: input.DirNone})
	if s.Mode() != "help" {
		t.Fatalf("空白幀把說明關掉了")
	}
	before := s.World().Party.Y
	step(t, s, input.Input{Dir: input.DirUp})
	if s.Mode() != "map" {
		t.Fatalf("按鍵之後在 %s，預期回到 map", s.Mode())
	}
	if s.World().Party.Y != before {
		t.Error("關面板那一下順便走了一步")
	}
}

// TestSettingsTogglesMusic：F2 進設定，M 開關音樂、+／- 調音量、ESC 關閉。
//
// ⚠ ESC 在這裡是**關面板**。這一條同時釘住鐵則 1：任何一層的 ESC 都不結束遊戲。
func TestSettingsTogglesMusic(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Fn: input.FnSettings})
	if s.Mode() != "settings" {
		t.Fatalf("F2 之後在 %s，預期 settings", s.Mode())
	}
	if on, _ := s.MusicSetting(); !on {
		t.Fatal("音樂預設應該是開的")
	}

	step(t, s, input.Input{Dir: input.DirNone, Char: 'M'})
	if on, _ := s.MusicSetting(); on {
		t.Error("按 M 沒有關掉音樂")
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'm'})
	if on, _ := s.MusicSetting(); !on {
		t.Error("小寫 m 應該與大寫一樣")
	}

	_, vol0 := s.MusicSetting()
	step(t, s, input.Input{Dir: input.DirNone, Char: '-'})
	if _, v := s.MusicSetting(); v != vol0-1 {
		t.Errorf("按 - 之後音量 %d，預期 %d", v, vol0-1)
	}
	// 下限夾住：一直按不會變成負數。
	for i := 0; i < 20; i++ {
		step(t, s, input.Input{Dir: input.DirNone, Char: '-'})
	}
	if _, v := s.MusicSetting(); v != 0 {
		t.Errorf("按到底音量 %d，預期 0", v)
	}
	for i := 0; i < 20; i++ {
		step(t, s, input.Input{Dir: input.DirNone, Char: '+'})
	}
	if _, v := s.MusicSetting(); v != 10 {
		t.Errorf("按到頂音量 %d，預期 10", v)
	}

	ok, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	if err != nil || !ok {
		t.Fatalf("設定裡的 ESC 結束了遊戲：ok=%v err=%v", ok, err)
	}
	if s.Mode() != "map" {
		t.Errorf("ESC 之後在 %s，預期回到 map", s.Mode())
	}
}

// TestHelpFitsTheMessageWindow：說明面板要塞得進訊息視窗的六行，
// 而且字母欄只能是 ASCII。
//
// 兩條都不會噴錯，只會安靜地壞掉：超出的行被切掉、中文字母欄變成 Big5 亂碼。
func TestHelpFitsTheMessageWindow(t *testing.T) {
	const messageRows = 6 // `docs/re/25` §2：訊息視窗六行
	if n := len(helpLines) + 1; n > messageRows {
		t.Errorf("說明有 %d 行（含標題），訊息視窗只有 %d 行，多的會被切掉",
			n, messageRows)
	}
	for _, l := range helpLines {
		for i := 0; i < len(l.key); i++ {
			if l.key[i] >= 0x80 {
				t.Errorf("字母欄 %q 不是 ASCII——接進 Big5 串會變亂碼", l.key)
				break
			}
		}
	}
	// 說明文件與程式碼分兩處寫就會漂：這一條讓功能鍵被拿掉時有人會紅。
	joined := strings.Join(func() (k []string) {
		for _, l := range helpLines {
			k = append(k, l.key)
		}
		return
	}(), "|")
	for _, w := range []string{"F1", "F2", "F5", "F9", "F10", "ESC"} {
		if !strings.Contains(joined, w) {
			t.Errorf("說明沒有列出 %s（實際列了 %s）", w, joined)
		}
	}
}

// TestTitleScreenHasNoCommandBar：標題畫面不畫地圖的指令列。
//
// `Frame` 在標題那一支提早 return，`HiFrame` 卻是照模式旗標判斷——
// 兩邊的條件不一致時，中文畫面會在標題上多出一整條指令列，
// 還會蓋掉同一列的 `Start`（`docs/re/95`）。**低解畫面完全正常**，
// 所以這個 bug 只在有倚天字型的時候看得到。
func TestTitleScreenHasNoCommandBar(t *testing.T) {
	s := newScene(t)
	dir := os.Getenv("WL_ETEN")
	if dir == "" {
		dir = "../../workplace/eten"
	}
	if err := s.LoadFont(dir); err != nil {
		t.Skipf("沒有倚天字型（%v），這一條驗不到", err)
	}
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("沒有翻譯目錄（%v），這一條驗不到", err)
	}
	s.BeginTitle()

	// 指令列那一列（字元列 24）在高解畫面上是 y ∈ [24×16, 25×16)。
	// `Start` 只占最左邊幾格，右半邊必須全黑。
	on := 0
	for y := render.CmdRow * 16; y < (render.CmdRow+1)*16 && y < render.HiScreenHeight; y++ {
		for x := 12 * 16; x < render.HiScreenWidth; x++ {
			if s.HiFrame().At(x, y) != 0 {
				on++
			}
		}
	}
	if on != 0 {
		t.Errorf("標題畫面的指令列那一列右半邊有 %d 個像素——指令列畫上去了", on)
	}
}
