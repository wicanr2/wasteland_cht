package play

// Spec 34: catalogue values are presentation data. Switching languages must not
// rebuild the world, reset the journal or touch a save/asset identifier.
import (
	"fmt"
	"path/filepath"

	"github.com/wicanr2/wasteland_cht/internal/lang"
)

type localeBundle struct{ text, paragraphs *lang.Catalogue }

// LoadLanguages prepares both complete bundles before changing the active one.
// Custom single-catalogue callers can continue using LoadCatalogue/LoadJournal.
func (s *Scene) LoadLanguages(textDir, paragraphDir, initial string) error {
	if initial != "zh-Hant" && initial != "zh-Hans" {
		return fmt.Errorf("unsupported locale %q", initial)
	}
	if initial == "zh-Hans" && s.modern == nil {
		return fmt.Errorf("zh-Hans requires modern UI")
	}
	if s.journal == nil {
		return fmt.Errorf("journal is not initialized")
	}
	bundles := make(map[string]localeBundle, 2)
	for _, code := range []string{"zh-Hant", "zh-Hans"} {
		text, err := lang.Load(filepath.Join(textDir, code+".cat"))
		if err != nil {
			return err
		}
		paragraphs, err := lang.Load(filepath.Join(paragraphDir, "paragraphs-"+code+".cat"))
		if err != nil {
			return err
		}
		if text.Len() == 0 || paragraphs.Len() != 162 {
			return fmt.Errorf("incomplete locale %s", code)
		}
		bundles[code] = localeBundle{text, paragraphs}
	}
	s.locales = bundles
	s.applyLocale(initial)
	return nil
}

func (s *Scene) Locale() string { return s.locale }

func (s *Scene) LanguagesAvailable() bool { return s.modern != nil && len(s.locales) == 2 }

func (s *Scene) CanSwitchLocale() bool {
	if !s.LanguagesAvailable() {
		return false
	}
	if s.title {
		return !s.attract.active
	}
	view := *s
	view.settingsOpen = false
	return view.modernFreeMap()
}

func (s *Scene) SelectLocale(code string) error {
	if !s.CanSwitchLocale() {
		return fmt.Errorf("language switching requires title or free map")
	}
	if _, ok := s.locales[code]; !ok {
		return fmt.Errorf("unsupported locale %q", code)
	}
	s.applyLocale(code)
	return nil
}

func (s *Scene) applyLocale(code string) {
	b := s.locales[code]
	s.locale, s.cat, s.journal.paragraphs = code, b.text, b.paragraphs
	s.message, s.cjk = "", ""
	if s.settingsOpen {
		s.showSettings()
	} else if !s.title {
		s.sayPlace()
	}
	s.dirty = true
}
