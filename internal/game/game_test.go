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
	// ⚠ **不要用出廠位置當起點**：(55, 62) 三面是山（nibble 11，docs/re/62），
	// 四個方向各走 100 步只有 56 步走得成，防呆門檻會誤判成「測試沒動過」。
	// 改成從一格四鄰全開的地方出發，這樣「走不動」才是真的被擋。
	sx, sy := -1, -1
	for y := 1; y < block.Dim-1 && sx < 0; y++ {
		for x := 1; x < block.Dim-1; x++ {
			open := true
			for _, d := range [][2]int{{0, 0}, {0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
				n, _, _, err := block.At(x+d[0], y+d[1])
				if err != nil || blocking[n] {
					open = false
					break
				}
			}
			if open {
				sx, sy = x, y
				break
			}
		}
	}
	if sx < 0 {
		t.Fatal("資源 0 找不到四鄰全開的格子")
	}
	p := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}, Selected: 0,
		X: uint8(sx), Y: uint8(sy)}
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
	// 門檻只要「明顯不是零」就好：nibble 11 佔了 20,495 格（docs/re/62），
	// 直線走 100 步撞牆是常態，**不要把它當成走得順的指標**。
	if moved < 50 {
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

// 疊圖（docs/re/48、docs/spec/03 §2.9）。

func TestKindIconMatchesTableInImage(t *testing.T) {
	rom := openRom(t)
	table, err := rom.KindIconTable()
	if err != nil {
		t.Fatalf("讀種類→疊圖表：%v", err)
	}
	// 原版 ds:AA17h 的六個 byte。Go 這邊寫死的常數要與它逐項相同——
	// 這條測試存在的理由就是不讓那份寫死的數字沒人守。
	for k := 1; k < len(table); k++ {
		if got := KindIcon(EnemyKind(k)); got != table[k] {
			t.Errorf("種類 %d：程式回 %d，映像裡是 %d", k, got, table[k])
		}
	}
	if table[0] != 0 {
		t.Errorf("表的第 0 項應該是佔位的 0，得到 %d", table[0])
	}
}

func TestKindIconFallsBackToHumanoid(t *testing.T) {
	// `sub_14664` 的 `mov bl, 3` 在讀記錄之前就先擺好——
	// 種類查不到時走的是**原版的預設**，不是我們補的保險。
	for _, k := range []EnemyKind{0, 6, 9, 255} {
		if got := KindIcon(k); got != IconHumanoid {
			t.Errorf("種類 %d 應該退回 Humanoid（%d），得到 %d", k, IconHumanoid, got)
		}
	}
}

func TestRadiationOnlyAtNight(t *testing.T) {
	// 實機：05:56 有、06:00 沒有（docs/re/48 §4）。門檻兩端都要壓。
	cases := []struct {
		hour int
		want bool
	}{{0, true}, {5, true}, {6, false}, {12, false}, {17, false}, {18, true}, {23, true}}
	for _, c := range cases {
		if got := RadiationVisible(c.hour); got != c.want {
			t.Errorf("%02d 時：輻射標誌可見 %v，應該是 %v", c.hour, got, c.want)
		}
		icon, ok := CellIcon(9, 0, c.hour)
		if ok != c.want || (ok && icon != IconRadiation) {
			t.Errorf("%02d 時的 nibble 9：得到 (%d, %v)", c.hour, icon, ok)
		}
	}
}

func TestCellIconNibble4Threshold(t *testing.T) {
	// `cmp al, 0Ah`：< 10 才是疊圖，≥ 10 是圖磚編號（值 − 10）。
	// bit7 要先去掉——記錄裡那一位是別的用途。
	for _, c := range []struct {
		rec1 byte
		icon byte
		ok   bool
	}{
		{0x00, IconBlack, true},
		{0x07, IconParty, true},
		{0x09, IconOtherGroup, true},
		{0x0A, 0, false}, // 剛好到門檻 → 圖磚
		{0x60, 0, false},
		{0x87, IconParty, true}, // bit7 設起來也一樣是 7
	} {
		icon, ok := CellIcon(4, c.rec1, 0)
		if icon != c.icon || ok != c.ok {
			t.Errorf("記錄 +0x01 ＝ %#02x：得到 (%d, %v)，應該是 (%d, %v)",
				c.rec1, icon, ok, c.icon, c.ok)
		}
	}
}

func TestCellIconLootAndPlainTerrain(t *testing.T) {
	if icon, ok := CellIcon(5, 0, 12); !ok || icon != IconLoot {
		t.Errorf("nibble 5 應該畫寶箱（%d），得到 (%d, %v)", IconLoot, icon, ok)
	}
	// 其餘 nibble 一律不疊圖——原版是 `jnz loc_18088` 走一般圖磚那條。
	for _, n := range []byte{0, 1, 2, 3, 6, 8, 10, 11, 12, 15} {
		if _, ok := CellIcon(n, 0x07, 0); ok {
			t.Errorf("nibble %d 不該疊圖", n)
		}
	}
}

func TestViewIconsFindsRadiationOnMap(t *testing.T) {
	rom := openRom(t)
	// 資源 0 有 36 格 nibble 9（docs/re/48 §6）。把視窗擺到其中一格上，
	// 夜間看得到、白天看不到——同一個位置的前後對照。
	blk, err := rom.Block(0)
	if err != nil {
		t.Skipf("讀不到資源 0：%v", err)
	}
	var found bool
	var x, y int
	for yy := 0; yy < blk.Dim && !found; yy++ {
		for xx := 0; xx < blk.Dim; xx++ {
			if terrain, _, _, err := blk.At(xx, yy); err == nil && terrain == 9 {
				x, y, found = xx, yy, true
				break
			}
		}
	}
	if !found {
		t.Fatal("資源 0 找不到 nibble 9 的格子——docs/re/48 §6 說有 36 格")
	}
	w := NewWorld(blk, &Party{X: uint8(x), Y: uint8(y)}, nil)

	w.Clock.Hour = 2
	night := len(w.ViewIcons())
	w.Clock.Hour = 12
	day := len(w.ViewIcons())
	if night == 0 {
		t.Error("夜間視窗裡應該至少有一格輻射標誌")
	}
	if day >= night {
		t.Errorf("白天的疊圖數 %d 應該少於夜間的 %d", day, night)
	}
}
