package game

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// 段落手札（docs/spec/19、docs/re/33）。
//
// 原版把 162 段劇情印在紙本上，遊戲裡只給編號。**防拷不移植**：
// 玩家不必翻紙本、不做編號驗證——讀到編號就直接把正文顯示出來。

// ParagraphCount 是段落書的總段數。
const ParagraphCount = 162

// 陷阱段落：段落自己明講「你不該讀到這裡」，用來抓沒買正版的人。
// 遊戲**永遠不會**叫玩家去讀它們（docs/re/33 §3，一手資料）。
var trapParagraphs = map[int]bool{1: true, 22: true, 145: true}

// JournalSection 是手札的分區。
type JournalSection int

const (
	// SectionMain 是遊戲會引用的 82 段——實際的劇情主體。
	SectionMain JournalSection = iota
	// SectionAppendix 是其餘 80 段（陷阱、沒被用到的變體、火星誘餌）。
	// **保存**：它們是 1988 年防拷設計的一部分，不能刪。
	SectionAppendix
)

// ParagraphRefs 是「字串 key → 段落編號」的引用表。
//
// ⚠ 這張表是**編譯期**從英文原文抽的（tools/extract_paragraph_refs.py）。
// 不要在執行期解析翻譯過的文字——翻完之後那句話變成「請看第 23 段。」，
// 格式隨譯者而變，等於讓譯者的用字決定遊戲讀不讀得到段落。
type ParagraphRefs map[string]int

// LoadParagraphRefs 讀引用表（`<key>\t<編號>`，# 開頭是註解）。
func LoadParagraphRefs(path string) (ParagraphRefs, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	refs := ParagraphRefs{}
	sc := bufio.NewScanner(f)
	for line := 1; sc.Scan(); line++ {
		t := sc.Text()
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		key, num, ok := strings.Cut(t, "\t")
		if !ok {
			return nil, fmt.Errorf("%s:%d：沒有 TAB", path, line)
		}
		n, err := strconv.Atoi(num)
		if err != nil {
			return nil, fmt.Errorf("%s:%d：段落編號 %q 不是數字", path, line, num)
		}
		refs[key] = n
	}
	return refs, sc.Err()
}

// Lookup 查一條敘述會不會叫玩家去讀段落。
func (r ParagraphRefs) Lookup(key string) (int, bool) {
	n, ok := r[key]
	return n, ok
}

// Journal 是遊戲內的段落手札。
//
// **預設全開**，忠於紙本——原版的紙本也是全開的，它靠的是前言那段
// 「請你忍住別亂翻」的自律。已讀只是標記，不是解鎖條件。
type Journal struct {
	refs ParagraphRefs
	used map[int]bool // 遊戲會引用的編號
	read map[int]bool // 玩家已經讀過的
}

// NewJournal 依引用表建一份手札。
func NewJournal(refs ParagraphRefs) *Journal {
	j := &Journal{refs: refs, used: map[int]bool{}, read: map[int]bool{}}
	for _, n := range refs {
		j.used[n] = true
	}
	return j
}

// Refs 回傳引用表，讓呈現層查「這一則訊息有沒有引用段落」。
func (j *Journal) Refs() ParagraphRefs { return j.refs }

// Section 回報某一段收在哪一區。
func (j *Journal) Section(n int) JournalSection {
	if j.used[n] {
		return SectionMain
	}
	return SectionAppendix
}

// IsTrap 回報這一段是不是陷阱段落（防拷設計，遊戲永遠不會引用）。
func (j *Journal) IsTrap(n int) bool { return trapParagraphs[n] }

// Entries 列出某一區的段落編號（由小到大）。兩區加起來是完整的 162 段。
func (j *Journal) Entries(s JournalSection) []int {
	var out []int
	for n := 1; n <= ParagraphCount; n++ {
		if j.Section(n) == s {
			out = append(out, n)
		}
	}
	sort.Ints(out)
	return out
}

// MarkRead 記下玩家讀過某一段。
func (j *Journal) MarkRead(n int) {
	if n >= 1 && n <= ParagraphCount {
		j.read[n] = true
	}
}

// WasRead 回報讀過沒有。**只影響顯示**，不擋著不讓翻。
func (j *Journal) WasRead(n int) bool { return j.read[n] }

// Trigger 收一條敘述的 key：有引用就記成已讀並回傳編號。
func (j *Journal) Trigger(key string) (int, bool) {
	n, ok := j.refs.Lookup(key)
	if !ok {
		return 0, false
	}
	j.MarkRead(n)
	return n, true
}

// Text 拿編號換正文。
//
// zh 是中文正文、en 是英文原文；沒有中文就回英文，與其餘文字的 fallback 一致
// （docs/spec/19 §5）。兩邊都沒有時回 false——**不要顯示編號充數**。
func Text(zh, en map[int]string, n int) (string, bool) {
	if s, ok := zh[n]; ok && s != "" {
		return s, true
	}
	if s, ok := en[n]; ok && s != "" {
		return s, true
	}
	return "", false
}
