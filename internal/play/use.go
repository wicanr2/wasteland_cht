package play

// `USE` 指令的三層選單（`docs/re/92`）。
//
// 原版：挑人（一個人就不問）→ 印字串 4 等 `S`／`I`／`A` → 選具體的
// 技能／物品／屬性 → 拿它去試**腳下那一格**的條件閘串列。
//
// ⚠ 選單的呈現是**重製決策**：原版用清單框架（`sub_198F0`）＋ 上下鍵，
// 這裡用數字鍵。判定那一層照原版（`game.Party.UseGate`），沒有動。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// useStage 是三層選單走到哪一層。
type useStage int

const (
	useStageOff    useStage = iota
	useStageMember          // 挑人
	useStageKind            // S／I／A
	useStagePick            // 選技能／物品／屬性
	useStageDir             // **戰鬥限定**：問方向（`docs/re/108` §2）
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
	// combat 為真表示這一輪 `USE` 是**戰鬥裡的指令**，不是地圖上的動作。
	//
	// ⚠ 兩者差三件事：不用挑人（輪到誰就是誰）、選完之後要**再問一個方向**、
	// 而且**不當場施用**——把「選項＋方向」記成指令參數，動作在結算階段
	// （`docs/re/107` §1 的第二張跳表）。
	combat bool
	// pending 是戰鬥那條路選好、等著記進指令參數的那一項。
	pending useOption
	// page 是清單分到第幾頁（0 起算）。
	//
	// ⚠ **分頁是重製版的決定，不是原版行為**：原版清單選擇走 `sub_198F0`，
	// 那一支還沒逆向（`docs/re/92` §3），所以它超過九項時怎麼呈現是未知的。
	// 這裡不假裝知道——用數字鍵選、`0` 翻頁，都是重製版自己的介面。
	page int
}

// usePageSize 是一頁幾項。九是因為選項用數字鍵 `1`–`9`。
const usePageSize = 9

// usePages 是清單共幾頁（至少 1）。
func (u *useState) pages() int {
	if len(u.options) == 0 {
		return 1
	}
	return (len(u.options) + usePageSize - 1) / usePageSize
}

// pageSlice 是這一頁的選項。
func (u *useState) pageSlice() []useOption {
	lo := u.page * usePageSize
	if lo >= len(u.options) {
		return nil
	}
	hi := lo + usePageSize
	if hi > len(u.options) {
		hi = len(u.options)
	}
	return u.options[lo:hi]
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
		s.cjk = t + " " + s.memberMenu()
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
	s.use.page = 0
	s.showUseMenu()
	if len(s.cjk) > 0 {
		s.message = "" // 中文清單出來了就不要再疊一份英文
	}
	s.dirty = true
}

// useMenuCJK 是同一份清單的中文版（技能與物品名走 `exe:2:<條號>`）。
//
// 有一個名字沒翻就整份退回英文——**中英混在同一份清單裡最難讀**，
// 而且會讓「哪些還沒翻」看不出來。
func (s *Scene) useMenuCJK() string {
	if s.cat == nil {
		return ""
	}
	var out string
	for i, o := range s.use.pageSlice() {
		if o.nameSlot == 0 {
			return "" // 屬性那條沒有對應的原版字串
		}
		b, ok := s.cat.Lookup(lang.ExeKey(nameTable, o.nameSlot))
		if !ok {
			return ""
		}
		if len(out) > 0 {
			out += string(' ')
		}
		out += string(rune('1'+i)) + " " + singular(b)
	}
	// 翻頁那一行。⚠ **不要在這一行放阿拉伯數字的頁碼**：
	// 滑鼠的規則是「點到哪一格就送那一格的字元」（`docs/spec/29` §3），
	// 頁碼裡的 `2` 會被當成「選第 2 項」。頁碼用中文數字。
	if s.use.pages() > 1 {
		if more := s.uiText("use.morepage"); len(more) > 0 {
			out += "\r" + more + cjkPageLabel(s.use.page+1, s.use.pages())
		}
	}
	return out
}

// cjkPageLabel 是「（第二頁／共三頁）」這種**不含阿拉伯數字**的頁碼。
func cjkPageLabel(now, total int) string {
	num := func(n int) string {
		const digits = "〇一二三四五六七八九"
		r := []rune(digits)
		if n < 10 {
			return string(r[n])
		}
		if n < 20 {
			return "十" + map[bool]string{true: "", false: string(r[n%10])}[n%10 == 0]
		}
		return string(r[n/10]) + "十" + map[bool]string{true: "", false: string(r[n%10])}[n%10 == 0]
	}
	// UTF-8 之後不用先編碼——**畫的那一刻才轉 Big5**（`render.DrawRune`）。
	// 全形括號與「第／共／頁」都在倚天字庫裡，編譯期已經驗過。
	return "（第" + num(now) + "頁／共" + num(total) + "頁）"
}


