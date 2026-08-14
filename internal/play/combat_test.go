package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

func mkParty(cons ...int16) *game.Party {
	p := &game.Party{}
	for i, c := range cons {
		p.Members = append(p.Members, &game.Character{
			Name: string(rune('A' + i)), CON: c, MaxCON: 30, AC: 2,
		})
	}
	return p
}

func mkScene(cons ...int16) *CombatScene {
	b := game.NewBattle(mkParty(cons...), rng.New())
	b.AddEnemy(0, 0, &game.Enemy{HP: 10})
	return NewCombatScene(b)
}

// 驗收 1：進戰鬥模式是名單，整隊逃跑之後切回地圖。
func TestScreenMode(t *testing.T) {
	s := mkScene(20, 20)
	if s.Mode != ModeRoster {
		t.Fatalf("進戰鬥應該是名單模式，得到 %d", s.Mode)
	}
	s.PartyFlees(game.FleeLeft)
	if s.Mode != ModeMap {
		t.Fatalf("整隊逃跑之後應該切回地圖，得到 %d", s.Mode)
	}
	if !s.Done() {
		t.Error("整隊逃跑之後指令階段應該結束")
	}
}

// 驗收 2：六個欄位落在對的欄座標；CON ≤ 0 印狀態字。
func TestRosterColumns(t *testing.T) {
	p := mkParty(25)
	rows := Roster(p)
	if len(rows) != 1 {
		t.Fatalf("應該有 1 行，得到 %d", len(rows))
	}
	line := rows[0].Text()
	if len(line) != rosterCols {
		t.Fatalf("一行應該是 %d 欄，得到 %d", rosterCols, len(line))
	}
	// 表頭的欄位起點與資料行要對得上。
	for _, tc := range []struct {
		col  int
		want string
	}{
		{colName, "A"}, {colAC, "2"}, {colMaxCON, "30"}, {colCON, "25"},
	} {
		if got := line[tc.col : tc.col+len(tc.want)]; got != tc.want {
			t.Errorf("欄 %#x：得到 %q，預期 %q", tc.col, got, tc.want)
		}
	}
	// 表頭釘死的是**順序**，不是對齊（docs/re/15 §4）。
	at := -1
	for _, label := range []string{"NAME", "AC", "AMM", "MAX", "CON", "WEAPON"} {
		i := strings.Index(RosterHeader, label)
		if i <= at {
			t.Fatalf("表頭的欄位順序不對，%q 出現在 %d：%q", label, i, RosterHeader)
		}
		at = i
	}

	// CON ≤ 0 的五個帶都要印狀態字，不能印數字。
	for _, tc := range []struct {
		con  int16
		want string
	}{{0, "UNC"}, {-5, "UNC"}, {-11, "SER"}, {-20, "CRT"}, {-30, "MRT"}, {-45, "COM"}} {
		got := Roster(mkParty(tc.con))[0].CON
		if got != tc.want {
			t.Errorf("CON %d：印 %q，預期 %q", tc.con, got, tc.want)
		}
	}
}

// 驗收 3：指令階段每問一個人，名單就是最新狀態。
func TestRosterReflectsCurrentState(t *testing.T) {
	s := mkScene(20, 20)
	s.Battle.Party.Members[1].CON = -25
	rows := Roster(s.Battle.Party)
	if rows[1].CON != "CRT" {
		t.Errorf("第二個人已經重傷，名單應該印 CRT，得到 %q", rows[1].CON)
	}
}

// 驗收 4：每個選項都帶熱鍵，而且字母與指令碼對得上。
func TestMenuKeysMatchCommandCodes(t *testing.T) {
	for _, o := range CommandMenu(nil) {
		cmd, ok := game.CommandFromKey(o.Key)
		if !ok || cmd != o.Cmd {
			t.Errorf("選項 %q 的熱鍵 %q 對到 %d，宣告的是 %d", o.Label, o.Key, cmd, o.Cmd)
		}
	}
	// 翻譯只換文字，熱鍵與指令碼不動。
	zh := map[game.Command]string{game.CmdRun: "逃跑", game.CmdAttack: "攻擊"}
	for _, o := range CommandMenu(func(c game.Command) string { return zh[c] }) {
		cmd, _ := game.CommandFromKey(o.Key)
		if cmd != o.Cmd {
			t.Errorf("翻譯之後熱鍵對錯了：%q → %d，應該是 %d", o.Key, cmd, o.Cmd)
		}
	}
	if got := CommandMenu(func(c game.Command) string { return zh[c] })[0].Label; got != "逃跑" {
		t.Errorf("第一個選項應該被翻成「逃跑」，得到 %q", got)
	}
}

