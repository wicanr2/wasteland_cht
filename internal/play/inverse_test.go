package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
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
// ⚠ **範圍要用「這一次真的要畫的那段字」算**：中文與英文長度不同，
// 拿錯就會反白到隔壁欄或反白不足，而畫面上只是「反白的位置有點怪」。
func TestInverseSpans(t *testing.T) {
	r := RosterRow{CON: "SER", Weapon: "Crowbar",
		CONInverse: true, WeaponInverse: true}

	en := r.InverseAt(r.CON, r.Weapon)
	for _, c := range []struct {
		col  int
		want bool
	}{
		{colCON - 1, false}, {colCON, true}, {colCON + 2, true},
		{colCON + 3, false},
		{colWeapon - 1, false}, {colWeapon, true},
		{colWeapon + 6, true}, {colWeapon + 7, false},
	} {
		if got := en(c.col); got != c.want {
			t.Errorf("欄 %d 反白 ＝ %v，預期 %v", c.col, got, c.want)
		}
	}

	// 中文那一版：「重傷」兩格、「撬棍」兩格。
	zh := r.InverseAt("重傷", "撬棍")
	if !zh(colCON) || !zh(colCON+1) || zh(colCON+2) {
		t.Error("中文的體力欄反白範圍不是兩格")
	}
	if !zh(colWeapon+1) || zh(colWeapon+2) {
		t.Error("中文的武器欄反白範圍不是兩格")
	}

	// 兩個旗標都沒設就整行正常畫（回 nil，呼叫端少一層判斷）。
	plain := RosterRow{CON: "10"}
	if plain.InverseAt("10", "") != nil {
		t.Error("沒有問題的那一行不該有反白範圍")
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
	con, weapon := rosterFieldsCJK(line)
	if con != "重傷" {
		t.Errorf("切回來的體力欄是 %q，預期「重傷」", con)
	}
	if weapon != "撬棍" {
		t.Errorf("切回來的武器欄是 %q，預期「撬棍」", weapon)
	}
}