// useMenu 是第三層的清單（英文後備）。一頁九項，多的翻頁。
func (s *Scene) useMenu() string {
	var b strings.Builder
	for i, o := range s.use.pageSlice() {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%d %s", i+1, o.label)
	}
	if s.use.pages() > 1 {
		fmt.Fprintf(&b, "\r0 More")
	}
	return b.String()
}

// singular 取名字的單數形 ＝ **字根 ＋ 單數字尾**。
//
// 中英兩條路共用這一支——**改成 UTF-8 之後 Big5 那個版本不需要了**：
// 以前得另外寫一份逐 byte 切的，還要靠「`0x0A` 不可能是 Big5 尾位元組」
// 這個前提才安全。
//
// 名字用 `\n` 分隔**字根／單數字尾／複數字尾**（`docs/re/17` §4.1）：
//
//	Crowbar\n\ns\n      → Crowbar / Crowbars
//	Kni\nfe\nves\n      → Knife   / Knives
//
// ⚠ **只取字根是錯的**：`Crowbar` 那一類的單數字尾是空的所以看不出來，
// 但 `Kni\nfe\nves\n` 會變成 `Kni`。原版名單畫面印的是 `Knife`
// （`docs/re/47` §5 的實機截圖）。
//
// ⚠ 整串丟進選單會換行、把後面的項目擠掉，所以複數那一段一定要去掉。
func singular(raw string) string {
	parts := strings.Split(raw, "\n")
	out := parts[0]
	if len(parts) > 1 {
		out += parts[1]
	}
	return strings.TrimSpace(out)
}

// skillName 查技能名（`ds:B270h` 表的第 1–35 條，`docs/re/17` §4）。
func (s *Scene) skillName(id byte) string {
	if n := singular(s.nameString(int(id))); n != "" {
		return n
	}
	return fmt.Sprintf("skill %d", id)
}

// skillNameCJK／itemNameCJK 是同兩個名稱的中文（Big5）。
//
// 名稱帶單複數分段（`字根\x0A單數\x0A複數\x0A`，`docs/re/17` §4.1），
// 交給 `RenderBytes` 依 Count ＝ 1 取單數——**不要自己找 `\x0A` 切**，
// 譯文的分段位置與原文不一定一樣。
func (s *Scene) skillNameCJK(id byte) string {
	return strings.TrimSpace(s.cjkExe(nameTable, int(id), textlayout.Options{Count: 1}))
}

func (s *Scene) itemNameCJK(id byte) string {
	return strings.TrimSpace(s.cjkExe(nameTable, int(id)+itemNameBase,
		textlayout.Options{Count: 1}))
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
	// 戰鬥那條路不當場施用：先問方向，再把「選項＋方向」記成指令參數。
	if s.use.combat {
		s.use.pending = o
		s.use.stage = useStageDir
		s.showUseDirection()
		return
	}
	m := s.world.Party.Members[s.use.member]
	kind := s.use.kind
	s.use = useState{}
	// 選單收起來，它的中文也要跟著收。留著的話結果那一行會疊在
	// 上一層的清單上——畫面上是「1 鬥毆 2 攀爬…」底下印著結果。
	s.cjk = ""

	// 名字那一欄中文優先（清單裡看到的是中文，結果那一句也該是）。
	zhLabel := o.label
	if o.nameSlot != 0 {
		if b := strings.TrimSpace(s.cjkExe(nameTable, o.nameSlot,
			textlayout.Options{Count: 1})); b != "" {
			zhLabel = b
		}
	}
	say := func(en string, uiName string) {
		s.message = fmt.Sprintf(en, m.Name, o.label)
		s.cjkFmt(uiName, m.Name, zhLabel)
	}

	rec, _, err := s.world.Block.CellRecord(int(s.world.Party.X), int(s.world.Party.Y))
	if err != nil || len(rec) == 0 {
		say("%s uses %s. Nothing happens.", "use.nothinghappens")
		s.dirty = true
		return
	}
	res := s.world.Party.UseOn(s.world.RNG, rec, m, kind, o.id, s.world.Skills)
	switch {
	case res.Hit < 0:
		say("%s uses %s. Nothing happens.", "use.nothinghappens")
	case res.Passed:
		say("%s uses %s. It works!", "use.works")
		s.playSound(7)
	default:
		say("%s uses %s. It fails.", "use.fails")
	}
	// **收尾要改寫腳下那一格**（`0x13D23`／`0x13D7C` 的 `sub_17CFF`）。
	// 這一步以前沒接：黑色圓柱插了黑星鑰匙會說 It works!，但下一根圓柱
	// 永遠不會出現，科奇斯基地的啟動序列就停在第一站（`docs/re/100` §3）。
	s.world.PatchHere(rec, res.PatchAt)
	s.dirty = true
}

