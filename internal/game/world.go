package game

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 走一步的規則層（docs/spec/04、docs/re/26、27）。
// 這一層不認識畫面：它回傳「發生了什麼」，由呈現層決定怎麼顯示。

// Direction 是四個走法，編號照原版的跳表（ds:AAB1h）。
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

var delta = [4][2]int{
	Up:    {0, -1},
	Down:  {0, +1},
	Left:  {-1, 0},
	Right: {+1, 0},
}

// 地圖視窗（docs/re/25）：19 × 9 格，隊伍固定在第 (9, 4) 格。
//
// 遭遇掃描的範圍就是這一塊，距離表涵蓋的也正好是這個半徑（docs/re/39 §2）。
const (
	ViewCols    = 19
	ViewRows    = 9
	ViewOffsetX = 9
	ViewOffsetY = 4
)

// 目前確認會擋住移動的第 1 層 nibble（docs/re/26 §3）。
//
// ⚠ 這是**下界不是全集**：原版還有 sub_13E9B／sub_15CE0 兩支沒讀完，
// 所以其餘 nibble 一律當成可通行，遇到與原版不符的地圖要回來補。
// blocking 是**無條件**擋住移動的 nibble（docs/re/62 §1）。
//
// nibble 11 佔 42 張地圖的第二多（20,495 格）——**山、牆、水**。
// 漏掉它玩家會穿山，而症狀是「路線與原版不同」而不是明顯的錯誤。
//
// **nibble 2 不在這裡**：它是條件式的，見 gateNeedsCheck（docs/re/65）。
var blocking = map[byte]bool{3: true, 11: true, 15: true}

// nibble 2 是條件式的障礙（門、鎖、檢定）。
const nibbleCondition = 2

// Skills 是技能資料表，判定條件閘要用（呼叫端載入，規則層不讀執行檔）。
// 沒設的話技能型別的條件一律失敗。
type SkillSource = SkillTable

// gateNeedsCheck 回報這一格要不要跑條件判定（原版 sub_13E9B 的前兩道分支）。
//
// 三條路見 docs/re/65 §1.1；這一支只回答「要不要判定」，**不判定**——
// 判定有副作用（擲骰、消耗鑰匙），只能在真的要走過去時做一次。
func (w *World) gateNeedsCheck(x, y int) (need bool, rec []byte) {
	terrain, record, _, err := w.Block.At(x, y)
	if err != nil || terrain != nibbleCondition {
		return false, nil
	}
	r, err := w.Block.SectionRecord(int(terrain), int(record))
	if err != nil || len(r) == 0 {
		return false, nil
	}
	if r[0]&0x80 != 0 || r[0]&0x40 == 0 {
		return false, nil
	}
	return true, r
}

// nibble 10 是傳送，但記錄 +0x00 的 bit7 設起來時要先問玩家。
const nibbleTeleport = 10

// nibble 4 是條件式的障礙：記錄 +0x01 的 bit7 設起來才擋（docs/re/62 §1）。
const nibbleBarrier = 4

// EventKind 是走完一步之後要處理的事情種類。
type EventKind int

const (
	EventNone EventKind = iota
	EventMessage
	EventMenu
	EventTeleport
	EventEncounter
	EventChest     // nibble 5，內容生成留給規格 05
	EventGate      // nibble 2，條件串列留給規格 06
	EventFacility  // nibble 6，設施與腳本留給規格 07
	EventRadiation // nibble 9，結算在 ApplyRadiation（docs/spec/07 §6.4）
)

// Event 是規則層交給呈現層的結果。字串一律以編號表示，規則層不碰文字。
type Event struct {
	Kind    EventKind
	Nibble  byte
	Record  int   // 這一格的第 2 層值（記錄編號）
	Strings []int // 要顯示的字串編號
	Choices []int // 選單選項的字串編號
	To      [2]uint8
	Data    []byte // 這一格的 section 記錄（取不到時是 nil）
	// PatchAt 是這一類事件收尾要拿記錄的哪個位移去改寫這一格
	// （原版 `sub_169B1(al)` → `sub_17CFF`）；0 ＝ 這一類不改寫。
	PatchAt int
}

