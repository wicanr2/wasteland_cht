// Package viewer 是**資產檢視器**的場景，不是遊戲。
//
// 用途只有一個：把 assets 解出來的東西畫成原版的畫面，好拿去與 DOSBox 跑的
// 原版截圖對拍（docs/spec/01 §4、docs/spec/03 §4）。
//
// ⚠ **這裡沒有、也不得有任何遊戲規則**：移動只搬視窗原點，不判定能不能走、
// 不推進時鐘、不擲遭遇。那些要等 docs/spec/04 標 READY（CLAUDE.md §0）。
//
// 這個套件不依賴 Ebiten，所以 PNG 輸出可以在無頭環境跑。
package viewer

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// Mode 是要看什麼。
type Mode string

const (
	ModeMap     Mode = "map"
	ModeTitle   Mode = "title"
	ModePicture Mode = "pic"
	// ModeEnd 是結局畫面（`END.CPA`，docs/re/23 §6）。
	ModeEnd     Mode = "end"
)

// Viewer 實作 ui.Scene。
type Viewer struct {
	font   *assets.Font
	frame  *render.Frame
	dirty  bool
	mode   Mode
	block  *assets.Block
	gfx    *render.Graphics
	pic    *assets.Indexed
	status string
	ox, oy int
}

// New 建立檢視器。blockID 只在 ModeMap 用，picID 只在 ModePicture 用。
func New(rom *assets.Rom, mode Mode, blockID, picID int) (*Viewer, error) {
	font, err := rom.FontMain()
	if err != nil {
		return nil, err
	}
	v := &Viewer{font: font, mode: mode, frame: render.NewFrame(), dirty: true}

	switch mode {
	case ModeMap:
		b, err := rom.Block(blockID)
		if err != nil {
			return nil, err
		}
		tiles, err := rom.Tileset(b.Tileset)
		if err != nil {
			return nil, err
		}
		icons, err := rom.Icons()
		if err != nil {
			return nil, err
		}
		v.block = b
		v.gfx = &render.Graphics{Icons: icons, Tiles: tiles}
		// 視窗中心對準地圖中央（隊伍在第 (9,4) 格）。
		v.ox = b.Dim/2 - render.PartyCol
		v.oy = b.Dim/2 - render.PartyRow
		v.status = fmt.Sprintf("BLOCK %d  %dX%d  TILESET %d  STEP %.2f MIN",
			blockID, b.Dim, b.Dim, b.Tileset, b.StepMinutes())
	case ModeEnd:
		im, err := rom.End()
		if err != nil {
			return nil, err
		}
		v.pic = im
		v.status = "END.CPA  288X128"
	case ModeTitle:
		im, err := rom.Title()
		if err != nil {
			return nil, err
		}
		v.pic = im
		v.status = "TITLE.PIC  288X128"
	case ModePicture:
		list, err := rom.Pictures("allpics1")
		if err != nil {
			return nil, err
		}
		more, err := rom.Pictures("allpics2")
		if err != nil {
			return nil, err
		}
		list = append(list, more...)
		if picID < 0 || picID >= len(list) {
			return nil, fmt.Errorf("圖片編號 %d 超出範圍（共 %d 張）", picID, len(list))
		}
		v.pic = list[picID]
		v.status = fmt.Sprintf("ALLPICS %d/%d  96X84", picID, len(list))
	default:
		return nil, fmt.Errorf("不認識的模式：%q（要 map｜title｜pic）", mode)
	}
	return v, nil
}

// Update 只搬視窗原點——沒有任何判定，這是檢視器不是遊戲。
func (v *Viewer) Update(in input.Input) (bool, error) {
	if in.Action == input.ActionQuit {
		return false, nil
	}
	switch in.Dir {
	case input.DirUp:
		v.oy--
	case input.DirDown:
		v.oy++
	case input.DirLeft:
		v.ox--
	case input.DirRight:
		v.ox++
	default:
		return true, nil
	}
	v.dirty = true
	return true, nil
}

// Frame 合成一幀。
func (v *Viewer) Frame() *render.Frame {
	if !v.dirty {
		return v.frame
	}
	f := render.NewFrame()
	switch v.mode {
	case ModeMap:
		if err := f.DrawMap(v.block, v.gfx, v.ox, v.oy); err != nil {
			v.status = "ERROR: " + err.Error()
		}
	default:
		if v.pic.Width == render.ViewWidth && v.pic.Height == render.ViewHeight {
			_ = f.DrawPicture(v.pic)
		} else {
			// 96 × 84 的圖置中放進視窗
			x := render.ViewX + (render.ViewWidth-v.pic.Width)/2
			y := render.ViewY + (render.ViewHeight-v.pic.Height)/2
			f.DrawIndexed(v.pic, x, y, render.ViewClip())
		}
	}
	if out, err := textlayout.Layout([]byte(v.status), textlayout.Options{Width: render.MsgWidth}); err == nil {
		_ = f.DrawText(v.font, out.Lines)
	}
	v.frame = f
	v.dirty = false
	return v.frame
}
