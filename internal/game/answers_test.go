package game

import "testing"

// 答案清單的結束標記是「bit7 設起來」——**最後一個**才有。
// 拿「bit7 ＝ 0 就停」當條件會一個都讀不到，所以這裡釘死方向。
func TestParseQuestionEndMarker(t *testing.T) {
	// +0x00 bit7 沒設 ＝ 打字模式；答案是字串 20、21、22（最後一個帶 bit7）
	rec := []byte{0x0A, 0, 0, 20, 21, 22 | 0x80, 0x33, 0x44}
	q, ok := ParseQuestion(rec)
	if !ok {
		t.Fatal("應該讀得完")
	}
	if q.SingleKey {
		t.Fatal("bit7 沒設應該是打字模式")
	}
	if q.Prompt != 0x0A {
		t.Fatalf("題目字串編號 %d，應該是 10", q.Prompt)
	}
	want := []byte{20, 21, 22}
	if len(q.Answers) != len(want) {
		t.Fatalf("答案數 %d（%v），應該是 %d", len(q.Answers), q.Answers, len(want))
	}
	for i := range want {
		if q.Answers[i] != want[i] {
			t.Fatalf("第 %d 個答案是 %d，應該是 %d", i, q.Answers[i], want[i])
		}
	}
}

// 單鍵模式由 +0x00 的 bit7 決定。
func TestParseQuestionSingleKey(t *testing.T) {
	q, ok := ParseQuestion([]byte{0x0A | 0x80, 0, 0, 5 | 0x80})
	if !ok || !q.SingleKey {
		t.Fatalf("bit7 設起來應該是單鍵模式：%+v ok=%v", q, ok)
	}
	if q.Prompt != 0x0A {
		t.Fatalf("題目編號要去掉 bit7，得到 %d", q.Prompt)
	}
}

// 沒有結束標記就到記錄尾 ＝ 資料壞了，要回 false 而不是假裝讀完。
func TestParseQuestionUnterminated(t *testing.T) {
	if _, ok := ParseQuestion([]byte{1, 0, 0, 20, 21, 22}); ok {
		t.Fatal("沒有結束標記應該回 false")
	}
	if _, ok := ParseQuestion([]byte{1, 0, 0}); ok {
		t.Fatal("記錄太短應該回 false")
	}
}

// 照順序試，第一個相等的贏；全部不中回 len(answers)（＝ 答錯那一支）。
func TestMatchAnswer(t *testing.T) {
	answers := [][]byte{[]byte("BIRD"), []byte("DIPSTICK"), []byte("MUERTE")}
	if got := MatchAnswer([]byte("DIPSTICK"), answers); got != 1 {
		t.Fatalf("命中索引 %d，應該是 1", got)
	}
	if got := MatchAnswer([]byte("NOPE"), answers); got != len(answers) {
		t.Fatalf("答錯應該回 %d，得到 %d", len(answers), got)
	}
	// 逐 byte 全等：大小寫不折疊（輸入層已經轉大寫了）。
	if got := MatchAnswer([]byte("bird"), answers); got != len(answers) {
		t.Fatal("小寫不該命中——原版沒有大小寫折疊，轉大寫是輸入層做的")
	}
	// 前綴不算命中。
	if got := MatchAnswer([]byte("BIR"), answers); got != len(answers) {
		t.Fatal("前綴不該命中")
	}
}

// 分支位移 ＝ 3 + 答案數 + n × 2（0x1522F–0x15242）。
func TestAnswerBranch(t *testing.T) {
	q := Question{Answers: []byte{20, 21, 22}}
	for n, want := range map[int]int{0: 6, 1: 8, 2: 10, 3: 12} {
		if got := q.AnswerBranch(n); got != want {
			t.Fatalf("第 %d 個答案的分支是 %d，應該是 %d", n, got, want)
		}
	}
}

// 改寫地圖格的兩個 byte（docs/re/46 §4.1）。
func TestParseCellPatch(t *testing.T) {
	rec := []byte{0, 0, 0, 0x05, 0x21, 0x83, 0x44, 0xFE, 0x00, 0xFD, 0x00}

	p, reuse, ok := ParseCellPatch(rec, 3)
	if !ok || reuse {
		t.Fatalf("一般的兩個 byte 應該直接拆：ok=%v reuse=%v", ok, reuse)
	}
	if p.Skip || p.Terrain != 5 || p.Record != 0x21 {
		t.Fatalf("拆出來是 %+v，應該是 {5, 0x21, false}", p)
	}

	// bit7 設起來 ＝ 這一支什麼都不改。
	p, _, ok = ParseCellPatch(rec, 5)
	if !ok || !p.Skip {
		t.Fatalf("bit7 設起來應該是 Skip：%+v ok=%v", p, ok)
	}
	// ⚠ Terrain 只去掉 bit7，不遮成 4 bits——0x83 的低 7 位是 3。
	if p.Terrain != 3 {
		t.Fatalf("Terrain 應該是 0x83 & 0x7F ＝ 3，得到 %d", p.Terrain)
	}

	// 0xFE／0xFD 是「沿用上一次」的特例，交給呼叫端。
	for _, at := range []int{7, 9} {
		if _, reuse, ok := ParseCellPatch(rec, at); !ok || !reuse {
			t.Fatalf("位移 %d 應該回 reuse：ok=%v reuse=%v", at, ok, reuse)
		}
	}

	// 位移超出記錄要回 false，不要讀到別人的 byte。
	if _, _, ok := ParseCellPatch(rec, len(rec)-1); ok {
		t.Fatal("只剩一個 byte 應該回 false")
	}
}
