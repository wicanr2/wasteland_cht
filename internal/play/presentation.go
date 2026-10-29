package play

// The approved presentation is a view of existing text cells and actions.
// Rules, menu identifiers and save data remain on their existing paths (spec 33).
import (
	"fmt"
	"image"
	"image/draw"
	"sort"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/presentation"
	"github.com/wicanr2/wasteland_cht/internal/render"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

type modernHit struct {
	rect     image.Rectangle
	col, row int
	key      byte
}

type modernView struct {
	fonts          *presentation.Fonts
	family, pixels int
	frame          *image.RGBA
	last           *render.Frame
	hits           []modernHit
	fields         []presentation.Field
	scroll         int
	scrollKey      string
}

func (s *Scene) EnableModernUI() error {
	f, err := presentation.NewFonts()
	if err != nil {
		return err
	}
	s.modern = &modernView{fonts: f, pixels: 18}
	s.dirty = true
	return nil
}

func (s *Scene) ModernEnabled() bool { return s.modern != nil }

func (s *Scene) modernText(key, fallback string) string {
	if v := s.uiText("modern." + key); v != "" {
		return v
	}
	return fallback
}

func (s *Scene) modernPopup() bool {
	if s.settingsOpen || s.help || s.quitAsk || s.confirm.active || s.journalOpen {
		return true
	}
	// Portrait scenes already have a full-height menu. An additional dialog
	// would repeat the same options while hiding the portrait and party.
	if s.combat != nil || s.facility != nil {
		return false
	}
	return s.charView.stage != cvOff || s.roster.active || s.use.stage != useStageOff ||
		s.question.active || s.loot.active || s.order.active || s.disband || s.encAsk != 0 || s.asking != input.DirNone
}

func (s *Scene) modernFreeMap() bool {
	return !s.title && !s.ending.active && !s.wipe.active && s.combat == nil && s.facility == nil && !s.modernPopup()
}

func (s *Scene) ModernFrame() (*image.RGBA, error) {
	if s.modern == nil {
		return nil, fmt.Errorf("modern UI is not enabled")
	}
	v := s.modern
	low := s.Frame()
	if v.last == low && v.frame != nil {
		return v.frame, nil
	}
	h := s.HiFrame()
	c := presentation.NewCanvas(v.fonts, v.family)
	v.hits = v.hits[:0]
	s.modernToolbar(c)
	c.Label(s.modernText("title", "Wasteland"), image.Rect(20, 54, 220, 76), 20, presentation.Accent, false)
	mode := map[string]string{"original": s.modernText("original", "Original art"), "hd2-a": s.modernText("art-a", "HD2 A"), "hd2-b": s.modernText("art-b", "HD2 B")}[s.ArtMode()]
	c.Label(mode, image.Rect(232, 54, 560, 76), 14, presentation.Muted, false)
	c.Label(fmt.Sprintf("%02d:%02d", s.world.Clock.Hour, s.world.Clock.Minute), image.Rect(806, 54, 892, 76), 16, presentation.Muted, true)
	content := h.ToImage()
	offset := 0
	if s.faithful != nil {
		var err error
		content, err = s.ArtFrame()
		if err != nil {
			return nil, err
		}
		offset = faithfulContentY
	}
	portrait := s.combat != nil || s.facility != nil
	if portrait {
		c.Box(image.Rect(20, 80, 316, 406), presentation.Panel, presentation.Edge, 5)
		draw.Draw(c.Image, image.Rect(24, 84, 312, 372), content, image.Pt(24, 24+offset), draw.Src)
		c.Box(image.Rect(332, 80, 948, 406), presentation.Panel, presentation.Edge, 5)
		menuCells := h.TextCells
		if current := s.modernMessageCells(); len(current) != 0 {
			menuCells = current
		}
		if !s.modernPopup() {
			s.modernRows(c, menuCells, 1, 13, 15, 38, image.Rect(348, 90, 932, 402), 24, false)
		}
		s.modernRows(c, h.TextCells, 12, 12, 0, 13, image.Rect(32, 378, 304, 400), 24, false)
		for _, label := range s.boxLabels() {
			var r image.Rectangle
			var glyph string
			var key byte
			switch label.Key {
			case render.KeyArrowUp:
				r, glyph, key = image.Rect(872, 84, 900, 112), "↑", keyPrevPage
			case render.KeyArrowDown:
				r, glyph, key = image.Rect(904, 84, 932, 112), "↓", keyNextPage
			default:
				continue
			}
			c.Box(r, presentation.Panel, presentation.Edge, 3)
			c.Label(glyph, r.Inset(2), 16, presentation.Accent, true)
			v.hits = append(v.hits, modernHit{rect: r, key: key})
		}
	} else {
		c.Box(image.Rect(20, 80, 892, 472), presentation.RGB(0x090f14), presentation.RGB(0x667078), 5)
		draw.Draw(c.Image, presentation.MapRect, content, image.Pt(24, 24+offset), draw.Src)
		if s.attract.active {
			c.Fill(presentation.MapRect, presentation.Background)
			s.modernRows(c, h.TextCells, 0, 16, 0, 39, presentation.MapRect.Inset(12), 22, false)
		}
		s.modernMeter(c)
	}
	if portrait || s.mapRosterShown() {
		c.Box(image.Rect(16, 420, 948, 654), presentation.Panel, presentation.Edge, 6)
		s.modernRoster(c)
	} else {
		c.Box(image.Rect(16, 486, 948, 654), presentation.Panel, presentation.Edge, 6)
		c.Label(s.modernText("message", "Message"), image.Rect(32, 498, 140, 520), 14, presentation.Accent, false)
		c.Label(fmt.Sprintf(s.modernText("group", "Group %d"), s.groupID+1), image.Rect(816, 498, 932, 520), 14, presentation.Muted, true)
		c.Fill(image.Rect(32, 528, 932, 529), presentation.RGB(0x31414e))
		cells := s.modernMessageCells()
		rows := cellRows(cells)
		key := s.message + "\x00" + s.cjk
		if key != v.scrollKey {
			v.scrollKey = key
			v.scroll = 0
		}
		start := max(0, len(rows)-3-v.scroll)
		end := min(len(rows), start+3)
		for i, row := range rows[start:end] {
			s.modernRows(c, cells, row, row, s.msgRect().Col, s.msgRect().Col+s.msgRect().Width-1, image.Rect(32, 542+i*26, 916, 568+i*26), 26, false)
		}
		c.Label(s.modernText("scroll", "PageUp / PageDown"), image.Rect(32, 620, 620, 644), 13, presentation.Muted, false)
	}
	s.modernCommands(c)
	if s.title && !s.attract.active {
		c.Box(image.Rect(352, 542, 608, 590), presentation.RGB(0x30404c), presentation.Accent, 5)
		c.Label(s.modernText("start", "Start [S]"), image.Rect(368, 550, 592, 582), 18, presentation.Accent, true)
		v.hits = append(v.hits, modernHit{rect: image.Rect(352, 542, 608, 590), key: 'S'})
	}
	if s.modernPopup() {
		// Expand interactive text instead of hiding half of the six-line menus.
		c.Box(image.Rect(16, 154, 948, 654), presentation.Panel, presentation.Accent, 6)
		c.Label(s.modernText("dialog", "Menu"), image.Rect(32, 170, 780, 198), 16, presentation.Accent, false)
		c.Label("ESC", image.Rect(864, 170, 924, 198), 14, presentation.Muted, true)
		c.Fill(image.Rect(32, 208, 924, 209), presentation.Edge)
		v.hits = v.hits[:0]
		rect := s.msgRect()
		s.modernRows(c, s.modernMessageCells(), rect.Row, rect.LastRow(), rect.Col, rect.Col+rect.Width-1, image.Rect(32, 222, 924, 626), 28, false)
		v.hits = append(v.hits, modernHit{rect: image.Rect(852, 162, 936, 204), key: 0x1b})
		if s.settingsOpen && s.LanguagesAvailable() {
			r := image.Rect(612, 162, 844, 204)
			col := presentation.Text
			if !s.CanSwitchLocale() {
				col = presentation.Muted
			}
			c.Box(r, presentation.Panel, presentation.Edge, 4)
			c.Label("L "+s.uiText("settings.language")+s.uiText("language."+s.locale), r.Inset(8), 16, col, true)
			if s.CanSwitchLocale() {
				v.hits = append(v.hits, modernHit{rect: r, key: 'L'})
			}
		}
	}
	v.fields = c.Fields
	v.frame, v.last = c.Image, low
	return v.frame, nil
}

// Named fields share their header and value rectangles; proportional fonts must
// not infer tab stops from runs of spaces in the legacy text grid.
func (s *Scene) modernRoster(c *presentation.Canvas) {
	type column struct {
		lo, hi        int
		key, fallback string
	}
	columns := []column{{32, 72, "", "#"}, {80, 320, "combat.hdrname", "Name"}, {330, 390, "combat.hdrac", "AC"},
		{398, 458, "combat.hdrammo", "Ammo"}, {466, 538, "combat.hdrmax", "Max"}, {546, 634, "combat.hdrcon", "CON"}, {642, 932, "combat.hdrweapon", "Weapon"}}
	for i, col := range columns {
		label := col.fallback
		if v := s.uiText(col.key); v != "" {
			label = v
		}
		c.Label(label, image.Rect(col.lo, 432, col.hi, 458), 14, presentation.Muted, i >= 2 && i <= 5)
	}
	rows := Roster(s.world.Party, s.items, s.itemName)
	for i, row := range rows {
		if i >= 7 {
			break
		}
		row.Index = i + 1
		row.IndexInverse = row.Index == s.selectedMember()
		con := row.CON
		if key, ok := woundKeys[con]; ok {
			if t := s.uiText(key); t != "" {
				con = t
			}
		}
		con = strings.ReplaceAll(con, string(game.WoundDead[0]), "†")
		weapon := row.Weapon
		if slot, ok := equippedSlot(s.world.Party.Members[i]); ok {
			if t := s.itemNameCJK(slot.ID); t != "" {
				weapon = t
			}
		}
		values := []string{fmt.Sprintf("%d>", row.Index), row.Name, row.AC, row.Ammo, row.MaxCON, con, weapon}
		inverse := []bool{row.IndexInverse, false, false, false, row.CONInverse, false, row.WeaponInverse}
		y := 458 + i*26
		for j, col := range columns {
			r := image.Rect(col.lo, y, col.hi, y+26)
			fg := presentation.Text
			if inverse[j] {
				c.Box(r, presentation.Accent, presentation.Accent, 2)
				fg = presentation.Background
			}
			c.Label(values[j], r.Inset(2), s.modern.pixels, fg, j >= 2 && j <= 5)
		}
		s.modern.hits = append(s.modern.hits, modernHit{rect: image.Rect(32, y, 932, y+26), key: byte('1' + i)})
	}
}

func (s *Scene) modernToolbar(c *presentation.Canvas) {
	v := s.modern
	c.Fill(image.Rect(0, 0, 960, 44), presentation.Panel)
	c.Fill(image.Rect(0, 43, 960, 44), presentation.Edge)
	c.Box(image.Rect(12, 6, 204, 38), presentation.Panel, presentation.Edge, 5)
	c.Label(s.modernText("art", "Art"), image.Rect(24, 10, 58, 34), 16, presentation.Muted, false)
	for _, ctl := range presentation.Controls() {
		if ctl.ID == "original" || ctl.ID == "hd2-a" || ctl.ID == "hd2-b" {
			col := presentation.Muted
			if s.ArtMode() == ctl.ID {
				c.Box(ctl.Rect, presentation.RGB(0x30404c), presentation.Accent, 3)
				col = presentation.Accent
			}
			label := map[string]string{"original": s.modernText("original-short", "Orig"), "hd2-a": "A", "hd2-b": "B"}[ctl.ID]
			c.Label(label, ctl.Rect, 16, col, true)
		}
	}
	for _, r := range []image.Rectangle{image.Rect(212, 6, 316, 38), image.Rect(324, 6, 436, 38), image.Rect(444, 6, 564, 38), image.Rect(572, 6, 684, 38), image.Rect(692, 6, 772, 38), image.Rect(780, 6, 860, 38), image.Rect(868, 6, 948, 38)} {
		c.Box(r, presentation.Panel, presentation.Edge, 5)
	}
	family := s.modernText("rounded", "Round")
	if v.family == 1 {
		family = "Noto"
	}
	c.Label(s.modernText("font", "Font")+" "+family, image.Rect(218, 10, 310, 34), 16, presentation.Text, true)
	c.Label("−", image.Rect(330, 10, 354, 34), 16, presentation.Accent, true)
	c.Label(fmt.Sprintf("%d px", v.pixels), image.Rect(354, 10, 406, 34), 16, presentation.Text, true)
	c.Label("+", image.Rect(406, 10, 430, 34), 16, presentation.Accent, true)
	score := s.modernText("retro", "Retro")
	if s.settings.MusicVariant == "modern" {
		score = s.modernText("modern", "Modern")
	}
	c.Label(s.modernText("score", "Score")+" "+score, image.Rect(450, 10, 558, 34), 16, presentation.Text, true)
	vol := presentation.Text
	if !s.settings.MusicOn {
		vol = presentation.Muted
	}
	c.Label(s.modernText("volume-title", "Vol"), image.Rect(578, 10, 610, 34), 16, vol, true)
	c.Label("−", image.Rect(610, 10, 632, 34), 16, presentation.Accent, true)
	c.Label(fmt.Sprint(s.settings.MusicVol), image.Rect(632, 10, 656, 34), 16, vol, true)
	c.Label("+", image.Rect(656, 10, 678, 34), 16, presentation.Accent, true)
	c.Label(s.modernText("save", "Save F5"), image.Rect(698, 10, 766, 34), 15, presentation.Text, true)
	c.Label(s.modernText("load", "Load F9"), image.Rect(786, 10, 854, 34), 15, presentation.Text, true)
	c.Label(s.modernText("settings", "Setup F2"), image.Rect(874, 10, 942, 34), 15, presentation.Text, true)
}

func (s *Scene) modernMeter(c *presentation.Canvas) {
	c.Box(image.Rect(904, 80, 948, 472), presentation.Panel, presentation.Edge, 5)
	c.Label(s.modernText("radiation", "RAD"), image.Rect(908, 96, 944, 120), 12, presentation.Muted, true)
	reading := s.geigerReading()
	active := 0
	value := "—"
	if s.partyHasGeiger() && reading != render.MeterNoReading {
		active = max(0, 12-int(reading))
		value = fmt.Sprint(reading)
	}
	for i := 0; i < 12; i++ {
		col := presentation.RGB(0x293944)
		if 11-i < active {
			col = presentation.RGB(0x5bada4)
		}
		c.Box(image.Rect(917, 144+i*20, 935, 156+i*20), col, presentation.RGB(0x3a4c58), 2)
	}
	c.Label(value, image.Rect(908, 416, 944, 452), 18, presentation.Muted, true)
}

func (s *Scene) modernCommands(c *presentation.Canvas) {
	keys := "UEODVSRP"
	names := []string{"use", "encounter", "order", "split", "switch", "save-command", "radio", "journal"}
	fallback := []string{"Use", "Fight", "Order", "Split", "Group", "Save", "Radio", "Journal"}
	for i, key := range []byte(keys) {
		x := 16 + i*117
		r := image.Rect(x, 666, x+109, 704)
		c.Box(r, presentation.Panel, presentation.Edge, 5)
		col := presentation.Text
		if !s.modernFreeMap() {
			col = presentation.Muted
		}
		c.Label(string(key), image.Rect(x+10, 672, x+30, 698), 15, presentation.Accent, true)
		c.Label(s.modernText(names[i], fallback[i]), image.Rect(x+38, 672, x+99, 698), 16, col, true)
	}
}

func cellRows(cells map[[2]int]render.TextCell) []int {
	set := map[int]bool{}
	for _, c := range cells {
		set[c.Row] = true
	}
	rows := make([]int, 0, len(set))
	for r := range set {
		rows = append(rows, r)
	}
	sort.Ints(rows)
	return rows
}

func (s *Scene) modernMessageCells() map[[2]int]render.TextCell {
	h := render.NewHiFrame()
	h.CaptureText = true
	rect := s.msgRect()
	if s.message != "" {
		out, err := textlayout.Layout([]byte(s.message), textlayout.Options{Width: rect.Width})
		if err == nil {
			lines := out.Lines
			if s.cjkVisible() && len(lines) > 1 {
				lines = lines[:1]
			}
			for i, l := range lines {
				if rect.Row+i > rect.LastRow() {
					break
				}
				s.drawASCIILine(h, l.String(), rect.Col, rect.Row+i)
			}
		}
	}
	if s.journalOpen && s.journalHead != "" {
		s.drawCJKLine(h, s.journalHead, rect.Col, rect.Row)
	}
	if s.cjk != "" {
		s.walkBody(s.cjk, rect, func(col, row int, r rune) { h.DrawRune(nil, r, col, row, 15) })
	}
	return h.TextCells
}

func (s *Scene) modernRows(c *presentation.Canvas, cells map[[2]int]render.TextCell, first, last, minCol, maxCol int, safe image.Rectangle, lineHeight int, roster bool) {
	v := s.modern
	for row := first; row <= last; row++ {
		var line []render.TextCell
		for _, cell := range cells {
			if cell.Row == row && cell.Col >= minCol && cell.Col <= maxCol {
				line = append(line, cell)
			}
		}
		sort.Slice(line, func(i, j int) bool { return line[i].Col < line[j].Col })
		y := safe.Min.Y + (row-first)*lineHeight
		if y+lineHeight > safe.Max.Y {
			break
		}
		for at := 0; at < len(line); {
			end := at + 1
			for end < len(line) && line[end].Col-line[end-1].Col <= 2 && line[end].Inverse == line[at].Inverse {
				end++
			}
			var b strings.Builder
			col := line[at].Col
			for _, cell := range line[at:end] {
				for col < cell.Col {
					b.WriteByte(' ')
					col++
				}
				r := cell.Rune
				if r == 0x7f {
					r = '†'
				}
				if r < 0x20 {
					r = ' '
				}
				b.WriteRune(r)
				col++
			}
			step := v.pixels
			if roster {
				step = 22
			}
			x := safe.Min.X + (line[at].Col-minCol)*step
			r := image.Rect(x, y, safe.Max.X, y+lineHeight)
			if end < len(line) {
				r.Max.X = min(r.Max.X, safe.Min.X+(line[end].Col-minCol)*step-2)
			}
			fg := presentation.Text
			if line[at].Inverse {
				fg = presentation.Background
				width := min(r.Dx(), c.Measure(b.String(), v.pixels)+4)
				c.Box(image.Rect(x-2, y, x+width, y+lineHeight), presentation.Accent, presentation.Accent, 2)
			}
			c.Label(b.String(), r, v.pixels, fg, false)
			// Hit rectangles use the exact same measured prefix advances as the label.
			prefix := ""
			previous := line[at].Col
			for _, cell := range line[at:end] {
				for previous < cell.Col {
					prefix += " "
					previous++
				}
				left := x + c.Measure(prefix, v.pixels)
				prefix += string(cell.Rune)
				previous++
				right := x + c.Measure(prefix, v.pixels)
				hit := image.Rect(left, y, max(left+4, right), y+lineHeight).Intersect(r)
				if !hit.Empty() {
					v.hits = append(v.hits, modernHit{rect: hit, col: cell.Col, row: cell.Row})
				}
			}
			at = end
		}
	}
}

func (s *Scene) modernInput(in input.Input) (input.Input, bool) {
	v := s.modern
	if in.Scroll != input.ScrollNone && s.modernFreeMap() && !s.mapRosterShown() {
		if in.Scroll == input.ScrollUp {
			v.scroll = min(3, v.scroll+1)
		} else {
			v.scroll = max(0, v.scroll-1)
		}
		s.dirty = true
		return in, true
	}
	if !in.Mouse.Any() {
		return in, false
	}
	if in.Mouse.Right {
		return input.Input{Dir: input.DirNone, Action: input.ActionCancel}, false
	}
	p := image.Pt(in.Mouse.X, in.Mouse.Y)
	none := input.Input{Dir: input.DirNone}
	if _, err := s.ModernFrame(); err != nil {
		return none, true
	}
	for _, ctl := range presentation.Controls() {
		if !p.In(ctl.Rect) {
			continue
		}
		if s.quitAsk {
			return none, true
		}
		switch ctl.ID {
		case "original", "hd2-a", "hd2-b":
			if err := s.SelectArtMode(s.artRoot, ctl.ID); err != nil {
				s.settingsOpen = true
				s.message = "HD2 art pack unavailable. Keep current art mode. ESC close"
				s.cjk = s.uiText("settings.visualunavailable")
				if s.cjk != "" {
					s.message = ""
				}
			}
		case "font":
			v.family = 1 - v.family
		case "size-minus":
			v.pixels = max(16, v.pixels-2)
		case "size-plus":
			v.pixels = min(20, v.pixels+2)
		case "score":
			if s.settings.MusicVariant == "modern" {
				s.settings.MusicVariant = "retro"
			} else {
				s.settings.MusicVariant = "modern"
			}
		case "volume-minus":
			s.settings.MusicVol = max(0, s.settings.MusicVol-1)
		case "volume-plus":
			s.settings.MusicVol = min(10, s.settings.MusicVol+1)
		case "music", "music-value":
			s.settings.MusicOn = !s.settings.MusicOn
		case "save":
			none.Fn = input.FnQuickSave
			return none, false
		case "load":
			none.Fn = input.FnQuickLoad
			return none, false
		case "settings":
			none.Fn = input.FnSettings
			return none, false
		}
		if s.settingsOpen && ctl.ID != "original" && ctl.ID != "hd2-a" && ctl.ID != "hd2-b" {
			s.showSettings()
		}
		s.dirty = true
		return none, true
	}
	if p.Y < 80 {
		return none, true
	}
	for i := len(v.hits) - 1; i >= 0; i-- {
		h := v.hits[i]
		if !p.In(h.rect) {
			continue
		}
		if h.key != 0 {
			if h.key == 0x1b {
				none.Action = input.ActionCancel
			} else {
				none.Char = h.key
			}
			return none, false
		}
		none.Mouse = input.Mouse{X: h.col*24 + 12, Y: h.row*24 + 12, Left: true}
		return none, false
	}
	if s.modernPopup() {
		return none, true
	}
	if s.modernFreeMap() {
		for i, key := range []byte("UEODVSRP") {
			if p.In(image.Rect(16+i*117, 666, 125+i*117, 704)) {
				none.Char = key
				return none, false
			}
		}
		if p.In(image.Rect(32, 620, 620, 644)) {
			v.scroll = (v.scroll + 1) % 4
			s.dirty = true
			return none, true
		}
		if p.In(presentation.MapRect) && !s.mapRosterShown() {
			none.Mouse = input.Mouse{X: p.X, Y: p.Y - 60, Left: true}
			return none, false
		}
	}
	return none, true
}
