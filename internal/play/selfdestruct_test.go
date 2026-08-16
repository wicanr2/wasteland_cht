package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// selfDestructMap／selfDestructRecord 是全 42 張地圖裡唯一一筆啟動自毀的記錄
// （`docs/re/100` §2）。`TestSelfDestructScriptIsUniqueInData` 守著這個唯一性。
const (
	selfDestructMap    = 20
	selfDestructRecord = 4
)

// TestSelfDestructScriptIsUniqueInData 掃 42 張地圖，確認 opcode 35
// （啟動自毀 ＝ 結局的唯一入口）只有一筆記錄用得到。
//
// 這是**正面事實**，與 `TestEndingHasNoTriggerInData` 那條負面事實成對：
// 資料裡沒有任何一格是「設施第 4 種」，結局是倒數自己合成出來的。
// 兩條一起才說得完整——只留其中一條，下一輪都會讀成別的意思。
func TestSelfDestructScriptIsUniqueInData(t *testing.T) {
	rom := openRom(t)
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	var found [][2]int
	opcodes := map[int]int{}
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for i := 0; i < blk.SectionCount(6); i++ {
			rec, err := blk.SectionRecord(6, i)
			if err != nil || len(rec) == 0 || rec[0]&0x80 != 0 {
				continue // bit7 設起來的是設施，不走 section 0x10
			}
			op, err := blk.SectionEntry(0x10, int(rec[0]))
			if err != nil {
				continue
			}
			opcodes[int(op)]++
			if int(op) == game.OpStartTimer {
				found = append(found, [2]int{res.ID, i})
			}
		}
	}
	// 正對照：這一輪掃到的 opcode 要涵蓋已知一定存在的幾支，
	// 否則「只有一筆」可能只是掃描本身有洞。
	for _, op := range []int{game.OpBeep, game.OpDayNight, game.OpStatusAll} {
		if opcodes[op] == 0 {
			t.Fatalf("正對照失敗：opcode %d 一筆都沒掃到，掃描有洞", op)
		}
	}
	if len(found) != 1 {
		t.Fatalf("opcode %d 出現在 %v，預期只有一筆", game.OpStartTimer, found)
	}
	if found[0][0] != selfDestructMap || found[0][1] != selfDestructRecord {
		t.Errorf("啟動自毀的記錄在資源 %d 第 %d 筆，預期 %d／%d",
			found[0][0], found[0][1], selfDestructMap, selfDestructRecord)
	}
}

// TestSelfDestructCountdownEndsTheGame 走玩家的路：踩上啟動自毀的那一格，
// 再走滿 240 刻，結局要自己開始。
//
// 這一條是 T1 的驗收——**在它之前，遊戲玩不到結局**：結局那一段掛在設施跳表的
// 第 4 格，而沒有任何一筆地圖記錄指到那一格（`TestEndingHasNoTriggerInData`）。
// 真正的入口是主迴圈裡的 `sub_1CB30`（`docs/re/100`）。
func TestSelfDestructCountdownEndsTheGame(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(selfDestructMap, 1, 1); err != nil {
		t.Fatalf("換到地圖 %d 失敗：%v", selfDestructMap, err)
	}
	w := s.World()
	x, y := floorRun(t, w.Block)
	// 站在中間那一格，把左邊那一格換成自毀腳本（資料裡是走完四站啟動序列
	// 之後，nibble 12 記錄 3 把自己那一格改寫成這個值）。
	w.Party.X, w.Party.Y = uint8(x+1), uint8(y)
	if err := w.Block.SetCell(x, y, 6, selfDestructRecord); err != nil {
		t.Fatal(err)
	}
	denomBefore := w.Block.Header[0x2F]
	if denomBefore == 0 {
		t.Fatal("地圖 20 的遭遇分母本來就是 0，這條測不到腳本把它歸零")
	}

	step(t, s, input.Input{Dir: input.DirLeft})
	if !w.SelfDestruct.Armed {
		t.Fatal("踩上自毀腳本之後倒數沒有啟動")
	}
	if got := w.Block.Header[0x2F]; got != 0 {
		t.Errorf("遭遇分母 = %d，預期腳本把它歸零", got)
	}
	armedAt := w.Clock.Total

	// 來回走：一步一刻（地圖 20 的標頭 +0x36 ＝ 1），走滿 240 刻才會爆。
	for i := 0; i < game.SelfDestructTicks*2 && !s.ending.active; i++ {
		if w.Clock.Total-armedAt >= game.SelfDestructTicks-1 && s.ending.active {
			break
		}
		d := input.DirRight
		if i%2 == 1 {
			d = input.DirLeft
		}
		step(t, s, input.Input{Dir: d})
	}
	if !s.ending.active {
		t.Fatalf("走了 %d 刻結局還沒開始", w.Clock.Total-armedAt)
	}
	if elapsed := w.Clock.Total - armedAt; elapsed < game.SelfDestructTicks {
		t.Errorf("結局在第 %d 刻就開始了，預期至少 %d 刻", elapsed, game.SelfDestructTicks)
	}
	if w.SelfDestruct.Armed {
		t.Error("結局開始之後倒數旗標沒有清掉（原版 `0x1B6B0` 會清）")
	}
}

// floorRun 找一段左右相鄰、三格都走得過去的地板，回傳最左邊那一格。
func floorRun(t *testing.T, b *assets.Block) (int, int) {
	t.Helper()
	for y := 1; y < b.Dim-1; y++ {
		for x := 1; x < b.Dim-2; x++ {
			ok := true
			for dx := 0; dx < 3; dx++ {
				terrain, _, _, err := b.At(x+dx, y)
				if err != nil || terrain != 0 {
					ok = false
					break
				}
			}
			if ok {
				return x, y
			}
		}
	}
	t.Fatal("地圖上找不到連續三格的地板")
	return 0, 0
}
