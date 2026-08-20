package play

// 設施的選單互動（docs/spec/25、docs/re/42）。
//
// 規格 18 是規則、規格 23 是場景，這個檔是中間的互動：
// 按哪個鍵、走到哪一步。**規則一條都不在這裡重做。**

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/render"
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
	strWhoEnters    = 2  // 7：`Who wants to enter?`（商店）
	strWhoTreat     = 23 // 8：`Who wants treatment?`（醫生）
	strBeyondHelp   = 9  // 8：`\x0B is beyond my help.`
	strRecommend    = 10 // 8：`"Well \x0B, I would recommend:"`
	strTrainerWho   = 5  // 6：`Who wants to enter?`（訓練師）
	strNoCondition  = 6  // 6：`\x0B is in no condition to learn.`
	strTrainerHead  = 3  // 6：`   IQ PTS LVL   SKILL`
	strCantBuy      = 3  // 7：`\x0B can't buy anything.`
	strBuyOrSell    = 6  // 7：`Do you want to:` ＋ Buy／Sell
	strPriceItem    = 7  // 7：`   PRICE     ITEM`
	strInvFull      = 8  // 7：`Your inventory is full.`
	strSkillPoints  = 4  // 6：`Skill points = `
	strNoSkillPts   = 7  // 6：`Not enough skill points!`
	strNotSmart     = 9  // 6：`\x0b is not smart enough to learn anything here.`
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
	// StepWho 是商店的「誰要進去？」（字串 7:2）。
	// **隊伍只有一個人時不問**（`docs/re/42` §1），直接用他。
	StepWho
)

