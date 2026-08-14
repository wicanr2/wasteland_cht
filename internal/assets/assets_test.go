package assets

import (
	"bytes"
	"os"
	"testing"
)

// 這些測試吃玩家自備的原版資料（不入版控）。沒有就整包跳過，
// 但**跳過會明講**——沉默的跳過與通過長得一模一樣。
const (
	romDir    = "../../workplace/orig/wastland"
	imagePath = "../../workplace/analysis/unpacked/wl.merged.exe"
)

func openRom(t *testing.T) *Rom {
	t.Helper()
	if _, err := os.Stat(romDir); err != nil {
		t.Skipf("找不到原版資料目錄 %s，跳過（玩家自備）", romDir)
	}
	rom, err := Open(romDir)
	if err != nil {
		t.Fatalf("Open：%v", err)
	}
	return rom
}

func openWithImage(t *testing.T) *Rom {
	t.Helper()
	rom := openRom(t)
	if _, err := os.Stat(imagePath); err != nil {
		t.Skipf("找不到分析映像 %s，跳過（tools/unpack_exepack.py 的產物）", imagePath)
	}
	if err := rom.LoadImage(imagePath); err != nil {
		t.Fatalf("LoadImage：%v", err)
	}
	return rom
}

func TestOpenVerifiesHashes(t *testing.T) {
	rom := openRom(t)
	if len(rom.files) != len(KnownFiles) {
		t.Fatalf("載入 %d 個檔案，應該是 %d 個", len(rom.files), len(KnownFiles))
	}
}

func TestOpenRejectsTamperedFile(t *testing.T) {
	rom := openRom(t)
	dir := t.TempDir()
	for name := range KnownFiles {
		data, err := rom.File(name)
		if err != nil {
			t.Fatal(err)
		}
		out := append([]byte(nil), data...)
		if name == "info" {
			out[0] ^= 0xFF // 動一個 byte
		}
		if err := os.WriteFile(dir+"/"+name, out, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Open(dir); err == nil {
		t.Fatal("改過一個 byte 還是通過了——雜湊驗證沒有生效")
	}
}

// 42 個 MSQ 區塊全部解得開，而且三層的長度都對得上（docs/spec/01 §4）。
func TestAllBlocksDecode(t *testing.T) {
	rom := openWithImage(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatalf("Resources：%v", err)
	}
	if len(res) != 42 {
		t.Fatalf("解出 %d 個 MSQ 資源，應該是 42 個", len(res))
	}

	dims := map[int]int{}
	tilesets := map[int]bool{}
	strings := 0
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatalf("區塊 %d：%v", i, err)
		}
		if len(b.Terrain) != b.Dim*b.Dim || len(b.Record) != b.Dim*b.Dim || len(b.Graphic) != b.Dim*b.Dim {
			t.Fatalf("區塊 %d 的三層長度不一致", i)
		}
		dims[b.Dim]++
		tilesets[b.Tileset] = true
		strings += countNonEmpty(b.Strings)
	}
	if dims[32] != 38 || dims[64] != 4 {
		t.Fatalf("邊長分佈是 %v，應該是 32 × 38、64 × 4", dims)
	}
	if len(tilesets) != 9 {
		t.Fatalf("用到 %d 個圖磚組，應該是 9 個", len(tilesets))
	}
	// 字串保留空槽（編號就是索引），所以比對的是非空的條數。
	if strings != 4401 {
		t.Fatalf("區塊的非空字串共 %d 條，應該是 4,401 條", strings)
	}
}

// 未解區域要能原樣寫回：Raw 的加密段重新加密後，必須與原始檔 byte-for-byte 相同。
func TestBlockRawRoundTrip(t *testing.T) {
	rom := openWithImage(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatalf("區塊 %d：%v", i, err)
		}
		data, err := rom.File(r.File)
		if err != nil {
			t.Fatal(err)
		}
		span := data[r.Offset : r.Offset+r.TotalLen]
		checksum := le16(span, 4)

		// 重新加密：同一把 key、同一個長度
		enc := append([]byte(nil), b.Raw...)
		key := byte(checksum&0xFF) ^ byte(checksum>>8)
		for j := 0; j < b.EncLen && j < len(enc); j++ {
			enc[j] = b.Raw[j] ^ key
			key += 0x1F
		}
		if !bytes.Equal(enc, span[6:r.ReadLen]) {
			t.Fatalf("區塊 %d 的加密段 round-trip 不一致", i)
		}
	}
}

