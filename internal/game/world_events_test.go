package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// 驗收 1：42 張地圖的 section 3／5／15／16／17 都解得開，記錄落在區塊內。
func TestSectionRecordsResolve(t *testing.T) {
	rom := openRom(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatalf("列資源失敗：%v", err)
	}
	checked := 0
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatalf("區塊 %d：%v", i, err)
		}
		for _, typ := range []int{3, 5, 15, 16, 17} {
			n := b.SectionCount(typ)
			if n == 0 {
				continue
			}
			for idx := 0; idx < n; idx++ {
				if _, err := b.SectionRecord(typ, idx); err != nil {
					// 指標陣列裡可以有空洞（值為 0），那不是錯。
					continue
				}
				checked++
			}
		}
	}
	if checked == 0 {
		t.Fatal("一筆 section 記錄都沒解出來，測試沒有真的驗到東西")
	}
	t.Logf("解出 %d 筆 section 記錄", checked)
}

// 驗收 2：所有 nibble 2 的格子，條件串列都解析得開且值在範圍內。
func TestGatesParseAcrossAllMaps(t *testing.T) {
	rom := openRom(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatalf("列資源失敗：%v", err)
	}
	total, cells := 0, 0
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatalf("區塊 %d：%v", i, err)
		}
		for idx, terrain := range b.Terrain {
			if terrain != 2 {
				continue
			}
			// section 型別 ＝ 第 1 層的 nibble 本身（sub_169EB）。
			rec, err := b.SectionRecord(2, int(b.Record[idx]))
			if err != nil {
				continue // 這一格指到的記錄取不到，留給 §7 的未解
			}
			cells++
			for _, g := range ParseGates(rec) {
				if g.Type < 0 || g.Type > 7 {
					t.Fatalf("區塊 %d 的條件型別 %d 超出 0–7", i, g.Type)
				}
				if g.Difficulty < 0 || g.Difficulty > 31 {
					t.Fatalf("區塊 %d 的難度 %d 超出 0–31", i, g.Difficulty)
				}
				total++
			}
		}
	}
	if cells == 0 {
		t.Fatal("42 張地圖裡一個 nibble 2 的格子都沒有，掃描一定有問題")
	}
	t.Logf("%d 個格子、%d 條條件全部解析成功", cells, total)
}

// 驗收 3：門檻公式與骰的邊界。
func TestThresholdAndFumble(t *testing.T) {
	for _, tc := range []struct{ diff, want int }{{0, 15}, {1, 20}, {10, 65}, {31, 170}} {
		if got := Threshold(tc.diff); got != tc.want {
			t.Errorf("難度 %d 的門檻應該是 %d，得到 %d", tc.diff, tc.want, got)
		}
	}

	// 骰 < 5 一定失敗，不管屬性多高。
	r := rng.New()
	c := &Character{Level: 5}
	c.Attributes[AttrLuck] = 18
	fumbles, total := 0, 0
	for i := 0; i < 5000; i++ {
		res := c.AttributeCheck(r, recAttributes+AttrLuck, 0, true)
		if res.Roll > 0 && res.Roll < fumble {
			fumbles++
			if res.OK {
				t.Fatalf("骰 %d 小於 %d 卻成功了", res.Roll, fumble)
			}
		}
		total++
	}
	if fumbles == 0 {
		t.Fatal("5000 次檢定一次都沒骰出 < 5，分佈一定有問題")
	}
	t.Logf("%d/%d 次骰到 < 5", fumbles, total)
}

// 性別與等級檢定不擲骰。
func TestAttributeCheckSpecialCases(t *testing.T) {
	r := rng.New()
	c := &Character{Gender: 1, Level: 7}
	if res := c.AttributeCheck(r, recGender, 1, true); !res.OK || res.Roll != 0 {
		t.Fatalf("性別檢定應該直接比相等且不擲骰：%+v", res)
	}
	if res := c.AttributeCheck(r, recGender, 0, true); res.OK {
		t.Fatal("性別 1 不該通過「等於 0」的檢定")
	}
	if res := c.AttributeCheck(r, recLevel, 7, true); !res.OK || res.Roll != 0 {
		t.Fatalf("等級檢定應該比大小且不擲骰：%+v", res)
	}
	if res := c.AttributeCheck(r, recLevel, 8, true); res.OK {
		t.Fatal("等級 7 不該通過「≥ 8」的檢定")
	}
}

