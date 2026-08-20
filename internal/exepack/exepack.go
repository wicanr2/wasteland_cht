// Package exepack 把 Microsoft EXEPACK 打包的 16-bit DOS 執行檔還原，
// 並把 `wla.bin` 疊成執行時期的合成映像。
//
// 這是 `tools/unpack_exepack.py` 與 `tools/apply_overlay.py` 的 Go 版，
// **兩邊必須產生一模一樣的位元組**（`exepack_test.go` 拿雜湊對）。
// 會有 Go 版是因為公開包要能在沒有 Python 的機器上自己產生合成映像
// ——Windows 預設就沒有 Python，而少了合成映像遊戲一步都走不動
// （字型、九張字串表與資源定址常數都在執行檔裡，`docs/re/02`、`docs/re/03` §5）。
//
// 實作照格式本身，不依賴任何外部解包工具：每一步都能對原始 bytes 交代，
// 失敗時明確報錯，不產出一份看起來像樣的錯映像。
package exepack

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// errorMessage 是 stub 裡那句話，relocation 表緊接在它後面。
//
// ⚠ **後面沒有結尾的 NUL。** `wl.exe` 在 IDA 裡顯示成
// `Packed file is corrupt#`，那個 `#`（0x23）其實是第一組 relocation 的
// count（35 筆）低位元組。把它當成訊息的一部分往後跳，整張表會錯開一個 byte，
// 讀出來的 count 變成不合理的大數（`docs/re/02`）。
var errorMessage = []byte("Packed file is corrupt")

// MZ 是 DOS 執行檔的檔頭（只留這一支用得到的欄位）。
type MZ struct {
	CBLP, CP, CRLC, CPARHDR uint16
	SS, SP, IP, CS          uint16
	ImageSize, HeaderSize   int
}

// Header 是 EXEPACK 的 stub 檔頭（`RB` 簽章前面那 16 bytes）。
type Header struct {
	RealIP, RealCS   uint16
	MemStart         uint16
	ExepackSize      uint16
	RealSP, RealSS   uint16
	DestLenParagraph uint16
	SkipLenParagraph uint16
}

// Stats 是解包過程的計數，給報告用（與 Python 版同名同義）。
type Stats struct {
	Commands, Fills, Copies int
	TrailingFFPadding       int
	LiteralPrefixBytes      int
	PackedBytes             int
	UnpackedBytes           int
}

// ReadMZ 讀 MZ 檔頭。
func ReadMZ(data []byte) (MZ, error) {
	if len(data) < 0x1C {
		return MZ{}, fmt.Errorf("檔案太短，放不下 MZ 檔頭")
	}
	if !bytes.HasPrefix(data, []byte("MZ")) && !bytes.HasPrefix(data, []byte("ZM")) {
		return MZ{}, fmt.Errorf("不是 MZ 執行檔")
	}
	u := func(off int) uint16 { return binary.LittleEndian.Uint16(data[off:]) }
	m := MZ{
		CBLP: u(2), CP: u(4), CRLC: u(6), CPARHDR: u(8),
		SS: u(0x0E), SP: u(0x10), IP: u(0x14), CS: u(0x16),
	}
	last := int(m.CBLP)
	if last == 0 {
		last = 512
	}
	m.ImageSize = (int(m.CP)-1)*512 + last
	m.HeaderSize = int(m.CPARHDR) * 16
	if m.HeaderSize > len(data) || m.ImageSize > len(data) {
		return MZ{}, fmt.Errorf("MZ 檔頭寫的長度超出檔案本身")
	}
	return m, nil
}

// ParseHeader 讀 stub 開頭的 EXEPACK 檔頭。
func ParseHeader(stub []byte) (Header, error) {
	if len(stub) < 18 {
		return Header{}, fmt.Errorf("stub 太短")
	}
	if !bytes.Equal(stub[16:18], []byte("RB")) {
		return Header{}, fmt.Errorf("stub 段落沒有 EXEPACK 的 'RB' 簽章")
	}
	u := func(off int) uint16 { return binary.LittleEndian.Uint16(stub[off:]) }
	return Header{
		RealIP: u(0), RealCS: u(2), MemStart: u(4), ExepackSize: u(6),
		RealSP: u(8), RealSS: u(10),
		DestLenParagraph: u(12), SkipLenParagraph: u(14),
	}, nil
}

