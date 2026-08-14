package input

// 滑鼠：原版把游標位置換成「等同按了哪個鍵」（docs/re/43）。
//
// 原版的作法是一張 21 筆的熱區表（`ds:0CAEBh`，每筆 x1/y1/x2/y2/handler），
// 由 `sub_18860` 由前往後掃，**第一個命中的區域就決定結果並回傳**，
// 掃不到才去讀鍵盤（`sub_18EFE` 先問滑鼠再 `int 16h`）。
//
// 每個區域另外受一張 32 位元的遮罩控制（`ds:7DF3h` 低 16 位、`ds:7DF5h` 高 16 位）：
// 每個詢問畫面在等待輸入之前把遮罩設好，只有位元設起來的區域才會被試。
// 這就是「戰鬥選單時點地圖沒反應」的原因，不是座標判斷，是遮罩。

// RegionKind 是熱區算出按鍵的方式。
type RegionKind int

const (
	// RegionFixed 固定送出一個鍵（螢幕邊上的按鈕，鍵在 `ds:0CBBDh` 的 `+0x03`）。
	RegionFixed RegionKind = iota
	// RegionRoster 隊伍名單：列號換成 '1'…'7'。
	RegionRoster
	// RegionMapView 地圖視窗：四個三角象限換成方向鍵，中央小框是 ESC。
	RegionMapView
	// RegionMessageRows 訊息視窗：查該列由 `\x10` 登記的熱鍵。
	RegionMessageRows
)

// Region 是一塊熱區。座標是原版的 320×200 畫布，邊界含端點。
type Region struct {
	X1, Y1, X2, Y2 int
	Kind           RegionKind
	Key            byte // 只有 RegionFixed 用
}

// Regions 照原版表的順序，順序就是優先序（docs/re/43 §3）。
var Regions = []Region{
	{0, 104, 311, 175, RegionRoster, 0},
	{8, 8, 280, 120, RegionMapView, 0},
	{104, 5, 295, 100, RegionMessageRows, 0},
	{0, 189, 319, 196, RegionFixed, 0x0D},
	{298, 24, 319, 39, RegionFixed, 0xC8},
	{298, 80, 319, 95, RegionFixed, 0xD0},
	{298, 144, 319, 159, RegionFixed, 0xC9},
	{298, 176, 319, 191, RegionFixed, 0xD1},
	{128, 101, 167, 108, RegionFixed, '>'},
	{240, 101, 279, 108, RegionFixed, '<'},
	{128, 0, 159, 7, RegionFixed, 0x1B},
	{240, 0, 279, 7, RegionFixed, ' '},
	{128, 133, 159, 140, RegionFixed, 0x1B},
	{112, 181, 199, 188, RegionFixed, ' '},
	{112, 133, 191, 140, RegionFixed, ' '},
	{160, 101, 247, 108, RegionFixed, 'P'},
	{168, 101, 223, 108, RegionFixed, 0x0D},
	{128, 101, 167, 108, RegionFixed, 'P'},
	{240, 101, 295, 108, RegionFixed, 'D'},
	{128, 101, 159, 108, RegionFixed, ' '},
	{168, 0, 231, 4, RegionFixed, 0x12},
}

// 方向鍵，原版表在 `ds:C05Dh`：上、下、左、右。
var mapViewKeys = [4]byte{'I', 'K', 'J', 'L'}

// RowHotkeys 是訊息視窗「每一列各記一個熱鍵」的表（原版 `ds:8DDCh`，25 格）。
//
// 控制碼 `\x11` 清空、`\x10` 在目前列登記一格；文字視窗捲動時這張表跟著往前移，
// 所以列號永遠對得上畫面。**表比視窗高**（25 格 vs 12 列），捲動不需要特判邊界。
type RowHotkeys [25]byte

// Clear 對應控制碼 `\x11`。
func (r *RowHotkeys) Clear() { *r = RowHotkeys{} }

// Set 對應控制碼 `\x10`：把 ch 登記到第 row 列（**列號 1 起算**，原版要先減 1）。
func (r *RowHotkeys) Set(row int, ch byte) {
	if row < 1 || row > len(r) {
		return
	}
	r[row-1] = ch
}

// Scroll 把整張表往前移一格、尾巴補零，對應文字視窗捲動一行。
func (r *RowHotkeys) Scroll() {
	copy(r[:], r[1:])
	r[len(r)-1] = 0
}

// Screen 是一次點擊要用到的畫面狀態。
type Screen struct {
	// Mask 是熱區遮罩，第 i 位對應 Regions[i]。0 表示這個畫面不收滑鼠。
	Mask uint32
	// Rows 是訊息視窗的每列熱鍵表。
	Rows RowHotkeys
	// PartySize 是隊伍人數，決定名單區可以點到第幾列。
	PartySize int
	// EscAsSpace 時地圖視窗中央送出的 ESC 會變成空白（原版 `ds:8DDAh`）。
	EscAsSpace bool
}

// Click 把一次點擊換成「等同按下的鍵」。回傳 false 代表這一下不算按鍵，
// 呼叫者要繼續等鍵盤——**不是**代表點在畫面外。
func Click(x, y int, s Screen) (byte, bool) {
	for i, r := range Regions {
		if s.Mask&(1<<uint(i)) == 0 {
			continue
		}
		if x < r.X1 || y < r.Y1 || x > r.X2 || y > r.Y2 {
			continue
		}
		// 命中就定案，不再往後找——原版在這裡直接回傳。
		switch r.Kind {
		case RegionFixed:
			return r.Key, true
		case RegionRoster:
			if y < 0x7D {
				return 0, false
			}
			row := (y-0x7D)>>3 + 1
			if row > s.PartySize {
				return 0, false
			}
			return byte('0' + row), true
		case RegionMapView:
			return mapViewKey(x, y, s.EscAsSpace)
		case RegionMessageRows:
			ch := s.Rows[(y-5)>>3]
			if ch == 0 {
				return 0, false
			}
			return ch, true
		}
		return 0, false
	}
	return 0, false
}

// mapViewKey 把地圖視窗裡的座標換成方向鍵。四條邊界是兩組對角線
// （x ＝ y+0x50 與 x ＝ 0xD0−y／0xE0−y），中央 16×16 的小框是 ESC。
func mapViewKey(x, y int, escAsSpace bool) (byte, bool) {
	if y >= 0x38 && y < 0x48 && x >= 0x88 && x < 0x98 {
		if escAsSpace {
			return ' ', true
		}
		return 0x1B, true
	}
	var idx int
	switch {
	case x >= y+0x50:
		if x < 0xD0-y {
			idx = 0 // 上
		} else {
			idx = 3 // 右
		}
	default:
		if x < 0xE0-y {
			idx = 2 // 左
		} else {
			idx = 1 // 下
		}
	}
	return mapViewKeys[idx], true
}
