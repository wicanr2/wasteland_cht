package assets

import (
	"encoding/binary"
	"fmt"
)

// 存檔容器（docs/spec/05 §2、docs/re/30）。
//
// 存檔是 GAME1／GAME2 檔尾的一個 MSQ 資源：
//
//	+0x00  magic 'msq0'／'msq1'
//	+0x04  checksum（16-bit）＝ 0 − Σ 明文位元組
//	+0x06  0x800 加密段 ＝ 8 × 256 bytes
//	+0x806 0xA00 未加密段 ＝ **十組按鍵巨集**（10 × 256，docs/re/30 §6）
//
// ⚠ 存檔策略是**改寫不是重建**：Plain 保留整份明文，寫回時就地修改，
// 未解區域一個 byte 都不動（CLAUDE.md §4）。

const (
	savePlainLen  = 0x800 // 加密段
	saveTailLen   = 0xA00 // 未加密段
	saveRecordLen = 256
	saveRecords   = savePlainLen / saveRecordLen // 8：第 0 筆是全域狀態
)

// SaveOffset 是兩個資料檔裡存檔資源的位置（原版 sub_18744 的 seek 目標，cx:dx）。
var SaveOffset = map[string]int{"game1": 0x000253C5, "game2": 0x00028BC7}

// 全域狀態（第 0 筆）的欄位位移（docs/spec/05 §3）。
const (
	gblSlotGroup0 = 0x00 // 四組隊伍槽表，間隔 14
	gblSlotStride = 0x0E
	gblSlotGroups = 4
	gblVars       = 0x78 // ds:464Eh–465Bh 那 14 bytes
	gblPlace      = 0xD0 // 地點名稱（明文）＝ 記憶體的 ds:7201h
	gblPlaceLen   = 13
	gblSerial     = 0xF5 // 32-bit 存檔序號
	gblFontColor  = 0xFE
)

// Save 是一份解開的存檔。
type Save struct {
	File   string
	Offset int
	Magic  string
	Plain  []byte // 0x800 明文，含全域狀態與 7 筆角色記錄
	Tail   []byte // 0xA00 未加密段 ＝ 十組按鍵巨集
}

// 按鍵巨集在尾段裡的佈局（docs/re/43 §6）。
const (
	MacroCount  = 10
	MacroStride = 0x100
)

// Macro 取第 n 組按鍵巨集（0–9）的原始 bytes。
//
// ⚠ 巨集**不進 checksum**——改它不會讓存檔被拒收。但重製版照樣要
// round-trip 它（`CLAUDE.md` §4：未解的位元組要能原樣寫回，
// 何況這一段已經不是未解的了）。
func (s *Save) Macro(n int) ([]byte, bool) {
	if n < 0 || n >= MacroCount || len(s.Tail) < (n+1)*MacroStride {
		return nil, false
	}
	return s.Tail[n*MacroStride : (n+1)*MacroStride], true
}

// LoadSave 讀出指定資料檔尾端的存檔並驗 checksum。
func (r *Rom) LoadSave(file string) (*Save, error) {
	at, ok := SaveOffset[file]
	if !ok {
		return nil, fmt.Errorf("%s 沒有已知的存檔位移", file)
	}
	data, err := r.File(file)
	if err != nil {
		return nil, err
	}
	if at+6+savePlainLen+saveTailLen > len(data) {
		return nil, fmt.Errorf("%s 太短，放不下 %#x 的存檔", file, at)
	}
	if magic := string(data[at : at+3]); magic != "msq" {
		return nil, fmt.Errorf("%s 的 %#x 不是 msq 而是 %q", file, at, magic)
	}
	checksum := le16(data, at+4)

	plain := decryptSave(data[at+6:at+6+savePlainLen], checksum)
	if got := saveChecksum(plain); got != checksum {
		return nil, fmt.Errorf("%s 的存檔 checksum 對不上：檔案 %#04x、算出來 %#04x",
			file, checksum, got)
	}

	tail := make([]byte, saveTailLen)
	copy(tail, data[at+6+savePlainLen:])

	return &Save{
		File:   file,
		Offset: at,
		Magic:  string(data[at : at+4]),
		Plain:  plain,
		Tail:   tail,
	}, nil
}

// Bytes 把存檔重新編碼成檔案裡的樣子（magic ＋ checksum ＋ 密文 ＋ 尾段）。
// checksum 一定重算——改了內容不重算，原版會拒收。
func (s *Save) Bytes() []byte {
	checksum := saveChecksum(s.Plain)
	out := make([]byte, 0, 6+savePlainLen+saveTailLen)
	out = append(out, s.Magic...)
	out = binary.LittleEndian.AppendUint16(out, checksum)
	out = append(out, encryptSave(s.Plain, checksum)...)
	out = append(out, s.Tail...)
	return out
}

// saveChecksum ＝ 0 − Σ 明文位元組（16-bit，docs/re/30 §2.1）。
func saveChecksum(plain []byte) uint16 {
	var sum uint16
	for _, b := range plain {
		sum -= uint16(b)
	}
	return sum
}

// XOR 串流是對稱的，但分成兩支命名，讓呼叫端讀得出方向。
func decryptSave(raw []byte, checksum uint16) []byte { return xorSave(raw, checksum) }
func encryptSave(raw []byte, checksum uint16) []byte { return xorSave(raw, checksum) }

