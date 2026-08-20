package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 反白的兩個條件（`docs/re/111` §1）：狀態位元非 0 → 體力欄、
// 裝備武器的 bit7（卡彈）→ 武器名。
func TestRosterInverseFlags(t *testing.T) {
	healthy := &game.Character{Name: "A", CON: 10, MaxCON: 10}
	sick := &game.Character{Name: "B", CON: 10, MaxCON: 10, Status: game.StatusRadiation}
	jammed := &game.Character{Name: "C", CON: 10, MaxCON: 10, EquipIndex: 1,
		Items: []game.Slot{{ID: 13, Value: 0x80 | 3}}}
	p := &game.Party{Members: []*game.Character{healthy, sick, jammed}}

	rows := Roster(p, nil, nil)
	if len(rows) != 3 {
		t.Fatalf("排出 %d 行，預期 3 行", len(rows))
	}
	for i, want := range []struct{ con, weapon bool }{
		{false, false}, {true, false}, {false, true},
	} {
		if rows[i].CONInverse != want.con {
			t.Errorf("第 %d 行的體力欄反白 ＝ %v，預期 %v",
				i+1, rows[i].CONInverse, want.con)
		}
		if rows[i].WeaponInverse != want.weapon {
			t.Errorf("第 %d 行的武器欄反白 ＝ %v，預期 %v",
				i+1, rows[i].WeaponInverse, want.weapon)
		}
	}
}

// `InverseAt` 只反白欄位本身那幾格，不碰左右。
//
// ⚠ **狀態位元反白的是 `MAX` 欄不是 `CON` 欄**：`sub_1708B` 在 `0x17102` 開、
// `0x1711C` 印 MAXCON、`0x1711F` 就關掉，CON 那一欄是關掉之後才印的
// （`docs/re/111` §1）。反到 CON 那一欄的話畫面上照樣有一塊反白，
// 只是反錯了一格——沒有任何症狀看得出來。
//
// ⚠ **範圍要用「這一次真的要畫的那段字」算**：中文與英文長度不同，
// 拿錯就會反白到隔壁欄或反白不足。
func TestInverseSpans(t *testing.T) {
	r := RosterRow{MaxCON: "30", CON: "SER", Weapon: "Crowbar",
		CONInverse: true, WeaponInverse: true}

	en := r.InverseAt(enRoster, r.MaxCON, r.Weapon)
	for _, c := range []struct {
		col  int
		want bool
	}{
		{colMaxCON - 1, false}, {colMaxCON, true}, {colMaxCON + 1, true},
		{colMaxCON + 2, false},
		// CON 那一欄一格都不反白。
		{colCON, false}, {colCON + 1, false}, {colCON + 2, false},
		{colWeapon - 1, false}, {colWeapon, true},
		{colWeapon + 6, true}, {colWeapon + 7, false},
	} {
		if got := en(c.col); got != c.want {
			t.Errorf("欄 %d 反白 ＝ %v，預期 %v", c.col, got, c.want)
		}
	}

	// 中文那一版：欄座標另一套，「撬棍」兩格。
	zh := r.InverseAt(cjkRoster, "100", "撬棍")
	if !zh(cjkColMaxCON) || !zh(cjkColMaxCON+2) || zh(cjkColMaxCON+3) {
		t.Error("中文的上限欄反白範圍不是三格")
	}
	if zh(cjkColCON) {
		t.Error("中文那一版也不該反白 CON 欄")
	}
	if !zh(cjkColWeapon+1) || zh(cjkColWeapon+2) {
		t.Error("中文的武器欄反白範圍不是兩格")
	}

	// 兩個旗標都沒設就整行正常畫（回 nil，呼叫端少一層判斷）。
	plain := RosterRow{CON: "10"}
	if plain.InverseAt(enRoster, "10", "") != nil {
		t.Error("沒有問題的那一行不該有反白範圍")
	}
}

