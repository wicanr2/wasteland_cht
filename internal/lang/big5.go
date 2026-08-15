package lang

// UTF-8 → Big5：**中文角色名字是重製版的擴充**，原版只收 ASCII。
//
// 原版的名字欄是角色記錄 `+0x00` 起 13 bytes（`docs/re/15`），
// 顯示走倚天 16×15 字模（`docs/spec/10`）——存的本來就是 Big5 bytes，
// 所以中文名字不需要改記錄格式，只需要把輸入編成 Big5。

import (
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
