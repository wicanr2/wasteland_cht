package artpack

import "testing"

func TestCheckedInFaithfulBundle(t *testing.T) {
	bundle, err := Load("../../artpacks/faithful-hd")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Manifest.ID != "faithful-hd-v1" || len(bundle.Manifest.Assets) != 1157 {
		t.Fatalf("id=%q assets=%d", bundle.Manifest.ID, len(bundle.Manifest.Assets))
	}
	if bundle.Manifest.Canvas.Width != 960 || bundle.Manifest.Canvas.Height != 720 || bundle.Manifest.Canvas.Responsive {
		t.Fatalf("canvas=%+v", bundle.Manifest.Canvas)
	}
	complete, err := LoadFaithfulComplete(bundle, 0, 66)
	if err != nil {
		t.Fatal(err)
	}
	if len(complete.Map.Tiles) != 66 || len(complete.Map.Icons) != 10 || len(complete.Map.PartyWalk) != 12 || len(complete.Scenes) != 82 {
		t.Fatalf("tiles=%d icons=%d party=%d scenes=%d", len(complete.Map.Tiles), len(complete.Map.Icons), len(complete.Map.PartyWalk), len(complete.Scenes))
	}
	if complete.Title.Bounds().Dx() != 864 || complete.Title.Bounds().Dy() != 384 || complete.Ending.Bounds().Dx() != 864 || complete.Ending.Bounds().Dy() != 384 {
		t.Fatalf("title=%v ending=%v", complete.Title.Bounds(), complete.Ending.Bounds())
	}
}

func TestCheckedInFaithfulMapZero(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 0, 66)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 66 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}

func TestCheckedInFaithfulMapTilesetOne(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 1, 141)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 141 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}

func TestCheckedInFaithfulMapTilesetTwo(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 2, 163)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 163 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}

func TestCheckedInFaithfulMapTilesetThree(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 3, 107)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 107 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}

func TestCheckedInFaithfulMapTilesetFour(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 4, 127)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 127 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}

func TestCheckedInFaithfulMapTilesetFive(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 5, 118)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 118 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}

func TestCheckedInFaithfulMapTilesetSix(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 6, 90)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 90 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}

func TestCheckedInFaithfulMapTilesetSeven(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 7, 104)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 104 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}

func TestCheckedInFaithfulMapTilesetEight(t *testing.T) {
	set, err := LoadFaithfulMap("../../artpacks/faithful-hd/assets", 8, 135)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Tiles) != 135 || len(set.Icons) != 10 || len(set.PartyWalk) != 12 {
		t.Fatalf("tiles=%d icons=%d party=%d", len(set.Tiles), len(set.Icons), len(set.PartyWalk))
	}
}
