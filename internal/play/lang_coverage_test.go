package play

// 翻譯覆蓋率的量測與門檻寫在 `docs/re/83`。
import (
	"github.com/wicanr2/wasteland_cht/internal/input"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/lang"
)

// TestTranslationCoverage 盤點中文化：4,800 多條原文裡有幾條查得到中文，
// 以及**目錄裡有沒有對不上任何原文的孤兒 key**。
//
// 孤兒 key 是白做的翻譯：檔案裡有、遊戲永遠查不到。
// 它不會讓任何測試變紅，因為 `Lookup` 查不到只是回原文。
func TestTranslationCoverage(t *testing.T) {
	rom := openRom(t)
	cat, err := lang.Load("../../translations/zh-Hant.cat")
	if err != nil {
		t.Skipf("載不到翻譯目錄：%v", err)
	}

	// 刻意不翻的那些（`translations/untranslatable.tsv`）：
	// 純控制碼、空白、未用槽的解碼雜訊——原文就不是文字。
	skip := map[string]bool{}
	raw, err := os.ReadFile("../../translations/untranslatable.tsv")
	if err != nil {
		t.Fatalf("讀不到 untranslatable.tsv：%v", err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		skip[strings.SplitN(line, "\t", 2)[0]] = true
	}

	seen := map[string]bool{} // 這一輪對得上原文的 key
	var missing []string
	var total, hit int

	// 九張執行檔字串表
	tables, err := rom.ExeStrings()
	if err != nil {
		t.Fatalf("ExeStrings：%v", err)
	}
	for ti, tbl := range tables {
		for si, s := range tbl {
			if s == "" {
				continue
			}
			total++
			key := lang.ExeKey(ti, si)
			seen[key] = true
			if _, ok := cat.Lookup(key); ok {
				hit++
			} else if !skip[key] {
				missing = append(missing, key)
			}
		}
	}
	exeTotal, exeHit := total, hit

	// 42 個地圖區塊
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for slot, s := range blk.Strings {
			if s == "" {
				continue
			}
			total++
			key := lang.BlockKey(res.File, res.ID, slot)
			seen[key] = true
			if _, ok := cat.Lookup(key); ok {
				hit++
			} else if !skip[key] {
				missing = append(missing, key)
			}
		}
	}

	t.Logf("執行檔：%d／%d 條有中文", exeHit, exeTotal)
	t.Logf("地圖區塊：%d／%d 條有中文", hit-exeHit, total-exeTotal)
	t.Logf("合計：%d／%d 條（%.1f%%），目錄共 %d 條",
		hit, total, 100*float64(hit)/float64(total), cat.Len())

	// `ui:` 那組是**重製版自己的介面文字**，原版寫成 ASCII 字面值不走字串表，
	// 所以對不上任何原文——它們不是孤兒。逐條列出來檢查，
	// **不要整個前綴放行**：打錯一個字的 `ui:` key 一樣查不到，
	// 而症狀只是畫面上那一處是英文。
	uiKeys := 0
	for _, k := range uiCatalogueKeys {
		if _, ok := cat.Lookup(lang.UIKey(k)); !ok {
			t.Errorf("介面文字 ui:%s 沒有翻譯", k)
			continue
		}
		uiKeys++
	}

	// `place:` 與 `monster:` 兩組是**明文名字**（地圖記錄、存檔與地圖的
	// 明文名字表，都不在字串表裡）。這裡只把數量扣掉；「有沒有多餘或漏掉的」
	// 由 places_test.go 與 monsters_test.go 對著遊戲資料雙向驗——
	// 那比在這裡列一份手抄清單可靠。
	placeKeys := 0
	cat.Each(func(key string, _ string) {
		if strings.HasPrefix(key, "place:") || strings.HasPrefix(key, "monster:") {
			placeKeys++
		}
	})

	orphans := cat.Len() - hit - uiKeys - placeKeys
	t.Logf("介面文字（ui:）：%d 條", uiKeys)
	t.Logf("明文名字（place:／monster:）：%d 條", placeKeys)
	t.Logf("目錄裡對不上任何原文的 key：%d 條", orphans)
	if orphans < 0 {
		t.Fatalf("hit(%d) 超過目錄長度(%d)——key 算法有重複", hit, cat.Len())
	}
	// 孤兒 key 是白做的翻譯，一條都不該有。
	if orphans != 0 {
		t.Errorf("目錄裡有 %d 條孤兒 key（遊戲永遠查不到）", orphans)
	}
	// 沒翻到的必須全部在 untranslatable 清單裡——**「還沒翻」與
	// 「刻意不翻」要分得開**，不然覆蓋率永遠停在 99% 而沒有人知道差在哪。
	if len(missing) != 0 {
		t.Errorf("有 %d 條既沒中文也不在 untranslatable 清單裡，例如 %v",
			len(missing), missing[:min(5, len(missing))])
	}
	t.Logf("刻意不翻的 %d 條（untranslatable.tsv）", len(skip))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 走一步之後訊息要真的變成中文——**翻譯目錄有譯文不等於畫面上看得到**。
//
// key 的 slot 是**字串編號**（`Event.Strings`），不是記錄編號。
// 拿錯會每次都查不到，而症狀只是「畫面上是英文」，看起來像還沒翻。
func TestWalkShowsTranslatedMessage(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("沒有翻譯目錄（%v），跳過", err)
	}
	if err := s.LoadMap(4, 18, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
		t.Fatal(err)
	}
	if s.Message() == "" {
		t.Fatal("走上去沒有任何訊息——這一格應該有")
	}
	if len(s.cjk) == 0 {
		t.Errorf("訊息 %q 在目錄裡有譯文，畫面上卻是英文", s.Message())
	}
	t.Logf("英文 %q → 中文 %d bytes", s.Message(), len(s.cjk))
}