// messageListLen 回報記錄開頭那串訊息編號有幾條（`sub_16CD0` 的迴圈）。
//
// 從 +0x00 起逐 byte，**bit7 設的那一條是最後一條**；
// 個數同時就是收尾改寫要用的位移。
func messageListLen(record []byte) int {
	for i, b := range record {
		if b&0x80 != 0 {
			return i + 1
		}
	}
	return len(record)
}

// World 是「目前在哪張地圖、隊伍在哪、幾點了」。
type World struct {
	Block *assets.Block
	Party *Party
	Clock Clock
	RNG   *rng.State

	// ViewX/ViewY 是地圖視窗的原點（ds:464Eh/464Fh）。
	// 原版把它與隊伍座標分開存，這裡照做，不要用其中一個推另一個。
	ViewX, ViewY int
	// confirmed ＝ 玩家剛對「進新地點？」答了 Yes，下一步跳過確認閘。
	confirmed bool

	// carryTerrain／carryRecord 是原版的 `ds:46FCh`／`46FDh`：
	// 上一次被改寫的那一格**在改寫前**的 (nibble, 記錄)，
	// 給記錄裡的 `0xFE`／`0xFD` 沿用（docs/re/69 §9）。
	carryTerrain byte
	carryRecord  byte

	// lastNibble／lastRecord 是原版的 `ds:4716h`／`4717h`：上一次觸發的
	// (nibble, 記錄)。nibble 1 拿它避免連續踩同一種格子重複印訊息。
	lastNibble byte
	lastRecord byte
	// Skills 是技能資料表（條件閘的技能型別要用）。
	Skills SkillTable
	// SelfDestruct 是科奇斯基地的自毀倒數，也就是**結局的觸發點**。
	SelfDestruct SelfDestruct

	// MapID／GroupIndex／Groups 是腳本 opcode 0 要比對的東西：
	// 目前這張地圖的編號、目前是第幾組、四組各自的位置。
	// **由呈現層在換地圖／切組時填**（存檔的隊伍槽表是 `internal/play` 拿著的）。
	// 沒填的話 opcode 0 一律走「沒有別組站在上面」那一條。
	MapID      byte
	GroupIndex int
	Groups     [PartyGroups]GroupPos

	// Stash 是腳本 opcode 4 寄放的角色，opcode 39 領回（`docs/re/102` §6）。
	// **不進存檔**：原版放在執行檔資料段的 `ds:065Ch`，與自毀倒數同一類。
	Stash map[byte]StashSlot
}

// NewWorld 把隊伍放到指定座標並對齊視窗原點。
func NewWorld(b *assets.Block, p *Party, r *rng.State) *World {
	w := &World{Block: b, Party: p, RNG: r}
	w.syncView()
	return w
}

func (w *World) syncView() {
	w.ViewX = int(w.Party.X) - ViewOffsetX
	w.ViewY = int(w.Party.Y) - ViewOffsetY
}

// StepResult 是一次移動的完整結果。
type StepResult struct {
	// Blocked 是被擋住時要印的字串編號（0 ＝ 沒有訊息或沒被擋）。
	Blocked int
	// Ask 非 0 時這一步**停在原地等玩家回答**（進新地點的確認）。
	Ask int
	// AskString 是問之前先印的那一句（目標地點名，`0x16AEA` 的 `sub_16B17`）。
	// 原版兩句一起出現：「進入石英城。」＋「要進入新地點嗎？」
	AskString int
	// Gate 是踩上 nibble 2 之後條件閘跑出來的結果（沒過的人已經受罰）。
	Gate      GateResult
	Moved     bool // 被擋住時是 false，而且時鐘與遭遇都不會動
	Periodic  bool // 這一步跨過 16 刻，跑過體力處理
	Encounter bool
	Event     Event
	// Script 是 nibble 6 的腳本指令跑完的結果（Op ＝ −1 表示沒跑）。
	Script ScriptResult
}

