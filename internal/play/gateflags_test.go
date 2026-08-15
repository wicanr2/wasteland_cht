package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// nibble2Records 收全部 42 張地圖裡 nibble 2 指到的相異記錄。
func nibble2Records(t *testing.T) [][]byte {
	t.Helper()
	rom := openRom(t)
	resources, err := rom.Resources()
	if err != nil {
		t.Fatalf("列舉資源失敗：%v", err)
	}
	var out [][]byte
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		seen := map[byte]bool{}
		for y := 0; y < blk.Dim; y++ {
			for x := 0; x < blk.Dim; x++ {
				terrain, idx, _, err := blk.At(x, y)
				if err != nil || terrain != 2 || seen[idx] {
					continue
				}
				seen[idx] = true
				rec, err := blk.SectionRecord(2, int(idx))
				if err == nil && len(rec) > 0 {
					out = append(out, rec)
				}
			}
		}
	}
	return out
}

// TestGateFlagsAppearInShippedData 記錄 +0x00 低位那四個旗標在出貨資料裡的分布。
//
// 四條分支都有記錄會走到——**沒有一條是死碼**，所以四條都得實作
// （docs/re/69）。數字寫進測試，解析出問題會直接紅。
func TestGateFlagsAppearInShippedData(t *testing.T) {
	recs := nibble2Records(t)
	if len(recs) != 424 {
		t.Fatalf("nibble 2 的相異記錄 = %d，預期 424", len(recs))
	}
	want := map[byte]int{
		game.GateAnyPass:    76,
		game.GateCondPatch:  61,
		game.GateWholeParty: 15,
		game.GateEachMember: 139,
	}
	for mask, n := range want {
		got := 0
		for _, r := range recs {
			if r[0]&mask != 0 {
				got++
			}
		}
		if got != n {
			t.Errorf("旗標 %#02x：%d 筆，預期 %d", mask, got, n)
		}
	}
}

// TestCondPatchTableIsRealData 驗「0xFF 之後接一張逐條件改寫表」。
//
// GateCondPatch（`& 8`）的收尾用 `sub_142B1` 算出來的位移去改寫地圖格：
// 位移 ＝ 0xFF 的位置 + 1 + 2n。判準是**取到的值合不合法**——
// 改寫對的第一個 byte 只能是 bit7 設（不改／沿用）或 nibble ≤ 0x0F，
// 0x10–0x7F 是不可能的值，佔隨機資料的 44%。
//
// ⚠ 「記錄長度放得下」不能當判準：`SectionRecord` 的切片一路延伸到區段結尾，
// 任何位移都放得下，測出來 424 筆全過——**恆真的判準證明不了任何事**。
func TestCondPatchTableIsRealData(t *testing.T) {
	recs := nibble2Records(t)
	// legal 回報這個 byte 能不能當改寫對的第 1 層。
	legal := func(b byte) bool { return b >= 0x80 || b <= 0x0F }

	count := func(hasFlag bool, shift int) (pairs, bad int) {
		for _, rec := range recs {
			if (rec[0]&game.GateCondPatch != 0) != hasFlag {
				continue
			}
			gates := game.ParseGates(rec)
			for i := range gates {
				at := game.CondPatchOffsetFor(rec, i) + shift
				if at < 0 || at+1 >= len(rec) {
					continue
				}
				pairs++
				if !legal(rec[at]) {
					bad++
				}
			}
		}
		return
	}

	pairs, bad := count(true, 0)
	if pairs != 599 {
		t.Fatalf("GateCondPatch 的改寫對 = %d，預期 599", pairs)
	}
	if bad != 0 {
		t.Errorf("%d／%d 對取到不可能的第 1 層值", bad, pairs)
	}

	// 負對照一：同樣的位移，套在**沒有** bit3 的記錄上（那裡不是改寫表）。
	ctlPairs, ctlBad := count(false, 0)
	if ctlBad*10 < ctlPairs { // 少於一成不合法就代表判準太鬆
		t.Errorf("負對照太乾淨（%d／%d 不合法）——判準可能沒有鑑別力",
			ctlBad, ctlPairs)
	}
	// 負對照二：位移錯開一格，落到記錄索引那個 byte 上。
	for _, shift := range []int{-1, 1} {
		_, offBad := count(true, shift)
		if offBad == 0 {
			t.Errorf("位移 %+d 也全部合法——公式對不對看不出來", shift)
		}
	}
	t.Logf("GateCondPatch：%d 對全合法；同位移的負對照 %d／%d 不合法",
		pairs, ctlBad, ctlPairs)
}

// TestCondPatchRewritesCellByPartySize 拿地圖 4 (1,2) 驗 sub_142B1 的位移公式。
//
// 那一格的記錄有四條「比隊伍人數」的條件（1／2／3／4 人，型別 3 不擲骰），
// 而 `& 8` 讓它依**通過的是哪一條**改寫這一格：
//
//	1–2 人 → nibble 12 記錄 0（被發現，直接開打）
//	3–4 人 → nibble 2  記錄 1（還有一次伏擊的機會）
//
// 出廠存檔是 4 人，所以走上去之後那一格要變成 nibble 2 記錄 1。
// **位移公式錯一格，改寫的目標就會變**——這是這條公式最直接的驗收。
func TestCondPatchRewritesCellByPartySize(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(4, 2, 2); err != nil {
		t.Fatalf("載入地圖 4 失敗：%v", err)
	}
	w := s.World()
	if n := len(w.Party.Members); n != 4 {
		t.Fatalf("隊伍 %d 人，這個測試要 4 人", n)
	}
	if terrain, idx, _, _ := w.Block.At(1, 2); terrain != 2 || idx != 0 {
		t.Fatalf("走進去之前 (1,2) 是 nibble %d 記錄 %d，預期 2／0", terrain, idx)
	}

	if _, err := s.Update(input.Input{Dir: input.DirLeft}); err != nil {
		t.Fatalf("走進 (1,2) 失敗：%v", err)
	}
	if x, y := w.Party.X, w.Party.Y; x != 1 || y != 2 {
		t.Fatalf("走完在 (%d,%d)，預期 (1,2)", x, y)
	}
	if got := s.Message(); !strings.Contains(got, "three outlaws") {
		t.Errorf("訊息 = %q，預期含 three outlaws（記錄 +0x01）", got)
	}
	terrain, idx, _, _ := w.Block.At(1, 2)
	if terrain != 2 || idx != 1 {
		t.Errorf("改寫後 (1,2) 是 nibble %d 記錄 %d，預期 2／1", terrain, idx)
	}
}
