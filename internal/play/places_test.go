package play

import (
	"strings"
	"testing"


	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestFacilitySignsAreTranslated：42 張地圖上每一塊設施招牌都查得到中文。
//
// 招牌不在字串表裡（是地圖記錄裡的明文 ASCII），所以翻譯覆蓋率的那把尺
// 量不到它——漏一塊的症狀是「走進某家店招牌是英文」，沒有任何錯誤訊息。
func TestFacilitySignsAreTranslated(t *testing.T) {
	s := sceneWithCatalogue(t)
	rom := s.rom
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	var missing []string
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for rec := 0; rec < 16; rec++ {
			raw, err := blk.SectionRecord(6, rec)
			if err != nil || len(raw) == 0 {
				continue
			}
			f, ok := game.ParseFacility(raw)
			if !ok {
				continue
			}
			name := strings.TrimSpace(f.Name)
			// ⚠ **沒有格子指到的 section 記錄讀出來是雜訊**（`cmd/wl-atlas`
			// 踩過同一個坑）：那些「招牌」是隨機位元組，不是漏翻。
			// 招牌照定義是可印 ASCII，不是的就不算。
			if name == "" || seen[name] || !printableASCII(name) {
				continue
			}
			seen[name] = true
			if len(s.placeCJK(f.Name)) == 0 {
				missing = append(missing, name)
			}
		}
	}
	if len(seen) == 0 {
		t.Fatal("一塊招牌都沒掃到，這一條沒驗到東西")
	}
	if len(missing) > 0 {
		t.Errorf("%d／%d 塊招牌沒有中文：%s",
			len(missing), len(seen), strings.Join(missing, "、"))
	}
	t.Logf("%d 塊招牌全部查得到中文", len(seen))
}

// TestNoStalePlaceKeys：目錄裡的每一條招牌都要對得到遊戲資料裡真的存在的招牌。
//
// 與 TestFacilitySignsAreTranslated 是**反向**的一條：那條抓漏翻，
// 這條抓多寫的。招牌是手寫進 tsv 的，資料改了或當初打錯字時，
// 多出來的那條永遠查不到，而且不會有任何症狀。
func TestNoStalePlaceKeys(t *testing.T) {
	s := sceneWithCatalogue(t)
	signs := map[string]bool{}
	resources, err := s.rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	for _, res := range resources {
		blk, err := s.rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for rec := 0; rec < 16; rec++ {
			raw, err := blk.SectionRecord(6, rec)
			if err != nil || len(raw) == 0 {
				continue
			}
			if f, ok := game.ParseFacility(raw); ok {
				if n := strings.TrimSpace(f.Name); n != "" && printableASCII(n) {
					signs[n] = true
				}
			}
		}
	}
	// 存檔裡的地點名也算（開場那一句）。
	if s.save != nil {
		if n := strings.TrimSpace(s.save.Place()); n != "" {
			signs[n] = true
		}
	}

	var stale []string
	s.cat.Each(func(key string, _ []byte) {
		name, ok := strings.CutPrefix(key, "place:")
		if !ok {
			return
		}
		if !signs[name] {
			stale = append(stale, name)
		}
	})
	if len(stale) > 0 {
		t.Errorf("%d 條招牌譯文對不到資料裡的招牌：%s",
			len(stale), strings.Join(stale, "、"))
	}
	t.Logf("資料裡 %d 塊招牌，目錄裡沒有多餘的", len(signs))
}

// TestSignLookupTrimsPadding：招牌在資料裡是補過空白的定長欄位。
//
// `Ranger Ctr.` 實際上是 `"Ranger Ctr. "`。不修掉就永遠查不到，
// 而查不到是**靜靜顯示英文**——看起來像「還沒翻」而不像壞掉。
func TestSignLookupTrimsPadding(t *testing.T) {
	s := sceneWithCatalogue(t)
	want := s.placeCJK("Ranger Ctr.")
	if len(want) == 0 {
		t.Fatal("連沒有空白的名字都查不到")
	}
	for _, padded := range []string{"Ranger Ctr. ", " Ranger Ctr.", "  Ranger Ctr.  "} {
		if got := s.placeCJK(padded); string(got) != string(want) {
			t.Errorf("%q 查不到（資料裡就是這樣補空白的）", padded)
		}
	}
}

// TestRangerCentreSignShowsCJK：走進遊俠中心，招牌那一行要是中文。
//
// 從**玩家的第一個按鍵**進場，不從場景中間的函式進去：
// 從中間進場的測試驗不到接線，而接線斷掉正是這一類缺口的成因。
func TestRangerCentreSignShowsCJK(t *testing.T) {
	s := sceneWithCatalogue(t)
	w := s.World()
	w.Teleport(55, 63)
	s.Invalidate()
	step(t, s, input.Input{Dir: input.DirUp})

	if s.facility == nil {
		t.Fatalf("沒進到設施，停在 %s", s.Mode())
	}
	if len(s.facility.CJKLines) == 0 || len(s.facility.CJKLines[0]) == 0 {
		t.Fatalf("招牌那一行沒有中文（英文是 %q）", s.facility.Lines[0])
	}
	want := s.placeCJK("Ranger Ctr.")
	if string(s.facility.CJKLines[0]) != string(want) {
		t.Errorf("招牌是 % X，預期 % X", s.facility.CJKLines[0], want)
	}
}

// TestRosterHeaderIsTranslated：戰鬥名單的表頭走 `ui:combat.hdr*`。
//
// 表頭是原版寫死的 ASCII（`RosterHeader`），字串表裡沒有它，
// 所以它是**翻譯覆蓋率量不到的第二個地方**。
func TestRosterHeaderIsTranslated(t *testing.T) {
	s := sceneWithCatalogue(t)
	hdr := rosterHeaderCJK(s.uiText)
	if len(hdr) == 0 {
		t.Fatal("表頭組不出中文")
	}
	// 六個標籤都要在，而且欄位要落在值的欄座標上。
	if n := cjkCells(hdr); n > rosterCols {
		t.Errorf("表頭佔 %d 格，超過 %d 格的名單寬度", n, rosterCols)
	}
	for _, want := range []struct {
		col  int
		name string
	}{
		{colName, "combat.hdrname"}, {colAC, "combat.hdrac"},
		{colAmmo, "combat.hdrammo"}, {colMaxCON, "combat.hdrmax"},
		{colCON, "combat.hdrcon"}, {colWeapon, "combat.hdrweapon"},
	} {
		label := s.uiText(want.name)
		at := cellIndex(hdr, want.col)
		if at < 0 || !strings.HasPrefix(string(hdr[at:]), string(label)) {
			t.Errorf("欄 %d 不是 %s（% X）", want.col, want.name, label)
		}
	}
}

// printableASCII 判斷這串是不是全部可印 ASCII。
func printableASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] > 0x7E {
			return false
		}
	}
	return true
}

// cellIndex 把「第 n 格」換算成 Big5 串裡的 byte 位置。
func cellIndex(b []byte, cell int) int {
	n := 0
	for i := 0; i < len(b); i++ {
		if n == cell {
			return i
		}
		if b[i] >= 0x80 {
			i++
		}
		n++
	}
	return -1
}

// sceneWithCatalogue 開一個載了翻譯目錄的場景。
//
// 招牌與表頭這幾條**沒有目錄就什麼都驗不到**，所以目錄載不到直接 skip，
// 不要讓它們變成「綠燈但沒測到東西」。
func sceneWithCatalogue(t *testing.T) *Scene {
	t.Helper()
	s := newScene(t)
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("沒有翻譯目錄（%v）", err)
	}
	return s
}