// unpackImage 反向解 RLE。
//
// ⚠ **EXEPACK 從映像尾端往前寫**，所以來源與目的的索引都是遞減的。
// 照一般「從頭往後」的直覺寫會得到一份長度正確、內容全錯的映像。
func unpackImage(packed []byte, destLen int) ([]byte, Stats, error) {
	src := len(packed)
	for src > 0 && packed[src-1] == 0xFF { // 尾端填充
		src--
	}
	st := Stats{TrailingFFPadding: len(packed) - src, PackedBytes: len(packed)}

	out := make([]byte, destLen)
	dst := destLen
	for {
		if src < 3 {
			return nil, st, fmt.Errorf("壓縮資料在讀到結束旗標前就用完了")
		}
		command := packed[src-1]
		length := int(binary.LittleEndian.Uint16(packed[src-3:]))
		src -= 3

		switch {
		case command&0xFE == 0xB0: // fill
			if src < 1 {
				return nil, st, fmt.Errorf("fill 命令缺少填充值")
			}
			value := packed[src-1]
			src--
			if dst-length < 0 {
				return nil, st, fmt.Errorf("fill 超出目的緩衝區")
			}
			for i := dst - length; i < dst; i++ {
				out[i] = value
			}
			dst -= length
			st.Fills++
		case command&0xFE == 0xB2: // copy
			if src-length < 0 || dst-length < 0 {
				return nil, st, fmt.Errorf("copy 超出緩衝區")
			}
			copy(out[dst-length:dst], packed[src-length:src])
			src -= length
			dst -= length
			st.Copies++
		default:
			return nil, st, fmt.Errorf("未知的 EXEPACK 命令 0x%02X", command)
		}

		st.Commands++
		if command&0x01 != 0 { // 最後一個命令
			break
		}
	}

	// 剩下的前段是未經 RLE 的原始資料，直接搬過去。
	if src > dst {
		return nil, st, fmt.Errorf("剩餘來源資料放不進目的緩衝區")
	}
	copy(out[dst-src:dst], packed[:src])
	st.LiteralPrefixBytes = src
	st.UnpackedBytes = len(out)
	return out, st, nil
}

// reloc 是一筆 relocation（段、位移）。
type reloc struct{ segment, offset uint16 }

// parseRelocations 讀 stub 尾端那張 16 組的 relocation 表。
func parseRelocations(stub []byte) ([]reloc, error) {
	idx := bytes.Index(stub, errorMessage)
	if idx < 0 {
		return nil, fmt.Errorf("stub 裡找不到 %q，無法定位 reloc table", errorMessage)
	}
	pos := idx + len(errorMessage)
	var out []reloc
	for high := 0; high < 16; high++ {
		if pos+2 > len(stub) {
			return nil, fmt.Errorf("reloc table 在讀完 16 組之前就超出 stub")
		}
		count := int(binary.LittleEndian.Uint16(stub[pos:]))
		pos += 2
		for i := 0; i < count; i++ {
			if pos+2 > len(stub) {
				return nil, fmt.Errorf("reloc table 讀到 stub 之外")
			}
			out = append(out, reloc{
				segment: uint16(high) << 12,
				offset:  binary.LittleEndian.Uint16(stub[pos:]),
			})
			pos += 2
		}
	}
	// 表之後應該只剩零。有非零資料表示解析起點錯了——
	// **這道檢查就是上面那個 `#` 陷阱的守門員**。
	for _, b := range stub[pos:] {
		if b != 0 {
			return nil, fmt.Errorf("reloc table 之後還有非零資料（stub offset 0x%X），解析起點可能錯了", pos)
		}
	}
	return out, nil
}

