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
var blocking = map[byte]bool{2: true, 3: true, 15: true}

// nibble 10 是傳送，但記錄 +0x00 的 bit7 設起來時要先問玩家。
const nibbleTeleport = 10

// EventKind 是走完一步之後要處理的事情種類。
type EventKind int

const (
	EventNone EventKind = iota
	EventMessage
	EventMenu
	EventTeleport
	EventEncounter
	EventChest    // nibble 5，內容生成留給規格 05
	EventGate     // nibble 2，條件串列留給規格 06
	EventFacility // nibble 6，設施與腳本留給規格 07
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
	Moved     bool // 被擋住時是 false，而且時鐘與遭遇都不會動
	Periodic  bool // 這一步跨過 16 刻，跑過體力處理
	Encounter bool
	Event     Event
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
	if !w.passable(nx, ny) {
		return res, nil // 被擋住：什麼都不推進
	}

	w.Party.X, w.Party.Y = uint8(nx), uint8(ny)
	w.syncView()
	res.Moved = true

	res.Periodic = w.Clock.Advance(w.stepTime(), w.Block.StepTick())
	if res.Periodic {
		w.Party.Tick16(w.Clock.Tick)
	}

	res.Encounter = w.rollEncounter()
	res.Event = w.trigger(nx, ny)
	return res, nil
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
	if terrain == nibbleTeleport {
		// bit7 設起來時原版會先問玩家，這一層回報可以走，
		// 由 trigger 產生 EventTeleport 讓上層決定。
		_ = record
	}
	return true
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
		// 遠看才顯示的描述——站上去反而不印（0x16CD0 比對隊伍座標）。
		ev.Kind = EventNone
	case 2:
		ev.Kind = EventGate
	case 4, 9, 12:
		ev.Kind = EventMessage
		ev.Strings = []int{int(record)}
	case 5:
		ev.Kind = EventChest
	case 6:
		ev.Kind = EventFacility
	case 8:
		ev.Kind = EventMenu
	case nibbleTeleport:
		ev.Kind = EventTeleport
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
