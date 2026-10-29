package play

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/render"
)

// 這是 preview 的狀態與文字保護，不代替新繪圖塊的逐格／實際畫面審查。
func TestHD2TilePreviewKeepsWorldAndHUD(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(1, 16, 16); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	dir := filepath.Join(root, "tileset-3")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	im := image.NewRGBA(image.Rect(0, 0, 48, 48))
	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			im.SetRGBA(x, y, color.RGBA{R: 210, G: 32, B: 190, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, im); err != nil {
		t.Fatal(err)
	}
	for i := range s.gfx.Tiles {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%03d.png", i)), buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	before := s.HiFrame().ToImage()
	clock, partyX, partyY := s.world.Clock, s.world.Party.X, s.world.Party.Y
	raw := append([]byte(nil), s.world.Block.Raw...)
	terrain := append([]byte(nil), s.world.Block.Terrain...)
	record := append([]byte(nil), s.world.Block.Record...)
	graphic := append([]byte(nil), s.world.Block.Graphic...)
	if err := s.LoadHD2TilePreview(root); err != nil {
		t.Fatal(err)
	}
	after, err := s.FaithfulPreviewFrame()
	if err != nil {
		t.Fatal(err)
	}
	if s.ArtMode() != "original" || s.world.Clock != clock ||
		s.world.Party.X != partyX || s.world.Party.Y != partyY ||
		!bytes.Equal(raw, s.world.Block.Raw) || !bytes.Equal(terrain, s.world.Block.Terrain) ||
		!bytes.Equal(record, s.world.Block.Record) || !bytes.Equal(graphic, s.world.Block.Graphic) {
		t.Fatal("preview 改變玩家模式或規則狀態")
	}
	clip := image.Rect(render.ViewX*3, (render.ViewY+1)*3,
		(render.ViewX+render.ViewWidth)*3, (render.ViewY+render.ViewHeight)*3)
	changed := 0
	for y := 0; y < after.Bounds().Dy(); y++ {
		for x := 0; x < after.Bounds().Dx(); x++ {
			if before.At(x, y) != after.At(x, y) {
				if !image.Pt(x, y).In(clip) {
					t.Fatalf("preview 覆蓋地圖外文字／外框：%d,%d", x, y)
				}
				changed++
			}
		}
	}
	if changed == 0 {
		t.Fatal("preview 沒有使用候選圖塊")
	}
}

func TestPartyArtStepSequence(t *testing.T) {
	s := &Scene{artPartyFrame: 1}
	s.beginPartyArtStep(3)
	if s.artPartyDir != 3 || s.artPartyFrame != 0 || s.artPartyTicks != 3 {
		t.Fatalf("開始左移：dir=%d frame=%d ticks=%d", s.artPartyDir, s.artPartyFrame, s.artPartyTicks)
	}
	want := []int{1, 2, 1}
	for i, frame := range want {
		if !s.tickPartyArt() {
			t.Fatalf("tick %d 沒回報改變", i)
		}
		if s.artPartyFrame != frame {
			t.Fatalf("tick %d frame=%d，預期 %d", i, s.artPartyFrame, frame)
		}
	}
	if s.tickPartyArt() {
		t.Fatal("動畫停下後仍回報改變")
	}
}

func TestReimaginedTilesOverlapInsteadOfExposingGridSeams(t *testing.T) {
	if reimaginedTileStepX >= reimaginedTileWidth {
		t.Fatalf("橫向步距 %d 沒有小於圖寬 %d", reimaginedTileStepX, reimaginedTileWidth)
	}
	if reimaginedTileStepY >= reimaginedTileHeight {
		t.Fatalf("縱向步距 %d 沒有小於圖高 %d", reimaginedTileStepY, reimaginedTileHeight)
	}
}

func TestReimaginedHUDKeepsTextInsideItsOwnBorder(t *testing.T) {
	frame, content, source := reimaginedHUDRects(image.Rect(20, 540, 1260, 708))
	if !content.In(frame) || content.Min.X-frame.Min.X < 12 || content.Min.Y-frame.Min.Y < 12 ||
		frame.Max.X-content.Max.X < 12 || frame.Max.Y-content.Max.Y < 12 {
		t.Fatalf("HUD 文字安全區沒有完整內縮：frame=%v content=%v", frame, content)
	}
	if source.Min.X <= 0 || source.Max.X >= 960 {
		t.Fatalf("來源仍包含原版左右邊框：%v", source)
	}
}
