package play

// `USE` 指令的三層選單（`docs/re/92`）。
//
// 原版：挑人（一個人就不問）→ 印字串 4 等 `S`／`I`／`A` → 選具體的
// 技能／物品／屬性 → 拿它去試**腳下那一格**的條件閘串列。
//
// ⚠ 選單的呈現是**重製決策**：原版用清單框架（`sub_198F0`）＋ 上下鍵，
// 這裡用數字鍵。判定那一層照原版（`game.Party.UseGate`），沒有動。

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/lang"
)

// useStage 是三層選單走到哪一層。
type useStage int

const (
	useStageOff    useStage = iota
	useStageMember          // 挑人
	useStageKind            // S／I／A
	useStagePick            // 選技能／物品／屬性
)

// nameTable 是技能與物品名那張表（`ds:B270h`）在 `ExeStrings()` 裡的編號。
const nameTable = 2

// useOption 是清單裡的一項：顯示用的名字，加上要交給 UseGate 的編號。
type useOption struct {
	label string
	id    byte
	// nameSlot 是這個名字在 `ds:B270h` 表裡的條號；0 ＝ 這一項不走那張表
	// （屬性名是 Go 這邊的常數）。中文版查 `exe:2:<nameSlot>`。
	nameSlot int
}

// useState 是 `USE` 進行中的狀態。
type useState struct {
	stage   useStage
	member  int // 隊伍索引
	kind    game.UseKind
	options []useOption
}

// beginUse 是按下 `U` 之後的第一步。
//
// **隊伍只有一個能行動的人就不問**（`0x13A89` 的 `cmp al, 1`）。
func (s *Scene) beginUse() {
	s.use = useState{stage: useStageMember}
	var only int = -1
	n := 0
	for i, m := range s.world.Party.Members {
		if m != nil && !m.Down() {
			n++
			only = i
		}
	}
	if n == 0 {
		s.use = useState{}
		s.message = "Nobody can do that."
		s.cjk = s.uiText("use.nobody")
		s.dirty = true
		return
	}
	if n == 1 {
		s.pickUseMember(only)
		return
	}
	s.message = "Which player? " + s.memberMenu()
	if t := s.uiText("use.which"); len(t) > 0 {
		s.cjk = append(append([]byte{}, t...), ' ')
		s.cjk = append(s.cjk, s.memberMenu()...)
		s.message = ""
	}
	s.dirty = true
}

// memberMenu 是隊員清單（1 起算，倒下的人標出來但選不了）。
func (s *Scene) memberMenu() string {
	var b strings.Builder
	for i, m := range s.world.Party.Members {
		if m == nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		if m.Down() {
			fmt.Fprintf(&b, "(%d %s)", i+1, m.Name)
			continue
		}
		fmt.Fprintf(&b, "%d %s", i+1, m.Name)
	}
	return b.String()
}

// pickUseMember 選定隊員，進第二層。
func (s *Scene) pickUseMember(i int) {
	s.use.member = i
	s.use.stage = useStageKind
	// 字串 4 的顯示順序是 Item／Skill／Attribute，而**索引照字母表 SIA**
	// （`docs/re/92` §2）——這裡顯示照原版、按鍵照字母表。
	s.message = "Use: Item / Skill / Attribute  (I/S/A, ESC cancel)"
	if t := s.uiText("use.kind"); len(t) > 0 {
		s.cjk = t
		s.message = ""
	}
	s.dirty = true
}

// pickUseKind 選定種類，把清單建起來。
func (s *Scene) pickUseKind(k game.UseKind) {
	m := s.world.Party.Members[s.use.member]
	s.use.kind = k
	s.use.options = s.use.options[:0]
	switch k {
	case game.UseSkill:
		for _, sk := range m.Skills {
			if sk.ID == 0 {
				continue
			}
			s.use.options = append(s.use.options, useOption{
				label: s.skillName(sk.ID), id: sk.ID, nameSlot: int(sk.ID)})
		}
	case game.UseItem:
		// UseGate 比的是**物品 ID**（`sub_14090` 先把槽號換成 ID）。
		for _, it := range m.Items {
			if it.ID == 0 {
				continue
			}
			s.use.options = append(s.use.options, useOption{
				label: s.itemName(it.ID), id: it.ID, nameSlot: int(it.ID) + itemNameBase})
		}
	case game.UseAttribute:
		// 屬性那條的參數是**角色記錄位移**（`docs/re/32` §4），不是屬性索引。
		for i, name := range game.AttributeNames {
			s.use.options = append(s.use.options, useOption{
				label: name, id: byte(0x0E + i)})
		}
	}
	if len(s.use.options) == 0 {
		s.use = useState{}
		s.message = "Nothing to use."
		s.cjk = s.uiText("use.nothing")
		s.dirty = true
		return
	}
	s.use.stage = useStagePick
	s.message = s.useMenu()
	s.cjk = s.useMenuCJK()
	if len(s.cjk) > 0 {
		s.message = "" // 中文清單出來了就不要再疊一份英文
	}
	s.dirty = true
}

