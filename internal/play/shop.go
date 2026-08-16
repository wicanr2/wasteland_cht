package play

// 設施的選單互動（docs/spec/25、docs/re/42）。
//
// 規格 18 是規則、規格 23 是場景，這個檔是中間的互動：
// 按哪個鍵、走到哪一步。**規則一條都不在這裡重做。**

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// 設施訊息的原版字串表與編號（`docs/re/17` §3）。
// ⚠ **商店在表 7、醫生在表 8、訓練在表 6**——拿表 1 去查會安靜地取到別的句子。
const (
	exeTableShop    = 7
	exeTableDoctor  = 8
	exeTableTrainer = 6

	strOutOfStock   = 1  // 7：`We are temporarily out of stock.`
	strYouHave      = 4  // 7：`You have $`
	strNothingWant  = 5  // 7：`You don't have anything they want!`
	strBuyOrSell    = 6  // 7：`Do you want to:` ＋ Buy／Sell
	strPriceItem    = 7  // 7：`   PRICE     ITEM`
	strInvFull      = 8  // 7：`Your inventory is full.`
	strSkillPoints  = 4  // 6：`Skill points = `
	strNoSkillPts   = 7  // 6：`Not enough skill points!`
	strDiseaseFirst = 1  // 8：第 1–8 條 ＝ 八個狀態位元的病名
	strExamPrice    = 11 // 8：`Exam $`
	strHealing      = 12 // 8：`Healing`
	strCuring       = 13 // 8：`Curing`
	strNeedToHeal   = 15 // 8：`You need to heal `
	strPointsYouCan = 16 // 8：` points. You can:` ＋ Heal／Continue
	strNoDiseases   = 17 // 8：`You have no diseases.`
	strNoMoney      = 18 // 8：`You don't have enough money.`
	strWhichCure    = 20 // 8：`Which to cure at $`
	strCureTail     = 21 // 8：`:` ＋ 清單開頭
)

// FacilityStep 是設施選單的一個狀態。
type FacilityStep int

const (
	StepMain FacilityStep = iota // 主選單
	StepBuy                      // 買的清單
	StepSell                     // 賣的清單
	StepHeal                     // 醫生：逐點治療
	StepCure                     // 醫生：選一種病
)

// 商店與醫生的按鍵（docs/re/42 §1、§5）。原版比對的是靜態字母表，
// 不跟著翻譯走（docs/re/40 §4）——所以這裡也寫死。
const (
	keyBuy      = 'B'
	keySell     = 'S'
	keyNextChar = 'P'
	keyHeal     = 'H' // 醫生主選單的 Healing，也是治療迴圈的「Heal 1 point」
	keyContinue = 'C' // 治療迴圈的 Continue
	keyCure     = 'C' // 醫生主選單的 Curing（與 Continue 同一個字母，但在不同層）
	keyExam     = 'E' // 醫生主選單的 Exam
	keyPrevPage = 'I' // 上一頁（docs/re/53 §3）
	keyNextPage = 'K' // 下一頁
	keyEscape   = 0x1B
)

// shopState 是設施選單的狀態。放在 FacilityScene 裡，
// 離開設施就整個丟掉——原版也沒有跨場次保留的東西。
type shopState struct {
	Step FacilityStep
	Who  int // 目前在櫃檯的隊伍成員索引
	Page int // 目前這一頁的起始列（docs/re/53 §4）
}

// PageRows 是一頁最多列幾件。
//
// ⚠ **原版沒有把它當成清單框架的常數**：`ds:469Eh` 全檔只有醫生的疾病表
// 設過（值 9），其餘清單畫到索引上限為止（docs/re/53 §2）。
// 這裡統一用 9 是因為**選擇鍵只有 `'1'`–`'9'`**——那才是真正的限制。
const PageRows = 9

