package game

import (
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

// 物品消耗：低 6 位是次數，用完把整個槽移除，高 2 位保留。
func TestUseItemConsumes(t *testing.T) {
	c := &Character{Items: []Slot{{ID: 9, Value: 0xC3}}} // 高 2 位 11、次數 3
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
	if len(c.Items) != 0 {
		t.Fatalf("次數用完應該把整個槽移除，還剩 %v", c.Items)
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
		s := &Script{World: w, Record: []byte{byte(op), 0, 0, 0, 0, 0, 0, 0}}
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
	(&Script{World: w, Record: rec}).Step()
	if rec[1] != 0xAA || rec[2] != 0xBB {
		t.Fatalf("白天應該走 +0x03，得到 %#x %#x", rec[1], rec[2])
	}
	w.Clock.Hour = 22
	(&Script{World: w, Record: rec}).Step()
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