// 序號反白（`ds:471Fh`，`docs/re/128`）只蓋序號與 `>` 那幾格，
// 而且**與另外兩個旗標各自獨立**——原版是三個開關、三段範圍。
func TestIndexInverseSpan(t *testing.T) {
	r := RosterRow{Index: 1, Name: "Hell Razor", MaxCON: "28", CON: "28",
		IndexInverse: true}
	inv := r.InverseAt(enRoster, r.MaxCON, r.Weapon)
	if inv == nil {
		t.Fatal("序號反白的那一行應該要有反白範圍")
	}
	for _, c := range []struct {
		col  int
		want bool
	}{
		{colIndex - 1, false}, {colIndex, true}, {colIndex + 1, true},
		{colIndex + 2, false},
		// 序號反白不該波及名字與其他欄。
		{colName, false}, {colMaxCON, false}, {colWeapon, false},
	} {
		if got := inv(c.col); got != c.want {
			t.Errorf("欄 %d 反白 ＝ %v，預期 %v", c.col, got, c.want)
		}
	}

	// 兩位數的序號多佔一格。
	ten := RosterRow{Index: 10, IndexInverse: true}
	inv10 := ten.InverseAt(enRoster, "", "")
	if !inv10(colIndex+2) || inv10(colIndex+3) {
		t.Error("序號 10 的反白範圍應該是三格（`10>`）")
	}

	// 沒被選中就不反白。
	plain := RosterRow{Index: 2, CON: "10"}
	if plain.InverseAt(enRoster, "10", "") != nil {
		t.Error("沒被選中的那一行不該有反白範圍")
	}
}

// 序號反白的來源：戰鬥是正在下令的那個人，設施是站在櫃檯前的那個人；
// 「誰要進去？」那一步一個都不反白（實機 `42-shop.png`）。
func TestSelectedMemberFollowsTheCounter(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if got := s.selectedMember(); got != 0 {
		t.Errorf("地圖上不該有人被選中，得到 %d", got)
	}
	// 高池鎮的商店：走上傳送格答 Y 進去，停在「誰要進去？」。
	if err := s.LoadMap(10, 30, 25); err != nil {
		t.Fatalf("載入高池鎮失敗：%v", err)
	}
	for _, k := range []byte{'i', 'Y'} {
		if _, err := s.Update(input.Input{Char: k}); err != nil {
			t.Fatalf("送 %q：%v", k, err)
		}
	}
	if s.facility == nil {
		t.Fatal("沒進到設施畫面")
	}
	if got := s.selectedMember(); got != 0 {
		t.Errorf("「誰要進去？」那一步不該有人反白，得到 %d", got)
	}
	// 選第二個人：反白跟著他走。
	if _, err := s.Update(input.Input{Char: '2'}); err != nil {
		t.Fatalf("選人：%v", err)
	}
	if got := s.selectedMember(); got != 2 {
		t.Errorf("櫃檯前是第 2 個人，反白卻在 %d", got)
	}
}

// 中文那一行切回兩欄的字：切點是欄座標，與排版函式共用同一組常數。
func TestRosterFieldsCJK(t *testing.T) {
	r := RosterRow{Index: 1, Name: "海爾", AC: "0", Ammo: "18",
		MaxCON: "28", CON: "重傷", Weapon: "撬棍"}
	line := rosterRowCJK(r, func(k string) string {
		if k == "wound.ser" {
			return "重傷"
		}
		return ""
	}, nil, 0, false)
	if line == "" {
		t.Skip("排不出中文那一版（缺 ui: 文字）")
	}
	maxCON, weapon := rosterFieldsCJK(line)
	if maxCON != "28" {
		t.Errorf("切回來的上限欄是 %q，預期 28", maxCON)
	}
	if weapon != "撬棍" {
		t.Errorf("切回來的武器欄是 %q，預期「撬棍」", weapon)
	}
}