// Step 走一步。順序照原版，不可重排（docs/spec/04 §3）：
// 閘 → 座標 → 時鐘 → 音效 → 遭遇 → 事件。
func (w *World) Step(dir Direction) (StepResult, error) {
	var res StepResult
	if dir < Up || dir > Right {
		return res, fmt.Errorf("方向 %d 不在 0–3", dir)
	}

	nx := int(w.Party.X) + delta[dir][0]
	ny := int(w.Party.Y) + delta[dir][1]
	// 第三道閘：進新地點之前要先問（原版會停在原地等 Yes／No）。
	if !w.confirmed && w.confirmNeeded(nx, ny) {
		res.Ask = AskEnterString
		res.AskString = w.teleportMessage(nx, ny)
		return res, nil
	}
	w.confirmed = false
	// 條件閘：真的要走過去時才判定一次（會擲骰、會消耗物品）。
	if need, rec := w.gateNeedsCheck(nx, ny); need {
		if g := w.Party.EvalGate(w.RNG, rec, w.Skills); g.Blocked {
			if len(rec) > 1 {
				res.Blocked = int(rec[1])
			}
			return res, nil
		}
	}
	if !w.passable(nx, ny) {
		// 被擋住：什麼都不推進，但**原版會印那一格記錄的訊息**
		// （`This mountain is in your way.`，docs/re/62 §2）——
		// 印死的 "BLOCKED." 會讓對拍差一整句話。
		res.Blocked = w.blockedMessage(nx, ny)
		return res, nil
	}

	w.Party.X, w.Party.Y = uint8(nx), uint8(ny)
	w.syncView()
	res.Moved = true
	w.Party.PlayerStepped = true // ds:916Bh ← 方向 − 4，方向 0–3 都是非零

	res.Periodic = w.Clock.Advance(w.stepTime(), w.Block.StepTick())
	if res.Periodic {
		w.Party.Tick16(w.Clock.Tick)
	}

	res.Encounter = w.rollEncounter()

	// 事件是**迴圈**：nibble 6 的腳本指令回報 CF ＝ 0（Continue）時，
	// 原版會把這一格換成「下一步」之後繼續跑（`docs/re/34` §1）。
	// 沙漠高溫就是這樣的兩段：腳本 opcode 3 依晝夜把格子換成
	// nibble 2 記錄 7–9 或 10–12，換完立刻跑那一格的條件閘（docs/re/75）。
	for i := 0; i < maxEventChain; i++ {
		res.Event = w.trigger(nx, ny)

		// nibble 2 的**事件**那一側：記錄 +0x00 的 bit6 設起來就跑條件串列，
		// 沒過的人各自受罰（docs/re/67 §3）。移動閘只管 bit7 沒設的格子，
		// 而沙漠那 163 格是 0xE1——bit7 設所以走得過去，bit6 設所以踩上去有事。
		if res.Event.Kind == EventGate && len(res.Event.Data) > 0 &&
			res.Event.Data[0]&0x40 != 0 {
			res.Gate = w.Party.EvalGate(w.RNG, res.Event.Data, w.Skills)
			w.applyCellPatch(nx, ny, res.Event.Data, res.Gate.PatchAt)
			break
		}
		if res.Event.PatchAt <= 0 {
			break
		}
		// nibble 12 會先改寫**別的格子**，再改寫自己這一格。
		if res.Event.Nibble == 12 {
			w.applyBatchPatch(nx, ny, res.Event.Data)
		}
		// nibble 6 的腳本先跑指令（可能改寫 +0x01／+0x02 ＝ 下一步），
		// 再由下面那一行用位移 1 把這一格換成下一步。
		chain := false
		if res.Event.Nibble == 6 {
			res.Script = w.runScript(res.Event.Data)
			chain = res.Script.Handled && res.Script.Continue
		}
		// nibble 1（位移 ＝ 訊息條數）與 nibble 6（位移 1）跑完都會改寫這一格。
		w.applyCellPatch(nx, ny, res.Event.Data, res.Event.PatchAt)
		if !chain {
			break
		}
	}
	// ds:4716h／4717h：記住這一次觸發的是哪一種格子。
	w.lastNibble, w.lastRecord = res.Event.Nibble, byte(res.Event.Record)
	return res, nil
}

