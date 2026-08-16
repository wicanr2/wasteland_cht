// Package play 是**可以走的遊戲場景**：把 internal/game 的規則接到畫面上。
//
// 與 internal/viewer 的分工：viewer 只負責把資產畫出來對拍，一條規則都沒有；
// play 走的是規則層——能不能進、時鐘、遭遇、事件全部照 docs/spec/04 與 07。
//
// 這個套件不依賴 Ebiten，所以無頭也跑得動（PNG 對拍用同一份場景）。
package play

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/game/rng"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/render"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// Scene 是一個可以走的世界。
type Scene struct {
	rom   *assets.Rom
	font  *assets.Font
	gfx   *render.Graphics
	world *game.World
	save  *assets.Save
	// saveDir 是 `Save` 指令要寫回的資料目錄；空的就不寫檔（見 SetSaveDir）。
	saveDir string

	frame   *render.Frame
	dirty   bool
	message string

	// eten 是倚天點陣字。沒有的話中文路徑整條關掉，遊戲照樣跑
	// （字型檔玩家自備，docs/spec/10 §4）。
	eten *assets.ETenFont
	// cjk 是這一步要顯示的中文（Big5），空的就用 message 走英文路徑。
	cjk []byte
	// cat 是翻譯目錄。沒有就整條中文關掉，遊戲跑英文（docs/spec/11 §7）。
	cat *lang.Catalogue
	// blockFile／blockID 是目前這張地圖的來源，組 key 用。
	blockFile string
	blockID   int

	// pics 是 ALLPICS1 的圖（設施畫面用 0–3，docs/re/29 §5.4）。
	pics []*assets.Indexed
	// anims 與 pics 同索引，是每張圖的局部動畫（規格 26）。
	anims []assets.PicAnim
	// player 與 animMask 只在設施模式有效：動畫是**累積 XOR**，
	// 所以要留一張至今播過的遮罩，重畫一幀時整張再疊一次。
	player   *render.PicPlayer
	animMask []byte
	// back 是傳送的回程位置（隊伍槽表 +0x0B–+0x0D，docs/re/60 §3）。
	back game.Return

	// spawn 是遭遇生成的三張表（`docs/re/78`），第一次用到才從映像讀。
	spawn   game.SpawnTables
	spawnOK bool
	// asking 非 DirNone 時畫面停在「Enter new location?」等 Y／N（docs/re/64）。
	asking input.Direction
	// roster 是 Ranger Center 角色管理畫面的狀態（`docs/re/72` §3）。
	roster rosterState

	// disband 為 true 時停在「誰要離隊」的選擇上。
	disband bool

	// encAsk 非 0 時停在 `ENC` 的「別組要不要打」Y／N 上，值 ＝ 組號 ＋ 1。
	encAsk int

	// ending 是結局播放的狀態（`docs/re/96`）。
	ending endingState

	// wipe 是全隊倒下的死亡畫面（`docs/spec/28`、`docs/re/99`）。
	wipe wipeState

	// journalHead 是手札的標題列（中文，Big5）。空的就走英文 message。
	journalHead []byte

	// placeIntro 是開場那句地點名的英文原文（存檔的全域狀態，`docs/re/30`）。
	// 留著是因為**它要在三個不同的時機重印**：建場景、載完翻譯目錄、
	// 從標題畫面按 Start 進地圖。
	placeIntro string

	// title 為 true 時停在標題畫面的主選單（`docs/re/95`）。
	title    bool
	titlePic *assets.Indexed

	// groupID 是目前操作的隊伍組（0–3，`docs/re/93` §2 的四組上限）。
	groupID int

	// order 是 `ORDER` 指令的重排狀態（`docs/re/93` §1）。
	order orderState

	// use 是 `USE` 指令的三層選單狀態（`docs/re/92`）。
	use useState

	// question 是 nibble 8 的問答（`docs/re/46` §4）：密語、暗號、控制面板。
	question questionState

	// 重製版自己加的三個面板與快速存檔（原版都沒有）。
	// **ESC 一層都不會結束遊戲，F10 是唯一的離開手勢**
	// （`esc-cancel-f10-quit-autosave`）。
	help         bool          // F1
	settingsOpen bool          // F2
	quitAsk      bool          // F10 的 Y／N
	settings     settingsState // 音樂開關與音量
	quickPath    string        // F5／F9 的存檔檔案；空 ＝ 沒開

	// portrait 是這場遭遇要顯示的敵人肖像圖編號（−1 ＝ 沒有）。
	portrait int

	// journal 是段落手札（`docs/spec/19`）。載不到就是 nil——
	// 沒有正文時遊戲照跑，只是讀不到段落。
	journal *Journal
	// journalOpen／journalAt 是手札模式的狀態（重製版的畫面，原版沒有）。
	journalOpen bool
	journalAt   int

	// sound 是下一幀要播的 PC 喇叭音效編號（`docs/re/44` §6），−1 ＝ 沒有。
	// `TakeSound` 取走就清掉，所以同一個觸發不會播兩次。
	sound int

	// items 是物品資料表（存檔區那一份，docs/re/45 §2）。
	// **武器傷害要靠它**——沒有它每個人的傷害都是 0，戰鬥永遠打不完。
	items game.ItemTable
	// itemsRaw 是同一張表的明文位元組。**庫存是遊戲狀態**（賣一件 +1），
	// 存檔時要連它一起寫回去，所以留一份原始資料在手上。
	itemsRaw []byte
	// combat 非 nil 時畫面在戰鬥（docs/spec/21）；
	// facility 非 nil 時畫面在設施（docs/spec/23）。兩者不會同時成立。
	combat   *CombatScene
	facility *FacilityScene
	// snapshot 是打之前每個角色的經驗值，收尾時相減用。
	snapshot xpSnapshot
}

// LoadJournal 載入段落手札：引用表（編譯期產物）與中文正文目錄。
//
// 兩個都可以載不到——那時讀到編號只會顯示原本那句英文，遊戲照跑
// （`docs/spec/19` §6：**沒翻的段落不能變成空白頁**）。
func (s *Scene) LoadJournal(refsPath, catPath string) error {
	refs, err := game.LoadParagraphRefs(refsPath)
	if err != nil {
		return err
	}
	s.journal = NewJournal(refs, nil)
	return s.journal.LoadParagraphs(catPath)
}

// LoadCatalogue 載入翻譯目錄；載不到就維持英文，不當成錯誤。
func (s *Scene) LoadCatalogue(path string) error {
	c, err := lang.Load(path)
	if err != nil {
		return err
	}
	s.cat = c
	s.dirty = true
	// ⚠ **順序**：`New` 比 `LoadCatalogue` 早跑，所以開場那句地點名在建場景
	// 的時候還查不到中文（目錄還沒載）。載完目錄要回頭補一次，
	// 否則畫面第一句永遠是英文——而那看起來像「這一句沒翻」。
	//
	// ⚠ 標題畫面時不補：那時訊息視窗不該有字，補上去會印在標題圖上。
	// 進地圖那一下由 `updateTitle` 補（`docs/re/95`）。
	if !s.title {
		s.sayPlace()
	}
	return nil
}

// sayPlace 印開場那句地點名：查得到中文就印中文，查不到印英文原文。
//
// ⚠ **不要用「現在的訊息是不是等於地點名」當條件。** 那個寫法在
// `cmd/wasteland` 的順序（`New` → `BeginTitle` → `LoadCatalogue`）下不成立——
// `BeginTitle` 已經把訊息清空了，於是整句話消失。
// 而 `wl-shot` 的順序剛好相反，截圖看起來完全正常：**這個 bug 只在
// 真的開視窗玩的時候看得到**。所以這裡改成看自己記下來的 placeIntro。
func (s *Scene) sayPlace() {
	if s.placeIntro == "" {
		return
	}
	if zh := s.placeCJK(s.placeIntro); len(zh) > 0 {
		s.message, s.cjk = "", zh
	} else {
		s.message, s.cjk = s.placeIntro, nil
	}
	s.dirty = true
}

// LoadFont 載入倚天字型；載不到就維持英文，不當成錯誤。
func (s *Scene) LoadFont(dir string) error {
	f, err := assets.LoadETen(dir)
	if err != nil {
		return err
	}
	s.eten = f
	s.dirty = true
	return nil
}

