package play

// 畫面覆蓋率的量測與門檻寫在 `docs/re/84`。
import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

var inputRight = input.Input{Dir: input.DirRight}

func flip(in input.Input) input.Input {
	if in.Dir == input.DirRight {
		return input.Input{Dir: input.DirLeft}
	}
	return input.Input{Dir: input.DirRight}
}

// TestEveryMapRenders 是呈現層的門檻：**42 張地圖每一張都畫得出來**。
//
// 缺圖磚、缺字型、座標算錯都會在這裡炸——而在一般測試裡不會，
// 因為那些測試只跑一兩張地圖。
func TestEveryMapRenders(t *testing.T) {
	rom := openRom(t)
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	drawn, blank := 0, 0
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			t.Errorf("資源 %d 載不進來：%v", res.ID, err)
			continue
		}
		// 從地圖中央開始，避開邊界。
		mid := uint8(blk.Dim / 2)
		if err := s.LoadMap(res.ID, mid, mid); err != nil {
			t.Errorf("資源 %d LoadMap 失敗：%v", res.ID, err)
			continue
		}
		f := s.Frame()
		if f == nil {
			t.Errorf("資源 %d 的 Frame 是 nil", res.ID)
			continue
		}
		if len(f.Pix) != render.ScreenWidth*render.ScreenHeight {
			t.Errorf("資源 %d 的畫面是 %d 個像素，預期 %d",
				res.ID, len(f.Pix), render.ScreenWidth*render.ScreenHeight)
			continue
		}
		// 值域：EGA 16 色。
		for i, v := range f.Pix {
			if v > 15 {
				t.Fatalf("資源 %d 的像素 %d 是 %d，超出 EGA 16 色", res.ID, i, v)
			}
		}
		// 全 0 代表什麼都沒畫出來。
		nonZero := 0
		for _, v := range f.Pix {
			if v != 0 {
				nonZero++
			}
		}
		if nonZero == 0 {
			blank++
			t.Errorf("資源 %d 畫出來是全黑", res.ID)
			continue
		}
		drawn++
		if drawn <= 3 {
			t.Logf("  資源 %2d：%d／%d 個像素非零（%.1f%%）",
				res.ID, nonZero, len(f.Pix), 100*float64(nonZero)/float64(len(f.Pix)))
		}
	}
	t.Logf("42 張地圖：畫得出來 %d 張、全黑 %d 張", drawn, blank)
	if drawn != len(resources) {
		t.Errorf("只有 %d／%d 張畫得出來", drawn, len(resources))
	}
}

// TestEveryFacilityRenders：23 家設施的畫面都要畫得出來。
//
// 設施畫面有圖片與局部動畫（規格 26），缺圖或動畫解析錯會在這裡炸。
func TestEveryFacilityRenders(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	resources, _ := rom.Resources()
	drawn := 0
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for i := 0; i < blk.SectionCount(6); i++ {
			rec, err := blk.SectionRecord(6, i)
			if err != nil || len(rec) == 0 || rec[0]&0x80 == 0 {
				continue
			}
			if fs := s.EnterFacility(rec); fs == nil {
				continue
			}
			f := s.Frame()
			if f == nil || len(f.Pix) != render.ScreenWidth*render.ScreenHeight {
				t.Errorf("資源 %d 記錄 %d 的設施畫面不完整", res.ID, i)
				continue
			}
			// 動畫推一拍，確認不會炸。
			s.TickAnim()
			drawn++
		}
	}
	s.LeaveFacility()
	t.Logf("設施畫面：%d 家畫得出來", drawn)
	if drawn == 0 {
		t.Error("一家設施都畫不出來")
	}
}

// TestCombatRenders：打起來之後畫面也要出得來。
func TestCombatRenders(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(0, 12, 2); err != nil {
		t.Fatal(err)
	}
	dir := inputRight
	for i := 0; i < 400 && !s.InCombat(); i++ {
		if _, err := s.Update(dir); err != nil {
			t.Fatal(err)
		}
		dir = flip(dir)
	}
	if !s.InCombat() {
		t.Skip("沒打起來")
	}
	f := s.Frame()
	if f == nil || len(f.Pix) != render.ScreenWidth*render.ScreenHeight {
		t.Fatal("戰鬥畫面不完整")
	}
	nonZero := 0
	for _, v := range f.Pix {
		if v != 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Error("戰鬥畫面是全黑")
	}
	t.Logf("戰鬥畫面：%d／%d 個像素非零", nonZero, len(f.Pix))
}

