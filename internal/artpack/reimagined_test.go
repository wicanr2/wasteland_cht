package artpack

import "testing"

func TestCheckedInReimaginedBundle(t *testing.T) {
	bundle, err := Load("../../artpacks/reimagined")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Manifest.ID != "wasteland-reimagined-v1" || len(bundle.Manifest.Assets) != 1905 {
		t.Fatalf("id=%q assets=%d", bundle.Manifest.ID, len(bundle.Manifest.Assets))
	}
	c := bundle.Manifest.Canvas
	if c.Width != 1280 || c.Height != 720 || !c.Responsive || c.MaxViewCols != 25 || c.MaxViewRows != 15 {
		t.Fatalf("canvas=%+v", c)
	}
	complete, err := LoadReimaginedComplete(bundle, 0, 66)
	if err != nil {
		t.Fatal(err)
	}
	if len(complete.Map.Tiles) != 66 || len(complete.Map.Icons) != 10 || len(complete.Scenes) != 82 {
		t.Fatalf("tiles=%d icons=%d scenes=%d", len(complete.Map.Tiles), len(complete.Map.Icons), len(complete.Scenes))
	}
	if len(complete.Characters) != 728 || len(complete.Weapons) != 32 {
		t.Fatalf("characters=%d weapons=%d", len(complete.Characters), len(complete.Weapons))
	}
	if got := complete.Title.Bounds().Size(); got.X != 1280 || got.Y != 720 {
		t.Fatalf("title=%v", complete.Title.Bounds())
	}
}
