package play

// 設施場景：踩上去到走出來（docs/spec/23、docs/re/29 §5.4）。
//
// 規格 09 是規則（價格、治療、學習），規格 18 是買賣與治療的迴圈，
// 這個檔只做場景：載圖、印地點名、離開時切回地圖。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/render"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// recStockGroup 是商店記錄裡「這家店賣哪一組東西」的位移（`0x1BEA2`，`docs/re/118`）。
const recStockGroup = 0x06

// greetingAt 回報這一種設施的招呼語字串編號放在記錄的第幾個 byte
// （`docs/re/119` §1）。−1 ＝ 這一種沒有招呼語。
func greetingAt(k game.FacilityKind) int {
	switch k {
	case game.FacilityShop:
		return 0x05 // `0x1BEB7`：抄進 ds:DBF4h
	case game.FacilityDoctor, game.FacilityTrainer:
		return 0x03 // `0x1C29F`／`0x1BBCF`：直接餵給 sub_178A0
	}
	return -1
}

// asksWho 回報這一種設施進場要不要先問「誰要進去」（`sub_1721B`）。
func asksWho(k game.FacilityKind) bool {
	switch k {
	case game.FacilityShop, game.FacilityDoctor, game.FacilityTrainer:
		return true
	}
	return false
}

// ableCount 數這一組有幾個人能行動（`sub_172BB` 那個條件）。
func ableCount(p *game.Party) int {
	if p == nil {
		return 0
	}
	n := 0
	for _, m := range p.Members {
		if m != nil && game.CanCommand(m) {
			n++
		}
	}
	return n
}

// facilityPicture 是每個設施進場要載的 ALLPICS 圖（docs/re/29 §5.4 那張表）。
//
// ⚠ 第五種（FacilityEnding）**沒有圖**——原版那一支（0x1B4F0）連
// `sub_190A6` 都沒叫，因為它根本不是設施畫面：走進去就播結局
// （`docs/re/96`）。用 -1 表示「沒有」，不要猜一個編號。
var facilityPicture = [game.FacilityCount]int{
	game.FacilityDoctor:  0,
	game.FacilityShop:    1,
	game.FacilityTrainer: 2,
	game.FacilityRoster:  3,
	game.FacilityEnding: -1,
}

// FacilityScene 是一個設施畫面的狀態。
type FacilityScene struct {
	Facility game.Facility
	Picture  int // ALLPICS 的圖片編號；-1 ＝ 這種設施沒有圖
	Lines    []string

	// state 是選單的狀態（docs/spec/25）。離開設施就整個丟掉——
	// 原版也沒有跨場次保留的東西。
	state *shopState
	note  string // 這一步要多印的一行（「背包滿了」那類）

	// Skills 是訓練師教得了的技能。**這家店教什麼由呼叫端給**——
	// 原版的清單走 `sub_1BDFF`，那一支與商店的清單框架同源、還沒逆向
	// （docs/re/52 §4），所以不在這裡假裝算得出來。
	Skills []TrainableSkill

	// ItemName／SkillName 由 Scene 進場時接上（`docs/re/17` §4 的名稱表）。
	// 沒接就退回 `item 43` 這種編號——**清單印編號玩家看不懂買的是什麼**，
	// 但那是「還沒接上」的樣子，不是假裝有名字。
	ItemName  func(byte) string
	SkillName func(byte) string
	// CJKPlace 把招牌的英文原名換成中文（`internal/play/places.go`）。
	// **招牌不在字串表裡**，所以它走查表不走目錄 key。
	CJKPlace func(string) string
	// CJKName 是同兩張表的中文（Big5）。查不到回 nil，那一行就退回英文。
	CJKItemName  func(byte) string
	CJKSkillName func(byte) string

	// SetStock 改一筆物品的店家庫存（買賣之後）。由 Scene 接上——
	// 庫存住在**物品表**上，跟著存檔走，不是設施場景的狀態。
	SetStock func(id, v byte)

	// Str 查原版字串表的**原文**、CJK 查譯文、UI 查重製版自己的介面文字。
	// 三個都可以是 nil——那時設施畫面就是英文字面，遊戲照跑。
	Str func(table, n int) string
	CJK      func(table, n int, opt textlayout.Options) string
	UI  func(name string) string

	// Greeting／GreetingCJK 是這家店的招呼語（地圖記錄 `+0x05` 指到的
	// **這張地圖自己的字串**，`0x1BEB7`）。查不到就是空的，那一行不印。
	Greeting    string
	GreetingCJK string

	// CJKLines 與 Lines 一一對應的中文（Big5）。某一行查不到就是 nil，
	// 那一行改畫英文——**不要整片放棄**，設施畫面每一行的來源都不一樣。
	CJKLines []string

	noteCJK string // note 那一行的中文
}

