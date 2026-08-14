// Package rng 是原版的亂數產生器與擲骰層（docs/spec/02-rng-and-dice.md）。
//
// 產生器是五個位元組的進位鏈（`sub_18E6B`）：一個 inc、一個 add、四個 adc，
// **沒有乘、除或移位**。四支擲骰函式全部建立在它上面。
//
// 這一層不認識畫面也不認識資產，可以單獨測。
package rng

import "fmt"

// State 對應原版的 ds:465Ch–4660h 五個位元組。
//
// 映像裡的初值是全零，而且全檔沒有設種子的程式碼——熵來自鍵盤輪詢的次數
// （每輪詢一次就推進一次）。零值的 State 就是原版剛開機的狀態。
type State struct {
	s [5]byte
}

// New 回傳原版開機時的狀態（全零）。
func New() *State { return &State{} }

// Next 推進一次並回傳一個位元組。
func (r *State) Next() byte {
	r.s[0]++
	a := r.s[0]
	carry := byte(0)
	for i := 1; i < 5; i++ {
		sum := uint16(a) + uint16(r.s[i]) + uint16(carry)
		a = byte(sum)
		carry = byte(sum >> 8)
		r.s[i] = a
	}
	return a
}

// Snapshot 取出五個位元組，Restore 放回去。
//
// ⚠ 原版的存檔**不含**這五個位元組（docs/re/27 §7），讀檔之後 RNG 是接著
// 當時記憶體裡的狀態跑。remake 若把狀態存進存檔，是**刻意偏離原版**的設計決定，
// 要在存檔格式文件裡標明，不要當成還原原版行為。
func (r *State) Snapshot() [5]byte { return r.s }
func (r *State) Restore(s [5]byte) { r.s = s }

// D6 回傳 1..6。
//
// 取低三位得 0..7，6 與 7 直接重抽——是拒絕取樣不是取模，所以六面等機率。
func (r *State) D6() int {
	for {
		v := r.Next() & 7
		if v < 6 {
			return int(v) + 1
		}
	}
}

// Roll 回傳 1..n（閉區間、等機率）。
//
// ⚠ 邊界照原版：**n ≤ 1 時原樣回傳 n**，所以 Roll(0) ＝ 0、Roll(1) ＝ 1。
// 呼叫端有依賴這個行為（docs/re/13 §3.2）。
func (r *State) Roll(n int) int {
	if n <= 1 {
		return n
	}
	if n > 255 {
		panic(fmt.Sprintf("Roll 的面數是 8-bit，收到 %d", n))
	}
	mask := byte(0)
	for v := n; v != 0; v >>= 1 {
		mask = mask<<1 | 1 // 不小於 n 的最小 2^k−1
	}
	for {
		v := r.Next() & mask
		if int(v) < n {
			return int(v) + 1
		}
	}
}

// SumD6 回傳 base ＋ n 顆 d6。
//
// 原版的累加器是 8-bit 低位 ＋ 8-bit 高位（`sub_19D86`），所以結果是 16-bit；
// 這裡用 int 但夾在 16-bit，避免顆數大時與原版分歧。
func (r *State) SumD6(base, n int) int {
	total := base
	for i := 0; i < n; i++ {
		total += r.D6()
	}
	return total & 0xFFFF
}

// PairD6 擲一對 d6 並累加，**同點就再擲一對**（`sub_19C84`）。
//
// 最小值 3（1+2），理論期望 8.4。這是規則層的通用檢定骰，11 個呼叫端。
func (r *State) PairD6() int {
	sum := 0
	for {
		a, b := r.D6(), r.D6()
		sum += a + b
		if a != b {
			return sum
		}
	}
}

// AttributeRoll 是角色建立用的「5d6 取最高三顆」（`sub_1CAD1`），值域 3–18。
func (r *State) AttributeRoll() int {
	var dice [5]int
	for i := range dice {
		dice[i] = r.D6()
	}
	// 只有五顆，插入排序最直接。
	for i := 1; i < len(dice); i++ {
		for j := i; j > 0 && dice[j] > dice[j-1]; j-- {
			dice[j], dice[j-1] = dice[j-1], dice[j]
		}
	}
	return dice[0] + dice[1] + dice[2]
}
