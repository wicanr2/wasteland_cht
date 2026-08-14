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

	s := &Scene{
		rom:   rom,
		font:  font,
		gfx:   &render.Graphics{Icons: icons, Tiles: tiles},
		world: game.NewWorld(block, party, rng.New()),
		save:  save,
		dirty: true,
	}
	s.world.Clock = clock
	s.message = save.Place()
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
		return "A PLACE."
	}
	if res.Encounter {
		return "YOU ARE BEING ATTACKED!"
	}
	return ""
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
