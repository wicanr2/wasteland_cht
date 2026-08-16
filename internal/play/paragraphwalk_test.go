package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestWalkOntoParagraphShowsText：**真的走上去**那一格，段落正文要出現。
//
// `TestParagraphTriggersHaveText` 驗的是資料（每個編號都查得到中文），
// 走上去那條路是另一回事——訊息的字串編號怎麼變成 key、key 怎麼查引用表、
// 查到之後正文有沒有真的送進畫面，中間任何一段錯了都只會顯示英文那句
// 「Read paragraph 23.」，看起來像「這段還沒翻」。
//
// 做法：逐張地圖、逐格從隔壁走上去，看 `journalAt` 有沒有被設起來
// （`maybeParagraph` 命中才會設）。**不挑固定座標**——段落引用大多掛在
// nibble 1 的敘述串列上，而那一種的記錄邊界還沒解，靜態掃會掃到雜訊。
func TestWalkOntoParagraphShowsText(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadJournal(
		"../../docs/re/generated/paragraph-refs.tsv",
		"../../translations/paragraphs-zh-Hant.cat"); err != nil {
		t.Fatalf("載手札：%v", err)
	}
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}

	const stepBudget = 20000 // 找到就停；上界只是不讓它跑到天荒地老
	steps := 0
	dirs := []struct {
		dx, dy int
		dir    input.Direction
	}{
		{0, 1, input.DirUp}, {0, -1, input.DirDown},
		{1, 0, input.DirLeft}, {-1, 0, input.DirRight},
	}

	for _, res := range resources {
		if err := s.LoadMap(res.ID, 1, 1); err != nil {
			continue
		}
		w := s.World()
		for y := 0; y < w.Block.Dim; y++ {
			for x := 0; x < w.Block.Dim; x++ {
				terrain, _, _, err := w.Block.At(x, y)
				// 敘述類的三種：1 ＝ 串列、4 ＝ 印一句、9 ＝ 印一句（輻射）。
				if err != nil || (terrain != 1 && terrain != 4 && terrain != 9) {
					continue
				}
				for _, d := range dirs {
					nx, ny := x+d.dx, y+d.dy
					if nx < 0 || ny < 0 || nx >= w.Block.Dim || ny >= w.Block.Dim {
						continue
					}
					if !w.Passable(nx, ny) {
						continue
					}
					if steps++; steps > stepBudget {
						t.Fatalf("走了 %d 步都沒有踩到引用段落的格子", stepBudget)
					}
					w.Teleport(uint8(nx), uint8(ny))
					s.journalAt = 0
					if _, err := s.Update(input.Input{Dir: d.dir}); err != nil {
						continue
					}
					if s.journalAt == 0 {
						continue
					}
					// 命中：正文要真的在畫面那條路徑上。
					n := s.journalAt
					if len(s.journal.Text(n)) == 0 {
						t.Fatalf("資源 %d (%d,%d) 引用第 %d 段，卻查不到正文",
							res.ID, x, y, n)
					}
					if len(s.CJK()) == 0 {
						t.Fatalf("資源 %d (%d,%d) 引用第 %d 段，畫面上卻沒有正文；"+
							"訊息是 %q", res.ID, x, y, n, s.Message())
					}
					t.Logf("資源 %d (%d,%d) → 第 %d 段，正文 %d bytes（走了 %d 步找到）",
						res.ID, x, y, n, len(s.CJK()), steps)
					return
				}
			}
		}
	}
	t.Fatalf("42 張地圖走了 %d 步，一格引用段落的都沒踩到", steps)
}