// SetCJK 設定這一步要顯示的中文訊息（Big5）。翻譯目錄接上之前，
// 這是讓呈現層拿到中文的唯一入口——規則層永遠只給編號。
func (s *Scene) SetCJK(b []byte) {
	s.cjk = b
	s.dirty = true
}

// HiFrame 合成 640 × 400 的畫面：原版素材 nearest 2× 放大，
// 中文用倚天 16 × 15 直繪（docs/spec/10 §2）。
//
// 一個中文字剛好佔原版一個字元格，所以訊息視窗仍然是 6 行 × 38 格。
// cjkVisible 回報這一幀會不會畫中文（有字型 ＋ 有內容）。
func (s *Scene) cjkVisible() bool { return s.eten != nil && len(s.cjk) > 0 }

// hiText 回報「文字層要不要整層改走高解畫布」。
//
// 條件是**有倚天的半形 ASCII 字模**：那時英數與中文同高、同一套設計，
// 整層搬上去才會齊。只有漢字沒有 ASCFONT 時維持原樣——
// 那時英數只能用遊戲原版的 8 × 8 放大，搬上去也不會比較好看。
//
// ⚠ 開了這個之後，低解那張 **不可以再畫同一批字**：先畫再蓋會留殘影
// （8 × 8 放大的筆劃比 24 點的粗，蓋不乾淨）。
func (s *Scene) hiText() bool { return s.eten != nil && s.eten.HasASCII() }

// drawASCIILine 在高解畫面上畫一整行純 ASCII。
func (s *Scene) drawASCIILine(h *render.HiFrame, text string, col, row int) {
	for i := 0; i < len(text); i++ {
		if text[i] != ' ' {
			s.drawASCII(h, text[i], col+i, row)
		}
	}
}

func (s *Scene) HiFrame() *render.HiFrame {
	h := render.NewHiFrame()
	h.Upscale(s.Frame())
	if s.eten == nil {
		return h
	}
	// ⚠ 這一行要與 `Frame` 的條件一致，**再加上標題畫面**：`Frame` 在標題那一支
	// 提早 return，這裡卻是照著模式旗標判斷，漏掉 `s.title` 就會把指令列
	// 畫到標題畫面上，還會蓋掉同一列的 `Start`（`docs/re/95`）。
	if !s.title && !s.wipe.active && s.facility == nil && s.combat == nil &&
		!s.ending.active {
		s.drawCJKLine(h, s.uiText("cmd.bar"), 0, render.CmdRow)
	}
	if s.journalOpen && len(s.journalHead) > 0 {
		s.drawCJKLine(h, s.journalHead, render.MsgCol, render.MsgRow)
	}
	// 死亡畫面的地點名那一行（與設施招牌同一個位置）。
	if s.wipe.active {
		if zh := s.wipePlaceCJK(); len(zh) > 0 {
			s.drawCJKLine(h, zh, render.FacilityNameCol, render.FacilityNameRow)
		}
	}
	// 戰鬥名單的表頭：原版是寫死的 ASCII，中文走 `ui:combat.hdr*`
	// （與指令列同一條路，8 × 8 的字模畫不出中文）。
	if s.combat != nil {
		if hdr := rosterHeaderCJK(s.uiText); len(hdr) > 0 {
			s.drawCJKLine(h, hdr, 0, render.RosterHeaderRow)
		}
	}
	// 設施畫面那幾行（店名、選單、清單）走自己的座標，不在訊息視窗裡。
	if s.facility != nil {
		for i, l := range s.facility.CJKLines {
			if len(l) == 0 {
				continue // 這一行沒有中文，由 drawFacility 畫英文
			}
			s.drawCJKLine(h, l, render.FacilityNameCol, render.FacilityNameRow+i)
		}
	}
	s.drawHiTextLayer(h)
	if len(s.cjk) == 0 {
		return h
	}
	// 英文訊息占掉第一行時，中文從第二行起——不要疊上去。
	col, row := render.MsgCol, render.MsgRow
	if s.message != "" || len(s.journalHead) > 0 {
		row++
	}
	// ⚠ **不能整串兩兩配對。** 譯文裡會夾 ASCII（人名、數字、標點），
	// 把它當成 Big5 的高位元組會讓**之後整行都錯位**——症狀是畫面上
	// 出現一串看得懂筆畫卻不成字的東西，而不是空白。
	for i := 0; i < len(s.cjk); {
		if col >= render.MsgCol+render.MsgWidth {
			col = render.MsgCol
			row++
		}
		if row > render.MsgRowEnd {
			break // 訊息視窗滿了；分頁是控制碼的事（docs/re/14 §4）
		}
		c := s.cjk[i]
		switch {
		case c == '\r' || c == '\n':
			// 原版的斷行控制碼：換一行、不佔格。
			col = render.MsgCol
			row++
			i++
			continue
		case c < 0x80:
			s.drawASCII(h, c, col, row)
			i++
		default:
			if i+1 >= len(s.cjk) {
				i++ // 落單的高位元組：跳過，不要拿下一輪的 byte 湊
				continue
			}
			h.DrawCJK(s.eten, s.cjk[i], s.cjk[i+1], col, row, 15)
			i += 2
		}
		col++
	}
	return h
}

// drawHiTextLayer 把原本畫在低解那張上的英文與數字，改用倚天半形字模
// 畫在高解畫布上——**這是英數與中文對齊的那一半**。
//
// 對應關係要與 `Frame` 一一對上：那邊在 `hiText()` 為真時跳過的每一項，
// 這裡都要補回來。少補一項的症狀是「那一行不見了」。
func (s *Scene) drawHiTextLayer(h *render.HiFrame) {
	if !s.hiText() {
		return
	}
	// 時鐘：外框上緣，切模式不影響它（`docs/re/27` §4）。
	if !s.ending.active && !s.wipe.active {
		c := s.world.Clock
		s.drawASCIILine(h, fmt.Sprintf("%02d:%02d", c.Hour, c.Minute),
			render.ClockCol, render.ClockRow)
	}
	// 戰鬥名單那幾行（表頭是中文，走另一條）。
	if s.combat != nil {
		for i, r := range Roster(s.world.Party) {
			row := render.RosterHeaderRow + 1 + i
			if row > render.MsgRow-1 {
				break // 名單與訊息視窗在字元列上會撞（docs/spec/03 §3）
			}
			s.drawASCIILine(h, r.Text(), 0, row)
		}
	}
	// 設施那幾行裡沒有中文的（店名以外的清單多半是英文物品名）。
	if s.facility != nil {
		for i, l := range s.facility.Lines {
			if i < len(s.facility.CJKLines) && len(s.facility.CJKLines[i]) > 0 {
				continue
			}
			s.drawASCIILine(h, l, render.FacilityNameCol, render.FacilityNameRow+i)
		}
	}
	// 訊息視窗的英文。有中文正文時只留第一行當標題（與 `Frame` 同一條規則）。
	if s.message == "" {
		return
	}
	out, err := textlayout.Layout([]byte(s.message), textlayout.Options{Width: render.MsgWidth})
	if err != nil {
		return
	}
	lines := out.Lines
	if s.cjkVisible() && len(lines) > 1 {
		lines = lines[:1]
	}
	for i, l := range lines {
		if render.MsgRow+i > render.MsgRowEnd {
			break
		}
		s.drawASCIILine(h, l.String(), render.MsgCol, render.MsgRow+i)
	}
}

