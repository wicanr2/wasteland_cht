package game

// 遭遇掃描與遭遇佇列（docs/spec/15、docs/re/39）。
//
// 規格 13 從「已知有一場遭遇」開始建敵人；這一層補上更前面一步——
// **哪一格會冒出遭遇**。

// 佇列的形狀（ds:A96Fh，docs/re/39 §1）：4 個隊伍組 × 4 個槽。
const (
	QueueGroups        = 4
	QueueSlotsPerGroup = 4
)

// 遭遇格的第 1 層 nibble（docs/re/39 §2）。
//
// 這兩個值同時也是 ds:AA87h 那張踩踏跳表裡的項目——同一個 nibble，
// 走過去會觸發，看得見時就會被掃到。
const (
	nibbleEncounterA = 3
	nibbleEncounterB = 15
)

// 遭遇記錄的兩個距離門檻（docs/re/39 §3）。
const (
	recNoticeRange = 0x00 // 察覺距離上限：超過就整格略過
	recActiveRange = 0x01 // 主動距離：不超過就一定進佇列
)

// 接戰值（sub_13878 → ds:46CCh，docs/re/39 §4）。
const (
	EngageNone  = 0x00 // 不能行動（CON ≤ 0）
	EngageClose = 0x0F // 能行動
	EngageFar   = 0xFE // 能行動且射程條件成立
)

// QueueEntry 是佇列裡的一筆。Used 為 false 表示這一格是空的。
type QueueEntry struct {
	X, Y     byte
	Resource byte
	Distance byte // 與最近那一組隊伍的距離（docs/re/39 §4）
	Used     bool
}

// EncounterQueue 是「目前視野裡有哪些遭遇」的快照。
//
// ⚠ **不是待辦清單。** 原版每次掃描一進來就把 64 bytes 全部清成 0
// （docs/re/39 §1），所以這裡也每次重建，不累積。
type EncounterQueue struct {
	Slots [QueueGroups * QueueSlotsPerGroup]QueueEntry
}

// Clear 清空整張佇列。
func (q *EncounterQueue) Clear() { *q = EncounterQueue{} }

// Find 依 (資源, x, y) 找一筆（sub_14AEA 只比這三個 byte，不比優先值）。
func (q *EncounterQueue) Find(resource, x, y byte) (int, bool) {
	for i, e := range q.Slots {
		if e.Used && e.X == x && e.Y == y && e.Resource == resource {
			return i, true
		}
	}
	return 0, false
}

// Insert 把一筆放進第 group 組的四個槽，依距離由近到遠排。
//
// 四槽都滿而且新的一筆不比最遠的那筆近 → 回 false（原版接著會重挑下一近的組，
// 這裡只回報放不進去，由呼叫端決定）。
func (q *EncounterQueue) Insert(group int, e QueueEntry) bool {
	if group < 0 || group >= QueueGroups {
		return false
	}
	base := group * QueueSlotsPerGroup
	e.Used = true

	// 先找空槽。
	for i := base; i < base+QueueSlotsPerGroup; i++ {
		if !q.Slots[i].Used {
			q.Slots[i] = e
			q.sortGroup(group)
			return true
		}
	}
	// 滿了：找最遠的那一格，比它近才擠掉。
	worst := base
	for i := base + 1; i < base+QueueSlotsPerGroup; i++ {
		if q.Slots[i].Distance > q.Slots[worst].Distance {
			worst = i
		}
	}
	if e.Distance >= q.Slots[worst].Distance {
		return false
	}
	q.Slots[worst] = e
	q.sortGroup(group)
	return true
}

// sortGroup 讓同一組的四個槽依距離由近到遠（0x148A8 是「插在第一個比它遠的前面」）。
//
// ⚠ 空槽要排在最後面。空槽的距離是 0，照數值排會被當成「最近的」擠到前面去，
// 而且症狀只有在同組還有空位時才看得到。
func (q *EncounterQueue) sortGroup(group int) {
	base := group * QueueSlotsPerGroup
	key := func(i int) int {
		if !q.Slots[i].Used {
			return 0x100
		}
		return int(q.Slots[i].Distance)
	}
	for i := base + 1; i < base+QueueSlotsPerGroup; i++ {
		for j := i; j > base && key(j) < key(j-1); j-- {
			q.Slots[j], q.Slots[j-1] = q.Slots[j-1], q.Slots[j]
		}
	}
}

// Engagement 是一個隊伍組的接戰值：該組成員接戰值的最大值（docs/re/39 §4）。
//
// far 由呼叫端提供，判準是 `EngageFarFor`（`sub_13878`，`docs/re/123` §2）：
// **裝備的武器有射程而且彈匣不是空的**。呼叫端要拿得到物品表才算得出來，
// 所以留成參數。
func Engagement(members []*Character, far func(int) bool) int {
	best := EngageNone
	for i, m := range members {
		if !CanCommand(m) {
			continue
		}
		v := EngageClose
		if far != nil && far(i) {
			v = EngageFar
		}
		if v > best {
			best = v
		}
	}
	return best
}

// PartyGroupState 是一個隊伍組在這次掃描裡的狀態。
//
// Present 對應原版 sub_149F7 的三道早退（不在這張地圖／這一格不在該組的視窗內／
// 該組沒有敵人可打）。這三件事 remake 目前只有玩家那一組算得出來，
// 其餘三組由呼叫端提供——**不要在這裡編一個出來**。
type PartyGroupState struct {
	// Present ＝ 這一組**在這張地圖上**（`docs/re/39` §4 的第一道判斷）。
	// 不在同一張地圖的組整個不參與這次掃描。
	Present bool
	// X／Y 是這一組的座標。**距離要用各組自己的座標算**——
	// 拿目前這組的座標去算所有組，別組永遠會得到錯的距離。
	X, Y   int
	Engage int // 該組成員接戰值的最大值（EngageNone／EngageClose／EngageFar）
}