func xorSave(raw []byte, checksum uint16) []byte {
	out := make([]byte, len(raw))
	key := byte(checksum) ^ byte(checksum>>8)
	for i, b := range raw {
		out[i] = b ^ key
		key += 0x1F
	}
	return out
}

// Record 回傳第 n 筆 256 bytes 的切片（0 ＝ 全域狀態，1–7 ＝ 角色）。
// 回傳的是 Plain 的切片，改它就是改存檔本體。
func (s *Save) Record(n int) ([]byte, error) {
	if n < 0 || n >= saveRecords {
		return nil, fmt.Errorf("記錄編號 %d 超出 0–%d", n, saveRecords-1)
	}
	return s.Plain[n*saveRecordLen : (n+1)*saveRecordLen], nil
}

// Serial 是 32-bit 存檔序號；兩份輪替時比它，大的比較新。
func (s *Save) Serial() uint32 { return le32(s.Plain, gblSerial) }

// Place 是地點名稱（明文 ASCII，NUL 結尾）。
func (s *Save) Place() string {
	raw := s.Plain[gblPlace : gblPlace+gblPlaceLen]
	for i, b := range raw {
		if b == 0 {
			return string(raw[:i])
		}
	}
	return string(raw)
}

// SlotGroup 是一組隊伍槽表（14 bytes，docs/spec/05 §3.1）。
type SlotGroup struct {
	Members  [8]byte // 角色記錄編號，0 ＝ 空槽
	X, Y     byte
	MapID    byte
	// +0x0B–+0x0D 是**回程**的座標與地圖，不是傳送目的地——
	// 踩上傳送格的第一件事就是把目前位置寫進去（docs/re/60 §3）。
	BackX    byte
	BackY    byte
	BackMap  byte
	RawIndex int
}

// SlotGroups 解出四組隊伍槽表。
func (s *Save) SlotGroups() [gblSlotGroups]SlotGroup {
	var out [gblSlotGroups]SlotGroup
	for i := range out {
		base := gblSlotGroup0 + i*gblSlotStride
		raw := s.Plain[base : base+gblSlotStride]
		g := SlotGroup{RawIndex: base}
		copy(g.Members[:], raw[0:8])
		g.X, g.Y, g.MapID = raw[8], raw[9], raw[10]
		g.BackX, g.BackY, g.BackMap = raw[11], raw[12], raw[13]
		out[i] = g
	}
	return out
}

// Globals 回傳 ds:464Eh–465Bh 那 14 bytes 的切片（可就地改）。
func (s *Save) Globals() []byte { return s.Plain[gblVars : gblVars+14] }

// PickNewer 在兩份存檔裡挑序號大的那一份（原版 sub_18744 的規則）。
func PickNewer(a, b *Save) *Save {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	case b.Serial() > a.Serial():
		return b
	default:
		return a
	}
}

// 物品資料表（docs/re/45 §2）。
//
// ⚠ **它不在執行檔裡，在存檔區**：映像裡 `ds:7A31h` 那段是零，表是開檔時
// 從「存檔資源 ＋ 0x1206 ＋ 槽位移」讀進來的（`0x186A9`）。每個存檔槽各一份，
// 所以它是**可變的遊戲狀態**——賣東西改到的庫存欄位要跟著存回去。
const (
	itemTableDelta = 0x1206 // 存檔資源 → 物品表資源
	itemTableLen   = 0x2F8  // 760 ＝ 95 × 8
)

// ItemSlotOffsets 是 `ds:BE20h` 那張表：三個存檔槽各自的物品表位移。
// 第 4 筆與第 3 筆相同（＝ 檔尾），所以實際只有三個槽。
var ItemSlotOffsets = [3]int{0x0000, 0x02FE, 0x05FC}

// LoadItemTable 讀出某個存檔槽的物品資料表（解密後的 760 bytes）。
// 用原版自己的條件驗 checksum，不符就回錯誤——不要為了讓流程過而放寬。
func (r *Rom) LoadItemTable(file string, slot int) ([]byte, error) {
	base, ok := SaveOffset[file]
	if !ok {
		return nil, fmt.Errorf("%s 沒有已知的存檔位移", file)
	}
	if slot < 0 || slot >= len(ItemSlotOffsets) {
		return nil, fmt.Errorf("存檔槽 %d 不在 0–%d", slot, len(ItemSlotOffsets)-1)
	}
	data, err := r.File(file)
	if err != nil {
		return nil, err
	}
	at := base + itemTableDelta + ItemSlotOffsets[slot]
	if at+6+itemTableLen > len(data) {
		return nil, fmt.Errorf("%s 太短，放不下 %#x 的物品表", file, at)
	}
	if magic := string(data[at : at+3]); magic != "msq" {
		return nil, fmt.Errorf("%s 的 %#x 不是 msq 而是 %q", file, at, magic)
	}
	checksum := le16(data, at+4)
	plain := decryptSave(data[at+6:at+6+itemTableLen], checksum)
	if got := saveChecksum(plain); got != checksum {
		return nil, fmt.Errorf("%s 槽 %d 的物品表 checksum 對不上：檔案 %#04x、算出來 %#04x",
			file, slot, checksum, got)
	}
	return plain, nil
}
