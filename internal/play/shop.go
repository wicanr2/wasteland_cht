package play

// 設施的選單互動（docs/spec/25、docs/re/42）。
//
// 規格 18 是規則、規格 23 是場景，這個檔是中間的互動：
// 按哪個鍵、走到哪一步。**規則一條都不在這裡重做。**

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// FacilityStep 是設施選單的一個狀態。
type FacilityStep int

const (
	StepMain FacilityStep = iota // 主選單
	StepBuy                      // 買的清單
	StepSell                     // 賣的清單
)

// 商店與醫生的按鍵（docs/re/42 §1、§5）。原版比對的是靜態字母表，
// 不跟著翻譯走（docs/re/40 §4）——所以這裡也寫死。
const (
	keyBuy      = 'B'
	keySell     = 'S'
	keyNextChar = 'P'
	keyHeal     = 'H'
	keyContinue = 'C'
	keyEscape   = 0x1B
)

// shopState 是設施選單的狀態。放在 FacilityScene 裡，
// 離開設施就整個丟掉——原版也沒有跨場次保留的東西。
type shopState struct {
	Step  FacilityStep
	Who   int // 目前在櫃檯的隊伍成員索引
	Stock map[byte]byte
}

// Key 送一個按鍵給設施選單。回傳 false 表示要離開設施。
//
// ⚠ **「背包滿了」與「沒有東西賣」都回主迴圈，不是離開**（docs/spec/25 §2）——
// 那是原版的正常路徑，不是錯誤分支。
func (f *FacilityScene) Key(k byte, p *game.Party, items game.ItemTable) bool {
	if f.state == nil {
		f.state = &shopState{Stock: map[byte]byte{}}
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
			f.note = "Your inventory is full."
			st.Step = StepMain
			return
		}
		if len(f.buyList(items)) == 0 {
			// 全部缺貨。原版的 sub_1C140 回 0 之後也是印字串 8 回主迴圈。
			f.note = "We are temporarily out of stock."
			st.Step = StepMain
			return
		}
		st.Step = StepBuy
	case keySell:
		if len(f.sellable(p, items)) == 0 {
			f.note = "You don't have anything they want!"
			st.Step = StepMain
			return
		}
		st.Step = StepSell
	default:
		if k >= '1' && k <= '9' {
			switch st.Step {
			case StepSell:
				f.sellOne(p, items, int(k-'1'))
			case StepBuy:
				f.buyOne(p, items, int(k-'1'))
			}
		}
	}
}

func (f *FacilityScene) doctorKey(k byte, p *game.Party) {
	st := f.state
	switch k {
	case keyNextChar:
		st.Who = nextAble(p, st.Who)
	case keyContinue:
		st.Step = StepMain
	case keyHeal:
		h := game.HealSession{Facility: f.Facility, Char: f.member(p)}
		if ok, reason := h.HealOne(); !ok {
			f.note = reason
		} else {
			f.note = ""
		}
	}
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
			return f.stockOf(id, it.Stock)
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
		f.note = "You don't have enough money."
		return
	}
	f.state.Stock[e.ID] = stock
	f.note = fmt.Sprintf("Bought for $%d.", e.Price)
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
		f.learnOne(p, int(k-'1'))
	}
}

// learnOne 學清單上的第 n 個技能。
func (f *FacilityScene) learnOne(p *game.Party, n int) {
	c := f.member(p)
	if c == nil {
		return
	}
	if c.SkillPts == 0 {
		f.note = "You have no skill points."
		return
	}
	list := f.Skills
	if n < 0 || n >= len(list) {
		return
	}
	ok, reason := c.LearnSkill(list[n].ID, list[n].Data)
	if !ok {
		f.note = reason
		return
	}
	f.note = fmt.Sprintf("Learned skill %d.", list[n].ID)
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
	stock, sold := game.Sell(c, e.Slot, price, f.stockOf(e.Item, item.Stock))
	if sold {
		f.state.Stock[e.Item] = stock
		f.note = fmt.Sprintf("Sold for $%d.", price)
	}
}

// stockOf 取這家店目前的庫存（賣過的以會話裡的為準）。
func (f *FacilityScene) stockOf(id, base byte) byte {
	if f.state == nil {
		return base
	}
	if v, ok := f.state.Stock[id]; ok {
		return v
	}
	return base
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
	f.Lines = f.Lines[:0]
	if f.Facility.Name != "" {
		f.Lines = append(f.Lines, f.Facility.Name)
	}
	c := f.member(p)
	if c == nil {
		return
	}
	f.Lines = append(f.Lines, fmt.Sprintf("%s  You have $%d", c.Name, c.Money))

	switch {
	case f.Facility.Kind == game.FacilityTrainer:
		f.Lines = append(f.Lines, fmt.Sprintf("Skill points: %d", c.SkillPts))
		for i, sk := range f.Skills {
			if i >= 9 {
				break
			}
			f.Lines = append(f.Lines,
				fmt.Sprintf("%d) skill %d  cost %d", i+1, sk.ID,
					game.SkillCost(sk.Data.BaseCost, int(c.SkillLevel(sk.ID))+1)))
		}
	case f.Facility.Kind == game.FacilityDoctor:
		h := game.HealSession{Facility: f.Facility, Char: c}
		f.Lines = append(f.Lines,
			fmt.Sprintf("%d points. You can:  Heal 1 point / Continue", h.Remaining()))
	case f.state.Step == StepSell:
		f.Lines = append(f.Lines, "   PRICE     ITEM")
		for i, e := range f.sellable(p, items) {
			mark := " "
			if e.Equipped {
				mark = "*" // 裝備中要標出來，但賣得掉（docs/re/42 §3.1）
			}
			f.Lines = append(f.Lines, fmt.Sprintf("%d)%s item %d", i+1, mark, e.Item))
		}
	case f.state.Step == StepBuy:
		f.Lines = append(f.Lines, "   PRICE     ITEM")
		for i, e := range f.buyList(items) {
			if i >= 9 {
				break // 選擇鍵只有 '1'–'9'；分頁是清單框架的事（未解）
			}
			f.Lines = append(f.Lines, fmt.Sprintf("%d) $%-6d item %d", i+1, e.Price, e.ID))
		}
	default:
		f.Lines = append(f.Lines, "Do you want to:  Buy / Sell")
	}
	if f.note != "" {
		f.Lines = append(f.Lines, f.note)
	}
}
