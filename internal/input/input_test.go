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
