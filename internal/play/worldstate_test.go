package play

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

func TestDefeatedEncounterStaysCleared(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(10, 28, 2); err != nil {
		t.Fatal(err)
	}
	c, err := s.StartEncounter()
	if err != nil || c == nil {
		t.Fatalf("真實遭遇不存在：%v", err)
	}
	x, y := c.EncX, c.EncY
	rec := c.EncRecord
	for _, e := range c.Battle.Enemies {
		if e != nil {
			e.HP = 0
		}
	}
	s.FinishEncounter()
	for _, at := range []int{3, 4, 5, 6, 7, 8} {
		if rec[at] != 0 {
			t.Fatalf("清場後 record[%d]=%d", at, rec[at])
		}
	}
	assertCleared := func() {
		t.Helper()
		q := s.world.ScanEncounters(s.scanGroups()).Queue
		if _, found := q.Find(10, byte(x), byte(y)); found {
			t.Fatal("已清除的遭遇仍進入戰鬥佇列")
		}
	}
	assertCleared()
	if err := s.LoadMap(0, 55, 62); err != nil {
		t.Fatal(err)
	}
	if err := s.LoadMap(10, 28, 2); err != nil {
		t.Fatal(err)
	}
	assertCleared()
	s.SetQuickSavePath(filepath.Join(t.TempDir(), "world.wlq"))
	if err := s.QuickSave(); err != nil {
		t.Fatal(err)
	}
	fresh := newScene(t)
	fresh.SetQuickSavePath(s.QuickSavePath())
	if err := fresh.QuickLoad(); err != nil {
		t.Fatal(err)
	}
	s = fresh
	assertCleared()
}

func TestEncounterKeepsSurvivors(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(10, 28, 2); err != nil {
		t.Fatal(err)
	}
	c, err := s.StartEncounter()
	if err != nil || c == nil {
		t.Fatalf("真實遭遇不存在：%v", err)
	}
	c.Battle.Enemies = [game.EnemySlots]*game.Enemy{}
	c.Battle.AddEnemy(0, 0, &game.Enemy{HP: 7})
	c.Battle.AddEnemy(0, 1, &game.Enemy{HP: 0})
	s.FinishEncounter()
	if c.EncRecord[4] != 1 || c.EncRecord[3] == 0 {
		t.Fatalf("中止時遺失存活敵人：%v", c.EncRecord[:9])
	}
}

func TestEncounterUsesActivePartyGroup(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(10, 28, 2); err != nil {
		t.Fatal(err)
	}
	// 縮小接線條件：目前組改為 1，0 組沒有成員。
	// 場景仍使用真實 JACKIE 遭遇與正常 E 指令。
	s.groupID = 1
	g := s.save.SlotGroups()[0]
	clear(s.save.Plain[g.RawIndex : g.RawIndex+7])
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	if !s.InCombat() || !hireOfferOf(s.Combat()) {
		t.Fatalf("第 1 組無法與 JACKIE 互動：mode=%s", s.Mode())
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'H'})
	if !s.combat.HirePicking() {
		t.Fatal("NPC 雇用選單未開啟")
	}
}

func TestQuickLoadLeavesWipeAndTitle(t *testing.T) {
	for _, mode := range []string{"wipe", "title", "confirm", "character"} {
		t.Run(mode, func(t *testing.T) {
			s := newScene(t)
			s.SetQuickSavePath(filepath.Join(t.TempDir(), "quick.wlq"))
			if err := s.QuickSave(); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "wipe":
				s.beginWipe()
			case "title":
				s.BeginTitle()
			case "confirm":
				step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
			case "character":
				step(t, s, input.Input{Dir: input.DirNone, Char: '1'})
			}
			step(t, s, input.Input{Dir: input.DirNone, Fn: input.FnQuickLoad})
			if s.Mode() != "map" || !s.partyCanAct() {
				t.Fatalf("F9 後 mode=%s, canAct=%v", s.Mode(), s.partyCanAct())
			}
		})
	}
}

func TestWorldStateRejectsCorruptionWithoutMutation(t *testing.T) {
	s := newScene(t)
	if err := s.LoadMap(10, 28, 2); err != nil {
		t.Fatal(err)
	}
	raw, err := s.encodeWorldState()
	if err != nil {
		t.Fatal(err)
	}
	var state worldState
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatal(err)
	}
	state.Maps[len(state.Maps)-1].Terrain[0] = 255
	bad, _ := json.Marshal(state)
	if _, err := s.decodeWorldState(bad); err == nil {
		t.Fatal("接受非法地形")
	}
	after, _ := s.encodeWorldState()
	if !bytes.Equal(raw, after) {
		t.Fatal("驗證失敗卻改到現有世界")
	}
}

func TestQuickVersionOneStillLoads(t *testing.T) {
	s := newScene(t)
	s.SetQuickSavePath(filepath.Join(t.TempDir(), "v1.wlq"))
	if err := s.QuickSave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(s.QuickSavePath())
	if err != nil {
		t.Fatal(err)
	}
	parts, _, err := parseQuick(raw)
	if err != nil {
		t.Fatal(err)
	}
	raw = raw[:len(raw)-4-len(parts[3])]
	binary.LittleEndian.PutUint16(raw[4:6], 1)
	if err := os.WriteFile(s.QuickSavePath(), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	s.BeginTitle()
	if err := s.QuickLoad(); err != nil || s.Mode() != "map" {
		t.Fatalf("舊檔讀取失敗：%v, mode=%s", err, s.Mode())
	}
	// v2 不可用空的第四段掩蓋遺失的地圖狀態。
	binary.LittleEndian.PutUint16(raw[4:6], quickVersion)
	raw = binary.LittleEndian.AppendUint32(raw, 0)
	if _, _, err := parseQuick(raw); err == nil {
		t.Fatal("v2 接受空白地圖狀態")
	}
}

func TestWorldSaveKeepsActiveParty(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'D'})
	step(t, s, input.Input{Dir: input.DirNone, Char: '4'})
	if err := s.SwitchGroup(1); err != nil {
		t.Fatal(err)
	}
	name := s.world.Party.Members[0].Name
	s.SetQuickSavePath(filepath.Join(t.TempDir(), "split.wlq"))
	if err := s.QuickSave(); err != nil {
		t.Fatal(err)
	}
	fresh := newScene(t)
	fresh.SetQuickSavePath(s.QuickSavePath())
	if err := fresh.QuickLoad(); err != nil {
		t.Fatal(err)
	}
	if fresh.groupID != 1 || len(fresh.world.Party.Members) != 1 || fresh.world.Party.Members[0].Name != name {
		t.Fatal("讀檔未回到目前操作的分隊")
	}
}
