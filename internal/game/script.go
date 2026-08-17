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
	OpOverlay      = 2  // overlay slot 18 ＝ 對調兩張圖形（docs/re/104）
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
	OpBeep         = 26 // jmp sub_1142B ＝ 播音效 7（docs/re/44 §6）
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
	// Sound 是要播的音效編號（0–8，`docs/re/44` §6），−1 表示沒有。
	// 規則層只回報編號，**要不要播由呈現層決定**。
	Sound int
	// Countdown 非 nil 時要印自毀倒數（opcode 14，`docs/re/102` §3）。
	// 同一條規矩：規則層只算數字，字串怎麼拼由呈現層決定。
	Countdown *Countdown
	// Swap 非 nil 時要對調兩張圖形（opcode 2，`docs/re/104`）。
	// 規則層只回報「哪兩個編號」，真的動圖是呈現層的事。
	Swap *IconSwap
}

// IconSwap 是 opcode 2 要求的圖形對調（`sub_10036` ＝ overlay slot 18）。
//
// 編號走的是**疊圖與圖磚連續的那一張表**（`seg003:0x420`，`docs/re/24` §2.3）：
// 0–9 是 `IC0_9.WLF` 的疊圖，≥10 是圖磚（圖磚編號 ＝ 值 − 10）。
type IconSwap struct{ A, B byte }

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

// FacilityKindCount 是 `ds:A4E0h` 跳表前面那幾個設施項的數量。
//
// 索引 `FacilityKindCount + n` 就是 **opcode n**（`ds:A4EAh` 只是
// `ds:A4E0h + 10`，`docs/re/79` §1）。
const FacilityKindCount = 5

