package game

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game/rng"
)

// `+0x09` 的兩個條件**都要成立**（`0x132D8` 的 bit1、`0x132DD` 的編號非零）。
//
// ⚠ 只看其中一個會把「有敵人但不能雇用」的格子當成可以雇用——
// 而那時畫面上會出現一個根本不在那裡的 NPC。
func TestReadHireOffer(t *testing.T) {
	cases := []struct {
		name string
		v    byte
		want HireOffer
	}{
		{"兩個都成立", 0x13, HireOffer{NPC: 1, Valid: true}},
		{"沒設 bit1", 0x11, HireOffer{NPC: 1}},
		{"編號是 0", 0x03, HireOffer{NPC: 0}},
		{"出廠常見值", 0x01, HireOffer{NPC: 0}},
		{"高編號", 0x6F, HireOffer{NPC: 6, Valid: true}},
	}
	for _, tc := range cases {
		rec := make([]byte, 0x10)
		rec[recHireField] = tc.v
		if got := ReadHireOffer(rec); got != tc.want {
			t.Errorf("%s（%#04x）＝ %+v，預期 %+v", tc.name, tc.v, got, tc.want)
		}
	}
	// 記錄短到沒有那一格時不能當成「可以雇用」。
	if got := ReadHireOffer([]byte{1, 2, 3}); got.Valid {
		t.Error("記錄太短卻說可以雇用")
	}
}

// `+0x31` 是 0 就**直接加入，連骰都不擲**（`0x13304`）。
func TestTryHireFreeSkipsTheRoll(t *testing.T) {
	r := rng.New()
	before := *r
	hirer := &Character{Level: 1}
	npc := &Character{Level: 99}
	out := TryHire(hirer, npc, 0, r)
	if !out.Joined || !out.Free {
		t.Fatalf("門檻 0 應該直接加入：%+v", out)
	}
	if *r != before {
		t.Error("擲骰了——門檻 0 那條路不該動亂數")
	}
}

// 骰 < 5 直接失敗（`0x1333B`），**而且不看雙方的數字**。
func TestTryHireLowRollFails(t *testing.T) {
	hirer := &Character{Level: 99}
	hirer.Attributes[AttrCharisma] = 18
	hirer.Attributes[AttrIQ] = 18
	npc := &Character{Level: 1}
	// 找一顆會擲出 < 5 的種子。
	for seed := 0; seed < 500; seed++ {
		r := rng.New()
		for i := 0; i < seed; i++ {
			r.D6()
		}
		probe := *r
		if HireRoll(&probe) >= hireRollFloor {
			continue
		}
		out := TryHire(hirer, npc, 1, r)
		if out.Joined {
			t.Fatalf("骰 %d 小於 %d 卻加入了：%+v", out.Roll, hireRollFloor, out)
		}
		return
	}
	t.Skip("這條亂數序列裡找不到小於 5 的一擲")
}

// 檢定的公式：**魅力在招募者那一側算兩次**（`docs/re/110` §2.1）。
func TestTryHireFormula(t *testing.T) {
	hirer := &Character{Level: 3}
	hirer.Attributes[AttrCharisma] = 16
	hirer.Attributes[AttrIQ] = 10
	npc := &Character{Level: 2}
	npc.Attributes[AttrCharisma] = 8
	npc.Attributes[AttrIQ] = 8
	out := TryHire(hirer, npc, 5, rng.New())
	// 招募者 ＝ (16+10)/2 ＋ 16 ＋ 3 ＝ 32（再加骰）
	// NPC    ＝ (8+8)/2 ＋ 5 ＋ 2 ＝ 15
	if out.Theirs != 15 {
		t.Errorf("NPC 那一側 ＝ %d，預期 15", out.Theirs)
	}
	if out.Roll < hireRollFloor {
		t.Skip("這一擲太小，公式驗不到")
	}
	if want := 32 + out.Roll; out.Ours != want {
		t.Errorf("招募者那一側 ＝ %d，預期 %d（骰 %d）", out.Ours, want, out.Roll)
	}
	if !out.Joined {
		t.Errorf("32 ＋ 骰 對 15 應該成功：%+v", out)
	}
}

// 那顆骰是「擲到不同點才停」（`sub_19C84`）：值域從 3 起跳（1+2），
// **不會是 2**（那要兩顆同點，而同點會繼續擲）。
func TestHireRollNeverStopsOnDoubles(t *testing.T) {
	r := rng.New()
	seen := map[int]int{}
	for i := 0; i < 2000; i++ {
		v := HireRoll(r)
		if v < 3 {
			t.Fatalf("擲出 %d，值域最小是 3", v)
		}
		seen[v]++
	}
	if len(seen) < 5 {
		t.Errorf("只出現 %d 種值，分布不像可變長度的骰：%v", len(seen), seen)
	}
}

// 雇用來的隊員要**留著整筆 256 bytes**，而且 `+0x31` 被改寫成招募者的魅力。
func TestHireNPCKeepsTheWholeRecord(t *testing.T) {
	rec := make([]byte, 0x100)
	copy(rec, "REDHAWK")
	rec[recNPCPrice] = 9
	rec[0x77] = 0xAB // 未解區域：要原樣留著
	c := HireNPC(rec, 17)
	if c == nil {
		t.Fatal("建不出隊員")
	}
	if c.Name != "REDHAWK" {
		t.Errorf("名字 ＝ %q", c.Name)
	}
	if len(c.Source) != 0x100 {
		t.Fatalf("Source ＝ %d bytes，預期 256", len(c.Source))
	}
	if c.Source[recNPCPrice] != 17 {
		t.Errorf("+0x31 ＝ %d，預期寫成招募者的魅力 17", c.Source[recNPCPrice])
	}
	if c.Source[0x77] != 0xAB {
		t.Error("未解區域被動到了")
	}
	// **是副本不是別名**：改了來源不該影響隊員。
	rec[0x77] = 0
	if c.Source[0x77] != 0xAB {
		t.Error("Source 是來源記錄的別名，不是副本")
	}
}

// 候選數是「有敵人的組」，**不是「可以雇用的組」**（`loc_1382B`）。
func TestHireCandidatesCountsGroupsWithEnemies(t *testing.T) {
	var g [EnemyGroups]SpawnGroup
	g[0] = SpawnGroup{Type: 3, Count: 2}
	g[2] = SpawnGroup{Type: 5, Count: 1}
	if n := HireCandidates(g); n != 2 {
		t.Errorf("候選 ＝ %d，預期 2", n)
	}
	var empty [EnemyGroups]SpawnGroup
	if n := HireCandidates(empty); n != 0 {
		t.Errorf("空的遭遇候選 ＝ %d，預期 0", n)
	}
}
