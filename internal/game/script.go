package game

// 地圖腳本直譯器（nibble 6，docs/spec/07 §5、docs/re/34）。
//
// 記錄本身就是指令：+0x00 是 opcode、+0x01/+0x02 是「下一步」、
// +0x03 起是參數。每個指令回報要不要繼續。

// OpCount 是指令表的長度（ds:A4EAh，44 筆，第 45 個 word 是 0）。
const OpCount = 44

// 指令編號。只列有實作的，其餘走 default。
const (
	OpMatchPlace   = 0  // 比對目前地圖與座標
	OpBranch       = 1  // 依條件選 +0x03 或 +0x05 當下一步
	OpOverlay      = 2  // overlay 畫面呼叫，語意未解
	OpDayNight     = 3  // 晝夜分支
	OpStash        = 4  // 寄放角色並蓋時間戳
	OpFetchRecord  = 5  // 取 section 3 的另一筆
	OpCopyRecord   = 6  // 複製 section 5 的一筆到另一筆
	OpNeedItem     = 7  // 沒帶物品 0x2F 的人扣血
	OpPlace9       = 8  // 往 section 5 第 9 筆放東西
	OpPlace25      = 9  // 同上，第 25 筆
	OpStatusOne    = 10 // 對指定角色加狀態
	OpStatusFirst  = 11 // 對第一個角色加狀態
	OpStatusAll    = 12 // 對全隊加狀態
	OpNop          = 13
	OpCountdown    = 14 // 印倒數並停住
	OpMatchCoord   = 15 // 比對這一格的座標
	OpSwapGender   = 16 // 全隊變性
	OpStatusAll2   = 17 // 對全隊加狀態（迴圈上限用另一個欄位）
	OpKindsDec     = 18
	OpDenomAdd2    = 19
	OpSlotsDec     = 20
	OpDenomSub2    = 21
	OpSlotsInc     = 22
	OpKindsInc11   = 23
	OpKindsInc12   = 24
	OpKindsInc17   = 25
	OpWait         = 26 // jmp sub_1142B，語意未解
	OpDenomSet20   = 27
	OpDenomAdd10   = 28
	OpFillRange    = 29
	OpFillOne      = 30
	OpAbort        = 31 // stc; retn
	OpFillPair     = 32
	OpFillOnes     = 33
	OpPlace9Param  = 34
	OpStartTimer   = 35
	OpMatchNeigh   = 36 // 檢查相鄰三格
	OpPlace15      = 37
	OpFillFirst10  = 38
	OpElapsed      = 39 // 距離時間戳過了多久
	OpDenomSet0    = 40
	OpDenomSet30   = 41
	OpDenomSet100  = 42
	OpPartyHasItem = 43
)

// 記錄區標頭裡會被腳本改的三個欄位（docs/re/34 §3）。
const (
	hdrEncounterDenom = 0x2F
	hdrEncounterKinds = 0x31
	hdrEncounterSlots = 0x32
)

// ScriptResult 是跑完一個指令之後的狀態。
type ScriptResult struct {
	Op       int
	Continue bool // 對應原版的 CF ＝ 0
	Handled  bool // false 表示這個指令還沒實作
	Message  int  // 要顯示的字串編號，−1 表示沒有
}

// Script 是一次腳本執行的上下文。規則層不碰畫面，
// 要顯示什麼透過 ScriptResult 交出去。
type Script struct {
	World *World
	// Record 是目前這筆記錄（＝ 直譯器的 PC）。
	Record []byte
	// Op 是這一筆的指令編號。**不是 `Record[0]`**——`Record[0]` 是
	// section `0x10` 的索引，取出來的 word 才是 opcode
	// （`sub_12C80` 的 `sub_17CB1(bl ＝ 0x10)`，docs/re/75 §2）。
	// 用 NewScript 建立就會查好。
	Op int
}

// NewScript 從記錄 `+0x00` 查出 opcode 並建好上下文。
//
// 查不到（沒有 section 0x10 或索引超出）時 Op ＝ −1，`Step` 會回
// `Handled ＝ false`——**不要當成 nop**。
func NewScript(w *World, record []byte) *Script {
	s := &Script{World: w, Record: record, Op: -1}
	if w == nil || w.Block == nil || len(record) == 0 {
		return s
	}
	if v, err := w.Block.SectionEntry(0x10, int(record[0])); err == nil {
		s.Op = int(v)
	}
	return s
}

