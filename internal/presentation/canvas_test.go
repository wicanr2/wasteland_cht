package presentation

import (
	"image"
	"strings"
	"testing"
)

func TestModernTextInkIsBoundedAtEverySize(t *testing.T) {
	f, err := NewFonts()
	if err != nil {
		t.Fatal(err)
	}
	for family := 0; family < 2; family++ {
		for _, px := range []int{12, 13, 14, 15, 16, 18, 20} {
			c := NewCanvas(f, family)
			for _, text := range []string{"已讀取快速存檔。", "Hell Razor CON 25/25 $ 1,250", strings.Repeat("龜鬱體ABC，。", 20), "gy（）?", "†"} {
				for _, r := range []image.Rectangle{image.Rect(10, 10, 210, 38), image.Rect(10, 50, 70, 78)} {
					c.Label(text, r, px, Text, true)
				}
			}
			if len(c.Fields) == 0 {
				t.Fatal("no text was rendered")
			}
			for _, field := range c.Fields {
				if !field.Ink.In(field.Safe) {
					t.Errorf("family=%d px=%d field=%+v", family, px, field)
				}
			}
		}
	}
}

func TestModernToolbarHasDistinctBoundedTargets(t *testing.T) {
	controls := Controls()
	for i, a := range controls {
		if !a.Rect.In(image.Rect(0, 0, Width, 44)) {
			t.Fatal(a)
		}
		for _, b := range controls[i+1:] {
			if !a.Rect.Intersect(b.Rect).Empty() {
				t.Fatalf("overlapping targets: %s %s", a.ID, b.ID)
			}
		}
	}
}