// 驗收 4：技能等級永遠不超過角色等級。
func TestTrainNeverExceedsCharacterLevel(t *testing.T) {
	r := rng.New()
	c := &Character{Level: 3, Skills: []Slot{{ID: 1, Value: 1}}}
	for i := 0; i < 20000; i++ {
		c.TrainSkill(r, 1, 31)
		if int(c.SkillLevel(1)) > int(c.Level) {
			t.Fatalf("技能等級 %d 超過角色等級 %d", c.SkillLevel(1), c.Level)
		}
	}
	if c.SkillLevel(1) != c.Level {
		t.Fatalf("練了兩萬次應該頂到角色等級 %d，得到 %d", c.Level, c.SkillLevel(1))
	}
}

// 驗收 5：42 張地圖裡 nibble 6 的記錄，opcode 都在 0–43。
func TestScriptOpcodesInRange(t *testing.T) {
	rom := openRom(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatalf("列資源失敗：%v", err)
	}
	seen := map[int]int{}
	cells := 0
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatalf("區塊 %d：%v", i, err)
		}
		for idx, terrain := range b.Terrain {
			if terrain != 6 {
				continue
			}
			rec, err := b.SectionRecord(6, int(b.Record[idx]))
			if err != nil {
				continue
			}
			cells++
			sel := rec[0]
			if sel&0x80 != 0 {
				// bit7 ＝ 設施畫面，索引 & 0x7F 進另一張 5 筆的表。
				if int(sel&0x7F) >= 5 {
					t.Errorf("區塊 %d 的設施索引 %d 超出 0–4", i, sel&0x7F)
				}
				continue
			}
			// bit7 ＝ 0：再經 section 型別 0x10 的陣列換成 opcode，
			// 陣列裡存的是 opcode 本身不是位移（docs/re/34 §1）。
			op, err := b.SectionEntry(0x10, int(sel))
			if err != nil {
				continue
			}
			seen[int(op)]++
		}
	}
	t.Logf("%d 個 nibble 6 的格子，用到 %d 種不同的值", cells, len(seen))
	for op := range seen {
		if op >= OpCount {
			t.Errorf("opcode %d 超出 0–%d", op, OpCount-1)
		}
	}
}

// 物品消耗：低 6 位是次數，用完把那一槽清成 0，高 2 位保留。
//
// **清空不是移除**：物品陣列是固定 30 槽、0 ＝ 空（`docs/re/15`），
// 賣掉一件也只是把那兩個 byte 清成 0（`docs/re/42` §3）。
// 後面那一槽的**槽號不能變**——角色記錄 `+0x1F`／`+0x25` 存的就是槽號。
func TestUseItemConsumes(t *testing.T) {
	c := &Character{Items: []Slot{{ID: 9, Value: 0xC3}, {ID: 7, Value: 1}}}
	for want := byte(2); want >= 1; want-- {
		if !c.UseItem(9) {
			t.Fatal("應該找得到這件物品")
		}
		if got := c.Items[0].Value; got != 0xC0|want {
			t.Fatalf("用一次之後應該是 %#02x，得到 %#02x", 0xC0|want, got)
		}
	}
	if !c.UseItem(9) {
		t.Fatal("最後一次應該還是找得到")
	}
	if c.Items[0] != (Slot{}) {
		t.Fatalf("次數用完那一槽應該清成 0，得到 %v", c.Items[0])
	}
	if c.Items[1] != (Slot{ID: 7, Value: 1}) {
		t.Fatalf("後面那一槽不該被搬動，得到 %v", c.Items[1])
	}
	if c.UseItem(9) {
		t.Fatal("沒有這件物品了不該回 true")
	}
}