// New 從出廠存檔開一個場景：挑序號大的那一份、讀出隊伍與所在地圖。
func New(rom *assets.Rom) (*Scene, error) {
	font, err := rom.FontMain()
	if err != nil {
		return nil, err
	}

	a, err := rom.LoadSave("game1")
	if err != nil {
		return nil, err
	}
	b, err := rom.LoadSave("game2")
	if err != nil {
		return nil, err
	}
	save := assets.PickNewer(a, b)

	party, mapID, err := loadParty(save)
	if err != nil {
		return nil, err
	}
	clock := loadClock(save)

	block, err := rom.BlockByID(mapID)
	if err != nil {
		return nil, fmt.Errorf("載入地圖 %d：%w", mapID, err)
	}
	tiles, err := rom.Tileset(block.Tileset)
	if err != nil {
		return nil, err
	}
	icons, err := rom.Icons()
	if err != nil {
		return nil, err
	}
	masks, err := rom.Masks()
	if err != nil {
		return nil, err
	}

	s := &Scene{
		rom:   rom,
		font:  font,
		gfx:   &render.Graphics{Icons: icons, Masks: masks, Tiles: tiles},
		world: game.NewWorld(block, party, rng.New()),
		save:  save,
		dirty:    true,
		sound:    -1,
		portrait: -1,
	}
	s.world.Clock = clock
	// 手札預設是空的（沒有引用表、沒有正文），由呼叫端 LoadJournal 載入——
	// **路徑不寫死在這裡**，與 LoadCatalogue／LoadFont 一致。
	s.journal = NewJournal(game.ParagraphRefs{}, nil)
	// 物品表跟著存檔走（每個存檔槽一份）。載不到就維持空表——
	// 傷害會是 0，但遊戲跑得動，而且下面這行的錯誤會留在訊息裡。
	// 設施畫面的圖。載不到就不畫圖，其餘照跑。
	// ALLPICS 是**兩個檔一個編號空間**：allpics1 有 33 張、allpics2 有 49 張，
	// 相加正好是 `docs/re/23` §4 的 82 張——所以 allpics1 在前、allpics2 接在後面。
	//
	// ⚠ 分界點是**張數相加推出來的（強證據）**，沒有從 `sub_184E8` 的選檔邏輯
	// 直讀。設施圖的編號落在 allpics1 那一段，所以先前只載一個檔也看不出問題；
	// 敵人肖像會用到 44 這種編號，載一個檔就是空白。
	if pics, err := rom.Pictures("allpics1"); err == nil {
		s.pics = pics
	}
	if pics2, err := rom.Pictures("allpics2"); err == nil {
		s.pics = append(s.pics, pics2...)
	}
	if anims, err := rom.PictureAnims("allpics1"); err == nil {
		s.anims = anims
	}
	// 技能資料表：條件閘的技能型別要用（docs/re/32 §2）。
	// 載不到就讓技能型別的條件一律失敗，其餘照跑。
	if raw, err := rom.SkillTableRaw(); err == nil {
		s.world.Skills = game.SkillBytes(raw)
	}
	s.placeIntro = save.Place()
	s.sayPlace()
	if raw, err := rom.LoadItemTable(save.File, 0); err == nil {
		s.items, s.itemsRaw = game.ParseItemTable(raw), raw
	} else {
		// 載不到就維持空表：傷害會是 0，但遊戲跑得動。
		// **錯誤要留在畫面上**，不要靜靜吞掉——那會變成「戰鬥打不完」的怪症狀。
		s.message = "ITEM TABLE: " + err.Error()
	}
	s.asking = input.DirNone // 零值是 DirUp，會被誤判成「正在問」
	s.settings = defaultSettings()
	s.blockFile, s.blockID = block.Resource.File, block.Resource.ID
	// 回程從存檔的隊伍槽表讀（+0x0B–+0x0D，docs/re/60 §3）。
	g := save.SlotGroups()[0]
	s.back = game.Return{X: g.BackX, Y: g.BackY, MapID: g.BackMap}
	return s, nil
}

// loadClock 從全域狀態的 ds:464Eh–465Bh 副本讀出時鐘（docs/spec/05 §3）。
//
//	相對位移 2–4  → 32-bit 累計的低三個 byte
//	相對位移 10   → 分的小數
//	相對位移 11   → 分
//	相對位移 12   → 時
func loadClock(save *assets.Save) game.Clock {
	g := save.Globals()
	return game.Clock{
		Frac:   g[10],
		Minute: g[11],
		Hour:   g[12],
		Total:  uint32(g[2]) | uint32(g[3])<<8 | uint32(g[4])<<16,
	}
}

// loadParty 從存檔的全域狀態與角色記錄建出隊伍。
func loadParty(save *assets.Save) (*game.Party, int, error) {
	return loadPartyGroup(save, 0) // 出廠只有第 0 組有人（docs/spec/05 §3.1）
}

// loadPartyGroup 載入第 n 組隊伍。
//
// 每一組槽表**各自帶座標與地圖**（`docs/spec/05` §3.1）——切組不是只換人，
// 是換到那一組所在的地方（`docs/re/93` §3）。
func loadPartyGroup(save *assets.Save, n int) (*game.Party, int, error) {
	groups := save.SlotGroups()
	if n < 0 || n >= len(groups) {
		return nil, 0, fmt.Errorf("隊伍組 %d 超出範圍（共 %d 組）", n, len(groups))
	}
	g := groups[n]

	p := &game.Party{X: g.X, Y: g.Y, Selected: 0}
	for _, id := range g.Members {
		if id == 0 {
			continue
		}
		raw, err := save.Record(int(id))
		if err != nil {
			return nil, 0, fmt.Errorf("角色記錄 %d：%w", id, err)
		}
		p.Members = append(p.Members, game.LoadCharacter(raw))
	}
	if len(p.Members) == 0 {
		return nil, 0, fmt.Errorf("存檔裡的第 %d 組隊伍是空的", n)
	}
	return p, int(g.MapID), nil
}

// LoadMap 換一張地圖並把隊伍放到指定座標。
//
// 給驗證工具用（`cmd/wl-play` 的 `map=N`）：起始地圖沒有靜態遭遇格，
// 要驗戰鬥流程得換到有的那幾張（`docs/re/51`）。
func (s *Scene) LoadMap(id int, x, y uint8) error {
	// **用 ID 不是切片索引**：遊戲裡的地圖編號是資源目錄的 ID（docs/re/63）。
	b, err := s.rom.BlockByID(id)
	if err != nil {
		return err
	}
	tiles, err := s.rom.Tileset(b.Tileset)
	if err != nil {
		return err
	}
	s.gfx.Tiles = tiles
	s.world.EnterMap(b, x, y)
	s.blockFile, s.blockID = b.Resource.File, b.Resource.ID
	s.dirty = true
	return nil
}

// playSound 排一個音效給呈現層。
//
// 一幀只留最後一個：原版的呼叫端之間隔著計時器中斷，同一幀連觸發兩個
// 在原版是聽不到前一個的。
func (s *Scene) playSound(n int) { s.sound = n }

// TakeSound 取走這一幀要播的音效編號（`ui.Sounder`），−1 ＝ 沒有。
func (s *Scene) TakeSound() int {
	n := s.sound
	s.sound = -1
	return n
}

// Asking 回報畫面是不是停在「Enter new location?」等 Y／N。
func (s *Scene) Asking() bool { return s.asking != input.DirNone }

// PollRNG 推進一次亂數產生器，等同原版的鍵盤輪詢（規格 02 §1.1）。
//
// ⚠ **這是熵的唯一來源。** 原版 `sub_18EFE` 每輪詢一次就推進一次，
// 所以序列取決於玩家花多久按鍵；不叫它的話**每一局的遭遇序列會完全相同**
// （產生器沒有種子，初值全零）。呈現層每幀叫一次。
//
// 無頭工具（`cmd/wl-shot`、`cmd/wl-play`）刻意不叫，保持可重現。
func (s *Scene) PollRNG() {
	if w := s.world; w != nil && w.RNG != nil {
		w.RNG.Next()
	}
}

// Message 是訊息視窗這一步顯示的字（英文路徑）。
// 中文走 cjk，兩者不會同時空著。
func (s *Scene) Message() string { return s.message }

// CJK 是這一步的中文訊息（Big5）。空的表示這一步走英文路徑。
// 給驗收工具看的——畫面自己直接讀內部欄位。
func (s *Scene) CJK() []byte { return s.cjk }

// Mode 是畫面現在停在哪一層，順序與 Update 的路由完全一致
// （`docs/spec/24`）。**驗收工具用**：訊息列看不出「進了設施但選單沒開」
// 這種接線斷掉的情況，模式看得出來。
func (s *Scene) Mode() string {
	switch {
	case s.quitAsk:
		return "quit"
	case s.wipe.active:
		return "wipe"
	case s.help:
		return "help"
	case s.settingsOpen:
		return "settings"
	case s.title:
		return "title"
	case s.ending.active:
		return "ending"
	case s.roster.active:
		switch {
		case s.roster.naming:
			return "roster:naming"
		case s.roster.del:
			return "roster:delete"
		}
		return "roster"
	case s.facility != nil:
		return "facility:" + s.facility.Facility.Name
	case s.combat != nil:
		return "combat"
	case s.journalOpen:
		return fmt.Sprintf("journal:%d", s.journalAt)
	case s.use.stage != useStageOff:
		return fmt.Sprintf("use:%d", int(s.use.stage))
	case s.question.active:
		if s.question.q.SingleKey {
			return "question:key"
		}
		return "question:typed"
	case s.order.active:
		return "order"
	case s.disband:
		return "disband"
	case s.encAsk != 0:
		return "encask"
	case s.asking != input.DirNone:
		return "asking"
	}
	return "map"
}

