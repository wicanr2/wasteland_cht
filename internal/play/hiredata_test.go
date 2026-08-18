package play

import (
	"sort"
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 出貨資料裡真的有可雇用的遭遇——而且在 **section 3**，不是 section 15
// （`docs/re/114` §1）。
//
// 這一道守的是一個**正面事實**：14 筆靜態遭遇帶著雇用旗標與 NPC 編號，
// 名字就是遊戲裡那批可招募的人。少一筆代表解讀變了，多一筆代表找到新的。
//
// ⚠ 兩個 section 都是 **12 bytes 一筆**，而且踩上去都走那支什麼都不做的處理函式，
// 所以「翻錯 section」的症狀是**安靜的零命中**——與「資料裡沒有」長得一模一樣。
func TestHireableEncountersInShippedData(t *testing.T) {
	rom := openRom(t)
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}

	const (
		encSection    = 3  // 出貨資料裡的靜態遭遇
		spawnSection  = 15 // 執行期隨機遭遇的槽
		npcSection    = 17
		flagOffset    = 0x09
		friendlyBit   = 0x02
		encRecordSize = 12
	)
	// recordsOf 回傳某個 section 裡**長到有 `+0x09` 的**記錄。
	//
	// ⚠ **不要用「長度剛好 12」當條件。** 遭遇記錄多數是 12 bytes，
	// 但資料裡也有 11 bytes 的，而每個 section 的最後一筆量不到長度
	// （沒有下一個指標）。用等號會安靜地漏掉六個 NPC，
	// 而症狀是「資料裡沒有可雇用的遭遇」——與解讀錯了長得一模一樣。
	recordsOf := func(blk interface {
		SectionCount(int) int
		SectionEntry(int, int) (uint16, error)
		SectionRecord(int, int) ([]byte, error)
	}, typ int) map[int][]byte {
		out := map[int][]byte{}
		n := blk.SectionCount(typ)
		for i := 0; i < n; i++ {
			rec, err := blk.SectionRecord(typ, i)
			if err != nil || len(rec) <= flagOffset {
				continue
			}
			if i+1 < n {
				a, e1 := blk.SectionEntry(typ, i)
				b, e2 := blk.SectionEntry(typ, i+1)
				if e1 == nil && e2 == nil && b > a && int(b-a) <= flagOffset {
					continue // 這一筆短到沒有 +0x09
				}
			}
			if rec[0] == 0 {
				continue // `+0x00` ＝ 0 是空槽，原版十幾處都先測它（`docs/re/37` §2.1）
			}
			out[i] = rec
		}
		return out
	}
	_ = encRecordSize

	var named []string
	unnamed, spawnHits := 0, 0
	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		for i, rec := range recordsOf(blk, encSection) {
			if rec[flagOffset]&friendlyBit == 0 || rec[flagOffset]>>4 == 0 {
				continue
			}
			npc, err := blk.SectionRecord(npcSection, int(rec[flagOffset]>>4))
			if err != nil || len(npc) < 14 {
				unnamed++
				continue
			}
			name := strings.TrimRight(string(npc[:14]), "\x00 ")
			if name == "" || !asciiOnly(name) {
				unnamed++
				continue
			}
			named = append(named, name)
			_ = i
		}
		// 隨機遭遇的槽**一筆都不該有**——`sub_16890` 生成時不碰 `+0x09`。
		for _, rec := range recordsOf(blk, spawnSection) {
			if rec[flagOffset]&friendlyBit != 0 {
				spawnHits++
			}
		}
	}

	sort.Strings(named)
	t.Logf("帶雇用旗標的靜態遭遇：%d 筆 → %s", len(named), strings.Join(named, "、"))
	t.Logf("旗標設著但查不到 NPC 記錄的：%d 筆", unnamed)

	want := []string{
		"ACE", "CHRISTINA", "COVENANT", "DAN CITRINE", "DR. MIKE SCOT",
		"FELICIA", "JACKIE", "MAD DOG FARGO", "MAYOR PEDROS", "METAL MANIAC",
		"MORT", "RALF", "REDHAWK", "VAX",
	}
	if strings.Join(named, "|") != strings.Join(want, "|") {
		t.Errorf("可雇用的 NPC 變了：\n得到 %v\n預期 %v", named, want)
	}
	// 正對照：如果 section 3 掃出 0 筆，上面那一條也會紅，但**下面這一條
	// 才分得出「解讀錯了」與「資料真的沒有」**。
	if spawnHits != 0 {
		t.Errorf("section %d（隨機遭遇的槽）有 %d 筆帶旗標——"+
			"`docs/re/114` §1 說那裡一筆都沒有", spawnSection, spawnHits)
	}
}


