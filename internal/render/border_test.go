package render

import "testing"

// 計量表由上到下的字模（`docs/re/120` §3）。索引 0 ＝ 字元列 9。
func TestMeterGlyphs(t *testing.T) {
	// 沒有蓋氏計數器：底座換一組，讀數一律當成沒有 → 整條空的。
	got := MeterGlyphs(0, false)
	want := [MeterRows]int{
		meterCapPlain + 2, meterCapPlain, // 列 9、10：頂蓋
		meterEmpty, meterEmpty, meterEmpty, meterEmpty, // 列 11–14：空管
		meterNoCounter + 2, meterNoCounter, // 列 15、16：底座
	}
	if got != want {
		t.Errorf("沒有計數器時應該是空管：\n得到 %#x\n期望 %#x", got, want)
	}

	// 有計數器但視野裡沒有輻射格（保留值）→ 一樣是空管，只有底座不同。
	got = MeterGlyphs(MeterNoReading, true)
	want[MeterRows-1], want[MeterRows-2] = meterHaveCounter, meterHaveCounter+2
	if got != want {
		t.Errorf("沒有讀數時應該是空管：\n得到 %#x\n期望 %#x", got, want)
	}

	// 讀數 0（踩在輻射格上）→ 9 階全滿：四格 ＋ 半格，
	// 而半格頂到頂之後改成把頂蓋換成亮的那一組（`0x17F25` 的 `jz`）。
	got = MeterGlyphs(0, true)
	want = [MeterRows]int{
		meterCapLit + 2, meterCapLit,
		meterFullLow, meterFullLow, meterFullHigh, meterFullHigh,
		meterHaveCounter + 2, meterHaveCounter,
	}
	if got != want {
		t.Errorf("讀數 0 應該滿格：\n得到 %#x\n期望 %#x", got, want)
	}

	// 讀數 8（退一階）→ n ＝ 8：四格滿、沒有半格，頂蓋是暗的。
	got = MeterGlyphs(8, true)
	want[0], want[1] = meterCapPlain+2, meterCapPlain
	if got != want {
		t.Errorf("讀數 8 應該四格滿、頂蓋是暗的：\n得到 %#x\n期望 %#x", got, want)
	}

	// 讀數 24 → step 3、n ＝ 6：三格滿 ＋ 沒有半格。列 > 12 的那兩格換色。
	got = MeterGlyphs(24, true)
	want = [MeterRows]int{
		meterCapPlain + 2, meterCapPlain,
		meterEmpty, meterFullLow, meterFullHigh, meterFullHigh,
		meterHaveCounter + 2, meterHaveCounter,
	}
	if got != want {
		t.Errorf("讀數 24 應該三格滿：\n得到 %#x\n期望 %#x", got, want)
	}

	// 讀數 ≥ 72（step ≥ 9）→ 空管，與沒有讀數一樣。
	if MeterGlyphs(72, true) != MeterGlyphs(MeterNoReading, true) {
		t.Error("讀數 72 以上與沒有讀數應該畫成一樣")
	}
}