// Key 送一個按鍵給設施選單。回傳 false 表示要離開設施。
//
// ⚠ **「背包滿了」與「沒有東西賣」都回主迴圈，不是離開**（docs/spec/25 §2）——
// 那是原版的正常路徑，不是錯誤分支。
func (f *FacilityScene) Key(k byte, p *game.Party, items game.ItemTable) bool {
	if f.state == nil {
		f.state = &shopState{}
	}
	st := f.state
	if k == keyEscape {
		if st.Step != StepMain {
			st.Step = StepMain // 清單 → 主迴圈
			f.refresh(p, items)
			return true
		}
		return false // 主迴圈 → 離開設施
	}

	// 翻頁對每一種清單都一樣（docs/re/53 §4）。
	if st.Step != StepMain {
		switch k {
		case keyPrevPage:
			if st.Page -= PageRows; st.Page < 0 {
				st.Page = 0
			}
			f.refresh(p, items)
			return true
		case keyNextPage:
			if st.Page+PageRows < f.rowCount(p, items) {
				st.Page += PageRows
			}
			f.refresh(p, items)
			return true
		}
	}

	switch f.Facility.Kind {
	case game.FacilityShop:
		f.shopKey(k, p, items)
	case game.FacilityDoctor:
		f.doctorKey(k, p)
	case game.FacilityTrainer:
		f.trainerKey(k, p)
	default:
		// 訓練師與其餘兩支還沒解流程（docs/re/42 §7）——
		// 任何鍵都不做事，ESC 才離開。**不要發明選單。**
	}
	f.refresh(p, items)
	return true
}

func (f *FacilityScene) shopKey(k byte, p *game.Party, items game.ItemTable) {
	st := f.state
	switch k {
	case keyNextChar:
		st.Who = nextAble(p, st.Who)
	case keyBuy:
		if _, ok := game.FirstEmptyItemSlot(f.member(p).Items); !ok {
			f.setNote("Your inventory is full.", exeTableShop, strInvFull)
			st.Step = StepMain
			return
		}
		if len(f.buyList(items)) == 0 {
			// 全部缺貨。原版的 sub_1C140 回 0 之後也是印字串 8 回主迴圈。
			f.setNote("We are temporarily out of stock.", exeTableShop, strOutOfStock)
			st.Step = StepMain
			return
		}
		st.Step, st.Page = StepBuy, 0
	case keySell:
		if len(f.sellable(p, items)) == 0 {
			f.setNote("You don't have anything they want!", exeTableShop, strNothingWant)
			st.Step = StepMain
			return
		}
		st.Step, st.Page = StepSell, 0
	default:
		if k >= '1' && k <= '9' {
			switch st.Step {
			case StepSell:
				f.sellOne(p, items, st.Page+int(k-'1'))
			case StepBuy:
				f.buyOne(p, items, st.Page+int(k-'1'))
			}
		}
	}
}

// doctorKey 是醫生（docs/re/42 §5、§6）。
//
// ⚠ **主選單與治療迴圈是兩層。** 主選單的三個選項來自字串表 8 的
// 第 11／12／13 條（`Exam $` ／ `Healing` ／ `Curing`，各帶一個 `\x10`
// 熱鍵標記），所以熱鍵是 E／H／C；進了 Healing 之後才是
// 「Heal 1 point ／ Continue」（第 16 條），那一層的 H 與 C 意思不同。
// 兩層寫成一層的話，「C」會在錯的地方被吃掉。
func (f *FacilityScene) doctorKey(k byte, p *game.Party) {
	st := f.state
	if k == keyNextChar {
		st.Who = nextAble(p, st.Who)
		return
	}
	switch st.Step {
	case StepHeal:
		switch k {
		case keyContinue:
			st.Step = StepMain
		case keyHeal:
			h := game.HealSession{Facility: f.Facility, Char: f.member(p)}
			if ok, reason := h.HealOne(); !ok {
				f.setNoteReason(reason)
			} else {
				f.note, f.noteCJK = "", nil
			}
		}
	case StepCure:
		if k >= '1' && k <= '9' {
			f.cureOne(p, f.state.Page+int(k-'1'))
		}
	default: // StepMain
		switch k {
		case keyExam:
			ok, reason := f.Facility.Exam(f.member(p))
			if !ok {
				f.setNoteReason(reason)
			} else {
				f.setNoteUI("Examined.", "facility.examined")
			}
		case keyHeal:
			st.Step = StepHeal
		case keyCure:
			if len(game.Diseases(f.member(p))) == 0 {
				// 沒有病就不開選單（docs/re/42 §6），留在主選單。
				f.setNote("You have no diseases.", exeTableDoctor, strNoDiseases)
				return
			}
			st.Step = StepCure
		}
	}
}

