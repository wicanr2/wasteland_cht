package play

// 商店的庫存表（`docs/re/118`）。
//
// 物品表在資料檔裡有**四份**：`game1` 的存檔資源 `+0x1206` 起三份
// （位移 0／`0x2FE`／`0x5FC`），`game2` 一份。四份的價錢、傷害、類別完全相同，
// **只有庫存那一欄（`+0x02`）不一樣**——它是「這一組商店賣什麼」。
//
// 進商店時原版把設施記錄的 `+0x06` 寫進 `ds:46C4h`，`sub_185E6` 比對
// `ds:46C5h`（目前載進來的那一組），不同才重新讀。**離開商店不會換回去**，
// 所以它是一個活著的全域狀態，跟著存檔一起被寫回去。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// itemStockGroups 是設施記錄 `+0x06` 認得的四個值。
//
// ⚠ 出貨資料實際用到的是 0（交易車廂、農業站雜貨店）、1（高池鎮商店）、
// 2（石英城商場）、4（維加斯的市場與黑市）——**3 沒有人用**。
// 醫生的 `+0x06` 是別的東西（每點治療費），不要拿去當庫存組。
const itemStockGroups = 5

// itemStockFile 回報第 n 組庫存住在哪個檔案（`0x18611` 的 `cmp … 4`）。
func itemStockFile(n byte) string {
	if n == 4 {
		return "game2"
	}
	return "game1"
}

// itemStockSlot 回報第 n 組在那個檔案裡的第幾份（`ds:BE20h` 的位移表）。
func itemStockSlot(n byte) int {
	if n == 4 {
		return 0 // game2 只有一份
	}
	return int(n)
}

// loadItemStock 換一組庫存表（`sub_185E6`）。
//
// 已經是這一組就什麼都不做——原版也是先比 `ds:46C5h` 再決定要不要讀。
// **這一步不能省**：省了的話賣東西的 `+1` 會寫進上一家店的那一份。
func (s *Scene) loadItemStock(n byte) error {
	if n >= itemStockGroups {
		return fmt.Errorf("庫存組 %d 不在 0–%d", n, itemStockGroups-1)
	}
	if n == s.itemStock && len(s.itemsRaw) > 0 {
		return nil
	}
	if s.rom == nil {
		return nil
	}
	raw, err := s.rom.LoadItemTable(itemStockFile(n), itemStockSlot(n))
	if err != nil {
		return err
	}
	s.items, s.itemsRaw = game.ParseItemTable(raw), raw
	s.itemStock = n
	return nil
}
