package game

import (
	"os"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 原版資料不入版控，沒有就跳過（與 internal/assets 的測試同一套規矩）。
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

// 驗收 3：時鐘的推進量對得上標頭（docs/spec/04 §8）。
func TestClockAdvance(t *testing.T) {
	// 荒野：每步 4 分鐘 ＝ 標頭 0x0400。走 15 步剛好一小時。
	var c Clock
	for i := 0; i < 15; i++ {
		c.Advance(4*256, 1)
	}
	if c.Hour != 1 || c.Minute != 0 {
		t.Fatalf("荒野走 15 步應該是 01:00，得到 %02d:%02d", c.Hour, c.Minute)
	}

	// 室內：每步 15 秒 ＝ 0.25 分鐘 ＝ 標頭 0x0040。走 240 步一小時。
	c = Clock{}
	for i := 0; i < 240; i++ {
		c.Advance(64, 1)
	}
	if c.Hour != 1 || c.Minute != 0 {
		t.Fatalf("室內走 240 步應該是 01:00，得到 %02d:%02d", c.Hour, c.Minute)
	}
}

func TestClockWrapAndNight(t *testing.T) {
	c := Clock{Hour: 23, Minute: 59}
	c.Advance(1*256, 0)
	if c.Hour != 0 || c.Minute != 0 {
		t.Fatalf("跨日應該回到 00:00，得到 %02d:%02d", c.Hour, c.Minute)
	}
	for _, tc := range []struct {
		hour  uint8
		night bool
	}{{0, true}, {5, true}, {6, false}, {17, false}, {18, true}, {23, true}} {
		c := Clock{Hour: tc.hour}
		if c.Night() != tc.night {
			t.Errorf("%d 時的晝夜判定錯了", tc.hour)
		}
	}
}

// 每 16 刻觸發一次週期處理。
func TestClockPeriodic(t *testing.T) {
	var c Clock
	hits := 0
	for i := 0; i < 64; i++ {
		if c.Advance(256, 1) {
			hits++
		}
	}
	if hits != 4 {
		t.Fatalf("走 64 步、每步 1 刻應該觸發 4 次，得到 %d", hits)
	}
}

// 驗收 4：健康角色回滿的步數 ＝ (MaxCON − 1) × 64 / StepTick。
func TestNaturalHealing(t *testing.T) {
	c := &Character{CON: 1, MaxCON: 5}
	p := &Party{Members: []*Character{c}, Selected: -1}
	var clk Clock
	for i := 0; i < (5-1)*64; i++ {
		if clk.Advance(256, 1) {
			p.Tick16(clk.Tick)
		}
	}
	if c.CON != c.MaxCON {
		t.Fatalf("走 %d 步之後 CON 應該回滿 %d，得到 %d", (5-1)*64, c.MaxCON, c.CON)
	}
}

// 選中的角色回復慢一倍（遮罩 0x7F 而不是 0x3F）。
func TestSelectedHealsSlower(t *testing.T) {
	a := &Character{CON: 1, MaxCON: 99}
	b := &Character{CON: 1, MaxCON: 99}
	p := &Party{Members: []*Character{a, b}, Selected: 0}
	var clk Clock
	for i := 0; i < 640; i++ {
		if clk.Advance(256, 1) {
			p.Tick16(clk.Tick)
		}
	}
	if a.CON >= b.CON {
		t.Fatalf("選中的角色應該回得比較慢：選中 %d、未選中 %d", a.CON, b.CON)
	}
}

// 生病的人會惡化，而且掉破 −50 直接歸零（死亡）。
func TestDiseaseWorsensToDeath(t *testing.T) {
	c := &Character{CON: 10, MaxCON: 20, Status: StatusRabies}
	p := &Party{Members: []*Character{c}, Selected: -1}
	var clk Clock
	for i := 0; i < 64*200 && c.CON != 0; i++ {
		if clk.Advance(256, 1) {
			p.Tick16(clk.Tick)
		}
	}
	if c.CON != 0 {
		t.Fatalf("生病的人最後應該死亡（CON ＝ 0），得到 %d", c.CON)
	}

	// 低四位的狀態不會惡化。
	d := &Character{CON: 10, MaxCON: 10, Status: StatusRadiation}
	p = &Party{Members: []*Character{d}, Selected: -1}
	clk = Clock{}
	for i := 0; i < 64*20; i++ {
		if clk.Advance(256, 1) {
			p.Tick16(clk.Tick)
		}
	}
	if d.CON != 10 {
		t.Fatalf("低四位的狀態不該讓 CON 變動，得到 %d", d.CON)
	}
}

// 驗收 2：從出廠位置往四個方向各走 100 步，不出界、不踩進會擋的 nibble。
func TestStepStaysInsideAndRespectsGates(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(0)
	if err != nil {
		t.Fatalf("載入區塊 0 失敗：%v", err)
	}
	p := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}, Selected: 0, X: 55, Y: 62}
	w := NewWorld(block, p, rng.New())

	moved := 0
	for _, dir := range []Direction{Up, Down, Left, Right} {
		for i := 0; i < 100; i++ {
			before := [2]uint8{p.X, p.Y}
			res, err := w.Step(dir)
			if err != nil {
				t.Fatalf("走一步失敗：%v", err)
			}
			if int(p.X) >= block.Dim || int(p.Y) >= block.Dim {
				t.Fatalf("走出邊界：(%d, %d)，地圖是 %d × %d", p.X, p.Y, block.Dim, block.Dim)
			}
			terrain, _, _, err := block.At(int(p.X), int(p.Y))
			if err != nil {
				t.Fatalf("讀格子失敗：%v", err)
			}
			if blocking[terrain] {
				t.Fatalf("踩進了會擋住的 nibble %d，位置 (%d, %d)", terrain, p.X, p.Y)
			}
			if !res.Moved && [2]uint8{p.X, p.Y} != before {
				t.Fatalf("被擋住卻動了座標")
			}
			if res.Moved {
				moved++
			}
		}
	}
	// 全程都被擋住的話，上面的檢查會空轉通過——這道防呆擋掉那種假綠。
	if moved < 100 {
		t.Fatalf("400 步裡只走成了 %d 步，測試沒有真的動過", moved)
	}
}

