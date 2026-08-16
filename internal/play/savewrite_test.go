package play

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestSaveCommandWritesToDisk：按 `S` 之後，**下一次開遊戲要接得上**。
//
// 抓到的錯是 `cmdSave` 只呼叫 `StoreTo`（改記憶體裡的明文）就報
// 「Game saved.」——測試全綠、畫面也說存好了，而檔案一個 byte 都沒動。
// 這種謊只有真的關掉再開才看得見，所以門檻要走完整條路：
// 寫檔 → 重新開一份 Rom → 從存檔開場 → 位置與時鐘對得上。
func TestSaveCommandWritesToDisk(t *testing.T) {
	src := os.Getenv("WL_DATA")
	if src == "" {
		src = "../../workplace/orig/wastland"
	}
	if _, err := os.Stat(src); err != nil {
		t.Skipf("找不到原版資料目錄 %s，跳過", src)
	}
	const image = "../../workplace/analysis/unpacked/wl.merged.exe"

	// 原版目錄是唯讀的（`CLAUDE.md` §4：不覆蓋原版資料），複製一份出來寫。
	dir := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	rom, err := assets.Open(dir)
	if err != nil {
		t.Skipf("開啟複製出來的資料目錄失敗：%v", err)
	}
	if err := rom.LoadImage(image); err != nil {
		t.Skipf("載入解包映像失敗：%v", err)
	}
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	s.SetSaveDir(dir)

	for i := 0; i < 3; i++ {
		if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
			t.Fatalf("第 %d 步：%v", i+1, err)
		}
	}
	wantX, wantY := s.World().Party.X, s.World().Party.Y
	wantClock := s.World().Clock
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'S'}); err != nil {
		t.Fatalf("按 S：%v", err)
	}
	if got := s.Message(); got != "Game saved." {
		t.Fatalf("存檔訊息是 %q，預期 \"Game saved.\"", got)
	}

	// 重新開一份——寫過之後雜湊當然對不上，走 OpenModified。
	rom2, err := assets.OpenModified(dir)
	if err != nil {
		t.Fatalf("重開資料目錄失敗：%v", err)
	}
	if err := rom2.LoadImage(image); err != nil {
		t.Fatal(err)
	}
	s2, err := New(rom2)
	if err != nil {
		t.Fatalf("從寫出去的存檔開場失敗：%v", err)
	}
	if x, y := s2.World().Party.X, s2.World().Party.Y; x != wantX || y != wantY {
		t.Errorf("重開後在 (%d,%d)，存檔時是 (%d,%d)", x, y, wantX, wantY)
	}
	if got := s2.World().Clock; got.Hour != wantClock.Hour || got.Minute != wantClock.Minute {
		t.Errorf("重開後時鐘 %02d:%02d，存檔時是 %02d:%02d",
			got.Hour, got.Minute, wantClock.Hour, wantClock.Minute)
	}
}

// TestSaveWithoutDirSaysSo：沒有可寫目錄時**不准報「存好了」**。
func TestSaveWithoutDirSaysSo(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'S'}); err != nil {
		t.Fatal(err)
	}
	if got := s.Message(); got == "Game saved." {
		t.Errorf("沒給 save-dir 卻報 %q——那是謊", got)
	}
}