func TestExeStrings(t *testing.T) {
	rom := openWithImage(t)
	tables, err := rom.ExeStrings()
	if err != nil {
		t.Fatalf("ExeStrings：%v", err)
	}
	if len(tables) != 9 {
		t.Fatalf("解出 %d 張表，應該是 9 張", len(tables))
	}
	slots, nonEmpty := 0, 0
	for _, tb := range tables {
		slots += len(tb)
		nonEmpty += countNonEmpty(tb)
	}
	// 一組固定四個槽，最後一組不一定用滿——所以槽數與非空條數要分開看。
	if slots != 444 || nonEmpty != 426 {
		t.Fatalf("執行檔字串是 %d 個槽、%d 條非空，應該是 444／426", slots, nonEmpty)
	}
	// 第一張表的第一條是開場字幕，用它確認解碼方向沒錯。
	if len(tables[0]) < 2 || !bytes.Contains([]byte(tables[0][1]), []byte("Electronic Arts")) {
		t.Fatalf("開場字幕解出來的是 %q", tables[0])
	}
}

func TestPicturesAndTiles(t *testing.T) {
	rom := openRom(t)

	title, err := rom.Title()
	if err != nil {
		t.Fatalf("Title：%v", err)
	}
	if title.Width != 288 || title.Height != 128 {
		t.Fatalf("標題畫面是 %d × %d", title.Width, title.Height)
	}
	if used := colorsUsed(title); used != 16 {
		t.Fatalf("標題畫面用到 %d 種顏色，應該是 16 種", used)
	}

	pics := 0
	for _, name := range []string{"allpics1", "allpics2"} {
		list, err := rom.Pictures(name)
		if err != nil {
			t.Fatalf("%s：%v", name, err)
		}
		for _, p := range list {
			if p.Width != 96 || p.Height != 84 {
				t.Fatalf("%s 有一張是 %d × %d", name, p.Width, p.Height)
			}
		}
		pics += len(list)
	}
	if pics != 82 {
		t.Fatalf("解出 %d 張圖，應該是 82 張", pics)
	}

	want := []int{66, 141, 163, 107, 127, 118, 90, 104, 135}
	for n, count := range want {
		tiles, err := rom.Tileset(n)
		if err != nil {
			t.Fatalf("圖磚組 %d：%v", n, err)
		}
		if len(tiles) != count {
			t.Fatalf("圖磚組 %d 有 %d 張，應該是 %d 張", n, len(tiles), count)
		}
		if tiles[0].Width != 16 || tiles[0].Height != 16 {
			t.Fatalf("圖磚組 %d 的圖磚是 %d × %d", n, tiles[0].Width, tiles[0].Height)
		}
	}
}

// 地圖第 3 層的圖形編號：0–9 是 IC0_9.WLF，≥10 是圖磚（編號 − 10）。
// 全 42 張地圖都不得指到範圍外。
func TestGraphicIndexInRange(t *testing.T) {
	rom := openWithImage(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	icons, err := rom.Icons()
	if err != nil {
		t.Fatalf("Icons：%v", err)
	}
	if len(icons) != 10 {
		t.Fatalf("IC0_9.WLF 解出 %d 張，應該是 10 張", len(icons))
	}
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatal(err)
		}
		tiles, err := rom.Tileset(b.Tileset)
		if err != nil {
			t.Fatalf("區塊 %d 的圖磚組：%v", i, err)
		}
		limit := len(icons) + len(tiles)
		for _, g := range b.Graphic {
			if int(g) >= limit {
				t.Fatalf("區塊 %d 的圖形編號 %d 超出 %d（%d 疊圖 ＋ %d 圖磚）",
					i, g, limit, len(icons), len(tiles))
			}
		}
		if int(b.OutsideGraphic()) >= limit {
			t.Fatalf("區塊 %d 的地圖外圖形 %d 超出範圍", i, b.OutsideGraphic())
		}
	}
}

