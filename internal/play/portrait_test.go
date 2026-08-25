package play

// 肖像框：一張圖 ＋ 一行說明（`docs/re/115`）。

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/wicanr2/wasteland_cht/internal/render"
)

// 說明字串置中在 12 格裡（`0x126A5`–`0x126BA`）。
//
// ⚠ 置中算的是**格數**不是 byte 數：一個中文字兩個 byte、只佔一格。
// 拿 len 去算的話中文那一行會往左偏一半，而畫面上看起來只是「沒對準」。
func TestPortraitCaptionCentered(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int
	}{
		// 實機截圖 `24-encyes.png`：`Ranger` 六格，起點 ＝ 1 ＋ (6 − 3) ＝ 4。
		{"Ranger", render.FacilityNameCol + 3},
		{"變種人", render.FacilityNameCol + 5},
		{"Mutants", render.FacilityNameCol + 3},
		// 12 格以上不置中（原版 `cmp al, 0x0C` 之後直接用起點）。
		{"123456789012", render.FacilityNameCol},
		{"12345678901234", render.FacilityNameCol},
	} {
		if got := captionCol(tc.text); got != tc.want {
			t.Errorf("%q 的起點是 %d，預期 %d", tc.text, got, tc.want)
		}
	}
}

// 中文說明放不下 12 格時**先退成中文那一段**，不要硬截。
//
// ⚠ 硬截的症狀是「變種人（Muta」——看起來只是「這個名字比較長」。
func TestFitCaption(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"變種人", "變種人"},
		{"變種人（Mutant）", "變種人（Mutant）"},    // 11 格，剛好放得下
		{"掠奪者（Marauder）", "掠奪者"},          // 13 格 → 退成中文那一段
		{"一二三四五六七八九十一二", "一二三四五六七八九十一二"},  // 剛好 12 格
		{"一二三四五六七八九十一二三", "一二三四五六七八九十一二"}, // 沒有括號 → 截
	} {
		if got := fitCaption(tc.in); got != tc.want {
			t.Errorf("fitCaption(%q) ＝ %q，預期 %q", tc.in, got, tc.want)
		}
		if n := utf8.RuneCountInString(fitCaption(tc.in)); n > portraitCaptionCells {
			t.Errorf("%q 之後還是 %d 格", tc.in, n)
		}
	}
}

func TestFacilityNameUsesCaptionSafeRect(t *testing.T) {
	got := facilityDisplayLine(0, "遊俠中心（Ranger Center.）")
	if got != "遊俠中心" {
		t.Fatalf("設施招牌 = %q，預期保留框內中文名", got)
	}
	if n := utf8.RuneCountInString(got); n > portraitCaptionCells {
		t.Fatalf("設施招牌仍佔 %d 格，超過 %d", n, portraitCaptionCells)
	}
	if got := facilityDisplayLine(1, "C建立  D刪除  P開始遊戲"); got != "C建立  D刪除  P開始遊戲" {
		t.Fatalf("選單列不應被招牌規則修改：%q", got)
	}
}

// 沒有敵人的那一回合（空地 `ENC`）畫的是**遊俠**：圖 8、說明 `Ranger`。
//
// ⚠ 這一條擋的是「留白」。原版在一組敵人都挑不到時走的是後備分支
// （`0x12660` 印字串 96、`0x12665` 載圖 8），不是什麼都不畫——
// 而畫面上「什麼都不畫」看起來只是「這一場沒有敵人」。
func TestEmptyRoundShowsRanger(t *testing.T) {
	s := newScene(t)
	if _, err := s.beginEmptyRound(); err != nil {
		t.Fatalf("開不了空回合：%v", err)
	}
	if got := s.portraitPicture(); got != rangerPicture {
		t.Errorf("圖是 %d，預期 %d", got, rangerPicture)
	}
	en, _ := s.portraitCaption()
	if en != "Ranger" {
		t.Errorf("說明是 %q，預期 Ranger", en)
	}
	// 圖那一塊真的畫上去了。
	f := s.Frame()
	nonZero := 0
	for y := render.FacilityPicY; y < render.FacilityPicY+render.FacilityPicHeight; y++ {
		for x := render.FacilityPicX; x < render.FacilityPicX+render.FacilityPicWidth; x++ {
			if f.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Error("空回合沒有畫遊俠那張圖")
	}
	// 說明那一列也要有字。
	row := render.FacilityNameRow * render.CharHeight
	ink := 0
	for y := row; y < row+render.CharHeight; y++ {
		for x := 0; x < render.FacilityPicWidth; x++ {
			if f.At(x, y) != 0 {
				ink++
			}
		}
	}
	if ink == 0 {
		t.Error("說明那一列是空的")
	}
}

// 有敵人時圖與說明**走同一個判斷**：都是第一組還活著的那一隻。
//
// ⚠ 兩邊各記一份的話，敵人死光之後圖還是敵人、字已經是遊俠，
// 而畫面上看起來完全正常。
func TestPortraitFollowsTheSameEnemy(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(4, 18, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartEncounter(); err != nil {
		t.Fatal(err)
	}
	e, _ := s.combat.firstEnemy()
	if e == nil {
		t.Fatal("打起來了卻沒有敵人")
	}
	if got := s.portraitPicture(); got != int(e.Data.Portrait) {
		t.Errorf("圖是 %d，預期敵人資料的 +0x07 ＝ %d", got, e.Data.Portrait)
	}
	en, _ := s.portraitCaption()
	if en == "" || en == "Ranger" {
		t.Errorf("有敵人時說明不該是 %q", en)
	}
	if en != s.combat.enemyLabel(e) {
		t.Errorf("說明 %q 與訊息裡的稱呼 %q 不一致", en, s.combat.enemyLabel(e))
	}

	// 全部打死之後退回遊俠那一支。
	for i := 0; i < 100; i++ {
		if x := s.combat.Battle.Enemy(i % 40); x != nil {
			x.HP = 0
		}
	}
	if got := s.portraitPicture(); got != rangerPicture {
		t.Errorf("敵人死光之後圖是 %d，預期 %d", got, rangerPicture)
	}
	en, _ = s.portraitCaption()
	if !strings.Contains(en, "Ranger") {
		t.Errorf("敵人死光之後說明是 %q，預期 Ranger", en)
	}
}