// 驗收 5：一行一個選項（原版結構），每一行的第一個字元就是熱鍵。
func TestMenuIsOnePerLine(t *testing.T) {
	opts := CommandMenu(nil)
	lines := MenuLines(opts)
	if len(lines) != len(opts) {
		t.Fatalf("應該一行一個選項，%d 個選項得到 %d 行", len(opts), len(lines))
	}
	for i, l := range lines {
		if l[0] != opts[i].Key {
			t.Errorf("第 %d 行的第一個字元是 %q，熱鍵是 %q", i, l[0], opts[i].Key)
		}
		if !strings.Contains(l, opts[i].Label) {
			t.Errorf("第 %d 行掉了選項文字 %q：%q", i, opts[i].Label, l)
		}
	}

	// 中文選項也一樣：第一個字元仍然是熱鍵（\x10 捕捉的就是它）。
	zh := map[game.Command]string{
		game.CmdRun: "逃跑", game.CmdUse: "使用", game.CmdHire: "雇用",
		game.CmdEvade: "迴避", game.CmdAttack: "攻擊",
		game.CmdWeapon: "換武器", game.CmdLoad: "裝填",
	}
	zhLines := MenuLines(CommandMenu(func(c game.Command) string { return zh[c] }))
	for i, l := range zhLines {
		if l[0] != opts[i].Key {
			t.Errorf("中文第 %d 行的第一個字元是 %q，熱鍵應該還是 %q", i, l[0], opts[i].Key)
		}
	}

	// 標題 ＋ 空行 ＋ 七個選項 ＝ 9 行，訊息視窗只有 6 行——要回報放不下。
	if MenuFits(append([]string{"X, choose:", ""}, zhLines...), 38, 6) {
		t.Error("9 行塞不進 6 行的訊息視窗，應該回報 false")
	}
	// 38 格一行放得下任何一個中文選項（一格一個字，docs/spec/10 §3）。
	if !MenuFits(zhLines, 38, 7) {
		t.Error("七行 × 38 格應該放得下")
	}
}

// 驗收 5：一場指令階段跑得完，倒下的人不會被問到。
func TestCommandPhaseSkipsDownedMembers(t *testing.T) {
	s := mkScene(20, 0, 20) // 中間那個倒下了
	asked := []int{}
	for !s.Done() {
		asked = append(asked, s.Turn)
		if lines := s.Prompt(nil); len(lines) == 0 {
			t.Fatal("還沒問完卻拿不到提示")
		}
		if !s.Choose('E', true) {
			t.Fatalf("成員 %d 選迴避應該被接受", s.Turn)
		}
	}
	if len(asked) != 2 || asked[0] != 0 || asked[1] != 2 {
		t.Fatalf("應該只問成員 0 與 2，實際問了 %v", asked)
	}
	if s.Phase.Cmd[1] != game.CmdNone {
		t.Errorf("倒下的成員不該留下指令，得到 %d", s.Phase.Cmd[1])
	}
	if s.Phase.Defence(0) != s.Phase.Cmd[0].DefenceBase() {
		t.Error("選了迴避，門檻應該跟著變")
	}
}

// 沒有武器選攻擊要被要求重選，而且游標不前進。
func TestUnarmedAttackReasks(t *testing.T) {
	s := mkScene(20, 20)
	if s.Choose('A', false) {
		t.Fatal("沒武器選攻擊不該被接受")
	}
	if s.Turn != 0 {
		t.Fatalf("被要求重選時游標不該前進，得到 %d", s.Turn)
	}
	if len(s.Log) == 0 {
		t.Error("應該留下一條訊息")
	}
	if !s.Choose('A', true) {
		t.Fatal("有武器就該接受")
	}
	if s.Turn != 1 {
		t.Fatalf("接受之後應該換下一個人，得到 %d", s.Turn)
	}
}

// 不在字母表裡的按鍵一律不接受，游標也不動。
func TestUnknownKeyIsRejected(t *testing.T) {
	s := mkScene(20)
	for _, k := range []byte{'Z', '1', 0, ' '} {
		if s.Choose(k, true) {
			t.Errorf("按鍵 %q 不該被接受", k)
		}
		if s.Turn != 0 {
			t.Fatalf("被拒絕之後游標不該動，得到 %d", s.Turn)
		}
	}
}
