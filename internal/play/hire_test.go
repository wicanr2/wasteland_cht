package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/game/rng"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// hireScene 造一場「有一組敵人、地圖上有 NPC 記錄」的戰鬥。
//
// 遭遇記錄是**手造的**：這一組驗的是「拿到那個旗標之後會怎樣」，
// 把公式與分支釘在原地。**出貨資料裡真的有 14 筆帶旗標的遭遇**
// （`docs/re/114` §4），那一側由 `hiredata_test.go` 端到端守著。
func hireScene(t *testing.T, blockID int) *CombatScene {
	t.Helper()
	rom := openRomForPlay(t)
	blk, err := rom.Block(blockID)
	if err != nil {
		t.Skipf("開圖 %d：%v", blockID, err)
	}
	hirer := &game.Character{Name: "HELL RAZOR", Level: 5, CON: 20, MaxCON: 20}
	hirer.Attributes[game.AttrCharisma] = 18
	hirer.Attributes[game.AttrIQ] = 18
	p := &game.Party{Members: []*game.Character{hirer}}
	r := rng.New()
	w := game.NewWorld(blk, p, r)
	b := game.NewBattle(p, r)
	b.AddEnemy(0, 0, &game.Enemy{HP: 8, Data: game.EnemyData{Base: 8}})
	c := NewCombatScene(b)
	c.World = w
	rec := make([]byte, 0x10)
	rec[0x09] = 0x13 // bit1 設、NPC 記錄編號 1
	c.EncRecord = rec
	return c
}

// 雇用成功：NPC **整個人**加入隊伍，那一組從戰場上消失。
func TestHireJoinsTheParty(t *testing.T) {
	c := hireScene(t, 34) // 圖 34 的 section 17 記錄 1 ＝ REDHAWK
	before := len(c.World.Party.Members)
	c.Phase.Set(0, game.CmdHire, 0)
	out := c.resolveCommands()

	if got := len(c.World.Party.Members); got != before+1 {
		t.Fatalf("隊伍 %d 人，預期 %d：%v", got, before+1, out.EN)
	}
	joined := c.World.Party.Members[before]
	if joined.Name != "REDHAWK" {
		t.Errorf("加入的是 %q，預期 REDHAWK", joined.Name)
	}
	// **整筆記錄都要帶著**（存檔時要原樣寫回，`docs/re/110` §3）。
	if len(joined.Source) != 0x100 {
		t.Errorf("Source ＝ %d bytes，預期 256", len(joined.Source))
	}
	// 招募者的魅力寫進 +0x31。
	if joined.Source[0x31] != 18 {
		t.Errorf("+0x31 ＝ %d，預期 18", joined.Source[0x31])
	}
	if c.Battle.GroupAlive(0) {
		t.Error("那一組加入了隊伍，不該還留在戰場上")
	}
	if len(out.EN) == 0 || !strings.Contains(strings.Join(out.EN, " "), "REDHAWK") {
		t.Errorf("訊息裡沒有提到加入的人：%v", out.EN)
	}
}

// 名冊滿了要**在擲骰之前**就擋掉（`0x132CE`），而且說得出原因。
func TestHireStopsAtSevenMembers(t *testing.T) {
	c := hireScene(t, 34)
	for len(c.World.Party.Members) < game.HireCap {
		c.World.Party.Members = append(c.World.Party.Members,
			&game.Character{Name: "X", CON: 10, MaxCON: 10})
	}
	c.Phase.Set(0, game.CmdHire, 0)
	out := c.resolveCommands()
	if got := len(c.World.Party.Members); got != game.HireCap {
		t.Fatalf("滿了還加得進去：%d 人", got)
	}
	if !strings.Contains(strings.Join(out.EN, " "), "No room") {
		t.Errorf("沒有說「名冊沒有空位」：%v", out.EN)
	}
}

// 那一格不提供雇用時要說「試著雇用……但失敗了」，**不能靜靜地什麼都沒發生**。
func TestHireWithoutOfferSaysItFailed(t *testing.T) {
	c := hireScene(t, 34)
	c.EncRecord[0x09] = 0x01 // 出貨資料裡最常見的值：不能雇用
	before := len(c.World.Party.Members)
	c.Phase.Set(0, game.CmdHire, 0)
	out := c.resolveCommands()
	if len(c.World.Party.Members) != before {
		t.Fatal("不能雇用卻加了人")
	}
	if !strings.Contains(strings.Join(out.EN, " "), "tries to hire") {
		t.Errorf("沒有印失敗那一句：%v", out.EN)
	}
}

// 指令階段：按 `H` 要開「哪一組？」，選一組之後才輪到下一個人。
func TestHirePickerTakesAGroup(t *testing.T) {
	c := hireScene(t, 34)
	c.Turn = 0
	if !c.Choose('H', true) {
		t.Fatal("按 H 沒有開選單")
	}
	if !c.HirePicking() {
		t.Fatal("選單沒開")
	}
	if c.Turn != 0 {
		t.Fatal("選單開著時不該換人")
	}
	if !c.PickHire('1') {
		t.Fatal("選單沒吃掉數字鍵")
	}
	if c.HirePicking() {
		t.Error("選完了選單還開著")
	}
	if c.Phase.Cmd[0] != game.CmdHire {
		t.Errorf("指令 ＝ %v，預期 Hire", c.Phase.Cmd[0])
	}
	_ = input.KeyNone
}

// 雇用來的隊員要**存得下去**：配一個空的記錄編號、填進隊伍槽表、
// 整筆 256 bytes 原樣寫下去（`docs/re/110` §3）。
//
// ⚠ 少了槽表那一步的症狀是「這一局玩得到、存完就不見了」，
// 而**存檔當下不會有任何錯誤**。
func TestHiredMemberSurvivesSave(t *testing.T) {
	s := newScene(t)
	save := s.Save()
	if save == nil {
		t.Skip("沒有存檔")
	}
	before := len(s.World().Party.Members)

	rec := make([]byte, 0x100)
	copy(rec, "REDHAWK")
	rec[0x0E] = 12 // 力量
	rec[0x1B], rec[0x1D] = 40, 40
	rec[0x24] = 3    // 等級
	rec[0x77] = 0xAB // 未解區域
	joined := game.HireNPC(rec, 15)
	s.World().Party.Members = append(s.World().Party.Members, joined)

	if err := s.StoreTo(save); err != nil {
		t.Fatalf("寫存檔：%v", err)
	}

	// 槽表要多一格，而且指到一筆真的寫過的記錄。
	g := save.SlotGroups()[0]
	id := g.Members[before]
	if id == 0 {
		t.Fatal("隊伍槽表沒有補上新隊員")
	}
	raw, err := save.Record(int(id))
	if err != nil {
		t.Fatal(err)
	}
	if got := game.LoadCharacter(raw); got.Name != "REDHAWK" {
		t.Errorf("記錄 %d 的名字 ＝ %q，預期 REDHAWK", id, got.Name)
	}
	if raw[0x77] != 0xAB {
		t.Errorf("未解區域 ＝ %#x，預期原樣寫下去（0xAB）", raw[0x77])
	}
	if raw[0x31] != 15 {
		t.Errorf("+0x31 ＝ %d，預期招募者的魅力 15", raw[0x31])
	}
}
