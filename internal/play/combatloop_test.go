package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// TestCombatRunsToCompletion 是戰鬥迴圈的端到端門檻：
// 從地圖走到遭遇、下令、打到結束、收尾回地圖，整條路要跑得完。
//
// 前三個子系統（腳本 `docs/re/76`、遭遇 `docs/re/78`、設施 `docs/re/79`）
// 都有覆蓋率門檻，戰鬥沒有——單元測試綠不代表這條路通得了
// （`docs/re/77` 的教訓：`rollEncounter` 全綠，而遭遇根本打不起來）。
func TestCombatRunsToCompletion(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(0, 12, 2); err != nil {
		t.Fatalf("載入地圖 0 失敗：%v", err)
	}

	// 走到打起來為止（空曠處，docs/re/78 §4）。
	dir := input.DirRight
	steps := 0
	for ; steps < 400 && !s.InCombat(); steps++ {
		if _, err := s.Update(input.Input{Dir: dir}); err != nil {
			t.Fatalf("第 %d 步：%v", steps, err)
		}
		if dir == input.DirRight {
			dir = input.DirLeft
		} else {
			dir = input.DirRight
		}
	}
	if !s.InCombat() {
		t.Fatal("走了 400 步一次遭遇都沒有")
	}
	c := s.Combat()
	if c == nil {
		t.Fatal("InCombat 為 true 但 Combat() 是 nil")
	}
	if n := c.Battle.EnemiesLeft(); n == 0 {
		t.Fatal("打起來了卻一隻敵人都沒有")
	}
	t.Logf("第 %d 步打起來，敵人 %d 隻", steps, c.Battle.EnemiesLeft())

	// 全隊下「攻擊」，然後把回合跑到結束。
	rounds := 0
	for ; rounds < 200; rounds++ {
		c.BeginCommands()
		for !c.Done() {
			if !c.Choose('A', true) { // A ＝ Attack
				c.Choose(' ', true) // 選不到就跳過這個人
			}
		}
		res := c.ResolveRound()
		if res.Over {
			t.Logf("第 %d 回合結束，%s", rounds+1,
				map[bool]string{true: "打贏了", false: "全隊倒下"}[res.Won])
			break
		}
	}
	if rounds >= 200 {
		t.Fatal("200 回合還沒打完——戰鬥迴圈收不了尾")
	}

	// 收尾要回得了地圖。
	out := s.FinishEncounter()
	if !out.Fought {
		t.Error("FinishEncounter 說沒打過")
	}
	if s.InCombat() {
		t.Error("收尾之後還在戰鬥狀態")
	}
	t.Logf("收尾：全滅=%v 經驗 %v", out.Wiped, out.XPGained)
}

// TestCombatManyBattles 連打多場，抓「偶爾收不了尾」與異常結果。
//
// 一場打得完證明路是通的，但**偶發的卡死或零傷害只有多打幾場才看得到**。
func TestCombatManyBattles(t *testing.T) {
	rom := openRom(t)
	const want = 12
	fought, won, stuck := 0, 0, 0
	rounds := map[int]int{}

	for attempt := 0; attempt < 40 && fought < want; attempt++ {
		s, err := New(rom)
		if err != nil {
			t.Fatalf("開場失敗：%v", err)
		}
		// 每一場換一個起點，讓亂數序列不同。
		if err := s.LoadMap(0, uint8(12+attempt%3), uint8(2+attempt%2)); err != nil {
			t.Fatal(err)
		}
		dir := input.DirRight
		for i := 0; i < 400 && !s.InCombat(); i++ {
			if _, err := s.Update(input.Input{Dir: dir}); err != nil {
				break
			}
			if dir == input.DirRight {
				dir = input.DirLeft
			} else {
				dir = input.DirRight
			}
		}
		if !s.InCombat() {
			continue
		}
		fought++
		c := s.Combat()
		n := 0
		for ; n < 200; n++ {
			c.BeginCommands()
			for !c.Done() {
				if !c.Choose('A', true) {
					c.Choose(' ', true)
				}
			}
			if res := c.ResolveRound(); res.Over {
				if res.Won {
					won++
				}
				break
			}
		}
		if n >= 200 {
			stuck++
			continue
		}
		rounds[n+1]++
		s.FinishEncounter()
	}

	t.Logf("打了 %d 場：贏 %d、卡住 %d，回合數分布 %v", fought, won, stuck, rounds)
	if fought < want {
		t.Errorf("只打到 %d 場（目標 %d）——遭遇生成可能退化了", fought, want)
	}
	if stuck != 0 {
		t.Errorf("%d 場在 200 回合內結束不了", stuck)
	}
}