// setNote 設一行註記：英文字面 ＋ 原版字串表的譯文。
func (f *FacilityScene) setNote(en string, table, n int) {
	f.note = en
	f.noteCJK = f.zh(table, n, textlayout.Options{})
}

// setNoteUI 設一行註記，中文走重製版的 `ui:`（原版沒有對應字串的那些）。
func (f *FacilityScene) setNoteUI(en, uiName string, args ...any) {
	f.note = en
	f.noteCJK = ""
	if f.UI == nil {
		return
	}
	if b := f.UI(uiName); len(b) > 0 {
		f.noteCJK = fmt.Sprintf(b, args...)
	}
}

// setNoteReason 設一行來自規則層的失敗理由。
//
// 規則層回的是中文字面（`internal/game` 的 `"錢不夠"`），不是 Big5——
// 那是給開發者看的，**畫面上要走目錄**。查不到就照原樣顯示。
func (f *FacilityScene) setNoteReason(reason string) {
	f.note = reason
	f.noteCJK = ""
}

// zh 查原版字串表第 table 張第 n 條的譯文。
func (f *FacilityScene) zh(table, n int, opt textlayout.Options) string {
	if f.CJK == nil {
		return ""
	}
	return f.CJK(table, n, opt)
}

// zhItem／zhSkill 是清單那一欄的中文名。
func (f *FacilityScene) zhItem(id byte) string {
	if f.CJKItemName == nil {
		return ""
	}
	return f.CJKItemName(id)
}

func (f *FacilityScene) zhSkill(id byte) string {
	if f.CJKSkillName == nil {
		return ""
	}
	return f.CJKSkillName(id)
}

// itemLabel／skillLabel 是清單那一欄要印的字，沒接名稱表就退回編號。
func (f *FacilityScene) itemLabel(id byte) string {
	if f.ItemName != nil {
		return f.ItemName(id)
	}
	return fmt.Sprintf("item %d", id)
}

func (f *FacilityScene) skillLabel(id byte) string {
	if f.SkillName != nil {
		return f.SkillName(id)
	}
	return fmt.Sprintf("skill %d", id)
}

// TrainableSkill 是訓練師教得了的一個技能。
type TrainableSkill struct {
	ID   byte
	Data game.SkillData
}