// Confirm 讓下一次 Step 跳過確認閘（玩家答了 Yes，docs/re/64）。
func (w *World) Confirm() { w.confirmed = true }

// runScript 跑 nibble 6 的腳本指令（設施不走這裡）。
//
// 跑完由呼叫端用位移 1 改寫這一格——指令改的 `+0x01`／`+0x02` 就是「下一步」，
// 而這台直譯器的程式計數器就是地圖格本身（docs/re/71 §5.1）。
func (w *World) runScript(record []byte) ScriptResult {
	if len(record) == 0 {
		return ScriptResult{Op: -1, Message: -1, Sound: -1}
	}
	// bit7 設且 kind < 5 才是設施；kind ≥ 5 是**直接指定 opcode**的腳本
	// （`docs/re/79` §1），NewScript 會處理這兩條路。
	if record[0]&0x80 != 0 && int(record[0]&0x7F) < FacilityKindCount {
		return ScriptResult{Op: -1, Message: -1, Sound: -1}
	}
	return NewScript(w, record).Step()
}

// PatchHere 用記錄的指定位移改寫**隊伍腳下**那一格（原版 `sub_169B1`）。
//
// 傳送的收尾走這一條（位移 4）：記錄 `+0x04`／`+0x05` 是「進去之後
// 這一格變成什麼」，設施就是這樣從傳送格變出來的（docs/re/73）。
func (w *World) PatchHere(record []byte, at int) {
	w.applyCellPatch(int(w.Party.X), int(w.Party.Y), record, at)
}

// Passable 回報這一格走不走得進去（四道閘裡與地形有關的那幾道）。
//
// 給驗證工具尋路用（`cmd/wl-play` 的 `path=`）。**不含**傳送與事件的副作用，
// 所以它回 true 不代表走過去不會發生別的事。
func (w *World) Passable(x, y int) bool { return w.passable(x, y) }

// applyCellPatch 照條件閘的收尾改寫這一格（原版 sub_17CFF）。
//
// 記錄在 `at`／`at+1` 給新的第 1 層與第 2 層。第一個 byte 的 bit7 設 ＝ 不改；
// `0xFE`／`0xFD` 改用 `carry`——**上一次被改寫的那一格在改寫前是什麼**
// （原版 `ds:46FCh`／`46FDh`，docs/re/69 §9）。
//
// 那個暫存是全域的：跨格、跨地圖都不清空，但它住在資料段不在存檔裡，
// 所以重新載入就回到零值。
func (w *World) applyCellPatch(x, y int, record []byte, at int) {
	p, reuse, ok := ParseCellPatch(record, at)
	if !ok || p.Skip {
		return
	}
	if reuse {
		p.Terrain, p.Record = w.carryTerrain, w.carryRecord
	}
	if p.Terrain > 0x0F {
		return // 溢出 4 bits 的資料異常，交給 SetCell 擋（這裡先不改）
	}
	// 原版在寫下去**之前**先讀這一格目前的值存起來（`0x17D24` 的 sub_17CD2）。
	if now, rec, _, err := w.Block.At(x, y); err == nil {
		w.carryTerrain, w.carryRecord = now, rec
	}
	_ = w.Block.SetCell(x, y, p.Terrain, p.Record)
}

// batchPatchStride 是 nibble 12 批次改寫表的一筆長度（旗標、x、y、第 1 層、第 2 層）。
const batchPatchStride = 5

