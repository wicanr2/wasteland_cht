package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// `AMM` 那一欄的三道閘（`0x170C4`–`0x170F3`，`docs/re/103` §1）。
//
// ⚠ 少一道就會在不該有數字的地方印數字，而「多印一個數字」不會讓任何
// 既有斷言變紅——那兩欄以前是空的。
func TestRosterAmmoColumn(t *testing.T) {
	// 物品表：ID 1 是遠程武器（類別 2）、ID 2 是近戰（類別 1）。
	items := game.ItemTable{
		{}, {Class: 2}, {Class: 1},
	}
	row := func(equip byte, slot game.Slot) RosterRow {
		c := &game.Character{Name: "A", CON: 20, MaxCON: 20, EquipIndex: equip,
			Items: []game.Slot{slot}}
		return Roster(&game.Party{Members: []*game.Character{c}}, items, nil)[0]
	}
	for _, tc := range []struct {
		name  string
		equip byte
		slot  game.Slot
		want  string
	}{
		{"沒有裝備", 0, game.Slot{ID: 1, Value: 7}, "0"},
		{"遠程武器印低 6 位", 1, game.Slot{ID: 1, Value: 7}, "7"},
		{"高 2 位不算次數", 1, game.Slot{ID: 1, Value: 0x47}, "7"},
		{"bit7 設起來 → 0", 1, game.Slot{ID: 1, Value: 0x87}, "0"},
		{"近戰武器沒有彈藥欄", 1, game.Slot{ID: 2, Value: 7}, "0"},
	} {
		if got := row(tc.equip, tc.slot).Ammo; got != tc.want {
			t.Errorf("%s：得到 %q，預期 %q", tc.name, got, tc.want)
		}
	}
}

// `WEAPON` 那一欄是裝備武器的名字（`0x17165`）。
//
// ⚠ **`+0x1F` 是 1 起算的**：原版算的是 `0xBB ＋ 2n`，而物品陣列從 `+0xBD` 起。
// 當成 0 起算會整批差一格，而症狀是「顯示的武器是背包裡的下一件」——
// 看起來像資料的問題。
func TestRosterWeaponColumnIsOneBased(t *testing.T) {
	names := map[byte]string{11: "knife", 22: "pistol"}
	c := &game.Character{Name: "A", CON: 20, MaxCON: 20, EquipIndex: 2,
		Items: []game.Slot{{ID: 11}, {ID: 22}}}
	got := Roster(&game.Party{Members: []*game.Character{c}}, nil,
		func(id byte) string { return names[id] })[0].Weapon
	if got != "pistol" {
		t.Errorf("EquipIndex 2 應該指到第 1 格（pistol），得到 %q", got)
	}

	c.EquipIndex = 0
	got = Roster(&game.Party{Members: []*game.Character{c}}, nil,
		func(id byte) string { return names[id] })[0].Weapon
	if got != "" {
		t.Errorf("沒有裝備時武器欄要留白，得到 %q", got)
	}
}

// 名單資料行的狀態字要走翻譯目錄（`ui:wound.*`）。
//
// ⚠ 表頭中文化了、資料行沒有的話，畫面上會是「中文表頭 ＋ 英文縮寫」——
// 而覆蓋率統計看不到這種洞（那五個字串是寫死在 Go 裡的，不在原版字串表）。
func TestRosterRowWoundWordsAreTranslated(t *testing.T) {
	s := newScene(t)
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("沒有翻譯目錄（%v）", err)
	}
	for _, w := range []string{"UNC", "SER", "CRT", "MRT", "COM"} {
		key, ok := woundKeys[w]
		if !ok {
			t.Fatalf("%s 沒有登記翻譯鍵", w)
		}
		if len(s.uiText(key)) == 0 {
			t.Errorf("%s（%s）查不到中文", w, key)
		}
	}

	// 整行組出來要有中文，而且欄位還在原本的格子上。
	r := RosterRow{Name: "A", AC: "2", Ammo: "7", MaxCON: "30", CON: "SER", Weapon: "pistol"}
	line := rosterRowCJK(r, s.uiText, nil, 0, false)
	if len(line) == 0 {
		t.Fatal("組不出中文那一行")
	}
	if cjkCells(line) > rosterCols {
		t.Errorf("一行佔 %d 格，超過 %d", cjkCells(line), rosterCols)
	}
	// 狀態字那一欄不能還是英文。
	if strings.Contains(string(line), "SER") {
		t.Errorf("狀態字沒有翻：%q", line)
	}

	// ⚠ 死亡是**字模**不是文字，不走翻譯——原樣留著。
	dead := rosterRowCJK(RosterRow{Name: "A", AC: "2", MaxCON: "30",
		CON: game.WoundDead}, s.uiText, nil, 0, false)
	if !strings.Contains(string(dead), game.WoundDead) {
		t.Errorf("死亡那一格應該保留骷髏字模：%q", dead)
	}
}
