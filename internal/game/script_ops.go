package game

// 腳本 opcode 需要的世界狀態（`docs/re/34`、`docs/re/102`）。
//
// 這一批 opcode **出貨資料裡一格都沒有指到**（`docs/re/76`）——它們是
// 「走不到的程式碼路徑」。實作它們的理由不是玩家會踩到，而是
// **直譯器要能完整地跑**：`Handled ＝ false` 的每一個分支都是一個
// 「照跑但什麼都沒發生」的洞，而洞的數量只有靠實作才會減少。
//
// ⚠ **只照讀出來的寫，不補故事。** 讀不動的（op 2 的 overlay 呼叫）
// 就留著 `Handled ＝ false`，不要猜一個看起來合理的行為填進去。

// GroupPos 是一支隊伍在世界上的位置（存檔的隊伍槽表 `+0x08`／`+0x09`／`+0x0A`，
// `docs/re/60` §3）。腳本 opcode 0 要拿它比對座標。
type GroupPos struct {
	X, Y   byte
	MapID  byte
	Active bool // 這一組有人
}

// PartyGroups 是隊伍組數上限（存檔的四組槽表）。
const PartyGroups = 4

// StashSlot 是 opcode 4 寄放的一個角色。
//
// 原版把整個 256 byte 的記錄抄到 `ds:065Ch ＋ n×256`、時間戳寫進
// `ds:0A5Ch ＋ n×4`（`0x1A54B`）。抄完之後**在副本上**清掉物品陣列
// （`+0xBD`–`+0xF8`）與裝備／護甲／護甲值／金錢——原本那個隊員不受影響。
type StashSlot struct {
	Char *Character
	At   uint32 // 寄放當下的 Clock.Total
}

// stashReturnTicks 是 opcode 39 的門檻（`0x1AC3F` 的 `sbb al, 2Dh`，
// 比的是 32-bit 時間差的**第二個 byte**）。
const stashReturnTicks = 0x2D << 8

// stashMaxParty 是 opcode 39 肯把人放回隊伍的人數上限
// （`0x1AC51` 的 `cmp al, 7`：`≥ 7` 就走另一條）。
const stashMaxParty = 7

// needItemID 是 opcode 7 掃的那件物品（`0x1A6A4` 的 `mov al, 2Fh`）。
const needItemID = 0x2F

// needItemCON 是沒帶那件物品的人被設成的 CON（`0x1A6BD` 的 `0FFFBh`）。
const needItemCON = -5

// Countdown 是 opcode 14 要顯示的倒數（`0x1A7E8`）。
//
// 原版印的是「字串 `0x11` ＋ 分 ＋ `:` ＋ 秒（補 0 兩位）＋ `.` ＋ 百分秒
// ＋ 字串 `0x12`」，**印完 `stc` 停住腳本**。
//
// ⚠ 規則層只算數字，字串怎麼拼由呈現層決定——與音效同一條規矩。
type Countdown struct {
	Minutes int
	Seconds int
	// Hundredths 來自 `ds:CF4Ah` 那張四格表（`00 0F 1E 2D` ＝ 0／15／30／45），
	// 索引是剩餘刻數的低 2 位。
	Hundredths int
	// Head／Tail 是包在數字兩邊的字串編號（`0x11`／`0x12`）。
	Head, Tail int
}

// countdownFraction 是 `ds:CF4Ah` 的四個 byte。
var countdownFraction = [4]int{0, 15, 30, 45}

// 倒數訊息前後的兩個字串編號（`0x1A835`／`0x1A866`）。
const (
	countdownHeadString = 0x11
	countdownTailString = 0x12
)

// countdownOf 把「已經過了幾刻」換成畫面上的 mm:ss.ff。
//
// 原版的算法：`剩餘 ＝ 0xF0 − 經過`，低 2 位查表當百分秒，
// 其餘右移兩位得到秒數，再除以 60 拆成分與秒（`0x1A803`–`0x1A833`）。
func countdownOf(elapsed uint32) Countdown {
	remain := int(SelfDestructTicks) - int(elapsed)
	if remain < 0 {
		remain = 0
	}
	total := remain >> 2
	return Countdown{
		Minutes:    total / 60,
		Seconds:    total % 60,
		Hundredths: countdownFraction[remain&3],
		Head:       countdownHeadString,
		Tail:       countdownTailString,
	}
}

