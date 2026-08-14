package play

import (
	"os"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

func openRom(t *testing.T) *assets.Rom {
	t.Helper()
	dir := os.Getenv("WL_DATA")
	if dir == "" {
		dir = "../../workplace/orig/wastland"
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("找不到原版資料目錄 %s，跳過", dir)
	}
	rom, err := assets.Open(dir)
	if err != nil {
		t.Skipf("開啟原版資料失敗：%v", err)
	}
	if err := rom.LoadImage("../../workplace/analysis/unpacked/wl.merged.exe"); err != nil {
		t.Skipf("載入解包映像失敗：%v", err)
	}
	return rom
}

// 從出廠存檔開場：隊伍、座標、地圖都要對得上 docs/spec/05 的驗收。
func TestNewFromShippedSave(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	w := s.World()
	if len(w.Party.Members) != 4 {
		t.Fatalf("出廠應該有四個隊員，得到 %d", len(w.Party.Members))
	}
	if w.Party.X != 55 || w.Party.Y != 62 {
		t.Fatalf("出廠座標應該是 (55, 62)，得到 (%d, %d)", w.Party.X, w.Party.Y)
	}
	if w.Party.Members[0].Name != "Hell Razor" {
		t.Fatalf("第一個隊員應該是 Hell Razor，得到 %q", w.Party.Members[0].Name)
	}
	// 出廠時鐘是 01:00（docs/re/30 §5），從全域狀態的副本讀出來。
	if w.Clock.Hour != 1 || w.Clock.Minute != 0 {
		t.Fatalf("出廠時鐘應該是 01:00，得到 %02d:%02d", w.Clock.Hour, w.Clock.Minute)
	}
}

// 真的走得動：按方向鍵會改座標、推進時鐘，而且畫得出一幀。
func TestWalkAdvancesWorld(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	w := s.World()
	start := [2]uint8{w.Party.X, w.Party.Y}
	before := w.Clock

	moved := 0
	dirs := []input.Direction{input.DirUp, input.DirDown, input.DirLeft, input.DirRight}
	for i := 0; i < 200; i++ {
		cur := [2]uint8{w.Party.X, w.Party.Y}
		ok, err := s.Update(input.Input{Dir: dirs[i%len(dirs)]})
		if err != nil {
			t.Fatalf("走一步失敗：%v", err)
		}
		if !ok {
			t.Fatal("方向鍵不該讓場景結束")
		}
		if [2]uint8{w.Party.X, w.Party.Y} != cur {
			moved++
		}
	}
	if moved == 0 {
		t.Fatal("200 次方向鍵一步都沒走成，接線一定有問題")
	}
	if [2]uint8{w.Party.X, w.Party.Y} == start && moved > 0 {
		t.Log("走了一圈又回到起點，這是四個方向輪流的正常結果")
	}
	if w.Clock == before {
		t.Fatal("走了路時鐘卻沒動")
	}

	f := s.Frame()
	if f == nil {
		t.Fatal("畫不出一幀")
	}
	// 畫面不能整片是同一個顏色——那代表地圖沒畫上去。
	first := f.At(0, 0)
	same := true
	for y := 0; y < 200 && same; y++ {
		for x := 0; x < 320; x++ {
			if f.At(x, y) != first {
				same = false
				break
			}
		}
	}
	if same {
		t.Fatal("整幀只有一種顏色，地圖沒畫出來")
	}
}

// F10 離開。
func TestQuitStopsScene(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	ok, err := s.Update(input.Input{Action: input.ActionQuit})
	if err != nil || ok {
		t.Fatalf("F10 應該讓場景結束：ok=%v err=%v", ok, err)
	}
}
