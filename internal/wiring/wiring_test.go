package wiring

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// 這一道檢查擋的是本專案漏過三次的同一件事：**RE 解出來了，卻沒人回頭接**
//（Agility 與對手行動值躺在 docs/re/21、37，敵人目標選擇躺在 docs/re/20 §1.1，
// 而 remake 三處都傳著寫死的常數，編得過、測得過、玩得動）。
//
// 判準只有一個、而且是機器數得到的：**程式碼或測試裡有沒有 `docs/re/NN` 的引用**。
// 這證明不了實作正確（那是各子系統門檻測試的事），但它擋得住「整份忘了接」。

const statusDoc = "docs/re/00-wiring-status.md"

type row struct {
	num    int
	status string
	note   string
	line   int
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// 表格列：| 07 | [標題](07-xxx.md) | 已接 | `internal/...` |
var rowRe = regexp.MustCompile(`^\|\s*(\d{1,3})\s*\|([^|]*)\|\s*([^|]+?)\s*\|(.*)\|\s*$`)

func parseStatus(t *testing.T, root string) []row {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, statusDoc))
	if err != nil {
		t.Fatalf("讀不到接線表：%v", err)
	}
	var out []row
	for i, line := range strings.Split(string(raw), "\n") {
		m := rowRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		out = append(out, row{
			num:    n,
			status: strings.TrimSpace(m[3]),
			note:   strings.TrimSpace(strings.Trim(m[4], "|")),
			line:   i + 1,
		})
	}
	if len(out) == 0 {
		t.Fatal("接線表一列都解不出來——表格式改了就要一起改這支測試")
	}
	return out
}

// notesOnDisk 回傳 docs/re/ 底下所有 RE 筆記的編號（不含 00-*）。
func notesOnDisk(t *testing.T, root string) map[int]string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "docs/re"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[int]string{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		i := strings.Index(name, "-")
		if i < 0 {
			continue
		}
		n, err := strconv.Atoi(name[:i])
		if err != nil || n == 0 {
			continue
		}
		out[n] = name
	}
	return out
}

// citations 掃程式碼、測試與工具，回傳每個編號被哪些檔案引用。
//
// ⚠ 掃描範圍與副檔名是這道檢查的**過濾器**——加了新的程式碼目錄卻沒加進來，
// 症狀會是「明明接上了卻說沒引用」。任何過濾閾值都會製造假零（CLAUDE.md §1.4）。
var citeRe = regexp.MustCompile(`docs/re/(\d{1,3})`)

