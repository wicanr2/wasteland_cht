package play

// 角色畫面（docs/re/131）：地圖上按 `1`–`7`（或滑鼠點名片行）開那個人的
// 資料頁／物品頁／技能頁；物品頁能裝備、裝填、卸卡彈、丟棄、交給隊友。
//
// ⚠ 呈現是重製決策：原版是名單模式＋清單框架（`sub_19130`），這裡照
// `USE` 的慣例走訊息視窗＋數字鍵（`docs/re/92` §7 同一條）；
// Enter／空白循環三頁、ESC 一律直接關（原版是物品頁 ESC → 技能頁 → 離開）。
// 物品重排（原版 0xFE 那條）沒有做。**判定那一層照原版**：
// 裝備 `sub_1949E`、裝填 `sub_196DB`、卸卡彈 `sub_19ACD`、交給 `0x19299`。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// 角色畫面用到的原版字串編號（都在表 2，docs/re/131 §8）。
const (
	strCVUnjam    = 134 // `Unjam?  Yes / No`
	strCVNoTrade  = 135 // `\x0B doesn't want to trade.`
	strCVName     = 137 // `NAME: \x0B`
	strCVSex      = 138 // `SEX: `
	strCVNation   = 139 // `NATIONALITY: `
	strCVPoolDiv  = 140 // `P)ool  D)iv cash`
	strCVReload   = 141 // `Reload?  Yes / No`
	strCVDoWhat   = 142 // `Do you want to:  Drop  Trade  Equip/unequip`
	strCVNobody   = 143 // `No one to trade with.`
	strCVWhoWants = 144 // `Who wants it?`
	strCVNationT  = 145 // 145–149 ＝ U.S…Chinese；性別 162／163
	strCVFull     = 150 // `\x0B can't carry any more.`
	strCVItemHdr  = 151 // `    ITEM`
	strCVSkillHdr = 153 // `   LVL   SKILL`
	strCVGenderT  = 162
	strCVUnjamOK  = 160 // `\x0B unjammed the weapon.`
	strCVUnjamNo  = 161 // `\x0B can't unjam the weapon.`
)

// cvPageSize 是物品頁一頁幾項：訊息視窗只有六行，表頭佔一行。
// 原版的清單框架有自己的視窗（一頁 0x12 列），這裡照重製版的視窗縮。
const cvPageSize = 5

// charViewStage 是角色畫面走到哪一層。
type charViewStage int

const (
	cvOff charViewStage = iota
	cvInfo
	cvItems
	cvSkills
	cvAction // 選了一件物品，等 D／T／E（或 Reload?／Unjam? 的 Y／N）
	cvTradeWho
)

// charView 是角色畫面的狀態。
type charView struct {
	stage  charViewStage
	member int
	page   int // 物品頁的分頁（0 起算，一頁九項）
	sel    int // 選中的物品槽（cvAction／cvTradeWho 用）
	// ask 是 cvAction 停在哪個問句。
	ask cvAsk
}

// beginCharView 開第 i 個人的角色畫面（原版 sub_12760 → sub_19130）。
func (s *Scene) beginCharView(i int) {
	if i < 0 || i >= len(s.world.Party.Members) || s.world.Party.Members[i] == nil {
		return
	}
	s.rosterOn = false // 名單蓋著訊息視窗，角色畫面要用那一塊
	s.charView = charView{stage: cvInfo, member: i, sel: -1}
	s.showCVInfo()
}

// cvChar 是畫面上這個人。
func (s *Scene) cvChar() *game.Character { return s.world.Party.Members[s.charView.member] }