// 被擋住時時鐘不動——這是原版的順序，不可放寬。
func TestBlockedStepAdvancesNothing(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(0)
	if err != nil {
		t.Fatalf("載入區塊 0 失敗：%v", err)
	}
	// 找一個四周至少有一面是牆的位置。
	var px, py int = -1, -1
	var dir Direction
	for y := 1; y < block.Dim-1 && px < 0; y++ {
		for x := 1; x < block.Dim-1; x++ {
			if t0, _, _, _ := block.At(x, y); blocking[t0] {
				continue
			}
			for d, dd := range delta {
				if t1, _, _, _ := block.At(x+dd[0], y+dd[1]); blocking[t1] {
					px, py, dir = x, y, Direction(d)
					break
				}
			}
			if px >= 0 {
				break
			}
		}
	}
	if px < 0 {
		t.Skip("這張地圖找不到相鄰的牆")
	}

	p := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}, X: uint8(px), Y: uint8(py)}
	w := NewWorld(block, p, rng.New())
	before := w.Clock
	res, err := w.Step(dir)
	if err != nil {
		t.Fatalf("走一步失敗：%v", err)
	}
	if res.Moved {
		t.Fatalf("(%d, %d) 往 %d 應該被擋住", px, py, dir)
	}
	if w.Clock != before {
		t.Fatalf("被擋住卻推進了時鐘：%+v → %+v", before, w.Clock)
	}
}

// 視窗原點永遠是隊伍座標減 (9, 4)。
func TestViewOriginFollowsParty(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(0)
	if err != nil {
		t.Fatalf("載入區塊 0 失敗：%v", err)
	}
	p := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}, X: 55, Y: 62}
	w := NewWorld(block, p, rng.New())
	for i := 0; i < 50; i++ {
		if _, err := w.Step(Direction(i % 4)); err != nil {
			t.Fatalf("走一步失敗：%v", err)
		}
		if w.ViewX != int(p.X)-ViewOffsetX || w.ViewY != int(p.Y)-ViewOffsetY {
			t.Fatalf("視窗原點沒跟上：隊伍 (%d, %d)、原點 (%d, %d)", p.X, p.Y, w.ViewX, w.ViewY)
		}
	}
}