// 腳本：遭遇控制那一組真的會改標頭，而且守住上下限。
func TestScriptEncounterControl(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(0)
	if err != nil {
		t.Fatalf("載入區塊 0 失敗：%v", err)
	}
	p := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}, X: 55, Y: 62}
	w := NewWorld(block, p, rng.New())

	block.Header[hdrEncounterDenom] = 50
	run := func(op int) ScriptResult {
		s := &Script{World: w, Record: []byte{byte(op), 0, 0, 0, 0, 0, 0, 0}, Op: op}
		return s.Step()
	}
	if r := run(OpDenomSet0); !r.Handled || block.Header[hdrEncounterDenom] != 0 {
		t.Fatalf("op 40 應該把分母設成 0，得到 %d", block.Header[hdrEncounterDenom])
	}
	if r := run(OpDenomSet100); !r.Handled || block.Header[hdrEncounterDenom] != 100 {
		t.Fatalf("op 42 應該把分母設成 100，得到 %d", block.Header[hdrEncounterDenom])
	}
	// 分母 0 的時候不擲遭遇。
	block.Header[hdrEncounterDenom] = 0
	if w.rollEncounter() {
		t.Fatal("分母為 0 不該擲出遭遇")
	}
	// 上限守得住。
	block.Header[hdrEncounterKinds] = 0x0B
	run(OpKindsInc11)
	if block.Header[hdrEncounterKinds] != 0x0B {
		t.Fatalf("op 23 的上限是 0x0B，卻變成 %#x", block.Header[hdrEncounterKinds])
	}
	// 中止指令回報不要繼續。
	if r := run(OpAbort); r.Continue {
		t.Fatal("op 31 應該停住腳本")
	}
	// 未實作的指令要明講，不能假裝成 nop。
	if r := run(OpOverlay); r.Handled {
		t.Fatal("op 2 還沒實作，不該回報 Handled")
	}
}

// 腳本的晝夜分支換的是「下一步」。
func TestScriptDayNightBranch(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(0)
	if err != nil {
		t.Fatalf("載入區塊 0 失敗：%v", err)
	}
	p := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}}
	w := NewWorld(block, p, rng.New())

	rec := []byte{OpDayNight, 0, 0, 0xAA, 0xBB, 0xCC, 0xDD, 0}
	w.Clock.Hour = 12
	(&Script{World: w, Record: rec, Op: OpDayNight}).Step()
	if rec[1] != 0xAA || rec[2] != 0xBB {
		t.Fatalf("白天應該走 +0x03，得到 %#x %#x", rec[1], rec[2])
	}
	w.Clock.Hour = 22
	(&Script{World: w, Record: rec, Op: OpDayNight}).Step()
	if rec[1] != 0xCC || rec[2] != 0xDD {
		t.Fatalf("夜間應該走 +0x05，得到 %#x %#x", rec[1], rec[2])
	}
}

var _ = assets.SectionTypes

// ds:916Bh：只有玩家自己走出來的那一步之後，檢定成功才給經驗值
// （docs/re/32 §7.1）。自動步不給——這是原版的防刷。
func TestCheckXPFollowsPlayerStep(t *testing.T) {
	r := rng.New()
	award := func(awardXP bool) uint32 {
		c := &Character{Level: 5}
		c.Attributes[AttrLuck] = 18
		for i := 0; i < 200; i++ {
			c.AttributeCheck(r, recAttributes+AttrLuck, 0, awardXP)
		}
		return c.XP
	}
	if got := award(true); got == 0 {
		t.Fatal("玩家走過一步之後，兩百次檢定應該累積到經驗值")
	}
	if got := award(false); got != 0 {
		t.Fatalf("自動步之後不該給經驗值，卻拿到 %d", got)
	}
}

// 走一步要把「這一步是玩家走的」記起來（ds:916Bh ← 方向 − 4）。
func TestStepMarksPlayerStepped(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(0)
	if err != nil {
		t.Fatalf("載入區塊 0 失敗：%v", err)
	}
	party := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}, X: 55, Y: 62}
	w := NewWorld(block, party, rng.New())
	if w.Party.PlayerStepped {
		t.Fatal("還沒走就不該是 true")
	}
	for dir := Up; dir <= Right; dir++ {
		if res, err := w.Step(dir); err == nil && res.Moved {
			if !w.Party.PlayerStepped {
				t.Fatal("走成一步之後應該記成玩家步")
			}
			return
		}
	}
	t.Skip("四個方向都走不動，換一張地圖再測")
}

