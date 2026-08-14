package game

import (
	"os"
	"path/filepath"
	"testing"
)

const refsPath = "../../docs/re/generated/paragraph-refs.tsv"

func loadRefs(t *testing.T) ParagraphRefs {
	t.Helper()
	refs, err := LoadParagraphRefs(refsPath)
	if err != nil {
		t.Fatalf("讀引用表失敗：%v", err)
	}
	return refs
}

// 驗收 1／2：引用表是從英文原文抽的；82 個編號都在，執行期拼的那條不在。
func TestParagraphRefs(t *testing.T) {
	refs := loadRefs(t)
	if len(refs) != 83 {
		t.Errorf("應該有 83 條引用（docs/re/33 §1），得到 %d", len(refs))
	}
	seen := map[int]bool{}
	for _, n := range refs {
		if n < 1 || n > ParagraphCount {
			t.Fatalf("段落編號 %d 超出 1–%d", n, ParagraphCount)
		}
		seen[n] = true
	}
	if len(seen) != 82 {
		t.Errorf("應該是 82 個不同編號，得到 %d", len(seen))
	}
	// ⚠ 這一條的編號是選單選完才拼上去的，不該被猜一個填進去。
	if _, ok := refs["blk:game2:15:61"]; ok {
		t.Error("blk:game2:15:61 的編號是執行期拼的，不該出現在表裡")
	}
}

// 驗收 3：兩區加起來是完整的 162 段，一段都不刪。
func TestJournalSectionsCoverEverything(t *testing.T) {
	j := NewJournal(loadRefs(t))
	main, appendix := j.Entries(SectionMain), j.Entries(SectionAppendix)
	if len(main) != 82 {
		t.Errorf("正文應該是 82 段，得到 %d", len(main))
	}
	if len(main)+len(appendix) != ParagraphCount {
		t.Fatalf("兩區加起來應該是 %d 段，得到 %d + %d",
			ParagraphCount, len(main), len(appendix))
	}
	all := map[int]bool{}
	for _, n := range append(append([]int{}, main...), appendix...) {
		if all[n] {
			t.Fatalf("第 %d 段出現在兩區", n)
		}
		all[n] = true
	}
	for n := 1; n <= ParagraphCount; n++ {
		if !all[n] {
			t.Fatalf("第 %d 段兩區都沒有", n)
		}
	}
}

// 驗收 4：陷阱段落落在附錄，而且標得出來。
func TestTrapParagraphsAreInAppendix(t *testing.T) {
	j := NewJournal(loadRefs(t))
	for _, n := range []int{1, 22, 145} {
		if !j.IsTrap(n) {
			t.Errorf("第 %d 段應該標成陷阱", n)
		}
		if j.Section(n) != SectionAppendix {
			t.Errorf("第 %d 段（陷阱）應該在附錄", n)
		}
	}
	// 一般段落不該被誤標。
	if j.IsTrap(23) {
		t.Error("第 23 段不是陷阱")
	}
}

// 驗收 5：已讀只影響顯示，不擋著不讓翻。
func TestJournalReadIsAMarkerNotALock(t *testing.T) {
	j := NewJournal(loadRefs(t))
	// 沒讀過也翻得到——手札預設全開。
	if len(j.Entries(SectionMain)) == 0 {
		t.Fatal("一開始就該翻得到正文")
	}
	if j.WasRead(23) {
		t.Error("一開始不該有已讀標記")
	}
	var key string
	for k, n := range j.refs {
		if n == 23 {
			key = k
		}
	}
	if key == "" {
		t.Skip("這份引用表裡沒有第 23 段")
	}
	n, ok := j.Trigger(key)
	if !ok || n != 23 {
		t.Fatalf("觸發應該回第 23 段，得到 %d（ok=%v）", n, ok)
	}
	if !j.WasRead(23) {
		t.Error("觸發之後該記成已讀")
	}
	// 觸發前後翻得到的段落數不變。
	if len(j.Entries(SectionMain)) != 82 {
		t.Error("已讀不該改變翻得到的段落數")
	}
	if _, ok := j.Trigger("blk:game1:99:99"); ok {
		t.Error("沒有引用的 key 不該觸發")
	}
}

// 驗收 6：沒有中文就顯示英文；兩邊都沒有時不顯示編號充數。
func TestParagraphTextFallback(t *testing.T) {
	zh := map[int]string{1: "中文"}
	en := map[int]string{1: "English", 2: "Only English"}
	if s, ok := Text(zh, en, 1); !ok || s != "中文" {
		t.Errorf("有中文就該用中文，得到 %q", s)
	}
	if s, ok := Text(zh, en, 2); !ok || s != "Only English" {
		t.Errorf("沒中文該退回英文，得到 %q", s)
	}
	if _, ok := Text(zh, en, 3); ok {
		t.Error("兩邊都沒有時該回 false，不要拿編號充數")
	}
	// 空字串當成沒有。
	if _, ok := Text(map[int]string{4: ""}, map[int]string{}, 4); ok {
		t.Error("空字串該當成沒有")
	}
}

// 壞掉的引用表要報錯，不要靜靜載入一半。
func TestLoadParagraphRefsRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	for _, body := range []string{"沒有TAB的一行\n", "blk:a:1:1\t不是數字\n"} {
		p := filepath.Join(dir, "bad.tsv")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadParagraphRefs(p); err == nil {
			t.Errorf("%q 應該報錯", body)
		}
	}
	if _, err := LoadParagraphRefs(filepath.Join(dir, "沒有這個檔")); err == nil {
		t.Error("檔案不存在應該報錯")
	}
}
