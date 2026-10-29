package play

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/lang"
)

func multilingualScene(t *testing.T) *Scene {
	t.Helper()
	s := modernScene(t)
	if err := s.LoadJournal("../../docs/re/generated/paragraph-refs.tsv", "../../translations/paragraphs-zh-Hant.cat"); err != nil {
		t.Fatal(err)
	}
	if err := s.LoadLanguages("../../translations", "../../translations", "zh-Hant"); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestLocaleCataloguesPreserveKeysControlsAndFonts(t *testing.T) {
	s := multilingualScene(t)
	formats := regexp.MustCompile(`%(?:[-+#0 ]*\d*(?:\.\d+)?)?[a-zA-Z%]`)
	for _, prefix := range []string{"", "paragraphs-"} {
		a, err := lang.Load("../../translations/" + prefix + "zh-Hant.cat")
		if err != nil {
			t.Fatal(err)
		}
		b, err := lang.Load("../../translations/" + prefix + "zh-Hans.cat")
		if err != nil {
			t.Fatal(err)
		}
		if a.Len() != b.Len() {
			t.Fatal("locale key counts differ")
		}
		a.Each(func(key, text string) {
			other, ok := b.Lookup(key)
			if !ok || other == "" {
				t.Fatalf("missing %s", key)
			}
			for _, code := range []byte{10, 12, 13, 14, 15, 16} {
				if strings.Count(text, string(code)) != strings.Count(other, string(code)) {
					t.Errorf("control %d changed in %s", code, key)
				}
			}
			if strings.Join(formats.FindAllString(text, -1), "|") != strings.Join(formats.FindAllString(other, -1), "|") {
				t.Errorf("format changed in %s", key)
			}
			for _, value := range []string{text, other} {
				for _, r := range value {
					if unicode.IsPrint(r) {
						for family := 0; family < 2; family++ {
							if !s.modern.fonts.HasRune(family, r) {
								t.Errorf("font %d missing %U %s", family, r, key)
							}
						}
					}
				}
			}
		})
	}
}

func TestLocaleToolbarSwitchPreservesSaveAndJournal(t *testing.T) {
	s := multilingualScene(t)
	s.SetQuickSavePath(filepath.Join(t.TempDir(), "before.wlqs"))
	if err := s.QuickSave(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(s.QuickSavePath())
	if err != nil {
		t.Fatal(err)
	}
	journal, world := s.journal.Journal, s.world
	modernClick(t, s, 908, 22)
	modernShot(t, s, "locale-hant-settings")
	modernClick(t, s, 728, 183)
	if s.Locale() != "zh-Hans" || !strings.Contains(s.cjk, "语言") {
		t.Fatalf("locale button failed: %s %q", s.Locale(), s.cjk)
	}
	modernShot(t, s, "locale-hans-settings")
	if s.journal.Journal != journal || s.world != world {
		t.Fatal("language switch replaced rule state")
	}
	if s.modern.pixels != 18 || s.modern.family != 0 {
		t.Fatal("language switch changed font")
	}
	step(t, s, key('L'))
	if s.Locale() != "zh-Hant" {
		t.Fatal("language keyboard toggle failed")
	}
	step(t, s, key('L'))
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	if err := s.QuickSave(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(s.QuickSavePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("language switching changed serialized game state")
	}
	step(t, s, key('1'))
	modernShot(t, s, "locale-hans-character")
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionConfirm})
	modernShot(t, s, "locale-hans-items")
	if !strings.Contains(s.cjk, "弹匣") {
		t.Fatalf("item view did not use Hans catalogue: %q", s.cjk)
	}
	if err := s.SelectLocale("zh-Hant"); err == nil {
		t.Fatal("switch accepted unfinished item action")
	}
}

func TestLocaleLoadFailureIsAtomicAndClassicIsExplicit(t *testing.T) {
	s := multilingualScene(t)
	cat, paragraphs, locale := s.cat, s.journal.paragraphs, s.locale
	if err := s.LoadLanguages("../../translations", t.TempDir(), "zh-Hans"); err == nil {
		t.Fatal("missing bundle accepted")
	}
	if s.cat != cat || s.journal.paragraphs != paragraphs || s.locale != locale {
		t.Fatal("partial locale load committed")
	}
	classic := newScene(t)
	if err := classic.LoadLanguages("../../translations", "../../translations", "zh-Hans"); err == nil {
		t.Fatal("classic claims Hans support")
	}
}