// World 讓測試與 cmd/wl-shot 拿得到規則層的狀態。
func (s *Scene) World() *game.World { return s.world }

// Invalidate 讓下一次 Frame 重畫。外部直接改了世界狀態時要叫它。
func (s *Scene) Invalidate() { s.dirty = true }

// Update 收一幀的輸入，依目前的模式路由（docs/spec/24）。
//
// ⚠ **模式中的按鍵不要「順便」轉給地圖**：原版在名單模式下方向鍵不走路
// （`sub_17FEE` 在旗標非 0 時擋住地圖繪製，docs/re/25 §2.5）——
// 轉下去會變成「在戰鬥裡走路」。
func (s *Scene) Update(in input.Input) (bool, error) {
	// 離開確認蓋在所有模式上面：**它自己收 Y／N，其餘按鍵一律不往下傳**。
	if s.quitAsk {
		return s.updateQuit(in)
	}
	// F10 是唯一的離開手勢，而且**不直接離開**——先問一句、答 Y 才先存檔再退出
	// （`esc-cancel-f10-quit-autosave` 鐵則 2、3、4）。
	if in.Action == input.ActionQuit {
		s.openQuit()
		return true, nil
	}
	if s.help {
		return s.updateHelp(in)
	}
	if s.settingsOpen {
		return s.updateSettings(in)
	}
	// 功能鍵不管在哪一層都叫得出來（與 ESC 的「退一層」語意分開）。
	switch in.Fn {
	case input.FnHelp:
		s.openHelp()
		return true, nil
	case input.FnSettings:
		s.openSettings()
		return true, nil
	case input.FnQuickSave:
		return s.doQuickSave()
	case input.FnQuickLoad:
		return s.doQuickLoad()
	}
	if s.title {
		return s.updateTitle(in)
	}
	if s.ending.active {
		return s.updateEnding(in)
	}
	if s.wipe.active {
		return s.updateWipe(in)
	}
	if s.roster.active {
		return s.updateRoster(in)
	}
	if s.facility != nil {
		return s.updateFacility(in)
	}
	if s.combat != nil {
		return s.updateCombat(in)
	}
	if s.journalOpen {
		return s.updateJournal(in)
	}
	if s.use.stage != useStageOff {
		return s.updateUse(in)
	}
	if s.question.active {
		return s.updateQuestion(in)
	}
	if s.order.active {
		return s.updateOrder(in)
	}
	if s.disband {
		return s.updateDisband(in)
	}
	if s.encAsk != 0 {
		return s.updateEncAsk(in)
	}
	ok, err := s.updateMap(in)
	if !ok || err != nil {
		return ok, err
	}
	// 主迴圈每一輪檢查一次自毀倒數（`0x16C28` 的 `call sub_1CB30`，
	// 位置就在全隊陣亡那道檢查前面）。**結局是這樣進來的**，
	// 不是踩到某一格——那一格只負責啟動倒數（`docs/re/100`）。
	if s.world.SelfDestructDue() {
		s.world.DisarmSelfDestruct()
		s.BeginEnding()
		return true, nil
	}
	// 全隊倒下的三分支（`0x16C2B`，就接在自毀那道後面）。
	s.checkWipe()
	return true, nil
}

// updateFacility 是設施模式：只有離開。買賣與治療的選單這一版不接
// （docs/spec/23 §5、docs/spec/24 §5）。
func (s *Scene) updateFacility(in input.Input) (bool, error) {
	k := byte(0)
	switch {
	case in.Action == input.ActionCancel:
		k = 0x1B // ESC：清單 → 主選單 → 離開，一層一層退
	case in.Char != 0:
		k = input.Upper(in.Char)
	}
	if k == 0 {
		return true, nil
	}
	if !s.facility.Key(k, s.world.Party, s.items) {
		s.LeaveFacility()
		s.message = ""
	}
	s.dirty = true
	return true, nil
}

// updateCombat 是戰鬥模式：逐人問指令，問完就結算一回合。
func (s *Scene) updateCombat(in input.Input) (bool, error) {
	c := s.combat
	if in.Action == input.ActionCancel {
		// ESC 退回上一個人（docs/spec/14）。已經在第一個就整場不動。
		return true, nil
	}
	if in.Char != 0 && !c.Done() {
		// armed：裝備欄還沒解到能判斷（docs/spec/22 §5），一律當成有武器。
		c.Choose(input.Upper(in.Char), true)
		s.dirty = true
	}
	if c.Done() {
		res := c.ResolveRound()
		s.dirty = true
		if n := len(res.Lines); n > 0 {
			s.message = res.Lines[n-1]
			s.cjk = nil
			if len(res.CJK) == n {
				s.cjk = res.CJK[n-1]
			}
		}
		if res.Over {
			out := s.FinishEncounter()
			s.message, s.cjk = s.combatOverMessage(res.Won, out)
			return true, nil
		}
		c.BeginCommands()
	}
	// 指令階段：把「換誰下令」那一行與選單畫出來。
	// 沒有這一步玩家看不到能按什麼——原版每問一個人就重印一次（`docs/re/40` §1）。
	s.showCombatPrompt()
	return true, nil
}

// showCombatPrompt 把目前這個人的提示與指令選單放進訊息視窗。
//
// ⚠ **排法是重製決策**：原版每個選項自成一行、標題加空行共九行，
// 而訊息視窗只有六行——「原版怎麼容納」還沒逆向（`docs/re/40` §5）。
// 這裡把七個選項排成一行（中文剛好放得下 38 格），與地圖指令列同一種做法。
// 熱鍵字母照 `docs/re/40` §4.1 留在每個選項前面，不跟著翻譯走。
func (s *Scene) showCombatPrompt() {
	c := s.combat
	if c == nil || c.Done() || c.Turn < 0 || c.Turn >= len(c.Battle.Party.Members) {
		return
	}
	m := c.Battle.Party.Members[c.Turn]
	if m == nil {
		return
	}
	en := m.Name + ", choose: " + strings.Join(MenuLines(CommandMenu(nil)), " ")
	// 中文：名字 ＋ 字串 55（`, 選擇：` ＋ 七個 `\x10<文字>`）。
	// `RenderBytes` 會把 `\x10` 拿掉、把 `\x0D` 變成換行，熱鍵字母留著；
	// 換行再壓成空白排成一行（見上面的重製決策）。
	var zh []byte
	if b := c.zhStr(strChoose, textlayout.Options{}); b != nil {
		zh = append([]byte(m.Name), oneLine(b)...)
	}
	// 這一輪已經有話要說（遭遇開始、上一回合的結果）就接在後面，
	// **不要蓋掉**——玩家要同時看到發生什麼與能按什麼。
	switch {
	case zh != nil && len(s.cjk) > 0:
		s.cjk = append(append(append([]byte{}, s.cjk...), '\n'), zh...)
		s.message = ""
	case zh != nil:
		s.cjk, s.message = zh, ""
	case s.message != "":
		s.message += " " + en
	default:
		s.message = en
	}
	s.dirty = true
}

// oneLine 把換行壓成空白，並去掉頭尾空白與連續空白。
//
// 指令選單在原版是**每個選項一行**（七行），而訊息視窗只有六行、
// 原版怎麼容納還沒逆向（`docs/re/40` §5）。中文一個字一格，
// 七個選項排成一行是 35 格左右，38 格的視窗放得下。
func oneLine(b []byte) []byte {
	out := make([]byte, 0, len(b))
	space := true // 開頭的空白一律吃掉
	for _, c := range b {
		if c == '\n' || c == '\r' || c == ' ' {
			if !space {
				out = append(out, ' ')
				space = true
			}
			continue
		}
		out = append(out, c)
		space = false
	}
	for len(out) > 0 && out[len(out)-1] == ' ' {
		out = out[:len(out)-1]
	}
	return out
}