// NewScript 從記錄 `+0x00` 查出 opcode 並建好上下文。
//
// 兩條路（`sub_12C80`）：
//
//	bit7 沒設 → `+0x00` 是 section 0x10 的索引，取出的 word 才是 opcode
//	bit7 設   → `+0x00 & 0x7F` **直接**索引 `ds:A4E0h`；≥ 5 就是 opcode −5
//
// 查不到時 Op ＝ −1，`Step` 會回 `Handled ＝ false`——**不要當成 nop**。
func NewScript(w *World, record []byte) *Script {
	s := &Script{World: w, Record: record, Op: -1}
	if w == nil || w.Block == nil || len(record) == 0 {
		return s
	}
	if record[0]&0x80 != 0 {
		if kind := int(record[0] & 0x7F); kind >= FacilityKindCount {
			s.Op = kind - FacilityKindCount
		}
		return s // kind < 5 是設施，不走直譯器
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
		return ScriptResult{Op: -1, Message: -1, Sound: -1}
	}
	op := s.Op
	res := ScriptResult{Op: op, Continue: true, Handled: true, Message: -1, Sound: -1}
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

	case OpMatchPlace:
		// **有沒有另一支隊伍站在記錄 +0x07 起那串座標上**（`0x1A470`）。
		// 有 → +0x03，沒有 → +0x05。
		if s.otherGroupOnList() {
			branch(3)
		} else {
			branch(5)
		}

	case OpBranch:
		// 第一個隊員站著 → +0x03，倒下 → +0x05（`0x1A4F4` 的 `sub_172BB(1)`；
		// 索引是 1 起算的，所以那是 `Members[0]`）。
		up := len(s.World.Party.Members) > 0 &&
			s.World.Party.Members[0] != nil && !s.World.Party.Members[0].Down()
		if up {
			branch(3)
		} else {
			branch(5)
		}

	case OpOverlay:
		// **把兩張圖形對調**（`0x1A515` → overlay slot 18，`docs/re/104`）。
		// 出貨資料六筆都是「隊伍的疊圖（7）↔ 一張地形圖磚」——隊伍在地圖上
		// 變成那塊地形；參數對調再跑一次就換回來。
		res.Swap = &IconSwap{A: arg(0), B: arg(1)}

	case OpStash:
		// 把第一個隊員抄一份存進第 `+0x03` 格並蓋時間戳（`0x1A54B`）。
		s.World.stashFirstMember(arg(0))

	case OpNeedItem:
		// 全隊掃物品 0x2F（`0x1A699`）：**沒有**的人 CON ← −5（舊值進 +0x26）。
		// 只要有一個人帶著就走 +0x05，一個都沒有才走 +0x03。
		anyHas := false
		for _, c := range s.World.Party.Members {
			if c == nil {
				continue
			}
			if c.HasItem(needItemID) {
				anyHas = true
				continue
			}
			c.PreHurt = c.CON
			c.CON = needItemCON
		}
		if anyHas {
			branch(5)
		} else {
			branch(3)
		}

	case OpPlace9, OpPlace9Param:
		// section 5 的第 9 筆 ← (0x5E, +0x03, +0x04)。
		// **op 8 與 op 34 是同一段程式碼的兩份**（`0x1A6F5`／`0x1AABF`）。
		if !s.placeInSection5(9, arg(0), arg(1)) {
			res.Handled = false
		}

	case OpPlace25:
		// 同上，第 25 筆，而且第二個參數是**常數 0**（`0x1A765`）。
		if !s.placeInSection5(25, arg(0), 0) {
			res.Handled = false
		}

	case OpPlace15:
		if !s.placeInSection5(15, arg(0), arg(1)) {
			res.Handled = false
		}

	case OpCountdown:
		// 印自毀倒數然後**停住**（`0x1A7E8` 收尾是 `stc`）。
		cd := countdownOf(s.World.Clock.Total - s.World.SelfDestruct.At)
		res.Countdown = &cd
		res.Continue = false

	case OpFillPair:
		// section 3 的第 0x1D 筆**往回到第 0 筆**，每筆寫兩個欄位
		// （`0x1AA3F`）：`+0x03 ← +0x03`、`+0x09 ← +0x04`。
		a, b := arg(0), arg(1)
		for i := 0x1D; i >= 0; i-- {
			r, err := s.World.Block.SectionRecord(3, i)
			if err != nil || len(r) < 10 {
				continue
			}
			r[3], r[9] = a, b
		}

	case OpMatchNeigh:
		// 三格都是 (nibble 4, 記錄 2) → 下一步直接設成 (0x0C, 3)（`0x1AB35`）。
		// ⚠ 這一個**不走 branch**：它把字面值寫進 +0x01/+0x02。
		if s.neighboursMatch() && len(s.Record) > 2 {
			s.Record[1], s.Record[2] = 0x0C, 3
		}

	case OpElapsed:
		// 寄放的角色過了多久（`0x1AC03`）。三條路，而且**目標不是常見的
		// +0x03/+0x05**：時候未到 → +0x06、隊伍滿了 → +0x08、領回來 → +0x04。
		switch {
		case s.World.stashElapsed(arg(0)) < stashReturnTicks:
			branch(6)
		case len(s.World.Party.Members) >= stashMaxParty:
			branch(8)
		default:
			s.World.unstash(arg(0))
			branch(4)
		}

	case OpFillRange, OpFillOne, OpFillOnes, OpFillFirst10:
		// section 3 的一段記錄，每筆 `+0x09` 換成一個值（docs/re/34 §2）：
		//
		//	op 29 → 第 0x10–0x21 筆 ← 記錄 +0x03
		//	op 30 → 第 0x22 筆      ← 記錄 +0x03
		//	op 33 → 第 0x21–0x28 筆 ← **常數 1**
		//	op 38 → 第 0–9 筆       ← 記錄 +0x03
		lo, hi, val := 0, 0, arg(0)
		switch op {
		case OpFillRange:
			lo, hi = 0x10, 0x21
		case OpFillOne:
			lo, hi = 0x22, 0x22
		case OpFillOnes:
			lo, hi, val = 0x21, 0x28, 1
		case OpFillFirst10:
			lo, hi = 0, 9
		}
		for i := lo; i <= hi; i++ {
			r, err := s.World.Block.SectionRecord(3, i)
			if err != nil || len(r) < 10 {
				continue
			}
			r[9] = val
		}

	case OpCopyRecord:
		// 把 section 5 的第 `+0x03` 筆整筆複製到第 `+0x04` 筆
		//（逐 byte，遇 `0x5E`／`0xDE`／`0xFF` 收尾，docs/re/34 §2）。
		src, err1 := s.World.Block.SectionRecord(5, int(arg(0)))
		dst, err2 := s.World.Block.SectionRecord(5, int(arg(1)))
		if err1 != nil || err2 != nil {
			res.Handled = false
			break
		}
		for i := 0; i < len(src) && i < len(dst); i++ {
			b := src[i]
			dst[i] = b
			if b == 0x5E || b == 0xDE || b == 0xFF {
				break
			}
		}

	case OpBeep:
		// `jmp sub_1142B` ＝ `sub_1CBD3(al ＝ 7)`：高低雙音 ×6
		// （`docs/re/44` §6 的音效 7）。出貨資料裡 30 格指到它，
		// 是**有格子指到的未實作 opcode 裡最多的一個**。
		res.Sound = 7

	case OpStartTimer:
		// **啟動科奇斯基地的自毀**（`0x1AB0E`）：記下現在的時鐘、
		// 把遭遇分母歸零，再播兩次音效 7。240 刻之後主迴圈的
		// `sub_1CB30` 會合成一個 kind 4 的分派——那就是結局
		//（`docs/re/100`）。全 42 張地圖只有一筆記錄用這個 opcode。
		s.World.ArmSelfDestruct()
		res.Sound = 7 // 原版連播兩次（`call sub_1142B; jmp sub_1142B`）

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
		// **44 種 opcode 現在全部有實作**（`docs/re/102`、`docs/re/104`），
		// 所以走到這裡只有一種可能：`Op` 是 section `0x10` 查出來的越界值
		// （1282／2271／26478／29813，`docs/re/76` §3）——
		// 那代表「這一筆記錄根本不是腳本」。明確回報，不要假裝成 nop。
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
//
// 編號 0 是空槽的值，不是一件物品——固定 30 槽的陣列裡幾乎一定有空槽，
// 不擋掉的話「身上有沒有第 0 件」永遠回 true。
func (c *Character) HasItem(id byte) bool {
	if id == 0 {
		return false
	}
	for _, it := range c.Items {
		if it.ID == id {
			return true
		}
	}
	return false
}
