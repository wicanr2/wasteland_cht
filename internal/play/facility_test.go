package play

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// hasNibble 回報這張地圖有沒有某一種第 1 層 nibble 的格子。
func hasNibble(b *assets.Block, nibble byte) bool {
	for _, t := range b.Terrain {
		if t == nibble {
			return true
		}
	}
	return false
}

// 設施場景（docs/spec/23 §6 的驗收條件）。

// 驗收 3：bit7 沒設的記錄是腳本指令，不是第 0 種設施。
func TestScriptRecordIsNotFacility(t *testing.T) {
	s := &Scene{}
	if fs := s.EnterFacility([]byte{0x00, 1, 2, 3, 4, 5, 6, 7}); fs != nil {
		t.Errorf("bit7 沒設卻當成設施 %v", fs.Facility.Kind)
	}
	if s.InFacility() {
		t.Error("沒進設施，InFacility 卻是 true")
	}
}

// 驗收 2：圖片編號與 docs/re/29 §5.4 那張表一致。
func TestFacilityPictureNumbers(t *testing.T) {
	want := map[game.FacilityKind]int{
		game.FacilityDoctor: 0, game.FacilityShop: 1,
		game.FacilityTrainer: 2, game.FacilityRoster: 3,
		game.FacilityUnknown: -1, // 這一種原版沒有載圖，不要猜一個
	}
	for kind, pic := range want {
		rec := make([]byte, 32)
		rec[0] = 0x80 | byte(kind)
		s := &Scene{}
		fs := s.EnterFacility(rec)
		if fs == nil {
			t.Fatalf("設施 %d 解不開", kind)
		}
		if fs.Picture != pic {
			t.Errorf("設施 %d 的圖應該是 %d，得到 %d", kind, pic, fs.Picture)
		}
	}
}

// 驗收 5：42 張地圖裡所有 bit7 設起來的 nibble 6 記錄都解得開，
// 而且設施編號不超出 0–4。
func TestAllFacilityRecordsParse(t *testing.T) {
	rom := openRomForPlay(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatalf("列資源：%v", err)
	}
	seen := map[game.FacilityKind]int{}
	names := 0
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			continue
		}
		// ⚠ **要掃整張 section 6 陣列，不能只掃「目前有格子指到」的記錄。**
		// 只掃被格子指到的會找到 2 筆（訓練師與存檔），看起來像「這遊戲
		// 沒有商店也沒有醫生」；掃整張是 7 筆，四種都在。差別在於設施格
		// 多半是**答對密語或腳本改寫之後**才指過去的（docs/re/46 §4.1）。
		if !hasNibble(b, 6) {
			continue // 沒有 nibble 6 的區塊，section 6 指標多半指到別人
		}
		for n := 0; n < b.SectionCount(6); n++ {
			rec, err := b.SectionRecord(6, n)
			if err != nil || len(rec) == 0 || rec[0]&0x80 == 0 {
				continue
			}
			f, ok := game.ParseFacility(rec)
			if !ok {
				t.Errorf("區塊 %d：bit7 設起來卻解不開（%#02x）", i, rec[0])
				continue
			}
			seen[f.Kind]++
			// 驗收 1：名字要是可讀的（NUL 以外不得有控制碼）。
			for _, ch := range f.Name {
				if ch < 0x20 || ch == 0x7F {
					t.Errorf("區塊 %d 的設施 %d 名字含控制碼 %#02x：%q",
						i, f.Kind, ch, f.Name)
					break
				}
			}
			if f.Name != "" {
				names++
			}
		}
	}
	// 四種有實作的設施都要找得到。少了商店或醫生就是掃描漏了
	// （多半又只掃了「目前有格子指到」的記錄）。
	for _, k := range []game.FacilityKind{
		game.FacilityDoctor, game.FacilityShop,
		game.FacilityTrainer, game.FacilityRoster,
	} {
		if seen[k] == 0 {
			t.Errorf("42 張地圖裡找不到設施種類 %d——掃描漏了", k)
		}
	}
	t.Logf("設施分布 %v，其中 %d 筆有地點名", seen, names)
}

// 驗收 4：離開設施不動世界狀態。
func TestLeaveFacilityKeepsWorld(t *testing.T) {
	s := openScene(t)
	w := s.World()
	x, y, clock := w.Party.X, w.Party.Y, w.Clock

	rec := make([]byte, 32)
	rec[0] = 0x80 | byte(game.FacilityShop)
	copy(rec[0x07:], "Shop\x00")
	if s.EnterFacility(rec) == nil {
		t.Fatal("進不了設施")
	}
	if !s.InFacility() {
		t.Fatal("InFacility 應該是 true")
	}
	s.LeaveFacility()
	if s.InFacility() {
		t.Error("離開之後 InFacility 還是 true")
	}
	if w.Party.X != x || w.Party.Y != y || w.Clock != clock {
		t.Error("離開設施不該動到座標或時鐘")
	}
}