// buildEXE 用解出來的映像與 relocation 表組一份未壓縮的 MZ 執行檔。
func buildEXE(image []byte, h Header, relocs []reloc) []byte {
	relocBytes := make([]byte, 0, len(relocs)*4)
	for _, r := range relocs {
		relocBytes = binary.LittleEndian.AppendUint16(relocBytes, r.offset)
		relocBytes = binary.LittleEndian.AppendUint16(relocBytes, r.segment)
	}
	headerLen := 0x1C + len(relocBytes)
	headerParagraphs := (headerLen + 15) / 16
	headerSize := headerParagraphs * 16
	total := headerSize + len(image)

	header := make([]byte, headerSize)
	put := func(off int, v uint16) { binary.LittleEndian.PutUint16(header[off:], v) }
	put(0x00, 0x5A4D)                      // MZ
	put(0x02, uint16(total%512))           // e_cblp
	put(0x04, uint16((total+511)/512))     // e_cp
	put(0x06, uint16(len(relocs)))         // e_crlc
	put(0x08, uint16(headerParagraphs))    // e_cparhdr
	put(0x0A, 0x0000)                      // e_minalloc
	put(0x0C, 0xFFFF)                      // e_maxalloc
	put(0x0E, h.RealSS)                    // e_ss
	put(0x10, h.RealSP)                    // e_sp
	put(0x12, 0x0000)                      // e_csum
	put(0x14, h.RealIP)                    // e_ip
	put(0x16, h.RealCS)                    // e_cs
	put(0x18, 0x001C)                      // e_lfarlc
	put(0x1A, 0x0000)                      // e_ovno
	copy(header[0x1C:], relocBytes)
	return append(header, image...)
}

// Unpack 把打包的 `wl.exe` 還原成未壓縮的 MZ 執行檔。
func Unpack(data []byte) ([]byte, Stats, error) {
	mz, err := ReadMZ(data)
	if err != nil {
		return nil, Stats{}, err
	}
	body := data[mz.HeaderSize:mz.ImageSize]
	stubOffset := int(mz.CS) * 16
	if stubOffset >= len(body) {
		return nil, Stats{}, fmt.Errorf("e_cs 指到映像之外")
	}
	packed, stub := body[:stubOffset], body[stubOffset:]

	h, err := ParseHeader(stub)
	if err != nil {
		return nil, Stats{}, err
	}
	if int(h.ExepackSize) != len(stub) {
		return nil, Stats{}, fmt.Errorf("exepack_size %d 與 stub 實際長度 %d 不符",
			h.ExepackSize, len(stub))
	}
	image, st, err := unpackImage(packed, int(h.DestLenParagraph)*16)
	if err != nil {
		return nil, st, err
	}
	relocs, err := parseRelocations(stub)
	if err != nil {
		return nil, st, err
	}
	return buildEXE(image, h, relocs), st, nil
}

// ApplyOverlay 把 `wla.bin` 疊到解包映像的 `CS:0000`，產生合成映像。
//
// 為什麼要這一步：`start` 開機時把 `wla.bin` 讀進 `CS:0000`（`docs/re/03` §5），
// 之後 `call` 過去。直接看 `wl.unpacked.exe` 的那 8 KB 是**還沒被覆蓋的**內容，
// 分析它等於分析一段永遠不會執行的程式碼。
//
// 回傳的映像是本專案合成的產物，不是原版檔案。
func ApplyOverlay(base, overlay []byte) ([]byte, error) {
	if !bytes.HasPrefix(base, []byte("MZ")) {
		return nil, fmt.Errorf("基底不是 MZ 執行檔")
	}
	if len(base) < 0x0A {
		return nil, fmt.Errorf("基底太短")
	}
	headerSize := int(binary.LittleEndian.Uint16(base[0x08:])) * 16
	if headerSize+len(overlay) > len(base) {
		return nil, fmt.Errorf("overlay 超出映像範圍")
	}
	out := make([]byte, len(base))
	copy(out, base)
	copy(out[headerSize:headerSize+len(overlay)], overlay)
	return out, nil
}