// useMenuCJK 是同一份清單的中文版（技能與物品名走 `exe:2:<條號>`）。
//
// 有一個名字沒翻就整份退回英文——**中英混在同一份清單裡最難讀**，
// 而且會讓「哪些還沒翻」看不出來。
func (s *Scene) useMenuCJK() []byte {
	if s.cat == nil {
		return nil
	}
	var out []byte
	for i, o := range s.use.options {
		if i >= 9 {
			break
		}
		if o.nameSlot == 0 {
			return nil // 屬性那條沒有對應的原版字串
		}
		b, ok := s.cat.Lookup(lang.ExeKey(nameTable, o.nameSlot))
		if !ok {
			return nil
		}
		if len(out) > 0 {
			out = append(out, ' ')
		}
		out = append(out, byte('1'+i), ' ')
		out = append(out, singularBytes(b)...)
	}
	return out
}

// singularBytes 是 `singular` 的 byte 版：譯文同樣用 `\x0A` 分單複數
// （`docs/re/28`），清單只要第一段。
func singularBytes(raw []byte) []byte {
	if i := bytes.IndexByte(raw, '\n'); i >= 0 {
		raw = raw[:i]
	}
	return bytes.TrimSpace(raw)
}

// useMenu 是第三層的清單（最多九項一頁——**超過的部分還沒做分頁**）。
func (s *Scene) useMenu() string {
	var b strings.Builder
	for i, o := range s.use.options {
		if i >= 9 {
			b.WriteString(" …")
			break
		}
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%d %s", i+1, o.label)
	}
	return b.String()
}

// singular 取名字的第一段。
//
// ⚠ 名字用 `\n` 分隔**字根／單數字尾／複數字尾**（`docs/re/17` §4.1），
// 整串丟進選單會換行、把後面的項目擠掉。與敵人名同一套處理。
func singular(raw string) string {
	if i := strings.IndexByte(raw, '\n'); i >= 0 {
		raw = raw[:i]
	}
	return strings.TrimSpace(raw)
}

// skillName 查技能名（`ds:B270h` 表的第 1–35 條，`docs/re/17` §4）。
func (s *Scene) skillName(id byte) string {
	if n := singular(s.nameString(int(id))); n != "" {
		return n
	}
	return fmt.Sprintf("skill %d", id)
}

// itemName 查物品名。同一張表的 36–130 是物品段落。
func (s *Scene) itemName(id byte) string {
	if n := singular(s.nameString(int(id) + itemNameBase)); n != "" {
		return n
	}
	return fmt.Sprintf("item %d", id)
}

// itemNameBase 是物品名在 `ds:B270h` 表裡的偏移：索引 ＝ 物品 ID ＋ 36
// （技能佔了 1–35，物品 ID 從 **0** 起算）。
//
// 正對照：出廠角色的槽 0 是 ID 13 → `M1911A1 45 pistol`、
// 槽 1–8 是 ID 30 → `45 clip`，正好是 `docs/re/21` §5.1 的
// 「手槍 ＋ 八個彈匣」。偏移寫成 35 會整批位移一格
// （手槍變 `RPG-7`、彈匣變 `Meson cannon`），而**畫面上看起來完全正常**。
const itemNameBase = 36

// applyUse 拿選中的東西去試腳下那一格的條件閘。
func (s *Scene) applyUse(o useOption) {
	m := s.world.Party.Members[s.use.member]
	kind := s.use.kind
	s.use = useState{}
	// 選單收起來，它的中文也要跟著收。留著的話結果那一行會疊在
	// 上一層的清單上——畫面上是「1 鬥毆 2 攀爬…」底下印著結果。
	s.cjk = nil

	rec, _, err := s.world.Block.CellRecord(int(s.world.Party.X), int(s.world.Party.Y))
	if err != nil || len(rec) == 0 {
		s.message = fmt.Sprintf("%s uses %s. Nothing happens.", m.Name, o.label)
		s.dirty = true
		return
	}
	hit, passed := s.world.Party.UseGate(s.world.RNG, rec, m, kind, o.id, s.world.Skills)
	switch {
	case hit < 0:
		// 沒有吻合的條件——原版就是什麼都不發生。
		s.message = fmt.Sprintf("%s uses %s. Nothing happens.", m.Name, o.label)
	case passed:
		s.message = fmt.Sprintf("%s uses %s. It works!", m.Name, o.label)
		s.playSound(7)
	default:
		s.message = fmt.Sprintf("%s uses %s. It fails.", m.Name, o.label)
	}
	s.dirty = true
}

// updateUse 是 `USE` 進行中的按鍵。
func (s *Scene) updateUse(in input.Input) (bool, error) {
	if in.Action == input.ActionCancel {
		s.use = useState{}
		s.message, s.cjk = "", nil
		s.dirty = true
		return true, nil
	}
	ch := input.Upper(in.Char)
	switch s.use.stage {
	case useStageMember:
		if ch >= '1' && ch <= '9' {
			i := int(ch - '1')
			if i < len(s.world.Party.Members) {
				if m := s.world.Party.Members[i]; m != nil && !m.Down() {
					s.pickUseMember(i)
				}
			}
		}
	case useStageKind:
		switch ch {
		case 'S':
			s.pickUseKind(game.UseSkill)
		case 'I':
			s.pickUseKind(game.UseItem)
		case 'A':
			s.pickUseKind(game.UseAttribute)
		}
	case useStagePick:
		if ch >= '1' && ch <= '9' {
			if i := int(ch - '1'); i < len(s.use.options) {
				s.applyUse(s.use.options[i])
			}
		}
	}
	return true, nil
}
