package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestDesertHeatViaScript 驗沙漠高溫（docs/re/75）。
//
// 地圖 0 的 (47,62) 是 nibble 6 記錄 1 ＝ 腳本；`+0x00` 是 section 0x10 的
// 索引，取出來的 opcode 是 **3（晝夜分支）**：白天把 `+0x03`／`+0x04`、
// 夜間把 `+0x05`／`+0x06` 搬進 `+0x01`／`+0x02`，再由位移 1 改寫這一格——
// 於是那一格當場變成 nibble 2 記錄 7–9（白天）或 10–12（夜間），
// **同一步**就印出高溫訊息。
func TestDesertHeatViaScript(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(0, 48, 62); err != nil {
		t.Fatalf("載入地圖 0 失敗：%v", err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirLeft}); err != nil {
		t.Fatalf("走進 (47,62) 失敗：%v", err)
	}
	if got := s.Message(); !strings.Contains(got, "warm") {
		t.Fatalf("踩上腳本格的訊息 ＝ %q，預期含 warm", got)
	}
	// 開場是 01:00，走過去仍是夜間 → 選的是記錄 10–12 那一組
	// （白天那組扣 2／4／6 顆 d6，夜間是 1／2／3）。
	if !strings.Contains(s.Message(), "It is very warm") {
		t.Errorf("夜間第一階段的訊息 ＝ %q，預期 It is very warm.", s.Message())
	}
	// 條件閘收尾用 `fd fd fd fd`（沿用改寫前的值）把這一格**改回原樣**，
	// 所以下次再踩會重跑一次——高溫是每一步都算的，不是一次性的。
	terrain, idx, _, err := s.World().Block.At(47, 62)
	if err != nil {
		t.Fatal(err)
	}
	if terrain != 6 || idx != 1 {
		t.Errorf("跑完之後 (47,62) 是 nibble %d 記錄 %d，預期改回 6／1",
			terrain, idx)
	}
}