// 商店與醫生的按鍵（docs/re/42 §1、§5）。原版比對的是靜態字母表，
// 不跟著翻譯走（docs/re/40 §4）——所以這裡也寫死。
const (
	keyBuy      = 'B'
	keySell     = 'S'
	// keyPool 是商店的 `P`：**集中金錢**（畫面外框下緣就寫著 `POOL MONEY`，
	// `sub_19B81`，`docs/re/117` §3）。
	keyPool = 'P'
	// keyNextChar 是醫生與訓練師的 `P`：換下一個人。
	//
	// ⚠ **這兩支還沒對過實機**——商店那個 `P` 原本也記成「換人」，
	// 是截圖上的 `POOL MONEY` 六個字推翻的。要下結論之前先照 §5 的指令跑一次。
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

// moreLabel 是清單還有下一頁時畫的那一個字（`docs/re/117` §2.2）。
//
// 原版把它畫在選單框內最後一列的左緣，而且套的是框邊標籤那套冷色字模
// （實機 `47-buy2.png`）。這裡是接在清單後面當一行——**落點同一列**，
// 只是字模走一般文字，還沒換成標籤那一套。
const moreLabel = "MORE!"

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
		// ⚠ **「誰要進去？」那一層的 ESC 是離開**（原版：選不到人就走人，
		// `docs/re/42` §1），不是退回主迴圈——它本來就是最外面那一層。
		if st.Step != StepMain && st.Step != StepWho {
			st.Step = StepMain // 清單 → 主迴圈
			f.refresh(p, items)
			return true
		}
		return false // 主迴圈（或選人那一層）→ 離開設施
	}

	// 翻頁對每一種清單都一樣（docs/re/53 §4）。
	//
	// ⚠ **判準是「現在有沒有清單」，不是「在哪一層」**：訓練師的技能清單掛在
	// 主迴圈上（36 個技能、一頁九列），拿 `Step` 當判準的話第 10 個技能以後
	// 永遠學不到，而畫面上只是「清單就這麼長」。原版的清單按鍵處理
	// （`sub_16D34`）是三種設施共用的，訓練師的清單也有上下箭頭
	// （`docs/re/129` §4 的 `0x1C8C2`）。
	if st.Step != StepWho && f.rowCount(p, items) > 0 {
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

	// 「誰要進去／誰要治療」那一層三種設施共用（`sub_1721B`，`docs/re/119`）：
	// 只收號碼，選到不能行動的人要印一句話退回來重選。
	if st.Step == StepWho {
		f.whoKey(k, p)
		f.refresh(p, items)
		return true
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

// whoKey 是「誰要進去？」那一層的按鍵（三種設施共用）。
func (f *FacilityScene) whoKey(k byte, p *game.Party) {
	if k < '1' || k > '9' {
		return
	}
	n := int(k - '1')
	if n >= len(p.Members) || p.Members[n] == nil {
		return
	}
	if !game.CanCommand(p.Members[n]) {
		en, table, id := f.cantMessage()
		f.setNote(p.Members[n].Name+en, table, id)
		return
	}
	f.state.Who, f.state.Step = n, StepMain
}

// whoPrompt 是那一層要印的問句（每種設施自己的字串）。
func (f *FacilityScene) whoPrompt() (en string, table, id int) {
	switch f.Facility.Kind {
	case game.FacilityDoctor:
		return "Who wants treatment?", exeTableDoctor, strWhoTreat
	case game.FacilityTrainer:
		return "Who wants to enter?", exeTableTrainer, strTrainerWho
	}
	return "Who wants to enter?", exeTableShop, strWhoEnters
}

// cantMessage 是「這個人不能用這家店」那一句（接在名字後面）。
func (f *FacilityScene) cantMessage() (en string, table, id int) {
	switch f.Facility.Kind {
	case game.FacilityDoctor:
		return " is beyond my help.", exeTableDoctor, strBeyondHelp
	case game.FacilityTrainer:
		return " is in no condition to learn.", exeTableTrainer, strNoCondition
	}
	return " can't buy anything.", exeTableShop, strCantBuy
}

func (f *FacilityScene) shopKey(k byte, p *game.Party, items game.ItemTable) {
	st := f.state
	switch k {
	case keyPool:
		// 把其他隊員身上的錢全部搬給櫃檯前這個人（`sub_19B81`）。
		game.PoolMoney(p, st.Who)
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
	if k == keyPool {
		// 醫生也有 `POOL MONEY`（實機截圖 `54-doc-menu.png` 的外框下緣）。
		game.PoolMoney(p, st.Who)
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
				f.note, f.noteCJK = "", ""
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
	// ⚠ **訓練師沒有 `P`**：外框下緣只有 `MORE!`，清單迴圈也沒有設
	// `ds:470Eh`（`0x1BC42` 那一段）——實機截圖 `61-tr-menu.png`。
	if k >= '1' && k <= '9' {
		f.learnOne(p, f.state.Page+int(k-'1'))
	}
	_ = st
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
	//
	// ⚠ **中文那一份可能是好幾行**：原版的字串裡有換行控制碼（`\x0D`），
	// `textlayout.Render` 把它換成 `\n`。設施畫面是逐行畫的，不像訊息視窗
	// 會自己斷行——不拆開的話那幾個 `\n` 會被當成字畫出來，
	// 畫面上是「你要：」前後各一個方框（實機對拍看出來的，`docs/re/117` §4）。
	add := func(en string, zh string) {
		parts := splitLines(zh)
		f.Lines = append(f.Lines, en)
		f.CJKLines = append(f.CJKLines, parts[0])
		for _, extra := range parts[1:] {
			f.Lines = append(f.Lines, "")
			f.CJKLines = append(f.CJKLines, extra)
		}
	}
	num := func(n int) string { return fmt.Sprintf("%d", n) }
	// ui 是重製版自己排的那幾行（清單的每一列、價格欄位的順序）。
	ui := func(name string, args ...any) string {
		if f.UI == nil {
			return ""
		}
		b := f.UI(name)
		if b == "" {
			return ""
		}
		return fmt.Sprintf(b, args...)
	}

	if f.Facility.Name != "" {
		// 招牌是**地圖記錄裡的明文 ASCII**，不在字串表裡，所以走查表
		// （`internal/play/places.go`）。查不到就照原樣顯示英文。
		var zh string
		if f.CJKPlace != nil {
			zh = f.CJKPlace(f.Facility.Name)
		}
		add(f.Facility.Name, zh)
	}
	// 招呼語：**地圖記錄 `+0x05` 指到這張地圖自己的字串**（`0x1BEB7` 把它
	// 存進 `ds:DBF4h`），不是執行檔字串表——所以每家店的招呼詞不一樣。
	if f.Greeting != "" {
		// 選單區只有 24 欄，原版會折行（實機截圖上「Welcome to the /
		// infirmary.」就是兩列）。折不了的長字硬斷，不要讓它畫出面板。
		en := wrapCells(f.Greeting, render.PanelWidth)
		zh := wrapCells(f.GreetingCJK, render.PanelWidth)
		// ⚠ **有中文就整段走中文。** 兩邊的折行數不一樣（英文「Welcome to the ／
		// infirmary.」兩列、中文「歡迎來到醫護所。」一列），逐行配對會讓多出來的
		// 那一行英文沒有中文可蓋，於是卡在兩行中文之間。
		if len(zh) > 0 {
			en = make([]string, len(zh))
		}
		for i := 0; i < len(en) || i < len(zh); i++ {
			var a, b string
			if i < len(en) {
				a = en[i]
			}
			if i < len(zh) {
				b = zh[i]
			}
			add(a, b)
		}
	}
	if f.state.Step == StepWho {
		// ⚠ **不列名字**：號碼與名字在底下的隊伍名單上（`1>`…`4>`），
		// 原版就是靠那一份選人。在選單區再列一次會佔掉四行。
		en, table, id := f.whoPrompt()
		add(en, f.zh(table, id, textlayout.Options{}))
		if f.note != "" {
			add(f.note, f.noteCJK)
		}
		return
	}
	c := f.member(p)
	if c == nil {
		return
	}
	// 「某某 你有 $N」：名字 ＋ 字串 7:4（`You have $`）＋ 數字。
	add(fmt.Sprintf("%s  You have $%d", c.Name, c.Money),
		zhJoin(c.Name, "  ",
			f.zh(exeTableShop, strYouHave, textlayout.Options{}), num(int(c.Money))))

	switch {
	case f.Facility.Kind == game.FacilityTrainer:
		add(fmt.Sprintf("Skill points = %d", c.SkillPts),
			zhJoin(f.zh(exeTableTrainer, strSkillPoints, textlayout.Options{}),
				num(int(c.SkillPts))))
		// 表頭與三欄照原版：IQ 需求、這一級要幾點、目前等級
		// （字串 6:3，實機截圖 `61-tr-menu.png`）。
		add("   IQ PTS LVL   SKILL",
			f.zh(exeTableTrainer, strTrainerHead, textlayout.Options{}))
		from, to := f.page(len(f.Skills))
		for i, sk := range f.Skills[from:to] {
			cost := game.SkillCost(sk.Data.BaseCost, int(c.SkillLevel(sk.ID))+1)
			lvl := int(c.SkillLevel(sk.ID))
			var zh string
			if n := f.zhSkill(sk.ID); n != "" {
				zh = ui("facility.skillrow", i+1, int(sk.Data.IQ), cost, lvl, string(n))
			}
			add(fmt.Sprintf("%d> %2d %3d %3d  %s",
				i+1, sk.Data.IQ, cost, lvl, f.skillLabel(sk.ID)), zh)
		}
		f.addMore(add, ui, to, len(f.Skills))
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
			var zh string
			if name != "" {
				zh = ui("facility.row", i+1, string(name))
			}
			add(fmt.Sprintf("%d> %s", i+1, f.diseaseLabel(bit)), zh)
		}
		f.addMore(add, ui, len(game.Diseases(c)), len(game.Diseases(c)))
	case f.Facility.Kind == game.FacilityDoctor:
		// 主選單是**條件式的**：醫生只列這個人現在用得到的項目
		// （`0x1C2CE`–`0x1C2FF` 的三個旗標，`docs/re/119` §2）。
		// 健康的人只有 `Exam $N` ——三個都印的話玩家會以為隨時能治療。
		name := func() string { return c.Name }
		add(fmt.Sprintf("\"Well %s, I would recommend:\"", c.Name),
			f.zh(exeTableDoctor, strRecommend, textlayout.Options{Name: name}))
		price := int(f.Facility.Price(0x05))
		add(fmt.Sprintf("Exam $%d", price),
			zhJoin(f.zh(exeTableDoctor, strExamPrice, textlayout.Options{}), num(price)))
		if c.CON < c.MaxCON {
			add("Healing", f.zh(exeTableDoctor, strHealing, textlayout.Options{}))
		}
		if c.Status != 0 {
			add("Curing", f.zh(exeTableDoctor, strCuring, textlayout.Options{}))
		}
	case f.state.Step == StepSell:
		add("   PRICE     ITEM", f.zh(exeTableShop, strPriceItem, textlayout.Options{}))
		list := f.sellable(p, items)
		from, to := f.page(len(list))
		for i, e := range list[from:to] {
			mark := " "
			if e.Equipped {
				mark = "*" // 裝備中要標出來，但賣得掉（docs/re/42 §3.1）
			}
			var zh string
			if n := f.zhItem(e.Item); n != "" {
				zh = ui("facility.sellrow", i+1, mark, string(n))
			}
			add(fmt.Sprintf("%d>%s %s", i+1, mark, f.itemLabel(e.Item)), zh)
		}
		f.addMore(add, ui, to, len(list))
	case f.state.Step == StepBuy:
		add("   PRICE     ITEM", f.zh(exeTableShop, strPriceItem, textlayout.Options{}))
		list := f.buyList(items)
		from, to := f.page(len(list))
		for i, e := range list[from:to] {
			var zh string
			if n := f.zhItem(e.ID); n != "" {
				zh = ui("facility.buyrow", i+1, int(e.Price), string(n))
			}
			// 版面照原版：列號是 `N>`（不是 `N)`）、價錢**右對齊**在 `PRICE`
			// 那一欄底下、名字從第 13 欄起（實機截圖 `44-zoom.png`）。
			// 原版沒有印 `$`。
			add(fmt.Sprintf("%d> %8d  %s", i+1, e.Price, f.itemLabel(e.ID)), zh)
	}
		f.addMore(add, ui, to, len(list))
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

// splitLines 把一段可能多行的譯文拆成逐行，並**去掉頭尾的空行**。
//
// 頭尾的空行來自原版字串的換行控制碼：那些字串假設自己接在別人後面印，
// 所以開頭先換一行。逐行畫的時候那一行就變成畫面上真的一列空白，
// 選單會整個往下掉一格。
func splitLines(s string) []string {
	parts := strings.Split(s, "\n")
	for len(parts) > 1 && parts[0] == "" {
		parts = parts[1:]
	}
	for len(parts) > 1 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// addMore 在清單還有下一頁時補一行 `MORE!`（`docs/re/117` §2.2）。
//
// ⚠ **判斷用的是「這一頁畫到第幾列」與總數**，不是頁碼：
// 拿頁碼算的話最後一頁剛好滿的時候也會冒出 `MORE!`，
// 而畫面上看起來只是「還有東西」——按下去卻沒有下一頁。
func (f *FacilityScene) addMore(add func(string, string), ui func(string, ...any) string,
	shown, total int) {
	if shown >= total {
		return
	}
	var zh string
	if ui != nil {
		zh = ui("facility.more")
	}
	add(moreLabel, zh)
}

// wrapCells 把一段字折成每行最多 cols 格（**一個 rune 一格**，`docs/spec/10` §3）。
//
// 有空白就在空白處斷，沒有就硬斷——原版的選單區只有 24 欄，
// 不折行的話字會畫到面板外面，而畫面上看起來像「這一行怎麼跑到名單上」。
func wrapCells(s string, cols int) []string {
	if s == "" || cols <= 0 {
		return nil
	}
	var out []string
	for _, para := range splitLines(s) {
		cells := []rune(para)
		for len(cells) > cols {
			cut := cols
			for i := cols; i > 0; i-- {
				if cells[i-1] == ' ' {
					cut = i
					break
				}
			}
			out = append(out, strings.TrimRight(string(cells[:cut]), " "))
			cells = cells[cut:]
		}
		out = append(out, string(cells))
	}
	return out
}

// Paged 回報畫面上現在有沒有一份**翻得了頁**的清單。
//
// 這決定選單框右邊那兩個上下箭頭畫不畫（`docs/re/129` §4）。
func (f *FacilityScene) Paged(p *game.Party, items game.ItemTable) bool {
	if f.state == nil || f.state.Step == StepWho {
		return false
	}
	return f.rowCount(p, items) > PageRows
}

// Who 回報**站在櫃檯前的是第幾個人**（0 起算），以及現在有沒有人在櫃檯。
//
// 「誰要進去？」那一步回 false：那時候還沒有人被選中
// （`docs/re/128` §3，實機 `42-shop.png` 名單上沒有反白的序號）。
func (f *FacilityScene) Who() (int, bool) {
	if f.state == nil || f.state.Step == StepWho {
		return 0, false
	}
	return f.state.Who, true
}

// HasPool 回報這一層設施**現在**有沒有 `P`（集中金錢）。
//
// 商店與醫生有、**訓練師沒有**（`docs/re/119` §3：實機的訓練師畫面上
// 外框下緣沒有 `POOL MONEY` 那個標籤）。
//
// ⚠ **「誰要進去？」那一步還沒有。** 實機截圖裡問話那一張
// （`workplace/dosbox/shots/42-shop.png`）外框下緣是空的，選完人的下一張
// （`43-menu.png`）才出現 `POOL MONEY`——那個鍵屬於櫃檯前那個人，
// 還沒選人就還沒有人可以收錢。
func (f *FacilityScene) HasPool() bool {
	if f.Facility.Kind == game.FacilityTrainer {
		return false
	}
	return f.state == nil || f.state.Step != StepWho
}
