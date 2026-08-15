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
	// items 是物品資料表（存檔區那一份，docs/re/45 §2）。
	// **武器傷害要靠它**——沒有它每個人的傷害都是 0，戰鬥永遠打不完。
	items game.ItemTable
	// combat 非 nil 時畫面在戰鬥（docs/spec/21）；
	// facility 非 nil 時畫面在設施（docs/spec/23）。兩者不會同時成立。
	combat   *CombatScene
	facility *FacilityScene
	// snapshot 是打之前每個角色的經驗值，收尾時相減用。
	snapshot xpSnapshot
}

// LoadCatalogue 載入翻譯目錄；載不到就維持英文，不當成錯誤。
func (s *Scene) LoadCatalogue(path string) error {
	c, err := lang.Load(path)
	if err != nil {
		return err
	}
	s.cat = c
	s.dirty = true
	return nil
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
func (s *Scene) HiFrame() *render.HiFrame {
	h := render.NewHiFrame()
	h.Upscale(s.Frame())
	if s.eten == nil || len(s.cjk) == 0 {
		return h
	}
	col, row := render.MsgCol, render.MsgRow
	for i := 0; i+1 < len(s.cjk); i += 2 {
		if col >= render.MsgCol+render.MsgWidth {
			col = render.MsgCol
			row++
		}
		if row > render.MsgRowEnd {
			break // 訊息視窗滿了；分頁是控制碼的事（docs/re/14 §4）
		}
		h.DrawCJK(s.eten, s.cjk[i], s.cjk[i+1], col, row, 15)
		col++
	}
	return h
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
		dirty: true,
	}
	s.world.Clock = clock
	// 物品表跟著存檔走（每個存檔槽一份）。載不到就維持空表——
	// 傷害會是 0，但遊戲跑得動，而且下面這行的錯誤會留在訊息裡。
	// 設施畫面的圖。載不到就不畫圖，其餘照跑。
	if pics, err := rom.Pictures("allpics1"); err == nil {
		s.pics = pics
	}
	if anims, err := rom.PictureAnims("allpics1"); err == nil {
		s.anims = anims
	}
	// 技能資料表：條件閘的技能型別要用（docs/re/32 §2）。
	// 載不到就讓技能型別的條件一律失敗，其餘照跑。
	if raw, err := rom.SkillTableRaw(); err == nil {
		s.world.Skills = game.SkillBytes(raw)
	}
	s.message = save.Place()
	if raw, err := rom.LoadItemTable(save.File, 0); err == nil {
		s.items = game.ParseItemTable(raw)
	} else {
		// 載不到就維持空表：傷害會是 0，但遊戲跑得動。
		// **錯誤要留在畫面上**，不要靜靜吞掉——那會變成「戰鬥打不完」的怪症狀。
		s.message = "ITEM TABLE: " + err.Error()
	}
	s.asking = input.DirNone // 零值是 DirUp，會被誤判成「正在問」
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
	groups := save.SlotGroups()
	g := groups[0] // 出廠只有第 0 組有人（docs/spec/05 §3.1）

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
		return nil, 0, fmt.Errorf("存檔裡的第 0 組隊伍是空的")
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
	// F10 任何模式都能離開。
	if in.Action == input.ActionQuit {
		return false, nil
	}
	if s.facility != nil {
		return s.updateFacility(in)
	}
	if s.combat != nil {
		return s.updateCombat(in)
	}
	return s.updateMap(in)
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
		if len(res.Lines) > 0 {
			s.message = res.Lines[len(res.Lines)-1]
		}
		if res.Over {
			out := s.FinishEncounter()
			s.message = combatOverMessage(res.Won, out)
			return true, nil
		}
		c.BeginCommands()
	}
	return true, nil
}

// combatOverMessage 是戰鬥結束那一行。
func combatOverMessage(won bool, out EncounterResult) string {
	if !won {
		return "The party has fallen."
	}
	total := uint32(0)
	for _, xp := range out.XPGained {
		total += xp
	}
	if total == 0 {
		return "The battle is over."
	}
	return fmt.Sprintf("The party gains %d experience.", total)
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
func (s *Scene) exeString(n int) string {
	tables, err := s.rom.ExeStrings()
	if err != nil || len(tables) < 2 || n < 0 || n >= len(tables[1]) {
		return ""
	}
	return tables[1][n]
}

// walk 真正走一步並處理結果。
func (s *Scene) walk(dir game.Direction) (bool, error) {
	res, err := s.world.Step(dir)
	if err != nil {
		return true, err
	}
	// 停下來問「Enter new location?」：不移動、不推進時間。
	if res.Ask != 0 {
		s.asking = dirOf(dir)
		s.message = s.exeString(res.Ask)
		s.dirty = true
		return true, nil
	}
	s.dirty = true
	s.message = s.describe(res)
	s.cjk = s.translate(res)

	// 條件閘的收尾訊息：通過印記錄 +0x02、沒過且沒人受罰印 +0x03（docs/re/69）。
	if res.Gate.Message > 0 && res.Gate.Message < len(s.world.Block.Strings) {
		s.message = s.world.Block.Strings[res.Gate.Message]
	}
	// 罰到人就照原版報一句（執行檔字串表 1 第 99 條 " gets hurt for "）。
	if len(res.Gate.Failed) > 0 {
		if line := s.gateHurtLine(res.Gate); line != "" {
			s.message = line
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
			s.message = "YOU ARE BEING ATTACKED!"
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
		return "TELEPORT."
	case game.EventChest:
		return "SOMETHING IS HERE."
	case game.EventMenu:
		return "CHOOSE."
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

// translate 查這一步的訊息有沒有中文。查不到就回 nil，顯示原文
// （docs/spec/11 §7：半成品的中文化要能玩）。
func (s *Scene) translate(res game.StepResult) []byte {
	if s.cat == nil || res.Event.Kind != game.EventMessage {
		return nil
	}
	key := lang.BlockKey(s.blockFile, s.blockID, res.Event.Record)
	if b, ok := s.cat.Lookup(key); ok {
		return b
	}
	return nil
}

// Frame 合成一幀：地圖視窗 ＋ 時鐘 ＋ 訊息。
func (s *Scene) Frame() *render.Frame {
	if !s.dirty && s.frame != nil {
		return s.frame
	}
	f := render.NewFrame()
	switch {
	case s.facility != nil:
		s.drawFacility(f)
	case s.combat != nil:
		s.drawRoster(f)
	default:
		s.drawMap(f)
	}
	// 時鐘在外框上緣，不屬於地圖視窗——切模式不影響它（docs/re/27 §4）。
	_ = f.DrawClock(s.font, int(s.world.Clock.Hour), int(s.world.Clock.Minute))

	if s.message != "" {
		out, err := textlayout.Layout([]byte(s.message), textlayout.Options{Width: render.MsgWidth})
		if err == nil {
			_ = f.DrawText(s.font, out.Lines)
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

// drawRoster 畫戰鬥那一種：地圖視窗換成隊伍名單（docs/re/40 §1）。
func (s *Scene) drawRoster(f *render.Frame) {
	row := render.RosterHeaderRow
	_ = f.DrawLineAt(s.font, RosterHeader, 0, row)
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