// cureOne 治好清單上的第 n 種病。
func (f *FacilityScene) cureOne(p *game.Party, n int) {
	c := f.member(p)
	list := game.Diseases(c)
	if n < 0 || n >= len(list) {
		return
	}
	if ok, reason := f.Facility.Cure(c, list[n]); !ok {
		f.setNoteReason(reason)
		return
	}
	f.setNoteUI("Cured.", "facility.cured")
	if len(game.Diseases(c)) == 0 {
		f.state.Step = StepMain // 病治完了就回主選單
	}
}

// rowCount 是目前這一種清單總共有幾列（翻頁上限用）。
func (f *FacilityScene) rowCount(p *game.Party, items game.ItemTable) int {
	switch f.state.Step {
	case StepBuy:
		return len(f.buyList(items))
	case StepSell:
		return len(f.sellable(p, items))
	case StepCure:
		return len(game.Diseases(f.member(p)))
	}
	if f.Facility.Kind == game.FacilityTrainer {
		return len(f.Skills)
	}
	return 0
}

// page 把一整份清單切出目前這一頁。
//
// ⚠ 回傳的是**切片**，所以呼叫端拿到的索引是頁內列號；
// 真正的索引要加回 `state.Page`（docs/re/53 §7：列與索引是兩件事）。
func (f *FacilityScene) page(n int) (from, to int) {
	from = f.state.Page
	if from > n {
		from = 0
	}
	to = from + PageRows
	if to > n {
		to = n
	}
	return from, to
}

// buyList 是這家店現在賣得出來的東西（含折價後的價格）。
func (f *FacilityScene) buyList(items game.ItemTable) []game.BuyListEntry {
	return game.BuyList(items,
		func(base uint16) uint16 { return game.ShopPrice(base, f.Facility.Record[0x03]) },
		func(id byte) byte {
			it, ok := items.Get(id)
			if !ok {
				return game.StockNone
			}
			return it.Stock // 庫存就在表上，賣過的已經改進去了
		})
}

// buyOne 買下清單上的第 n 件。
func (f *FacilityScene) buyOne(p *game.Party, items game.ItemTable, n int) {
	list := f.buyList(items)
	if n < 0 || n >= len(list) {
		return
	}
	e := list[n]
	stock, ok := game.Buy(f.member(p), e.ID, uint32(e.Price), e.Stock)
	if !ok {
		// 錢不夠或背包滿了——**不是錯誤，是正常路徑**（docs/spec/25 §2）。
		f.setNote("You don't have enough money.", exeTableDoctor, strNoMoney)
		return
	}
	f.setStock(e.ID, stock)
	f.setNoteUI(fmt.Sprintf("Bought for $%d.", e.Price), "facility.bought", e.Price)
}

// trainerKey 是訓練師（docs/re/52 §2）。
//
// ⚠ **三條「走不通」都回到選人，不是離開設施**：不能行動、技能點數 0、
// 沒有可學的。寫成錯誤分支會讓玩家莫名其妙被踢出去。
func (f *FacilityScene) trainerKey(k byte, p *game.Party) {
	st := f.state
	switch {
	case k == keyNextChar:
		st.Who = nextAble(p, st.Who)
	case k >= '1' && k <= '9':
		f.learnOne(p, f.state.Page+int(k-'1'))
	}
}

// learnOne 學清單上的第 n 個技能。
func (f *FacilityScene) learnOne(p *game.Party, n int) {
	c := f.member(p)
	if c == nil {
		return
	}
	if c.SkillPts == 0 {
		f.setNote("You have no skill points.", exeTableTrainer, strNoSkillPts)
		return
	}
	list := f.Skills
	if n < 0 || n >= len(list) {
		return
	}
	ok, reason := c.LearnSkill(list[n].ID, list[n].Data)
	if !ok {
		f.setNoteReason(reason)
		return
	}
	name := f.skillLabel(list[n].ID)
	f.setNoteUI(fmt.Sprintf("Learned %s.", name), "facility.learned", name)
}