// updateUse 是 `USE` 進行中的按鍵。
func (s *Scene) updateUse(in input.Input) (bool, error) {
	if in.Action == input.ActionCancel {
		s.use = useState{}
		s.message, s.cjk = "", ""
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
		// `0` 或上下鍵翻頁。**這是重製版的介面**——原版的清單選擇（`sub_198F0`）
		// 還沒逆向，超過九項時它怎麼做是未知的（`docs/re/92` §3）。
		if ch == '0' || in.Dir == input.DirUp || in.Dir == input.DirDown {
			if s.use.pages() > 1 {
				s.use.page = (s.use.page + 1) % s.use.pages()
				s.showUseMenu()
			}
			return true, nil
		}
		if ch >= '1' && ch <= '9' {
			// ⚠ 數字是**這一頁的第幾項**，不是整份清單的第幾項。
			if i := s.use.page*usePageSize + int(ch-'1'); i < len(s.use.options) {
				s.applyUse(s.use.options[i])
			}
		}
	case useStageDir:
		if d, ok := useDirection(ch, in.Dir); ok {
			s.commitCombatUse(d)
		}
	}
	return true, nil
}

// useDirection 把按鍵換成九向表的索引（`docs/re/108` §2）。
//
// 原版的字母表是 `I K J L ␣` ＋ 四個方向鍵的擴充掃描碼，
// 而 `0x1260E` 的 `al −= 5` **把方向鍵折進字母鍵那五格**——
// 兩套鍵是同一件事，不是八個方向。⚠ 表裡第 5–8 格是對角，但**選不到**。
func useDirection(ch byte, dir input.Direction) (byte, bool) {
	switch ch {
	case 'I':
		return useDirUp, true
	case 'K':
		return useDirDown, true
	case 'J':
		return useDirLeft, true
	case 'L':
		return useDirRight, true
	case ' ':
		return useDirStay, true
	}
	switch dir {
	case input.DirUp:
		return useDirUp, true
	case input.DirDown:
		return useDirDown, true
	case input.DirLeft:
		return useDirLeft, true
	case input.DirRight:
		return useDirRight, true
	}
	return 0, false
}

// 九向位移表 `ds:AAB1h` 的前五格（`docs/re/108` §2）。
const (
	useDirUp    byte = 0
	useDirDown  byte = 1
	useDirLeft  byte = 2
	useDirRight byte = 3
	useDirStay  byte = 4
)

// useDirDelta 是那五格各自的 (dx, dy)。
func useDirDelta(d byte) (int, int) {
	switch d {
	case useDirUp:
		return 0, -1
	case useDirDown:
		return 0, 1
	case useDirLeft:
		return -1, 0
	case useDirRight:
		return 1, 0
	}
	return 0, 0
}

// showUseDirection 問方向——原版印的是 `ds:A469h` 的 `Which way?`。
func (s *Scene) showUseDirection() {
	s.message = "Which way?  I K J L or arrows, space = here"
	s.cjk = s.uiText("use.whichway")
	if len(s.cjk) > 0 {
		s.message = ""
	}
	s.dirty = true
}

// commitCombatUse 把「選項 ＋ 方向」記成指令參數，編號另存一格
// （`docs/re/108` §1：`ds:46DAh ← (選項 << 4) | 方向`、
// `ds:A9FDh[角色編號] ← 編號`）。
func (s *Scene) commitCombatUse(dir byte) {
	c := s.combat
	o, kind, member := s.use.pending, s.use.kind, s.use.member
	s.use = useState{}
	s.cjk = ""
	if c == nil {
		return
	}
	c.SetUse(member, byte(kind), o.id, dir)
	s.showCombatPrompt()
	s.dirty = true
}

// showUseMenu 把目前這一頁的清單放上畫面。
func (s *Scene) showUseMenu() {
	s.message = s.useMenu()
	s.cjk = s.useMenuCJK()
	if len(s.cjk) > 0 {
		s.message = "" // 有中文就不要再印英文（兩份會疊在同一個視窗）
	}
	s.dirty = true
}
