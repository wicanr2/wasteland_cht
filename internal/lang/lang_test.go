package lang

import (
	"unicode/utf8"
	"os"
	"testing"
)

func catPath() string {
	if p := os.Getenv("WL_CAT"); p != "" {
		return p
	}
	return "../../translations/zh-Hant.cat"
}

// 驗收 5：查得到的回中文、查不到的回 false。
func TestLoadAndLookup(t *testing.T) {
	c, err := Load(catPath())
	if err != nil {
		t.Skipf("沒有翻譯目錄（%v），跳過", err)
	}
	if c.Len() == 0 {
		t.Fatal("目錄是空的")
	}
	found := 0
	for _, key := range []string{ExeKey(2, 0), ExeKey(8, 1)} {
		if v, ok := c.Lookup(key); ok {
			found++
			if len(v) == 0 {
				t.Errorf("%s 查到了但內容是空的", key)
			}
			// 目錄裡是 **UTF-8**（Big5 只出現在畫的那一刻）。
			if !utf8.ValidString(v) {
				t.Errorf("%s 不是合法的 UTF-8：% x", key, v)
			}
		}
	}
	t.Logf("目錄共 %d 條，抽驗到 %d 條", c.Len(), found)

	if _, ok := c.Lookup("blk:game1:999:999"); ok {
		t.Fatal("不存在的 key 不該查得到")
	}
}

// nil 的目錄可以安全使用——沒有 .cat 檔時遊戲要照跑。
func TestNilCatalogueIsSafe(t *testing.T) {
	var c *Catalogue
	if _, ok := c.Lookup("exe:0:0"); ok {
		t.Fatal("nil 目錄不該查到東西")
	}
	if c.Len() != 0 {
		t.Fatal("nil 目錄的長度應該是 0")
	}
}

func TestLoadRejectsGarbage(t *testing.T) {
	tmp := t.TempDir() + "/bad.cat"
	if err := os.WriteFile(tmp, []byte("not a catalogue at all"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(tmp); err == nil {
		t.Fatal("亂七八糟的檔案應該被拒絕")
	}
	if _, err := Load(t.TempDir() + "/does-not-exist.cat"); err == nil {
		t.Fatal("不存在的檔案應該回錯誤")
	}
}

func TestKeyFormat(t *testing.T) {
	if got := BlockKey("game1", 12, 37); got != "blk:game1:12:37" {
		t.Errorf("BlockKey 錯了：%s", got)
	}
	if got := ExeKey(8, 1); got != "exe:8:1" {
		t.Errorf("ExeKey 錯了：%s", got)
	}
}