// sellable 是這個人身上賣得掉的東西。
func (f *FacilityScene) sellable(p *game.Party, items game.ItemTable) []game.SellListEntry {
	c := f.member(p)
	if c == nil {
		return nil
	}
	return game.SellList(c.Items,
		func(id byte) bool {
			// 庫存 0 的品項照樣賣得掉——「賣得掉」看的是物品本身，
			// 不是這家店現在有沒有貨（docs/re/42 §4）。
			_, ok := items.Get(id)
			return ok
		},
		func(slot int) bool {
			return slot == int(c.EquipIndex) || slot == int(c.ArmorIndex)
		})
}

// sellOne 賣掉清單上的第 n 件。
func (f *FacilityScene) sellOne(p *game.Party, items game.ItemTable, n int) {
	list := f.sellable(p, items)
	if n < 0 || n >= len(list) {
		return
	}
	c := f.member(p)
	e := list[n]
	item, ok := items.Get(e.Item)
	if !ok {
		return
	}
	price := f.Facility.SellPrice(item.Price)
	stock, sold := game.Sell(c, e.Slot, price, item.Stock)
	if sold {
		f.setStock(e.Item, stock)
		f.setNoteUI(fmt.Sprintf("Sold for $%d.", price), "facility.sold", price)
	}
}

// setStock 改一筆的店家庫存。
//
// ⚠ **不要留在會話裡。** 物品表在存檔區、每個存檔槽一份（`docs/re/45` §2），
// 庫存 `+0x02` 是遊戲狀態：賣一件 `+1`、買一件 `−1`（`docs/re/42` §3、§4）。
// 記在設施場景上的話，走出店門那一刻就沒了。
func (f *FacilityScene) setStock(id, v byte) {
	if f.SetStock != nil {
		f.SetStock(id, v)
	}
}

// member 是目前在櫃檯的人。
func (f *FacilityScene) member(p *game.Party) *game.Character {
	if p == nil || f.state.Who < 0 || f.state.Who >= len(p.Members) {
		return nil
	}
	return p.Members[f.state.Who]
}

// nextAble 從 from 之後找下一個能行動的人（繞一圈）。
//
// ⚠ CON ≤ 0 的人不能進商店（`sub_172BB`，docs/re/42 §1）。
func nextAble(p *game.Party, from int) int {
	n := len(p.Members)
	for i := 1; i <= n; i++ {
		j := (from + i) % n
		if game.CanCommand(p.Members[j]) {
			return j
		}
	}
	return from
}

