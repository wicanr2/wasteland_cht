package play

import (
	"fmt"
	"image"

	"github.com/wicanr2/wasteland_cht/internal/artpack"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// LoadFaithfulPreview enables an explicitly incomplete vertical-slice renderer.
// It is intentionally not wired to player settings until the full bundle gate passes.
func (s *Scene) LoadFaithfulPreview(root string) error {
	set, err := artpack.LoadFaithfulMap(root, s.world.Block.Tileset, len(s.gfx.Tiles))
	if err != nil {
		return err
	}
	s.faithfulPreview = set
	return nil
}

func (s *Scene) FaithfulPreviewFrame() (*image.RGBA, error) {
	if s.faithfulPreview == nil {
		return nil, fmt.Errorf("尚未載入 faithful-hd preview")
	}
	base := s.HiFrame().ToImage()
	if s.title || s.ending.active || s.wipe.active || s.facility != nil || s.combat != nil {
		return base, nil // scene replacements are a separate incomplete vertical chain
	}
	if err := render.DrawFaithfulMap(base, s.world.Block, s.faithfulPreview.Tiles,
		s.faithfulPreview.Icons, s.world.ViewX, s.world.ViewY); err != nil {
		return nil, err
	}
	for _, ic := range s.world.ViewIcons() {
		if err := render.DrawFaithfulIcon(base, s.faithfulPreview.Icons, ic.Icon, ic.Col, ic.Row); err != nil {
			return nil, err
		}
	}
	partyIndex := s.artPartyDir*3 + s.artPartyFrame
	if partyIndex < 0 || partyIndex >= len(s.faithfulPreview.PartyWalk) {
		partyIndex = 1
	}
	if err := render.DrawFaithfulIcon(base, s.faithfulPreview.PartyWalk, byte(partyIndex),
		render.PartyCol, render.PartyRow); err != nil {
		return nil, err
	}
	return base, nil
}

func (s *Scene) beginPartyArtStep(dir int) {
	s.artPartyDir, s.artPartyFrame, s.artPartyTicks = dir, 0, 3
	s.dirty = true
}

func (s *Scene) tickPartyArt() bool {
	if s.artPartyTicks <= 0 {
		return false
	}
	s.artPartyTicks--
	s.artPartyFrame++
	if s.artPartyFrame > 2 {
		s.artPartyFrame = 1
	}
	if s.artPartyTicks == 0 {
		s.artPartyFrame = 1
	}
	s.dirty = true
	return true
}
