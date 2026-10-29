package play

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

func modernScene(t *testing.T) *Scene {
	t.Helper()
	s := newScene(t)
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Fatal(err)
	}
	if err := s.EnableModernUI(); err != nil {
		t.Fatal(err)
	}
	if s.eten != nil {
		t.Fatal("modern UI test must not depend on ETen")
	}
	return s
}

func modernShot(t *testing.T, s *Scene, name string) {
	t.Helper()
	im, err := s.ModernFrame()
	if err != nil {
		t.Fatal(err)
	}
	if im.Bounds() != image.Rect(0, 0, 960, 720) {
		t.Fatal(im.Bounds())
	}
	for _, f := range s.modern.fields {
		if !f.Ink.In(f.Safe) {
			t.Fatalf("unsafe text: %+v", f)
		}
	}
	if dir := os.Getenv("WL_MODERN_SHOTS"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, im)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func modernClick(t *testing.T, s *Scene, x, y int) {
	t.Helper()
	step(t, s, input.Input{Dir: input.DirNone, Mouse: input.Mouse{X: x, Y: y, Left: true}})
}

func TestModernMapPreservesViewportWithoutETen(t *testing.T) {
	s := newScene(t)
	before := s.HiFrame().ToImage()
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Fatal(err)
	}
	if err := s.EnableModernUI(); err != nil {
		t.Fatal(err)
	}
	modernShot(t, s, "map-original")
	after := s.modern.frame
	for y := 0; y < 384; y++ {
		for x := 0; x < 864; x++ {
			if before.RGBAAt(24+x, 24+y) != after.RGBAAt(24+x, 84+y) {
				t.Fatalf("viewport changed at %d,%d", x, y)
			}
		}
	}
	if s.modernText("rounded", "") != "圓體" {
		t.Fatal("built-in Chinese UI unavailable")
	}
}

func TestModernToolbarOwnsClicksAndQuickSave(t *testing.T) {
	s := modernScene(t)
	s.SetQuickSavePath(filepath.Join(t.TempDir(), "quick.wlqs"))
	x, y, clock := s.world.Party.X, s.world.Party.Y, s.world.Clock
	for i := 0; i < 3; i++ {
		modernClick(t, s, 420, 22)
	}
	if s.modern.pixels != 20 {
		t.Fatal(s.modern.pixels)
	}
	modernClick(t, s, 250, 22)
	if s.modern.family != 1 {
		t.Fatal("font toggle failed")
	}
	modernClick(t, s, 340, 22)
	if s.modern.pixels != 18 {
		t.Fatal(s.modern.pixels)
	}
	modernClick(t, s, 500, 22)
	if s.settings.MusicVariant != "modern" {
		t.Fatal("score control failed")
	}
	if s.world.Party.X != x || s.world.Party.Y != y || s.world.Clock != clock {
		t.Fatal("toolbar click reached world")
	}
	modernClick(t, s, 730, 22)
	b, err := os.ReadFile(s.QuickSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b, []byte("WLQS")) {
		t.Fatal("wrong quicksave")
	}
	step(t, s, input.Input{Dir: input.DirUp})
	modernClick(t, s, 820, 22)
	if s.world.Party.X != x || s.world.Party.Y != y {
		t.Fatal("toolbar quickload did not restore position")
	}
	modernShot(t, s, "loaded")
}

func TestModernMenusRemainClickableAndQuitOwnsToolbar(t *testing.T) {
	s := modernScene(t)
	step(t, s, key('1'))
	modernShot(t, s, "character-info")
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionConfirm})
	modernShot(t, s, "character-items")
	if s.charView.stage != cvItems {
		t.Fatal(s.charView.stage)
	}
	if s.cjk == "" {
		t.Fatal("item translation still depends on ETen")
	}
	clicked := false
	for _, h := range s.modern.hits {
		cell, ok := s.modernMessageCells()[[2]int{h.col, h.row}]
		if ok && cell.Rune == '1' {
			modernClick(t, s, (h.rect.Min.X+h.rect.Max.X)/2, (h.rect.Min.Y+h.rect.Max.Y)/2)
			clicked = true
			break
		}
	}
	if !clicked || s.charView.stage != cvAction || s.charView.sel != 0 {
		t.Fatalf("item click failed: clicked=%v stage=%d sel=%d", clicked, s.charView.stage, s.charView.sel)
	}
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionQuit})
	modernShot(t, s, "quit-confirm")
	family, px := s.modern.family, s.modern.pixels
	modernClick(t, s, 250, 22)
	modernClick(t, s, 420, 22)
	modernClick(t, s, 820, 22)
	if !s.quitAsk || family != s.modern.family || px != s.modern.pixels {
		t.Fatal("quit confirmation allowed toolbar click-through")
	}
}

func TestModernFontCoversCatalogue(t *testing.T) {
	s := modernScene(t)
	chars := map[rune]bool{}
	s.cat.Each(func(_, text string) {
		for _, r := range text {
			if unicode.IsPrint(r) {
				chars[r] = true
			}
		}
	})
	for _, r := range "−†…" {
		chars[r] = true
	}
	for family := 0; family < 2; family++ {
		for r := range chars {
			if !s.modern.fonts.HasRune(family, r) {
				t.Errorf("font=%d missing %U %q", family, r, r)
			}
		}
	}
}

