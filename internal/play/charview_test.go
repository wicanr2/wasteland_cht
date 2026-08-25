package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 地圖上按數字鍵開角色畫面（docs/re/131 §1）；Enter 循環三頁、ESC 關閉。
func TestCharViewOpensAndCycles(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(key('1')); err != nil {
		t.Fatal(err)
	}
	if s.charView.stage != cvInfo {
		t.Fatalf("按 1 應該開資料頁，stage=%d（訊息 %q）", s.charView.stage, s.Message())
	}
	if !strings.Contains(s.Message(), "Hell Razor") && !strings.Contains(s.cjk, "Hell Razor") {
		t.Errorf("資料頁該有名字，訊息 %q／%q", s.Message(), s.cjk)
	}
	for _, want := range []charViewStage{cvItems, cvSkills, cvInfo} {
		if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionConfirm}); err != nil {
			t.Fatal(err)
		}
		if s.charView.stage != want {
			t.Fatalf("Enter 之後應該在 %d，得到 %d", want, s.charView.stage)
		}
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionCancel}); err != nil {
		t.Fatal(err)
	}
	if s.charView.stage != cvOff {
		t.Fatal("ESC 應該關掉角色畫面")
	}
	// 超過人數的數字不開（sub_12760 的人數檢查）。
	if _, err := s.Update(key('7')); err != nil {
		t.Fatal(err)
	}
	if s.charView.stage != cvOff {
		t.Fatal("第 7 個人不存在，不該開")
	}
}

// 物品頁選彈匣 → Reload?（sub_196DB）；選其他 → D/T/E；E 裝備。
func TestCharViewItemActions(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	// Angela Deth（2 號）：VP91Z ＋ 9mm 彈匣，武器已裝備。
	if _, err := s.Update(key('2')); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionConfirm}); err != nil {
		t.Fatal(err)
	}
	if s.charView.stage != cvItems {
		t.Fatalf("應該在物品頁，stage=%d", s.charView.stage)
	}
	c := s.World().Party.Members[1]
	rows := s.cvItemRows()
	// 找一格彈匣（裝備武器的 +0x07）。
	w := slotIndexOf(c.EquipIndex)
	d, _ := s.items.Get(c.Items[w].ID)
	clipRow := -1
	for n, slot := range rows {
		if c.Items[slot].ID == d.Ammo {
			clipRow = n
			break
		}
	}
	if clipRow < 0 || clipRow > 8 {
		t.Skipf("第一頁沒有彈匣（row=%d）", clipRow)
	}
	if _, err := s.Update(key(byte('1' + clipRow))); err != nil {
		t.Fatal(err)
	}
	if s.charView.ask != askReload {
		t.Fatalf("選彈匣應該問 Reload?，ask=%d（訊息 %q %q）", s.charView.ask, s.Message(), s.cjk)
	}
	before := c.Items[w].Value & 0x3F
	if _, err := s.Update(key('Y')); err != nil {
		t.Fatal(err)
	}
	if got := c.Items[w].Value & 0x3F; got != d.Capacity&0x3F {
		t.Errorf("裝填後武器應該滿容量 %d，得到 %d（原本 %d）", d.Capacity, got, before)
	}

	// 換一件非彈匣、非裝備的：D/T/E 選單 → E 裝備。
	rows = s.cvItemRows()
	pick := -1
	for n, slot := range rows {
		if c.Items[slot].ID != d.Ammo && slot != w {
			pick = n
			break
		}
	}
	if pick < 0 {
		t.Skip("沒有可裝備的別件")
	}
	slot := rows[pick]
	// 翻到那一頁（`0` ＝ 下一頁），再按這一頁的第幾項。
	for i := 0; i < pick/cvPageSize; i++ {
		if _, err := s.Update(key('0')); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Update(key(byte('1' + pick%cvPageSize))); err != nil {
		t.Fatal(err)
	}
	if s.charView.ask != askMenu {
		t.Fatalf("應該出 D/T/E 選單，ask=%d", s.charView.ask)
	}
	if _, err := s.Update(key('E')); err != nil {
		t.Fatal(err)
	}
	if got := slotIndexOf(c.EquipIndex); got != slot {
		t.Errorf("裝備後 EquipIndex 應該指到槽 %d，得到 %d", slot, got)
	}
}

// 交給隊友：+0x29 ＝ 0 不檢定，物品直接搬過去（docs/re/131 §7）。
func TestCharViewTrade(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	giver := s.World().Party.Members[0]
	giver.RecordUsed = 0 // 跳過拒絕檢定，讓結果可預期
	if _, err := s.Update(key('1')); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirNone, Action: input.ActionConfirm}); err != nil {
		t.Fatal(err)
	}
	rows := s.cvItemRows()
	if len(rows) == 0 {
		t.Skip("沒東西可給")
	}
	id := giver.Items[rows[0]].ID
	if _, err := s.Update(key('1')); err != nil {
		t.Fatal(err)
	}
	// 可能先問 Reload?（第一件常是武器不會，但保險起見答 N 直到選單）。
	for s.charView.ask != askMenu {
		if _, err := s.Update(key('N')); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Update(key('T')); err != nil {
		t.Fatal(err)
	}
	if s.charView.stage != cvTradeWho {
		t.Fatalf("四個人應該問給誰，stage=%d", s.charView.stage)
	}
	recv := s.World().Party.Members[2]
	has := func(c *game.Character) int {
		n := 0
		for _, sl := range c.Items {
			if sl.ID == id {
				n++
			}
		}
		return n
	}
	gBefore, rBefore := has(giver), has(recv)
	if _, err := s.Update(key('3')); err != nil {
		t.Fatal(err)
	}
	if has(giver) != gBefore-1 || has(recv) != rBefore+1 {
		t.Fatalf("物品 %d 沒搬過去：giver %d→%d、recv %d→%d",
			id, gBefore, has(giver), rBefore, has(recv))
	}
}
