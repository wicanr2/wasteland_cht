package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 空白鍵在地圖畫面展開／收起隊伍名單（原版 `ds:46B9h`，`docs/re/126` §2）。
func TestSpaceTogglesMapRoster(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: ' '}); err != nil {
		t.Fatal(err)
	}
	if !s.rosterOn {
		t.Fatal("空白鍵沒有展開名單")
	}
	// 展開時標籤只剩 ROSTER OFF（`0x16BAB` 的 `0x200B`，docs/re/129 §4）。
	labels := s.boxLabels()
	if len(labels) != 1 {
		t.Fatalf("名單展開時應該只有一個標籤，得到 %d 個", len(labels))
	}
	if s.Frame() == nil {
		t.Fatal("名單展開時畫不出一幀")
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Char: ' '}); err != nil {
		t.Fatal(err)
	}
	if s.rosterOn {
		t.Fatal("再按一次空白沒有收起名單")
	}
}

// 名單開著時一有新訊息就自動關回去（實機 94–97 截圖：撞牆印 Smack!
// 的那一步，標籤變回 ROSTER ON、訊息視窗取回畫面）。
func TestMapRosterFoldsOnMessage(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	s.rosterOn = true
	// 往一個會被擋住的方向走到印出訊息為止（出廠位置四周有山）。
	for _, d := range []input.Direction{input.DirUp, input.DirLeft, input.DirDown, input.DirRight} {
		if _, err := s.Update(input.Input{Dir: d}); err != nil {
			t.Fatal(err)
		}
		if s.message != "" || s.cjk != "" {
			break
		}
	}
	if s.message == "" && s.cjk == "" {
		t.Skip("四個方向都沒印訊息，換個起點再驗")
	}
	if s.rosterOn {
		t.Fatal("印了訊息之後名單應該自動收起")
	}
}

// 創角要停在 `Keep this char?`：答 N 整組重擲、答 Y 才入隊（docs/re/21 §5）。
func TestCreateKeepLoopRerolls(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	s.beginRoster()
	deleteOneForCreate(t, s)
	if _, err := s.Update(key('C')); err != nil {
		t.Fatal(err)
	}
	for _, ch := range "AB" {
		if _, err := s.Update(key(byte(ch))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionConfirm}); err != nil {
		t.Fatal(err)
	}
	if !s.roster.keep || s.roster.cand == nil {
		t.Fatal("Enter 之後沒有停在 Keep this char? 上")
	}
	before := len(s.World().Party.Members)
	// 答 N 五次：每次都要還在 keep 迴圈裡、不入隊；屬性至少要變過一次
	// （5d6 取三連擲 35 次全部相同的機率可以忽略）。
	first := s.roster.cand.Attributes
	changed := false
	for i := 0; i < 5; i++ {
		if _, err := s.Update(key('N')); err != nil {
			t.Fatal(err)
		}
		if !s.roster.keep || s.roster.cand == nil {
			t.Fatal("答 N 之後應該重擲並留在 Keep this char? 上")
		}
		if s.roster.cand.Attributes != first {
			changed = true
		}
	}
	if !changed {
		t.Error("重擲五次屬性一次都沒變——rollCandidate 沒有真的重擲")
	}
	if len(s.World().Party.Members) != before {
		t.Fatal("答 N 不該把人放進隊伍")
	}
	if _, err := s.Update(key('Y')); err != nil {
		t.Fatal(err)
	}
	if s.roster.keep {
		t.Fatal("答 Y 之後不該還停在 Keep this char? 上")
	}
	if len(s.World().Party.Members) != before+1 {
		t.Fatal("答 Y 之後隊伍應該多一個人")
	}
}
