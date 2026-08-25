package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

const FaithfulTileSize = TileSize * HiScale // 48

// DrawFaithfulMap replaces only the 288×128 map viewport on a 960×600 RGBA
// frame. HUD and CJK geometry stay on the existing verified HiFrame path.
func DrawFaithfulMap(dst draw.Image, b *assets.Block, tiles, icons []image.Image, originX, originY int) error {
	clip := image.Rect(ViewX*HiScale, (ViewY+1)*HiScale, (ViewX+ViewWidth)*HiScale, (ViewY+ViewHeight)*HiScale)
	draw.Draw(dst, clip, &image.Uniform{C: color.Black}, image.Point{}, draw.Src)
	for row := 0; row < ViewRows; row++ {
		for col := 0; col < ViewCols; col++ {
			mx, my := originX+col, originY+row
			code := b.OutsideGraphic()
			if mx >= 0 && my >= 0 && mx < b.Dim && my < b.Dim {
				code = b.Graphic[my*b.Dim+mx]
			}
			var src image.Image
			if int(code) < len(icons) {
				src = icons[code]
			} else {
				i := int(code) - 10
				if i < 0 || i >= len(tiles) {
					return fmt.Errorf("圖形編號 %d 超出 faithful 資產（%d icons＋%d tiles）", code, len(icons), len(tiles))
				}
				src = tiles[i]
			}
			x := ViewX*HiScale - FaithfulTileSize/2 + col*FaithfulTileSize
			y := ViewY*HiScale - FaithfulTileSize/2 + row*FaithfulTileSize
			r := image.Rect(x, y, x+FaithfulTileSize, y+FaithfulTileSize).Intersect(clip)
			if !r.Empty() {
				draw.Draw(dst, r, src, image.Pt(r.Min.X-x, r.Min.Y-y), draw.Over)
			}
		}
	}
	return nil
}

func DrawFaithfulIcon(dst draw.Image, icons []image.Image, icon byte, col, row int) error {
	if int(icon) >= len(icons) {
		return fmt.Errorf("faithful 疊圖 %d 超出範圍（%d）", icon, len(icons))
	}
	x := ViewX*HiScale - FaithfulTileSize/2 + col*FaithfulTileSize
	y := ViewY*HiScale - FaithfulTileSize/2 + row*FaithfulTileSize
	clip := image.Rect(ViewX*HiScale, (ViewY+1)*HiScale, (ViewX+ViewWidth)*HiScale, (ViewY+ViewHeight)*HiScale)
	r := image.Rect(x, y, x+FaithfulTileSize, y+FaithfulTileSize).Intersect(clip)
	if !r.Empty() {
		draw.Draw(dst, r, icons[icon], image.Pt(r.Min.X-x, r.Min.Y-y), draw.Over)
	}
	return nil
}
