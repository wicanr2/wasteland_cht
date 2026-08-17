package render

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 腳本 opcode 2 的圖形對調（overlay slot 18，`docs/re/104`）。
func mkGraphics(icons, tiles int) *Graphics {
	g := &Graphics{}
	for i := 0; i < icons; i++ {
		g.Icons = append(g.Icons, &assets.Indexed{Width: 1, Height: 1, Pix: []byte{byte(i)}})
		mask := make([]bool, 4)
		mask[0] = true // 每一張遮罩都不是全 0，才分得出有沒有被換掉
		g.Masks = append(g.Masks, mask)
	}
	for i := 0; i < tiles; i++ {
		g.Tiles = append(g.Tiles, &assets.Indexed{
			Width: 1, Height: 1, Pix: []byte{byte(100 + i)}})
	}
	return g
}

// 疊圖與圖磚在原版是**同一張連續的表**（`seg003:0x420`，`docs/re/24` §2.3），
// 所以編號 7 與編號 0x2C 對調要真的跨過那條界線。
func TestGraphicsSwapCrossesIconTileBoundary(t *testing.T) {
	g := mkGraphics(10, 100)
	tile := g.Tiles[0x2C-10]
	icon := g.Icons[7]

	if err := g.Swap(7, 0x2C); err != nil {
		t.Fatalf("對調 7／0x2C：%v", err)
	}
	if g.Icons[7] != tile {
		t.Error("編號 7 應該換成那張圖磚")
	}
	if g.Tiles[0x2C-10] != icon {
		t.Error("那張圖磚應該換成原本的疊圖")
	}

	// ⚠ **這是開關不是單向操作**：同一對再跑一次就回原狀（原版做的是 xchg）。
	if err := g.Swap(0x2C, 7); err != nil {
		t.Fatalf("換回來：%v", err)
	}
	if g.Icons[7] != icon || g.Tiles[0x2C-10] != tile {
		t.Error("參數對調再跑一次應該完全回原狀")
	}
}

// 遮罩那一半：編號 < 10 的那一張與暫存格對調，而暫存格是**全 0**
//（映像裡 `seg003:0xDF60` 那 32 bytes 全零）——換上去等於不透明。
func TestGraphicsSwapParksTheMask(t *testing.T) {
	g := mkGraphics(10, 100)
	before := g.Masks[7]

	if err := g.Swap(7, 0x2C); err != nil {
		t.Fatal(err)
	}
	for i, b := range g.Masks[7] {
		if b {
			t.Fatalf("換進來的遮罩應該全 0（不透明），第 %d 格是 true", i)
		}
	}
	if err := g.Swap(0x2C, 7); err != nil {
		t.Fatal(err)
	}
	if &g.Masks[7][0] != &before[0] {
		t.Error("換回來之後應該是原本那一張遮罩")
	}
}

// 兩個編號都 ≥ 10 時**完全不動遮罩**（`0x10C88`–`0x10C91` 的兩道比較）。
func TestGraphicsSwapLeavesMasksAloneForTiles(t *testing.T) {
	g := mkGraphics(10, 100)
	var before [][]bool
	for _, m := range g.Masks {
		before = append(before, m)
	}
	if err := g.Swap(0x2C, 0x5D); err != nil {
		t.Fatal(err)
	}
	for i := range g.Masks {
		if &g.Masks[i][0] != &before[i][0] {
			t.Errorf("第 %d 張遮罩被動到了——兩個編號都 ≥ 10 時不該碰遮罩", i)
		}
	}
}

// 編號超出範圍要**報錯不吞**：那代表資料或索引算錯了。
func TestGraphicsSwapRejectsOutOfRange(t *testing.T) {
	g := mkGraphics(10, 4)
	if err := g.Swap(7, 200); err == nil {
		t.Error("編號 200 超出 10 疊圖 ＋ 4 圖磚，應該報錯")
	}
}
