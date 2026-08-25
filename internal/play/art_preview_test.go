package play

import (
	"image"
	"testing"
)

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