// batchPatchEnd 回報 nibble 12 的批次表結束在哪個位移。
//
// 表從 +0x01 起，每筆 5 bytes，**旗標的 bit7 設起來的那一筆是最後一筆**；
// 結束位移同時就是「改寫自己這一格」要用的位移（`0x12C48` 的 `al ← bl + 5`）。
func batchPatchEnd(record []byte) int {
	at := 1
	for at+batchPatchStride-1 < len(record) {
		flags := record[at]
		at += batchPatchStride
		if flags&0x80 != 0 {
			break
		}
	}
	return at
}

// applyBatchPatch 跑 nibble 12 的批次改寫（`0x12BE5`–`0x12C46`，docs/re/71）。
//
// 每筆 5 bytes：
//
//	+0 旗標：bit0 ＝ 座標相對**隊伍**（否則相對地圖原點）、bit7 ＝ 最後一筆
//	+1／+2 目標座標（要加上基準）
//	+3／+4 新的第 1 層與第 2 層（走 sub_17CFF，所以 0xFE／0xFD 一樣適用）
//
// 踩一格可以改寫地圖上任意多格——**遠端改寫**是這個遊戲少數不作用在腳下的機制。
func (w *World) applyBatchPatch(px, py int, record []byte) {
	for at := 1; at+batchPatchStride-1 < len(record); at += batchPatchStride {
		flags := record[at]
		baseX, baseY := 0, 0
		if flags&1 != 0 {
			baseX, baseY = px, py
		}
		x := baseX + int(record[at+1])
		y := baseY + int(record[at+2])
		if x >= 0 && y >= 0 && x < w.Block.Dim && y < w.Block.Dim {
			w.applyCellPatch(x, y, record, at+3)
		}
		if flags&0x80 != 0 {
			return
		}
	}
}

// blockedMessage 回報擋住這一步的那一格要印的字串編號（0 ＝ 不印）。
//
// 原版在第四道閘（sub_15CE0）擋住之後呼叫 sub_16D1A(bl ＝ 0)，
// 那一支讀的是記錄 +0x00，值 0 就不印（docs/re/58 §1、docs/re/62 §2）。
func (w *World) blockedMessage(x, y int) int {
	if x < 0 || y < 0 || x >= w.Block.Dim || y >= w.Block.Dim {
		return 0
	}
	terrain, record, _, err := w.Block.At(x, y)
	if err != nil {
		return 0
	}
	if need, rec := w.gateNeedsCheck(x, y); need && len(rec) > 1 {
		return int(rec[1]) // 條件閘的訊息在 +0x01（docs/re/65 §1）
	}
	rec, err := w.Block.SectionRecord(int(terrain), int(record))
	if err != nil || len(rec) == 0 {
		return 0
	}
	return int(rec[0])
}

// IdleStep 是「原地不動的一步」（方向碼 4，docs/re/26 §1.1）。
//
// 捲動跳表的第 5 筆是 `clc; retn`——隊伍不動。兩個觸發點都在主迴圈：
// 節拍計數器滿 0x400（0x16B91），或 ds:46E1h 非 0（0x16B3F）。
//
// ⚠ **目前沒有人呼叫它，而且先不要接。** 實機站著 45 秒時鐘一分鐘都沒走
// （docs/re/47 §6.1），所以下面這個「時鐘前進」與原版對不上；
// 原版那條路走 sub_1651A(4)，還沒讀到底（docs/re/26 §1.2 記了下一個入口）。
// 接上去之前先把它讀完，不要因為「規格裡有」就接。
//
// ⚠ **這一步不算玩家走的**：PlayerStepped 清成 false，之後的檢定就不給
// 經驗值（ds:916Bh，docs/re/32 §7.1）。這是原版的防刷，不要順手改掉。
func (w *World) IdleStep() StepResult {
	var res StepResult
	w.Party.PlayerStepped = false

	res.Periodic = w.Clock.Advance(w.stepTime(), w.Block.StepTick())
	if res.Periodic {
		w.Party.Tick16(w.Clock.Tick)
	}
	res.Encounter = w.rollEncounter()
	// Moved 維持 false：座標沒動，也不重跑踩上去的事件。
	return res
}