// 原地不動的一步（方向碼 4，docs/re/26 §1.1）：座標不動，
// 但時鐘與遭遇照跑，而且**不算玩家走的**。
func TestIdleStepDoesNotMoveButAdvancesTime(t *testing.T) {
	rom := openRom(t)
	block, err := rom.Block(0)
	if err != nil {
		t.Fatalf("載入區塊 0 失敗：%v", err)
	}
	party := &Party{Members: []*Character{{CON: 20, MaxCON: 20}}, X: 55, Y: 62}
	w := NewWorld(block, party, rng.New())

	// 先走一步讓 PlayerStepped 變 true。
	for dir := Up; dir <= Right; dir++ {
		if res, err := w.Step(dir); err == nil && res.Moved {
			break
		}
	}
	if !w.Party.PlayerStepped {
		t.Skip("四個方向都走不動，換一張地圖再測")
	}

	x, y := w.Party.X, w.Party.Y
	before := w.Clock
	res := w.IdleStep()

	if res.Moved {
		t.Fatal("原地一步不該回報 Moved")
	}
	if w.Party.X != x || w.Party.Y != y {
		t.Fatalf("座標動了：(%d,%d) → (%d,%d)", x, y, w.Party.X, w.Party.Y)
	}
	if w.Clock == before {
		t.Fatal("時鐘應該照樣前進")
	}
	if w.Party.PlayerStepped {
		t.Fatal("原地一步之後不該還算玩家步——那會讓站著不動也能刷經驗值")
	}
}

// 驗收（規格 07 §6.4）：42 張地圖的 nibble 9 格子，記錄的兩個 byte 都拿得到，
// 而且統計與 docs/re/55 §4 的獨立計數相同（211 格、84 格無視護甲）。
//
// 數字寫死是刻意的：這一批是**資料**，換了解析方式就會變，
// 變了要當成解析錯而不是「資料本來就這樣」。
func TestRadiationRecordsAcrossAllMaps(t *testing.T) {
	rom := openRom(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatalf("列資源失敗：%v", err)
	}
	cells, bypass := 0, 0
	dice := map[byte]int{}
	for i := range res {
		b, err := rom.Block(i)
		if err != nil {
			t.Fatalf("區塊 %d：%v", i, err)
		}
		dim := int(b.Dim)
		for y := 0; y < dim; y++ {
			for x := 0; x < dim; x++ {
				terrain, record, _, err := b.At(x, y)
				if err != nil || terrain != 9 {
					continue
				}
				cells++
				rec, err := b.SectionRecord(9, int(record))
				if err != nil || len(rec) < 2 {
					t.Fatalf("資源 %d (%d, %d)：nibble 9 的記錄取不到（%v）", i, x, y, err)
				}
				if RadiationBypassesArmour(rec) {
					bypass++
				}
				dice[RadiationDice(rec)]++
			}
		}
	}
	if cells != 211 {
		t.Errorf("nibble 9 有 %d 格，docs/re/48 §6 與 docs/re/55 §4 都是 211", cells)
	}
	if bypass != 84 {
		t.Errorf("無視護甲的有 %d 格，docs/re/55 §4 是 84", bypass)
	}
	for n := range dice {
		if n < 2 || n > 10 {
			t.Errorf("骰數 %d 落在 2–10 之外", n)
		}
	}
	t.Logf("211 格、%d 格無視護甲、骰數分布 %v", bypass, dice)
}

// 驗收（規格 07 §6.4）：無視護甲時吸收是 0；不論扣多少血，每個人都會中毒。
func TestApplyRadiation(t *testing.T) {
	newWorld := func() *World {
		return &World{
			Party: &Party{Members: []*Character{
				{CON: 5, MaxCON: 20, AC: 200}, // AC 高到吸收幾乎必然大於傷害
				{CON: 5, MaxCON: 20, AC: 200},
			}},
			RNG: rng.New(),
		}
	}

	// +0x00 bit0 ＝ 1 → 無視護甲：AC 再高也擋不住。
	w := newWorld()
	hits := w.ApplyRadiation([]byte{23, 10})
	for _, h := range hits {
		if h.Absorb != 0 {
			t.Errorf("無視護甲時吸收應該是 0，得到 %d", h.Absorb)
		}
		if h.Applied != h.Rolled || h.Applied < 10 {
			t.Errorf("10 顆 d6 應該全額扣：擲 %d、扣 %d", h.Rolled, h.Applied)
		}
	}
	if w.Party.Members[0].CON >= 5 {
		t.Error("無視護甲卻沒扣到 CON")
	}
	if w.Party.Members[0].CON > 0 {
		t.Logf("CON 扣到 %d（可以是負的）", w.Party.Members[0].CON)
	}

	// +0x00 bit0 ＝ 0 → 照常吸收：AC 200 顆 d6 幾乎不可能被打穿。
	w = newWorld()
	hits = w.ApplyRadiation([]byte{22, 2})
	for _, h := range hits {
		if h.Absorb == 0 {
			t.Error("沒有無視護甲卻沒算吸收")
		}
		if h.Applied != 0 {
			t.Errorf("AC 200 對 2 顆 d6 竟然扣了 %d", h.Applied)
		}
	}
	// **扣不扣血與中不中毒無關**——這一條最容易寫成「有扣血才中毒」。
	for i, c := range w.Party.Members {
		if c.Status&StatusRadiation == 0 {
			t.Errorf("第 %d 個人沒有中輻射毒", i)
		}
		if c.CON != 5 {
			t.Errorf("第 %d 個人不該掉血，CON ＝ %d", i, c.CON)
		}
	}
}

