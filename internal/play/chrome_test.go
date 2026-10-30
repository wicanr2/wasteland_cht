package play

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

func hdChromeScene(t *testing.T) *Scene {
	t.Helper()
	for _, mode := range []string{"hd2-a", "hd2-b"} {
		if _, err := os.Stat("../../artpacks/" + mode + "/manifest.json"); err != nil {
			t.Skip("local-only HD2 art is required for this scene check")
		}
	}
	s := multilingualScene(t)
	s.SetArtRoot("../../artpacks")
	return s
}

func TestHDChromeTogglePreservesViewportAndSave(t *testing.T) {
	s := hdChromeScene(t)
	if !s.modern.hdChrome || s.modernHDChrome() {
		t.Fatal("A/B preference must default on without changing original art")
	}
	s.SetQuickSavePath(filepath.Join(t.TempDir(), "chrome.wlq"))
	for _, mode := range []string{"hd2-a", "hd2-b"} {
		if err := s.SelectArtMode(s.artRoot, mode); err != nil {
			t.Fatal(err)
		}
		if !s.modernHDChrome() {
			t.Fatal("HD chrome did not follow A/B")
		}
		modernShot(t, s, "chrome-"+mode+"-on")
		on := s.modern.frame
		if err := s.QuickSave(); err != nil {
			t.Fatal(err)
		}
		before, _ := os.ReadFile(s.QuickSavePath())
		world, journal, clock := s.world, s.journal.Journal, s.world.Clock
		modernClick(t, s, 908, 22)
		modernShot(t, s, "chrome-"+mode+"-settings-on")
		modernClick(t, s, 478, 183)
		if s.modernHDChrome() || !s.settingsOpen {
			t.Fatal("frame button did not toggle in settings")
		}
		step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
		modernShot(t, s, "chrome-"+mode+"-off")
		off := s.modern.frame
		for y := 84; y < 468; y++ {
			for x := 24; x < 888; x++ {
				if on.RGBAAt(x, y) != off.RGBAAt(x, y) {
					t.Fatalf("%s frame changed viewport at %d,%d", mode, x, y)
				}
			}
		}
		if on.RGBAAt(16, 76) == off.RGBAAt(16, 76) {
			t.Fatal("frame toggle did not visibly change the bezel")
		}
		if err := s.QuickSave(); err != nil {
			t.Fatal(err)
		}
		after, _ := os.ReadFile(s.QuickSavePath())
		if !bytes.Equal(before, after) || s.world != world || s.journal.Journal != journal || s.world.Clock != clock {
			t.Fatal("frame button changed game state or serialized save")
		}
		modernClick(t, s, 908, 22)
		step(t, s, key('C'))
		if !s.modernHDChrome() {
			t.Fatal("C did not restore HD chrome")
		}
		step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	}
	if err := s.SelectArtMode(s.artRoot, "original"); err != nil {
		t.Fatal(err)
	}
	if s.modernHDChrome() {
		t.Fatal("original art incorrectly acquired the new chrome")
	}
}

func TestHDChromeSettingsAndModes(t *testing.T) {
	s := hdChromeScene(t)
	if err := s.SelectArtMode(s.artRoot, "hd2-b"); err != nil {
		t.Fatal(err)
	}
	for _, locale := range []string{"zh-Hant", "zh-Hans"} {
		if err := s.SelectLocale(locale); err != nil {
			t.Fatal(err)
		}
		for family := 0; family < 2; family++ {
			for _, px := range []int{16, 18, 20} {
				for _, enabled := range []bool{true, false} {
					s.modern.family, s.modern.pixels, s.modern.hdChrome = family, px, enabled
					s.openSettings()
					modernShot(t, s, fmt.Sprintf("chrome-settings-%s-%d-%d-%v", locale, family, px, enabled))
					// The actual frame must not enter any visible popup text ink.
					for _, f := range s.modern.fields {
						if f.Safe.Min.Y >= 162 && f.Safe.Max.Y <= 646 && !f.Ink.In(image.Rect(24, 162, 940, 646)) {
							t.Fatalf("bezel overlaps text: %+v", f)
						}
					}
					step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
				}
			}
		}
	}
	s.modern.hdChrome = false
	if err := s.SelectArtMode(s.artRoot, "hd2-a"); err != nil {
		t.Fatal(err)
	}
	if s.modernHDChrome() {
		t.Fatal("art switch reset the chosen frame preference")
	}
	s.SetQuickSavePath(filepath.Join(t.TempDir(), "chrome.wlq"))
	if err := s.QuickSave(); err != nil {
		t.Fatal(err)
	}
	s.beginWipe()
	step(t, s, input.Input{Dir: input.DirNone, Fn: input.FnQuickLoad})
	if s.modern.hdChrome || s.Mode() != "map" {
		t.Fatal("F9 did not retain frame preference or leave death mode")
	}
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionQuit})
	modernClick(t, s, 478, 183)
	if s.modern.hdChrome || !s.quitAsk {
		t.Fatal("frame control click bypassed quit ownership")
	}
}

func TestHDChromeCombatPortraitIsNotCovered(t *testing.T) {
	s := hdChromeScene(t)
	if err := s.SelectArtMode(s.artRoot, "hd2-b"); err != nil {
		t.Fatal(err)
	}
	step(t, s, key('E'))
	step(t, s, key('Y'))
	if !s.InCombat() {
		t.Fatal("normal empty round did not open combat UI")
	}
	s.modern.pixels = 20
	s.dirty = true
	modernShot(t, s, "chrome-combat-on")
	on := s.modern.frame
	s.modern.hdChrome, s.dirty = false, true
	modernShot(t, s, "chrome-combat-off")
	for y := 84; y < 336; y++ {
		for x := 24; x < 312; x++ {
			if on.RGBAAt(x, y) != s.modern.frame.RGBAAt(x, y) {
				t.Fatalf("bezel covered portrait at %d,%d", x, y)
			}
		}
	}
}