func (w *World) stepTime() uint16 {
	h := w.Block.Header
	return uint16(h[0x34]) | uint16(h[0x35])<<8
}

// passable 是 docs/spec/04 §3.1 的四道閘。
func (w *World) passable(x, y int) bool {
	if x < 0 || y < 0 || x >= w.Block.Dim || y >= w.Block.Dim {
		return false
	}
	terrain, record, _, err := w.Block.At(x, y)
	if err != nil {
		return false
	}
	if blocking[terrain] {
		return false
	}
	// ⚠ **條件格在這裡一律當成可走。** 判定有副作用（擲骰、消耗鑰匙），
	// 而 passable 會被尋路等唯讀路徑呼叫——判定放在 Step 裡做一次
	// （docs/re/65 §1.1）。

	if terrain == nibbleBarrier {
		// nibble 4：記錄 +0x01 的 bit7 設起來才擋（門、關卡），
		// 沒設的是一般的疊圖格，可以走（docs/re/62 §1）。
		if rec, err := w.Block.SectionRecord(int(terrain), int(record)); err == nil &&
			len(rec) > 1 && rec[1]&0x80 != 0 {
			return false
		}
	}
	return true
}

// maxEventChain 是一步之內最多連跑幾個事件。
//
// 腳本把這一格換成「下一步」之後會繼續跑，原版靠指令自己回報 CF ＝ 1 收手；
// remake 加一個上限，資料壞掉時才不會卡死。**正常資料碰不到這個上限**。
const maxEventChain = 8

// AskEnterString 是「Enter new location?」的字串編號（執行檔字串表 1 第 103 條，
// `sub_16AD5` 的 `mov al, 67h`）。
const AskEnterString = 0x67

// confirmNeeded 回報走進這一格之前要不要先問玩家（第三道閘 sub_16AD5）。
//
// ⚠ **判準是記錄 `+0x00` 的 bit6**，不是 bit7——原版是 `shl al, 1` 之後看符號
// （docs/re/64 §1）。Quartz 入口的 `+0x00` 是 `0x41`，bit6 設起來所以會問。
func (w *World) confirmNeeded(x, y int) bool {
	terrain, record, _, err := w.Block.At(x, y)
	if err != nil || terrain != nibbleTeleport {
		return false
	}
	rec, err := w.Block.SectionRecord(int(terrain), int(record))
	if err != nil || len(rec) == 0 {
		return false
	}
	return rec[0]&0x40 != 0
}

// teleportString 取出 nibble 10 記錄要印的字串編號（`+0x00` 的低 6 位）。
func teleportString(rec []byte) int {
	if len(rec) == 0 {
		return 0
	}
	return int(rec[0] & 0x3F)
}

// teleportMessage 是同一件事的座標版，給確認流程用。
func (w *World) teleportMessage(x, y int) int {
	terrain, record, _, err := w.Block.At(x, y)
	if err != nil || terrain != nibbleTeleport {
		return 0
	}
	rec, err := w.Block.SectionRecord(int(terrain), int(record))
	if err != nil {
		return 0
	}
	return teleportString(rec)
}

// rollEncounter 照 docs/spec/04 §5：分母為 0 就不擲。
func (w *World) rollEncounter() bool {
	denom := w.Block.Header[0x2F]
	if denom == 0 || w.RNG == nil {
		return false
	}
	return w.RNG.Roll(int(denom)) == 1
}

