// 指令 wl-atlas 把 42 個地圖區塊裡「玩家走得到的東西」倒成一份 JSON。
//
// 存在的理由是攻略（`docs/walkthrough/`）要有可稽核的來源：攻略正文寫的
// 每一句「這裡有商店」「這扇門要開鎖」「密語是 MUERTE」，都要指得回這份
// 機器產出的清單，而不是靠人記憶或抄別人的攻略。
//
//	tools/go.sh run ./cmd/wl-atlas -dir workplace/orig/wastland \
//	    -lang translations/zh-Hant.cat -out docs/walkthrough/generated/atlas.json
//
// 這支**只讀不寫遊戲資料**，也不加語意推論：nibble 與記錄 bytes 照原樣倒出來，
// 該怎麼讀寫在 `docs/spec/07`。整理成 markdown 是 `tools/summarize_walkthrough.py`
// 的事（`CLAUDE.md` §1.2 的兩段式）。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

func main() {
	dir := flag.String("dir", "workplace/orig/wastland", "原版資料目錄")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	catPath := flag.String("lang", "", "翻譯目錄（.cat），給了就一併倒出中文")
	out := flag.String("out", "", "輸出的 JSON 路徑（必填）")
	flag.Parse()

	if err := run(*dir, *imagePath, *catPath, *out); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}

// Atlas 是整份輸出。
type Atlas struct {
	Dir   string     `json:"dir"`
	Maps  []MapInfo  `json:"maps"`
	Notes []string   `json:"notes"`
	Cells CellCensus `json:"cell_census"`
}

// CellCensus 是全 42 張地圖的 nibble 統計，用來確認沒有漏掉哪一類格子。
type CellCensus struct {
	ByNibble map[string]int `json:"by_nibble"`
}

// MapInfo 是一張地圖。
type MapInfo struct {
	Resource   int          `json:"resource"`
	File       string       `json:"file"`
	Dim        int          `json:"dim"`
	Tileset    int          `json:"tileset"`
	EncDenom   int          `json:"encounter_denominator"`
	StepMin    float64      `json:"step_minutes"`
	Strings    []StringInfo `json:"strings"`
	Facilities []Facility   `json:"facilities"`
	Teleports  []Teleport   `json:"teleports"`
	Gates      []GateInfo   `json:"gates"`
	Menus      []MenuInfo   `json:"menus"`
	Chests     []CellUse    `json:"chests"`
	Messages   []MessageUse `json:"messages"`
}

// StringInfo 是這張地圖的一條字串。
type StringInfo struct {
	Slot int    `json:"slot"`
	Key  string `json:"key"`
	EN   string `json:"en"`
	ZH   string `json:"zh,omitempty"`
}

// CellUse 記「哪些格子指到這一筆記錄」。
//
// Bytes 只留開頭 recBytes 個——`SectionRecord` 回傳的是「從記錄起點到區塊尾」，
// 記錄邊界沒有長度欄位（`docs/re/16` §3），整段倒出來會是幾十 KB 的噪音。
type CellUse struct {
	Nibble    int      `json:"nibble"`
	Record    int      `json:"record"`
	CellCount int      `json:"cell_count"`
	Cells     [][2]int `json:"cells"`
	Bytes     string   `json:"bytes"`
}

// recBytes 是每筆記錄倒出的 byte 數。
const recBytes = 32

// Facility 是 nibble 6 的一筆：設施或腳本。
type Facility struct {
	CellUse
	JumpIndex int    `json:"jump_index"`
	Kind      string `json:"kind"`           // "facility"、"script" 或 "script-indexed"
	Opcode    int    `json:"opcode"`         // 腳本才有，−1 表示不是腳本
	Name      string `json:"name,omitempty"` // 設施才有：招牌上的字
}

// Teleport 是 nibble 10 的一筆。
type Teleport struct {
	CellUse
	ToMap     int  `json:"to_map"`
	ToRes     int  `json:"to_resource"` // 地圖編號查 ds:BF1Ch 之後的資源編號
	ToX       int  `json:"to_x"`
	ToY       int  `json:"to_y"`
	Relative  bool `json:"relative"`
	Back      bool `json:"back"`
	AsksFirst bool `json:"asks_first"`
}

// GateInfo 是 nibble 2 的一筆：移動閘與它的條件串列。
type GateInfo struct {
	CellUse
	MessageSlot int    `json:"message_slot"`
	Conditions  []Cond `json:"conditions"`
}