// combatOverMessage 是戰鬥結束那一行（英文 ＋ 中文）。
//
// 三句都是重製版自己的收尾（原版打完是回地圖、經驗值逐人前後相減報，
// `docs/re/51` §3），所以走 `ui:`。
func (s *Scene) combatOverMessage(won bool, out EncounterResult) (string, []byte) {
	if !won {
		return "The party has fallen.", s.uiText("combat.fallen")
	}
	total := uint32(0)
	for _, xp := range out.XPGained {
		total += xp
	}
	if total == 0 {
		return "The battle is over.", s.uiText("combat.over")
	}
	en := fmt.Sprintf("The party gains %d experience.", total)
	var zh []byte
	if f := s.uiText("combat.partyxp"); len(f) > 0 {
		zh = []byte(fmt.Sprintf(string(f), total))
	}
	return en, zh
}

// updateMap 是地圖模式：方向鍵走一步。
func (s *Scene) updateMap(in input.Input) (bool, error) {
	// 停在「進新地點？」的確認上時，只收 Y／N（原版是同步問，docs/re/64 §1）。
	if s.asking != input.DirNone {
		switch input.Upper(in.Char) {
		case 'Y':
			d := s.asking
			s.asking = input.DirNone
			s.world.Confirm()
			return s.walk(gameDir(d))
		case 'N':
			s.asking = input.DirNone
			s.message = ""
			s.dirty = true
		}
		if in.Action == input.ActionCancel {
			s.asking = input.DirNone
			s.message = ""
			s.dirty = true
		}
		return true, nil
	}

	var dir game.Direction
	switch in.Dir {
	case input.DirUp:
		dir = game.Up
	case input.DirDown:
		dir = game.Down
	case input.DirLeft:
		dir = game.Left
	case input.DirRight:
		dir = game.Right
	default:
		// 不是方向就看是不是指令列的一項（`docs/re/91`）。
		// **順序不能顛倒**：原版的 IKJL 是方向鍵，而指令的首字母
		// （U E O D V S R）與它們不重疊，所以先問方向再問指令。
		if input.Upper(in.Char) == JournalKey {
			s.openJournal(s.journalAt)
			return true, nil
		}
		if c := CommandFor(in.Char); c >= 0 {
			return s.runCommand(c)
		}
		return true, nil
	}

	return s.walk(dir)
}

// gameDir 是 dirOf 的反向。
func gameDir(d input.Direction) game.Direction {
	switch d {
	case input.DirUp:
		return game.Up
	case input.DirDown:
		return game.Down
	case input.DirLeft:
		return game.Left
	default:
		return game.Right
	}
}

// dirOf 把規則層的方向換回輸入層的方向（確認流程要記住原本按了哪個鍵）。
func dirOf(d game.Direction) input.Direction {
	switch d {
	case game.Up:
		return input.DirUp
	case game.Down:
		return input.DirDown
	case game.Left:
		return input.DirLeft
	default:
		return input.DirRight
	}
}

// exeString 取執行檔字串表 1 的第 n 條（`sub_16CB2` 那一族用的就是這張）。
// nameString 查「技能、物品、介面」那張表（`ds:B270h` ＝ 第 2 張，`docs/re/17` §4）。
//
// ⚠ **不是 `exeString` 那張**：`exeString` 走的是第 1 張（無線電、隊伍、戰鬥）。
// 兩張表的索引空間完全不同，拿錯會安靜地取到別的句子——
// 技能 1 在這張是 `Brawling`，在那張是 `Radio?YesNo`。
func (s *Scene) nameString(n int) string {
	if s.rom == nil {
		return "" // 沒有映像的場景（單元測試用）查不到名字
	}
	tables, err := s.rom.ExeStrings()
	if err != nil || len(tables) < 3 || n < 0 || n >= len(tables[2]) {
		return ""
	}
	return tables[2][n]
}

func (s *Scene) exeString(n int) string { return s.exeStringN(exeTable1, n) }

// exeStringN 取執行檔第 table 張字串表的第 n 條（`docs/re/17` §3 的九張表）。
//
// 設施、角色管理、技能那幾組訊息**不在表 1**（商店 7、醫生 8、
// 角色管理 3、訓練 6），拿表 1 去查會安靜地取到別的句子。
func (s *Scene) exeStringN(table, n int) string {
	if s.rom == nil {
		return ""
	}
	tables, err := s.rom.ExeStrings()
	if err != nil || table < 0 || table >= len(tables) || n < 0 || n >= len(tables[table]) {
		return ""
	}
	return tables[table][n]
}

// walk 真正走一步並處理結果。
func (s *Scene) walk(dir game.Direction) (bool, error) {
	res, err := s.world.Step(dir)
	if err != nil {
		return true, err
	}
	// 停下來問「Enter new location?」：不移動、不推進時間。
	// **先印目標地點名**（`0x16AEA` 的 `sub_16B17`）——原版兩句一起出現，
	// 只印問句會少掉「進入石英城。」那一行。
	if res.Ask != 0 {
		s.asking = dirOf(dir)
		s.say(res.Ask, textlayout.Options{})
		if n := res.AskString; n > 0 && n < len(s.world.Block.Strings) {
			zh := s.cjkLookup(lang.BlockKey(s.blockFile, s.blockID, n), textlayout.Options{})
			// 有中文就只加中文——`sayT` 已經把英文那一份清掉了，
			// 兩份疊上去訊息視窗六行放不下（`docs/spec/10` §2）。
			if s.cjk != nil && zh != nil {
				s.cjk = append(append([]byte{}, zh...), s.cjk...)
			} else if s.cjk == nil {
				s.message = s.world.Block.Strings[n] + s.message
			}
		}
		return true, nil
	}
	s.dirty = true
	s.message = s.describe(res)
	s.cjk = s.translate(res)
	// 訊息裡引用了段落就把正文顯示出來（`docs/spec/19` §3）。
	// 引用表的 key 與翻譯目錄同一組編號。
	for _, n := range s.stringSlots(res) {
		if n > 0 && s.maybeParagraph(lang.BlockKey(s.blockFile, s.blockID, n)) {
			break
		}
	}

	// 踩上 nibble 8 就進問答（密語、暗號、控制面板，`docs/re/46` §4）。
	if res.Moved && res.Event.Kind == game.EventMenu {
		s.beginQuestion(res.Event.Data)
	}

	// 走一步的點擊聲（音效 1，`0x16575` 在 `sub_1656D` 裡，docs/re/44 §6）。
	// **只有真的移動了才響**——被擋住時原版走的是別條路。
	if res.Moved {
		s.playSound(1)
	}
	// 腳本指令自己要求的音效（op 26／35 走 `sub_1142B` 播 7）蓋過腳步聲。
	if res.Script.Sound >= 0 {
		s.playSound(res.Script.Sound)
	}

	// 條件閘的收尾訊息：通過印記錄 +0x02、沒過且沒人受罰印 +0x03（docs/re/69）。
	if res.Gate.Message > 0 && res.Gate.Message < len(s.world.Block.Strings) {
		s.message = s.world.Block.Strings[res.Gate.Message]
	}
	// 罰到人就照原版報一句（執行檔字串表 1 第 99 條 " gets hurt for "）。
	if len(res.Gate.Failed) > 0 {
		if line := s.gateHurtLine(res.Gate); line != "" {
			s.message = line
			if zh := s.gateHurtCJK(res.Gate); zh != nil {
				s.message, s.cjk = "", zh
			}
		}
	}

	// 踩到傳送格就換地圖（docs/spec/07 §6.7）。**這一步到此為止**——
	// 原版換完地圖回 CF ＝ 1，不捲動也不掃遭遇。
	if res.Moved && res.Event.Kind == game.EventTeleport {
		if err := s.doTeleport(res.Event.Data); err != nil {
			s.message = "ERROR: " + err.Error()
			return true, nil
		}
		// 傳送收尾把腳下改寫成 nibble 6 之後，那一格的事件要跑起來
		// （原版 `loc_16A90` 尾端的 `sub_15DAF`）——**商店與醫生就是這樣進去的**。
		// 設施記錄的 `+0x01`／`+0x02` 是 `fd fd`（不改），所以不會繞回傳送格。
		s.enterFacilityHere()
		return true, nil
	}

	// 踩到輻射格就結算（docs/spec/07 §6.4）。訊息已經在 describe 裡了，
	// 這裡只補傷害那一句——**扣血是規則層做的，這裡不重算**。
	if res.Moved && res.Event.Kind == game.EventRadiation {
		if hits := s.world.ApplyRadiation(res.Event.Data); len(hits) > 0 {
			total := 0
			for _, h := range hits {
				total += h.Applied
			}
			if total > 0 {
				s.message = fmt.Sprintf("%s (%d damage)", s.message, total)
			}
		}
	}

	// 踩到設施格就進設施畫面（docs/spec/23）。
	// bit7 沒設的是腳本指令，`EnterFacility` 會回 nil——那條路不歸這裡管。
	if res.Moved && res.Event.Kind == game.EventFacility {
		s.EnterFacility(res.Event.Data)
	}

	// 走一步先跑**遭遇生成**（`sub_16890`，docs/re/78）：擲中就在附近的空地
	// 放一格 nibble 15。沒有這一步，下面的掃描永遠掃不到東西——
	// 出貨地圖上一格敵人都沒有。
	if res.Moved && !s.InFacility() {
		if tbl, ok := s.spawnTables(); ok {
			s.world.SpawnEncounter(tbl)
		}
	}

	// 走一步之後掃遭遇（docs/re/51 §2）。掃描說沒有可打的就什麼都不做——
	// **擲骰說「觸發」不等於真的打得起來**，還要視窗裡有敵人格、
	// 距離過得了記錄的兩道門檻（docs/spec/15）。
	if res.Moved && res.Encounter && !s.InFacility() {
		if c, err := s.StartEncounter(); err != nil {
			s.message = "ERROR: " + err.Error()
		} else if c != nil {
			s.sayAttacked()
			s.showCombatPrompt()
		}
	}
	return true, nil
}