func TestModernAllSettingsAndMenuFieldsFit(t *testing.T) {
	s := modernScene(t)
	for family := 0; family < 2; family++ {
		for _, px := range []int{16, 18, 20} {
			s.modern.family, s.modern.pixels = family, px
			s.openSettings()
			modernShot(t, s, strings.Join([]string{"settings", string(rune('0' + family)), string(rune('0' + px/2))}, "-"))
			step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
		}
	}
}

func TestModernPortraitMenusAndSevenRosterRows(t *testing.T) {
	s := modernScene(t)
	step(t, s, key('E'))
	step(t, s, key('Y'))
	if !s.InCombat() {
		t.Fatal("normal empty encounter did not open combat UI")
	}
	t.Logf("Existing combat text, not synthesized: %q", s.cjk)
	for len(s.world.Party.Members) < 7 {
		copy := *s.world.Party.Members[0]
		copy.Name = "長名字測試角色"
		s.world.Party.Members = append(s.world.Party.Members, &copy)
	}
	for _, px := range []int{16, 18, 20} {
		s.modern.pixels = px
		s.dirty = true
		modernShot(t, s, fmt.Sprintf("combat-seven-%d", px))
		for _, heading := range []string{"姓名", "防護", "彈藥", "上限", "體力", "武器"} {
			found := false
			for _, field := range s.modern.fields {
				if field.Text == heading {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("roster heading absent: %s", heading)
			}
		}
		if len(s.modern.fields) == 0 {
			t.Fatal("no fields")
		}
	}
}

func TestModernIdleCombatDoesNotRepeatPrompt(t *testing.T) {
	s := modernScene(t)
	step(t, s, key('E'))
	step(t, s, key('Y'))
	beforeEN, beforeCJK := s.message, s.cjk
	for i := 0; i < 60; i++ {
		step(t, s, input.Input{Dir: input.DirNone})
	}
	if s.message != beforeEN || s.cjk != beforeCJK {
		t.Fatal("idle window polls appended repeated combat menus")
	}
	modernShot(t, s, "combat-idle")
}

func TestModernPortraitShowsThirteenthMenuRow(t *testing.T) {
	s := modernScene(t)
	step(t, s, key('E'))
	step(t, s, key('Y'))
	s.cjk = strings.Repeat("上一列\n", 12) + "Z 最後一列可選"
	s.message = ""
	s.modern.pixels = 20
	s.dirty = true
	modernShot(t, s, "combat-last-row")
	found := false
	for _, f := range s.modern.fields {
		if f.Text == "Z 最後一列可選" {
			found = true
		}
	}
	if !found {
		t.Fatal("thirteenth menu row was hidden or truncated")
	}
}

func TestModernRealFacilityRecordsRender(t *testing.T) {
	s := modernScene(t)
	resources, err := s.rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[game.FacilityKind]bool{}
	for _, res := range resources {
		b, err := s.rom.BlockByID(res.ID)
		if err != nil {
			t.Fatal(err)
		}
		for n := 0; n < b.SectionCount(6); n++ {
			rec, err := b.SectionRecord(6, n)
			if err != nil || len(rec) == 0 || rec[0]&0x80 == 0 {
				continue
			}
			kind := game.FacilityKind(rec[0] & 0x7f)
			if kind > game.FacilityRoster || seen[kind] {
				continue
			}
			// Each fixture starts with a clean player mode; LoadMap alone is a
			// diagnostic map change and deliberately does not dismiss old menus.
			s = modernScene(t)
			if err := s.LoadMap(res.ID, uint8(b.Dim/2), uint8(b.Dim/2)); err != nil {
				t.Fatal(err)
			}
			if s.EnterFacility(rec) == nil {
				continue
			}
			s.modern.pixels = 20
			s.dirty = true
			modernShot(t, s, fmt.Sprintf("facility-%d-20", kind))
			seen[kind] = true
			s.LeaveFacility()
		}
	}
	if len(seen) != 4 {
		t.Fatalf("facility types=%v", seen)
	}
	t.Log("Controlled renderer fixtures use actual facility records; not normal-player-path proof.")
}

func TestModernFacilityViaStartingPlayerPath(t *testing.T) {
	s := modernScene(t)
	x, y := s.world.Party.X, s.world.Party.Y
	n, _, rec, err := s.world.Block.At(int(x), int(y))
	t.Logf("shipped start (%d,%d), nibble=%d, record=%v, err=%v", x, y, n, rec, err)
	step(t, s, input.Input{Dir: input.DirUp})
	step(t, s, input.Input{Dir: input.DirDown})
	t.Logf("normal Up/Down: facility=%v roster=%v asking=%v", s.InFacility(), s.roster.active, s.asking)
	step(t, s, key('Y'))
	t.Logf("after Y: map=%d position=%d,%d facility=%v roster=%v asking=%v", s.blockID, s.world.Party.X, s.world.Party.Y, s.InFacility(), s.roster.active, s.asking)
	if !s.InFacility() || !s.roster.active {
		t.Fatal("normal start route did not reach Ranger Center")
	}
	if s.modernPopup() {
		t.Fatal("Ranger Center options should not cover the portrait twice")
	}
	modernShot(t, s, "facility-player-up-down")
	step(t, s, key('P'))
	if s.InFacility() || s.roster.active {
		t.Fatal("Ranger Center Play did not return to map")
	}
	modernShot(t, s, "facility-player-return")
}
