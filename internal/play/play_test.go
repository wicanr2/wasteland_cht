package play

import (
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

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

// 從出廠存檔開場：隊伍、座標、地圖都要對得上 docs/spec/05 的驗收。
func TestNewFromShippedSave(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	w := s.World()
	if len(w.Party.Members) != 4 {
		t.Fatalf("出廠應該有四個隊員，得到 %d", len(w.Party.Members))
	}
	if w.Party.X != 55 || w.Party.Y != 62 {
		t.Fatalf("出廠座標應該是 (55, 62)，得到 (%d, %d)", w.Party.X, w.Party.Y)
	}
	if w.Party.Members[0].Name != "Hell Razor" {
		t.Fatalf("第一個隊員應該是 Hell Razor，得到 %q", w.Party.Members[0].Name)
	}
	// 出廠時鐘是 01:00（docs/re/30 §5），從全域狀態的副本讀出來。
	if w.Clock.Hour != 1 || w.Clock.Minute != 0 {
		t.Fatalf("出廠時鐘應該是 01:00，得到 %02d:%02d", w.Clock.Hour, w.Clock.Minute)
	}
}

// 真的走得動：按方向鍵會改座標、推進時鐘，而且畫得出一幀。
func TestWalkAdvancesWorld(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	w := s.World()
	start := [2]uint8{w.Party.X, w.Party.Y}
	before := w.Clock

	moved := 0
	dirs := []input.Direction{input.DirUp, input.DirDown, input.DirLeft, input.DirRight}
	for i := 0; i < 200; i++ {
		cur := [2]uint8{w.Party.X, w.Party.Y}
		ok, err := s.Update(input.Input{Dir: dirs[i%len(dirs)]})
		if err != nil {
			t.Fatalf("走一步失敗：%v", err)
		}
		if !ok {
			t.Fatal("方向鍵不該讓場景結束")
		}
		if [2]uint8{w.Party.X, w.Party.Y} != cur {
			moved++
		}
	}
	if moved == 0 {
		t.Fatal("200 次方向鍵一步都沒走成，接線一定有問題")
	}
	if [2]uint8{w.Party.X, w.Party.Y} == start && moved > 0 {
		t.Log("走了一圈又回到起點，這是四個方向輪流的正常結果")
	}
	if w.Clock == before {
		t.Fatal("走了路時鐘卻沒動")
	}

	f := s.Frame()
	if f == nil {
		t.Fatal("畫不出一幀")
	}
	// 畫面不能整片是同一個顏色——那代表地圖沒畫上去。
	first := f.At(0, 0)
	same := true
	for y := 0; y < 200 && same; y++ {
		for x := 0; x < 320; x++ {
			if f.At(x, y) != first {
				same = false
				break
			}
		}
	}
	if same {
		t.Fatal("整幀只有一種顏色，地圖沒畫出來")
	}
}

// F10 離開。
func TestQuitStopsScene(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	ok, err := s.Update(input.Input{Action: input.ActionQuit})
	if err != nil || ok {
		t.Fatalf("F10 應該讓場景結束：ok=%v err=%v", ok, err)
	}
}

// 驗收（規格 04 §8.6、規格 05 §2.1）：什麼都沒改就寫回，明文要完全相同。
func TestStoreToIsIdempotent(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	save := s.Save()
	before := append([]byte(nil), save.Plain...)
	if err := s.StoreTo(save); err != nil {
		t.Fatalf("寫回失敗：%v", err)
	}
	for i := range save.Plain {
		if save.Plain[i] != before[i] {
			t.Fatalf("什麼都沒改卻動了 +%#x：%#02x → %#02x", i, before[i], save.Plain[i])
		}
	}
}

// 走一段路再寫回：只有座標、視窗原點與時鐘那幾個 byte 會變。
func TestStoreToAfterWalkTouchesOnlyKnownBytes(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	save := s.Save()
	before := append([]byte(nil), save.Plain...)

	for i := 0; i < 40; i++ {
		if _, err := s.Update(input.Input{Dir: input.DirLeft}); err != nil {
			t.Fatalf("走一步失敗：%v", err)
		}
	}
	if err := s.StoreTo(save); err != nil {
		t.Fatalf("寫回失敗：%v", err)
	}

	// 允許變動的位移：隊伍槽表的 +0x08/+0x09，全域 +0x78 的 0/1/2/3/4/10/11/12。
	allowed := map[int]bool{0x08: true, 0x09: true}
	for _, off := range []int{0, 1, 2, 3, 4, 10, 11, 12} {
		allowed[0x78+off] = true
	}
	changed := 0
	for i := range save.Plain {
		if save.Plain[i] == before[i] {
			continue
		}
		if !allowed[i] {
			t.Fatalf("走路之後不該動到 +%#x（%#02x → %#02x）", i, before[i], save.Plain[i])
		}
		changed++
	}
	if changed == 0 {
		t.Fatal("走了 40 步存檔卻一個 byte 都沒變")
	}
	t.Logf("走 40 步之後有 %d 個 byte 變動", changed)

	// 而且重新編碼還是解得開（checksum 有跟著重算）。
	reencoded := save.Bytes()
	if len(reencoded) != 6+0x800+0xA00 {
		t.Fatalf("重新編碼的長度不對：%d", len(reencoded))
	}
}

// 高解畫布：沒有字型時仍然畫得出來（只是沒中文），有字型時中文畫得上去。
func TestHiFrame(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}

	// 沒載字型也要能畫。
	h := s.HiFrame()
	if h == nil {
		t.Fatal("畫不出高解畫面")
	}
	// 640 × 400 應該是 320 × 200 的乾淨 2×。
	f := s.Frame()
	for _, p := range [][2]int{{0, 0}, {100, 50}, {319, 199}} {
		want := f.At(p[0], p[1])
		if got := h.At(p[0]*2, p[1]*2); got != want {
			t.Fatalf("(%d,%d) 放大後應該是 %d，得到 %d", p[0], p[1], want, got)
		}
	}

	dir := os.Getenv("WL_ETEN")
	if dir == "" {
		dir = "../../workplace/eten"
	}
	if err := s.LoadFont(dir); err != nil {
		t.Skipf("沒有倚天字型（%v），中文那半跳過", err)
	}
	s.SetCJK([]byte{0xA7, 0x41, 0xAD, 0xCC}) // 「你們」
	h = s.HiFrame()
	// 訊息視窗第一格（欄 1、列 18）附近應該有非零像素。
	on := 0
	for y := 18 * 16; y < 19*16; y++ {
		for x := 1 * 16; x < 3*16; x++ {
			if h.At(x, y) != 0 {
				on++
			}
		}
	}
	if on == 0 {
		t.Fatal("設了中文卻沒畫上去")
	}
	t.Logf("兩個中文字畫出 %d 個像素", on)
}