// Cond 是條件串列的一條。
type Cond struct {
	Kind  string `json:"kind"`
	ID    int    `json:"id"`
	Value int    `json:"value"`
	Raw   string `json:"raw"`
}

// MenuInfo 是 nibble 8 的一筆：問答。
type MenuInfo struct {
	CellUse
	// Valid ＝ 這一筆真的是問答（見 menuInfo 的三道篩）。false 的不要拿去寫攻略。
	Valid      bool     `json:"valid"`
	Mode       string   `json:"mode"` // "單鍵" 或 "打字"
	PromptSlot int      `json:"prompt_slot"`
	PromptEN   string   `json:"prompt_en"`
	PromptZH   string   `json:"prompt_zh,omitempty"`
	AnswerSlot []int    `json:"answer_slots"`
	Answers    []string `json:"answers"`
}

// MessageUse 是 nibble 1／4／9／12 的一筆：踩上去印字。
type MessageUse struct {
	CellUse
	Slots []int `json:"slots"`
}

func run(dir, imagePath, catPath, out string) error {
	if out == "" {
		return fmt.Errorf("要指定 -out")
	}
	rom, err := assets.Open(dir)
	if err != nil {
		return err
	}
	if err := rom.LoadImage(imagePath); err != nil {
		return err
	}
	var cat *lang.Catalogue
	if catPath != "" {
		if cat, err = lang.Load(catPath); err != nil {
			return err
		}
	}

	res, err := rom.Resources()
	if err != nil {
		return err
	}
	atlas := Atlas{
		Dir:   dir,
		Cells: CellCensus{ByNibble: map[string]int{}},
		Notes: []string{
			"nibble 語意見 docs/spec/07 §2；記錄 bytes 未加解讀，照原樣倒出",
			"傳送目的地的 to_map 是**地圖編號**，bit7 設起來的要再查 ds:BF1Ch（docs/re/61）",
			"條件串列的解讀依 docs/re/65；kind 未知時只給 raw",
		},
	}

	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			return fmt.Errorf("區塊 %d：%w", i, err)
		}
		mi, err := collect(rom, b, cat, atlas.Cells.ByNibble)
		if err != nil {
			return fmt.Errorf("區塊 %d：%w", i, err)
		}
		atlas.Maps = append(atlas.Maps, mi)
	}

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	buf, err := json.MarshalIndent(atlas, "", "  ")
	if err != nil {
		return err
	}
	buf = append(buf, '\n')
	if err := os.WriteFile(out, buf, 0o644); err != nil {
		return err
	}
	fmt.Printf("%d 張地圖 → %s（%d bytes）\n", len(atlas.Maps), out, len(buf))
	return nil
}

