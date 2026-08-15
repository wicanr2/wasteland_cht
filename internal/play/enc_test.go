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
	// 出廠存檔只有一組人，站在沒有敵人的格子上 → 三條合法出口之一。
	msg := s.Message()
	ok := strings.Contains(msg, "Nothing to fight") ||
		strings.Contains(msg, "ATTACKED") ||
		msg == encOffMapPrompt
	if !ok {
		t.Fatalf("ENC 的訊息不在三條合法出口裡：%q", msg)
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
