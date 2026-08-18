package play

// 長名字的側車檔（`WLNAMES`）。
//
// ⚠ **原版的名字欄是 13 bytes**（角色記錄 `+0x00`–`+0x0C`，`docs/re/15`），
// 而且**那一格不能加長**：`+0x0E` 起就是七個屬性，往後接著金錢、性別、
// 國籍、AC、MaxCON、CON。名字擴到 30 bytes 會一路蓋到 `+0x1D`，存檔整個壞掉。
//
// 所以做法是：**原版存檔照舊寫截斷過的 13 bytes**（原版遊戲讀得懂、
// round-trip 不變），完整名字另外存一個重製版自己的檔。
// 先例是快速存檔的 `WLQS`（規格 27）——**不碰玩家的原版資料**是同一條原則。
//
// 側車檔不在就退回存檔裡那 13 bytes，遊戲照跑。

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"
)

const (
	namesFile    = "WLNAMES"
	namesMagic   = "WLNM"
	namesVersion = 1
	// NameSlots 是角色記錄的筆數（存檔的 0x800 段是 8 × 256，第 0 筆是全域）。
	NameSlots = 7
)

// longNames 是每個角色槽的完整名字（UTF-8）。空字串 ＝ 那一槽沒有長名字，
// 用存檔裡那 13 bytes 就好。
type longNames [NameSlots]string

// namesPath 是側車檔的位置——**與存檔同一個目錄**。
func namesPath(dir string) string { return filepath.Join(dir, namesFile) }

// loadLongNames 讀側車檔。檔案不在、壞掉、版本不合都回空的——
// **不要因此讓遊戲開不起來**，退回存檔裡的短名字就好。
func loadLongNames(dir string) longNames {
	var out longNames
	if dir == "" {
		return out
	}
	raw, err := os.ReadFile(namesPath(dir))
	if err != nil || len(raw) < 6 || string(raw[:4]) != namesMagic {
		return out
	}
	if binary.LittleEndian.Uint16(raw[4:]) != namesVersion {
		return out
	}
	at := 6
	for i := 0; i < NameSlots; i++ {
		if at+2 > len(raw) {
			break
		}
		n := int(binary.LittleEndian.Uint16(raw[at:]))
		at += 2
		if at+n > len(raw) {
			break
		}
		// ⚠ **檔案可能被改壞**：不是合法 UTF-8 就丟掉那一槽，
		// 不要把半個字塞進畫面（`render.DrawRune` 查不到字模會整格消失）。
		if s := string(raw[at : at+n]); utf8.ValidString(s) {
			out[i] = s
		}
		at += n
	}
	return out
}

// storeLongNames 寫側車檔。
//
// **每一槽都寫**（沒有長名字就寫空字串），這樣槽號與位置一一對應，
// 讀的時候不必猜哪一槽對哪一個角色。
func storeLongNames(dir string, names longNames) error {
	if dir == "" {
		return nil
	}
	buf := append([]byte(namesMagic), 0, 0)
	binary.LittleEndian.PutUint16(buf[4:], namesVersion)
	for _, n := range names {
		buf = binary.LittleEndian.AppendUint16(buf, uint16(len(n)))
		buf = append(buf, n...)
	}
	if err := os.WriteFile(namesPath(dir), buf, 0o644); err != nil {
		return fmt.Errorf("寫 %s：%w", namesFile, err)
	}
	return nil
}
