package play

// 地點招牌的中文（`translations/zh-Hant/places.tsv`）。
//
// **這些字不在字串表裡。** 設施的招牌來自地圖記錄裡的明文 ASCII
// （`docs/re/29` §5.4），畫面開場那一句地點名來自存檔的全域狀態
// （`docs/re/30` 的 `gblPlace`）。兩者都是原版直接寫在資料裡的英文，
// 抽字串的工具看不到它們，所以中文化走「**顯示時查表**」：
// 拿英文原名當 key 去翻譯目錄查，查不到就照原樣顯示英文。
//
// ⚠ **不改玩家的存檔。** 另一條路是把中文寫回存檔那 16 個 byte，
// 但存檔策略是改寫不是重建（`CLAUDE.md` §4），而且那份資料玩家還要拿去
// 跑原版。查表這條路一個 byte 都不碰。

import (
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/lang"
)

// placeCJK 把地點的英文原名換成中文（Big5）。查不到回 nil。
//
// 查不到**不是錯誤**：資料裡的地點名還沒盤點完（劇情改寫會生出新的），
// 沒有譯文時畫面上顯示原本那串英文，遊戲照跑。
func (s *Scene) placeCJK(name string) string {
	// ⚠ **資料裡的招牌是補過空白的定長欄位**：地圖記錄與存檔裡的
	// `Ranger Ctr.` 實際上是 `"Ranger Ctr. "`（尾端一個空白）。
	// 不修掉就永遠查不到，而且**查不到是靜靜顯示英文**，看起來像「還沒翻」。
	name = strings.TrimSpace(name)
	if s.cat == nil || name == "" {
		return ""
	}
	if b, ok := s.cat.Lookup(lang.PlaceKey(name)); ok {
		return b
	}
	return ""
}
