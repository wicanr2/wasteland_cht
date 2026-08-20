package play

// 商店的庫存表（`docs/re/118`）。

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 進商店要照記錄 `+0x06` 換那一組庫存表。
//
// ⚠ 這是 DOSBox 實機對拍抓出來的：高池鎮商店用第 1 組，而 remake 一律用
// 第 0 組，於是清單上多了一件原版沒有的 `Engine`（$500），
// **而畫面上看起來只是「這家店貨比較多」**。
func TestShopStockGroupComesFromRecord(t *testing.T) {
	s := newScene(t)
	// 地圖 10 的 (30, 25) 往上走一步是高池鎮商店的傳送格。
	if err := s.LoadMap(10, 30, 25); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Char: 'Y'}); err != nil {
		t.Fatal(err)
	}
	if s.facility == nil {
		t.Fatalf("沒進商店：%q", s.Message())
	}
	if got := s.itemStock; got != 1 {
		t.Errorf("載進來的是第 %d 組，預期第 1 組", got)
	}
	if _, err := s.Update(input.Input{Char: 'B'}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Join(s.facility.Lines, "\n")
	// 第 1 組沒有 Engine（第 0 組才有），而且 Match 要排得進第一頁。
	if strings.Contains(lines, "Engine") {
		t.Errorf("第 1 組不該賣 Engine：\n%s", lines)
	}
	if !strings.Contains(lines, "Match") {
		t.Errorf("第一頁應該有 Match：\n%s", lines)
	}
}

// 四組庫存表**只有庫存那一欄不一樣**。
//
// 這一條是「進商店換整張表」這個做法的前提：價錢或傷害要是也跟著換，
// 換表就會改到戰鬥，而那種偏差在畫面上看不出來。
func TestItemStockGroupsDifferOnlyInStock(t *testing.T) {
	rom := openRom(t)
	base, err := rom.LoadItemTable("game1", 0)
	if err != nil {
		t.Skipf("讀不到第 0 組：%v", err)
	}
	for _, g := range []byte{1, 2, 4} {
		raw, err := rom.LoadItemTable(itemStockFile(g), itemStockSlot(g))
		if err != nil {
			t.Fatalf("第 %d 組讀不到：%v", g, err)
		}
		if len(raw) != len(base) {
			t.Fatalf("第 %d 組長度 %d，第 0 組 %d", g, len(raw), len(base))
		}
		for i := range raw {
			if raw[i] == base[i] {
				continue
			}
			if i%8 != 2 {
				t.Errorf("第 %d 組的物品 %d 欄位 +0x%02X 不一樣（%02X vs %02X）——"+
					"只有庫存（+0x02）可以不同", g, i/8, i%8, raw[i], base[i])
			}
		}
	}
	_ = assets.ItemSlotOffsets
	_ = game.StockNone
}