// EnterFacility 從地圖進設施。
//
// 回 nil 表示這一格不是設施（記錄 `+0x00` 的 bit7 沒設 → 那是腳本指令，
// 走規格 07 的腳本直譯器，不是這裡）。
func (s *Scene) EnterFacility(record []byte) *FacilityScene {
	f, ok := game.ParseFacility(record)
	if !ok {
		return nil
	}
	// 選單狀態一開始就建好：`Frame()` 與 `Key()` 都會讀它，
	// 只在 Key 裡懶初始化的話，先畫再按就會踩到 nil。
	// 第 4 種不是設施，是**結局**（`ds:A4E0h[4]` → `0x1B4F0`）。
	if f.Kind == game.FacilityEnding {
		s.BeginEnding()
		return nil
	}
	// 商店有自己的庫存表：記錄 `+0x06` 選第幾組（`docs/re/118`）。
	// ⚠ **只有商店**——醫生的 `+0x06` 是每點治療費，拿去當庫存組會載到別的表。
	if f.Kind == game.FacilityShop && len(record) > recStockGroup {
		if err := s.loadItemStock(record[recStockGroup]); err != nil {
			s.message = "ITEM TABLE: " + err.Error()
		}
	}
	fs := &FacilityScene{
		Facility: f,
		Picture:  facilityPicture[f.Kind],
		state:    &shopState{},
	}
	fs.CJKPlace = s.placeCJK
	if f.Name != "" {
		fs.Lines = append(fs.Lines, f.Name)
		// ⚠ 這裡與 `shop.go` 的選單重建**兩處都要設**：進場先擺一份
		// （角色管理那條路不重建選單），商店那條路每次重建又會蓋掉，
		// 所以那邊也要走同一個查表 hook。少設一處的症狀是
		// 「有些設施的招牌是中文、有些是英文」。
		fs.CJKLines = append(fs.CJKLines, fs.CJKPlace(f.Name))
	}
	// 招呼語：記錄裡的一個 byte 指到**這張地圖自己的字串**（`docs/re/119` §1）。
	// ⚠ **位移每一種設施不一樣**：商店 `+0x05`、醫生與訓練師 `+0x03`——
	// 拿錯位移會印出別的句子（訓練師的 `+0x05` 落在名字欄位裡）。
	if at := greetingAt(f.Kind); at >= 0 && len(record) > at &&
		s.world != nil && s.world.Block != nil {
		if n := int(record[at]); n > 0 && n < len(s.world.Block.Strings) {
			// ⚠ 原文帶著控制碼（開頭常是 `\x07` 換頁），逐字畫的話會在畫面上
			// 變成一個方框——走 `textlayout.Render` 先把控制碼吃掉。
			fs.Greeting = strings.TrimSpace(
				textlayout.Render(s.world.Block.Strings[n], textlayout.Options{}))
			fs.GreetingCJK = s.cjkLookup(
				lang.BlockKey(s.blockFile, s.blockID, n), textlayout.Options{})
		}
	}
	// 商店、醫生、訓練師三種都先問「誰要進去／誰要治療」（`sub_1721B`，
	// `docs/re/119` §1）。**一個人就不問**，原版也是直接用他。
	if asksWho(f.Kind) && s.world != nil && ableCount(s.world.Party) > 1 {
		fs.state.Step = StepWho
	}
	fs.ItemName, fs.SkillName = s.itemName, s.skillName
	fs.CJKItemName, fs.CJKSkillName = s.itemNameCJK, s.skillNameCJK
	fs.SetStock = s.setStock
	fs.Str = s.exeStringN
	fs.CJK = s.cjkExe
	fs.UI = s.uiText
	if f.Kind == game.FacilityTrainer {
		fs.Skills = s.trainableSkills()
	}
	s.facility = fs
	// 進了設施就把地圖那一步的訊息收掉。設施畫面從字元列 12 起
	// （`docs/re/54`），與訊息視窗（列 18–23）重疊——留著會有一行
	// 「TELEPORT.」壓在商品清單中間。
	s.message, s.cjk = "", ""
	// 動畫從頭起：遮罩清空、播放器重建（規格 26 §3）。
	s.animMask = make([]byte, render.FacilityPicWidth*render.FacilityPicHeight)
	s.player = nil
	if fs.Picture >= 0 && fs.Picture < len(s.anims) {
		s.player = render.NewPicPlayer(s.anims[fs.Picture])
	}
	// 進場就把主選單畫出來。原版是走進去**當下**就印
	// 「Do you want to: Buy / Sell」那一行（`docs/re/42` §1 的主迴圈
	// 開頭就在印），不是等玩家先按一個鍵——只在 `Key()` 裡 refresh 的話，
	// 玩家進去看到的是一片只有店名的畫面，不知道能按什麼。
	//
	// 角色管理（設施 3）不走這一支：它的選單是 `CREATE DELETE PLAY`，
	// 套商店那份會印出一行不存在的 Buy／Sell。
	// （`s.world` 為 nil 的是只驗版面的測試場景，那時沒有隊伍可印。）
	switch f.Kind {
	case game.FacilityShop, game.FacilityDoctor, game.FacilityTrainer:
		if s.world != nil {
			fs.refresh(s.world.Party, s.items)
		}
	}
	// 設施 3 進去就是角色管理畫面（`docs/re/72` §3）。
	// ⚠ 這一行要留在**進場的唯一入口**：兩條路（踩到 nibble 6、傳送收尾
	// 改寫腳下）都經過這裡，掛在其中一條會有一條走不到——
	// 症狀是走進 Ranger Center 之後 C／D／P 通通沒反應。
	s.enterRosterIfNeeded()
	s.dirty = true
	return fs
}

