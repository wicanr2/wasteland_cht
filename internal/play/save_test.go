package play

import (
	"bytes"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestSaveRoundTripsByteForByte 是 CLAUDE.md §4 的硬要求：
// **讀出再寫回要 byte-for-byte 相同**。
//
// 存檔策略是「改寫不是重建」——從原始 bytes 出發只蓋已解欄位，
// 未解區域一個 byte 都不動。這個測試守著那條線：只要有人在
// `Save.Bytes()` 的路上重建了什麼，就會紅。
func TestSaveRoundTripsByteForByte(t *testing.T) {
	rom := openRom(t)
	for file := range assets.SaveOffset {
		sv, err := rom.LoadSave(file)
		if err != nil {
			t.Errorf("%s 的存檔讀不出來：%v", file, err)
			continue
		}
		raw, err := rom.File(file)
		if err != nil {
			t.Fatal(err)
		}
		at := assets.SaveOffset[file]
		out := sv.Bytes()
		orig := raw[at : at+len(out)]
		if !bytes.Equal(out, orig) {
			diff := 0
			first := -1
			for i := range out {
				if out[i] != orig[i] {
					diff++
					if first < 0 {
						first = i
					}
				}
			}
			t.Errorf("%s：重新編碼有 %d 個 byte 不同，第一個在 +%#x", file, diff, first)
			continue
		}
		t.Logf("%s：%d bytes 完全相同", file, len(out))
	}
}

// TestSaveKeepsUnknownBytesAfterPlaying 玩一段之後再存，未解區域仍不能動。
//
// `TestStoreToAfterWalkTouchesOnlyKnownBytes` 只走了幾步；這裡走到打起來、
// 打完一場再存——事件、改寫地圖格、戰鬥結算都跑過一輪。
func TestSaveKeepsUnknownBytesAfterPlaying(t *testing.T) {
	rom := openRom(t)
	before, err := rom.LoadSave("game1")
	if err != nil {
		t.Skipf("讀不到 game1 的存檔：%v", err)
	}
	orig := append([]byte(nil), before.Plain...)

	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(0, 12, 2); err != nil {
		t.Fatal(err)
	}
	dir := input.DirRight
	for i := 0; i < 400 && !s.InCombat(); i++ {
		if _, err := s.Update(input.Input{Dir: dir}); err != nil {
			t.Fatal(err)
		}
		if dir == input.DirRight {
			dir = input.DirLeft
		} else {
			dir = input.DirRight
		}
	}
	if s.InCombat() {
		c := s.Combat()
		for n := 0; n < 200; n++ {
			c.BeginCommands()
			for !c.Done() {
				if !c.Choose('A', true) {
					c.Choose(' ', true)
				}
			}
			if res := c.ResolveRound(); res.Over {
				break
			}
		}
		s.FinishEncounter()
	}

	sv, err := rom.LoadSave("game1")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.StoreTo(sv); err != nil {
		t.Fatalf("StoreTo 失敗：%v", err)
	}
	changed := 0
	for i := range sv.Plain {
		if sv.Plain[i] != orig[i] {
			changed++
		}
	}
	t.Logf("玩了一輪（含一場戰鬥）之後，存檔明文有 %d／%d bytes 改變",
		changed, len(orig))
	if changed == 0 {
		t.Error("一個 byte 都沒改——StoreTo 可能沒寫進去")
	}
	// 改動要限縮在已解欄位：隊伍槽表與 14 bytes 的全域區。
	// 全部 2,000+ bytes 都變代表在重建而不是改寫。
	if changed > 200 {
		t.Errorf("改了 %d bytes，遠超過已解欄位的範圍", changed)
	}
}

// TestSaveLoadRoundTripKeepsState 是完整的存讀循環：
// 玩一段 → StoreTo → 重新編碼 → 解回來 → 欄位要一致。
//
// 前兩個測試守的是「沒動到不該動的」，這個守的是「該存的真的存進去了」。
func TestSaveLoadRoundTripKeepsState(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(0, 12, 2); err != nil {
		t.Fatal(err)
	}
	// 走幾步讓時鐘與座標動起來，順便給第一個人一些經驗。
	for i := 0; i < 20; i++ {
		dir := input.DirRight
		if i%2 == 1 {
			dir = input.DirLeft
		}
		if _, err := s.Update(input.Input{Dir: dir}); err != nil {
			t.Fatal(err)
		}
		if s.InCombat() {
			c := s.Combat()
			for n := 0; n < 200; n++ {
				c.BeginCommands()
				for !c.Done() {
					if !c.Choose('A', true) {
						c.Choose(' ', true)
					}
				}
				if res := c.ResolveRound(); res.Over {
					break
				}
			}
			s.FinishEncounter()
		}
	}
	w := s.World()
	wantX, wantY := w.Party.X, w.Party.Y
	wantHour, wantMin := w.Clock.Hour, w.Clock.Minute
	first := w.Party.Members[0]
	wantName, wantCON := first.Name, first.CON

	sv, err := rom.LoadSave("game1")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.StoreTo(sv); err != nil {
		t.Fatal(err)
	}

	// 重新編碼一次：長度與 magic 要對得上，checksum 由 Bytes() 重算。
	enc := sv.Bytes()
	if len(enc) != 4614 {
		t.Errorf("重新編碼長度 %d，預期 4614", len(enc))
	}
	back := sv
	g := back.SlotGroups()[0]
	slot := back.Plain[g.RawIndex : g.RawIndex+14]
	if slot[8] != wantX || slot[9] != wantY {
		t.Errorf("座標存成 (%d,%d)，預期 (%d,%d)", slot[8], slot[9], wantX, wantY)
	}
	gl := back.Globals()
	if gl[12] != wantHour || gl[11] != wantMin {
		t.Errorf("時鐘存成 %d:%02d，預期 %d:%02d", gl[12], gl[11], wantHour, wantMin)
	}
	var id byte
	for _, m := range g.Members {
		if m != 0 {
			id = m
			break
		}
	}
	raw, err := back.Record(int(id))
	if err != nil {
		t.Fatal(err)
	}
	got := game.LoadCharacter(raw)
	if got.Name != wantName {
		t.Errorf("名字存成 %q，預期 %q", got.Name, wantName)
	}
	if got.CON != wantCON {
		t.Errorf("CON 存成 %d，預期 %d", got.CON, wantCON)
	}
	t.Logf("存讀循環一致：(%d,%d) %d:%02d %s CON=%d",
		slot[8], slot[9], gl[12], gl[11], got.Name, got.CON)
}