// 驗收：nibble 9 的訊息編號是**記錄 +0x00**，不是第 2 層的記錄編號。
//
// 對照 docs/re/48 §5：資源 0 的輻射格訊息是 `The ground seems to glow here.`。
// 拿第 2 層值當編號會查到完全不相干的字串——這一條就是在擋那個。
func TestRadiationMessageComesFromRecord(t *testing.T) {
	rom := openRom(t)
	b, err := rom.Block(0)
	if err != nil {
		t.Fatalf("區塊 0：%v", err)
	}
	w := &World{Block: b, Party: &Party{Members: []*Character{{CON: 10}}}, RNG: rng.New()}
	dim := int(b.Dim)
	found := 0
	for y := 0; y < dim && found == 0; y++ {
		for x := 0; x < dim; x++ {
			terrain, _, _, err := b.At(x, y)
			if err != nil || terrain != 9 {
				continue
			}
			ev := w.trigger(x, y)
			if ev.Kind != EventRadiation {
				t.Fatalf("(%d, %d) 的 nibble 9 沒有走輻射那條：%v", x, y, ev.Kind)
			}
			if len(ev.Strings) != 1 {
				t.Fatalf("(%d, %d) 沒有訊息編號", x, y)
			}
			if n := ev.Strings[0]; n != int(ev.Data[0]) {
				t.Fatalf("訊息編號 %d ≠ 記錄 +0x00 的 %d", n, ev.Data[0])
			}
			// 原版的字串前後帶 \r（換行控制碼），比對內容不比對外框。
			s := b.Strings[ev.Strings[0]]
			if !strings.Contains(s, "The ground seems to glow here.") {
				t.Fatalf("資源 0 的輻射訊息是 %q，docs/re/48 §5 是 "+
					"`The ground seems to glow here.`", s)
			}
			found++
			break
		}
	}
	if found == 0 {
		t.Fatal("資源 0 找不到 nibble 9 的格子——docs/re/48 §6 說有 36 格")
	}
}

// 驗收（規格 07 §6.4）：穿著 Rad suit（物品 41）的人完全不受輻射影響。
//
// 這一條是**實跑出來的**：沒有它，走進資源 0 的輻射帶三步全隊倒地，
// 那條帶子有 36 格連續——遊戲會變成走不過去（docs/re/55 §3）。
func TestRadSuitBlocksRadiation(t *testing.T) {
	suited := &Character{CON: 30, MaxCON: 30, ArmorIndex: 3,
		Items: []Slot{{}, {}, {}, {ID: ItemRadSuit}}}
	bare := &Character{CON: 30, MaxCON: 30}
	robed := &Character{CON: 30, MaxCON: 30, ArmorIndex: 3,
		Items: []Slot{{}, {}, {}, {ID: 42}}} // Robe：一樣是護甲，但擋不了輻射

	w := &World{Party: &Party{Members: []*Character{suited, bare, robed}}, RNG: rng.New()}
	hits := w.ApplyRadiation([]byte{23, 10}) // 訊息 23 ＝ 奇數 → 無視護甲

	if !hits[0].Immune || hits[0].Applied != 0 {
		t.Errorf("穿 Rad suit 的人受了 %d 點傷", hits[0].Applied)
	}
	if suited.Status&StatusRadiation != 0 {
		t.Error("穿 Rad suit 的人不該中輻射毒——原版整個人跳過")
	}
	if suited.CON != 30 {
		t.Errorf("穿 Rad suit 的人 CON 變成 %d", suited.CON)
	}
	for i, c := range []*Character{bare, robed} {
		if c.CON >= 30 {
			t.Errorf("第 %d 個沒穿 Rad suit 的人竟然沒掉血", i)
		}
		if c.Status&StatusRadiation == 0 {
			t.Errorf("第 %d 個沒穿 Rad suit 的人沒中毒", i)
		}
	}
}

