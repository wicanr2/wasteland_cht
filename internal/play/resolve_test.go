package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 戰鬥的七個指令**按下去要有事發生**（`docs/re/107`）。
//
// ⚠ 這一條擋的是一種特別安靜的漏接：下令階段的處理程式寫好了、也有單元測試，
// 但 `internal/play` 沒有人叫它們，於是玩家按 `W`／`L` 什麼都不會發生，
// **畫面上完全看不出異狀**。接線表也擋不住——它只數「有沒有人引用那份筆記」。
func TestWeaponCommandActuallyChangesEquipment(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	c := s.combat
	m := c.Battle.Party.Members[c.Turn]
	if m == nil || len(m.Items) == 0 {
		t.Skip("第一個人不能行動")
	}
	before := m.EquipIndex

	// `W` 要開清單——不開的話參數留 0，結算階段什麼都不做。
	step(t, s, input.Input{Dir: input.DirNone, Char: 'W'})
	if !c.WeaponPicking() {
		t.Fatal("按 W 沒有開清單")
	}
	// 選第一項。
	step(t, s, input.Input{Dir: input.DirNone, Char: '1'})
	if c.WeaponPicking() {
		t.Fatal("選完之後清單還開著")
	}
	if c.Phase.Cmd[0] != game.CmdWeapon {
		t.Fatalf("指令碼應該是換武器，得到 %d", c.Phase.Cmd[0])
	}
	if c.Phase.Arg[0] == 0 {
		t.Fatal("參數留 0：結算階段會什麼都不做，而畫面上看不出來")
	}

	// 其餘的人都下攻擊令，把這一回合跑完。
	for !c.Done() {
		step(t, s, input.Input{Dir: input.DirNone, Char: 'A'})
	}
	if m.EquipIndex == before {
		t.Errorf("跑完一回合裝備沒變（還是 %d）——結算階段沒有接上", before)
	}
}

// 清單開著時**數字鍵歸清單**，不能被當成指令熱鍵。
func TestWeaponPickSwallowsKeys(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	c := s.combat
	turn := c.Turn
	step(t, s, input.Input{Dir: input.DirNone, Char: 'W'})
	if !c.WeaponPicking() {
		t.Skip("這個人身上沒東西")
	}
	// `A` 在指令選單是「攻擊」；清單開著的時候它什麼都不該做。
	step(t, s, input.Input{Dir: input.DirNone, Char: 'A'})
	if !c.WeaponPicking() {
		t.Error("清單被別的按鍵關掉了")
	}
	if c.Turn != turn {
		t.Error("清單開著卻換人了")
	}
	// ESC 收清單、回到指令選單（原版回 0xFF ＝ 取消，重問這個人）。
	step(t, s, input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	if c.WeaponPicking() {
		t.Error("ESC 沒有收掉清單")
	}
	if c.Turn != turn {
		t.Error("取消之後應該重問同一個人")
	}
}

// 清單開著的時候，**面板上看得到的是清單**。
//
// ⚠ 原版不清面板，新的文字接在後面、滿了就捲（`docs/re/106`），
// 所以指令選單留在訊息字串裡是**對的**——要驗的是「捲完之後看得到什麼」，
// 不是「字串裡有沒有」。
func TestWeaponPickIsWhatYouSee(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'W'})
	if !s.combat.WeaponPicking() {
		t.Skip("這個人身上沒東西")
	}
	msg := s.Message()
	if msg == "" {
		t.Skip("這一輪走的是中文那條路")
	}
	rect := s.msgRect()
	rows := map[int]string{}
	eachMessageCell(msg, rect, rect.Row, func(col, row int, r rune) {
		rows[row] += string(r)
	})
	var body []string
	for r := rect.Row; r <= rect.LastRow(); r++ {
		body = append(body, rows[r])
	}
	seen := strings.Join(body, "\n")
	if !strings.Contains(seen, "which item?") {
		t.Errorf("面板上看不到清單標題：\n%s", seen)
	}
	// 最後一列一定要是清單的最後一行——它是玩家要按的那幾項。
	last := strings.TrimSpace(body[len(body)-1])
	if last == "" || strings.HasSuffix(last, "Load/unjam") {
		t.Errorf("面板最後一列是 %q，清單被擠出去了：\n%s", last, seen)
	}
}

// 裝填在結算階段真的填（`docs/re/107` §3）。
func TestLoadCommandActuallyReloads(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	c := s.combat
	// 找一個「裝著吃彈匣的武器、而且身上有彈匣」的人。
	idx, target := -1, (*game.Character)(nil)
	for i, m := range c.Battle.Party.Members {
		if m == nil || m.Down() {
			continue
		}
		slot, ok := equippedSlot(m)
		if !ok {
			continue
		}
		d, ok := c.Items.Get(slot.ID)
		if !ok || d.Ammo == 0 {
			continue
		}
		if _, found := game.FindItem(itemIDs(m), d.Ammo); found {
			idx, target = i, m
			break
		}
	}
	if target == nil {
		t.Skip("出廠隊伍裡沒有人裝著吃彈匣的武器")
	}
	// 先把彈數打掉，才看得出「填滿」。
	w, _ := equippedSlot(target)
	target.Items[target.EquipIndex-1].Value = w.Value &^ 0x3F

	for !c.Done() {
		ch := byte('A')
		if c.Turn == idx {
			ch = 'L'
		}
		step(t, s, input.Input{Dir: input.DirNone, Char: ch})
	}
	if got := target.Items[target.EquipIndex-1].Value & 0x3F; got == 0 {
		t.Error("跑完一回合彈數還是 0——裝填沒有接上")
	}
}

