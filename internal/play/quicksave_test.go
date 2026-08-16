package play

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestQuickSaveRoundTrip：F5 存、走遠一點、F9 讀回來，狀態要回到存的那一刻。
//
// ⚠ 快速存檔**寫的是自己的檔**，一個 byte 都不碰玩家的原版資料目錄
// （那份驗 SHA-256，寫過就開不起來）。這一條順便釘住這件事：
// 測試給的路徑在 t.TempDir() 底下。
func TestQuickSaveRoundTrip(t *testing.T) {
	s := newScene(t)
	path := filepath.Join(t.TempDir(), "quick.wlq")
	s.SetQuickSavePath(path)

	w := s.World()
	step(t, s, input.Input{Dir: input.DirUp})
	savedX, savedY := w.Party.X, w.Party.Y
	savedTotal := w.Clock.Total

	if _, err := s.Update(input.Input{Dir: input.DirNone, Fn: input.FnQuickSave}); err != nil {
		t.Fatalf("F5：%v", err)
	}
	if st, err := os.Stat(path); err != nil || st.Size() == 0 {
		t.Fatalf("快速存檔沒寫出來：%v", err)
	}

	// 走開，狀態要真的變了，否則下面驗不出讀檔有沒有作用。
	for i := 0; i < 5; i++ {
		step(t, s, input.Input{Dir: input.DirUp})
	}
	w = s.World()
	if w.Party.X == savedX && w.Party.Y == savedY && w.Clock.Total == savedTotal {
		t.Fatal("走了五步狀態沒變，這一條測不到讀檔")
	}

	if _, err := s.Update(input.Input{Dir: input.DirNone, Fn: input.FnQuickLoad}); err != nil {
		t.Fatalf("F9：%v", err)
	}
	w = s.World()
	if w.Party.X != savedX || w.Party.Y != savedY {
		t.Errorf("讀回來在 (%d,%d)，預期 (%d,%d)", w.Party.X, w.Party.Y, savedX, savedY)
	}
	if w.Clock.Total != savedTotal {
		t.Errorf("讀回來的時鐘 %d，預期 %d", w.Clock.Total, savedTotal)
	}
	if s.Mode() != "map" {
		t.Errorf("讀完檔停在 %s，預期回到地圖模式", s.Mode())
	}
}

// TestQuickLoadRejectsGarbage：讀到不是快速存檔的東西要說出來，不能把世界弄壞。
func TestQuickLoadRejectsGarbage(t *testing.T) {
	s := newScene(t)
	path := filepath.Join(t.TempDir(), "junk.wlq")
	if err := os.WriteFile(path, []byte("this is not a save"), 0o644); err != nil {
		t.Fatal(err)
	}
	s.SetQuickSavePath(path)
	before := s.World().Party.X

	if _, err := s.Update(input.Input{Dir: input.DirNone, Fn: input.FnQuickLoad}); err != nil {
		t.Fatalf("F9 不該回錯誤（訊息要留在畫面上）：%v", err)
	}
	if s.World().Party.X != before {
		t.Error("讀了壞檔卻動到了世界")
	}
	if got := s.Message(); got == "" {
		t.Error("讀壞檔沒有任何訊息")
	}
}

// TestF10SavesBeforeQuitting：鐵則 4 —— 選 Y 之後**先存檔才退出**。
func TestF10SavesBeforeQuitting(t *testing.T) {
	s := newScene(t)
	path := filepath.Join(t.TempDir(), "quit.wlq")
	s.SetQuickSavePath(path)
	step(t, s, input.Input{Dir: input.DirUp})

	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionQuit})
	ok, err := s.Update(input.Input{Dir: input.DirNone, Char: 'Y'})
	if err != nil || ok {
		t.Fatalf("選 Y 應該離開：ok=%v err=%v", ok, err)
	}
	if st, err := os.Stat(path); err != nil || st.Size() == 0 {
		t.Fatalf("離開前沒有存檔：%v", err)
	}
}