// placeInSection5 是 opcode 8／9／34／37 共用的那一段（`0x1A6F5` 等四處）：
// 往 section 5 的第 n 筆寫 `+0x02 ← 0x5E`、`+0x03 ← a`、`+0x04 ← b`。
//
// `0x5E` 是記錄的終止碼之一（見 `OpCopyRecord` 的收尾條件），
// 寫在 `+0x02` 等於**把那一筆截斷成只剩兩個參數**。
func (s *Script) placeInSection5(index int, a, b byte) bool {
	rec, err := s.World.Block.SectionRecord(5, index)
	if err != nil || len(rec) < 5 {
		return false
	}
	rec[2], rec[3], rec[4] = 0x5E, a, b
	return true
}

// otherGroupOnList 是 opcode 0 的判斷（`0x1A470`）：
// **有沒有另一支隊伍站在記錄 `+0x07` 起那串座標的某一格上**。
//
// 清單是 (x, y) 成對排到 `0xFF` 為止。只有「不是自己這一組」且
// 「在同一張地圖」的組參與比對。
//
// ⚠ 原版的迴圈上界是 `ds:4657h`（組數），而且是**倒著跑到 0**。
// 這裡改成掃四個槽並跳過空的組——結果一樣，而空組本來就比不中。
func (s *Script) otherGroupOnList() bool {
	w := s.World
	for g := 0; g < PartyGroups; g++ {
		p := w.Groups[g]
		if !p.Active || g == w.GroupIndex || p.MapID != w.MapID {
			continue
		}
		for at := 7; at+1 < len(s.Record); at += 2 {
			if s.Record[at] == 0xFF {
				break
			}
			if s.Record[at] == p.X && s.Record[at+1] == p.Y {
				return true
			}
		}
	}
	return false
}

// stashFirstMember 是 opcode 4（`0x1A54B`）：把**第一個隊員**的記錄
// 抄一份進第 n 格，副本上的物品、裝備、護甲與金錢清掉，並蓋上時間戳。
//
// ⚠ **原本那個隊員一個欄位都不動**——清除全部發生在副本上
// （寫入目標是 `ds:4661h`，不是 `ds:46B5h`）。照抄之前很容易讀反。
func (w *World) stashFirstMember(slot byte) {
	if len(w.Party.Members) == 0 || w.Party.Members[0] == nil {
		return
	}
	c := *w.Party.Members[0] // 淺拷貝之後再換掉兩個 slice，等於深拷貝
	c.Items = nil
	c.Skills = append([]Slot(nil), w.Party.Members[0].Skills...)
	c.EquipIndex, c.ArmorIndex, c.AC, c.Money = 0, 0, 0, 0
	if w.Stash == nil {
		w.Stash = map[byte]StashSlot{}
	}
	w.Stash[slot] = StashSlot{Char: &c, At: w.Clock.Total}
}

// unstash 是 opcode 39 命中之後那一段（`0x1AC5E`–`0x1ACD0`）：
// 把寄放的角色抄進隊伍的下一個空位、人數 ＋1。
//
// ⚠ **原版還會叫玩家打一個名字**（`0x1AC93` 起那段輸入迴圈，13 個字寫回
// 記錄 `+0x00`）。重製版沒有接那個輸入，**沿用副本裡原本的名字**——
// 這是省略一個提示，不是換一條規則；要補的話入口在這裡。
func (w *World) unstash(slot byte) bool {
	st, ok := w.Stash[slot]
	if !ok || st.Char == nil {
		return false
	}
	c := *st.Char
	c.Items = append([]Slot(nil), st.Char.Items...)
	c.Skills = append([]Slot(nil), st.Char.Skills...)
	w.Party.Members = append(w.Party.Members, &c)
	delete(w.Stash, slot)
	return true
}

// stashElapsed 回報第 n 格寄放了多久。沒有那一格時回 0。
func (w *World) stashElapsed(slot byte) uint32 {
	st, ok := w.Stash[slot]
	if !ok {
		return 0
	}
	return w.Clock.Total - st.At
}

// neighbourCells 是 opcode 36 檢查的三格（`0x1AB35` 的 `bl` ＝ 9／10／11、
// `dl` ＝ 0x13）——**地圖上的絕對座標，不是相對隊伍的鄰格**。
var neighbourCells = [3][2]int{{9, 0x13}, {10, 0x13}, {11, 0x13}}

// neighbourTerrain／neighbourRecord 是那三格都要符合的值（`cmp al, 4`／`cmp dl, 2`）。
const (
	neighbourTerrain = 4
	neighbourRecord  = 2
)

// neighboursMatch 回報那三格是不是都是 (nibble 4, 記錄 2)。
func (s *Script) neighboursMatch() bool {
	for _, c := range neighbourCells {
		terrain, record, _, err := s.World.Block.At(c[0], c[1])
		if err != nil || terrain != neighbourTerrain || record != neighbourRecord {
			return false
		}
	}
	return true
}