// itemIDs 把物品陣列攤成 ID 陣列（`game.FindItem` 吃的形狀）。
func itemIDs(c *game.Character) []byte {
	out := make([]byte, len(c.Items))
	for i := range c.Items {
		out[i] = c.Items[i].ID
	}
	return out
}

// 名單那一欄與傷害計算**一定要算出同一件武器**（`docs/re/103` §3 的 1-based）。
//
// ⚠ 兩邊各自算索引的時候，畫面上完全看不出來：出廠的 Angela Deth
// 名單印 `VP91Z 9mm`，而攻擊拿的是她的**彈匣**（Dice ＝ 0，等於白打）。
func TestWeaponColumnAndDamageAgree(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	c := s.combat
	n := 0
	for _, m := range c.Battle.Party.Members {
		if m == nil || m.EquipIndex == 0 {
			continue
		}
		slot, ok := equippedSlot(m)
		if !ok {
			t.Errorf("%s 有 EquipIndex %d 卻取不到那一格", m.Name, m.EquipIndex)
			continue
		}
		want, _ := c.Items.Get(slot.ID)
		if got := c.weaponOf(m); got != want {
			t.Errorf("%s：名單那一欄是物品 %d，傷害用的卻是別的（Dice %d vs %d）",
				m.Name, slot.ID, got.Dice, want.Dice)
		}
		n++
	}
	if n == 0 {
		t.Skip("沒有人裝備武器")
	}
}

// 結算階段每一句話**都要有中文**（`docs/re/98` §2 的三層：目錄有譯文 →
// 消費端查得到 → 控制碼解得開）。
//
// ⚠ 覆蓋率測試只量第一層。這一條量第二、三層——三層任何一層斷了，
// 症狀都一樣是「畫面上是英文」。
func TestResolveMessagesHaveChinese(t *testing.T) {
	s := newScene(t)
	// ⚠ `newScene` **不載目錄**——不載的話 `c.CJK` 仍然是一支非 nil 的函式，
	// 只是每次都回 nil。用「函式在不在」當判準會讓這條測試永遠綠。
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		t.Skipf("沒有翻譯目錄：%v", err)
	}
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	c := s.combat
	m := c.Battle.Party.Members[0]
	for _, tc := range []struct {
		name string
		res  game.ResolveResult
	}{
		{"換裝備", game.ResolveResult{Message: game.MsgSwapsEquipment}},
		{"迴避", game.ResolveResult{Message: game.MsgEvades}},
		{"裝填", game.ResolveResult{Message: game.MsgReloads}},
		{"沒彈匣", game.ResolveResult{Message: game.MsgNoMoreClips}},
		{"不能裝填", game.ResolveResult{Message: game.MsgCantBeReloaded}},
		{"卡彈", game.ResolveResult{Message: game.MsgWeaponJammed, Table2: true}},
	} {
		if zh := c.zhResolve(m, tc.res); len(zh) == 0 {
			t.Errorf("%s（字串 %d）查不到中文", tc.name, tc.res.Message)
		}
		if en := resolveText(m.Name, tc.res); en == "" {
			t.Errorf("%s（字串 %d）沒有英文那一句", tc.name, tc.res.Message)
		}
	}
}

// 指令被打回票時那句話要**進訊息區**，不是只留在 `Log` 裡。
//
// ⚠ 只 append 到 `Log` 等於沒說：畫面上只會看到選單又出現一次，
// 玩家不知道自己按的那個鍵怎麼了。
func TestRejectedCommandSaysWhy(t *testing.T) {
	s := newScene(t)
	step(t, s, input.Input{Dir: input.DirNone, Char: 'E'})
	step(t, s, input.Input{Dir: input.DirNone, Char: 'Y'})
	if !s.InCombat() {
		t.Skip("這一格開不了戰鬥")
	}
	turn := s.combat.Turn
	step(t, s, input.Input{Dir: input.DirNone, Char: 'H'}) // 還沒做的指令
	if s.combat.Turn != turn {
		t.Fatal("打回票應該重問同一個人")
	}
	if len(s.cjk) > 0 {
		// 中文那條路：訊息接在提示前面。
		if !strings.Contains(s.cjk, s.uiText("combat.notyet")) {
			t.Error("中文訊息沒有出現在面板上")
		}
		return
	}
	if !strings.Contains(s.Message(), "not implemented") {
		t.Errorf("英文訊息沒有出現在面板上：%q", s.Message())
	}
}