// Step 跑一個指令。回傳 Handled ＝ false 時，呼叫者要當成「還沒做」處理，
// 不要當成 nop——那會讓遊戲安靜地跑錯。
func (s *Script) Step() ScriptResult {
	if len(s.Record) < 3 {
		return ScriptResult{Op: -1, Message: -1}
	}
	op := s.Op
	res := ScriptResult{Op: op, Continue: true, Handled: true, Message: -1}
	if op < 0 || op >= OpCount {
		res.Handled = false
		res.Continue = false
		return res
	}
	hdr := s.World.Block.Header

	arg := func(i int) byte {
		if 3+i < len(s.Record) {
			return s.Record[3+i]
		}
		return 0
	}
	// 分支就是「把某兩個 byte 搬進 +0x01/+0x02」。
	branch := func(at int) {
		if at+1 < len(s.Record) {
			s.Record[1], s.Record[2] = s.Record[at], s.Record[at+1]
		}
	}
	clamp := func(off int, delta int, lo, hi int) {
		v := int(hdr[off]) + delta
		if v < lo || v > hi {
			return
		}
		hdr[off] = byte(v)
	}

	switch op {
	case OpNop:
		// 什麼都不做。

	case OpAbort:
		res.Continue = false

	case OpDayNight:
		// 6–17 時走 +0x03，其餘走 +0x05。
		if s.World.Clock.Hour >= DawnHour && s.World.Clock.Hour < DuskHour {
			branch(3)
		} else {
			branch(5)
		}

	case OpMatchCoord:
		if arg(4) == s.World.Party.X && arg(5) == s.World.Party.Y {
			branch(3)
		} else {
			branch(5)
		}

	case OpSwapGender:
		for _, c := range s.World.Party.Members {
			c.Gender ^= 1
		}

	case OpStatusFirst, OpStatusAll, OpStatusAll2, OpStatusOne:
		bit := statusMask(arg(0))
		switch op {
		case OpStatusFirst:
			if c := s.World.Party.Current(); c != nil {
				c.Status |= bit
			}
		case OpStatusOne:
			if c := s.World.Party.Current(); c != nil {
				c.Status |= bit
			}
		default:
			for _, c := range s.World.Party.Members {
				c.Status |= bit
			}
		}

	case OpKindsDec:
		clamp(hdrEncounterKinds, -1, 1, 255)
	case OpSlotsDec:
		clamp(hdrEncounterSlots, -1, 1, 255)
	case OpSlotsInc:
		clamp(hdrEncounterSlots, +1, 0, 0x15)
	case OpKindsInc11:
		clamp(hdrEncounterKinds, +1, 0, 0x0B)
	case OpKindsInc12:
		clamp(hdrEncounterKinds, +1, 0, 0x0C)
	case OpKindsInc17:
		clamp(hdrEncounterKinds, +1, 0, 0x11)
	case OpDenomAdd2:
		if hdr[hdrEncounterDenom] != 0 {
			clamp(hdrEncounterDenom, +2, 0, 255)
		}
	case OpDenomSub2:
		if hdr[hdrEncounterDenom] != 0 {
			clamp(hdrEncounterDenom, -2, 1, 255)
		}
	case OpDenomAdd10:
		clamp(hdrEncounterDenom, +10, 0, 0xC8)
	case OpDenomSet20:
		hdr[hdrEncounterDenom] = 0x14
	case OpDenomSet0:
		hdr[hdrEncounterDenom] = 0
	case OpDenomSet30:
		hdr[hdrEncounterDenom] = 0x1E
	case OpDenomSet100:
		hdr[hdrEncounterDenom] = 0x64

	case OpPartyHasItem:
		// 記錄 +0x07 起是一串物品編號，0xFF 結束。
		found := false
		for at := 7; at < len(s.Record) && s.Record[at] != 0xFF && !found; at++ {
			for _, c := range s.World.Party.Members {
				if c.HasItem(s.Record[at]) {
					found = true
					break
				}
			}
		}
		if found {
			branch(3)
		} else {
			branch(5)
		}

	default:
		// 還沒實作的指令：明確回報，不要假裝成 nop。
		res.Handled = false
	}
	return res
}

// statusMask 把腳本參數（0–7）換成狀態位元（ds:CF4Eh ＝ 01 02 04 … 80）。
func statusMask(index byte) uint8 {
	if index > 7 {
		return 0
	}
	return 1 << index
}

// HasItem 回傳角色身上有沒有這件物品（不消耗）。
func (c *Character) HasItem(id byte) bool {
	for _, it := range c.Items {
		if it.ID == id {
			return true
		}
	}
	return false
}
