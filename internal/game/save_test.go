package game

import (
	"bytes"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 驗收 1／2：讀出再寫回，byte-for-byte 相同（docs/spec/05 §6）。
func TestSaveRoundTrip(t *testing.T) {
	rom := openRom(t)
	for _, file := range []string{"game1", "game2"} {
		save, err := rom.LoadSave(file)
		if err != nil {
			t.Fatalf("%s 讀存檔失敗：%v", file, err)
		}
		orig, err := rom.File(file)
		if err != nil {
			t.Fatalf("讀 %s 失敗：%v", file, err)
		}
		at := assets.SaveOffset[file]
		want := orig[at : at+len(save.Bytes())]
		if got := save.Bytes(); !bytes.Equal(got, want) {
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("%s 寫回不一致，第一個差異在 +%#x：%#02x vs %#02x",
						file, i, got[i], want[i])
				}
			}
			t.Fatalf("%s 寫回長度不一致：%d vs %d", file, len(got), len(want))
		}
	}
}

// 驗收 3：出廠的四個 Ranger。
func TestShippedRangers(t *testing.T) {
	rom := openRom(t)
	save, err := rom.LoadSave("game1")
	if err != nil {
		t.Fatalf("讀存檔失敗：%v", err)
	}
	want := []string{"Hell Razor", "Angela Deth", "Thrasher", "Snake Vargas"}
	for i, name := range want {
		raw, err := save.Record(i + 1)
		if err != nil {
			t.Fatalf("取第 %d 筆失敗：%v", i+1, err)
		}
		c := LoadCharacter(raw)
		if c.Name != name {
			t.Errorf("第 %d 個角色應該是 %q，得到 %q", i+1, name, c.Name)
		}
		if c.Level != 1 {
			t.Errorf("%s 的等級應該是 1，得到 %d", c.Name, c.Level)
		}
		if c.Rank != "Private" {
			t.Errorf("%s 的階級應該是 Private，得到 %q", c.Name, c.Rank)
		}
		if c.CON != c.MaxCON {
			t.Errorf("%s 出廠應該滿血：CON %d、MAXCON %d", c.Name, c.CON, c.MaxCON)
		}
		if c.XP != 0 {
			t.Errorf("%s 出廠經驗值應該是 0，得到 %d", c.Name, c.XP)
		}
		for a, v := range c.Attributes {
			if v < 3 || v > 18 {
				t.Errorf("%s 的 %s ＝ %d，超出 5d6 取三的值域 3–18",
					c.Name, AttributeNames[a], v)
			}
		}
		// 建立時技能點 ＝ IQ（docs/re/21 §5）。
		if c.SkillPts != c.Attributes[AttrIQ] {
			t.Logf("%s 的技能點 %d 與 IQ %d 不同（可能已經花掉過）",
				c.Name, c.SkillPts, c.Attributes[AttrIQ])
		}
	}
	// 第 5–7 筆應該是空的。
	for i := 5; i <= 7; i++ {
		raw, _ := save.Record(i)
		if c := LoadCharacter(raw); c.Name != "" {
			t.Errorf("第 %d 筆應該是空的，卻有名字 %q", i, c.Name)
		}
	}
}

// 驗收 4：隊伍槽表。
func TestSlotGroups(t *testing.T) {
	rom := openRom(t)
	save, err := rom.LoadSave("game1")
	if err != nil {
		t.Fatalf("讀存檔失敗：%v", err)
	}
	groups := save.SlotGroups()
	var members []byte
	for _, m := range groups[0].Members {
		if m != 0 {
			members = append(members, m)
		}
	}
	if !bytes.Equal(members, []byte{1, 2, 3, 4}) {
		t.Errorf("第 0 組成員應該是 1–4，得到 %v", members)
	}
	if groups[0].X != 55 || groups[0].Y != 62 {
		t.Errorf("第 0 組座標應該是 (55, 62)，得到 (%d, %d)", groups[0].X, groups[0].Y)
	}
	if groups[0].MapID != 0 {
		t.Errorf("第 0 組地圖應該是 0，得到 %d", groups[0].MapID)
	}
	for i := 1; i < len(groups); i++ {
		for _, m := range groups[i].Members {
			if m != 0 {
				t.Errorf("第 %d 組應該是空的，卻有成員 %d", i, m)
			}
		}
	}
}

// 驗收 5：改一個欄位再寫回，只有那幾個 byte 與 checksum 變動。
func TestEditOnlyTouchesKnownBytes(t *testing.T) {
	rom := openRom(t)
	save, err := rom.LoadSave("game1")
	if err != nil {
		t.Fatalf("讀存檔失敗：%v", err)
	}
	before := append([]byte(nil), save.Plain...)

	raw, _ := save.Record(1)
	c := LoadCharacter(raw)
	c.Money = 12345
	c.StoreTo(raw)

	var changed []int
	for i := range save.Plain {
		if save.Plain[i] != before[i] {
			changed = append(changed, i)
		}
	}
	// 金錢在第 1 筆的 +0x15–+0x17 ＝ 明文的 0x115–0x117。
	want := []int{0x115, 0x116, 0x117}
	if len(changed) > len(want) {
		t.Fatalf("改金錢卻動了 %d 個 byte：%v", len(changed), changed)
	}
	for _, at := range changed {
		if at < want[0] || at > want[len(want)-1] {
			t.Fatalf("改金錢卻動到 +%#x", at)
		}
	}
	if got := LoadCharacter(raw).Money; got != 12345 {
		t.Fatalf("寫回之後讀出來是 %d", got)
	}
}

