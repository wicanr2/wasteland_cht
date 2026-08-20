package play

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
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
	confirmed(t, s, 'S')
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
	confirmed(t, s, 'S')
	if got := s.Message(); got == "Game saved." {
		t.Errorf("沒給 save-dir 卻報 %q——那是謊", got)
	}
}

// TestShopStockSurvivesSave：賣一件東西給店家，**存檔重開之後庫存要留著**。
//
// 物品表是存檔區旁邊的另一個資源、每個存檔槽一份（`docs/re/45` §2），
// 庫存 `+0x02` 是遊戲狀態（`docs/re/42` §3 賣一件 `+1`）。
// 記在設施場景上的話走出店門就沒了，而且不會有任何測試變紅。
func TestShopStockSurvivesSave(t *testing.T) {
	src := os.Getenv("WL_DATA")
	if src == "" {
		src = "../../workplace/orig/wastland"
	}
	if _, err := os.Stat(src); err != nil {
		t.Skipf("找不到原版資料目錄 %s，跳過", src)
	}
	const image = "../../workplace/analysis/unpacked/wl.merged.exe"

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

	// 挑一個庫存**不是「無限」**的品項來改。出廠的表裡只有 0（缺貨）
	// 與 0xFF（無限）兩種，賣一件給店家就是把 0 變成 1（`docs/re/42` §3）。
	target, before := -1, byte(0)
	for id := 0; id < len(s.items); id++ {
		if s.items[id].Stock != game.StockUnlimited {
			target, before = id, s.items[id].Stock
			break
		}
	}
	if target < 0 {
		t.Skip("這份存檔裡每一項的庫存都是無限，測不出差別")
	}
	s.setStock(byte(target), before+1)

	confirmed(t, s, 'S')
	if got := s.Message(); got != "Game saved." && len(s.CJK()) == 0 {
		t.Fatalf("存檔訊息是 %q，看起來沒寫出去", got)
	}

	rom2, err := assets.OpenModified(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rom2.LoadImage(image); err != nil {
		t.Fatal(err)
	}
	s2, err := New(rom2)
	if err != nil {
		t.Fatalf("從寫出去的存檔開場失敗：%v", err)
	}
	if got := s2.items[target].Stock; got != before+1 {
		t.Errorf("品項 %d 的庫存重開後是 %d，存檔時是 %d", target, got, before+1)
	}
}

// TestSaveWritesCurrentMapWhereTheOriginalReadsIt：**目前地圖要寫進全域狀態**。
//
// ⚠ 原版讀檔（`sub_18744`）只把全域狀態那 14 bytes 抄回 `ds:464Eh`，
// 接著 `sub_18350` 比 `ds:4655h`（相對位移 7）決定載哪一張地圖——
// **它一眼都沒看隊伍槽表的 `+0x0A`**（`docs/re/117`）。
//
// 只寫槽表的話原版會用**舊地圖 ＋ 新座標**開場，而畫面上看起來只是
// 「傳送到奇怪的地方」。這一條是 DOSBox 實機對拍抓出來的。
func TestSaveWritesCurrentMapWhereTheOriginalReadsIt(t *testing.T) {
	dir, rom := copyDataDir(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	s.SetSaveDir(dir)
	const (
		wantMap  = 21
		wantX    = 2
		wantY    = 22
		gblMapAt = 7
	)
	if err := s.LoadMap(wantMap, wantX, wantY); err != nil {
		t.Fatal(err)
	}
	confirmed(t, s, 'S')
	if got := s.Message(); got != "Game saved." {
		t.Fatalf("存檔訊息是 %q", got)
	}

	// 直接看那一格：這是原版真正會讀的地方。
	rom2, err := assets.OpenModified(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rom2.LoadImage(testImage); err != nil {
		t.Fatal(err)
	}
	a, err := rom2.LoadSave("game1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := rom2.LoadSave("game2")
	if err != nil {
		t.Fatal(err)
	}
	sv := assets.PickNewer(a, b)
	if got := sv.Globals()[gblMapAt]; got != wantMap {
		t.Errorf("全域狀態的目前地圖 ＝ %d，預期 %d", got, wantMap)
	}
	// 隊伍槽表那一份也要在（切組時用得到）。
	if got := sv.SlotGroups()[0].MapID; got != wantMap {
		t.Errorf("隊伍槽表的地圖 ＝ %d，預期 %d", got, wantMap)
	}

	// 端到端：重開之後真的在那張地圖上。
	s2, err := New(rom2)
	if err != nil {
		t.Fatalf("從寫出去的存檔開場失敗：%v", err)
	}
	if got := s2.World().Block.Resource.ID; got != wantMap {
		t.Errorf("重開後在地圖 %d，預期 %d", got, wantMap)
	}
}

// copyDataDir 複製一份可寫的原版資料目錄（原版目錄唯讀，`CLAUDE.md` §4）。
func copyDataDir(t *testing.T) (string, *assets.Rom) {
	t.Helper()
	src := os.Getenv("WL_DATA")
	if src == "" {
		src = "../../workplace/orig/wastland"
	}
	if _, err := os.Stat(src); err != nil {
		t.Skipf("找不到原版資料目錄 %s，跳過", src)
	}
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
	if err := rom.LoadImage(testImage); err != nil {
		t.Skipf("載入解包映像失敗：%v", err)
	}
	return dir, rom
}

// testImage 是解包合成映像的路徑（測試共用）。
const testImage = "../../workplace/analysis/unpacked/wl.merged.exe"
