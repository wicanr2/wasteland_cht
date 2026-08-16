package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
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
	rows := Roster(p, nil, nil)
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

	// CON ≤ 0 的六個帶都要印狀態字，不能印數字。
	//
	// ⚠ **CON ＝ 0 是死亡（等級 5），不是昏迷**——那一格印的是原版主文字
	// 字型第 `0x7F` 格的骷髏字模（`docs/re/17` §4.4）。昏迷（`UNC`）是等級 0，
	// 也就是 CON 在 −1…−10 之間、還救得回來的那一段（`docs/re/99` §3.1）。
	for _, tc := range []struct {
		con  int16
		want string
	}{{0, game.WoundDead}, {-5, "UNC"}, {-10, "UNC"}, {-11, "SER"}, {-19, "SER"},
		{-20, "CRT"}, {-29, "CRT"}, {-30, "MRT"}, {-39, "MRT"}, {-40, "COM"}, {-45, "COM"}} {
		got := Roster(mkParty(tc.con), nil, nil)[0].CON
		if got != tc.want {
			t.Errorf("CON %d：印 %q，預期 %q", tc.con, got, tc.want)
		}
	}
}

// 驗收 3：指令階段每問一個人，名單就是最新狀態。
func TestRosterReflectsCurrentState(t *testing.T) {
	s := mkScene(20, 20)
	s.Battle.Party.Members[1].CON = -25
	rows := Roster(s.Battle.Party, nil, nil)
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

// 遭遇迴圈（docs/spec/21 §5 的驗收條件）。

func openRomForPlay(t *testing.T) *assets.Rom {
	t.Helper()
	rom, err := assets.Open("../../workplace/orig/wastland")
	if err != nil {
		t.Skipf("找不到原版資料：%v", err)
	}
	if err := rom.LoadImage("../../workplace/analysis/unpacked/wl.merged.exe"); err != nil {
		t.Skipf("載入分析映像失敗：%v", err)
	}
	return rom
}

func openScene(t *testing.T) *Scene {
	t.Helper()
	rom := openRomForPlay(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場景：%v", err)
	}
	return s
}

// 驗收 1：視窗裡沒有遭遇格就不會進戰鬥。
func TestNoEncounterCellNoCombat(t *testing.T) {
	s := openScene(t)
	w := s.World()
	// 出廠位置（Ranger Ctr. 外面）掃一次；有沒有遭遇格由資料決定，
	// 所以這裡驗的是「掃到的格數 ＝ 0 就一定回 nil」這條蘊含。
	var groups [game.QueueGroups]game.PartyGroupState
	groups[0] = game.PartyGroupState{Present: true, Engage: game.EngageFar,
		X: int(s.World().Party.X), Y: int(s.World().Party.Y)}
	scan := w.ScanEncounters(groups)
	c, err := s.StartEncounter()
	if err != nil {
		t.Fatalf("StartEncounter：%v", err)
	}
	if _, ok := scan.Queue.Nearest(0); !ok && c != nil {
		t.Error("佇列是空的，卻進了戰鬥")
	}
	if s.InCombat() != (c != nil) {
		t.Error("InCombat 與回傳值不一致")
	}
}

// 驗收 3：經驗值用前後相減算，一個角色一筆。
func TestEncounterXPIsBeforeAfterDelta(t *testing.T) {
	s := openScene(t)
	if len(s.World().Party.Members) == 0 {
		t.Skip("出廠存檔沒有隊員")
	}
	s.snapshot = s.takeXP()
	s.combat = &CombatScene{}

	m := s.World().Party.Members[0]
	before := m.XP
	m.XP += 137

	res := s.FinishEncounter()
	if got := res.XPGained[m.Name]; got != 137 {
		t.Errorf("%s 應該拿到 137 點經驗值，得到 %d", m.Name, got)
	}
	if len(res.XPGained) != 1 {
		t.Errorf("只有一個人拿到經驗值，卻報了 %d 筆", len(res.XPGained))
	}
	if m.XP != before+137 {
		t.Error("收尾不該動到角色的經驗值")
	}
	if s.InCombat() {
		t.Error("收尾之後應該不在戰鬥裡")
	}
}

// 驗收 4：打完回得去——座標、時鐘、地圖都不變。
func TestEncounterKeepsWorldState(t *testing.T) {
	s := openScene(t)
	w := s.World()
	x, y, clock, dim := w.Party.X, w.Party.Y, w.Clock, w.Block.Dim

	s.snapshot = s.takeXP()
	s.combat = &CombatScene{}
	s.FinishEncounter()

	if w.Party.X != x || w.Party.Y != y {
		t.Errorf("座標被動到了：(%d, %d) → (%d, %d)", x, y, w.Party.X, w.Party.Y)
	}
	if w.Clock != clock {
		t.Errorf("時鐘被動到了：%v → %v", clock, w.Clock)
	}
	if w.Block.Dim != dim {
		t.Error("地圖被換掉了")
	}
}

// 驗收 5：全滅要分辨得出來，不能靜悄悄回地圖。
func TestEncounterReportsWipe(t *testing.T) {
	s := openScene(t)
	if len(s.World().Party.Members) == 0 {
		t.Skip("出廠存檔沒有隊員")
	}
	s.snapshot = s.takeXP()
	s.combat = &CombatScene{}
	for _, m := range s.World().Party.Members {
		m.CON = 0
	}
	if res := s.FinishEncounter(); !res.Wiped {
		t.Error("全隊 CON ＝ 0 應該回報全滅")
	}
}

// 端到端：把隊伍擺到遭遇格旁邊，戰鬥要真的生得出來。
//
// 這條與上面那幾條的分工：上面驗的是「接合的規則對不對」，
// 這一條驗的是**拿原版資料真的跑得起來**——測試全綠不等於接得起來。
func TestEncounterSpawnsFromRealMap(t *testing.T) {
	s := openScene(t)
	w := s.World()

	// ⚠ **出廠那張地圖（資源 0）一格靜態遭遇格都沒有。** 42 張地圖合計
	// 365 個 nibble 3、**0 個 nibble 15**——nibble 15 是執行時才長出來的
	// （`sub_16890` 生成，docs/spec/13）。所以這條測試要自己換一張有的地圖，
	// 不能拿出廠位置去試，否則會永遠 skip 而看起來像通過。
	var fx, fy, found = -1, -1, -1
	for id := 0; id < 50 && found < 0; id++ {
		b, err := s.rom.Block(id)
		if err != nil {
			continue
		}
		for y := 0; y < b.Dim && found < 0; y++ {
			for x := 0; x < b.Dim; x++ {
				terrain, _, _, err := b.At(x, y)
				if err == nil && game.IsEncounterCell(terrain) {
					fx, fy, found = x, y, id
					w.EnterMap(b, uint8(x), uint8(y))
					break
				}
			}
		}
	}
	if found < 0 {
		t.Fatal("42 張地圖裡一格遭遇格都找不到——掃描或資料解讀有問題")
	}

	w.Teleport(uint8(fx), uint8(fy))
	c, err := s.StartEncounter()
	if err != nil {
		t.Fatalf("StartEncounter：%v", err)
	}
	if c == nil {
		t.Fatalf("站在遭遇格 (%d, %d) 上卻沒有戰鬥", fx, fy)
	}
	n := 0
	for _, e := range c.Battle.Enemies {
		if e != nil {
			n++
		}
	}
	if n == 0 {
		t.Error("戰鬥開了，但一個敵人都沒有")
	}
	t.Logf("遭遇格 (%d, %d) 生出 %d 個敵人", fx, fy, n)
}

// 回合結算（docs/spec/22 §6 的驗收條件）。

func mkBattle(t *testing.T, enemyHP uint16, ac byte) *CombatScene {
	t.Helper()
	p := mkParty(30, 30)
	for _, m := range p.Members {
		// 屬性放在死區（9–13），AttrModifier 不加不減——
		// 讓測試只驗流程，不被屬性修正拖著跑。
		for i := range m.Attributes {
			m.Attributes[i] = 11
		}
		m.AC = ac
	}
	b := game.NewBattle(p, rng.New())
	b.AddEnemy(0, 0, &game.Enemy{HP: enemyHP, Data: game.EnemyData{Base: 20, DiceN: 1, XPMul: 0}})
	s := NewCombatScene(b)
	// 給一把武器。**沒有武器每個人的傷害都是 0**，戰鬥永遠打不完——
	// 這一點測試裡要顯式給，不要靠零值碰運氣。
	s.Items = game.ItemTable{{}, {Dice: 3, Class: 1}}
	for _, m := range p.Members {
		m.EquipIndex = 1
	}
	return s
}

// 驗收 4：敵人死光 → Over && Won。
func TestRoundEndsWhenEnemiesDie(t *testing.T) {
	s := mkBattle(t, 1, 0)
	for i := range s.Phase.Cmd {
		s.Phase.Cmd[i] = game.CmdAttack
	}
	// 直接把敵人打死，驗的是「結束判定」而不是命中率。
	s.Battle.Enemy(0).HP = 0
	res := s.ResolveRound()
	if !res.Over || !res.Won {
		t.Errorf("敵人都死了應該是 Over && Won，得到 over=%v won=%v", res.Over, res.Won)
	}
}

// 驗收 4：隊伍全倒 → Over && !Won。
func TestRoundEndsWhenPartyFalls(t *testing.T) {
	s := mkBattle(t, 100, 0)
	for _, m := range s.Battle.Party.Members {
		m.CON = 0
	}
	res := s.ResolveRound()
	if !res.Over || res.Won {
		t.Errorf("隊伍全倒應該是 Over && !Won，得到 over=%v won=%v", res.Over, res.Won)
	}
}

// 驗收 1：行動順序與 Battle.Order() 一致（同一顆種子跑兩次結果相同）。
func TestRoundOrderIsStable(t *testing.T) {
	one := mkBattle(t, 100, 0)
	one.Battle.RNG = rng.New()
	// **只有下攻擊令的隊員才排進行動表**（docs/re/90 §2），
	// 所以要先下令——不下令的話行動表裡只會有敵人。
	for i := range one.Phase.Cmd {
		one.Phase.Cmd[i] = game.CmdAttack
	}
	one.ResolveRound()
	got := len(one.Battle.Order())
	if got == 0 {
		t.Fatal("一回合結束後行動表是空的——BeginRound 沒排進去")
	}
	// 兩個隊員 ＋ 一個敵人 ＝ 三個行動者。
	if got != 3 {
		t.Errorf("行動表應該有 3 個單位，得到 %d", got)
	}

	// 反向：沒人下攻擊令時，行動表裡只剩敵人。
	two := mkBattle(t, 100, 0)
	two.Battle.RNG = rng.New()
	two.ResolveRound()
	if n := len(two.Battle.Order()); n != 1 {
		t.Errorf("沒下令時只有 1 個敵人該排進去，得到 %d", n)
	}
}

// 驗收 3：打死敵人給的經驗值等於 KillXP()。
func TestRoundAwardsKillXP(t *testing.T) {
	s := mkBattle(t, 1, 0)
	m := s.Battle.Party.Members[0]
	before := m.XP
	e := s.Battle.Enemy(0)
	want := e.Data.KillXP()

	// 直接跑攻擊那一段，避開命中率的隨機性：連打到死為止。
	for i := 0; i < 200 && e.HP > 0; i++ {
		s.Phase.Cmd[0] = game.CmdAttack
		s.partyActs(game.Combatant{Slot: game.EnemySlots, IsParty: true})
	}
	if e.HP > 0 {
		t.Skip("200 次都沒打中——命中率暫代路徑，換一輪再測")
	}
	if got := m.XP - before; got != want {
		t.Errorf("擊殺經驗值應該是 %d，得到 %d", want, got)
	}
}

// 端到端：拿原版資料開一場戰鬥，跑到結束。
//
// 上面那幾條驗規則，這一條驗**接得起來**——物品表載得到、武器有骰數、
// 回合跑得完。少任何一環症狀都是「戰鬥打不完」，而單元測試不會紅。
func TestRealBattleTerminates(t *testing.T) {
	s := openScene(t)
	w := s.World()

	var found = -1
	for id := 0; id < 50 && found < 0; id++ {
		b, err := s.rom.Block(id)
		if err != nil {
			continue
		}
		for y := 0; y < b.Dim && found < 0; y++ {
			for x := 0; x < b.Dim; x++ {
				if terrain, _, _, err := b.At(x, y); err == nil && game.IsEncounterCell(terrain) {
					w.EnterMap(b, uint8(x), uint8(y))
					found = id
					break
				}
			}
		}
	}
	if found < 0 {
		t.Fatal("找不到遭遇格")
	}

	c, err := s.StartEncounter()
	if err != nil || c == nil {
		t.Fatalf("開不了戰鬥：c=%v err=%v", c, err)
	}
	if len(c.Items) == 0 {
		t.Fatal("物品表沒載到——武器傷害會是 0，戰鬥永遠打不完")
	}

	rounds := 0
	for rounds < game.MaxRounds {
		for i := range c.Phase.Cmd {
			c.Phase.Cmd[i] = game.CmdAttack
		}
		res := c.ResolveRound()
		rounds++
		if res.Over {
			t.Logf("地圖 %d 的戰鬥在第 %d 回合結束，隊伍%s", found, rounds,
				map[bool]string{true: "贏", false: "輸"}[res.Won])
			return
		}
		c.BeginCommands()
	}
	t.Errorf("跑了 %d 回合還沒結束——回合迴圈收斂不了", rounds)
}