// 驗收 6：兩份輪替挑序號大的。
func TestPickNewer(t *testing.T) {
	rom := openRom(t)
	a, err := rom.LoadSave("game1")
	if err != nil {
		t.Fatalf("讀 game1 失敗：%v", err)
	}
	b, err := rom.LoadSave("game2")
	if err != nil {
		t.Fatalf("讀 game2 失敗：%v", err)
	}
	if a.Serial() <= b.Serial() {
		t.Fatalf("出廠時 game1 的序號應該比較大：%d vs %d", a.Serial(), b.Serial())
	}
	if assets.PickNewer(a, b) != a {
		t.Fatalf("應該挑 game1")
	}
	if got := a.Place(); got != "Ranger Ctr." && got != "Ranger Ctr. " {
		t.Errorf("地點名稱是 %q", got)
	}
}

// 升級門檻與連升（docs/re/31 §1）。
func TestLevelThresholds(t *testing.T) {
	for _, tc := range []struct {
		level int
		xp    uint32
	}{{2, 1024}, {3, 3072}, {4, 6144}, {5, 10240}, {6, 15360}, {10, 46080}} {
		if got := XPForLevel(tc.level); got != tc.xp {
			t.Errorf("升到等級 %d 應該要 %d 點，得到 %d", tc.level, tc.xp, got)
		}
	}

	c := &Character{Level: 1, SkillPts: 0}
	c.AddXP(3072) // 剛好夠到等級 3
	if n := c.LevelUp(); n != 2 {
		t.Fatalf("3072 點應該一次升兩級，得到 %d", n)
	}
	if c.Level != 3 || c.SkillPts != 2 {
		t.Fatalf("升完應該是等級 3、技能點 2，得到 %d／%d", c.Level, c.SkillPts)
	}
	// 升級不扣經驗值。
	if c.XP != 3072 {
		t.Fatalf("升級不該扣經驗值，剩 %d", c.XP)
	}
}

func TestXPSaturates(t *testing.T) {
	c := &Character{XP: maxUint24 - 5}
	c.AddXP(100)
	if c.XP != maxUint24 {
		t.Fatalf("經驗值應該飽和在 %#x，得到 %#x", uint32(maxUint24), c.XP)
	}
}

// 技能費用：基礎 × 2^(L−1)，飽和 0xFF。
func TestSkillCost(t *testing.T) {
	for _, tc := range []struct {
		base  byte
		level int
		want  byte
	}{{1, 1, 1}, {1, 2, 2}, {1, 3, 4}, {1, 5, 16}, {3, 1, 3}, {3, 3, 12}, {1, 9, 0xFF}} {
		if got := SkillCost(tc.base, tc.level); got != tc.want {
			t.Errorf("基礎 %d 升到等級 %d 應該是 %d，得到 %d",
				tc.base, tc.level, tc.want, got)
		}
	}
}

// 技能資料的兩個欄位擠在一個 byte 裡。
func TestParseSkillData(t *testing.T) {
	// Brawling：0x19 → IQ 3、費用 1、屬性 0x10（Luck）。
	d := ParseSkillData(0x19, 0x10)
	if d.IQ != 3 || d.BaseCost != 1 || d.Attribute != 0x10 {
		t.Fatalf("拆錯了：%+v", d)
	}
}

func TestLearnSkill(t *testing.T) {
	c := &Character{SkillPts: 10}
	c.Attributes[AttrIQ] = 3
	data := ParseSkillData(0x19, 0x10) // IQ 3、基礎費用 1

	if ok, why := c.LearnSkill(1, data); !ok {
		t.Fatalf("應該學得起來：%s", why)
	}
	if c.SkillLevel(1) != 1 || c.SkillPts != 9 {
		t.Fatalf("學完應該是等級 1、剩 9 點，得到 %d／%d", c.SkillLevel(1), c.SkillPts)
	}
	if ok, _ := c.LearnSkill(1, data); !ok || c.SkillLevel(1) != 2 || c.SkillPts != 7 {
		t.Fatalf("升到等級 2 要花 2 點：等級 %d、剩 %d", c.SkillLevel(1), c.SkillPts)
	}

	// IQ 不夠。
	dumb := &Character{SkillPts: 99}
	dumb.Attributes[AttrIQ] = 2
	if ok, why := dumb.LearnSkill(1, data); ok || why != "IQ 不足" {
		t.Fatalf("IQ 2 不該學得起 IQ 3 的技能，得到 ok=%v why=%q", ok, why)
	}

	// 技能點不夠。
	poor := &Character{SkillPts: 0}
	poor.Attributes[AttrIQ] = 18
	if ok, why := poor.LearnSkill(1, data); ok || why != "技能點不足" {
		t.Fatalf("沒有技能點不該學得起來，得到 ok=%v why=%q", ok, why)
	}
}