// spawnTables 是遭遇生成的三張表，從執行檔映像讀一次就好。
func (s *Scene) spawnTables() (game.SpawnTables, bool) {
	if s.spawnOK {
		return s.spawn, true
	}
	raw, err := s.rom.SpawnTablesRaw()
	if err != nil {
		return game.SpawnTables{}, false
	}
	copy(s.spawn.Near[:], raw[0])
	copy(s.spawn.Far[:], raw[1])
	copy(s.spawn.Dist[:], raw[2])
	s.spawnOK = true
	return s.spawn, true
}

// enterFacilityHere 在傳送收尾之後檢查腳下那一格，是設施就進去。
//
// 傳送記錄的 `+0x04`／`+0x05` 把落點改寫成 (nibble 6, 設施記錄)，
// 22 筆設施——商店、醫生、圖書館、訓練——全部靠這一步（docs/re/73）。
func (s *Scene) enterFacilityHere() {
	w := s.world
	terrain, idx, _, err := w.Block.At(int(w.Party.X), int(w.Party.Y))
	if err != nil || terrain != 6 {
		return
	}
	rec, err := w.Block.SectionRecord(6, int(idx))
	if err != nil || len(rec) == 0 || rec[0]&0x80 == 0 {
		return // bit7 沒設的是腳本，不是設施
	}
	s.EnterFacility(rec)
}

// describe 把一步的結果變成訊息視窗要顯示的字。
// 規則層只給編號，文字在這裡才查出來——中文化改的是這一層與翻譯目錄。
func (s *Scene) describe(res game.StepResult) string {
	if !res.Moved {
		// 原版印的是那一格記錄 +0x00 指的訊息（`This mountain is in your way.`），
		// 不是一句固定的 BLOCKED（docs/re/62 §2）。
		if n := res.Blocked; n > 0 && n < len(s.world.Block.Strings) {
			return s.world.Block.Strings[n]
		}
		return "BLOCKED."
	}
	switch res.Event.Kind {
	case game.EventMessage, game.EventRadiation:
		// 字串編號在 Event.Strings（nibble 4／9 是記錄 +0x00，可能是 0 ＝ 不印）。
		// nibble 1 的訊息**可以有很多條**，原版逐條印出來。
		var out []string
		for _, n := range res.Event.Strings {
			if n >= 0 && n < len(s.world.Block.Strings) {
				if line := s.world.Block.Strings[n]; line != "" {
					out = append(out, line)
				}
			}
		}
		return strings.Join(out, "")
	case game.EventTeleport:
		// 記錄 +0x00 的低 6 位就是「進入石英城。」那一句（`sub_16B17`）。
		// 會先問「進新地點？」的傳送格由 walk 印，這裡是不問的那一種。
		if len(res.Event.Strings) > 0 {
			if n := res.Event.Strings[0]; n > 0 && n < len(s.world.Block.Strings) {
				return s.world.Block.Strings[n]
			}
		}
		return ""
	case game.EventChest:
		return "SOMETHING IS HERE."
	case game.EventMenu:
		// 題目由 beginQuestion 印（`docs/re/46` §4）——這裡回空的，
		// 不然那一句會蓋在題目上面。
		return ""
	case game.EventGate:
		// 走得過去的條件格踩上去會印記錄 +0x01 的訊息（docs/re/66）。
		if len(res.Event.Strings) > 0 {
			if n := res.Event.Strings[0]; n > 0 && n < len(s.world.Block.Strings) {
				return s.world.Block.Strings[n]
			}
		}
		return ""
	case game.EventFacility:
		// bit7 設起來的是設施畫面，沒設的是腳本指令（docs/spec/09 §2）。
		if f, ok := game.ParseFacility(res.Event.Data); ok {
			return f.Name
		}
		return "SOMETHING HAPPENS."
	}
	// ⚠ **這裡不報遭遇。** 擲骰說「觸發」不等於打得起來——還要視窗裡有敵人格、
	// 距離過得了記錄的兩道門檻（docs/spec/15）。掃描落空卻印「被攻擊」，
	// 玩家會看到一句什麼都沒發生的警告。訊息由 updateMap 依 StartEncounter
	// 的結果決定。
	return ""
}

// StoreTo 把目前的狀態寫回存檔的明文。
//
// **改寫不是重建**（CLAUDE.md §4）：只蓋已解欄位，未解區域一個 byte 都不動。
// 沒有走過路、沒有改過角色時，寫回去應該與讀進來完全相同。
func (s *Scene) StoreTo(save *assets.Save) error {
	groups := save.SlotGroups()
	g := groups[0]

	// 隊伍座標與所在地圖（隊伍槽表 +0x08/+0x09/+0x0A）。
	slot := save.Plain[g.RawIndex : g.RawIndex+14]
	slot[8] = s.world.Party.X
	slot[9] = s.world.Party.Y
	slot[10] = byte(s.blockID) // 所在地圖（docs/re/60 §3 的槽表 +0x0A）

	// 時鐘與視窗原點（ds:464Eh–465Bh 的副本）。
	gl := save.Globals()
	gl[0] = byte(s.world.ViewX)
	gl[1] = byte(s.world.ViewY)
	gl[2] = byte(s.world.Clock.Total)
	gl[3] = byte(s.world.Clock.Total >> 8)
	gl[4] = byte(s.world.Clock.Total >> 16)
	gl[10] = s.world.Clock.Frac
	gl[11] = s.world.Clock.Minute
	gl[12] = s.world.Clock.Hour

	// 角色記錄：就地蓋回已解欄位。
	i := 0
	for _, id := range g.Members {
		if id == 0 {
			continue
		}
		if i >= len(s.world.Party.Members) {
			break
		}
		raw, err := save.Record(int(id))
		if err != nil {
			return err
		}
		s.world.Party.Members[i].StoreTo(raw)
		i++
	}
	return nil
}

// Save 回傳目前這一份存檔（呼叫者負責寫檔）。
func (s *Scene) Save() *assets.Save { return s.save }

// setStock 改一筆物品的店家庫存，解析後的表與明文一起改。
//
// 兩份都要改：規則層讀的是解析後的表，寫回存檔用的是明文。
// 只改一邊會造成「這一場看得到、存完就沒了」或者反過來。
func (s *Scene) setStock(id, v byte) {
	if !s.items.SetStock(id, v) {
		return
	}
	// 明文的索引比表的索引大一（`sub_17AE0` 的基址是表首 ＋ 8）。
	if at := (int(id) + 1) * 8; at+2 < len(s.itemsRaw) {
		s.itemsRaw[at+2] = v
	}
}

