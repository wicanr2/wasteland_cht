package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

func key(ch byte) input.Input { return input.Input{Dir: input.DirNone, Char: ch} }

// USE 的三層選單要走得完：挑人 → S/I/A → 選項目 → 判定。
//
// **三條分支要一起有**，缺一條就是「按了沒反應」——那是最難查的半成品
// （docs/re/79 §5）。
func TestUseWalksAllThreeBranches(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		key  byte
		kind game.UseKind
		name string
	}{
		{'S', game.UseSkill, "技能"},
		{'I', game.UseItem, "物品"},
		{'A', game.UseAttribute, "屬性"},
	} {
		if _, err := s.Update(key('U')); err != nil {
			t.Fatal(err)
		}
		if s.use.stage == useStageOff {
			t.Fatalf("%s：按 U 沒有開始 USE", tc.name)
		}
		// 出廠隊伍不只一人，所以會先問挑誰。
		if s.use.stage == useStageMember {
			if !strings.Contains(s.Message(), "Which player") {
				t.Errorf("%s：沒問挑誰，訊息是 %q", tc.name, s.Message())
			}
			if _, err := s.Update(key('1')); err != nil {
				t.Fatal(err)
			}
		}
		if s.use.stage != useStageKind {
			t.Fatalf("%s：挑完人應該問種類，stage=%d", tc.name, s.use.stage)
		}
		if _, err := s.Update(key(tc.key)); err != nil {
			t.Fatal(err)
		}
		if s.use.stage != useStagePick {
			t.Fatalf("%s：按 %c 之後應該列清單，stage=%d（訊息 %q）",
				tc.name, tc.key, s.use.stage, s.Message())
		}
		if s.use.kind != tc.kind {
			t.Errorf("%s：種類應該是 %d，得到 %d", tc.name, tc.kind, s.use.kind)
		}
		if len(s.use.options) == 0 {
			t.Fatalf("%s：清單是空的", tc.name)
		}
		t.Logf("%s：%d 項——%s", tc.name, len(s.use.options), s.Message())

		// 選第一項 → 問方向（docs/re/132）→ 原地施用 → 狀態歸零。
		if _, err := s.Update(key('1')); err != nil {
			t.Fatal(err)
		}
		if s.use.stage != useStageDir {
			t.Fatalf("%s：選完應該問方向，stage=%d", tc.name, s.use.stage)
		}
		if _, err := s.Update(key(' ')); err != nil {
			t.Fatal(err)
		}
		if s.use.stage != useStageOff {
			t.Errorf("%s：選完之後應該回到地圖，stage=%d", tc.name, s.use.stage)
		}
		if s.Message() == "" {
			t.Errorf("%s：判定完了卻沒有訊息", tc.name)
		}
	}
}

// 屬性那條傳的是**角色記錄位移**（0x0E–0x14），不是屬性索引（docs/re/32 §4）。
func TestUseAttributeUsesRecordOffset(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(key('U')); err != nil {
		t.Fatal(err)
	}
	if s.use.stage == useStageMember {
		if _, err := s.Update(key('1')); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Update(key('A')); err != nil {
		t.Fatal(err)
	}
	if len(s.use.options) != game.AttrCount {
		t.Fatalf("屬性應該有 %d 項，得到 %d", game.AttrCount, len(s.use.options))
	}
	for i, o := range s.use.options {
		if want := byte(0x0E + i); o.id != want {
			t.Errorf("第 %d 個屬性的參數應該是記錄位移 %#x，得到 %#x", i, want, o.id)
		}
	}
}

// ESC 隨時可以取消，而且不會留下半開的狀態。
func TestUseCancels(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(key('U')); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel}); err != nil {
		t.Fatal(err)
	}
	if s.use.stage != useStageOff {
		t.Error("ESC 沒有取消 USE")
	}
	// 取消之後方向鍵要能正常走路。
	x := s.World().Party.X
	if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
		t.Fatal(err)
	}
	if s.World().Party.X == x {
		t.Error("取消之後走不動——狀態沒清乾淨")
	}
}

// USE 進行中方向鍵不走路（與手札、戰鬥同一條規矩，docs/spec/24）。
func TestUseBlocksWalking(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(key('U')); err != nil {
		t.Fatal(err)
	}
	x, y := s.World().Party.X, s.World().Party.Y
	for i := 0; i < 4; i++ {
		if _, err := s.Update(input.Input{Dir: input.DirRight}); err != nil {
			t.Fatal(err)
		}
	}
	if s.World().Party.X != x || s.World().Party.Y != y {
		t.Errorf("USE 進行中卻走了路：(%d,%d) → (%d,%d)",
			x, y, s.World().Party.X, s.World().Party.Y)
	}
}

// 物品名的偏移要對——差一格會整批換成別的武器，而畫面看起來完全正常。
//
// 正對照是出廠裝備（docs/re/21 §5.1）：手槍 ＋ 八個彈匣。
func TestItemNamesMatchStartingKit(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	m := s.World().Party.Members[0]
	var got []string
	for _, it := range m.Items {
		if it.ID != 0 {
			got = append(got, s.itemName(it.ID))
		}
	}
	if len(got) == 0 {
		t.Fatal("出廠角色身上沒有東西")
	}
	t.Logf("出廠裝備：%v", got[:min(len(got), 4)])
	if !strings.Contains(got[0], "pistol") {
		t.Errorf("第一件應該是手槍，得到 %q", got[0])
	}
	clips := 0
	for _, n := range got {
		if strings.Contains(n, "clip") {
			clips++
		}
	}
	if clips != 8 {
		t.Errorf("應該有八個彈匣，得到 %d", clips)
	}
}