// trigger 依第 1 層的 nibble 分派（docs/re/26 §5）。
// 七個 nibble 什麼都不做，其餘八個各有處理。
func (w *World) trigger(x, y int) Event {
	terrain, record, _, err := w.Block.At(x, y)
	if err != nil {
		return Event{}
	}
	ev := Event{Nibble: terrain, Record: int(record)}
	// section 型別 ＝ 這一格的 nibble 本身（docs/spec/07 §2）。
	ev.Data, _ = w.Block.SectionRecord(int(terrain), int(record))

	switch terrain {
	case 1:
		// 氛圍敘述（`sub_16CD0`，docs/re/70）：訊息**可以有很多條**，
		// 從記錄 +0x00 起逐 byte，bit7 設的那一條是最後一條。
		// 連續踩到同一種 (nibble, 記錄) 就不重印——原版比對 ds:4716h／4717h。
		ev.Kind = EventMessage
		n := messageListLen(ev.Data)
		if w.lastNibble != terrain || w.lastRecord != record {
			for i := 0; i < n; i++ {
				if s := ev.Data[i] & 0x7F; s != 0 {
					ev.Strings = append(ev.Strings, int(s))
				}
			}
		}
		// 訊息串列的長度就是收尾改寫要用的位移（`sub_169B1(al ＝ 個數)`）。
		ev.PatchAt = n
	case 2:
		// nibble 2 的事件處理與移動閘是**同一支**（sub_13EC9）：走得過去的格子
		// 踩上去還是會印記錄 +0x01 的訊息（沙漠高溫就是這樣，docs/re/66）。
		ev.Kind = EventGate
		if len(ev.Data) > 1 && ev.Data[1] != 0 {
			ev.Strings = []int{int(ev.Data[1])}
		}
	case 4, 9:
		// 字串編號是**記錄 +0x00**，不是這一格的第 2 層值——
		// 第 2 層是「第幾筆記錄」，兩者的值域差很遠（docs/re/29 §2、§5.1）。
		// 編號 0 ＝ 不印（sub_16D1A 的 `test al,al`）。
		ev.Kind = EventMessage
		if len(ev.Data) > 0 && ev.Data[0] != 0 {
			ev.Strings = []int{int(ev.Data[0])}
		}
		if terrain == 9 {
			ev.Kind = EventRadiation
		}
	case 12:
		// 字串編號是**記錄 +0x00**（`0x12BD0` 的 `mov bl,0` → `sub_17920`），
		// 不是這一格的第 2 層值。
		ev.Kind = EventMessage
		if len(ev.Data) > 0 && ev.Data[0] != 0 {
			ev.Strings = []int{int(ev.Data[0])}
		}
		// +0x01 起是**批次改寫表**，跑完之後再改寫自己這一格（docs/re/71）。
		ev.PatchAt = batchPatchEnd(ev.Data)
	case 5:
		ev.Kind = EventChest
	case 6:
		// 跑完設施／腳本之後用**位移 1** 改寫這一格（`0x12C70` 的
		// `mov al,1` → `sub_169B1`）。設施記錄那 30 筆幾乎都是 `fd fd`，
		// 也就是「改回原樣」——這是 0xFD 在出貨資料裡最大的用戶。
		ev.Kind = EventFacility
		ev.PatchAt = 1
	case 8:
		ev.Kind = EventMenu
	case nibbleTeleport:
		ev.Kind = EventTeleport
		// 記錄 +0x00 的**低 6 位是字串編號**（`sub_16B17`：
		// `mov bl,0` → `and al,3Fh` → `sub_17920`）。那句話就是
		// 「進入石英城。」——`0x16A21` 在 `ds:AAD3h ＝ 0` 時印它，
		// 也就是**不問「進新地點？」的傳送格**；會問的那一種改由
		// `0x16AEA` 在問之前先印（`Step` 的 `AskString`）。
		if s := teleportString(ev.Data); s != 0 {
			ev.Strings = []int{s}
		}
	default:
		ev.Kind = EventNone
	}
	return ev
}

// Teleport 把隊伍搬到指定座標（nibble 10 的效果，目的地由隊伍槽表提供）。
func (w *World) Teleport(x, y uint8) {
	w.Party.X, w.Party.Y = x, y
	w.syncView()
}

// EnterMap 換地圖：刻歸零，其餘時間不動。
func (w *World) EnterMap(b *assets.Block, x, y uint8) {
	w.Block = b
	w.Party.X, w.Party.Y = x, y
	w.syncView()
	w.Clock.EnterMap()
}