// 驗收（docs/re/62）：nibble 11 一律擋住，而且擋住時要回報那一格的訊息編號。
//
// 這一條是**實機對拍**逼出來的：同一串按鍵，原版被山擋住而 remake 穿了過去。
// nibble 11 有 20,495 格、42 張地圖全部都有——漏掉它整個地圖的形狀都不對。
func TestNibble11BlocksMovement(t *testing.T) {
	rom := openRom(t)
	b, err := rom.Block(0)
	if err != nil {
		t.Fatal(err)
	}
	// 找一格：站得住、右邊是 nibble 11。
	var px, py int = -1, -1
	for y := 1; y < b.Dim-1 && px < 0; y++ {
		for x := 1; x < b.Dim-2; x++ {
			here, _, _, err := b.At(x, y)
			if err != nil || blocking[here] {
				continue
			}
			right, _, _, err := b.At(x+1, y)
			if err != nil || right != 11 {
				continue
			}
			px, py = x, y
			break
		}
	}
	if px < 0 {
		t.Fatal("資源 0 找不到「站得住、右邊是 nibble 11」的格子")
	}

	w := &World{Block: b, Party: &Party{X: uint8(px), Y: uint8(py),
		Members: []*Character{{CON: 10}}}, RNG: rng.New()}
	res, err := w.Step(Right)
	if err != nil {
		t.Fatal(err)
	}
	if res.Moved {
		t.Fatalf("(%d, %d) 往右是 nibble 11，卻走過去了", px, py)
	}
	if w.Party.X != uint8(px) {
		t.Fatal("被擋住卻動了座標")
	}
	// 原版擋住時印的是記錄 +0x00 的訊息，不是一句固定的 BLOCKED。
	if res.Blocked <= 0 || res.Blocked >= len(b.Strings) {
		t.Fatalf("被擋住時回報的字串編號是 %d，取不到訊息", res.Blocked)
	}
	t.Logf("(%d, %d) 往右被擋，訊息：%q", px, py, b.Strings[res.Blocked])
}

// 驗收（docs/re/65）：nibble 2 是條件式的，三條路各自要對。
//
// 資料面：2,699 格裡 1,522 格 bit7 設（直接放行）、1,031 格 bit6 沒設（放行）、
// 只有 146 格要判定。**把整個 nibble 2 當成牆會擋掉 94% 本該能走的格子。**
func TestNibble2IsConditional(t *testing.T) {
	rom := openRom(t)
	res, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	var pass7, pass6, judged int
	for _, r := range res {
		b, err := rom.BlockByID(r.ID)
		if err != nil {
			continue
		}
		w := &World{Block: b, Party: &Party{Members: []*Character{{CON: 10}}}, RNG: rng.New()}
		for y := 0; y < b.Dim; y++ {
			for x := 0; x < b.Dim; x++ {
				terr, rec, _, err := b.At(x, y)
				if err != nil || terr != 2 {
					continue
				}
				d, err := b.SectionRecord(2, int(rec))
				if err != nil || len(d) == 0 {
					continue
				}
				need, _ := w.gateNeedsCheck(x, y)
				switch {
				case d[0]&0x80 != 0:
					pass7++
					if need {
						t.Fatalf("(%d,%d) 的 +0x00 是 %#02x（bit7 設）卻要判定", x, y, d[0])
					}
				case d[0]&0x40 == 0:
					pass6++
					if need {
						t.Fatalf("(%d,%d) 的 +0x00 是 %#02x（bit6 沒設）卻要判定", x, y, d[0])
					}
				default:
					judged++
					if !need {
						t.Fatalf("(%d,%d) 的 +0x00 是 %#02x 要判定，卻直接放行", x, y, d[0])
					}
				}
			}
		}
	}
	if pass7 != 1522 || pass6 != 1031 || judged != 146 {
		t.Errorf("分布變了：bit7 %d（1522）、bit6 沒設 %d（1031）、要判定 %d（146）",
			pass7, pass6, judged)
	}
	t.Logf("nibble 2：直接放行 %d ＋ %d，要判定 %d", pass7, pass6, judged)
}

