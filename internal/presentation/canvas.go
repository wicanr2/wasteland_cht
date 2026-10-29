// Package presentation renders the approved rounded interface without Ebiten.
package presentation

import (
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

//go:embed fonts/*
var files embed.FS

const Width, Height = 960, 720

var (
	Background = RGB(0x101820)
	Panel      = RGB(0x17232e)
	Edge       = RGB(0x46535d)
	Text       = RGB(0xe7ebed)
	Muted      = RGB(0x9daab5)
	Accent     = RGB(0xe2c58c)
	MapRect    = image.Rect(24, 84, 888, 468)
)

func RGB(v uint32) color.RGBA { return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255} }

var loadOnce sync.Once
var loaded [2]*opentype.Font
var loadError error

func loadFonts() {
	for i, name := range []string{"GenSenRounded-TW-Regular.otf.gz", "NotoSansCJKtc-Regular.otf.gz"} {
		b, err := files.ReadFile("fonts/" + name)
		if err == nil {
			var z *gzip.Reader
			z, err = gzip.NewReader(bytes.NewReader(b))
			if err == nil {
				b, err = io.ReadAll(z)
				_ = z.Close()
			}
		}
		if err == nil {
			loaded[i], err = opentype.Parse(b)
		}
		if err != nil {
			loadError = fmt.Errorf("load UI font %s: %w", name, err)
			return
		}
	}
}

type Fonts struct{ faces map[[2]int]font.Face }

func NewFonts() (*Fonts, error) {
	loadOnce.Do(loadFonts)
	if loadError != nil {
		return nil, loadError
	}
	return &Fonts{faces: make(map[[2]int]font.Face)}, nil
}

func (f *Fonts) Face(family, size int) font.Face {
	if family != 1 {
		family = 0
	}
	k := [2]int{family, size}
	if face := f.faces[k]; face != nil {
		return face
	}
	face, err := opentype.NewFace(loaded[family], &opentype.FaceOptions{Size: float64(size), DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		panic(err) // Font parsing was checked before enabling the presentation.
	}
	f.faces[k] = face
	return face
}

func (f *Fonts) HasRune(family int, r rune) bool {
	var b sfnt.Buffer
	g, err := loaded[family].GlyphIndex(&b, r)
	return err == nil && g != 0
}

type Field struct {
	Safe, Ink image.Rectangle
	Text      string
	Size      int
}

type Canvas struct {
	Image  *image.RGBA
	Fonts  *Fonts
	Family int
	Fields []Field
}

func NewCanvas(f *Fonts, family int) *Canvas {
	im := image.NewRGBA(image.Rect(0, 0, Width, Height))
	draw.Draw(im, im.Bounds(), image.NewUniform(Background), image.Point{}, draw.Src)
	return &Canvas{Image: im, Fonts: f, Family: family}
}

func (c *Canvas) Fill(r image.Rectangle, col color.Color) {
	draw.Draw(c.Image, r, image.NewUniform(col), image.Point{}, draw.Src)
}

func rounded(dst *image.RGBA, r image.Rectangle, radius float32, col color.Color) {
	if r.Empty() {
		return
	}
	x0, y0, x1, y1 := float32(0), float32(0), float32(r.Dx()), float32(r.Dy())
	v := vector.NewRasterizer(r.Dx(), r.Dy())
	v.MoveTo(x0+radius, y0)
	v.LineTo(x1-radius, y0)
	v.QuadTo(x1, y0, x1, y0+radius)
	v.LineTo(x1, y1-radius)
	v.QuadTo(x1, y1, x1-radius, y1)
	v.LineTo(x0+radius, y1)
	v.QuadTo(x0, y1, x0, y1-radius)
	v.LineTo(x0, y0+radius)
	v.QuadTo(x0, y0, x0+radius, y0)
	v.ClosePath()
	tmp := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	v.Draw(tmp, tmp.Bounds(), image.NewUniform(col), image.Point{})
	draw.Draw(dst, r, tmp, image.Point{}, draw.Over)
}

func (c *Canvas) Box(r image.Rectangle, fill, edge color.Color, radius int) {
	rounded(c.Image, r, float32(radius), edge)
	rounded(c.Image, r.Inset(1), float32(max(0, radius-1)), fill)
}

func (c *Canvas) Measure(text string, px int) int {
	return font.MeasureString(c.Fonts.Face(c.Family, px), text).Ceil()
}

// Label uses measured ink and advance, never painter clipping as an overflow policy.
func (c *Canvas) Label(text string, safe image.Rectangle, px int, col color.Color, centered bool) image.Rectangle {
	if safe.Empty() || text == "" {
		return image.Rectangle{}
	}
	face := c.Fonts.Face(c.Family, px)
	fits := func(s string) bool {
		b, advance := font.BoundString(face, s)
		return max(advance.Ceil(), b.Max.X.Ceil())-min(0, b.Min.X.Floor()) <= safe.Dx() && b.Max.Y.Ceil()-b.Min.Y.Floor() <= safe.Dy()
	}
	if !fits(text) {
		runes := []rune(text)
		for len(runes) > 0 && !fits(string(runes)+"…") {
			runes = runes[:len(runes)-1]
		}
		text = string(runes) + "…"
		if !fits(text) {
			return image.Rectangle{}
		}
	}
	b, advance := font.BoundString(face, text)
	w, h := max(advance.Ceil(), b.Max.X.Ceil())-min(0, b.Min.X.Floor()), b.Max.Y.Ceil()-b.Min.Y.Floor()
	x := safe.Min.X - min(0, b.Min.X.Floor())
	if centered {
		x += (safe.Dx() - w) / 2
	}
	y := safe.Min.Y + (safe.Dy()-h)/2 - b.Min.Y.Floor()
	d := font.Drawer{Dst: c.Image, Src: image.NewUniform(col), Face: face, Dot: fixed.P(x, y)}
	d.DrawString(text)
	ink := image.Rect(x+b.Min.X.Floor(), y+b.Min.Y.Floor(), x+b.Max.X.Ceil(), y+b.Max.Y.Ceil())
	c.Fields = append(c.Fields, Field{Safe: safe, Ink: ink, Text: text, Size: px})
	return ink
}

// Controls is the same geometry as the accepted native-size proposal.
type Control struct {
	ID   string
	Rect image.Rectangle
}

func Controls() []Control {
	return []Control{
		{"original", image.Rect(64, 9, 120, 35)}, {"hd2-a", image.Rect(124, 9, 158, 35)}, {"hd2-b", image.Rect(162, 9, 196, 35)},
		{"font", image.Rect(212, 6, 316, 38)}, {"size-minus", image.Rect(324, 6, 354, 38)}, {"size-plus", image.Rect(406, 6, 436, 38)},
		{"score", image.Rect(444, 6, 564, 38)}, {"music", image.Rect(572, 6, 610, 38)}, {"volume-minus", image.Rect(610, 6, 632, 38)}, {"music-value", image.Rect(632, 6, 656, 38)}, {"volume-plus", image.Rect(656, 6, 684, 38)},
		{"save", image.Rect(692, 6, 772, 38)}, {"load", image.Rect(780, 6, 860, 38)}, {"settings", image.Rect(868, 6, 948, 38)},
	}
}
