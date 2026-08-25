package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

func TestFaithfulArtModeSwitchIsAtomicAndFollowsMap(t *testing.T) {
	s := openScene(t)
	if got := s.ArtMode(); got != "original" {
		t.Fatalf("初始美術模式 = %q", got)
	}
	if err := s.SelectArtMode("../../artpacks", "faithful-hd"); err != nil {
		t.Fatalf("切到 faithful-hd：%v", err)
	}
	frame, err := s.ArtFrame()
	if err != nil {
		t.Fatalf("合成 faithful-hd：%v", err)
	}
	if got := frame.Bounds().Size(); got.X != FaithfulCanvasWidth || got.Y != FaithfulCanvasHeight {
		t.Fatalf("faithful-hd 畫布 = %v", got)
	}

	// 失敗切換不得破壞已經啟用的完整模式。
	if err := s.SelectArtMode(t.TempDir(), "faithful-hd"); err == nil {
		t.Fatal("不存在的 bundle 意外載入成功")
	}
	if got := s.ArtMode(); got != "faithful-hd" {
		t.Fatalf("失敗切換後模式 = %q", got)
	}
	if _, err := s.ArtFrame(); err != nil {
		t.Fatalf("失敗切換後舊畫面不可用：%v", err)
	}

	oldTileset := s.World().Block.Tileset
	found := false
	for id := 0; id < 128; id++ {
		if err := s.LoadMap(id, 1, 1); err == nil && s.World().Block.Tileset != oldTileset {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("找不到另一組 tileset 的地圖")
	}
	if _, err := s.ArtFrame(); err != nil {
		t.Fatalf("換 tileset 後合成 faithful-hd：%v", err)
	}

	if err := s.SelectArtMode("", "original"); err != nil {
		t.Fatalf("切回 original：%v", err)
	}
	if got := s.ArtMode(); got != "original" {
		t.Fatalf("切回後模式 = %q", got)
	}
}

func TestReimaginedArtModeIsResponsiveAndComplete(t *testing.T) {
	s := openScene(t)
	if err := s.SelectArtMode("../../artpacks", "reimagined"); err != nil {
		t.Fatalf("切到 reimagined：%v", err)
	}
	if got := len(s.reimagined.Characters); got != 7*8*13 {
		t.Fatalf("角色動作資產 = %d，預期 %d", got, 7*8*13)
	}
	if got := len(s.reimagined.Weapons); got != 4*8 {
		t.Fatalf("武器圖層 = %d，預期 %d", got, 4*8)
	}
	for _, size := range [][2]int{{960, 540}, {1280, 720}, {1600, 900}, {1920, 1080}} {
		s.SetArtViewport(size[0], size[1])
		frame, err := s.ArtFrame()
		if err != nil {
			t.Fatalf("合成 %dx%d：%v", size[0], size[1], err)
		}
		if got := frame.Bounds().Size(); got.X != size[0] || got.Y != size[1] {
			t.Fatalf("要求 %dx%d，實得 %v", size[0], size[1], got)
		}
	}

	s.BeginTitle()
	frame, err := s.ArtFrame()
	if err != nil || frame.Bounds().Dx() != 1920 {
		t.Fatalf("reimagined 標題合成：%v, %v", frame.Bounds(), err)
	}

	s2 := openScene(t)
	if err := s2.SelectArtMode("../../artpacks", "reimagined"); err != nil {
		t.Fatal(err)
	}
	s2.facility = &FacilityScene{Picture: 0}
	before := s2.artAnimFrame
	if !s2.TickAnim() || s2.artAnimFrame == before {
		t.Fatal("reimagined 設施互動動畫沒有隨 TickAnim 推進")
	}
	if _, err := s2.ArtFrame(); err != nil {
		t.Fatalf("reimagined 設施場景：%v", err)
	}
	s2.facility = nil
	if err := s2.LoadMap(4, 18, 2); err != nil {
		t.Fatal(err)
	}
	if combat, err := s2.StartEncounter(); err != nil || combat == nil {
		t.Fatalf("建立 reimagined 戰鬥：combat=%v err=%v", combat, err)
	}
	if len(s2.world.Party.Members) >= 3 {
		s2.combat.Phase.Cmd[0] = game.CmdAttack
		s2.world.Party.Members[1].CON = 1
		s2.world.Party.Members[2].CON = 0
	}
	if _, err := s2.ArtFrame(); err != nil {
		t.Fatalf("reimagined 攻擊／受傷／倒下狀態：%v", err)
	}
	s2.combat = nil
	s2.BeginEnding()
	if _, err := s2.ArtFrame(); err != nil {
		t.Fatalf("reimagined 結局：%v", err)
	}
}

func TestVisualSettingCyclesAllThreeModes(t *testing.T) {
	s := openScene(t)
	s.SetArtRoot("../../artpacks")
	s.openSettings()
	for _, want := range []string{"faithful-hd", "reimagined", "original"} {
		keep, err := s.updateSettings(input.Input{Char: 'V'})
		if err != nil || !keep {
			t.Fatalf("切到 %s：keep=%v err=%v", want, keep, err)
		}
		if got := s.ArtMode(); got != want {
			t.Fatalf("V 循環模式 = %q，預期 %q", got, want)
		}
	}
}

func TestEveryModernSceneRendersThroughPlayScene(t *testing.T) {
	s := openScene(t)
	for _, mode := range []string{"faithful-hd", "reimagined"} {
		if err := s.SelectArtMode("../../artpacks", mode); err != nil {
			t.Fatalf("載入 %s：%v", mode, err)
		}
		if mode == "reimagined" {
			s.SetArtViewport(960, 540)
		}
		for picture := 0; picture < 82; picture++ {
			s.facility = &FacilityScene{Picture: picture}
			if _, err := s.ArtFrame(); err != nil {
				t.Fatalf("%s 場景 %03d：%v", mode, picture, err)
			}
		}
	}
}

func TestEveryTilesetRendersInBothModernModes(t *testing.T) {
	for _, mode := range []string{"faithful-hd", "reimagined"} {
		s := openScene(t)
		if err := s.SelectArtMode("../../artpacks", mode); err != nil {
			t.Fatal(err)
		}
		if mode == "reimagined" {
			s.SetArtViewport(960, 540)
		}
		seen := map[int]bool{s.World().Block.Tileset: true}
		maps := 0
		for id := 0; id < 128; id++ {
			if err := s.LoadMap(id, 1, 1); err != nil {
				continue
			}
			maps++
			seen[s.World().Block.Tileset] = true
			if _, err := s.ArtFrame(); err != nil {
				t.Fatalf("%s 地圖 %d：%v", mode, id, err)
			}
		}
		if maps != 42 || len(seen) != 9 {
			t.Fatalf("%s：maps=%d tilesets=%d", mode, maps, len(seen))
		}
	}
}
