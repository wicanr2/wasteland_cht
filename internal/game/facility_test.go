package game

import "testing"

// 驗收 1：價格公式 基礎價 − (基礎價 >> n)。
func TestShopPrice(t *testing.T) {
	for _, tc := range []struct {
		base     uint16
		discount byte
		want     uint16
	}{
		// n ＝ 0 是**原價**：原版 `sub_1C1CC` 在 `dl ＝ 0` 直接 return，
		// 右移迴圈不跑（`docs/re/22` §3 的比例表：0 → 100%）。
		// 讓公式自己算會得到 0，於是指數 0 的店全館免費。
		{100, 0, 100},
		{100, 1, 50}, // 半價
		{100, 2, 75},
		{100, 3, 88},
		{1000, 2, 750},
	} {
		if got := ShopPrice(tc.base, tc.discount); got != tc.want {
			t.Errorf("基礎價 %d、折價指數 %d 應該是 %d，得到 %d",
				tc.base, tc.discount, tc.want, got)
		}
	}
}

// 一筆設施記錄：bit7 ＋ 編號、價格欄位、名稱。
func mkFacility(kind FacilityKind) []byte {
	rec := make([]byte, 32)
	rec[0] = 0x80 | byte(kind)
	rec[shopDiscount] = 1 // 半價
	rec[docHealPer] = 5
	rec[docExam] = 20
	rec[docCure] = 200
	copy(rec[facilityNameAt[kind]:], "Doc Smith\x00")
	return rec
}

func TestParseFacility(t *testing.T) {
	f, ok := ParseFacility(mkFacility(FacilityDoctor))
	if !ok || f.Kind != FacilityDoctor || f.Name != "Doc Smith" {
		t.Fatalf("拆錯了：ok=%v %+v", ok, f)
	}
	// bit7 沒設的是腳本指令，不是設施。
	if _, ok := ParseFacility([]byte{3, 0, 0, 0}); ok {
		t.Fatal("bit7 沒設不該被當成設施")
	}
	// 編號超出 0–4。
	if _, ok := ParseFacility([]byte{0x80 | 9, 0, 0, 0}); ok {
		t.Fatal("編號 9 不該通過")
	}
}

// 驗收 2：療傷 CON 10／MAXCON 28、單價 5 → 90。
func TestHeal(t *testing.T) {
	f, _ := ParseFacility(mkFacility(FacilityDoctor))
	c := &Character{CON: 10, MaxCON: 28, Money: 100}
	points, cost := f.HealCost(c)
	if points != 18 || cost != 90 {
		t.Fatalf("要治 18 點、花 90 元，得到 %d 點／%d 元", points, cost)
	}
	if ok, why := f.Heal(c); !ok {
		t.Fatalf("100 元應該付得起 90：%s", why)
	}
	if c.CON != c.MaxCON || c.Money != 10 {
		t.Fatalf("治完應該滿血、剩 10 元，得到 CON %d／錢 %d", c.CON, c.Money)
	}
	// 滿血的人不需要治。
	if ok, why := f.Heal(c); ok || why != ReasonNoHealNeeded {
		t.Fatalf("滿血不該能治：ok=%v why=%q", ok, why)
	}
	// 錢不夠整筆失敗，不扣款也不回血。
	poor := &Character{CON: 1, MaxCON: 28, Money: 10}
	if ok, why := f.Heal(poor); ok || why != ReasonNoMoney {
		t.Fatalf("錢不夠應該失敗：ok=%v why=%q", ok, why)
	}
	if poor.Money != 10 || poor.CON != 1 {
		t.Fatal("失敗了卻動到錢或血")
	}
}

// 驗收 3：治病只清指定的位元，錢不夠不清。
func TestCure(t *testing.T) {
	f, _ := ParseFacility(mkFacility(FacilityDoctor))
	c := &Character{Money: 500, Status: StatusRabies | StatusSewerRot}

	if ok, why := f.Cure(c, 5); !ok { // bit 5 ＝ Rabies
		t.Fatalf("應該治得起來：%s", why)
	}
	if c.Status != StatusSewerRot {
		t.Fatalf("只該清掉 Rabies，剩下的應該是 %#02x，得到 %#02x",
			uint8(StatusSewerRot), c.Status)
	}
	if c.Money != 300 {
		t.Fatalf("應該扣 200，剩 %d", c.Money)
	}

	if ok, why := f.Cure(c, 5); ok || why != ReasonNoSuchDisease {
		t.Fatalf("已經治好的不該再治：ok=%v why=%q", ok, why)
	}

	broke := &Character{Money: 10, Status: StatusRabies}
	if ok, why := f.Cure(broke, 5); ok || why != ReasonNoMoney {
		t.Fatalf("錢不夠應該失敗：ok=%v why=%q", ok, why)
	}
	if broke.Status != StatusRabies {
		t.Fatal("錢不夠卻把病清掉了")
	}
}

func TestExamAndBuy(t *testing.T) {
	f, _ := ParseFacility(mkFacility(FacilityDoctor))
	c := &Character{Money: 25}
	if ok, _ := f.Exam(c); !ok || c.Money != 5 {
		t.Fatalf("檢查費 20，應該剩 5，得到 %d", c.Money)
	}
	if ok, why := f.Exam(c); ok || why != ReasonNoMoney {
		t.Fatalf("剩 5 元不該付得起 20：ok=%v why=%q", ok, why)
	}

	shop, _ := ParseFacility(mkFacility(FacilityShop))
	buyer := &Character{Money: 100}
	if ok, why := shop.Buy(buyer, 40, 100); !ok { // 半價 → 50
		t.Fatalf("應該買得起：%s", why)
	}
	if buyer.Money != 50 || len(buyer.Items) != 1 || buyer.Items[0].ID != 40 {
		t.Fatalf("買完應該剩 50 元、多一件物品，得到 %d／%v", buyer.Money, buyer.Items)
	}
}
