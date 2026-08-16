package play

import (
	"sort"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// TestHitChanceSpread 拿出貨資料裡**全部**的敵人資料，對出廠隊伍算一次命中累加值，
// 把分布印出來。
//
// 公式（`docs/re/88`）只有三個變數：隊伍的 Brawling、隊伍的 Agility、
// 對手的行動值。三個裡有兩個來自隊伍，所以**分布窄不窄要看資料**——
// 如果 42 張地圖的敵人算出來全是同一個數字，那這個公式在實際遊戲裡等於沒有作用，
// 而那種「接上了但沒差別」的狀態不會有任何斷言會紅。
func TestHitChanceSpread(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	member := s.World().Party.Members[0]
	if member == nil {
		t.Fatal("出廠隊伍第一個人是 nil")
	}

	hist := map[uint16]int{}
	melee, clamped := 0, 0
	var kinds int

	for _, res := range resources {
		blk, err := rom.BlockByID(res.ID)
		if err != nil {
			continue
		}
		raw, err := blk.EnemyData()
		if err != nil {
			continue
		}
		// ⚠ `EnemyData()` 回的是「位移到區塊尾」，**沒有結尾**。
		// 真正的筆數是標頭 `+0x31`（生成器擲 1..種類數，`docs/re/78` §1），
		// 掃到區塊尾會把後面的資料當成敵人——實測會多出上萬筆、
		// 其中近半算出累加值 0，分布整個被雜訊蓋掉。
		if len(blk.Header) <= 0x31 {
			continue
		}
		kindsHere := int(blk.Header[0x31])
		for k := 1; k <= kindsHere && k*8+8 <= len(raw); k++ {
			off := k * 8
			d := game.ParseEnemyData(raw[off : off+8])
			// 基礎值走遊戲裡真的會用的那一個——沒有移動計畫 ＝ 50
			// （`ds:711Dh`，`docs/re/101` §5）。寫死 60 的話這份分布
			// 描述的是一個沒有人走到的世界。
			acc := game.HitChance(member, game.HitBase(game.NoMovePlan), d)
			hist[acc]++
			kinds++
			if d.Weapon == game.ClassMelee {
				melee++
			}
			if acc == 100 {
				clamped++
			}
		}
	}

	if kinds == 0 {
		t.Fatal("一筆敵人資料都沒讀到")
	}
	var keys []uint16
	for k := range hist {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	t.Logf("%d 筆敵人資料、%d 種累加值（近戰 %d 筆、夾到 100 的 %d 筆）",
		kinds, len(keys), melee, clamped)
	for _, k := range keys {
		t.Logf("  累加值 %3d：%d 筆", k, hist[k])
	}

	// 門檻一：公式要真的分得出差別。全部同一個值 ＝ 接上去等於沒接。
	if len(keys) < 2 {
		t.Errorf("%d 筆敵人只算出 %d 種累加值——命中公式沒有作用",
			kinds, len(keys))
	}
	// 門檻二：不能全部夾在 100（那等於必中）也不能全部是 0（那等於必失手）。
	if clamped == kinds {
		t.Errorf("每一筆都夾在 100 ＝ 對這個人必中")
	}
	if hist[0] == kinds {
		t.Errorf("每一筆都是 0 ＝ 對這個人必失手")
	}
	// 門檻三：出貨資料裡要有近戰類別的敵人，否則 ×4 那條路一次都沒走過，
	// 而它是這個公式裡唯一會溢位的地方。
	if melee == 0 {
		t.Errorf("42 張地圖一筆近戰敵人都沒有——8-bit ×4 那條路沒被走到")
	}
}