// trainableSkills 是訓練師列出來的技能。
//
// **不是每家店各有一份清單**——訓練師記錄裡只有 kind、下一步、招呼字串與
// 名稱（`docs/re/79` §2）。原版 `0x1BC30` 用清單框架列完整張技能資料表，
// 篩選（IQ 需求、費用、角色技能欄還有沒有空位）全都發生在**選完之後**：
// `0x1BC60` 選好才 `sub_1CA8D` 算費用、跟角色記錄 `+0x20` 的技能點數比。
//
// 推論等級：**強證據**（清單框架與選後檢查都直讀，「清單沒有篩選」是
// 從「篩選在選之後」推出來的）。
func (s *Scene) trainableSkills() []TrainableSkill {
	if s.rom == nil {
		return nil // 沒有映像的場景（單元測試用）就不列
	}
	raw, err := s.rom.SkillTableRaw()
	if err != nil {
		return nil
	}
	out := make([]TrainableSkill, 0, len(raw)/2)
	for i := 0; i+1 < len(raw); i += 2 {
		if i == 0 {
			// ⚠ **第 0 格不是技能**（與物品表第 0 筆同一種佔位），
			// 原版的清單第一列是 `Brawling`（實機截圖 `61-tr-menu.png`）。
			// 列出來的話畫面上會多一行沒有名字的 `skill 0`。
			continue
		}
		d := game.ParseSkillData(raw[i], raw[i+1])
		if d.BaseCost == 0 {
			continue // 費用 0 的槽不是可學的技能
		}
		out = append(out, TrainableSkill{ID: byte(i / 2), Data: d})
	}
	return out
}

// LeaveFacility 回地圖。
//
// 座標、時鐘、地圖都不用還原——踩到設施格之前隊伍就已經走完那一步了
// （規格 07 §6），設施只是接在後面跑。
// enterRosterIfNeeded 讓設施 3 走角色管理那條路（`docs/re/72` §3）。
func (s *Scene) enterRosterIfNeeded() {
	if s.facility != nil && s.facility.Facility.Kind == game.FacilityRoster {
		s.beginRoster()
	}
}

func (s *Scene) LeaveFacility() {
	s.facility = nil
	s.player, s.animMask = nil, nil
	s.dirty = true
}

// TickAnim 推進設施圖的局部動畫一拍（規格 26 §3）。
//
// 一拍 ≈ 55 ms（原版比對 BIOS 計時器 `0040:006C`）。**分頻是呈現層的事**——
// 這裡收到一次就推一拍，呼叫端自己決定多久叫一次。
// 回報畫面有沒有變，沒變就不必重畫。
func (s *Scene) TickAnim() bool {
	if s.facility == nil || s.player == nil {
		return false
	}
	elems := s.player.Tick()
	if len(elems) == 0 {
		return false
	}
	render.XorInto(s.animMask, render.FacilityPicWidth, elems)
	s.dirty = true
	return true
}

// InFacility 回報現在是不是在設施畫面。
func (s *Scene) InFacility() bool { return s.facility != nil }

// Facility 回傳目前的設施畫面（沒有就 nil）。
func (s *Scene) Facility() *FacilityScene { return s.facility }
