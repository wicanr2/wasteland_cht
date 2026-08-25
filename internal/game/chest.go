package game

// 寶箱（nibble 5）：內容生成在 `docs/re/29` §4、撿拾流程在 `docs/re/130`。
//
// 記錄佈局：`+0x00`／`+0x01` 是「全部拿完之後這一格變什麼」的改寫對
// （`0x15436` 的 `sub_169B1(0)`）；`+0x02` 起每兩個 byte 一組：
//
//	0x00    空格（拿走的東西清成 0），**續掃不是結束**（`0x15299` → `0x1532A`）
//	0xFF    清單結束（`0x1529E`）
//	< 0x80  物品類別，還沒擲出是哪一件
//	≥ 0x80  已決定（bit7 設），低 7 位是物品編號
//	0x5E    現金特例：占 3 bytes（0x5E、金額低、金額高），
//	        第一次踩到把兩個金額 byte 各 roll 一次並把 0x5E 改成 0xDE
//
// 第二個 byte 是數量：bit7 設 ＝ 還沒擲，`roll(低 7 位)` 後寫回。

// chestCash 是現金項的物品編號（`0x153A3` 的 `cmp al, 5Eh`）。
const chestCash = 0x5E

// chestPairs 逐組走訪寶箱清單，回傳每一組的位移（跳過空格、遇 0xFF 停）。
func chestPairs(data []byte) []int {
	var out []int
	for at := 2; at+1 < len(data); at += 2 {
		switch {
		case data[at] == 0xFF:
			return out
		case data[at] == 0:
			continue
		default:
			out = append(out, at)
		}
	}
	return out
}

// RollChest 把「還沒決定」的項目擲成具體內容並**寫回記錄**
// （`docs/re/29` §4：lazy 生成，同一個寶箱重看不會變）。
// data 是 `Block.SectionRecord` 回傳的活切片，寫回就是進地圖資料。
func (w *World) RollChest(tbl ItemTable, data []byte) {
	for _, at := range chestPairs(data) {
		b := data[at]
		if b&0x80 != 0 {
			continue // 已決定
		}
		if b == chestCash {
			// 現金：改成 0xDE，後面兩個金額 byte 各擲一次（`sub_15441` ×2）。
			data[at] = chestCash | 0x80
			for k := 1; k <= 2; k++ {
				if at+k < len(data) && data[at+k] > 0 {
					data[at+k] = byte(w.RNG.Roll(int(data[at+k])))
				}
			}
			continue
		}
		// 類別 → 擲出這個類別的第幾件（`sub_15453` ＋ `roll(件數)`）。
		if id, ok := w.rollChestItem(tbl, b); ok {
			data[at] = id | 0x80
		} else {
			data[at] = 0 // 這個類別一件都沒有：清掉（原版 ZF 那條路）
		}
		if at+1 < len(data) && data[at+1]&0x80 != 0 {
			n := int(data[at+1] & 0x7F)
			if n > 0 {
				data[at+1] = byte(w.RNG.Roll(n))
			} else {
				data[at+1] = 1
			}
		}
	}
}

// rollChestItem 從類別擲出具體物品：先數這個類別有幾件（物品編號從 94 往下
// 掃到 1，比對物品類別 ＝ +0x03 >> 3，`docs/re/45`；出貨資料的待擲類別
// 只有 1 ＝ 近戰），再 roll(件數) 取第 n 件（`docs/re/29` §4）。
func (w *World) rollChestItem(tbl ItemTable, category byte) (byte, bool) {
	count := 0
	for id := byte(chestCash); id >= 1; id-- {
		if d, ok := tbl.Get(id); ok && d.Class == ItemClass(category) {
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	want := w.RNG.Roll(count)
	seen := 0
	for id := byte(chestCash); id >= 1; id-- {
		if d, ok := tbl.Get(id); ok && d.Class == ItemClass(category) {
			seen++
			if seen == want {
				return id, true
			}
		}
	}
	return 0, false
}

// ChestEntry 是寶箱清單上的一項。
type ChestEntry struct {
	At    int  // 記錄裡那一組的位移
	ID    byte // 物品編號（已去 bit7）；chestCash ＝ 現金
	Count int  // 還剩幾件；現金時 ＝ 金額（16-bit）
}

// ChestEntries 列出還沒被拿走的項目（給呈現層畫清單）。
func ChestEntries(data []byte) []ChestEntry {
	var out []ChestEntry
	for _, at := range chestPairs(data) {
		id := data[at] & 0x7F
		e := ChestEntry{At: at, ID: id}
		if id == chestCash {
			e.Count = int(data[at+1])
			if at+2 < len(data) {
				e.Count |= int(data[at+2]) << 8
			}
		} else if at+1 < len(data) {
			e.Count = int(data[at+1])
		}
		out = append(out, e)
	}
	return out
}

// TakeChestEntry 讓 c 拿一件（`0x153CB`）。回傳 false ＝ 物品槽滿
// （原版印字串 18 `You can't carry any more.`）。
// 拿完的項目把第一個 byte 清成 0；現金整筆一次拿走並在第三個 byte
// 蓋 0xFF（`0x153B9`——讓步進 2 的掃描在那裡結束，原版就是這樣收的）。
func (w *World) TakeChestEntry(tbl ItemTable, data []byte, at int, c *Character) bool {
	if at < 2 || at >= len(data) || data[at] == 0 || data[at] == 0xFF {
		return true // 沒東西可拿，當作成功（呼叫端重畫清單就會消失）
	}
	id := data[at] & 0x7F
	if id == chestCash {
		amount := 0
		if at+1 < len(data) {
			amount = int(data[at+1])
		}
		if at+2 < len(data) {
			amount |= int(data[at+2]) << 8
			data[at+2] = 0xFF
		}
		c.Money += uint32(amount)
		data[at] = 0
		return true
	}
	slot, ok := FirstEmptyItemSlot(c.Items)
	if !ok {
		return false
	}
	var full byte
	if d, ok := tbl.Get(id); ok {
		full = d.Capacity
	}
	c.Items = putSlot(c.Items, slot, Slot{ID: id, Value: full})
	if at+1 < len(data) && data[at+1] > 0 {
		data[at+1]--
	}
	if at+1 >= len(data) || data[at+1] == 0 {
		data[at] = 0
	}
	return true
}

// ChestEmptied 回報清單是不是空了——空了要用位移 0 的改寫對把這一格
// 改掉（`0x15436` 的 `sub_169B1(0)`，布袋圖示就是這樣消失的）。
func (w *World) ChestEmptied(x, y int, data []byte) {
	if len(ChestEntries(data)) > 0 {
		return
	}
	w.applyCellPatch(x, y, data, 0)
}
