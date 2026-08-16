package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// openZhScene 開一個接上翻譯目錄的場景（沒有字型也沒關係，
// 這一組驗的是「中文有沒有查出來」，不是「畫得出來」）。
func openZhScene(t *testing.T) *Scene {
	t.Helper()
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("載不到翻譯目錄：%v", err)
	}
	return s
}

// TestPlayerMessagesAreTranslated：**玩家一路上會看到的訊息都要走出中文**。
//
// 譯文早就在目錄裡了，缺的一直是查表——而「沒查表」的症狀只是畫面上是英文，
// 看起來像「這句還沒翻」。每一項都是一條實際的玩家動作，
// 走完之後 `CJK()` 必須非空。
func TestPlayerMessagesAreTranslated(t *testing.T) {
	t.Run("進新地點的問句", func(t *testing.T) {
		s := openZhScene(t)
		if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Update(input.Input{Dir: input.DirDown}); err != nil {
			t.Fatal(err)
		}
		if !s.Asking() {
			t.Fatalf("沒有停在確認上，訊息是 %q", s.Message())
		}
		mustCJK(t, s, "Enter new location?")
	})

	t.Run("角色管理選單", func(t *testing.T) {
		s := openZhScene(t)
		for _, in := range []input.Input{
			{Dir: input.DirUp}, {Dir: input.DirDown},
			{Dir: input.DirNone, Char: 'Y'},
		} {
			if _, err := s.Update(in); err != nil {
				t.Fatal(err)
			}
		}
		if s.Mode() != "roster" {
			t.Fatalf("沒有進角色管理，模式是 %q", s.Mode())
		}
		mustCJK(t, s, "CREATE DELETE PLAY")
	})

	t.Run("商店主選單", func(t *testing.T) {
		s := openZhScene(t)
		if err := s.LoadMap(10, 30, 25); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'Y'}); err != nil {
			t.Fatal(err)
		}
		if !s.InFacility() {
			t.Fatalf("沒有進商店，訊息是 %q", s.Message())
		}
		// 設施畫面的中文不在訊息視窗，在自己的那幾行。
		zh := s.Facility().CJKLines
		if len(zh) == 0 {
			t.Fatal("設施沒有任何中文行")
		}
		got := 0
		for _, l := range zh {
			if len(l) > 0 {
				got++
			}
		}
		// 第一行是地點名（來自存檔資料，另一個題目），其餘都該有中文。
		if got < len(zh)-1 {
			t.Errorf("設施 %d 行裡只有 %d 行有中文：%q", len(zh), got, s.Facility().Lines)
		}
	})

	t.Run("USE 的結果那一句", func(t *testing.T) {
		s := openZhScene(t)
		for _, in := range []input.Input{
			{Dir: input.DirNone, Char: 'U'},
			{Dir: input.DirNone, Char: '1'},
			{Dir: input.DirNone, Char: 'S'},
			{Dir: input.DirNone, Char: '1'},
		} {
			if _, err := s.Update(in); err != nil {
				t.Fatal(err)
			}
		}
		mustCJK(t, s, "X uses Y")
	})

	t.Run("存檔訊息", func(t *testing.T) {
		s := openZhScene(t)
		if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'S'}); err != nil {
			t.Fatal(err)
		}
		mustCJK(t, s, "Game state updated")
	})

	t.Run("戰鬥的指令選單", func(t *testing.T) {
		s := openZhScene(t)
		if err := s.LoadMap(0, 12, 2); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 400 && !s.InCombat(); i++ {
			d := input.DirRight
			if i%2 == 1 {
				d = input.DirLeft
			}
			if _, err := s.Update(input.Input{Dir: d}); err != nil {
				t.Fatal(err)
			}
		}
		if !s.InCombat() {
			t.Skip("走了 400 步沒遇到敵人，這一輪測不到")
		}
		mustCJK(t, s, "遭到攻擊 ＋ 指令選單")
	})
}

func mustCJK(t *testing.T, s *Scene, what string) {
	t.Helper()
	if len(s.CJK()) == 0 {
		t.Errorf("「%s」沒有走出中文，畫面上會是英文：%q", what, s.Message())
	}
}
