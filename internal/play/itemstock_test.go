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
	// 進場先選人（隊伍不只一個人時原版會問「誰要進去？」）。
	if _, err := s.Update(input.Input{Char: '1'}); err != nil {
		t.Fatal(err)
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

// 進商店要先問「誰要進去？」（`docs/re/117` §2.1）。
//
// ⚠ 隊伍只有一個人時**不問**——原版直接用他。這一條兩邊都驗，
// 因為「多問一次」與「該問卻沒問」在畫面上都只是「選單長得不太一樣」。
func TestShopAsksWhoEnters(t *testing.T) {
	s := newScene(t)
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
	if got := s.facility.state.Step; got != StepWho {
		t.Fatalf("進場停在第 %d 層，預期 StepWho（%d）", got, StepWho)
	}
	lines := strings.Join(s.facility.Lines, "\n")
	if !strings.Contains(lines, "Who wants to enter?") {
		t.Errorf("沒有問「誰要進去？」：\n%s", lines)
	}
	// 招呼語是這張地圖自己的字串（記錄 `+0x05`）。
	if !strings.Contains(lines, "Welcome to the shop.") {
		t.Errorf("沒有招呼語：\n%s", lines)
	}

	// 選第 2 個人 → 換他站在櫃檯前。
	if _, err := s.Update(input.Input{Char: '2'}); err != nil {
		t.Fatal(err)
	}
	if got := s.facility.state.Who; got != 1 {
		t.Errorf("櫃檯前是第 %d 個人，預期第 1 個（0 起算）", got)
	}
	if got := s.facility.state.Step; got != StepMain {
		t.Errorf("選完人停在第 %d 層，預期主選單（%d）", got, StepMain)
	}
}

// `P` 是**集中金錢**，不是換人（`docs/re/117` §3：畫面下緣寫著 `POOL MONEY`）。
func TestPoolMoneyGathersEveryonesCash(t *testing.T) {
	p := &game.Party{Members: []*game.Character{
		{Name: "A", Money: 100, CON: 10, MaxCON: 10},
		{Name: "B", Money: 250, CON: 10, MaxCON: 10},
		{Name: "C", Money: 7, CON: 10, MaxCON: 10},
	}}
	if got := game.PoolMoney(p, 1); got != 357 {
		t.Errorf("集中之後 B 有 %d，預期 357", got)
	}
	for _, i := range []int{0, 2} {
		if p.Members[i].Money != 0 {
			t.Errorf("第 %d 個人還留著 %d", i, p.Members[i].Money)
		}
	}
	// 錢的欄位是 24 bit，加到滿就停在上限（不能溢位回小數字）。
	big := &game.Party{Members: []*game.Character{
		{Name: "A", Money: 1<<24 - 1}, {Name: "B", Money: 1000},
	}}
	if got := game.PoolMoney(big, 0); got != 1<<24-1 {
		t.Errorf("上限是 %d，得到 %d", 1<<24-1, got)
	}
}
