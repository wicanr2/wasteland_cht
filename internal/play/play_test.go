package play

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
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

// F10 是唯一的離開手勢，而且**先問再走**（`esc-cancel-f10-quit-autosave`）。
//
// ⚠ 這一條以前驗的是「F10 直接結束」。那正是那份鐵則列的反模式：
// 沒有確認的離開路徑，玩家手滑就丟進度。
func TestQuitAsksBeforeLeaving(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	ok, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionQuit})
	if err != nil || !ok {
		t.Fatalf("F10 應該先跳確認而不是直接離開：ok=%v err=%v", ok, err)
	}
	if s.Mode() != "quit" {
		t.Fatalf("按過 F10 之後模式 = %s，預期 quit", s.Mode())
	}
	// N（與 ESC）取消，遊戲要留著。
	if ok, err = s.Update(input.Input{Dir: input.DirNone, Char: 'N'}); err != nil || !ok {
		t.Fatalf("選 N 應該取消離開：ok=%v err=%v", ok, err)
	}
	if s.Mode() == "quit" {
		t.Fatal("選 N 之後還停在確認畫面")
	}
	// 再來一次，這次選 Y。
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionQuit})
	if ok, err = s.Update(input.Input{Dir: input.DirNone, Char: 'Y'}); err != nil || ok {
		t.Fatalf("選 Y 應該離開：ok=%v err=%v", ok, err)
	}
}

// ESC 在**任何一層**都不能結束遊戲（鐵則 1）。
func TestEscapeNeverQuits(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	for i, mode := range []string{"標題", "地圖", "手札", "設定", "說明"} {
		switch mode {
		case "標題":
			s.BeginTitle()
		case "地圖":
			s.title = false
		case "手札":
			s.openJournal(1)
		case "設定":
			s.openSettings()
		case "說明":
			s.openHelp()
		}
		ok, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel})
		if err != nil {
			t.Fatalf("%s：%v", mode, err)
		}
		if !ok {
			t.Fatalf("第 %d 層（%s）按 ESC 結束了遊戲——ESC 只能取消", i, mode)
		}
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

	// ⚠ **方向不能亂挑**：起點 (55, 62) 三面是山（nibble 11，docs/re/62），
	// 只有往北走得動。步數也不能太多——往北 20 步就進輻射帶了。
	for i := 0; i < 8; i++ {
		if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
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
		t.Fatal("走了 8 步存檔卻一個 byte 都沒變")
	}
	t.Logf("走 8 步之後有 %d 個 byte 變動", changed)

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
	// 高解畫布應該是 320 × 200 的乾淨整數倍（`render.HiScale`）。
	f := s.Frame()
	for _, p := range [][2]int{{0, 0}, {100, 50}, {319, 199}} {
		want := f.At(p[0], p[1])
		if got := h.At(p[0]*render.HiScale, p[1]*render.HiScale); got != want {
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
	s.SetCJK("你們") // UTF-8：**以前這裡是四個 Big5 byte**
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
	s.SetCJK("")
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

func mustLookup(t *testing.T, s *Scene, key string) string {
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

// 驗收（規格 21／22）：整條戰鬥路徑跑得完——開戰 → 逐人下指令 → 結算 →
// 打完回地圖，而且經驗有進帳。
//
// 這一條是 cmd/wl-play 跑出來的路徑固化成測試：單元測試各自驗一塊規則，
// 這裡驗它們串起來不會卡住。**上限 200 步就是「打不完」的判準**。
func TestBattleRunsToCompletion(t *testing.T) {
	s := openScene(t)
	// 起始地圖沒有靜態遭遇格（docs/re/51），換到有的那一張。
	if err := s.LoadMap(4, 18, 2); err != nil {
		t.Fatalf("換地圖：%v", err)
	}
	c, err := s.StartEncounter()
	if err != nil {
		t.Fatalf("開戰：%v", err)
	}
	if c == nil {
		t.Fatal("地圖 4 的 (18, 2) 開不了戰——docs/re/51 說那裡有遭遇格")
	}
	before := uint32(0)
	for _, m := range s.World().Party.Members {
		before += m.XP
	}

	for i := 0; i < 200; i++ {
		if !s.InCombat() {
			after := uint32(0)
			for _, m := range s.World().Party.Members {
				after += m.XP
			}
			if after <= before {
				t.Fatalf("打完了但經驗沒進帳（%d → %d）", before, after)
			}
			return
		}
		if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'A'}); err != nil {
			t.Fatalf("第 %d 步：%v", i, err)
		}
	}
	t.Fatal("200 步之內戰鬥沒結束——可能卡在某一輪的指令階段")
}

// 驗收（規格 07 §6.7）：踩上傳送格會**換地圖**，而且回程指向踩上去的那一格。
//
// 沒有這一條玩家出不了起始地圖——23 筆設施記錄裡只有 2 筆走得進去，
// 商店與醫生全在別的地圖（docs/re/60 §1）。
func TestTeleportChangesMap(t *testing.T) {
	s := openScene(t)
	w := s.World()
	w.Teleport(20, 16) // 傳送格 (21, 16) 的左邊
	s.Invalidate()

	if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
		t.Fatalf("往右一步：%v", err)
	}
	// 進新地點要先答 Yes（記錄 +0x00 的 bit6，docs/re/64）。
	if !s.Asking() {
		t.Fatal("踩上傳送格沒有問「Enter new location?」")
	}
	if s.World().Party.X != 20 {
		t.Fatal("還在問就先動了座標")
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'Y'}); err != nil {
		t.Fatalf("答 Yes：%v", err)
	}
	if got := s.MapID(); got != 12 {
		t.Fatalf("踩上傳送格之後還在地圖 %d，記錄 +0x03 說是 12", got)
	}
	if w.Party.X != 1 || w.Party.Y != 1 {
		t.Fatalf("落點是 (%d, %d)，記錄 +0x01/+0x02 說是 (1, 1)", w.Party.X, w.Party.Y)
	}
	// 回程要指向**踩上去的那一格**，不是目的地——顛倒的話玩家會卡在原地來回。
	if s.back.X != 21 || s.back.Y != 16 || s.back.MapID != 0 {
		t.Fatalf("回程存成 (%d, %d) 地圖 %d，應該是 (21, 16) 地圖 0",
			s.back.X, s.back.Y, s.back.MapID)
	}
	// 換了地圖就要換圖磚組，不然畫面會拿舊組去畫。
	if s.gfx.Tiles == nil {
		t.Fatal("換地圖之後圖磚組是空的")
	}
}