func collect(rom *assets.Rom, b *assets.Block, cat *lang.Catalogue, census map[string]int) (MapInfo, error) {
	mi := MapInfo{
		Resource: b.Resource.ID,
		File:     b.Resource.File,
		Dim:      b.Dim,
		Tileset:  b.Tileset,
		EncDenom: int(b.Header[0x2F]),
		StepMin:  b.StepMinutes(),
	}
	for slot, s := range b.Strings {
		si := StringInfo{Slot: slot, Key: lang.BlockKey(b.Resource.File, b.Resource.ID, slot), EN: s}
		si.ZH = zhString(b, cat, slot)
		mi.Strings = append(mi.Strings, si)
	}

	// 先掃全圖，把「哪些 (nibble, 記錄) 有格子指到」與座標收起來。
	type key struct{ n, r int }
	used := map[key][][2]int{}
	for y := 0; y < b.Dim; y++ {
		for x := 0; x < b.Dim; x++ {
			terrain, record, _, err := b.At(x, y)
			if err != nil {
				return mi, err
			}
			census[fmt.Sprintf("%d", terrain)]++
			k := key{int(terrain), int(record)}
			used[k] = append(used[k], [2]int{x, y})
		}
	}
	// **沒有格子指到的記錄也要列**：Base Cochise 的傳送格在出貨資料裡
	// 不存在，是劇情把某一格改寫成 nibble 10 之後才出現的（`docs/re/71`）。
	// 只列「現在踩得到的」會漏掉整段後期地圖。
	// ⚠ nibble 5 不能這樣列：section 邊界靠「第一個非空指標」推得
	// （`docs/re/16` §3.2），沒有格子佐證時讀出來的是雜訊。
	// nibble 8 列了，但每一筆要過 MenuInfo.Valid 那道篩——資源 0 沒有
	// 真正的問答，硬讀會得到「一句地圖敘述配一串不成話的答案」。
	for _, typ := range []int{2, 6, 8, 10} {
		for r := 0; r < b.SectionCount(typ); r++ {
			k := key{typ, r}
			if _, ok := used[k]; !ok {
				used[k] = nil
			}
		}
	}
	keys := make([]key, 0, len(used))
	for k := range used {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].n != keys[j].n {
			return keys[i].n < keys[j].n
		}
		return keys[i].r < keys[j].r
	})

	for _, k := range keys {
		cells := used[k]
		rec, _ := boundedRecord(b, k.n, k.r)
		cu := CellUse{Nibble: k.n, Record: k.r, CellCount: len(cells), Cells: cells, Bytes: hex(rec)}
		switch k.n {
		case 1, 4, 9, 12:
			mu := MessageUse{CellUse: cu}
			switch k.n {
			case 1:
				for i := 0; i < len(rec); i++ {
					if s := rec[i] & 0x7F; s != 0 {
						mu.Slots = append(mu.Slots, int(s))
					}
					if rec[i]&0x80 != 0 {
						break
					}
				}
			default:
				if len(rec) > 0 && rec[0] != 0 {
					mu.Slots = []int{int(rec[0])}
				}
			}
			mi.Messages = append(mi.Messages, mu)
		case 2:
			gi := GateInfo{CellUse: cu}
			if len(rec) > 1 {
				gi.MessageSlot = int(rec[1])
			}
			for _, g := range game.ParseGates(rec) {
				gi.Conditions = append(gi.Conditions, cond(g))
			}
			mi.Gates = append(mi.Gates, gi)
		case 5:
			mi.Chests = append(mi.Chests, cu)
		case 6:
			mi.Facilities = append(mi.Facilities, facility(b, cu, rec))
		case 8:
			mi.Menus = append(mi.Menus, menuInfo(b, cu, rec, cat))
		case 10:
			t := Teleport{CellUse: cu, ToMap: -1, ToRes: -1}
			if len(rec) >= 4 {
				t.Relative = rec[0]&0x80 != 0
				t.AsksFirst = rec[0]&0x40 != 0
				t.ToX, t.ToY = int(rec[1]), int(rec[2])
				if rec[3] == game.TeleportBackMarker {
					t.Back = true
				} else {
					t.ToMap = int(rec[3])
					if v, err := rom.ResolveMapID(rec[3]); err == nil {
						t.ToRes = int(v)
					}
				}
			}
			mi.Teleports = append(mi.Teleports, t)
		}
	}
	return mi, nil
}

// cond 把一條閘門條件翻成可讀的欄位。
//
// 型別 1／5／6／7 的判定路徑都是「找物品」（`internal/game/gates.go`），
// 為什麼分四個編號還沒解——這裡照原版保留編號，不合併。
func cond(g game.Gate) Cond {
	c := Cond{ID: int(g.Param), Value: g.Difficulty, Raw: fmt.Sprintf("type=%d diff=%d param=%02x", g.Type, g.Difficulty, g.Param)}
	switch g.Type {
	case game.GateSkill:
		c.Kind = "技能"
	case game.GateAttribute:
		c.Kind = "屬性"
	case game.GatePartySize:
		c.Kind = "隊伍人數"
	case game.GateMoney:
		c.Kind = "金錢"
	case 1, 5, 6, 7:
		c.Kind = "物品"
	default:
		c.Kind = "未解"
	}
	return c
}

// typableAnswer 回報這條字串**打得出來嗎**：原版的輸入緩衝區是 16 bytes，
// 輸入層又丟掉 `< 0x20` 的字元（`docs/re/46` §4），所以超過 15 個字元或含
// 控制碼的字串永遠比不中。
func typableAnswer(s string) bool {
	if strings.TrimSpace(s) == "" || len(s) > 15 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return false
		}
	}
	return true
}

// menuInfo 解一格 nibble 8。三道篩決定 Valid：
//
//	1. 答案清單讀得到結束標記（`ParseQuestion` 回 false 就是沒讀到）
//	2. 題目的字串編號在這個區塊的字串數之內
//	3. 打字題至少有一條答案打得出來——一條都打不出來的不是問答
//
// 沒有格子指到的記錄照樣列（對話會把格子改寫成下一題），但雜訊要擋掉。
func menuInfo(b *assets.Block, cu CellUse, rec []byte, cat *lang.Catalogue) MenuInfo {
	m := MenuInfo{CellUse: cu}
	q, ok := game.ParseQuestion(rec)
	if !ok {
		return m
	}
	m.Mode = "打字"
	if q.SingleKey {
		m.Mode = "單鍵"
	}
	m.Valid = int(q.Prompt) < len(b.Strings)
	if !q.SingleKey {
		typable := false
		for _, a := range q.Answers {
			if typableAnswer(blockString(b, int(a))) {
				typable = true
				break
			}
		}
		m.Valid = m.Valid && typable
	}
	m.PromptSlot = int(q.Prompt)
	m.PromptEN = blockString(b, int(q.Prompt))
	m.PromptZH = zhString(b, cat, int(q.Prompt))
	for _, a := range q.Answers {
		m.AnswerSlot = append(m.AnswerSlot, int(a))
		m.Answers = append(m.Answers, blockString(b, int(a)))
	}
	return m
}

