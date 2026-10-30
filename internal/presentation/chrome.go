package presentation

import "image"

// MetalBorder redraws the approved original-style silver bevel at native size.
// Spec 35: decoration occupies the border only, never the content rectangle.
func (c *Canvas) MetalBorder(r image.Rectangle, width int) {
	for n := 0; n < width; n++ {
		light, dark := RGB(0xe9edf0), RGB(0x11181e)
		switch {
		case n < 1:
		case n < 3:
			light, dark = RGB(0xb5bdc3), RGB(0x454f58)
		case n < width-2:
			light, dark = RGB(0x929ca5), RGB(0x66717b)
		case n < width-1:
			light, dark = RGB(0x454f58), RGB(0xb5bdc3)
		default:
			light, dark = RGB(0x11181e), RGB(0xe9edf0)
		}
		a := r.Inset(n)
		if a.Empty() {
			return
		}
		c.Fill(image.Rect(a.Min.X, a.Min.Y, a.Max.X, a.Min.Y+1), light)
		c.Fill(image.Rect(a.Min.X, a.Min.Y, a.Min.X+1, a.Max.Y), light)
		c.Fill(image.Rect(a.Min.X, a.Max.Y-1, a.Max.X, a.Max.Y), dark)
		c.Fill(image.Rect(a.Max.X-1, a.Min.Y, a.Max.X, a.Max.Y), dark)
	}
}
