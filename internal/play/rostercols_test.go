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

// 名單四行對著**原版實機截圖**（`docs/re/47` §5 錄下來的那一張）。
//
// ⚠ 這是這一批欄位唯一的外部 oracle：`AMM`／`WEAPON` 兩欄以前是空的，
// 沒有任何測試看得到它們對不對。四行一起驗，因為它們同時釘住三件事——
// 物品表 `+0x04`（容量）→ 物品槽附屬 byte 的初值 → `AMM` 那一欄，
// 以及 `+0x1F` 的 1 起算與名字的單複數拆法。
func TestRosterMatchesOriginalScreenshot(t *testing.T) {
	s := newScene(t)
	want := []RosterRow{
		{Index: 1, Name: "Hell Razor", AC: "0", Ammo: "0", MaxCON: "28", CON: "28", Weapon: "Crowbar"},
		{Index: 2, Name: "Angela Deth", AC: "0", Ammo: "18", MaxCON: "27", CON: "27", Weapon: "VP91Z 9mm pistol"},
		{Index: 3, Name: "Thrasher", AC: "0", Ammo: "0", MaxCON: "34", CON: "34", Weapon: "Knife"},
		{Index: 4, Name: "Snake Vargas", AC: "0", Ammo: "18", MaxCON: "31", CON: "31", Weapon: "VP91Z 9mm pistol"},
	}
	// ⚠ 行首是**序號 ＋ `>`**（`0x1709A` 印數字、`0x170A0` 印 `0x3E`）——
	// 原版畫面上是 `1>Hell Razor`。少了它整行往左差兩格。
	got := Roster(s.World().Party, s.items, s.itemName)
	if len(got) != len(want) {
		t.Fatalf("出廠隊伍應該有 %d 個人，得到 %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 行\n得到 %+v\n預期 %+v", i+1, got[i], want[i])
		}
	}
}

// 名字的單數形是**字根 ＋ 單數字尾**，不是只有字根（`docs/re/17` §4.1）。
//
// ⚠ 字尾是空的那一類（`Crowbar\n\ns\n`）兩種寫法結果一樣，
// 所以**一定要拿 `Kni\nfe\nves\n` 這種來驗**——只用 Crowbar 測會全綠。
func TestSingularKeepsTheSuffix(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"Kni\nfe\nves\n", "Knife"},
		{"Ax\n\nes\n", "Ax"},
		{"Crowbar\n\ns\n", "Crowbar"},
		{"Sledge hammer\n\ns\n", "Sledge hammer"},
		{"沒有分隔碼", "沒有分隔碼"},
		{"", ""},
	} {
		if got := singular(tc.raw); got != tc.want {
			t.Errorf("singular(%q) ＝ %q，預期 %q", tc.raw, got, tc.want)
		}
	}
}
