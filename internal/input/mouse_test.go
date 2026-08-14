package input

import "testing"

const allRegions = 0xFFFFFFFF

func TestMessageRowHotkey(t *testing.T) {
	var s Screen
	s.Mask = 1 << 2
	s.Rows.Set(1, 'Y')
	s.Rows.Set(2, 'N')

	// 第 1 列是 y 5–12，第 2 列是 13–20。
	for _, c := range []struct {
		y    int
		want byte
	}{{5, 'Y'}, {12, 'Y'}, {13, 'N'}, {20, 'N'}} {
		got, ok := Click(200, c.y, s)
		if !ok || got != c.want {
			t.Fatalf("y=%d 得到 %q,%v，應該是 %q", c.y, got, ok, c.want)
		}
	}
	// 沒登記熱鍵的列不算按鍵。
	if _, ok := Click(200, 30, s); ok {
		t.Fatal("空白列不該送出按鍵")
	}
}

func TestRowHotkeysScroll(t *testing.T) {
	var r RowHotkeys
	r.Set(1, 'A')
	r.Set(2, 'B')
	r.Scroll()
	if r[0] != 'B' || r[1] != 0 {
		t.Fatalf("捲動後是 %q %q，應該是 'B' 0", r[0], r[1])
	}
	r.Clear()
	if r[0] != 0 {
		t.Fatal("Clear 沒清乾淨")
	}
}

// 遮罩沒設的區域一律不試——這是「戰鬥選單時點地圖沒反應」的機制。
func TestMaskGatesRegions(t *testing.T) {
	var s Screen
	s.Rows.Set(1, 'Y')
	if _, ok := Click(200, 6, s); ok {
		t.Fatal("遮罩為 0 時不該有任何熱區生效")
	}
	s.Mask = 1 << 2
	if _, ok := Click(200, 6, s); !ok {
		t.Fatal("遮罩設起來之後訊息視窗該生效")
	}
}

// 熱區重疊時由表的順序決定：地圖視窗（第 1 筆）壓過訊息視窗（第 2 筆）。
func TestFirstMatchingRegionWins(t *testing.T) {
	var s Screen
	s.Mask = allRegions
	s.Rows.Set(2, 'Y')
	got, ok := Click(200, 20, s) // 兩塊都涵蓋這個點
	if !ok || got != 'L' {
		t.Fatalf("重疊處得到 %q,%v，應該是地圖視窗的 'L'", got, ok)
	}
}

func TestMapViewQuadrants(t *testing.T) {
	cases := []struct {
		x, y int
		want byte
	}{
		{144, 20, 'I'},  // 上
		{144, 110, 'K'}, // 下
		{20, 64, 'J'},   // 左
		{250, 64, 'L'},  // 右
	}
	s := Screen{Mask: 1 << 1}
	for _, c := range cases {
		got, ok := Click(c.x, c.y, s)
		if !ok || got != c.want {
			t.Fatalf("(%d,%d) 得到 %q，應該是 %q", c.x, c.y, got, c.want)
		}
	}
	if got, ok := Click(144, 64, s); !ok || got != 0x1B {
		t.Fatalf("中央小框得到 %q，應該是 ESC", got)
	}
	s.EscAsSpace = true
	if got, ok := Click(144, 64, s); !ok || got != ' ' {
		t.Fatalf("EscAsSpace 時中央小框得到 %q，應該是空白", got)
	}
}

func TestRosterRowLimitedByPartySize(t *testing.T) {
	s := Screen{Mask: 1 << 0, PartySize: 3}
	if got, ok := Click(100, 0x7D, s); !ok || got != '1' {
		t.Fatalf("第一列得到 %q,%v，應該是 '1'", got, ok)
	}
	if got, ok := Click(100, 0x7D+16, s); !ok || got != '3' {
		t.Fatalf("第三列得到 %q,%v，應該是 '3'", got, ok)
	}
	if _, ok := Click(100, 0x7D+24, s); ok {
		t.Fatal("超過隊伍人數的列不該送出按鍵")
	}
}

func TestFixedButtonKeys(t *testing.T) {
	s := Screen{Mask: allRegions}
	if got, ok := Click(310, 30, s); !ok || got != 0xC8 {
		t.Fatalf("右側上箭頭得到 %q，應該是 0xC8", got)
	}
	if got, ok := Click(10, 190, s); !ok || got != 0x0D {
		t.Fatalf("底列得到 %q，應該是 Enter", got)
	}
}