// TestEnemyCellGetsIcon 驗遭遇生成的敵人格會疊上圖示（docs/re/85）。
//
// 生成器把格子改成 nibble 15（`docs/re/78`），圖示要走
// 「section 15 記錄的種類 → 敵人資料表 +0x06 → ds:AA17h」這一鏈。
func TestEnemyCellGetsIcon(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	tbl, ok := s.spawnTables()
	if !ok {
		t.Fatal("讀不到遭遇生成的三張表")
	}
	w := s.World()
	w.Party.X, w.Party.Y = 12, 2
	var sx, sy int
	for i := 0; i < 3000; i++ {
		if r := w.SpawnEncounter(tbl); r.Placed {
			sx, sy = r.X, r.Y
			break
		}
	}
	if sx == 0 && sy == 0 {
		t.Skip("3000 次沒生成")
	}
	// 視窗擺到敵人格看得到的位置。
	w.Teleport(uint8(sx), uint8(sy+1))
	found := false
	for _, ic := range w.ViewIcons() {
		if w.ViewX+ic.Col == sx && w.ViewY+ic.Row == sy {
			found = true
			t.Logf("敵人格 (%d,%d) 疊上圖示 %d", sx, sy, ic.Icon)
		}
	}
	if !found {
		t.Errorf("敵人格 (%d,%d) 沒有疊圖", sx, sy)
	}
}

// 遭遇時要畫出敵人肖像（docs/re/37 §3.2 的 `+0x07`）。
//
// **編號解出來很久了，畫面上一直沒有**——這一條擋的是那個狀態。
func TestEncounterShowsPortrait(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadMap(4, 18, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartEncounter(); err != nil {
		t.Fatal(err)
	}
	if !s.InCombat() {
		t.Fatal("地圖 4 的 (18, 2) 開不了戰")
	}
	if s.portrait < 0 {
		t.Fatal("打起來了卻沒有肖像編號")
	}
	t.Logf("肖像圖編號 %d", s.portrait)

	// 圖的那一塊（視窗原點 8,8、96×84）要有非零像素。
	f := s.Frame()
	nonZero := 0
	for y := render.FacilityPicY; y < render.FacilityPicY+84; y++ {
		for x := render.FacilityPicX; x < render.FacilityPicX+96; x++ {
			if f.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	t.Logf("肖像區有 %d／%d 個非零像素", nonZero, 96*84)
	if nonZero == 0 {
		t.Error("肖像沒畫上去")
	}

	// 名單那幾列也要還在（兩者不該互相蓋掉）。
	rosterY := render.RosterHeaderRow * render.CharHeight
	roster := 0
	for y := rosterY; y < rosterY+8 && y < render.ScreenHeight; y++ {
		for x := 0; x < render.ScreenWidth; x++ {
			if f.At(x, y) != 0 {
				roster++
			}
		}
	}
	if roster == 0 {
		t.Error("名單表頭不見了——肖像蓋掉了它")
	}

	// 收尾之後肖像要清掉。
	s.FinishEncounter()
	if s.portrait >= 0 {
		t.Error("戰鬥結束了肖像編號還留著")
	}
}

// ALLPICS 是兩個檔一個編號空間，82 張都要載得到（docs/re/23 §4）。
//
// 先前只載 allpics1（33 張），設施圖剛好都落在那一段所以看不出問題——
// **敵人肖像用到 44，載一個檔就是一片空白**。
func TestAllPicturesLoaded(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	const want = 82
	if len(s.pics) != want {
		t.Errorf("應該載到 %d 張圖，得到 %d", want, len(s.pics))
	}
	// 每張都要有像素——空圖表示解碼在某一張上斷了。
	empty := 0
	for i, p := range s.pics {
		if p == nil || len(p.Pix) == 0 {
			t.Errorf("第 %d 張是空的", i)
			continue
		}
		nz := 0
		for _, v := range p.Pix {
			if v != 0 {
				nz++
			}
		}
		if nz == 0 {
			empty++
		}
	}
	t.Logf("%d 張圖，全黑的 %d 張", len(s.pics), empty)
	if empty != 0 {
		t.Errorf("有 %d 張全黑", empty)
	}
}
