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
	KeyF1
	KeyF2
	KeyF5
	KeyF9
	KeyF10
	KeyEnter
	KeySpace
)

// Direction 照原版的方向編號（docs/re/26 §2）。
type Direction int

// ⚠ **`Direction` 的零值是 `DirUp` 不是 `DirNone`。** 0–3 對應原版捲動跳表的
// 索引（`docs/re/26` §1.1），所以 `DirNone` 只能擺在 −1。
// **自己組 `Input` 時一定要明確寫 `Dir: DirNone`**——忘了就會變成「每一幀往上走」，
// 而症狀是遠處的怪事（按字母鍵卻在移動、訊息莫名其妙變成 BLOCKED）。
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
	ActionQuit   // F10：離開（要先跳確認、先存檔，見 Scene.updateQuit）
)

// Function 是功能鍵。
//
// **與 Action 分開**是刻意的：`ActionCancel` 有「退一層」的語意，會被每一層
// 子模式接住；功能鍵要的是「不管在哪一層都叫得出來」，兩者的路由不一樣。
type Function int

const (
	FnNone      Function = iota
	FnHelp               // F1
	FnSettings           // F2
	FnQuickSave          // F5
	FnQuickLoad          // F9
)

// Input 是一幀收到的輸入。
type Input struct {
	Dir    Direction
	Action Action
	// Char 是這一幀按下的可列印字元（選單用字首字母選項，docs/re/14 §5）。
	Char byte
	// Runes 是這一幀輸入的完整字元（含中文）。
	Runes []rune
	// Fn 是這一幀按下的功能鍵（F1／F2／F5／F9）。
	Fn Function
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
		case KeyF1:
			in.Fn = FnHelp
		case KeyF2:
			in.Fn = FnSettings
		case KeyF5:
			in.Fn = FnQuickSave
		case KeyF9:
			in.Fn = FnQuickLoad
		case KeyEnter, KeySpace:
			if in.Action == ActionNone {
				in.Action = ActionConfirm
			}
		}
	}
	for _, r := range runes {
		if r >= 0x20 && r < 0x7F && in.Char == 0 {
			in.Char = byte(r)
		}
	}
	// Runes 保留這一幀的完整輸入。**中文名字要用它**——`Char` 是單一 byte，
	// 一個中文字進不去（重製版的擴充，原版只收 ASCII）。
	if len(runes) > 0 {
		in.Runes = append(in.Runes[:0], runes...)
	}
	return in
}
