package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// `ENC` 按下去要真的走驅動器，不能還停在 "not wired yet"。
//
// 這一條擋的是**接線漏洞**：`StartEncounter` 早就寫好了，
// 缺的只是指令列這個入口（`docs/re/94` §4）。
func TestEncRunsTheDriver(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'E'}); err != nil {
		t.Fatalf("按 E：%v", err)
	}
	if strings.Contains(s.Message(), "not wired") {
		t.Fatalf("ENC 還沒接上：%q", s.Message())
	}
	// 出廠存檔只有一組人，站在沒有敵人的格子上 → 四條合法出口之一。
	msg := s.Message()
	ok := strings.Contains(msg, "Nothing to fight") ||
		strings.Contains(msg, "ATTACKED") ||
		msg == encOffMapPrompt ||
		msg == encNoEnemyPrompt
	if !ok {
		t.Fatalf("ENC 的訊息不在四條合法出口裡：%q", msg)
	}
}

// 這一組沒有敵人在打時，原版**還是會問**「要不要跑一個戰鬥回合」
//（`sub_11F76` 的 `0x11FD9`，字串 20 ＋ 76）。答 Y 就進指令階段。
//
// ⚠ 這條路以前整個沒接：remake 直接印「Nothing to fight here.」
// （那還是一句**重製版自己寫的**話，原版沒有這條字串）。
// 少掉的是一整個玩法——在空地上換武器、裝填、用道具都要靠它。
func TestEncOffersARoundWithNoEnemies(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	if s.Mode() != "confirm" {
		t.Fatalf("按 E 之後應該停在 Y／N，停在 %s（訊息 %q）", s.Mode(), s.Message())
	}
	if s.Message() != encNoEnemyPrompt && len(s.CJK()) == 0 {
		t.Errorf("問句不對：%q", s.Message())
	}

	// 答 N ＝ 什麼都不發生。
	step(t, s, input.Input{Dir: input.DirNone, Char: 'N'})
	if s.InCombat() {
		t.Fatal("答 N 卻進了戰鬥")
	}

	// 答 Y ＝ 進指令階段（沒有敵人也照問）。
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Fatal("答 Y 應該進指令階段")
	}
	c := s.Combat()
	if c.Battle.EnemiesLeft() != 0 {
		t.Errorf("這一回合不該有敵人，得到 %d", c.Battle.EnemiesLeft())
	}
	if c.Done() {
		t.Error("指令階段應該還在問人，不該一開始就結束")
	}
	// ⚠ 沒有敵人時 `Over()` 立刻成立，所以下完令這一回合就收掉——
	// 與原版一樣，不是卡在戰鬥裡。
	for !c.Done() {
		if !c.Choose('E', true) { // E ＝ Evade，任何人都選得了
			c.Choose(' ', true)
		}
	}
	if res := c.ResolveRound(); !res.Over || !res.Won {
		t.Errorf("沒有敵人的回合應該立刻結束：over=%v won=%v", res.Over, res.Won)
	}
}

// 別組不在這張地圖上時要問一句，答 N 就跳過（原版 `0x11DA6`）。
func TestEncAsksAboutOffMapParty(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	// 手動在第 1 組塞一個人，放到別張地圖上。
	groups := s.save.SlotGroups()
	slot := s.save.Plain[groups[1].RawIndex : groups[1].RawIndex+14]
	slot[0] = 1
	slot[10] = byte(s.blockID + 1)

	n, ok := s.encOffMapGroup()
	if !ok || n != 1 {
		t.Fatalf("應該找到第 1 組在別張地圖上，得到 (%d, %v)", n, ok)
	}

	s.encAsk = n + 1
	s.message = encOffMapPrompt
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: 'N'}); err != nil {
		t.Fatalf("按 N：%v", err)
	}
	if s.encAsk != 0 {
		t.Fatalf("答 N 之後還停在問句上（encAsk ＝ %d）", s.encAsk)
	}
	if s.groupID != 0 {
		t.Fatalf("答 N 不該切組，現在在第 %d 組", s.groupID)
	}
}