// 驗收：設施圖畫在視窗原點 (8, 8)，地點名在字元列 12（docs/re/54）。
func TestFacilityPictureGoesToWindowOrigin(t *testing.T) {
	s := openScene(t)
	if len(s.pics) <= 3 {
		t.Skip("ALLPICS1 沒載到")
	}
	rec := make([]byte, 32)
	rec[0] = 0x80 | byte(game.FacilityRoster)
	copy(rec[0x03:], "Ranger Ctr.\x00")
	if s.EnterFacility(rec) == nil {
		t.Fatal("進不了設施")
	}
	s.dirty = true
	f := s.Frame()

	// 圖的每一個像素都要落在 (8, 8) 起的那一塊。
	want := s.pics[3]
	diff := 0
	for y := 0; y < want.Height; y++ {
		for x := 0; x < want.Width; x++ {
			if f.At(render.FacilityPicX+x, render.FacilityPicY+y) != want.Pix[y*want.Width+x] {
				diff++
			}
		}
	}
	if diff != 0 {
		t.Errorf("設施圖沒有畫在 (8, 8)：%d 個像素不符", diff)
	}
	// 地點名那一列要有東西（字元列 12）。
	ink := 0
	for x := render.FacilityNameCol * render.CharWidth; x < 120; x++ {
		for y := render.FacilityNameRow * render.CharHeight; y < (render.FacilityNameRow+1)*render.CharHeight; y++ {
			if f.At(x, y) != 0 {
				ink++
			}
		}
	}
	if ink == 0 {
		t.Errorf("字元列 %d 應該印著地點名，卻是空的", render.FacilityNameRow)
	}
}

// 驗收（規格 26 §6 條 1／5）：進設施之後推 9 拍動畫，畫面與實機截圖
// **逐像素相同**，而且底圖那份解碼緩衝區一個 byte 都沒被動到。
//
// 這是整條路徑的對拍——EnterFacility → TickAnim → Frame，
// 不是只驗 PicPlayer 自己吐了什麼（`CLAUDE.md` §4：驗收要對原版）。
func TestFacilityAnimationMatchesHardwareShot(t *testing.T) {
	const shot = "../../workplace/dosbox/shots/db05.ppm"
	if _, err := os.Stat(shot); err != nil {
		t.Skipf("找不到實機截圖 %s，跳過。產生方式見 docs/re/54 §4", shot)
	}
	s := openScene(t)
	if len(s.pics) <= 3 || len(s.anims) <= 3 {
		t.Skip("ALLPICS1 沒載到")
	}
	want, err := readPPM(shot)
	if err != nil {
		t.Fatalf("讀截圖：%v", err)
	}
	before := append([]byte(nil), s.pics[3].Pix...)

	rec := make([]byte, 32)
	rec[0] = 0x80 | byte(game.FacilityRoster)
	copy(rec[0x03:], "Ranger Ctr.\x00")
	if s.EnterFacility(rec) == nil {
		t.Fatal("進不了設施")
	}
	// 腳本的延遲是 2,1,1,1,…，所以格 3 疊上去是第 9 次推進
	// （倒數 2 拍 → 格 0，之後每 2 拍一格）。
	for i := 0; i < 9; i++ {
		s.TickAnim()
	}
	f := s.Frame()

	diff := 0
	for y := 0; y < render.FacilityPicHeight; y++ {
		for x := 0; x < render.FacilityPicWidth; x++ {
			sx, sy := render.FacilityPicX+x, render.FacilityPicY+y
			o := (sy*render.ScreenWidth + sx) * 3
			if egaIndex(want[o], want[o+1], want[o+2]) != f.At(sx, sy) {
				diff++
			}
		}
	}
	if diff != 0 {
		t.Errorf("推 9 拍之後與實機截圖差 %d 個像素", diff)
	}
	// 規格 26 §6 條 5：動畫是疊在畫面上，不回頭改解碼緩衝區。
	for i, v := range s.pics[3].Pix {
		if v != before[i] {
			t.Fatalf("底圖被改掉了（第 %d 個像素 %d → %d）", i, before[i], v)
		}
	}
}

// egaIndex 把截圖的 RGB 換回索引；認不得回 0xFF，讓它一定不吻合。
func egaIndex(r, g, b byte) byte {
	for i, c := range assets.EGAPalette {
		if c.R == r && c.G == g && c.B == b {
			return byte(i)
		}
	}
	return 0xFF
}

// readPPM 讀 P6 的 320 × 200 截圖。
func readPPM(path string) ([]byte, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	r := bufio.NewReader(fh)
	var magic string
	var w, h, max int
	if _, err := fmt.Fscan(r, &magic, &w, &h, &max); err != nil {
		return nil, fmt.Errorf("讀標頭：%w", err)
	}
	if magic != "P6" || max != 255 {
		return nil, fmt.Errorf("只支援 8-bit 的 P6，這份是 %s／max %d", magic, max)
	}
	if w != render.ScreenWidth || h != render.ScreenHeight {
		return nil, fmt.Errorf("截圖是 %d × %d，畫面是 %d × %d",
			w, h, render.ScreenWidth, render.ScreenHeight)
	}
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}
	buf := make([]byte, w*h*3)
	for i := 0; i < len(buf); {
		n, err := r.Read(buf[i:])
		if err != nil {
			return nil, fmt.Errorf("讀像素（讀到第 %d byte）：%w", i, err)
		}
		i += n
	}
	return buf, nil
}
