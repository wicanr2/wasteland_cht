package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestTeleportPrintsLocationName 守住傳送格的那一句話。
//
// nibble 10 記錄 `+0x00` 的**低 6 位是字串編號**（`sub_16B17`：
// `mov bl,0` ／ `and al,3Fh` ／ `jmp sub_17920`），原版兩條路都會印它：
//
//	會問「進新地點？」的格子（bit6 ＝ 1）：`0x16AEA`，問之前先印
//	不問的格子（bit6 ＝ 0）：`0x16A21`，傳送那一支自己印
//
// 這個欄位以前沒接，兩條路都印死的 `TELEPORT.`——編得過、測得過、玩得動，
// 而玩家永遠看不到「進入石英城。」。
func TestTeleportPrintsLocationName(t *testing.T) {
	t.Run("會問的那一種", func(t *testing.T) {
		s := newScene(t)
		// (55,61) 的批次改寫把 (55,62) 換成傳送格（docs/re/72 §2）。
		step(t, s, input.Input{Dir: input.DirUp})
		step(t, s, input.Input{Dir: input.DirDown})
		if !s.Asking() {
			t.Fatal("走回去沒有停下來問")
		}
		if got := s.Message(); !strings.Contains(got, "Entering Ranger Center") {
			t.Errorf("訊息 = %q，預期含 Entering Ranger Center", got)
		}
	})

	t.Run("不問的那一種", func(t *testing.T) {
		s := newScene(t)
		// 地圖 10 的 (26,8) 是 nibble 10 記錄 2，`+0x00` 的 bit6 沒設、
		// 低 6 位 ＝ 24 ＝「You exit the cave.」。
		if err := s.LoadMap(10, 26, 7); err != nil {
			t.Fatalf("換到地圖 10 失敗：%v", err)
		}
		step(t, s, input.Input{Dir: input.DirDown})
		if got := s.Message(); !strings.Contains(got, "You exit the cave") {
			t.Errorf("訊息 = %q，預期含 You exit the cave", got)
		}
	})
}

func newScene(t *testing.T) *Scene {
	t.Helper()
	s, err := New(openRom(t))
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	return s
}

func step(t *testing.T, s *Scene, in input.Input) {
	t.Helper()
	if _, err := s.Update(in); err != nil {
		t.Fatalf("Update(%+v) 失敗：%v", in, err)
	}
}
