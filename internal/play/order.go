package play

// `ORDER` 指令：重排隊伍順序（`docs/re/93` §1）。
//
// 原版的做法是「整份抄到暫存區、原表清空、再逐一問下一個是誰」：
//
//	0x12AE9  隊伍槽表（8 格）→ ds:46CFh，原表填 0
//	0x12B15  sub_12B8A：把暫存區裡還沒放回去的人列出來
//	0x12B18  sub_1721B 挑一個 → 從暫存區移除 → 放進槽表的下一格
//	         排滿 ds:4653h（人數）就結束
//
// ⚠ **中途取消會留下空表**——原版沒有回頭路，所以這裡也照做：
// 進了 `ORDER` 就要排完。ESC 取消時把原本的順序整份放回去。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// orderState 是重排進行中的狀態。
type orderState struct {
	active bool
	// pending 是暫存區（`ds:46CFh`）：還沒放回去的人，nil 表示那一格已經放過。
	pending []*game.Character
	// placed 是已經排好的新順序。
	placed []*game.Character
	// backup 是進來之前的順序，取消時整份放回去。
	backup []*game.Character
}

// beginOrder 是按下 `O` 之後的第一步。
//
// **一個人不用排**（`0x12AE0` 的 `dec dl` ＝ 0 就 `retn`）。
func (s *Scene) beginOrder() {
	members := s.world.Party.Members
	n := 0
	for _, m := range members {
		if m != nil {
			n++
		}
	}
	if n <= 1 {
		s.sayEN("Nothing to reorder.", "order.nothing")
		s.dirty = true
		return
	}
	s.order = orderState{active: true}
	s.order.backup = append(s.order.backup, members...)
	for _, m := range members {
		if m != nil {
			s.order.pending = append(s.order.pending, m)
		}
	}
	s.message = s.orderPrompt()
	s.dirty = true
}

// orderPrompt 列出還沒放回去的人（原版 `sub_12B8A` 那一段）。
func (s *Scene) orderPrompt() string {
	var b strings.Builder
	for i, m := range s.order.pending {
		if m == nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%d %s", i+1, m.Name)
	}
	return fmt.Sprintf("Who is next? %s", b.String())
}

// updateOrder 是重排進行中的按鍵。
func (s *Scene) updateOrder(in input.Input) (bool, error) {
	if in.Action == input.ActionCancel {
		// 取消：把原本的順序整份放回去，不留半排好的狀態。
		copy(s.world.Party.Members, s.order.backup)
		s.order = orderState{}
		s.message = ""
		s.dirty = true
		return true, nil
	}
	ch := input.Upper(in.Char)
	if ch < '1' || ch > '9' {
		return true, nil
	}
	i := int(ch - '1')
	if i >= len(s.order.pending) || s.order.pending[i] == nil {
		return true, nil // 已經放過的格子選不了（原版 `jz` 回去重問）
	}
	s.order.placed = append(s.order.placed, s.order.pending[i])
	s.order.pending[i] = nil

	if len(s.order.placed) < len(s.order.backup)-s.nilCount() {
		s.message = s.orderPrompt()
		s.dirty = true
		return true, nil
	}
	// 排滿了：寫回隊伍，空格補 nil。
	for j := range s.world.Party.Members {
		if j < len(s.order.placed) {
			s.world.Party.Members[j] = s.order.placed[j]
		} else {
			s.world.Party.Members[j] = nil
		}
	}
	names := make([]string, 0, len(s.order.placed))
	for _, m := range s.order.placed {
		names = append(names, m.Name)
	}
	s.order = orderState{}
	s.message = "Order: " + strings.Join(names, ", ")
	s.dirty = true
	return true, nil
}

// nilCount 數備份裡有幾格是空的。
func (s *Scene) nilCount() int {
	n := 0
	for _, m := range s.order.backup {
		if m == nil {
			n++
		}
	}
	return n
}