// 走到 JACKIE 那一格開打，記錄真的讀得出雇用旗標——
// **端到端**，不是手造的記錄（`internal/play/hire_test.go` 那一組是單元測試）。
func asciiOnly(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] > 0x7E {
			return false
		}
	}
	return true
}

func TestJackieEncounterIsHireable(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	// 高池鎮北邊那一格，JACKIE 就在 (28, 1)。
	if err := s.LoadMap(10, 28, 2); err != nil {
		t.Fatal(err)
	}
	c, err := s.StartEncounter()
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("(28,2) 旁邊應該有一場遭遇——資料裡 3/11 就在 (28,1)")
	}
	offer := hireOfferOf(c)
	if !offer {
		t.Fatalf("這一場應該可以雇用，`+0x09` ＝ %#02x", c.EncRecord[9])
	}
	if !friendlyOf(c) {
		t.Error("可雇用的遭遇一定也是友善的（同一個位元）")
	}
	// 敵方回合不出手：友善的那一組一輪都不會攻擊。
	before := c.World.Party.Members[0].CON
	for i := 0; i < 3; i++ {
		c.ResolveRound()
	}
	if c.World.Party.Members[0].CON < before {
		t.Errorf("友善的那一組打了人：CON 從 %d 掉到 %d",
			before, c.World.Party.Members[0].CON)
	}
	// 打他們一下就翻臉，而且**改的是地圖區塊本身**。
	if !friendlyOf(c) {
		t.Fatal("還沒動手就已經翻臉了")
	}
	c.Phase.Set(0, game.CmdAttack, 0)
	c.resolveCommands()
	c.ResolveRound()
	if friendlyOf(c) {
		t.Error("打了他們還是友善的——`TurnHostile` 沒有接上")
	}
}

// hireOfferOf／friendlyOf 只是把 game 那兩支包一層，讓上面讀得順。
func hireOfferOf(c *CombatScene) bool { return game.ReadHireOffer(c.EncRecord).Valid }
func friendlyOf(c *CombatScene) bool  { return game.Friendly(c.EncRecord) }

// 從玩家的鍵盤走一遍：`H` → `1` → 那一組真的加入隊伍。
//
// ⚠ 這一道守的是**路由**，不是公式。`H` 開得起來、`beginHirePick` 也對，
// 但如果 `Scene.updateCombat` 沒有在「哪一組？」開著時把數字鍵交給
// `PickHire`，按下去就沒有任何反應——而畫面上只是選單又出現一次。
func TestHireFromKeyboardJoinsParty(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LoadMap(10, 28, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartEncounter(); err != nil {
		t.Fatal(err)
	}
	before := len(s.world.Party.Members)

	key := func(c byte) {
		t.Helper()
		if _, err := s.Update(input.Input{Dir: input.DirNone, Char: c}); err != nil {
			t.Fatal(err)
		}
	}
	key('H')
	if !s.combat.HirePicking() {
		t.Fatal("按 H 沒有開「哪一組？」")
	}
	key('1')
	if s.combat != nil && s.combat.HirePicking() {
		t.Fatal("按 1 之後選單還開著——數字鍵沒有交給 PickHire")
	}
	// 其餘的人隨便下個令，把這一回合結掉。
	for i := 0; i < before; i++ {
		key('E')
	}
	if got := len(s.world.Party.Members); got != before+1 {
		t.Fatalf("隊伍 %d 人，雇用之後應該是 %d 人", got, before+1)
	}
	joined := s.world.Party.Members[before]
	if joined.Name != "JACKIE" {
		t.Errorf("加入的是 %q，預期 JACKIE", joined.Name)
	}
	if len(joined.Source) != 0x100 {
		t.Errorf("新隊員的來源記錄 %d bytes，應該是整筆 256", len(joined.Source))
	}
	// ⚠ 他自己帶進來的經驗值**不算這一場賺的**。
	if strings.Contains(s.Message(), "experience") {
		t.Errorf("收尾訊息把新隊員的經驗值算進去了：%q", s.Message())
	}
}
