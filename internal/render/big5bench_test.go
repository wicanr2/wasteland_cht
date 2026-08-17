package render

import (
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/lang"
)

// 畫面上每個漢字每一幀都要轉一次 Big5（`DrawRune`）——**這是 UTF-8 那條路
// 唯一新增的執行期成本**，要量不要猜。
func BenchmarkRuneToBig5(b *testing.B) {
	rs := []rune("你們遭到攻擊！選擇逃跑使用雇用迴避攻擊換武器裝填")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range rs {
			lang.CachedRuneToBig5(r)
		}
	}
}
