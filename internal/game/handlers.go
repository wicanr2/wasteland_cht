package game

// 四支指令處理程式（docs/spec/17、docs/re/41）。
//
// ⚠ **它們都不執行動作。** 換武器沒有真的換、裝填沒有真的填——
// 只把「選了什麼」記下來，動作在結算階段。與迴避（處理程式是空的）同一套設計。
// Use 寫的那一格是唯一的例外（§Use）。

// HandlerResult 是一支處理程式的結果。
//
// Accepted 為 false 就重問這個人（原版是 stc），Message 空字串表示不留訊息
// ——Load 的「沒裝備武器」那一道就是靜靜結束（docs/re/41 §4）。
type HandlerResult struct {
	Accepted bool
	Arg      byte
	Message  string
}

func reask(msg string) HandlerResult { return HandlerResult{Message: msg} }
func accept(arg byte) HandlerResult  { return HandlerResult{Accepted: true, Arg: arg} }

// 原版用到的字串編號（字串表 1，docs/re/41 §6）。
// 這一層只給編號，文字在呈現層查翻譯目錄。
const (
	MsgNoOneInRange   = 57 // No one is within range.
	MsgWhichGroup     = 63 // Which group?
	MsgNothingToUse   = 64 // You don't have anything.
	MsgNoMoreClips    = 65 // has no more clips.
	MsgCantBeReloaded = 66 // 's weapon can't be reloaded.
)

// ItemSlots 是角色記錄的物品陣列格數（+0xBD 起，30 槽、stride 2）。
const ItemSlots = 30

// CountItems 數物品陣列裡有幾件（sub_1963A）。值為 0 的槽是空的。
//
// keep 為 nil 表示不過濾（原版 0x1963A 入口）；給了就只算通過的
// （原版 0x1963F 入口多一道物品類別的過濾）。
func CountItems(items []byte, keep func(byte) bool) int {
	n := 0
	for i := 0; i < len(items) && i < ItemSlots; i++ {
		if items[i] == 0 {
			continue
		}
		if keep != nil && !keep(items[i]) {
			continue
		}
		n++
	}
	return n
}

// FindItem 在物品陣列裡線性找一件東西（sub_1968C），回傳槽編號。
func FindItem(items []byte, want byte) (int, bool) {
	for i := 0; i < len(items) && i < ItemSlots; i++ {
		if items[i] == want {
			return i, true
		}
	}
	return 0, false
}

// HandleWeapon 是換武器（碼 3）：空手就重問，否則開物品清單。
//
// pick 是清單的選擇（回傳槽編號與有沒有選）。原版是 sub_19394，
// 它的回傳值編碼還沒逆向，所以由呼叫端提供。
func HandleWeapon(items []byte, pick func() (byte, bool)) HandlerResult {
	if CountItems(items, nil) == 0 {
		return reask(msg(MsgNothingToUse))
	}
	slot, ok := pick()
	if !ok {
		return reask("")
	}
	return accept(slot)
}

// HandleHire 是雇用（碼 1）。
//
// ⚠ 「可雇用的對象數」怎麼算還沒逆向（`loc_1382B`，docs/re/41 §7），
// 所以候選由呼叫端提供——**不要在這裡編一個判斷條件出來**。
func HandleHire(candidates int, pick func() (byte, bool)) HandlerResult {
	if candidates <= 0 {
		return reask(msg(MsgNoOneInRange))
	}
	n, ok := pick()
	if !ok {
		return reask("") // 原版回 0xFF ＝ 取消
	}
	return accept(n)
}

// HandleLoad 是裝填／排除卡彈（碼 6）的三道檢查（docs/re/41 §4）。
//
// ⚠ 順序固定，而且**第一道不留訊息**——沒裝備武器時原版是靜靜結束的。
func HandleLoad(armed bool, ammoType byte, items []byte) HandlerResult {
	if !armed {
		return reask("") // 靜靜結束，不印訊息
	}
	if ammoType == 0 {
		return reask(msg(MsgCantBeReloaded))
	}
	slot, ok := FindItem(items, ammoType)
	if !ok {
		return reask(msg(MsgNoMoreClips))
	}
	return accept(byte(slot))
}

// UseKind 是 Use 的第一層選擇。
//
// ⚠ **順序照字母表 `ds:A5E8h`（`53 49 41` ＝ `SIA`），不是字串 4 的顯示順序。**
// 字串 4 印的是「Use: Item / Skill / Attribute」，而按鍵是拿大寫化的鍵去線性掃
// 那張字母表、回傳**索引**（`sub_173B0`，`docs/re/46` §4）——
// 兩者順序不同，照顯示文字編號會把三條路全部對錯人
// （與 `docs/re/38` §2「選單顯示的順序不是指令碼」同一個坑）。
type UseKind byte

const (
	UseSkill UseKind = iota // 'S'
	UseItem                 // 'I'
	UseAttribute            // 'A'
)

// UseChoices 是每個角色一格的「這回合要用什麼」（原版 ds:A9FDh + 角色編號）。
//
// ⚠ **索引是角色編號不是隊伍位置**——同一個角色換到別的隊伍位置，這一格跟著他。
// 那個 byte 的欄位配置還沒逆向，所以這裡只存放與取出，不解釋內容。
type UseChoices map[byte]byte

// HandleUse 是使用（碼 7）：選完把一個 byte 記在該角色那一格。
//
// choose 回傳原版合成出來的那個 byte（`sub_12738` 的低半部與選擇合成，
// 合成規則還沒逆向），以及有沒有選。
func HandleUse(choices UseChoices, character byte, choose func() (byte, bool)) HandlerResult {
	v, ok := choose()
	if !ok {
		return reask("")
	}
	if choices != nil {
		choices[character] = v
	}
	return accept(v)
}

// msg 把字串編號包成「呈現層要查的 key」。規則層不碰文字本身。
func msg(id int) string { return "exe:1:" + itoa(id) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [4]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
