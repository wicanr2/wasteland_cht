package play

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

const (
	ReimaginedDefaultWidth  = 1280
	ReimaginedDefaultHeight = 720
	reimaginedTileWidth     = 64
	reimaginedTileHeight    = 48
	// 新版圖磚是 3/4 俯視素材，不應像正交方格一樣首尾硬接。橫向重疊 4 px、
	// 縱向重疊 8 px，讓下一格的前景蓋過上一格底部陰影，也消除逐列黑縫。
	reimaginedTileStepX = 56
	reimaginedTileStepY = 36
)

var reimaginedTileMasks = makeReimaginedTileMasks()

func makeReimaginedTileMasks() [2][2]*image.Alpha {
	var masks [2][2]*image.Alpha
	for left := 0; left < 2; left++ {
		for top := 0; top < 2; top++ {
			m := image.NewAlpha(image.Rect(0, 0, reimaginedTileWidth, reimaginedTileHeight))
			for y := 0; y < reimaginedTileHeight; y++ {
				for x := 0; x < reimaginedTileWidth; x++ {
					a := 255
					if left != 0 && x < reimaginedTileWidth-reimaginedTileStepX {
						a = x * 255 / (reimaginedTileWidth - reimaginedTileStepX)
					}
					if top != 0 && y < reimaginedTileHeight-reimaginedTileStepY {
						v := y * 255 / (reimaginedTileHeight - reimaginedTileStepY)
						if v < a {
							a = v
						}
					}
					m.SetAlpha(x, y, color.Alpha{A: uint8(a)})
				}
			}
			masks[left][top] = m
		}
	}
	return masks
}

// SetArtViewport 接收呈現層算出的 16:9 邏輯畫布。只有全面重構模式使用；
// 另外兩種模式的固定畫布不受視窗大小影響。
func (s *Scene) SetArtViewport(width, height int) {
	if width < 960 || height < 540 {
		width, height = 960, 540
	}
	if width > 1920 {
		width = 1920
	}
	if height > 1080 {
		height = 1080
	}
	s.artWidth, s.artHeight = width, height
}

func (s *Scene) reimaginedFrame() (*image.RGBA, error) {
	if s.reimagined == nil {
		return nil, fmt.Errorf("reimagined 資產尚未載入")
	}
	w, h := s.artWidth, s.artHeight
	if w == 0 || h == 0 {
		w, h = ReimaginedDefaultWidth, ReimaginedDefaultHeight
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.RGBA{8, 10, 11, 255}}, image.Point{}, draw.Src)

	switch {
	case s.title:
		drawScale(dst, dst.Bounds(), s.reimagined.Title)
		return dst, nil
	case s.ending.active:
		drawScale(dst, dst.Bounds(), s.reimagined.Ending)
		return dst, nil
	}

	hudHeight := clamp(h/4, 150, 240)
	content := image.Rect(20, 20, w-20, h-hudHeight-12)
	if s.wipe.active {
		s.drawReimaginedScene(dst, content, wipePicture)
	} else if s.facility != nil {
		s.drawReimaginedScene(dst, content, s.facility.Picture)
		s.drawReimaginedPartyPanel(dst, content, "interact")
	} else if s.combat != nil {
		s.drawReimaginedScene(dst, content, s.portraitPicture())
		s.drawReimaginedPartyPanel(dst, content, "combat")
	} else if err := s.drawReimaginedMap(dst, content); err != nil {
		return nil, err
	}
	s.drawReimaginedHUD(dst, image.Rect(20, h-hudHeight, w-20, h-12))
	return dst, nil
}

func (s *Scene) drawReimaginedPartyPanel(dst *image.RGBA, rect image.Rectangle, context string) {
	count := len(s.world.Party.Members)
	if count > 7 {
		count = 7
	}
	if count == 0 {
		return
	}
	spacing := clamp(rect.Dx()/(count+1), 104, 144)
	width := spacing * count
	startX := rect.Min.X + (rect.Dx()-width)/2
	for i := 0; i < count; i++ {
		member := s.world.Party.Members[i]
		if member == nil {
			continue
		}
		state, frame := "idle", 0
		direction := i % 8
		switch context {
		case "interact":
			state, frame = "interact", s.artAnimFrame%2
		case "combat":
			switch {
			case member.CON <= 0:
				state = "down"
			case member.CON < member.MaxCON/2:
				state, frame = "hurt", s.artAnimFrame%2
			case s.combat.Phase != nil && i < len(s.combat.Phase.Cmd) && s.combat.Phase.Cmd[i] == game.CmdAttack:
				state, frame = "attack", s.artAnimFrame%3
			}
		}
		character := s.reimagined.Character(i, direction, state, frame)
		if character == nil {
			continue
		}
		x := startX + i*spacing
		y := rect.Max.Y - 104
		drawScaledOver(dst, image.Rect(x, y, x+96, y+96), character, rect)
		if state != "down" {
			kind := s.reimaginedWeapon(member)
			if weapon := s.reimagined.Weapon(kind, direction); weapon != nil {
				drawScaledOver(dst, image.Rect(x, y, x+96, y+96), weapon, rect)
			}
		}
	}
}