// refresh 重排這個設施畫面要印的字。
func (f *FacilityScene) refresh(p *game.Party, items game.ItemTable) {
	f.Lines, f.CJKLines = f.Lines[:0], f.CJKLines[:0]
	// add 一次放一行：英文與中文並排，中文查不到就是 nil（那一行畫英文）。
	add := func(en string, zh []byte) {
		f.Lines = append(f.Lines, en)
		f.CJKLines = append(f.CJKLines, zh)
	}
	num := func(n int) []byte { return []byte(fmt.Sprintf("%d", n)) }
	// ui 是重製版自己排的那幾行（清單的每一列、價格欄位的順序）。
	ui := func(name string, args ...any) []byte {
		if f.UI == nil {
			return nil
		}
		b := f.UI(name)
		if len(b) == 0 {
			return nil
		}
		return []byte(fmt.Sprintf(string(b), args...))
	}

	if f.Facility.Name != "" {
		// 招牌是**地圖記錄裡的明文 ASCII**，不在字串表裡，所以走查表
		// （`internal/play/places.go`）。查不到就照原樣顯示英文。
		var zh []byte
		if f.CJKPlace != nil {
			zh = f.CJKPlace(f.Facility.Name)
		}
		add(f.Facility.Name, zh)
	}
	c := f.member(p)
	if c == nil {
		return
	}
	// 「某某 你有 $N」：名字 ＋ 字串 7:4（`You have $`）＋ 數字。
	add(fmt.Sprintf("%s  You have $%d", c.Name, c.Money),
		zhJoin([]byte(c.Name), []byte("  "),
			f.zh(exeTableShop, strYouHave, textlayout.Options{}), num(int(c.Money))))

	switch {
	case f.Facility.Kind == game.FacilityTrainer:
		add(fmt.Sprintf("Skill points: %d", c.SkillPts),
			zhJoin(f.zh(exeTableTrainer, strSkillPoints, textlayout.Options{}),
				num(int(c.SkillPts))))
		from, to := f.page(len(f.Skills))
		for i, sk := range f.Skills[from:to] {
			cost := game.SkillCost(sk.Data.BaseCost, int(c.SkillLevel(sk.ID))+1)
			var zh []byte
			if n := f.zhSkill(sk.ID); n != nil {
				zh = ui("facility.skillrow", i+1, string(n), cost)
			}
			add(fmt.Sprintf("%d) %s  cost %d", i+1, f.skillLabel(sk.ID), cost), zh)
		}
	case f.Facility.Kind == game.FacilityDoctor && f.state.Step == StepHeal:
		h := game.HealSession{Facility: f.Facility, Char: c}
		add(fmt.Sprintf("%d points. You can:  Heal 1 point / Continue", h.Remaining()),
			zhJoin(f.zh(exeTableDoctor, strNeedToHeal, textlayout.Options{}),
				num(h.Remaining()),
				f.zh(exeTableDoctor, strPointsYouCan, textlayout.Options{})))
	case f.Facility.Kind == game.FacilityDoctor && f.state.Step == StepCure:
		price := int(f.Facility.Price(0x06))
		add(fmt.Sprintf("Which to cure at $%d:", price),
			zhJoin(f.zh(exeTableDoctor, strWhichCure, textlayout.Options{}), num(price),
				f.zh(exeTableDoctor, strCureTail, textlayout.Options{})))
		for i, bit := range game.Diseases(c) {
			// 病名是字串表 8 的第 1–8 條（位元 0–7，`docs/re/35` §1）。
			name := f.zh(exeTableDoctor, strDiseaseFirst+bit, textlayout.Options{})
			var zh []byte
			if name != nil {
				zh = ui("facility.row", i+1, string(name))
			}
			add(fmt.Sprintf("%d) %s", i+1, f.diseaseLabel(bit)), zh)
		}
	case f.Facility.Kind == game.FacilityDoctor:
		price := int(f.Facility.Price(0x05))
		add(fmt.Sprintf("Exam $%d / Healing / Curing", price),
			zhJoin(f.zh(exeTableDoctor, strExamPrice, textlayout.Options{}), num(price),
				f.zh(exeTableDoctor, strHealing, textlayout.Options{}),
				f.zh(exeTableDoctor, strCuring, textlayout.Options{})))
	case f.state.Step == StepSell:
		add("   PRICE     ITEM", f.zh(exeTableShop, strPriceItem, textlayout.Options{}))
		list := f.sellable(p, items)
		from, to := f.page(len(list))
		for i, e := range list[from:to] {
			mark := " "
			if e.Equipped {
				mark = "*" // 裝備中要標出來，但賣得掉（docs/re/42 §3.1）
			}
			var zh []byte
			if n := f.zhItem(e.Item); n != nil {
				zh = ui("facility.sellrow", i+1, mark, string(n))
			}
			add(fmt.Sprintf("%d)%s %s", i+1, mark, f.itemLabel(e.Item)), zh)
		}
	case f.state.Step == StepBuy:
		add("   PRICE     ITEM", f.zh(exeTableShop, strPriceItem, textlayout.Options{}))
		list := f.buyList(items)
		from, to := f.page(len(list))
		for i, e := range list[from:to] {
			var zh []byte
			if n := f.zhItem(e.ID); n != nil {
				zh = ui("facility.buyrow", i+1, int(e.Price), string(n))
			}
			add(fmt.Sprintf("%d) $%-6d %s", i+1, e.Price, f.itemLabel(e.ID)), zh)
		}
	default:
		add("Do you want to:  Buy / Sell",
			f.zh(exeTableShop, strBuyOrSell, textlayout.Options{}))
	}
	if f.note != "" {
		add(f.note, f.noteCJK)
	}
}

// diseaseLabel 是病名的英文（字串表 8 的第 1–8 條，`docs/re/35` §1）。
// 查不到就退回位元編號——那是「還沒接上」的樣子，不是編一個病名。
func (f *FacilityScene) diseaseLabel(bit int) string {
	if f.Str != nil {
		if s := f.Str(exeTableDoctor, strDiseaseFirst+bit); s != "" {
			return s
		}
	}
	return fmt.Sprintf("status bit %d", bit)
}
