package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 移動閘的收尾（docs/re/69 §2 的 0x1406F）：**通過或擋住都改寫這一格**。
// 找一格真的要判定的移動閘（bit7 清、bit6 設），把條件改成可控的
// 「比錢」（型別 4，只比不扣），兩個方向各驗一次。
func gateCellForTest(t *testing.T) (blockID, x, y int) {
	rom := openRom(t)
	for b := 0; b < 42; b++ {
		block, err := rom.Block(b)
		if err != nil {
			continue
		}
		w := NewWorld(block, &Party{Members: []*Character{{CON: 20, MaxCON: 20}}}, rng.New())
		size := block.Dim
		for yy := 1; yy < size-1; yy++ {
			for xx := 1; xx < size-1; xx++ {
				if need, rec := w.gateNeedsCheck(xx, yy); need && len(rec) > 0x0C {
					return b, xx, yy
				}
			}
		}
	}
	t.Skip("42 張地圖找不到要判定的移動閘")
	return 0, 0, 0
}

func stepOntoGate(t *testing.T, blockID, x, y int, prep func(rec []byte)) (*World, StepResult, byte) {
	rom := openRom(t)
	block, err := rom.Block(blockID)
	if err != nil {
		t.Fatal(err)
	}
	p := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}, X: uint8(x - 1), Y: uint8(y)}
	w := NewWorld(block, p, rng.New())
	_, rec := w.gateNeedsCheck(x, y)
	if rec == nil {
		t.Fatal("這一格不再是移動閘")
	}
	prep(rec)
	res, err := w.Step(Right)
	if err != nil {
		t.Fatal(err)
	}
	terrain, _, _, err := block.At(x, y)
	if err != nil {
		t.Fatal(err)
	}
	return w, res, terrain
}

func TestGatePassRewritesCell(t *testing.T) {
	b, x, y := gateCellForTest(t)
	_, res, terrain := stepOntoGate(t, b, x, y, func(rec []byte) {
		rec[0] = 0x40                            // bit6 跑條件、其餘旗標清掉
		rec[2] = 9                               // 通過要印 +0x02
		rec[4], rec[5] = 0x04, 0                 // 通過的改寫對：nibble 4
		rec[0x0A], rec[0x0B], rec[0x0C] = 0x80, 0, 0xFF // 比錢 ≥ 0 ＝ 必過
	})
	if res.Gate.Blocked || !res.Moved {
		t.Fatalf("必過的閘卻沒走過去：%+v", res)
	}
	if res.Gate.Message != 9 {
		t.Errorf("通過應該印 +0x02（9），得到 %d", res.Gate.Message)
	}
	if terrain != 4 {
		t.Errorf("通過之後這一格應該被改寫成 nibble 4，得到 %d", terrain)
	}
}

func TestGateBlockRewritesCell(t *testing.T) {
	b, x, y := gateCellForTest(t)
	w, res, terrain := stepOntoGate(t, b, x, y, func(rec []byte) {
		rec[0] = 0x40
		rec[1] = 7                               // 進閘就印 +0x01
		rec[3] = 8                               // 沒過且沒人受罰印 +0x03
		rec[6], rec[7] = 0x05, 0                 // 擋住的改寫對：nibble 5
		rec[8] = 0                               // 不罰
		rec[0x0A], rec[0x0B], rec[0x0C] = 0x80, 255, 0xFF // 比錢 ≥ 255 ＝ 必敗
	})
	if !res.Gate.Blocked || res.Moved {
		t.Fatalf("必敗的閘卻走過去了：%+v", res)
	}
	if res.Blocked != 7 {
		t.Errorf("擋住應該印 +0x01（7），得到 %d", res.Blocked)
	}
	if res.Gate.Message != 8 {
		t.Errorf("沒人受罰應該印 +0x03（8），得到 %d", res.Gate.Message)
	}
	if terrain != 5 {
		t.Errorf("擋住也要改寫這一格（0x1406F），得到 nibble %d", terrain)
	}
	// 同一步不可以重複評一次（同一支 sub_13EC9 只跑一次）：
	// 改寫已經發生，事件那一側看到的是新 nibble，不會再進 EvalGate。
	_ = w
}