func (s *Scene) reimaginedWeapon(member *game.Character) string {
	if member == nil {
		return "blade"
	}
	slot, ok := equippedSlot(member)
	if !ok {
		return "blade"
	}
	item, ok := s.items.Get(slot.ID)
	if !ok {
		return "blade"
	}
	switch item.Class {
	case game.ClassMelee:
		return "blade"
	case game.ClassPistol, game.ClassLaserPist:
		return "pistol"
	case game.ClassATLight, game.ClassATHeavy, game.ClassEnergyHigh, game.ClassExplosive:
		return "launcher"
	default:
		return "rifle"
	}
}

func drawScaledOver(dst draw.Image, target image.Rectangle, src image.Image, clip image.Rectangle) {
	tmp := image.NewRGBA(image.Rect(0, 0, target.Dx(), target.Dy()))
	drawScale(tmp, tmp.Bounds(), src)
	r := target.Intersect(clip)
	if !r.Empty() {
		draw.Draw(dst, r, tmp, image.Pt(r.Min.X-target.Min.X, r.Min.Y-target.Min.Y), draw.Over)
	}
}

func (s *Scene) drawReimaginedScene(dst *image.RGBA, rect image.Rectangle, index int) {
	if index < 0 || index >= len(s.reimagined.Scenes) {
		return
	}
	drawContain(dst, rect, s.reimagined.Scenes[index])
}

func (s *Scene) drawReimaginedMap(dst *image.RGBA, rect image.Rectangle) error {
	set := s.reimagined.Map
	if set == nil {
		return fmt.Errorf("reimagined 地圖資產為空")
	}
	cols := clamp(rect.Dx()/reimaginedTileStepX+2, 18, 25)
	rows := clamp(rect.Dy()/reimaginedTileStepY+2, 8, 15)
	originX := int(s.world.Party.X) - cols/2
	originY := int(s.world.Party.Y) - rows/2
	mapW := (cols-1)*reimaginedTileStepX + reimaginedTileWidth
	mapH := (rows-1)*reimaginedTileStepY + reimaginedTileHeight
	startX := rect.Min.X + (rect.Dx()-mapW)/2
	startY := rect.Min.Y + (rect.Dy()-mapH)/2
	clip := rect

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			mx, my := originX+col, originY+row
			code := s.world.Block.OutsideGraphic()
			if mx >= 0 && my >= 0 && mx < s.world.Block.Dim && my < s.world.Block.Dim {
				code = s.world.Block.Graphic[my*s.world.Block.Dim+mx]
			}
			var src image.Image
			if int(code) < len(set.Icons) {
				src = set.Icons[code]
			} else {
				i := int(code) - 10
				if i < 0 || i >= len(set.Tiles) {
					return fmt.Errorf("圖形編號 %d 超出 reimagined 資產", code)
				}
				src = set.Tiles[i]
			}
			x := startX + col*reimaginedTileStepX
			y := startY + row*reimaginedTileStepY
			r := image.Rect(x, y, x+src.Bounds().Dx(), y+src.Bounds().Dy()).Intersect(clip)
			if !r.Empty() {
				left, top := 0, 0
				if col > 0 {
					left = 1
				}
				if row > 0 {
					top = 1
				}
				draw.DrawMask(dst, r, src, image.Pt(r.Min.X-x, r.Min.Y-y),
					reimaginedTileMasks[left][top], image.Pt(r.Min.X-x, r.Min.Y-y), draw.Over)
			}
		}
	}

	// 同一遊戲格內以小隊陣形呈現最多七人；規則座標仍只有一格。
	partyX := startX + (int(s.world.Party.X)-originX)*reimaginedTileStepX
	partyY := startY + (int(s.world.Party.Y)-originY)*reimaginedTileStepY
	// 縮進影片或小視窗時，暗色角色會融進棕色沙地。定位環不是遊戲規則物件，
	// 只是一層低彩度 HUD 提示；先畫陰影再畫雙環，讓主角在動態背景上仍可辨識。
	drawReimaginedPartyMarker(dst, partyX+reimaginedTileWidth/2, partyY+reimaginedTileHeight/2, clip)
	direction := [...]int{0, 2, 4, 6}[clamp(s.artPartyDir, 0, 3)]
	formation := [7]image.Point{{0, -24}, {-26, -10}, {26, -10}, {-36, 10}, {36, 10}, {-17, 27}, {17, 27}}
	count := len(s.world.Party.Members)
	if count > 7 {
		count = 7
	}
	for i := 0; i < count; i++ {
		state, frame := "idle", 0
		if s.artPartyTicks != 0 {
			state, frame = "walk", s.artPartyFrame%4
		}
		character := s.reimagined.Character(i, direction, state, frame)
		if character == nil {
			continue
		}
		x := partyX - 8 + formation[i].X
		y := partyY - 54 + formation[i].Y
		target := image.Rect(x, y, x+80, y+80)
		drawScaledOver(dst, target, character, clip)
		weapon := s.reimagined.Weapon(s.reimaginedWeapon(s.world.Party.Members[i]), direction)
		if weapon != nil {
			drawScaledOver(dst, target, weapon, clip)
		}
	}
	return nil
}