// SetSaveDir 指定指令列的 `Save` 要把存檔寫回哪個目錄。
//
// 空字串 ＝ **只更新記憶體、不寫檔**，無頭工具與測試用；
// 那時 `Save` 會照實說它沒有寫出去，不會報一句「存好了」。
func (s *Scene) SetSaveDir(dir string) { s.saveDir = dir }

// translate 查這一步的訊息有沒有中文。查不到就回 nil，顯示原文
// （docs/spec/11 §7：半成品的中文化要能玩）。
func (s *Scene) translate(res game.StepResult) []byte {
	if s.cat == nil {
		return nil
	}
	// ⚠ **key 的 slot 是字串編號，不是記錄編號。** `describe` 印的是
	// `Event.Strings` 指到的那幾條，翻譯要查同一組編號——
	// 拿 `Event.Record` 去查會**每次都查不到**，而症狀只是「畫面上是英文」，
	// 看起來像還沒翻。
	var out []byte
	for _, n := range s.stringSlots(res) {
		if n <= 0 {
			continue
		}
		// 控制碼要解掉再送去畫（見 cjkLookup）。選擇子用預設：
		// 單數、男性、him——**原版在印之前設的那三個變數還沒 RE**
		// （`ds:4687h`／`ds:470Bh`／`ds:471Ah`），地圖敘述大多不帶變形碼。
		if b := s.cjkLookup(lang.BlockKey(s.blockFile, s.blockID, n),
			textlayout.Options{}); b != nil {
			out = append(out, b...)
		}
	}
	return out
}

// stringSlots 回傳這一步實際印出來的字串編號，與 `describe` 取的是同一組。
func (s *Scene) stringSlots(res game.StepResult) []int {
	if !res.Moved {
		if res.Blocked > 0 {
			return []int{res.Blocked}
		}
		return nil
	}
	switch res.Event.Kind {
	case game.EventMessage, game.EventRadiation, game.EventTeleport:
		return res.Event.Strings
	case game.EventGate:
		if len(res.Event.Strings) > 0 {
			return res.Event.Strings[:1]
		}
	}
	return nil
}

// Frame 合成一幀：地圖視窗 ＋ 時鐘 ＋ 訊息。
func (s *Scene) Frame() *render.Frame {
	if !s.dirty && s.frame != nil {
		return s.frame
	}
	f := render.NewFrame()
	if s.title {
		s.drawTitle(f)
		s.frame = f
		s.dirty = false
		return f
	}
	if s.wipe.active {
		s.drawWipe(f)
	} else if s.ending.active {
		s.drawEnding(f)
	} else {
	switch {
	case s.facility != nil:
		s.drawFacility(f)
	case s.combat != nil:
		s.drawPortrait(f)
		s.drawRoster(f)
	default:
		s.drawMap(f)
	}
	}
	// 時鐘在外框上緣，不屬於地圖視窗——切模式不影響它（docs/re/27 §4）。
	// **結局沒有時鐘也沒有指令列**：那時候已經不在遊戲裡了。
	if !s.ending.active && !s.wipe.active && !s.hiText() {
		_ = f.DrawClock(s.font, int(s.world.Clock.Hour), int(s.world.Clock.Minute))
	}

	// 地圖模式才有指令列（`docs/re/91`）——戰鬥與設施有自己的選單。
	// **有中文字型時這一行改由 HiFrame 畫**（見 drawCommandBarCJK）：
	// 8 × 8 的字模畫不出中文，先畫英文再蓋會留下殘影。
	if s.facility == nil && s.combat == nil && !s.ending.active && !s.wipe.active &&
		s.eten == nil {
		_ = f.DrawLineAt(s.font, commandBar(), 0, render.CmdRow)
	}

	if s.message != "" && !s.hiText() {
		out, err := textlayout.Layout([]byte(s.message), textlayout.Options{Width: render.MsgWidth})
		if err == nil {
			lines := out.Lines
			// ⚠ 有中文正文要顯示時，英文訊息**只留第一行當標題**——
			// 兩者畫在同一個訊息視窗，全部畫出來會疊在一起
			// （中文是 16 × 15、英文是 8 × 8，疊起來兩邊都讀不了）。
			if s.cjkVisible() && len(lines) > 1 {
				lines = lines[:1]
			}
			_ = f.DrawText(s.font, lines)
		}
	}
	s.frame = f
	s.dirty = false
	return s.frame
}

// drawMap 畫地圖那一種（docs/spec/24 §4）。
func (s *Scene) drawMap(f *render.Frame) {
	if err := f.DrawMap(s.world.Block, s.gfx, s.world.ViewX, s.world.ViewY); err != nil {
		s.message = "ERROR: " + err.Error()
	}
	// 地形畫完之後才疊圖：寶箱、輻射區、nibble 4 的格（docs/spec/03 §2.9）。
	for _, ic := range s.world.ViewIcons() {
		if err := f.DrawIcon(s.gfx, ic.Icon, ic.Col, ic.Row); err != nil {
			s.message = "ERROR: " + err.Error()
		}
	}
	// 隊伍圖示疊在地圖上（docs/re/47 §5 對拍抓出來的缺口）。
	if err := f.DrawParty(s.gfx); err != nil {
		s.message = "ERROR: " + err.Error()
	}
}

// drawPortrait 畫敵人肖像（遭遇時，`docs/re/37` §3.2）。
//
// 位置與設施圖相同——兩者走同一支載入器（`docs/re/23` §4）。
// 名單從字元列 14（y ＝ 112）起，圖是 y ＝ 8–92，**不重疊**。
func (s *Scene) drawPortrait(f *render.Frame) {
	if s.portrait < 0 || s.pics == nil || s.portrait >= len(s.pics) {
		return
	}
	f.DrawIndexed(s.pics[s.portrait], render.FacilityPicX, render.FacilityPicY,
		render.ViewClip())
}

// drawRoster 畫戰鬥那一種：地圖視窗換成隊伍名單（docs/re/40 §1）。
func (s *Scene) drawRoster(f *render.Frame) {
	row := render.RosterHeaderRow
	// 有倚天字型與譯文時表頭由 HiFrame 畫中文——先畫英文再蓋會留殘影
	// （與指令列同一條，`docs/spec/10` §5）。
	if s.eten == nil || len(rosterHeaderCJK(s.uiText)) == 0 {
		_ = f.DrawLineAt(s.font, RosterHeader, 0, row)
	}
	if s.hiText() {
		return // 名單那幾行由 HiFrame 用倚天半形字畫
	}
	for i, r := range Roster(s.world.Party) {
		if row+1+i > render.MsgRow-1 {
			break // 名單與訊息視窗在字元列上會撞（docs/spec/03 §3）
		}
		_ = f.DrawLineAt(s.font, r.Text(), 0, row+1+i)
	}
}

// drawFacility 畫設施那一種：地圖視窗換成那張 ALLPICS 圖。
func (s *Scene) drawFacility(f *render.Frame) {
	fs := s.facility
	// 位置是**實機對拍量出來的**（docs/re/54）：圖在視窗原點 (8, 8)、
	// 96 × 84，地圖在右邊照常露出來——**不是整個視窗換成圖**。
	if fs.Picture >= 0 && s.pics != nil && fs.Picture < len(s.pics) {
		f.DrawIndexed(s.pics[fs.Picture], render.FacilityPicX, render.FacilityPicY,
			render.ViewClip())
		f.ApplyAnimMask(s.animMask, render.FacilityPicWidth,
			render.FacilityPicX, render.FacilityPicY)
	}
	for i, l := range fs.Lines {
		// 有中文的那一行交給 HiFrame 畫——8 × 8 的字模畫不出中文，
		// 先畫英文再蓋會留下殘影（與指令列同一條）。
		if s.eten != nil && i < len(fs.CJKLines) && len(fs.CJKLines[i]) > 0 {
			continue
		}
		if s.hiText() {
			continue // 純英文那幾行也交給 HiFrame，字才會與中文同高
		}
		_ = f.DrawLineAt(s.font, l, render.FacilityNameCol, render.FacilityNameRow+i)
	}
}