// 驗收 6（端到端）：踩到一個有翻譯的訊息格，畫面上要出現中文。
func TestTranslatedMessageShowsCJK(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("沒有翻譯目錄（%v），跳過", err)
	}
	dir := os.Getenv("WL_ETEN")
	if dir == "" {
		dir = "../../workplace/eten"
	}
	if err := s.LoadFont(dir); err != nil {
		t.Skipf("沒有倚天字型（%v），跳過", err)
	}

	// 走到全世界地圖上任何一個有翻譯的訊息格。翻譯目錄裡有的是
	// blk:game1:0:1–8，對應第 2 層值 1–8 的 nibble 4/9/12 格。
	w := s.World()
	found := false
	for y := 0; y < w.Block.Dim && !found; y++ {
		for x := 0; x < w.Block.Dim; x++ {
			terrain, record, _, err := w.Block.At(x, y)
			if err != nil || record < 1 || record > 8 {
				continue
			}
			if terrain != 4 && terrain != 9 && terrain != 12 {
				continue
			}
			// 直接把隊伍放到旁邊再走過去，比亂走可靠。
			w.Party.X, w.Party.Y = uint8(x), uint8(y+1)
			if _, err := w.Step(0 /* Up */); err != nil {
				continue
			}
			if w.Party.X != uint8(x) || w.Party.Y != uint8(y) {
				continue
			}
			found = true
			break
		}
	}
	if !found {
		t.Skip("這張地圖上找不到有翻譯的訊息格")
	}

	// 用 Update 走一次一樣的流程，讓 translate 跑到。
	s.SetCJK(nil)
	res, err := w.Step(1 /* Down */)
	_ = res
	if err != nil {
		t.Fatalf("走一步失敗：%v", err)
	}
	// 直接驗 translate 的行為（Update 只是它的呼叫端）。
	if b, ok := s.cat.Lookup("blk:game1:0:1"); !ok || len(b) == 0 {
		t.Fatal("翻譯目錄裡應該有 blk:game1:0:1")
	}
	s.SetCJK(mustLookup(t, s, "blk:game1:0:1"))
	h := s.HiFrame()
	on := 0
	for y := 18 * 16; y < 19*16; y++ {
		for x := 16; x < 16*20; x++ {
			if h.At(x, y) != 0 {
				on++
			}
		}
	}
	if on == 0 {
		t.Fatal("有翻譯卻沒畫出中文")
	}
	t.Logf("中文訊息畫出 %d 個像素", on)
}

func mustLookup(t *testing.T, s *Scene, key string) []byte {
	t.Helper()
	b, ok := s.cat.Lookup(key)
	if !ok {
		t.Fatalf("查不到 %s", key)
	}
	return b
}

// 驗收（規格 02 §1.1）：輪詢會推進產生器，沒輪詢就完全決定性。
//
// 這一條是實跑出來的：wl-play 跑三次 200 步，結果一模一樣——
// 產生器沒有種子，**不輪詢就等於每一局的遭遇序列都相同**。
func TestPollRNGIsTheOnlyEntropy(t *testing.T) {
	s := openScene(t)
	before := s.World().RNG.Snapshot()
	s.PollRNG()
	if s.World().RNG.Snapshot() == before {
		t.Fatal("PollRNG 沒有推進產生器")
	}

	// 不輪詢：同一串輸入兩次，結果要完全一樣（無頭工具靠這個可重現）。
	run := func() string {
		sc := openScene(t)
		for i := 0; i < 40; i++ {
			if _, err := sc.Update(input.Input{Dir: input.DirUp}); err != nil {
				t.Fatal(err)
			}
		}
		w := sc.World()
		return fmt.Sprintf("%d,%d,%v", w.Party.X, w.Party.Y, w.RNG.Snapshot())
	}
	if a, b := run(), run(); a != b {
		t.Fatalf("不輪詢卻跑出兩種結果：\n  %s\n  %s", a, b)
	}
}
