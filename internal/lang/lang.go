// Package lang 讀翻譯目錄（docs/spec/11 §6–§7）。
//
// 目錄是 tools/build_lang.py 編出來的，**內容是 UTF-8**：排版檢查與
// 「倚天畫不畫得出來」的 Big5 覆蓋檢查都在編譯期做完，執行期只要會讀。
//
// ⚠ **Big5 只出現在畫的那一刻**（`internal/render` 查字模時），
// 中間所有層都是 UTF-8。以前整條管線走 Big5 bytes，代價是任何拼接
// 都得先轉——漏一次就是「讀得出筆畫卻不成字」，而且只在畫面上看得到。
//
// 載不到目錄時整個中文關掉、遊戲跑英文；查不到的條目回原文。
// **半成品的中文化要能玩**——不留空白，也不顯示 key。
package lang

import (
	"encoding/binary"
	"fmt"
	"strings"
	"os"
)

const (
	magic   = "WLCAT\x00"
	version = 2
)

// Catalogue 是一份翻譯目錄。nil 的 Catalogue 可以安全使用（一律回 false）。
type Catalogue struct {
	entries map[string]string
}

// Load 讀一份 .cat。檔案不存在時回 (nil, err)，呼叫者要當成「沒有翻譯」處理。
func Load(path string) (*Catalogue, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) < len(magic)+6 || string(raw[:len(magic)]) != magic {
		return nil, fmt.Errorf("%s 不是翻譯目錄", path)
	}
	at := len(magic)
	if v := binary.LittleEndian.Uint16(raw[at:]); v != version {
		return nil, fmt.Errorf("%s 的版本是 %d，這支只讀 %d", path, v, version)
	}
	at += 2
	count := int(binary.LittleEndian.Uint32(raw[at:]))
	at += 4

	c := &Catalogue{entries: make(map[string]string, count)}
	for i := 0; i < count; i++ {
		key, next, err := readChunk(raw, at)
		if err != nil {
			return nil, fmt.Errorf("第 %d 條的 key：%w", i, err)
		}
		text, next2, err := readChunk(raw, next)
		if err != nil {
			return nil, fmt.Errorf("第 %d 條的內容：%w", i, err)
		}
		c.entries[string(key)] = string(text)
		at = next2
	}
	return c, nil
}

func readChunk(raw []byte, at int) ([]byte, int, error) {
	if at+2 > len(raw) {
		return nil, 0, fmt.Errorf("檔案在 %#x 就結束了", at)
	}
	n := int(binary.LittleEndian.Uint16(raw[at:]))
	at += 2
	if at+n > len(raw) {
		return nil, 0, fmt.Errorf("長度 %d 超出檔案", n)
	}
	return raw[at : at+n], at + n, nil
}

// Lookup 查一條翻譯，回傳 **UTF-8** 文字。查不到就回 false——
// 呼叫者要用原文，不要顯示空白。
func (c *Catalogue) Lookup(key string) (string, bool) {
	if c == nil {
		return "", false
	}
	v, ok := c.entries[key]
	return v, ok
}

// Len 是目錄裡有幾條。
func (c *Catalogue) Len() int {
	if c == nil {
		return 0
	}
	return len(c.entries)
}

// Each 逐條走過目錄。給**盤點與門檻測試**用——遊戲本體只查單一 key。
func (c *Catalogue) Each(f func(key, value string)) {
	if c == nil {
		return
	}
	for k, v := range c.entries {
		f(k, v)
	}
}

// BlockKey 是地圖區塊字串的 key（docs/spec/11 §3）。
func BlockKey(file string, resourceID, slot int) string {
	return fmt.Sprintf("blk:%s:%d:%d", file, resourceID, slot)
}

// ExeKey 是執行檔字串表的 key。
func ExeKey(table, slot int) string {
	return fmt.Sprintf("exe:%d:%d", table, slot)
}

// ParagraphKey 是段落手札正文的 key（`translations/paragraphs-zh-Hant.cat`）。
//
// 段落書不在原版的字串語料裡（它是紙本），所以自成一個目錄檔，
// 但格式與 key 的形狀刻意沿用同一套，讀檔器只要一份。
func ParagraphKey(n int) string {
	return fmt.Sprintf("para:%d", n)
}

// UIKey 是重製版自己的介面文字（`translations/*/ui.tsv`）。
//
// 原版把指令列與幾個選單寫成 ASCII 字面值、不走字串表，所以抽不出 key；
// 但它們玩家看得到，一樣要進語言資料檔（`CLAUDE.md` §3.3）。
func UIKey(name string) string { return "ui:" + name }

// PlaceKey 是地點招牌的 key。
//
// 地點名是**資料裡的明文 ASCII**（地圖記錄與存檔），不在字串表裡，
// 所以 key 就是英文原名本身——查表用的是畫面上原本要印的那串字。
func PlaceKey(name string) string { return "place:" + name }

// MonsterKey 是明文敵人名字的 key（`docs/re/114` §6）。
//
// ⚠ **控制碼要寫成 `\xNN`**，與 `tools/extract_monster_names.py` 的 `escape()`
// 同一套：名字用 `\n` 分單複數（`Juv\nenile\nies\n`），而翻譯檔是逐行的 TSV，
// key 裡塞真的換行就沒辦法逐行讀。`tools/build_lang.py` 的 `read_tsv`
// **只把譯文 unescape，key 原樣保留**，所以這一側也要用跳脫過的形式。
// 兩邊不一致的症狀是**每一條有單複數的名字都查不到**——而畫面上只是顯示英文。
func MonsterKey(raw string) string {
	var b strings.Builder
	b.WriteString("monster:")
	for i := 0; i < len(raw); i++ {
		if c := raw[i]; c < 0x20 || c == 0x7F {
			fmt.Fprintf(&b, "\\x%02X", c)
			continue
		}
		b.WriteByte(raw[i])
	}
	return b.String()
}
