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
	KeyPageUp
	KeyPageDown
	KeyY
	KeyN
	KeyV
)

// Mouse 是這一幀的滑鼠狀態（`docs/spec/29`）。
//
// ⚠ **座標是高解畫布上的像素**（960 × 600），不是視窗像素也不是字元格。
// 換算成格子是消費端的事——視窗放大倍率由呈現層吃掉。
//
// Left／Right 是**這一幀剛按下**（edge），不是按住：按住不放應該只算一次，
// 否則點一下會連走好幾格。
type Mouse struct {
	X, Y        int
	Left, Right bool
}

// Any 回報這一幀有沒有按鍵動作。
func (m Mouse) Any() bool { return m.Left || m.Right }

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

// Scroll 是這一幀的捲動請求（Page Up／Page Down）。
//
// **與 `Direction` 分開**是刻意的：方向鍵在地圖上是走路、在清單上是選項，
// 捲動要的是「不管那一層拿方向鍵做什麼，都能把看不到的字翻出來」。
// 兩者共用一個欄位的話，每加一個模式就要再判斷一次「這一次的上是哪一種上」。
type Scroll int

const (
	ScrollNone Scroll = iota
	ScrollUp
	ScrollDown
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
	// Scroll 是這一幀的捲動請求（Page Up／Page Down）。
	Scroll Scroll

	// Mouse 是這一幀的滑鼠。**滑鼠不是新的一條輸入路徑**：
	// 場景會把點擊翻成與鍵盤等價的 Input 再走同一條路（`docs/spec/29` §2）。
	Mouse Mouse
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
		case KeyPageUp:
			in.Scroll = ScrollUp
		case KeyPageDown:
			in.Scroll = ScrollDown
		case KeyEnter, KeySpace:
			if in.Action == ActionNone {
				in.Action = ActionConfirm
			}
		case KeyY:
			if in.Char == 0 {
				in.Char = 'Y'
			}
		case KeyN:
			if in.Char == 0 {
				in.Char = 'N'
			}
		case KeyV:
			if in.Char == 0 {
				in.Char = 'V'
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