// plainExe 把原版字串裡的控制碼去掉（\x0D 換成換行、\x0B 換成名字）。
func plainExe(t, name string) string {
	var b strings.Builder
	for i := 0; i < len(t); i++ {
		switch c := t[i]; {
		case c == 0x0D:
			b.WriteByte('\r')
		case c == 0x0B:
			b.WriteString(name)
		case c >= 0x20:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// exePair 取一條表 2 字串的英文與中文（都已去控制碼、代入名字）。
func (s *Scene) exePair(n int, name string) (en, zh string) {
	en = plainExe(s.exeStringN(nameTable, n), name)
	zh = plainExe(s.cjkExe(nameTable, n,
		textlayout.Options{Name: func() string { return name }}), name)
	return en, zh
}

// showCVInfo 畫資料頁：姓名、性別、國籍、階級／等級／經驗、金錢、屬性。
func (s *Scene) showCVInfo() {
	c := s.cvChar()
	sexN := strCVGenderT + int(c.Gender)   // 162／163
	natN := strCVNationT + int(c.Nation)   // 145–149
	var en, zh strings.Builder
	build := func(b *strings.Builder, pick func(n int) string, statsFmt string) {
		// 137 自帶 `\x0B`（名字佔位），plainExe 已代入，不要再接一次。
		fmt.Fprintf(b, "%s\r", pick(strCVName))
		fmt.Fprintf(b, "%s%s  %s%s\r", pick(strCVSex), pick(sexN), pick(strCVNation), pick(natN))
		fmt.Fprintf(b, "%s Lv%d EXP %d $%d\r", c.Rank, c.Level, c.XP, c.Money)
		if statsFmt != "" {
			a := c.Attributes
			fmt.Fprintf(b, statsFmt, a[0], a[1], a[2], a[3], a[4], a[5], a[6], c.MaxCON)
			b.WriteByte('\r')
		}
		b.WriteString(pick(strCVPoolDiv))
	}
	build(&en, func(n int) string { e, _ := s.exePair(n, c.Name); return e },
		"ST %d IQ %d LK %d SP %d AG %d DX %d CH %d CON %d")
	zhOK := true
	build(&zh, func(n int) string {
		_, z := s.exePair(n, c.Name)
		if z == "" {
			zhOK = false
		}
		return z
	}, s.uiText("roster.stats"))
	s.message, s.cjk = en.String(), ""
	if zhOK && s.eten != nil {
		s.message, s.cjk = "", zh.String()
	}
	s.dirty = true
}

// cvItemRows 是物品清單（槽號照原樣，空槽跳過）。
func (s *Scene) cvItemRows() []int {
	c := s.cvChar()
	var out []int
	for i, sl := range c.Items {
		if sl.ID != 0 {
			out = append(out, i)
		}
	}
	return out
}

// showCVItems 畫物品頁：一頁九項、`0`／`I`／`K` 翻頁，裝備中標 `*`。
func (s *Scene) showCVItems() {
	c := s.cvChar()
	rows := s.cvItemRows()
	pages := (len(rows) + cvPageSize - 1) / cvPageSize
	if pages < 1 {
		pages = 1
	}
	if s.charView.page >= pages {
		s.charView.page = pages - 1
	}
	lo := s.charView.page * cvPageSize
	hi := lo + cvPageSize
	if hi > len(rows) {
		hi = len(rows)
	}
	enHdr, zhHdr := s.exePair(strCVItemHdr, c.Name)
	var en, zh strings.Builder
	en.WriteString(strings.TrimSpace(enHdr))
	zh.WriteString(strings.TrimSpace(zhHdr))
	if pages > 1 {
		fmt.Fprintf(&en, " %d/%d (0)", s.charView.page+1, pages)
		fmt.Fprintf(&zh, " %d/%d (0)", s.charView.page+1, pages)
	}
	en.WriteByte('\r')
	zh.WriteByte('\r')
	for n, slot := range rows[lo:hi] {
		// 原版的清單只有名字沒有數量欄（表頭 151 就只有 `ITEM`）；
		// `*` 是重製版加的裝備中記號。
		mark := ' '
		if slotIndexOf(c.EquipIndex) == slot || slotIndexOf(c.ArmorIndex) == slot {
			mark = '*'
		}
		fmt.Fprintf(&en, "%d)%c%s\r", n+1, mark, s.itemName(c.Items[slot].ID))
		name := s.itemNameCJK(c.Items[slot].ID)
		if name == "" {
			name = s.itemName(c.Items[slot].ID)
		}
		fmt.Fprintf(&zh, "%d)%c%s\r", n+1, mark, name)
	}
	s.message, s.cjk = en.String(), ""
	if s.eten != nil {
		s.message, s.cjk = "", zh.String()
	}
	s.dirty = true
}

// slotIndexOf 把裝備索引（原版存的是 `shl 1 ＋ 0xBC` 之前的值）換成槽號。
// 0 ＝ 沒裝備 → 回 −1。
func slotIndexOf(idx byte) int {
	if idx == 0 {
		return -1
	}
	return int(idx) - 1
}

// showCVSkills 畫技能頁。
func (s *Scene) showCVSkills() {
	c := s.cvChar()
	enHdr, zhHdr := s.exePair(strCVSkillHdr, c.Name)
	var en, zh strings.Builder
	en.WriteString(strings.TrimSpace(enHdr) + "\r")
	zh.WriteString(strings.TrimSpace(zhHdr) + "\r")
	for _, sl := range c.Skills {
		if sl.ID == 0 {
			continue
		}
		fmt.Fprintf(&en, "%d %s\r", sl.Value, s.nameString(int(sl.ID)))
		name := strings.TrimSpace(s.cjkExe(nameTable, int(sl.ID), textlayout.Options{Count: 1}))
		if name == "" {
			name = s.nameString(int(sl.ID))
		}
		fmt.Fprintf(&zh, "%d %s\r", sl.Value, name)
	}
	s.message, s.cjk = en.String(), ""
	if s.eten != nil {
		s.message, s.cjk = "", zh.String()
	}
	s.dirty = true
}

// cvAsk 是 cvAction 的三段問句（原版順序檢查：彈匣 → 卡彈 → D/T/E 選單；
// 答 No 就落到下一段，docs/re/131 §3.1）。
type cvAsk int

const (
	askReload cvAsk = iota
	askUnjam
	askMenu
)

// pickCVItem 選中一件物品，從第一段問句開始。
func (s *Scene) pickCVItem(slot int) {
	s.charView.sel = slot
	s.charView.stage = cvAction
	s.cvAskFrom(askReload)
}

// cvAskFrom 從 step 起找下一個成立的問句。
func (s *Scene) cvAskFrom(step cvAsk) {
	c := s.cvChar()
	slot := s.charView.sel
	if step <= askReload {
		// 選中的是裝備武器吃的彈匣 → Reload?
		if w := slotIndexOf(c.EquipIndex); w >= 0 && w < len(c.Items) {
			if d, ok := s.items.Get(c.Items[w].ID); ok && d.Ammo != 0 &&
				c.Items[slot].ID == d.Ammo {
				s.charView.ask = askReload
				s.sayT(nameTable, strCVReload, textlayout.Options{})
				return
			}
		}
	}
	if step <= askUnjam && game.Jammed(c.Items[slot]) {
		s.charView.ask = askUnjam
		s.sayT(nameTable, strCVUnjam, textlayout.Options{})
		return
	}
	s.charView.ask = askMenu
	s.sayT(nameTable, strCVDoWhat, textlayout.Options{})
}

// updateCharView 是角色畫面的按鍵。
func (s *Scene) updateCharView(in input.Input) (bool, error) {
	cv := &s.charView
	c := s.cvChar()
	ch := input.Upper(in.Char)
	enter := in.Action == input.ActionConfirm || ch == '\r' || ch == ' '

	if in.Action == input.ActionCancel {
		switch cv.stage {
		case cvAction, cvTradeWho:
			cv.stage = cvItems
			s.showCVItems()
		default:
			s.charView = charView{}
			s.message, s.cjk = "", ""
			s.dirty = true
		}
		return true, nil
	}

	switch cv.stage {
	case cvInfo:
		switch {
		case enter:
			cv.stage = cvItems
			s.showCVItems()
		case ch == 'P':
			game.PoolMoney(s.world.Party, cv.member)
			s.showCVInfo()
		case ch == 'D':
			game.DivideCash(s.world.Party, cv.member)
			s.showCVInfo()
		}
	case cvItems:
		rows := s.cvItemRows()
		pages := (len(rows) + cvPageSize - 1) / cvPageSize
		switch {
		case enter:
			cv.stage = cvSkills
			s.showCVSkills()
		case ch == '0' && pages > 1:
			cv.page = (cv.page + 1) % pages
			s.showCVItems()
		case (ch == 'I' || in.Dir == input.DirUp) && cv.page > 0:
			cv.page--
			s.showCVItems()
		case (ch == 'K' || in.Dir == input.DirDown) && cv.page+1 < pages:
			cv.page++
			s.showCVItems()
		case ch >= '1' && ch <= '9':
			// 數字是**這一頁的第幾項**（與 USE 的清單同一條規則）。
			if n := int(ch - '1'); n < cvPageSize {
				if i := cv.page*cvPageSize + n; i < len(rows) {
					s.pickCVItem(rows[i])
				}
			}
		}
	case cvSkills:
		if enter {
			cv.stage = cvInfo
			s.showCVInfo()
		}
	case cvAction:
		s.updateCVAction(ch, c)
	case cvTradeWho:
		if ch >= '1' && ch <= '9' {
			i := int(ch - '1')
			if i != cv.member && i < len(s.world.Party.Members) &&
				s.world.Party.Members[i] != nil {
				s.cvTrade(i)
			}
		}
	}
	return true, nil
}

// updateCVAction 是「選了一件物品之後」的按鍵。
func (s *Scene) updateCVAction(ch byte, c *game.Character) {
	cv := &s.charView
	switch cv.ask {
	case askReload:
		switch ch {
		case 'Y':
			r := c.ReloadFrom(cv.sel, s.items)
			s.sayResolve(c.Name, r)
			cv.stage = cvItems
			s.dirty = true
		case 'N':
			s.cvAskFrom(askUnjam) // 落到下一段（卡彈／選單）
		}
		return
	case askUnjam:
		switch ch {
		case 'Y':
			if c.Unjam(s.world.RNG, cv.sel, s.items) == game.UnjamOK {
				s.sayT2Name(strCVUnjamOK, c.Name)
			} else {
				s.sayT2Name(strCVUnjamNo, c.Name)
			}
			cv.stage = cvItems
			s.dirty = true
		case 'N':
			s.cvAskFrom(askMenu)
		}
		return
	}
	switch ch {
	case 'D': // 丟棄：那一槽清 0（原版不動裝備索引與 AC，docs/re/131 §3.1）
		c.Items[cv.sel] = game.Slot{}
		cv.stage = cvItems
		s.showCVItems()
	case 'E': // 裝備／卸下（sub_1949E：清單列號 ＝ 槽號 ＋1）
		it, _ := s.items.Get(c.Items[cv.sel].ID)
		c.Equip(cv.sel+1, it)
		cv.stage = cvItems
		s.showCVItems()
	case 'T': // 交給隊友（0x19299）
		n := 0
		for _, m := range s.world.Party.Members {
			if m != nil {
				n++
			}
		}
		switch {
		case n <= 1:
			s.sayT(nameTable, strCVNobody, textlayout.Options{})
			cv.stage = cvItems
		case n == 2:
			// 兩個人不用問：對方 ＝ 自己 xor 3（1↔2，原版 0x192AA）。
			s.cvTrade(((cv.member + 1) ^ 3) - 1)
		default:
			cv.stage = cvTradeWho
			s.message = "Who wants it? " + s.memberMenu()
			if zh := s.cjkExe(nameTable, strCVWhoWants, textlayout.Options{}); zh != "" {
				s.cjk = plainExe(zh, "") + " " + s.memberMenu()
				s.message = ""
			}
			s.dirty = true
		}
	}
}

// cvTrade 把選中那一件交給第 to 個人（含拒絕檢定，docs/re/131 §7）。
func (s *Scene) cvTrade(to int) {
	cv := &s.charView
	giver, receiver := s.cvChar(), s.world.Party.Members[to]
	if game.TradeNeedsCheck(giver) &&
		game.TradeRefused(s.world.RNG, giver, receiver) {
		s.sayT2Name(strCVNoTrade, giver.Name)
	} else if !game.TradeItem(giver, receiver, cv.sel) {
		s.sayT2Name(strCVFull, receiver.Name)
	} else {
		cv.stage = cvItems
		s.showCVItems()
		return
	}
	cv.stage = cvItems
	s.dirty = true
}

// sayT2Name 印一條表 2 的字串並代入名字。
func (s *Scene) sayT2Name(n int, name string) { s.sayT2NameAt(nameTable, n, name) }

// sayResolve 把 ResolveResult 印出來（與戰鬥的 resolveText 同一組字串）。
func (s *Scene) sayResolve(name string, r game.ResolveResult) {
	if r.Message == 0 {
		return
	}
	table := exeTable1
	if r.Table2 {
		table = nameTable
	}
	s.sayT2NameAt(table, int(r.Message), name)
}

// sayT2NameAt 同 sayT2Name 但可指定表別。
func (s *Scene) sayT2NameAt(table, n int, name string) {
	s.message = plainExe(s.exeStringN(table, n), name)
	if zh := s.cjkExe(table, n,
		textlayout.Options{Name: func() string { return name }}); zh != "" {
		s.cjk, s.message = plainExe(zh, name), ""
	}
	s.dirty = true
}
