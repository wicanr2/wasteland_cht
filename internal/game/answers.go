package game

import "bytes"

// 打字回答（nibble 8，`0x15160`，docs/re/46 §4）。
//
// 記錄的形狀：
//
//	+0x00        題目的字串編號；**bit7 ＝ 用單鍵模式**
//	+0x03 起     一串答案的字串編號，最後一個的 bit7 是結束標記
//	之後         每個答案兩個 byte 的後續動作
//
// 這一層只做「數答案」與「比答案」，字串怎麼解碼是 assets 的事。

// answerListStart 是答案清單在記錄裡的起點。
const answerListStart = 0x03

// Question 是一格 nibble 8 的題目。
type Question struct {
	Prompt    byte   // 題目的字串編號（已經去掉 bit7）
	SingleKey bool   // 記錄 +0x00 的 bit7：按一個鍵就好
	Answers   []byte // 答案的字串編號，依序
}

// ParseQuestion 從地圖記錄拆出一題。record 是記錄本體。
//
// 清單以「bit7 設起來」當結束標記——**最後一個的 bit7 是 1**，
// 所以不能拿「bit7 ＝ 0 就停」當條件（那會一個都讀不到）。
func ParseQuestion(record []byte) (Question, bool) {
	var q Question
	if len(record) <= answerListStart {
		return q, false
	}
	q.SingleKey = record[0]&0x80 != 0
	q.Prompt = record[0] & 0x7F

	for i := answerListStart; i < len(record); i++ {
		b := record[i]
		q.Answers = append(q.Answers, b&0x7F)
		if b&0x80 != 0 {
			return q, true
		}
	}
	// 沒讀到結束標記就到記錄尾了——資料壞了，不要假裝讀完。
	return q, false
}

// MatchAnswer 比對玩家輸入與解好的答案，回傳命中的索引。
//
// **照順序試，第一個相等的贏**；全部不中回 len(answers)，
// 也就是原版落到「答錯」那一支的位置（0x15219 的 ds:0A651h）。
//
// 比對是逐 byte 全等（`sub_18D8E`）：沒有大小寫折疊——按鍵在輸入層
// 就轉成大寫了，答案在資料裡本來就是大寫（docs/re/46 §2.1、§3）。
func MatchAnswer(input []byte, answers [][]byte) int {
	for i, a := range answers {
		if bytes.Equal(input, a) {
			return i
		}
	}
	return len(answers)
}

// AnswerBranch 是命中第 n 個答案之後要走的那一格：
// 原版是 `sub_169B1(3 + 答案數 + n × 2)`（0x1522F–0x15242）。
func (q Question) AnswerBranch(n int) int {
	return answerListStart + len(q.Answers) + n*2
}
