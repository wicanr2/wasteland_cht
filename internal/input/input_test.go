package input

import "testing"

func TestDirections(t *testing.T) {
	cases := map[Key]Direction{
		KeyUp: DirUp, KeyDown: DirDown, KeyLeft: DirLeft, KeyRight: DirRight,
		KeyW: DirUp, KeyS: DirDown, KeyA: DirLeft, KeyD: DirRight,
	}
	for k, want := range cases {
		if in := Read([]Key{k}, nil); in.Dir != want {
			t.Fatalf("鍵 %d 對到方向 %d，應該是 %d", k, in.Dir, want)
		}
	}
	if in := Read(nil, nil); in.Dir != DirNone {
		t.Fatalf("沒按鍵時方向是 %d，應該是 DirNone", in.Dir)
	}
}

// ESC 是取消不是離開，離開要按 F10——按錯鍵不該丟進度。
func TestEscapeIsCancelNotQuit(t *testing.T) {
	if in := Read([]Key{KeyEscape}, nil); in.Action != ActionCancel {
		t.Fatalf("ESC 對到 %d，應該是 ActionCancel", in.Action)
	}
	if in := Read([]Key{KeyF10}, nil); in.Action != ActionQuit {
		t.Fatalf("F10 對到 %d，應該是 ActionQuit", in.Action)
	}
}

func TestChar(t *testing.T) {
	if in := Read(nil, []rune{'B'}); in.Char != 'B' {
		t.Fatalf("字元是 %q，應該是 'B'", in.Char)
	}
	if in := Read(nil, []rune{'\n', 'S'}); in.Char != 'S' {
		t.Fatalf("控制字元應該被跳過，得到 %q", in.Char)
	}
}

// Page Up／Page Down 走 `Scroll`，**不碰 `Dir`**。
//
// ⚠ 兩者共用一個欄位的話，手札裡的一頁捲動在地圖上會變成走一步——
// 而症狀是「按了 Page Down，隊伍往南走了一格」。
func TestPageKeysAreScrollNotDirection(t *testing.T) {
	for _, tc := range []struct {
		key  Key
		want Scroll
	}{
		{KeyPageUp, ScrollUp},
		{KeyPageDown, ScrollDown},
	} {
		in := Read([]Key{tc.key}, nil)
		if in.Scroll != tc.want {
			t.Errorf("鍵 %d 的 Scroll ＝ %d，預期 %d", tc.key, in.Scroll, tc.want)
		}
		if in.Dir != DirNone {
			t.Errorf("鍵 %d 不該產生方向（得到 %d）", tc.key, in.Dir)
		}
		if in.Action != ActionNone {
			t.Errorf("鍵 %d 不該產生動作（得到 %d）", tc.key, in.Action)
		}
	}
	if in := Read(nil, nil); in.Scroll != ScrollNone {
		t.Errorf("沒按鍵時 Scroll ＝ %d，應該是 ScrollNone", in.Scroll)
	}
}
