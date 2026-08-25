package play

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"path/filepath"

	"github.com/wicanr2/wasteland_cht/internal/artpack"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

const (
	FaithfulCanvasWidth  = 960
	FaithfulCanvasHeight = 720
	faithfulContentY     = (FaithfulCanvasHeight - render.HiScreenHeight) / 2
)

// SelectArtMode prepares and atomically commits a complete presentation.
// A failed load leaves both the active mode and decoded images untouched.
func (s *Scene) SelectArtMode(root, mode string) error {
	if mode == "original" {
		s.artMode, s.faithful, s.reimagined, s.artBundle = "original", nil, nil, nil
		s.dirty = true
		return nil
	}
	if mode != string(artpack.ModeFaithfulHD) && mode != string(artpack.ModeReimagined) {
		return fmt.Errorf("尚未提供美術模式 %q 的 renderer", mode)
	}
	bundle, err := artpack.Load(filepath.Join(root, mode))
	if err != nil {
		return err
	}
	if mode == string(artpack.ModeReimagined) {
		prepared, err := artpack.LoadReimaginedComplete(bundle, s.world.Block.Tileset, len(s.gfx.Tiles))
		if err != nil {
			return err
		}
		s.reimagined, s.faithful, s.artBundle, s.artMode = prepared, nil, bundle, mode
	} else {
		prepared, err := artpack.LoadFaithfulComplete(bundle, s.world.Block.Tileset, len(s.gfx.Tiles))
		if err != nil {
			return err
		}
		s.faithful, s.reimagined, s.artBundle, s.artMode = prepared, nil, bundle, mode
	}
	s.dirty = true
	return nil
}

func (s *Scene) ArtMode() string {
	if s.artMode == "" {
		return "original"
	}
	return s.artMode
}

// SetArtRoot 設定三種美術包的共同根目錄；空字串恢復專案預設值。
func (s *Scene) SetArtRoot(root string) {
	if root == "" {
		root = "artpacks"
	}
	s.artRoot = root
}

// ArtFrame returns the faithful-hd 4:3 frame. The existing 960x600 game
// composition is centered without cropping; modern RGB art replaces only the
// corresponding original-art rectangle.
func (s *Scene) ArtFrame() (*image.RGBA, error) {
	if s.ArtMode() == string(artpack.ModeReimagined) {
		return s.reimaginedFrame()
	}
	if s.faithful == nil || s.ArtMode() != string(artpack.ModeFaithfulHD) {
		return nil, fmt.Errorf("目前不是 faithful-hd 模式")
	}
	content := s.HiFrame().ToImage()
	set := s.faithful.Map
	switch {
	case s.title:
		drawFaithfulPicture(content, s.faithful.Title, render.ViewX*render.HiScale, render.ViewY*render.HiScale)
	case s.ending.active:
		drawFaithfulPicture(content, s.faithful.Ending, render.ViewX*render.HiScale, render.ViewY*render.HiScale)
	case s.wipe.active:
		s.drawFaithfulScene(content, wipePicture)
	case s.facility != nil:
		s.drawFaithfulScene(content, s.facility.Picture)
	case s.combat != nil:
		s.drawFaithfulScene(content, s.portraitPicture())
	default:
		if err := render.DrawFaithfulMap(content, s.world.Block, set.Tiles, set.Icons, s.world.ViewX, s.world.ViewY); err != nil {
			return nil, err
		}
		for _, ic := range s.world.ViewIcons() {
			if err := render.DrawFaithfulIcon(content, set.Icons, ic.Icon, ic.Col, ic.Row); err != nil {
				return nil, err
			}
		}
		party := s.artPartyDir*3 + s.artPartyFrame
		if party < 0 || party >= len(set.PartyWalk) {
			party = 1
		}
		if err := render.DrawFaithfulIcon(content, set.PartyWalk, byte(party), render.PartyCol, render.PartyRow); err != nil {
			return nil, err
		}
	}
	canvas := image.NewRGBA(image.Rect(0, 0, FaithfulCanvasWidth, FaithfulCanvasHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(0, faithfulContentY, render.HiScreenWidth, faithfulContentY+render.HiScreenHeight), content, image.Point{}, draw.Src)
	return canvas, nil
}

func (s *Scene) drawFaithfulScene(dst *image.RGBA, index int) {
	if index < 0 || index >= len(s.faithful.Scenes) {
		return
	}
	drawFaithfulPicture(dst, s.faithful.Scenes[index], render.FacilityPicX*render.HiScale, render.FacilityPicY*render.HiScale)
}

func drawFaithfulPicture(dst *image.RGBA, src image.Image, x, y int) {
	if src == nil {
		return
	}
	r := image.Rect(x, y, x+src.Bounds().Dx(), y+src.Bounds().Dy())
	draw.Draw(dst, r, src, src.Bounds().Min, draw.Src)
}
