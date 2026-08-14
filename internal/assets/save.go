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
//	+0x806 0xA00 未加密段（出廠全零，用途未解）
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
	Tail   []byte // 0xA00 未加密段，原樣保留
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
	TeleX    byte
	TeleY    byte
	Unknown  byte // +0x0D，語意未解
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
		g.TeleX, g.TeleY, g.Unknown = raw[11], raw[12], raw[13]
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
