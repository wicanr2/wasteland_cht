package game

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// 輻射格（第 1 層 nibble ＝ 9）的結算，照 docs/spec/07 §6.4。
//
// 原版在 0x14410：印記錄 +0x00 那條訊息，然後**對每個隊員**擲
// 記錄 +0x01 顆 d6 扣 CON，並加上 Radiation poisoning 的狀態位元。

// RadiationHit 是一個隊員這一次受到的輻射傷害。
type RadiationHit struct {
	Member  int  // 隊伍裡的序號
	Rolled  int  // 擲出來的傷害
	Absorb  int  // 護甲吸收（無視護甲時是 0）
	Applied int  // 實際扣掉的 CON
	Immune  bool // 穿著 Rad suit，整個跳過
}

// RadiationBypassesArmour 回報這一格的傷害無不無視護甲。
//
// ⚠ 判準是**記錄 +0x00（訊息編號）的 bit0**，不是輻射的通則——
// 211 個輻射格裡 84 格無視、127 格照常吸收（docs/re/55 §2、§4）。
// 為什麼是訊息編號的奇偶，原版沒有交代；**照抄，不要統一成其中一種**。
func RadiationBypassesArmour(rec []byte) bool {
	return len(rec) > 0 && rec[0]&1 == 1
}

// ItemRadSuit 是 Rad suit 的物品編號（原版 0x14432 的 `cmp al, 29h`，
// 物品表第 41 筆就叫 `Rad suit`）。
//
// ⚠ 判準是**裝備中的護甲是不是這一件**，不是背包裡有沒有——
// 原版取的是護甲槽（+0x25）指到的那一格（docs/re/55 §3）。
const ItemRadSuit = 41

// ImmuneToRadiation 回報這個人穿的護甲擋不擋輻射。
func (c *Character) ImmuneToRadiation() bool {
	if c.ArmorIndex == 0 {
		return false // 沒穿護甲：原版的 sub_196C4 回 0，直接受傷
	}
	i := int(c.ArmorIndex)
	if i >= len(c.Items) {
		return false
	}
	return c.Items[i].ID == ItemRadSuit
}

// RadiationDice 是這一格要擲幾顆 d6（記錄 +0x01）。
func RadiationDice(rec []byte) byte {
	if len(rec) < 2 {
		return 0
	}
	return rec[1]
}

// ApplyRadiation 對整隊結算一次輻射傷害，回傳每個人的結果。
//
// 每個人都會拿到 Radiation poisoning 的狀態位元，**扣多少血與有沒有中毒無關**。
func (w *World) ApplyRadiation(rec []byte) []RadiationHit {
	n := RadiationDice(rec)
	bypass := RadiationBypassesArmour(rec)
	out := make([]RadiationHit, 0, len(w.Party.Members))
	for i, c := range w.Party.Members {
		// 穿著 Rad suit 的人整個跳過——不扣血也不中毒（原版 0x14432 的 jz）。
		if c.ImmuneToRadiation() {
			out = append(out, RadiationHit{Member: i, Immune: true})
			continue
		}
		hit := RadiationHit{Member: i, Rolled: rollD6(w.RNG, int(n))}
		if !bypass {
			hit.Absorb = Absorb(w.RNG, c.AC)
		}
		if v := hit.Rolled - hit.Absorb; v > 0 {
			hit.Applied = v
			c.PreHurt = c.CON
			c.CON -= int16(v)
		}
		c.Status |= StatusRadiation
		out = append(out, hit)
	}
	return out
}

// rollD6 是 0 ＋ n 顆 d6；沒有亂數源時回 0（無頭測試會這樣）。
func rollD6(r *rng.State, n int) int {
	if r == nil {
		return 0
	}
	return r.SumD6(0, n)
}
