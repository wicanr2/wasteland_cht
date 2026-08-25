package play

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 地圖 USE Medic：改問「對誰」，只救 CON < 0 的隊員（docs/re/133 §2）。
func TestUseMedicHealsDownedMember(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	healer := s.World().Party.Members[0]
	for i := range healer.Skills {
		if healer.Skills[i].ID == 0 {
			healer.Skills[i] = game.Slot{ID: game.SkillMedic, Value: 3}
			break
		}
	}
	healer.Attributes[game.AttrIQ] = 30
	target := s.World().Party.Members[1]
	target.CON = -8

	send := func(c byte) {
		t.Helper()
		if _, err := s.Update(key(c)); err != nil {
			t.Fatal(err)
		}
	}
	send('U')
	send('1') // 施用者
	send('S')
	// 找 Medic 在清單的第幾項。
	pick := -1
	for i, o := range s.use.options {
		if o.id == game.SkillMedic {
			pick = i
			break
		}
	}
	if pick < 0 {
		t.Fatal("技能清單裡找不到 Medic")
	}
	pickMedic := func() {
		for p := 0; p < pick/usePageSize; p++ {
			send('0') // 翻頁
		}
		send(byte('1' + pick%usePageSize))
	}
	pickMedic()
	if s.use.stage != useStageHealWho {
		t.Fatalf("選 Medic 之後應該問對誰，stage=%d（訊息 %q）", s.use.stage, s.Message())
	}
	// 多試幾次（檢定會失敗）：每次重走整個選單。
	for i := 0; i < 40 && target.CON == -8; i++ {
		if s.use.stage != useStageHealWho {
			send('U')
			send('1')
			send('S')
			pickMedic()
		}
		send('2') // 目標
	}
	if target.CON != -4 {
		t.Fatalf("急救成功要把 −8 折半成 −4，得到 %d", target.CON)
	}
	if s.use.stage != useStageOff {
		t.Fatal("急救之後應該回地圖")
	}
}

// 角色畫面物品頁按 R 重排：點完定案、裝備索引跟著搬；ESC 整包不動。
func TestCharViewReorder(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	c := s.World().Party.Members[0]
	// 收斂成三件好驗證：手槍（裝備）、彈匣、刀。
	c.Items = make([]game.Slot, 30)
	c.Items[0] = game.Slot{ID: 13, Value: 7}
	c.Items[1] = game.Slot{ID: 30, Value: 1}
	c.Items[2] = game.Slot{ID: 4, Value: 1}
	c.EquipIndex = 1

	send := func(in input.Input) {
		t.Helper()
		if _, err := s.Update(in); err != nil {
			t.Fatal(err)
		}
	}
	send(key('1'))                                                    // 角色畫面
	send(input.Input{Dir: input.DirNone, Action: input.ActionConfirm}) // 物品頁
	send(key('R'))
	if s.charView.stage != cvReorder {
		t.Fatalf("按 R 應該進重排，stage=%d", s.charView.stage)
	}
	// ESC ＝ 整包不動。
	send(key('2'))
	send(input.Input{Dir: input.DirNone, Action: input.ActionCancel})
	if c.Items[0].ID != 13 {
		t.Fatal("ESC 取消之後順序不該變")
	}
	// 重來：點 刀(3)、彈匣(2)、手槍(1) → 新順序 4, 30, 13。
	send(key('R'))
	send(key('3'))
	send(key('2')) // 剩下的清單重編號：彈匣現在是第 2 項？點走刀之後剩 手槍(1) 彈匣(2)
	send(key('1')) // 手槍
	if s.charView.stage != cvItems {
		t.Fatalf("點完應該回物品頁，stage=%d", s.charView.stage)
	}
	want := []byte{4, 30, 13}
	for i, id := range want {
		if c.Items[i].ID != id {
			t.Fatalf("槽 %d 應該是 %d，得到 %d", i, id, c.Items[i].ID)
		}
	}
	if c.EquipIndex != 3 {
		t.Errorf("手槍排到第 3 位，裝備索引應該是 3，得到 %d", c.EquipIndex)
	}
}
