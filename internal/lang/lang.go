// Package lang 讀翻譯目錄（docs/spec/11 §6–§7）。
//
// 目錄是 tools/build_lang.py 編出來的：**Big5 編碼與排版檢查都在編譯時做完了**，
// 所以這裡只要會讀就好，不依賴任何編碼函式庫。
//
// 載不到目錄時整個中文關掉、遊戲跑英文；查不到的條目回原文。
// **半成品的中文化要能玩**——不留空白，也不顯示 key。
package lang

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	magic   = "WLCAT\x00"
	version = 1
)

// Catalogue 是一份翻譯目錄。nil 的 Catalogue 可以安全使用（一律回 false）。
type Catalogue struct {
	entries map[string][]byte
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

	c := &Catalogue{entries: make(map[string][]byte, count)}
	for i := 0; i < count; i++ {
		key, next, err := readChunk(raw, at)
		if err != nil {
			return nil, fmt.Errorf("第 %d 條的 key：%w", i, err)
		}
		text, next2, err := readChunk(raw, next)
		if err != nil {
			return nil, fmt.Errorf("第 %d 條的內容：%w", i, err)
		}
		c.entries[string(key)] = text
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

// Lookup 查一條翻譯，回傳 Big5 bytes。查不到就回 false——
// 呼叫者要用原文，不要顯示空白。
func (c *Catalogue) Lookup(key string) ([]byte, bool) {
	if c == nil {
		return nil, false
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

// BlockKey 是地圖區塊字串的 key（docs/spec/11 §3）。
func BlockKey(file string, resourceID, slot int) string {
	return fmt.Sprintf("blk:%s:%d:%d", file, resourceID, slot)
}

// ExeKey 是執行檔字串表的 key。
func ExeKey(table, slot int) string {
	return fmt.Sprintf("exe:%d:%d", table, slot)
}
