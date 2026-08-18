package play

import (
	"sort"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/lang"
)

// monsterNamesInData 掃 42 張地圖的明文名字表，回傳去重後的原文。
//
// ⚠ **第 0 條不算**（表前面就是一個 NUL），**沒有字母的也不算**——
// 表尾之後偶爾會多解出一兩個位元組（`tools/extract_monster_names.py` 同一條）。
func monsterNamesInData(t *testing.T, s *Scene) []string {
	t.Helper()
	resources, err := s.rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, res := range resources {
		blk, err := s.rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for i, n := range blk.MonsterNames() {
			if i == 0 || !hasLetter(n) {
				continue
			}
			seen[n] = true
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func hasLetter(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}

// 每一條明文敵人名字都查得到中文。
//
// 這批名字不在字串表裡（`docs/re/114` §6），翻譯覆蓋率那把尺量不到——
// 漏一條的症狀是「某種敵人在戰鬥訊息裡是英文」，沒有任何錯誤訊息。
func TestMonsterNamesAreTranslated(t *testing.T) {
	s := sceneWithCatalogue(t)
	names := monsterNamesInData(t, s)
	if len(names) == 0 {
		t.Fatal("一條名字都沒掃到，這一條沒驗到東西")
	}
	var missing []string
	for _, n := range names {
		if s.monsterCJK(n) == "" {
			missing = append(missing, strings.ReplaceAll(n, "\n", "|"))
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d／%d 條名字沒有中文：%s",
			len(missing), len(names), strings.Join(missing, "、"))
	}
	t.Logf("%d 條明文敵人名字全部查得到中文", len(names))
}

// 反過來：目錄裡不該有對不上任何資料的 `monster:` key。
//
// 多出來的是白做的翻譯——它不會讓任何測試變紅，因為查不到只是回原文。
func TestNoOrphanMonsterKeys(t *testing.T) {
	s := sceneWithCatalogue(t)
	inData := map[string]bool{}
	for _, n := range monsterNamesInData(t, s) {
		inData[lang.MonsterKey(n)] = true
	}
	var orphans []string
	s.cat.Each(func(key string, _ string) {
		if strings.HasPrefix(key, "monster:") && !inData[key] {
			orphans = append(orphans, strings.ReplaceAll(key, "\n", "|"))
		}
	})
	sort.Strings(orphans)
	if len(orphans) > 0 {
		t.Errorf("目錄裡有 %d 條對不上資料的 monster key：%s",
			len(orphans), strings.Join(orphans, "、"))
	}
}

// 專有名詞的體例：**中文（英文）**（使用者定案 2026-08-18）。
//
// 這一道只守「有括號的那些，括號裡就是原文」——**哪一條算專有名詞是人的判斷**，
// 機器分不出 `Woman` 與 `Weez`。守得住的是格式一致，不是分類正確。
func TestProperNounFormat(t *testing.T) {
	s := sceneWithCatalogue(t)
	names := monsterNamesInData(t, s)
	withParen := 0
	for _, raw := range names {
		zh := s.monsterCJK(raw)
		if zh == "" || !strings.Contains(zh, "（") {
			continue
		}
		withParen++
		// 原文的字根 ＋ 單數字尾就是玩家在畫面上看到的英文。
		want := "（" + singular(raw) + "）"
		if !strings.Contains(zh, want) {
			t.Errorf("%q 的譯文 %q 括號裡不是原文 %q",
				strings.ReplaceAll(raw, "\n", "|"), zh, singular(raw))
		}
	}
	if withParen == 0 {
		t.Fatal("一條帶括號的譯文都沒有——體例沒套上去")
	}
	t.Logf("%d／%d 條照「中文（英文）」的體例", withParen, len(names))
}

// 地點招牌套同一套體例。
func TestPlaceSignFormat(t *testing.T) {
	s := sceneWithCatalogue(t)
	n := 0
	s.cat.Each(func(key string, zh string) {
		if !strings.HasPrefix(key, "place:") {
			return
		}
		n++
		en := strings.TrimPrefix(key, "place:")
		if !strings.Contains(zh, "（"+en+"）") {
			t.Errorf("招牌 %q 的譯文 %q 沒有「中文（英文）」的括號", en, zh)
		}
	})
	if n == 0 {
		t.Fatal("一塊招牌都沒掃到")
	}
	t.Logf("%d 塊招牌照體例", n)
}
