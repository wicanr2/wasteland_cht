package lang

// UTF-8 → Big5：**中文角色名字是重製版的擴充**，原版只收 ASCII。
//
// 原版的名字欄是角色記錄 `+0x00` 起 13 bytes（`docs/re/15`），
// 顯示走倚天 16×15 字模（`docs/spec/10`）——存的本來就是 Big5 bytes，
// 所以中文名字不需要改記錄格式，只需要把輸入編成 Big5。

import (
	"sync"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// ToBig5 把一段 UTF-8 文字編成 Big5。
//
// 編不出來的字回 false——**不要用問號或空白代替**，
// 那會讓玩家看到一個自己沒打過的名字（`docs/spec/11` §7 的同一條原則）。
func ToBig5(s string) ([]byte, bool) {
	out, _, err := transform.Bytes(traditionalchinese.Big5.NewEncoder(), []byte(s))
	if err != nil {
		return nil, false
	}
	return out, true
}

// FromBig5 把 Big5 bytes 解回 UTF-8，給**驗收工具**印在終端機上用
// （`cmd/wl-play` 的 trace）。遊戲本體不需要這個方向——畫面直接吃 Big5
// 查倚天字模，中間不經過 UTF-8。
//
// 解不出來的回 false，呼叫端自己決定要不要退回十六進位。
func FromBig5(b []byte) (string, bool) {
	out, _, err := transform.Bytes(traditionalchinese.Big5.NewDecoder(), b)
	if err != nil {
		return "", false
	}
	return string(out), true
}

// RuneToBig5 是單一字元的版本，給逐鍵輸入用。
func RuneToBig5(r rune) ([]byte, bool) {
	if r < 0x20 {
		return nil, false
	}
	if r < 0x7F {
		return []byte{byte(r)}, true // ASCII 直接過
	}
	return ToBig5(string(r))
}

// —— 字元 → Big5 的快取 ————————————————————————————————————
//
// ⚠ **這是 UTF-8 那條路唯一新增的執行期成本。** 畫面上每個漢字每一幀都要
// 轉一次（`render.DrawRune`），而 `x/text` 的 transform 每次配置三塊記憶體。
// 量過：24 個字 72 次配置。一畫面兩百個字、每秒 60 幀 ＝ 每秒三萬六千次配置，
// 對遊戲迴圈是真的 GC 壓力。
//
// 快取讓它變成一次 map 查詢。**Big5 是固定對照表**，
// 同一個字永遠對到同一組 bytes，快取不會失效。
var (
	big5Mu    sync.RWMutex
	big5Cache = map[rune][2]byte{}
)

// CachedRuneToBig5 是 RuneToBig5 的快取版，給繪製這條熱路徑用。
//
// 只回兩個 byte 的漢字；ASCII 與編不出來的字讓呼叫端自己處理
// （回 ok ＝ false），這樣快取裡只放固定形狀的東西。
func CachedRuneToBig5(r rune) ([2]byte, bool) {
	big5Mu.RLock()
	b, ok := big5Cache[r]
	big5Mu.RUnlock()
	if ok {
		return b, b != [2]byte{}
	}
	var out [2]byte
	if enc, ok := ToBig5(string(r)); ok && len(enc) == 2 {
		out = [2]byte{enc[0], enc[1]}
	}
	big5Mu.Lock()
	big5Cache[r] = out
	big5Mu.Unlock()
	return out, out != [2]byte{}
}
