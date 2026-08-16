package game

import "testing"

// 傷勢等級的逐值對照（`sub_19A1D`，門檻表 `ds:CCCDh` ＝ `ff f5 ec e2 d8`）。
//
// `docs/re/99` §3.1 一度只到「分支結構確定、語意對照未完成」。
// 對完之後的邊界在 `docs/re/15` §3：等級由 **CON 低位 byte 與門檻的無號比較**
// 決定，而門檻 `F5 EC E2 D8` 當有號數就是 −11／−20／−30／−40。
//
// ⚠ **邊界值兩邊都要驗**。`cmp` 之後接的是 `cmc` ＋ `jnb` ＋ `jz`，
// 也就是「小於**或等於**」——只驗 −11 不驗 −10，寫成嚴格小於也會綠。
func TestWoundLevelEveryBand(t *testing.T) {
	for _, tc := range []struct {
		con  int16
		want int
	}{
		// 等級 5：CON 恰為 0（`+0x1D` 與 `+0x1E` 兩個 byte 都 0）。
		{0, 5},
		// 等級 0：昏迷，還救得回來（`Party.Wipe` 的 WipeSwitch 靠這一段）。
		{-1, 0}, {-5, 0}, {-10, 0},
		// 等級 1–4：門檻上下各驗一格。
		{-11, 1}, {-19, 1},
		{-20, 2}, {-29, 2},
		{-30, 3}, {-39, 3},
		{-40, 4}, {-45, 4}, {-120, 4},
		// CON 為正的人原版根本不查這張表（只有倒下才呼叫）。
		{1, 0}, {25, 0},
	} {
		if got := (&Character{CON: tc.con}).WoundLevel(); got != tc.want {
			t.Errorf("CON %d 的傷勢等級應該是 %d，得到 %d", tc.con, tc.want, got)
		}
	}
}

// 狀態字的索引就是等級（`ds:B233h` ＝ `85 9A 9B 9C 9D 84`，`docs/re/17` §4.4）。
//
// ⚠ 這一條擋的是**索引錯位**：等級 0 是 `UNC`、等級 5 是骷髏字模。
// 曾經寫成「0 ＝ 空字串、5 ＝ UNC」，於是死掉的人在名單上顯示成昏迷，
// 而昏迷的人什麼都不顯示——兩個都不會讓任何斷言變紅。
func TestWoundNamesAreIndexedByLevel(t *testing.T) {
	want := [6]string{"UNC", "SER", "CRT", "MRT", "COM", "\x7f"}
	if WoundNames != want {
		t.Errorf("狀態字表是 %q，預期 %q", WoundNames, want)
	}
	if WoundNames[5] != WoundDead {
		t.Error("等級 5 要用骷髏字模，不是文字")
	}
	// 每一個等級都要有字：空字串表示那一格沒對到訊息碼。
	for lvl, s := range WoundNames {
		if s == "" {
			t.Errorf("等級 %d 沒有狀態字", lvl)
		}
	}
}

// 「救得回來」的邊界：倒下、等級 0、狀態位元全 0（`0x16C3F`–`0x16C51`）。
//
// 原版的迴圈是「找到第一個不是『倒下且有傷勢或狀態』的人就跳出」，
// 跳出去代表這一隊還有救 → 自動換隊；走完整圈才是死亡畫面。
func TestWipeSalvageableBoundary(t *testing.T) {
	party := func(con int16, status byte) *Party {
		return &Party{Members: []*Character{{Name: "A", CON: con, Status: status}}}
	}
	for _, tc := range []struct {
		name   string
		con    int16
		status byte
		want   WipeState
	}{
		{"還站著", 5, 0, WipeNone},
		{"昏迷、沒有狀態 → 還有救", -10, 0, WipeSwitch},
		{"昏迷、但帶著狀態 → 沒救", -10, 1, WipeDead},
		{"掉出昏迷帶（等級 1）→ 沒救", -11, 0, WipeDead},
		{"CON ＝ 0（等級 5）→ 沒救", 0, 0, WipeDead},
	} {
		if got := party(tc.con, tc.status).Wipe(); got != tc.want {
			t.Errorf("%s：得到 %v，預期 %v", tc.name, got, tc.want)
		}
	}
}