func drawReimaginedPartyMarker(dst draw.Image, cx, cy int, clip image.Rectangle) {
	for y := -18; y <= 18; y++ {
		for x := -52; x <= 52; x++ {
			d := x*x*9 + y*y*64
			p := image.Pt(cx+x, cy+y)
			if !p.In(clip) {
				continue
			}
			if d <= 52*52*9 && d >= 48*48*9 {
				dst.Set(p.X, p.Y, color.RGBA{205, 166, 72, 210})
			}
		}
	}
}

func (s *Scene) drawReimaginedHUD(dst *image.RGBA, rect image.Rectangle) {
	panel := &image.Uniform{C: color.RGBA{12, 15, 16, 245}}
	draw.Draw(dst, rect, panel, image.Point{}, draw.Src)
	// 沿用已驗證的中文字形與遊戲訊息，但**不把原版外框一起縮放**。原本把
	// HiFrame 最底 120 px 整塊拉進來，會把中文指令壓到邊線上；這裡裁出
	// 欄 1–38、列 18–24 的內容安全區，再留 14 px 新版 HUD 內距。
	hi := s.HiFrame().ToImage()
	frame, content, source := reimaginedHUDRects(rect)
	drawScale(dst, content, hi.SubImage(source))
	drawReimaginedHUDBorder(dst, frame)
}

func reimaginedHUDRects(rect image.Rectangle) (frame, content, source image.Rectangle) {
	frame = rect
	content = image.Rect(rect.Min.X+14, rect.Min.Y+14, rect.Max.X-14, rect.Max.Y-14)
	source = image.Rect(render.HiScale*8, render.MsgRow*render.HiScale*8,
		render.HiScreenWidth-render.HiScale*8, render.HiScreenHeight)
	return
}

func drawReimaginedHUDBorder(dst draw.Image, rect image.Rectangle) {
	outer := color.RGBA{80, 91, 88, 255}
	inner := color.RGBA{184, 139, 64, 220}
	for i, c := range []color.RGBA{outer, inner} {
		r := image.Rect(rect.Min.X+i, rect.Min.Y+i, rect.Max.X-i, rect.Max.Y-i)
		draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+1), &image.Uniform{C: c}, image.Point{}, draw.Src)
		draw.Draw(dst, image.Rect(r.Min.X, r.Max.Y-1, r.Max.X, r.Max.Y), &image.Uniform{C: c}, image.Point{}, draw.Src)
		draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Min.X+1, r.Max.Y), &image.Uniform{C: c}, image.Point{}, draw.Src)
		draw.Draw(dst, image.Rect(r.Max.X-1, r.Min.Y, r.Max.X, r.Max.Y), &image.Uniform{C: c}, image.Point{}, draw.Src)
	}
}

func drawScale(dst draw.Image, target image.Rectangle, src image.Image) {
	if src == nil || target.Empty() {
		return
	}
	sb := src.Bounds()
	for y := target.Min.Y; y < target.Max.Y; y++ {
		sy := sb.Min.Y + (y-target.Min.Y)*sb.Dy()/target.Dy()
		for x := target.Min.X; x < target.Max.X; x++ {
			sx := sb.Min.X + (x-target.Min.X)*sb.Dx()/target.Dx()
			dst.Set(x, y, src.At(sx, sy))
		}
	}
}

func drawContain(dst draw.Image, target image.Rectangle, src image.Image) {
	if src == nil || target.Empty() {
		return
	}
	sb := src.Bounds()
	w, h := target.Dx(), target.Dy()
	if w*sb.Dy() > h*sb.Dx() {
		w = h * sb.Dx() / sb.Dy()
	} else {
		h = w * sb.Dy() / sb.Dx()
	}
	x := target.Min.X + (target.Dx()-w)/2
	y := target.Min.Y + (target.Dy()-h)/2
	drawScale(dst, image.Rect(x, y, x+w, y+h), src)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