func TestFonts(t *testing.T) {
	rom := openWithImage(t)

	mono, err := rom.FontMain()
	if err != nil {
		t.Fatalf("FontMain：%v", err)
	}
	if len(mono.Glyphs) != 128 {
		t.Fatalf("主文字字型有 %d 個字模，應該是 128 個", len(mono.Glyphs))
	}
	// 空白（ASCII 0x20）必須整格空白，A 必須有筆畫——方向錯了這兩項會同時壞。
	space, err := mono.GlyphForASCII(' ')
	if err != nil {
		t.Fatal(err)
	}
	if lit(space) != 0 {
		t.Fatalf("空白字模有 %d 個亮點", lit(space))
	}
	letterA, err := mono.GlyphForASCII('A')
	if err != nil {
		t.Fatal(err)
	}
	if n := lit(letterA); n < 8 || n > 40 {
		t.Fatalf("字母 A 的亮點數是 %d，不像字", n)
	}

	color, err := rom.FontColor()
	if err != nil {
		t.Fatalf("FontColor：%v", err)
	}
	if len(color.Glyphs) != 172 {
		t.Fatalf("彩色字型有 %d 個字模，應該是 172 個", len(color.Glyphs))
	}
}

// 時間欄位：荒野大地圖每步 4 分鐘、一般室內 15 秒（docs/re/27 §3）。
func TestStepTime(t *testing.T) {
	rom := openWithImage(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	counts := map[float64]int{}
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatal(err)
		}
		counts[b.StepMinutes()]++
	}
	if counts[0.25] != 28 || counts[4] != 1 {
		t.Fatalf("每步分鐘的分佈是 %v，應該有 28 張 0.25、1 張 4", counts)
	}
}

func countNonEmpty(list []string) int {
	n := 0
	for _, s := range list {
		if s != "" {
			n++
		}
	}
	return n
}

func colorsUsed(im *Indexed) int {
	var seen [16]bool
	for _, p := range im.Pix {
		seen[p&0x0F] = true
	}
	n := 0
	for _, s := range seen {
		if s {
			n++
		}
	}
	return n
}

func lit(g *Glyph) int {
	n := 0
	for _, p := range g.Pix {
		if p != 0 {
			n++
		}
	}
	return n
}

// 物品資料表在存檔區，三個槽各一份，checksum 都要過（docs/re/45 §2）。
func TestLoadItemTableAllSlots(t *testing.T) {
	r := openRom(t)
	for slot := range ItemSlotOffsets {
		raw, err := r.LoadItemTable("game1", slot)
		if err != nil {
			t.Fatalf("槽 %d：%v", slot, err)
		}
		if len(raw) != itemTableLen {
			t.Fatalf("槽 %d 長度 %d，應該是 %d", slot, len(raw), itemTableLen)
		}
	}
	// 槽超出範圍要回錯誤，不是靜靜回空表。
	if _, err := r.LoadItemTable("game1", len(ItemSlotOffsets)); err == nil {
		t.Fatal("槽超出範圍應該回錯誤")
	}
}