// boundedRecord 取一筆記錄，並且**把尾巴切在下一筆記錄的起點**。
//
// `Block.SectionRecord` 回傳「從記錄起點到區塊尾」——記錄本身沒有長度欄位
// （`docs/re/16` §3）。條件串列讀到 `0xFF` 就停，所以平常沒事；但**沒有終止碼
// 的記錄會一路吃進下一筆**，症狀是「這一格的條件跟隔壁那一格一模一樣」，
// 看起來完全正常。用同一個 section 裡下一個更大的指標當上界擋掉。
func boundedRecord(b *assets.Block, typ, index int) ([]byte, error) {
	rec, err := b.SectionRecord(typ, index)
	if err != nil {
		return nil, err
	}
	start := len(b.Raw) - len(rec)
	end := len(b.Raw)
	for i := 0; i < b.SectionCount(typ); i++ {
		off, err := b.SectionEntry(typ, i)
		if err != nil {
			break
		}
		if int(off) > start && int(off) < end {
			end = int(off)
		}
	}
	return b.Raw[start:end], nil
}

// facility 依 `NewScript`／`ParseFacility` 的兩條路分類（`docs/re/79` §1）：
//
//	bit7 沒設 → `+0x00` 是 section 0x10 的索引，取出來的 word 才是 opcode
//	bit7 設   → `+0x00 & 0x7F` < 5 是設施，≥ 5 是 opcode −5
//
// ⚠ **不能只看 `+0x00 & 0x7F`**：沙漠上那幾百格高溫腳本的 `+0x00` 是 1–3，
// 照設施表讀會變成「沙漠上有 439 格醫生」。
func facility(b *assets.Block, cu CellUse, rec []byte) Facility {
	f := Facility{CellUse: cu, JumpIndex: -1, Opcode: -1}
	if len(rec) == 0 {
		return f
	}
	if rec[0]&0x80 == 0 {
		f.Kind = "script-indexed"
		if v, err := b.SectionEntry(0x10, int(rec[0])); err == nil {
			f.Opcode = int(v)
			f.JumpIndex = int(v) + game.FacilityKindCount
		}
		return f
	}
	kind := int(rec[0] & 0x7F)
	f.JumpIndex = kind
	if kind >= game.FacilityKindCount {
		f.Kind = "script"
		f.Opcode = kind - game.FacilityKindCount
		return f
	}
	f.Kind = "facility"
	if fac, ok := game.ParseFacility(rec); ok {
		f.Name = fac.Name
	}
	return f
}

func blockString(b *assets.Block, slot int) string {
	if slot < 0 || slot >= len(b.Strings) {
		return ""
	}
	return b.Strings[slot]
}

func zhString(b *assets.Block, cat *lang.Catalogue, slot int) string {
	if cat == nil || slot < 0 || slot >= len(b.Strings) {
		return ""
	}
	v, ok := cat.Lookup(lang.BlockKey(b.Resource.File, b.Resource.ID, slot))
	if !ok {
		return ""
	}
	// 目錄裡是 Big5（畫面直接吃 Big5 查倚天字模），JSON 要 UTF-8——
	// **不轉的話 encoding/json 會把每個不合法的 byte 換成 U+FFFD**，
	// 症狀是輸出看起來有中文、實際上內容已經毀了。
	out, ok := lang.FromBig5(textlayout.RenderBytes(v, textlayout.Options{}))
	if !ok {
		return ""
	}
	return out
}

func hex(b []byte) string {
	const digits = "0123456789abcdef"
	if len(b) > recBytes {
		b = b[:recBytes]
	}
	out := make([]byte, 0, len(b)*3)
	for i, v := range b {
		if i > 0 {
			out = append(out, ' ')
		}
		out = append(out, digits[v>>4], digits[v&0xF])
	}
	return string(out)
}