// 驗收（docs/re/65）：那 146 格真的會跑條件判定，通過才走得過去。
//
// 挑的是**型別 3（隊伍人數）**：不擲骰、不消耗物品，所以結果是決定性的——
// 用擲骰型的條件寫測試會變成偶爾紅。
func TestConditionGateEvaluates(t *testing.T) {
	rom := openRom(t)
	b, err := rom.BlockByID(2)
	if err != nil {
		t.Fatal(err)
	}
	const gx, gy = 4, 11
	rec, err := b.SectionRecord(2, mustRecord(t, b, gx, gy))
	if err != nil {
		t.Fatal(err)
	}
	gates := ParseGates(rec)
	if len(gates) != 1 || gates[0].Type != GatePartySize || gates[0].Param != 1 {
		t.Fatalf("資源 2 的 (%d, %d) 條件變了：%+v", gx, gy, gates)
	}

	// 站在左邊往右走進那一格。
	step := func(members int) StepResult {
		p := &Party{X: gx - 1, Y: gy}
		for i := 0; i < members; i++ {
			p.Members = append(p.Members, &Character{CON: 20, MaxCON: 20})
		}
		w := NewWorld(b, p, rng.New())
		res, err := w.Step(Right)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	// 四個人：條件要求 1 人 → 過不了。
	if r := step(4); r.Moved {
		t.Error("隊伍 4 人卻過了「人數 ＝ 1」的條件閘")
	}
	// 一個人：通過。
	if r := step(1); !r.Moved {
		t.Errorf("隊伍 1 人卻沒過（Blocked ＝ %d）", r.Blocked)
	}
}

// mustRecord 取這一格的第 2 層值（記錄索引）。
func mustRecord(t *testing.T, b *assets.Block, x, y int) int {
	t.Helper()
	terrain, rec, _, err := b.At(x, y)
	if err != nil || terrain != 2 {
		t.Fatalf("(%d, %d) 不是 nibble 2（%d，%v）", x, y, terrain, err)
	}
	return int(rec)
}

// 驗收（docs/re/67 §3）：條件閘是**每個能行動的人都要過才放行**，
// 沒過的各自受罰——不是「有人過就好」。
func TestEvalGateShape(t *testing.T) {
	// 條件：隊伍人數 ＝ 1（型別 3）；懲罰：CON 固定減 1。
	rec := make([]byte, 16)
	rec[0x00] = 0xE1
	rec[0x08] = 0x80 | 0x1D // CON、固定值
	rec[0x09] = 0x80 | 1    // 減 1
	rec[0x0A] = byte(GatePartySize << 5)
	rec[0x0B] = 1
	rec[0x0C] = 0xFF

	// 兩個人：條件要求 1 人 → 兩個都沒過 → 擋住、兩個都扣 1。
	p := &Party{Members: []*Character{
		{CON: 10, MaxCON: 10}, {CON: 10, MaxCON: 10},
	}}
	g := p.EvalGate(rng.New(), rec, nil)
	if !g.Blocked || len(g.Failed) != 2 {
		t.Fatalf("兩個人都該沒過：Blocked=%v Failed=%+v", g.Blocked, g.Failed)
	}
	for i, c := range p.Members {
		if c.CON != 9 {
			t.Errorf("第 %d 個人的 CON 是 %d，該扣成 9", i, c.CON)
		}
	}

	// 一個人：通過 → 不擋、不扣。
	p = &Party{Members: []*Character{{CON: 10, MaxCON: 10}}}
	g = p.EvalGate(rng.New(), rec, nil)
	if g.Blocked || len(g.Failed) != 0 {
		t.Fatalf("一個人該通過：Blocked=%v Failed=%+v", g.Blocked, g.Failed)
	}
	if p.Members[0].CON != 10 {
		t.Errorf("通過卻扣了血：CON ＝ %d", p.Members[0].CON)
	}

	// 不能行動的人跳過，不算沒過。
	p = &Party{Members: []*Character{{CON: 10, MaxCON: 10}, {CON: 0}}}
	g = p.EvalGate(rng.New(), rec, nil)
	if len(g.Failed) != 1 || g.Failed[0].Member != 0 {
		t.Fatalf("CON ＝ 0 的人不該被算進去：%+v", g.Failed)
	}
}

// 驗收（docs/re/67 §1）：+0x08 的 bit7 決定固定或擲骰、+0x09 的 bit7 決定加減。
func TestGatePenaltyBits(t *testing.T) {
	mk := func(f, q byte) []byte {
		rec := make([]byte, 16)
		rec[0x08], rec[0x09] = f, q
		rec[0x0A] = 0xFF
		return rec
	}
	// 固定值減 3。
	c := &Character{CON: 20, MaxCON: 20}
	if _, amt := applyGatePenalty(rng.New(), c, mk(0x80|0x1D, 0x80|3)); amt != -3 || c.CON != 17 {
		t.Errorf("固定減 3：量 %d、CON %d", amt, c.CON)
	}
	// 固定值加 3。
	c = &Character{CON: 20, MaxCON: 20}
	if _, amt := applyGatePenalty(rng.New(), c, mk(0x80|0x1D, 3)); amt != 3 || c.CON != 23 {
		t.Errorf("固定加 3：量 %d、CON %d", amt, c.CON)
	}
	// 擲骰：2 顆 d6 減，落在 2–12。
	c = &Character{CON: 100, MaxCON: 100}
	_, amt := applyGatePenalty(rng.New(), c, mk(0x1D, 0x80|2))
	if amt > -2 || amt < -12 {
		t.Errorf("2 顆 d6 減，得到 %d", amt)
	}
	// 欄位 0 ＝ 什麼都不做。
	c = &Character{CON: 20, MaxCON: 20}
	if f, amt := applyGatePenalty(rng.New(), c, mk(0, 0x80|5)); f != 0 || amt != 0 || c.CON != 20 {
		t.Errorf("欄位 0 卻動了：%d %d %d", f, amt, c.CON)
	}
	// 金錢不會扣成負的。
	c = &Character{Money: 2}
	applyGatePenalty(rng.New(), c, mk(0x80|0x15, 0x80|10))
	if c.Money != 0 {
		t.Errorf("金錢扣過頭：%d", c.Money)
	}
}

// 驗收（docs/re/68）：條件閘收尾會改寫地圖格——全部人都過用 +0x04、
// 有人沒過用 +0x06，第一個 byte 的 bit7 設就不改。
func TestGateRewritesCell(t *testing.T) {
	rom := openRom(t)
	b, err := rom.BlockByID(0)
	if err != nil {
		t.Fatal(err)
	}
	// 找一格：站得住、右邊是指到記錄 2 的 nibble 2（+0x06 ＝ 0x0A 傳送）。
	var px, py = -1, -1
	for y := 1; y < b.Dim-1 && px < 0; y++ {
		for x := 1; x < b.Dim-2; x++ {
			if here, _, _, err := b.At(x, y); err != nil || blocking[here] {
				continue
			}
			if right, rec, _, err := b.At(x+1, y); err == nil && right == 2 && rec == 2 {
				px, py = x, y
				break
			}
		}
	}
	if px < 0 {
		t.Skip("資源 0 找不到指到記錄 2 的 nibble 2 格")
	}
	rec, err := b.SectionRecord(2, 2)
	if err != nil || len(rec) < 8 {
		t.Fatalf("記錄 2：%v", err)
	}
	if rec[0x06] != 0x0A {
		t.Fatalf("記錄 2 的 +0x06 是 %#02x，docs/re/68 §3 說是 0x0A", rec[0x06])
	}

	raw, _ := rom.SkillTableRaw()
	w := NewWorld(b, &Party{X: uint8(px), Y: uint8(py),
		Members: []*Character{{CON: 20, MaxCON: 20}}}, rng.New())
	w.Skills = SkillBytes(raw)
	if _, err := w.Step(Right); err != nil {
		t.Fatal(err)
	}
	// 條件是「技能 7 難度 6」，出廠角色沒有 → 沒過 → 用 +0x06 改寫。
	after, _, _, err := b.At(px+1, py)
	if err != nil {
		t.Fatal(err)
	}
	if after != 0x0A {
		t.Errorf("(%d, %d) 沒被改寫成 nibble 10，還是 %d", px+1, py, after)
	}
}
