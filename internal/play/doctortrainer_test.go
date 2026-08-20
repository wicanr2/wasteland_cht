package play

// 醫生與訓練師的實機對拍（`docs/re/119`）。

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// enterFacilityAt 走一步踩上傳送格、答 Y，回傳當時的場景。
func enterFacilityAt(t *testing.T, id, x, y int) *Scene {
	t.Helper()
	s := newScene(t)
	if err := s.LoadMap(id, uint8(x), uint8(y)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Char: 'Y'}); err != nil {
		t.Fatal(err)
	}
	if s.facility == nil {
		t.Fatalf("沒進設施：%q", s.Message())
	}
	return s
}

// 招呼語的位移**每一種設施不一樣**：商店 `+0x05`、醫生與訓練師 `+0x03`。
//
// ⚠ 拿錯位移不會壞掉，只會印出**別的句子**：醫生印成「Entering the
// infirmary.」（那是走進門那一步的地圖訊息）、訓練師印成隔壁店的招呼語，
// 而畫面上看起來完全正常。
func TestGreetingOffsetPerFacilityKind(t *testing.T) {
	for _, c := range []struct {
		name    string
		id      int
		x, y    int
		wantSay string
	}{
		{"商店", 10, 30, 25, "Welcome to the shop."},
		{"醫生", 10, 13, 30, "Welcome to the infirmary."},
		{"訓練師", 21, 24, 21, "Darwin Branch Library."},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := enterFacilityAt(t, c.id, c.x, c.y)
			if got := s.facility.Greeting; got != c.wantSay {
				t.Errorf("招呼語是 %q，預期 %q", got, c.wantSay)
			}
		})
	}
}

// 醫生也要問「誰要治療？」，而且主選單**只列這個人現在用得到的項目**。
//
// ⚠ 三個都印的話玩家會以為隨時能治療；原版是三個旗標各自判斷
// （`0x1C2CE`–`0x1C2FF`）。
func TestDoctorAsksWhoAndListsOnlyWhatApplies(t *testing.T) {
	s := enterFacilityAt(t, 10, 13, 30)
	if got := s.facility.state.Step; got != StepWho {
		t.Fatalf("進場停在第 %d 層，預期 StepWho（%d）", got, StepWho)
	}
	if lines := strings.Join(s.facility.Lines, "\n"); !strings.Contains(lines, "Who wants treatment?") {
		t.Errorf("醫生要問「誰要治療？」：\n%s", lines)
	}
	if _, err := s.Update(input.Input{Char: '1'}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Join(s.facility.Lines, "\n")
	if !strings.Contains(lines, "Exam $") {
		t.Errorf("沒有 Exam：\n%s", lines)
	}
	if !strings.Contains(lines, "I would recommend") {
		t.Errorf("沒有「I would recommend」那一行：\n%s", lines)
	}
	m := s.World().Party.Members[0]
	if m.CON == m.MaxCON && strings.Contains(lines, "Healing") {
		t.Errorf("這個人是滿血的，不該列 Healing：\n%s", lines)
	}

	// 受傷之後 Healing 才出現。
	m.CON = m.MaxCON - 5
	s.facility.refresh(s.World().Party, s.items)
	if lines := strings.Join(s.facility.Lines, "\n"); !strings.Contains(lines, "Healing") {
		t.Errorf("受傷了卻沒有 Healing：\n%s", lines)
	}
}

// 訓練師的清單有表頭與三欄（IQ／PTS／LVL），而且**第一列是 Brawling**。
//
// ⚠ 技能表第 0 格不是技能（與物品表第 0 筆同一種佔位），列出來畫面上會多
// 一行沒有名字的 `skill 0`——實機截圖第一列就是 `Brawling`。
func TestTrainerListMatchesOriginal(t *testing.T) {
	s := enterFacilityAt(t, 21, 24, 21)
	if _, err := s.Update(input.Input{Char: '1'}); err != nil {
		t.Fatal(err)
	}
	lines := s.facility.Lines
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "IQ PTS LVL   SKILL") {
		t.Errorf("沒有表頭：\n%s", joined)
	}
	if strings.Contains(joined, "skill 0") {
		t.Errorf("列出了不是技能的第 0 格：\n%s", joined)
	}
	var first string
	for _, l := range lines {
		if strings.HasPrefix(l, "1>") {
			first = l
			break
		}
	}
	if first == "" {
		t.Fatalf("找不到第一列：\n%s", joined)
	}
	// 實機：`1>    3   4    2  Brawling`（IQ 3、這一級 4 點、目前 2 級）。
	if !strings.Contains(first, "Brawling") {
		t.Errorf("第一列是 %q，預期 Brawling", first)
	}
	for _, want := range []string{" 3 ", " 4 ", " 2 "} {
		if !strings.Contains(first, want) {
			t.Errorf("第一列 %q 少了欄位 %q", first, want)
		}
	}
}