// uiCatalogueKeys 是重製版介面文字的完整清單（`translations/*/ui.tsv`）。
//
// 這一份與 tsv 是兩份要一起改的東西——**多一條會被當成孤兒、少一條會漏檢**。
var uiCatalogueKeys = []string{
	"cmd.bar",
	"journal.header", "journal.hint",
	"journal.decoy", "journal.appendix", "journal.epilogue",
	"journal.untranslated",
	"use.which", "use.kind", "use.whichway", "use.nothing", "use.nobody", "use.morepage",
	"use.nothinghappens", "use.works", "use.fails",
	"gate.party", "gate.damage",
	"combat.attacked", "combat.hit", "combat.fallen", "combat.over",
	"combat.which", "combat.partyxp", "hire.joins",
	"enc.inside", "enc.none", "enc.groupnone",
	"save.none", "save.done", "save.donenonames", "save.memoryonly", "save.nodata",
	"radio.nothing", "view.none", "view.switched",
	"disband.single", "disband.left", "order.nothing", "order.done",
	"roster.menu", "roster.name", "roster.nobody", "roster.joined", "roster.gone",
	"facility.examined", "facility.noheal", "facility.skillfull", "facility.cured", "facility.bought", "facility.sold",
	"facility.learned", "facility.row", "facility.skillrow",
	"facility.sellrow", "facility.buyrow", "facility.more",
	// F1 說明、F2 設定、F10 離開確認、F5／F9 快速存讀檔（重製版自己加的）。
	"combat.hdrname", "combat.hdrac", "combat.hdrammo",
	"combat.hdrmax", "combat.hdrcon", "combat.hdrweapon",
	// 名片行的傷勢狀態字（`docs/re/17` §4.4）。**死亡那一格不在這裡**——
	// 它是字型第 `0x7F` 格的骷髏字模，不是可以翻譯的文字。
	"wound.unc", "wound.ser", "wound.crt", "wound.mrt", "wound.com",
	"help.title", "help.move", "help.cmdbar",
	"help.panels", "help.f5f9", "help.quit",
	"settings.title", "settings.music", "settings.volume", "settings.sfx",
	"settings.retro", "settings.modern",
	"settings.on", "settings.off", "settings.close",
	"quit.ask", "quit.savefailed",
	"quick.saved", "quick.loaded", "quick.savefailed", "quick.loadfailed",
	// 全隊倒下的死亡畫面（`docs/spec/28`）。這兩句在原版是**明文 ASCII**，
	// 不在打包字串表裡，所以走 `ui:`。
	"wipe.place", "wipe.message",
}
