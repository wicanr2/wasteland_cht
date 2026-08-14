package rng

import "testing"

// 規格 02 §4 的驗收數列。
//
// 前七個輸出等於二項式係數 C(n+4,5) mod 256——五重進位鏈在還沒有位元組溢位前
// 就是計數器的五重前綴和。**這同時是「初值真的是全零」的獨立佐證**：
// 只要初始化寫錯，第一項就不會是 1。
func TestBinomialHead(t *testing.T) {
	want := []byte{1, 6, 21, 56, 126, 252, 206}
	r := New()
	for i, w := range want {
		if got := r.Next(); got != w {
			t.Fatalf("第 %d 個輸出是 %d，應該是 %d", i+1, got, w)
		}
	}
	// 第 8 項起與二項式分歧（進位開始回饋）：實際 25、二項式 24。
	if got := r.Next(); got != 25 {
		t.Fatalf("第 8 個輸出是 %d，應該是 25", got)
	}
}

func TestRollBoundaries(t *testing.T) {
	r := New()
	if got := r.Roll(0); got != 0 {
		t.Fatalf("Roll(0) ＝ %d，原版是 0", got)
	}
	if got := r.Roll(1); got != 1 {
		t.Fatalf("Roll(1) ＝ %d，原版是 1", got)
	}
}

func TestRollRange(t *testing.T) {
	r := New()
	seen := map[int]bool{}
	for i := 0; i < 20000; i++ {
		v := r.Roll(3)
		if v < 1 || v > 3 {
			t.Fatalf("Roll(3) 回傳 %d", v)
		}
		seen[v] = true
	}
	if len(seen) != 3 {
		t.Fatalf("Roll(3) 只出現 %d 種值", len(seen))
	}
	for i := 0; i < 20000; i++ {
		if v := r.Roll(100); v < 1 || v > 100 {
			t.Fatalf("Roll(100) 回傳 %d", v)
		}
	}
}

func TestD6Distribution(t *testing.T) {
	r := New()
	const n = 600000
	var count [7]int
	for i := 0; i < n; i++ {
		v := r.D6()
		if v < 1 || v > 6 {
			t.Fatalf("D6 回傳 %d", v)
		}
		count[v]++
	}
	for face := 1; face <= 6; face++ {
		ratio := float64(count[face]) / n
		if ratio < 1.0/6*0.97 || ratio > 1.0/6*1.03 {
			t.Fatalf("d6 的 %d 點佔 %.4f，偏差超過 3%%", face, ratio)
		}
	}
}

func TestSumAndPairAverages(t *testing.T) {
	r := New()
	const n = 300000

	total := 0
	for i := 0; i < n; i++ {
		total += r.SumD6(0, 6)
	}
	if avg := float64(total) / n; avg < 20.9 || avg > 21.1 {
		t.Fatalf("SumD6(0,6) 平均 %.2f，應該是 21.0 ± 0.1", avg)
	}

	total = 0
	min := 99
	for i := 0; i < n; i++ {
		v := r.PairD6()
		if v < min {
			min = v
		}
		total += v
	}
	if avg := float64(total) / n; avg < 8.3 || avg > 8.5 {
		t.Fatalf("PairD6 平均 %.2f，應該是 8.4 ± 0.1", avg)
	}
	if min != 3 {
		t.Fatalf("PairD6 的最小值是 %d，應該是 3（1+2）", min)
	}
}

func TestNoStateRepeatIn3M(t *testing.T) {
	r := New()
	seen := make(map[[5]byte]struct{}, 3000000)
	for i := 0; i < 3000000; i++ {
		r.Next()
		s := r.Snapshot()
		if _, dup := seen[s]; dup {
			t.Fatalf("第 %d 次呼叫時狀態重複：%v", i, s)
		}
		seen[s] = struct{}{}
	}
}

func TestAttributeRoll(t *testing.T) {
	r := New()
	const n = 200000
	total, min, max := 0, 99, 0
	for i := 0; i < n; i++ {
		v := r.AttributeRoll()
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		total += v
	}
	if min < 3 || max > 18 {
		t.Fatalf("5d6 取三的值域是 %d–%d，應該落在 3–18", min, max)
	}
	if avg := float64(total) / n; avg < 13.3 || avg > 13.6 {
		t.Fatalf("5d6 取三的平均是 %.2f，應該是 13.43 附近", avg)
	}
}

func TestSnapshotRestore(t *testing.T) {
	r := New()
	for i := 0; i < 100; i++ {
		r.Next()
	}
	saved := r.Snapshot()
	want := []byte{r.Next(), r.Next(), r.Next()}
	r.Restore(saved)
	for i, w := range want {
		if got := r.Next(); got != w {
			t.Fatalf("還原後第 %d 個輸出是 %d，應該是 %d", i, got, w)
		}
	}
}
