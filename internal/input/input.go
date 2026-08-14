// Package input 是與呈現層無關的輸入模型。
//
// 為什麼不直接用 Ebiten 的按鍵型別：**Ebiten 在沒有 DISPLAY 的環境會在
// package init 就 panic**，所以只要測試碰到那個套件，無頭環境就跑不了。
// 把按鍵對應放這裡，鍵位邏輯就能在 docker 裡測；`internal/ui` 只剩
// 「Ebiten 的鍵 → 這裡的鍵」一個對照 switch。
package input

// Key 是與函式庫無關的按鍵代碼。
type Key int

const (
	KeyNone Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyW
	KeyA
	KeyS
	KeyD
	KeyEscape
	KeyF10
	KeyEnter
	KeySpace
)

// Direction 照原版的方向編號（docs/re/26 §2）。
type Direction int

const (
	DirNone  Direction = -1
	DirUp    Direction = 0
	DirDown  Direction = 1
	DirLeft  Direction = 2
	DirRight Direction = 3
)

// Action 是非移動的輸入。
type Action int

const (
	ActionNone Action = iota
	ActionConfirm
	ActionCancel // ESC：取消／返回，**不是離開**
	ActionQuit   // F10：離開
)

// Input 是一幀收到的輸入。
type Input struct {
	Dir    Direction
	Action Action
	// Char 是這一幀按下的可列印字元（選單用字首字母選項，docs/re/14 §5）。
	Char byte
}

// Bindings 是預設對應。
//
// ⚠ **這是 remake 的設計決定，不是逆向結論**：原版用哪些鍵移動還沒從
// `sub_1651A` 的呼叫端讀出來（盤點 F1 只解到「等待按鍵」）。
var Bindings = map[Key]Direction{
	KeyUp:    DirUp,
	KeyDown:  DirDown,
	KeyLeft:  DirLeft,
	KeyRight: DirRight,
	KeyW:     DirUp,
	KeyS:     DirDown,
	KeyA:     DirLeft,
	KeyD:     DirRight,
}

// Read 把「這一幀剛按下的鍵」與輸入字元翻成 Input。
func Read(justPressed []Key, runes []rune) Input {
	in := Input{Dir: DirNone}
	for _, k := range justPressed {
		if d, ok := Bindings[k]; ok && in.Dir == DirNone {
			in.Dir = d
		}
		switch k {
		case KeyEscape:
			in.Action = ActionCancel
		case KeyF10:
			in.Action = ActionQuit
		case KeyEnter, KeySpace:
			if in.Action == ActionNone {
				in.Action = ActionConfirm
			}
		}
	}
	for _, r := range runes {
		if r >= 0x20 && r < 0x7F {
			in.Char = byte(r)
			break
		}
	}
	return in
}