// 出廠資料的幾筆已知值——這一組同時擋住「類別右移四次」與索引偏移錯位。
func TestItemTableKnownEntries(t *testing.T) {
	raw, err := openRom(t).LoadItemTable("game1", 0)
	if err != nil {
		t.Fatalf("讀物品表：%v", err)
	}
	// 索引 n 落在表的第 n+1 筆（基址 ＝ 表首 ＋ 8）。
	entry := func(id int) []byte { return raw[(id+1)*8 : (id+2)*8] }

	// 12 ＝ RPG-7：類別 9（反戰車重）、13 顆骰、彈匣 1、技能 11（AT weapon）。
	rpg := entry(12)
	if got := rpg[3] >> 3; got != 9 {
		t.Fatalf("RPG-7 的類別是 %d，應該是 9（右移三次不是四次）", got)
	}
	if rpg[6] != 13 || rpg[5] != 11 {
		t.Fatalf("RPG-7 骰數 %d、技能 %d，應該是 13／11", rpg[6], rpg[5])
	}
	// 13 ＝ M1911A1 45 pistol：彈匣 7、彈藥指向 30（45 clip）。
	pistol := entry(13)
	if pistol[4] != 7 || pistol[7] != 30 {
		t.Fatalf(".45 手槍容量 %d、彈藥 %d，應該是 7／30", pistol[4], pistol[7])
	}
	// 37 ＝ Kevlar vest：類別 15（護甲）、AC 4。
	vest := entry(37)
	if got := vest[3] >> 3; got != 15 || vest[6] != 4 {
		t.Fatalf("Kevlar vest 類別 %d、AC %d，應該是 15／4", got, vest[6])
	}
}

// SetCell 改寫兩層，並且擋住放不進 4 bits 的地形值——
// 原版是直接 or 進去的，遮掉的話資料異常永遠不會被發現（docs/re/46 §4.1）。
func TestSetCell(t *testing.T) {
	r := openWithImage(t)
	b, err := r.Block(0)
	if err != nil {
		t.Fatalf("載入區塊 0 失敗：%v", err)
	}
	x, y := 10, 10
	oldT, oldR, _, err := b.At(x, y)
	if err != nil {
		t.Fatalf("讀格子失敗：%v", err)
	}
	if err := b.SetCell(x, y, 3, 0x21); err != nil {
		t.Fatalf("改寫失敗：%v", err)
	}
	gotT, gotR, _, _ := b.At(x, y)
	if gotT != 3 || gotR != 0x21 {
		t.Fatalf("改寫後是 (%d, %#x)，應該是 (3, 0x21)", gotT, gotR)
	}
	if err := b.SetCell(x, y, 0x1F, 0); err == nil {
		t.Fatal("地形是 4 bits，0x1F 應該被擋下來")
	}
	if err := b.SetCell(-1, 0, 0, 0); err == nil {
		t.Fatal("越界應該回錯誤")
	}
	// 放回去，免得影響同一個 Block 的其他測試。
	if err := b.SetCell(x, y, oldT, oldR); err != nil {
		t.Fatalf("還原失敗：%v", err)
	}
}

// 存檔尾段那 0xA00 就是十組按鍵巨集（docs/re/30 §6、docs/re/43 §6）。
func TestSaveTailIsMacros(t *testing.T) {
	sv, err := openRom(t).LoadSave("game1")
	if err != nil {
		t.Fatalf("讀存檔：%v", err)
	}
	if len(sv.Tail) != MacroCount*MacroStride {
		t.Fatalf("尾段 %d bytes，應該是 %d ＝ 10 × 256", len(sv.Tail), MacroCount*MacroStride)
	}
	for n := 0; n < MacroCount; n++ {
		m, ok := sv.Macro(n)
		if !ok || len(m) != MacroStride {
			t.Fatalf("第 %d 組取不到：ok=%v len=%d", n, ok, len(m))
		}
		// 出廠全零——還沒有人錄過。
		for i, b := range m {
			if b != 0 {
				t.Fatalf("第 %d 組的第 %d 個 byte 是 %#x，出廠應該全零", n, i, b)
			}
		}
	}
	if _, ok := sv.Macro(MacroCount); ok {
		t.Fatal("超出範圍應該回 false")
	}
	// 巨集不進 checksum：改了尾段照樣編得回去。
	sv.Tail[0] = 'A'
	if got := sv.Bytes(); len(got) != 6+savePlainLen+saveTailLen {
		t.Fatalf("重新編碼的長度 %d 不對", len(got))
	}
}