// doTeleport 執行一次傳送（docs/spec/07 §6.7）。
//
// 回程資訊存在隊伍槽表的 +0x0B–+0x0D。remake 把它放在 Scene 上：
// 存檔寫回時再落到那三個 byte（規格 05）。
// teleportPatchAt 是傳送收尾改寫腳下那一格用的位移（`sub_169B1(4)`）。
const teleportPatchAt = 4

func (s *Scene) doTeleport(rec []byte) error {
	w := s.world
	here := game.Return{X: w.Party.X, Y: w.Party.Y, MapID: uint8(s.MapID())}
	// **先改寫腳下這一格**（`0x16A24` 的 `sub_169B1(4)`，docs/re/73）：
	// 記錄 `+0x04`／`+0x05` 是「進去之後這一格變成什麼」。
	// 22 筆設施就是靠這一步從傳送格變成 nibble 6 ——
	// 商店與醫生的入口全部在這裡。
	w.PatchHere(rec, teleportPatchAt)
	target, back := game.ResolveTeleport(rec, here, s.back)
	s.back = back
	// 編號 bit7 設起來的是**建築內部**，要先查表換成真正的資源編號
	// （docs/re/61）。忘了換會拿 130 這種值去索引 42 個區塊。
	id, err := s.rom.ResolveMapID(target.MapID)
	if err != nil {
		return err
	}
	target.MapID = id
	if int(target.MapID) == s.MapID() {
		// 同一張地圖：只搬座標，不重載。
		w.Teleport(target.X, target.Y)
		s.dirty = true
		return nil
	}
	return s.LoadMap(int(target.MapID), target.X, target.Y)
}

// MapID 是目前這張地圖的資源編號。
func (s *Scene) MapID() int { return s.blockID }

// gateHurtLine 把條件閘的懲罰寫成原版那句話。
//
// 原版逐個受罰的人各印一行（`sub_157D6` 的 `" gets hurt for "`，`docs/re/67`）；
// 訊息視窗只有幾行，這裡把同一批合成一行，**傷害量照實算總和**。
func (s *Scene) gateHurtLine(g game.GateResult) string {
	total := 0
	for _, h := range g.Failed {
		if h.Field == 0x1D && h.Amount < 0 {
			total -= h.Amount
		}
	}
	if total == 0 {
		return ""
	}
	name := "The party"
	if len(g.Failed) == 1 {
		if m := g.Failed[0].Member; m >= 0 && m < len(s.world.Party.Members) {
			name = s.world.Party.Members[m].Name
		}
	}
	return fmt.Sprintf("%s%s%d point of damage.", name, s.exeString(0x63), total)
}

// gateHurtCJK 是同一句的中文：名字 ＋ 字串 0x63（` gets hurt for `）＋ 點數。
// 查不到就回 nil，畫面留英文那一句。
func (s *Scene) gateHurtCJK(g game.GateResult) []byte {
	total := 0
	for _, h := range g.Failed {
		if h.Field == 0x1D && h.Amount < 0 {
			total -= h.Amount
		}
	}
	if total == 0 {
		return nil
	}
	name := s.uiText("gate.party")
	if len(g.Failed) == 1 {
		if m := g.Failed[0].Member; m >= 0 && m < len(s.world.Party.Members) {
			name = []byte(s.world.Party.Members[m].Name)
		}
	}
	return zhJoin(name,
		s.cjkExe(exeTable1, 0x63, textlayout.Options{}),
		[]byte(fmt.Sprintf("%d", total)),
		s.uiText("gate.damage"))
}

// cjkLookup 查一條譯文，並把原版的控制碼**解乾淨**再回傳。查不到回 nil。
//
// ⚠ 不解碼的話症狀很像「這句沒翻」：`\x0A` 就是 `\n`，畫面上會把單複數
// **兩段都印出來**、中間多一個換行；`\x0B`（名字）、`\x0F`（數量）、
// `\x10`（熱鍵標記）則變成怪字元。譯文裡這些碼很常見（`docs/re/28`）。
func (s *Scene) cjkLookup(key string, opt textlayout.Options) []byte {
	if s.cat == nil {
		return nil
	}
	b, ok := s.cat.Lookup(key)
	if !ok {
		return nil
	}
	return textlayout.RenderBytes(b, opt)
}

// cjkExe 查執行檔字串表第 table 張的第 n 條譯文。
func (s *Scene) cjkExe(table, n int, opt textlayout.Options) []byte {
	return s.cjkLookup(lang.ExeKey(table, n), opt)
}

// exeTable1 是無線電、隊伍、戰鬥那張表（`ds:AB3Eh`，`docs/re/17` §3）。
// `exeString` 取的就是它，中文走同一組編號。
const exeTable1 = 1

// say 設定這一步的訊息：英文照舊，中文查同一條原版字串的譯文。
//
// 兩邊都設是刻意的——沒有字型或沒有譯文時畫面還是有話可看
// （`docs/spec/11` §7：半成品的中文化要能玩）。
func (s *Scene) say(n int, opt textlayout.Options) { s.sayT(exeTable1, n, opt) }

// sayT 是 say 的多字串表版本。
func (s *Scene) sayT(table, n int, opt textlayout.Options) {
	s.message = s.exeStringN(table, n)
	s.cjk = s.cjkExe(table, n, opt)
	if s.cjk != nil {
		// 有中文就不要再畫英文——訊息視窗只有六行，兩份疊上去誰都讀不完
		// （英文 8 × 8、中文 16 × 15，`docs/spec/10` §2）。
		s.message = ""
	}
	s.dirty = true
}

// cjkFmt 把一條 `ui:` 格式字串填好參數放進中文那一格；
// 沒有翻譯就不動（畫面留英文那一句）。
//
// ⚠ 參數裡的名字可能本身就是 Big5（中文角色名），所以整串當位元組處理，
// 格式動詞一律是 ASCII 的 `%s`／`%d`。
func (s *Scene) cjkFmt(uiName string, args ...any) {
	f := s.uiText(uiName)
	if len(f) == 0 {
		return
	}
	s.cjk = []byte(fmt.Sprintf(string(f), args...))
	s.message = ""
	s.dirty = true
}

// sayEN 設一句沒有原版字串對應的話（重製版自己的），中文走 `ui:`。
func (s *Scene) sayEN(en string, uiName string) {
	s.message = en
	s.cjk = s.uiText(uiName)
	if s.cjk != nil {
		s.message = ""
	}
	s.dirty = true
}

// uiText 取重製版介面文字的中文（Big5）。沒有翻譯就回 nil。
func (s *Scene) uiText(name string) []byte {
	if s.cat == nil {
		return nil
	}
	if b, ok := s.cat.Lookup(lang.UIKey(name)); ok {
		return b
	}
	return nil
}

// drawASCII 在高解畫面上畫一個 ASCII 字元。
//
// **優先用倚天自己的半形字模**：它與漢字同高、同一套設計，中英混排才會齊。
// 沒有 `ASCFONT.*` 時退回遊戲原版的 8 × 8 字模放大——那個筆劃比中文粗，
// 是後備不是首選。
func (s *Scene) drawASCII(h *render.HiFrame, c byte, col, row int) {
	if h.DrawETenASCII(s.eten, c, col, row, 15) {
		return
	}
	h.DrawASCIIAt(s.font, c, col, row, 15)
}

// drawCJKLine 在高解畫面上畫一行中英混排的字（Big5）。
//
// ⚠ **逐 byte 判型別，不能整串兩兩配對**——這一行一定夾著熱鍵字母
// （「U 使用」），把 `U` 當成 Big5 高位元組會讓整行往後錯開。
func (s *Scene) drawCJKLine(h *render.HiFrame, text []byte, col, row int) {
	for i := 0; i < len(text); {
		c := text[i]
		if c < 0x80 {
			s.drawASCII(h, c, col, row)
			col++
			i++
			continue
		}
		if i+1 >= len(text) {
			break
		}
		h.DrawCJK(s.eten, text[i], text[i+1], col, row, 15)
		// ⚠ **一個中文字佔一格，不是兩格**（`docs/spec/10` §3：倚天 16 × 15
		// 剛好是放大後的一個字元格）。前進兩格會把整行拉開一倍——
		// 指令列與設施清單都會變成「使 用」這種疏排，而且尾巴掉出畫面。
		col++
		i += 2
	}
}
