package game

import "testing"

// 驗收 1：指令碼照熱鍵字母表，不照選單顯示順序。
//
// 對照組是原版 ds:A44Dh 那 8 個 byte，寫死在這裡——
// 拿常數自己對自己驗等於什麼都沒驗。
func TestCommandCodesFollowKeyTable(t *testing.T) {
	original := [8]byte{0x20, 0x48, 0x41, 0x57, 0x52, 0x45, 0x4C, 0x55} // " HAWRELU"
	if CommandKeys != original {
		t.Fatalf("熱鍵字母表與原版不符：%q vs %q", CommandKeys, original)
	}
	for _, tc := range []struct {
		key byte
		cmd Command
	}{
		{' ', CmdNone}, {'H', CmdHire}, {'A', CmdAttack}, {'W', CmdWeapon},
		{'R', CmdRun}, {'E', CmdEvade}, {'L', CmdLoad}, {'U', CmdUse},
	} {
		got, ok := CommandFromKey(tc.key)
		if !ok || got != tc.cmd {
			t.Errorf("按鍵 %q → %d（ok=%v），預期 %d", tc.key, got, ok, tc.cmd)
		}
	}
	if _, ok := CommandFromKey('Z'); ok {
		t.Error("字母表裡沒有 Z，應該回 false 讓呼叫端重問")
	}

	// 選單顯示順序是 Run/Use/Hire/Evade/Attack/Weapon/Load——刻意與指令碼不同。
	menuOrder := []Command{CmdRun, CmdUse, CmdHire, CmdEvade, CmdAttack, CmdWeapon, CmdLoad}
	same := true
	for i, c := range menuOrder {
		if int(c) != i+1 {
			same = false
		}
	}
	if same {
		t.Error("選單順序與指令碼一致了——那表示常數被改成照選單編號，規則會對錯人")
	}
}

// 驗收 2：CON ≤ 0 的成員不能下令。
func TestCanCommandFollowsCON(t *testing.T) {
	for _, tc := range []struct {
		con  int16
		want bool
	}{{20, true}, {1, true}, {0, false}, {-5, false}, {-60, false}} {
		if got := CanCommand(&Character{CON: tc.con}); got != tc.want {
			t.Errorf("CON %d：能下令 ＝ %v，預期 %v", tc.con, got, tc.want)
		}
	}
	if CanCommand(nil) {
		t.Error("nil 成員不該能下令")
	}
}

// 驗收 3：迴避不改任何狀態，但門檻基礎值變 60。
func TestDefenceBase(t *testing.T) {
	for _, tc := range []struct {
		cmd  Command
		want int
	}{
		{CmdEvade, 60}, {CmdAttack, 50}, {CmdNone, 40}, {CmdRun, 40},
		{CmdHire, 40}, {CmdWeapon, 40}, {CmdLoad, 40}, {CmdUse, 40}, {CmdPartyRan, 40},
	} {
		if got := tc.cmd.DefenceBase(); got != tc.want {
			t.Errorf("指令 %d：基礎值 %d，預期 %d", tc.cmd, got, tc.want)
		}
	}
}

// 驗收 4：整隊逃跑之後每個人的指令碼都是 8、參數都是同一個方向。
func TestPartyFlees(t *testing.T) {
	p := NewCommandPhase(4)
	p.Set(0, CmdAttack, 3)
	p.Set(1, CmdEvade, 0)

	p.PartyFlees(FleeLeft)
	if !p.Ran || p.RanTo != FleeLeft {
		t.Fatalf("應該記下整隊往 %d 跑，得到 Ran=%v RanTo=%d", FleeLeft, p.Ran, p.RanTo)
	}
	for i := range p.Cmd {
		if p.Cmd[i] != CmdPartyRan || p.Arg[i] != byte(FleeLeft) {
			t.Fatalf("成員 %d：指令 %d 參數 %d，預期 %d／%d",
				i, p.Cmd[i], p.Arg[i], CmdPartyRan, byte(FleeLeft))
		}
		// 逃跑之後不再享有迴避的門檻。
		if p.Defence(i) != baseDefault {
			t.Errorf("成員 %d 逃跑之後的門檻應該是預設值", i)
		}
	}
}

// 驗收 6：新開的指令階段每個人都是「沒下令」。
func TestNewCommandPhaseStartsClear(t *testing.T) {
	p := NewCommandPhase(7)
	for i := range p.Cmd {
		if p.Cmd[i] != CmdNone || p.Arg[i] != 0 {
			t.Fatalf("成員 %d 沒有歸零：%d／%d", i, p.Cmd[i], p.Arg[i])
		}
		if p.Defence(i) != baseDefault {
			t.Fatalf("沒下令的門檻應該是預設值 %d", baseDefault)
		}
	}
	if p.Set(7, CmdAttack, 0) {
		t.Error("編號越界應該回 false")
	}
	if p.Defence(-1) != baseDefault || p.Defence(99) != baseDefault {
		t.Error("越界的門檻應該回預設值而不是 panic")
	}
}

// 驗收 5：方向的兩組按鍵對到同五個方向（ds:A45Ch）。
func TestFleeDirectionKeys(t *testing.T) {
	original := [9]byte{0x49, 0x4B, 0x4A, 0x4C, 0x20, 0xC8, 0xD0, 0xCB, 0xCD}
	if FleeKeys != original {
		t.Fatalf("方向字母表與原版不符：%v vs %v", FleeKeys, original)
	}
	pairs := []struct {
		letter, arrow byte
		want          FleeDirection
	}{
		{'I', 0xC8, FleeUp}, {'K', 0xD0, FleeDown},
		{'J', 0xCB, FleeLeft}, {'L', 0xCD, FleeRight},
	}
	for _, p := range pairs {
		a, ok1 := FleeDirectionFromKey(p.letter)
		b, ok2 := FleeDirectionFromKey(p.arrow)
		if !ok1 || !ok2 || a != p.want || b != p.want {
			t.Errorf("%q 與掃描碼 %#x 應該都是 %d，得到 %d／%d", p.letter, p.arrow, p.want, a, b)
		}
	}
	if d, ok := FleeDirectionFromKey(' '); !ok || d != FleeStay {
		t.Errorf("空白鍵應該是「不動」，得到 %d（ok=%v）", d, ok)
	}
	if _, ok := FleeDirectionFromKey('Z'); ok {
		t.Error("字母表裡沒有 Z")
	}
}