func citations(t *testing.T, root string) map[int][]string {
	t.Helper()
	dirs := []string{"internal", "cmd", "tools", "translations"}
	exts := map[string]bool{".go": true, ".py": true, ".sh": true, ".tsv": true}
	out := map[int][]string{}
	for _, d := range dirs {
		base := filepath.Join(root, d)
		err := filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !exts[filepath.Ext(p)] {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			seen := map[int]bool{}
			for _, m := range citeRe.FindAllStringSubmatch(string(b), -1) {
				n, _ := strconv.Atoi(m[1])
				if n != 0 && !seen[n] {
					seen[n] = true
					out[n] = append(out[n], rel)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("掃 %s：%v", d, err)
		}
	}
	if len(out) == 0 {
		t.Fatal("一個引用都沒掃到——正對照失敗，先確認掃描範圍與正則")
	}
	return out
}

func TestWiringStatus(t *testing.T) {
	root := repoRoot(t)
	rows := parseStatus(t, root)
	notes := notesOnDisk(t, root)
	cites := citations(t, root)

	listed := map[int]row{}
	for _, r := range rows {
		if prev, dup := listed[r.num]; dup {
			t.Errorf("編號 %02d 在接線表出現兩次（第 %d 行與第 %d 行）", r.num, prev.line, r.line)
			continue
		}
		listed[r.num] = r
	}

	// 一：每份筆記都要登記。**新寫一份 RE 筆記卻沒接上，就是在這裡紅。**
	var missing []int
	for n := range notes {
		if _, ok := listed[n]; !ok {
			missing = append(missing, n)
		}
	}
	sort.Ints(missing)
	for _, n := range missing {
		t.Errorf("`docs/re/%s` 沒有登記在接線表——結論接進 remake 了嗎？"+
			"接了就寫「已接」＋引用點，沒接就寫「未接」＋缺什麼", notes[n])
	}

	// 二：表裡不能有幽靈列（筆記被刪或編號打錯）。
	for n, r := range listed {
		if _, ok := notes[n]; !ok {
			t.Errorf("接線表第 %d 行的編號 %02d 找不到對應筆記", r.line, n)
		}
	}

	for n, r := range listed {
		cited := cites[n]
		switch r.status {
		case "已接":
			// 三：說接了就要真的有人引用。
			if len(cited) == 0 {
				t.Errorf("編號 %02d 標「已接」，但程式碼、測試與工具裡沒有任何 `docs/re/%02d` 的引用",
					n, n)
			}
		case "未接", "不適用":
			// 四：理由不能空白。
			if r.note == "" {
				t.Errorf("編號 %02d 標「%s」卻沒寫理由", n, r.status)
			}
			// 五：反向——標成沒接卻被引用了，表示接上了忘了改狀態。
			if len(cited) > 0 {
				t.Errorf("編號 %02d 標「%s」，但 %s 引用了它——接上了就把狀態改成「已接」",
					n, r.status, strings.Join(cited, "、"))
			}
		default:
			t.Errorf("編號 %02d 的狀態是 %q，只能是「已接」「未接」「不適用」", n, r.status)
		}
	}

	var counts [3]int
	for _, r := range listed {
		switch r.status {
		case "已接":
			counts[0]++
		case "未接":
			counts[1]++
		case "不適用":
			counts[2]++
		}
	}
	t.Log(fmt.Sprintf("筆記 %d 份：已接 %d、未接 %d、不適用 %d",
		len(notes), counts[0], counts[1], counts[2]))
}

// TestPlaceholders 對帳「還在暫代的位置」：程式碼裡寫著「暫代」的檔案，
// 與接線表那一節列出來的檔案，必須是同一個集合。
//
// 兩個方向都要擋：新頂一個值卻沒登記，這張表就會漏掉；
// 解掉了卻沒清註解，表就會留著一筆假債。
// **靠人記得回來刪註解是靠不住的**——`weaponOf` 的「還沒接到物品表」
// 就在物品表接上之後又留了好幾輪。
func TestPlaceholders(t *testing.T) {
	root := repoRoot(t)

	inCode := map[string]bool{}
	for _, d := range []string{"internal", "cmd"} {
		err := filepath.Walk(filepath.Join(root, d), func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(p) != ".go" ||
				strings.HasSuffix(p, "_test.go") {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			if strings.Contains(string(b), "暫代") {
				rel, _ := filepath.Rel(root, p)
				inCode[filepath.ToSlash(rel)] = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	raw, err := os.ReadFile(filepath.Join(root, statusDoc))
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	i := strings.Index(body, "## 還在暫代的位置")
	if i < 0 {
		t.Fatal("接線表沒有「還在暫代的位置」那一節")
	}
	section := body[i:]
	if j := strings.Index(section[1:], "\n## "); j >= 0 {
		section = section[:j+1]
	}
	inDoc := map[string]bool{}
	for _, m := range regexp.MustCompile("`([^`]+\\.go)`").FindAllStringSubmatch(section, -1) {
		inDoc[m[1]] = true
	}
	if len(inDoc) == 0 {
		t.Fatal("「還在暫代的位置」一列都解不出來——正對照失敗")
	}

	for p := range inCode {
		if !inDoc[p] {
			t.Errorf("%s 裡有「暫代」，但接線表的「還在暫代的位置」沒列它——"+
				"頂了一個值就要登記缺哪一段 RE", p)
		}
	}
	for p := range inDoc {
		if !inCode[p] {
			t.Errorf("接線表說 %s 還在暫代，但那個檔案裡已經沒有「暫代」了——"+
				"解掉了就把這一列刪掉", p)
		}
	}
	t.Logf("暫代位置 %d 處", len(inDoc))
}
