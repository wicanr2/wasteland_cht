package play

// 指令列的 Y／N 確認（`sub_19B4F`，`docs/re/91` §2）。
//
// `Save`（`0x1A290`）與 `Radio`（`0x15260`）走的是**同一套**：
// 印一句訊息 → `sub_19B4F` 等 Y／N → 答 N 就走人。只有訊息編號不同
// （Save 是執行檔字串 `0x49`、Radio 是第 1 條）。
//
// ⚠ **兩個指令要一起做。** 只接一邊的話另一邊仍然是「按下去就發生」，
// 而那正是玩家最容易誤觸的兩個——`Save` 會蓋掉存檔、`Radio` 會用掉表揚。

import (
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// 兩個確認的訊息編號（執行檔字串表 1，`docs/re/91` §1、§2）。
const (
	confirmSaveString  = 0x49
	confirmRadioString = 1
)

// confirmState 是「正在等 Y／N」的狀態。
type confirmState struct {
	active bool
	// then 是答 Y 之後要做的事。
	then func() (bool, error)
}

// askConfirm 印一句話並停下來等 Y／N。
func (s *Scene) askConfirm(stringN int, then func() (bool, error)) (bool, error) {
	s.say(stringN, textlayout.Options{})
	s.confirm = confirmState{active: true, then: then}
	s.dirty = true
	return true, nil
}

// askConfirmText 是 askConfirm 的「訊息自己準備好」版本。
//
// 有些提問是**兩條原版字串接起來**的（`ENC` 的 20 ＋ 76，`docs/re/105`），
// 沒辦法用單一編號表示。zh 為 nil 就顯示 en。
func (s *Scene) askConfirmText(en string, zh []byte, then func() (bool, error)) (bool, error) {
	s.message, s.cjk = en, zh
	if zh != nil {
		s.message = ""
	}
	s.confirm = confirmState{active: true, then: then}
	s.dirty = true
	return true, nil
}

// updateConfirm 收 Y／N。
//
// ⚠ **ESC 與 N 都是取消**（規格 27 §1：ESC 一律只取消）。
// 其餘按鍵一律吃掉不往下傳——確認畫面上按方向鍵不該走路。
func (s *Scene) updateConfirm(in input.Input) (bool, error) {
	then := s.confirm.then
	switch {
	case input.Upper(in.Char) == 'Y' || in.Action == input.ActionConfirm:
		s.confirm = confirmState{}
		if then != nil {
			return then()
		}
	case input.Upper(in.Char) == 'N' || in.Action == input.ActionCancel:
		s.confirm = confirmState{}
		s.message, s.cjk = "", nil
		s.dirty = true
	}
	return true, nil
}