// ScanResult 是一次掃描的結果。
type ScanResult struct {
	Queue   EncounterQueue
	Scanned int // 掃過幾格（不含視窗外與地圖外）
	Hits    int // 命中 nibble 3／15 的格子數
}

// IsEncounterCell 回報這個第 1 層 nibble 是不是遭遇格。
func IsEncounterCell(nibble byte) bool {
	return nibble == nibbleEncounterA || nibble == nibbleEncounterB
}

// ScanEncounters 掃地圖視窗，把命中的遭遇放進佇列（docs/spec/15 §3）。
//
// groups 是四個隊伍組的狀態；距離最近的那一組會拿到這一筆。
func (w *World) ScanEncounters(groups [QueueGroups]PartyGroupState) ScanResult {
	var out ScanResult
	out.Queue.Clear()
	if w.Block == nil {
		return out
	}
	size := w.Block.Dim
	for row := 0; row < ViewRows; row++ {
		y := w.ViewY + row
		if y < 0 || y >= size {
			continue
		}
		for col := 0; col < ViewCols; col++ {
			x := w.ViewX + col
			if x < 0 || x >= size {
				continue
			}
			out.Scanned++
			terrain, record, _, err := w.Block.At(x, y)
			if err != nil || !IsEncounterCell(terrain) {
				continue
			}
			out.Hits++

			rec, err := w.Block.SectionRecord(int(terrain), int(record))
			if err != nil || len(rec) <= recActiveRange {
				continue
			}
			if g := ReadSpawnGroups(rec); g[0].Count == 0 &&
				g[1].Count == 0 && g[2].Count == 0 {
				continue // 這一格沒有敵人
			}

			// 對四組各評一次，挑距離最近的那一組（sub_14A65 從 0xFF 起跳）。
			//
			// ⚠ **距離對每一組各算一次**（`docs/re/39` §4）：
			// 原版寫進 `[0x7111 + 組]` 的是「該組到這一格的距離」，
			// 不在這張地圖或太遠的組寫 `0xFF`。拿目前這組的座標去算全部，
			// 別組的判斷會整個錯掉，而症狀只是「別組不會遇敵」。
			best, bestDist := -1, 0x100
			for g, st := range groups {
				if !st.Present {
					continue
				}
				dist, ok := Distance(x-st.X, y-st.Y)
				if !ok || dist > int(rec[recNoticeRange]) {
					continue // 超出察覺距離（視窗外也算）
				}
				if dist > int(rec[recActiveRange]) && dist >= st.Engage {
					continue // 這一組吃不到這場遭遇
				}
				if dist < bestDist {
					best, bestDist = g, dist
				}
			}
			if best < 0 {
				continue
			}
			out.Queue.Insert(best, QueueEntry{
				X: byte(x), Y: byte(y),
				Resource: byte(w.Block.Resource.ID), Distance: byte(bestDist),
			})
		}
	}
	return out
}

// Nearest 取某一組裡最近的那一筆。
//
// 槽內已經以距離由近到遠排好（sortGroup，docs/re/39 §4），
// 所以「最近」就是第一個 Used 的槽——**不要再排一次**，
// 原版的排序規則是「插在第一個比它遠的前面」，重排會換掉同距離的先後。
func (q *EncounterQueue) Nearest(group int) (QueueEntry, bool) {
	if group < 0 || group >= QueueGroups {
		return QueueEntry{}, false
	}
	for i := 0; i < QueueSlotsPerGroup; i++ {
		if e := q.Slots[group*QueueSlotsPerGroup+i]; e.Used {
			return e, true
		}
	}
	return QueueEntry{}, false
}

// EngageFarFor 回報這個人夠不夠格拿到 `EngageFar`（`sub_13878` 的 `0xFE` 那條）。
//
// 原版的三步（`0x1388B`–`0x1389B`）：
//
//	sub_196B2 → 裝備武器那一格的物品資料
//	sub_199F1 → 物品資料 +0x03 ＝ 類別；**是 0 就沒武器**，退回近戰
//	sub_19D2F → 類別在 ds:CD00h 的射程清單裡才繼續
//	sub_138C1 → 彈匣裡的發數（物品槽附屬 byte 的低 6 位）是 0
//	            就走 sub_1391E，把接戰值**壓回 0x0F**
//
// 所以「有射程」不只看武器種類，**還要有子彈**。三個條件缺一個都是 `EngageClose`，
// 而接戰值就是攻擊挑得到誰的距離上限（`docs/re/123` §2）——
// 少了這一段，空彈匣的槍照樣打得到遠處的敵人。
func EngageFarFor(c *Character, items ItemTable) bool {
	if c == nil || items == nil || c.EquipIndex == 0 {
		return false
	}
	i := int(c.EquipIndex)
	if i >= len(c.Items) {
		return false
	}
	slot := c.Items[i]
	if slot.ID == 0 {
		return false
	}
	d, ok := items.Get(slot.ID)
	if !ok || !d.Class.Ranged() {
		return false
	}
	// 附屬 byte 的低 6 位是剩餘發數（高 2 位另有用途，`docs/re/21` §5.1）。
	return slot.Value&0x3F != 0
}
