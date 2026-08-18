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
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
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
	s.showOrderPrompt()
}

// orderPromptStr 是原版印的那一條：**字串 15**（`0x12B10` 的 `al ← 0x0F`）。
//
// ⚠ **不要自己寫一句。** 這個位置原版說的是 `Pick a player:`，目錄裡早就翻好；
// 自己寫的話那條原版字串就永遠不會出現在畫面上，而且中文化覆蓋率也量不到
// （`docs/re/105` §5 是同一條規則的第一個實例）。
const orderPromptStr = 15

// orderNames 是還沒放回去的那幾個人，「編號 ＋ 名字」以空白分隔。
// 名字不翻譯，所以英文與中文兩條路共用這一份。
func (s *Scene) orderNames() string {
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
	return b.String()
}

// showOrderPrompt 把字串 15 ＋ 名單放進訊息區，中英兩條路各一份。
func (s *Scene) showOrderPrompt() {
	names := s.orderNames()
	s.message = strings.TrimSpace(oneLine(s.exeString(orderPromptStr)) + " " + names)
	s.cjk = ""
	if zh := s.cjkExe(exeTable1, orderPromptStr, textlayout.Options{}); len(zh) > 0 {
		s.cjk = zh + string(' ') + names
		s.message = ""
	}
	s.dirty = true
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
		s.showOrderPrompt()
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
	// ⚠ **原版排完不印任何一句**（`0x12B41` 之後只有 `sub_17033` 重畫名片行）。
	// 這一句是重製版加的確認，所以走 `ui:` 前綴——**標出來哪些話不是原版的**。
	s.message = "Order: " + strings.Join(names, ", ")
	s.cjk = ""
	if zh := s.uiText("order.done"); len(zh) > 0 {
		// UTF-8 之後分隔符直接寫就好——**這裡以前要先 `lang.ToBig5("、")`**，
		// 漏掉一次的症狀是「Vargas粻 hrasher」這種讀得出筆畫卻不成字的東西。
		s.cjk = zh + strings.Join(names, "、")
		s.message = ""
	}
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