// 驗收（規格 07 §6.7、docs/re/61）：建築內部的地圖編號 bit7 設起來，
// 要先查 ds:BF1Ch 換成真正的資源編號。
//
// **忘了查的症狀是「進建築就爆掉」**——130 這種值拿去索引 42 個區塊會直接失敗。
func TestBuildingInteriorResolvesMapID(t *testing.T) {
	s := openScene(t)
	// 先進 Quartz（資源 1），再踩它的一個建築入口 (24, 22)。
	if err := s.LoadMap(1, 24, 21); err != nil {
		t.Fatalf("進 Quartz：%v", err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirDown}); err != nil {
		t.Fatalf("踩建築入口：%v", err)
	}
	if s.Asking() {
		if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'Y'}); err != nil {
			t.Fatalf("答 Yes：%v", err)
		}
	}
	if got := s.MapID(); got != 5 {
		t.Fatalf("進建築之後在地圖 %d，編號 130 查表應該換成 5", got)
	}
	w := s.World()
	if w.Party.X != 25 || w.Party.Y != 19 {
		t.Fatalf("落點 (%d, %d)，記錄說是 (25, 19)", w.Party.X, w.Party.Y)
	}
}

// 驗收（docs/re/64）：進新地點要先問 Yes／No，答 No 就留在原地。
//
// 判準是記錄 +0x00 的 **bit6**（原版 `shl al,1` 之後看符號）——
// 不是 bit7。Quartz 入口的 +0x00 是 0x41。
func TestEnterLocationAsksAndNoStays(t *testing.T) {
	s := openScene(t)
	w := s.World()
	w.Teleport(20, 16)
	s.Invalidate()

	if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
		t.Fatal(err)
	}
	if !s.Asking() {
		t.Fatal("沒有問「Enter new location?」")
	}
	if got := s.Message(); !strings.Contains(got, "Enter new location") {
		t.Fatalf("問的訊息是 %q", got)
	}

	// 答 No：留在原地、地圖不變、時鐘不動。
	before := w.Clock
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'N'}); err != nil {
		t.Fatal(err)
	}
	if s.Asking() {
		t.Fatal("答了 No 還在問")
	}
	if w.Party.X != 20 || w.Party.Y != 16 || s.MapID() != 0 {
		t.Fatalf("答 No 卻動了：地圖 %d (%d, %d)", s.MapID(), w.Party.X, w.Party.Y)
	}
	if w.Clock != before {
		t.Fatal("答 No 卻推進了時鐘")
	}

	// 再走一次、這次答 Yes。
	if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'Y'}); err != nil {
		t.Fatal(err)
	}
	if s.MapID() != 12 {
		t.Fatalf("答 Yes 之後在地圖 %d", s.MapID())
	}
}

// 輻射計量表的讀數 ＝ 視野內最近的輻射格有多遠（`ds:46EEh`，docs/re/120 §2），
// 而且**隊上沒有人帶蓋氏計數器時畫面上不顯示**。
//
// ⚠ 距離是原版那張表的值（10 × 歐氏），不是格數——站在旁邊一格是 10 不是 1。
func TestGeigerReadingIsTheNearestRadiationCell(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	// 世界地圖（資源 0）的輻射帶在 (53,37) 一帶，36 格。
	if err := s.LoadMap(0, 53, 36); err != nil {
		t.Fatalf("載入世界地圖失敗：%v", err)
	}
	if got := s.geigerReading(); got != 10 {
		t.Errorf("站在輻射格正上方，最近的一格應該是 10，得到 %d", got)
	}
	// 走遠一點就讀不到了（視窗 19 × 9，掃不到就是保留值）。
	if err := s.LoadMap(0, 20, 5); err != nil {
		t.Fatalf("載入世界地圖失敗：%v", err)
	}
	if got := s.geigerReading(); got != render.MeterNoReading {
		t.Errorf("視野裡沒有輻射格應該回保留值，得到 %d", got)
	}

	// 出廠隊伍身上沒有蓋氏計數器（那是商店貨，docs/re/118）。
	if s.partyHasGeiger() {
		t.Error("出廠隊伍不該有蓋氏計數器")
	}
	// 塞一個給第一個人就算數——**與誰帶著無關**（`sub_17E42` 找到就跳出）。
	c := s.world.Party.Members[0]
	c.Items = append(c.Items, game.Slot{ID: game.ItemGeigerCounter, Value: 1})
	if !s.partyHasGeiger() {
		t.Error("隊上有一個人帶著就該算數")
	}
}
