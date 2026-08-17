package play

import (
	"sort"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// TestCatalogueRendersWithoutControlCodes：**每一條譯文解完之後都不該再有控制碼**。
//
// 譯文照原版保留了整套控制碼（`\x0A` 單複數近四百處、`\x0E` 三選一、
// `\x0B` 名字、`\x0F` 數量、`\x10` 熱鍵標記）。中文那條路以前是把原始位元組
// 直接送去畫，於是 `\x0A`（就是 `\n`）變成換行、單複數**兩段都印出來**，
// 其餘控制碼變成畫面上的怪字元——看起來像「這句翻得很怪」而不是像壞掉。
//
// 這一道守的是「解碼有沒有漏掉哪一個碼」：解完只准剩下 `\n` 與可見位元組。
func TestCatalogueRendersWithoutControlCodes(t *testing.T) {
	cat, err := lang.Load("../../translations/zh-Hant.cat")
	if err != nil {
		t.Skipf("載不到翻譯目錄：%v", err)
	}
	bad := map[rune][]string{}
	total := 0
	cat.Each(func(key, raw string) {
		total++
		out := textlayout.Render(raw, textlayout.Options{})
		for _, c := range out {
			if c >= 0x20 || c == '\n' {
				continue
			}
			if len(bad[c]) < 3 {
				bad[c] = append(bad[c], key)
			}
		}
	})
	if total == 0 {
		t.Fatal("目錄是空的")
	}
	if len(bad) == 0 {
		t.Logf("%d 條譯文全部解得乾淨", total)
		return
	}
	codes := make([]int, 0, len(bad))
	for c := range bad {
		codes = append(codes, int(c))
	}
	sort.Ints(codes)
	for _, c := range codes {
		t.Errorf("控制碼 %#02x 解完還留在畫面上，例如 %v", c, bad[rune(c)])
	}
}
