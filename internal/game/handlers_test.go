package game

import "testing"

func items(vals ...byte) []byte {
	out := make([]byte, ItemSlots)
	copy(out, vals)
	return out
}

// 驗收 4：物品陣列只認 30 槽，值 0 是空槽。
func TestCountAndFindItems(t *testing.T) {
	it := items(5, 0, 7, 0, 5)
	if n := CountItems(it, nil); n != 3 {
		t.Errorf("應該數到 3 件，得到 %d", n)
	}
	if n := CountItems(it, func(v byte) bool { return v == 5 }); n != 2 {
		t.Errorf("過濾之後應該是 2 件，得到 %d", n)
	}
	if n := CountItems(items(), nil); n != 0 {
		t.Errorf("空陣列應該是 0 件，得到 %d", n)
	}
	if slot, ok := FindItem(it, 7); !ok || slot != 2 {
		t.Errorf("7 應該在第 2 槽，得到 %d（ok=%v）", slot, ok)
	}
	if _, ok := FindItem(it, 9); ok {
		t.Error("陣列裡沒有 9")
	}
	// 超過 30 槽的部分不算——原版的迴圈到 0xF9 就停了。
	long := make([]byte, ItemSlots+5)
	for i := range long {
		long[i] = 1
	}
	if n := CountItems(long, nil); n != ItemSlots {
		t.Errorf("最多只該數 %d 槽，得到 %d", ItemSlots, n)
	}
	if _, ok := FindItem(append(make([]byte, ItemSlots), 3), 3); ok {
		t.Error("第 30 槽之後的東西不該找得到")
	}
}

// 驗收 3：Weapon 空手時不開選單。
func TestHandleWeapon(t *testing.T) {
	opened := false
	res := HandleWeapon(items(), func() (byte, bool) { opened = true; return 0, true })
	if res.Accepted || opened {
		t.Fatalf("空手時不該開選單也不該接受：%+v（開了＝%v）", res, opened)
	}
	if res.Message != "exe:1:64" {
		t.Errorf("應該留字串 64 的訊息，得到 %q", res.Message)
	}

	res = HandleWeapon(items(0, 0, 9), func() (byte, bool) { return 2, true })
	if !res.Accepted || res.Arg != 2 {
		t.Errorf("有東西就該接受第 2 槽，得到 %+v", res)
	}
	if res := HandleWeapon(items(9), func() (byte, bool) { return 0, false }); res.Accepted {
		t.Error("選單取消時應該重問")
	}
}

// 驗收 5：Load 的三道檢查照順序，而且第一道不留訊息。
func TestHandleLoadGates(t *testing.T) {
	for _, tc := range []struct {
		name     string
		armed    bool
		ammoType byte
		items    []byte
		accepted bool
		message  string
	}{
		{"沒裝備武器 → 靜靜結束", false, 3, items(3), false, ""},
		{"武器不吃彈匣", true, 0, items(3), false, "exe:1:66"},
		{"沒有彈匣", true, 3, items(9), false, "exe:1:65"},
		{"三道都過", true, 3, items(9, 3), true, ""},
	} {
		res := HandleLoad(tc.armed, tc.ammoType, tc.items)
		if res.Accepted != tc.accepted || res.Message != tc.message {
			t.Errorf("%s：得到 %+v，預期 accepted=%v message=%q",
				tc.name, res, tc.accepted, tc.message)
		}
	}
	// 接受時回傳的是彈匣在第幾槽。
	if res := HandleLoad(true, 3, items(9, 3)); res.Arg != 1 {
		t.Errorf("彈匣在第 1 槽，得到 %d", res.Arg)
	}
}

// 驗收 2：前提不成立就重問。
func TestHandleHire(t *testing.T) {
	res := HandleHire(0, func() (byte, bool) { return 0, true })
	if res.Accepted || res.Message != "exe:1:57" {
		t.Errorf("沒有對象時應該留字串 57 並重問，得到 %+v", res)
	}
	if res := HandleHire(2, func() (byte, bool) { return 1, true }); !res.Accepted || res.Arg != 1 {
		t.Errorf("有對象就該接受，得到 %+v", res)
	}
	if res := HandleHire(2, func() (byte, bool) { return 0, false }); res.Accepted || res.Message != "" {
		t.Errorf("取消時應該重問且不留訊息，得到 %+v", res)
	}
}

// 驗收 6：Use 那一格以角色編號為索引。
func TestHandleUseIndexesByCharacter(t *testing.T) {
	choices := UseChoices{}
	if res := HandleUse(choices, 5, func() (byte, bool) { return 0x42, true }); !res.Accepted {
		t.Fatalf("選了就該接受，得到 %+v", res)
	}
	if choices[5] != 0x42 {
		t.Errorf("應該記在角色 5 那一格，得到 %v", choices)
	}
	// 隊伍位置換了不影響——這一層只認角色編號。
	if _, ok := choices[0]; ok {
		t.Error("不該以隊伍位置當索引")
	}
	if res := HandleUse(choices, 5, func() (byte, bool) { return 0, false }); res.Accepted {
		t.Error("取消時應該重問")
	}
	if choices[5] != 0x42 {
		t.Error("取消不該蓋掉原本那一格")
	}
	// nil 也要安全。
	if res := HandleUse(nil, 1, func() (byte, bool) { return 7, true }); !res.Accepted {
		t.Error("choices 為 nil 時不該 panic，也該照樣接受")
	}
}

// 驗收 1：四支都不改任何狀態（Use 寫的那一格除外）。
func TestHandlersDoNotMutate(t *testing.T) {
	before := items(3, 9, 0, 4)
	snapshot := append([]byte(nil), before...)

	HandleWeapon(before, func() (byte, bool) { return 0, true })
	HandleLoad(true, 3, before)
	HandleHire(1, func() (byte, bool) { return 0, true })

	for i := range snapshot {
		if before[i] != snapshot[i] {
			t.Fatalf("處理程式動到了物品陣列：第 %d 槽 %d → %d", i, snapshot[i], before[i])
		}
	}
}
