// Package play 是**可以走的遊戲場景**：把 internal/game 的規則接到畫面上。
//
// 與 internal/viewer 的分工：viewer 只負責把資產畫出來對拍，一條規則都沒有；
// play 走的是規則層——能不能進、時鐘、遭遇、事件全部照 docs/spec/04 與 07。
//
// 這個套件不依賴 Ebiten，所以無頭也跑得動（PNG 對拍用同一份場景）。
package play

import (
	"fmt"

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

	block, err := rom.Block(mapID)
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
	s.message = save.Place()
	if raw, err := rom.LoadItemTable(save.File, 0); err == nil {
		s.items = game.ParseItemTable(raw)
	} else {
		// 載不到就維持空表：傷害會是 0，但遊戲跑得動。
		// **錯誤要留在畫面上**，不要靜靜吞掉——那會變成「戰鬥打不完」的怪症狀。
		s.message = "ITEM TABLE: " + err.Error()
	}
	s.blockFile, s.blockID = block.Resource.File, block.Resource.ID
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

// World 讓測試與 cmd/wl-shot 拿得到規則層的狀態。
func (s *Scene) World() *game.World { return s.world }

// Invalidate 讓下一次 Frame 重畫。外部直接改了世界狀態時要叫它。
func (s *Scene) Invalidate() { s.dirty = true }

// Update 走一步。ESC 取消、F10 離開（docs/spec/03 的按鍵模型）。
func (s *Scene) Update(in input.Input) (bool, error) {
	if in.Action == input.ActionQuit {
		return false, nil
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

	res, err := s.world.Step(dir)
	if err != nil {
		return true, err
	}
	s.dirty = true
	s.message = s.describe(res)
	s.cjk = s.translate(res)

	// 踩到設施格就進設施畫面（docs/spec/23）。
	// bit7 沒設的是腳本指令，`EnterFacility` 會回 nil——那條路不歸這裡管。
	if res.Moved && res.Event.Kind == game.EventFacility {
		s.EnterFacility(res.Event.Data)
	}

	// 走一步之後掃遭遇（docs/re/51 §2）。掃描說沒有可打的就什麼都不做——
	// **擲骰說「觸發」不等於真的打得起來**，還要視窗裡有敵人格、
	// 距離過得了記錄的兩道門檻（docs/spec/15）。
	if res.Moved && res.Encounter && !s.InFacility() {
		if c, err := s.StartEncounter(); err != nil {
			s.message = "ERROR: " + err.Error()
		} else if c != nil {
			s.message = c.Log[len(c.Log)-1]
		}
	}
	return true, nil
}

// describe 把一步的結果變成訊息視窗要顯示的字。
// 規則層只給編號，文字在這裡才查出來——中文化改的是這一層與翻譯目錄。
func (s *Scene) describe(res game.StepResult) string {
	if !res.Moved {
		return "BLOCKED."
	}
	switch res.Event.Kind {
	case game.EventMessage:
		if n := res.Event.Record; n >= 0 && n < len(s.world.Block.Strings) {
			return s.world.Block.Strings[n]
		}
	case game.EventTeleport:
		return "TELEPORT."
	case game.EventChest:
		return "SOMETHING IS HERE."
	case game.EventMenu:
		return "CHOOSE."
	case game.EventGate:
		return "BLOCKED BY SOMETHING."
	case game.EventFacility:
		// bit7 設起來的是設施畫面，沒設的是腳本指令（docs/spec/09 §2）。
		if f, ok := game.ParseFacility(res.Event.Data); ok {
			return f.Name
		}
		return "SOMETHING HAPPENS."
	}
	if res.Encounter {
		return "YOU ARE BEING ATTACKED!"
	}
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
