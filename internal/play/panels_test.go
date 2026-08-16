package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
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

// TestMusicTrackFollowsMode：曲子跟著場景走。
//
// **這張對照表是重製版的決定，不是逆向結論**——原版沒有背景音樂
// （九首 PC 喇叭音效，`docs/re/44`）。所以這一條驗的是我們自己訂的規則，
// 不是原版行為。
func TestMusicTrackFollowsMode(t *testing.T) {
	s := newScene(t)
	if got := s.MusicTrack(); got != "desert" {
		t.Errorf("地圖上放 %q，預期 desert", got)
	}
	s.BeginTitle()
	if got := s.MusicTrack(); got != "theme" {
		t.Errorf("標題畫面放 %q，預期 theme", got)
	}
}

// TestHelpListsTheKeysItClaims：說明列的按鍵字母要與實際收的鍵一致。
//
// 說明文件與程式碼分兩處寫就會漂：這一條讓「F5／F9 快速存讀檔」這種
// 敘述被改掉時有人會紅。
func TestHelpListsTheKeysItClaims(t *testing.T) {
	want := []string{"F1", "F2", "F5 / F9", "F10", "ESC"}
	var keys []string
	for _, l := range helpLines {
		keys = append(keys, l.key)
	}
	joined := strings.Join(keys, "|")
	for _, w := range want {
		if !strings.Contains(joined, w) {
			t.Errorf("說明沒有列出 %s（實際列了 %s）", w, joined)
		}
	}
}
