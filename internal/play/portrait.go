package play

// 肖像框：一張圖 ＋ 一行 12 格的說明（`docs/re/115`）。
//
// 原版沒有「戰鬥的肖像框」這個東西——它畫的是 `ds:46FAh`（目前這張圖）
// 與 `ds:7201h`（13 bytes 的說明），兩個都是**跨畫面模式活著**的全域，
// 戰鬥只是最後一個寫它們的人。所以這一支不叫 `combatPortrait`。

import (
	"unicode/utf8"

	"github.com/wicanr2/wasteland_cht/internal/render"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

const (
	// rangerPicture 是「一組敵人都挑不到」時的那張圖（`0x12665` 的 `al ← 8`）。
	rangerPicture = 8
	// strRangerName 是同一支分支印的字串（`0x12660` 的 `al ← 0x60`）＝ `Ranger`。
	strRangerName = 96
	// portraitCaptionCells 是說明字串的格數（`0x12693` 填 12 格，`ds:720Dh` 收尾）。
	portraitCaptionCells = 12
)

// portraitCaption 回這一格要印的說明（英文、中文各一份）。
//
// 誰的名字：**離隊伍最近的那一組敵人**（`sub_126D7` 取 `[ds:46C8h+3]` 的最小值）。
// remake 還沒有敵人在地圖上移動（`docs/re/87`），每一組的距離不存在，
// 所以取第一組還活著的——只有一組時兩者等價。
//
// ⚠ 一組都挑不到時**不是留白**：原版印 `Ranger`（字串 96）並載圖 8。
// 空地遭遇（`ENC` 問一句就進指令階段）走的正是這一支，實機截圖上看到的
// 遊俠圖與 `Ranger` 就是它（`docs/re/105` §2 的那張）。
func (s *Scene) portraitCaption() (en, cjk string) {
	if s.combat == nil {
		return "", ""
	}
	if e, _ := s.combat.firstEnemy(); e != nil {
		return s.combat.enemyLabel(e), s.combat.zhEnemy(e)
	}
	return s.exeString(strRangerName),
		s.cjkExe(exeTable1, strRangerName, textlayout.Options{})
}

// portraitPicture 回這一格要畫哪一張圖；-1 ＝ 這個模式沒有圖。
//
// ⚠ **與說明字串走同一個判斷**（`firstEnemy`）。兩邊各記一份的話，
// 敵人死光之後圖還是敵人、字已經變成遊俠——而畫面上看起來完全正常。
func (s *Scene) portraitPicture() int {
	if s.combat == nil {
		return -1
	}
	if e, _ := s.combat.firstEnemy(); e != nil {
		return int(e.Data.Portrait)
	}
	return rangerPicture
}

// captionCol 是說明字串置中之後的起點（`0x126A5`–`0x126BA`）。
//
// ⚠ 置中是**照格數**算的，不是 byte 數：一個中文字兩個 byte、只佔一格。
// 太長就不置中（原版是 `cmp al, 0x0C` 之後直接用 0）。
func captionCol(text string) int {
	n := utf8.RuneCountInString(text)
	if n >= portraitCaptionCells {
		return render.FacilityNameCol
	}
	return render.FacilityNameCol + (portraitCaptionCells/2 - n/2)
}

// fitCaption 把中文說明塞進 12 格。
//
// 譯名的體例是「中文（英文）」（`translations/zh-Hant/monsters.tsv`），
// 一定比原文長。放不下就**先退成只有中文那一段**，還是放不下才截斷——
// 直接截的話「變種人（Mutant）」會變成「變種人（Muta」，
// 而畫面上看起來只是「這個名字比較長」。
func fitCaption(zh string) string {
	if utf8.RuneCountInString(zh) <= portraitCaptionCells {
		return zh
	}
	if i := indexRune(zh, '（'); i > 0 {
		if head := zh[:i]; utf8.RuneCountInString(head) <= portraitCaptionCells {
			return head
		}
	}
	return clipCells(zh, portraitCaptionCells)
}

// indexRune 回第一個 r 的 byte 位置；沒有回 -1。
func indexRune(s string, r rune) int {
	for i, c := range s {
		if c == r {
			return i
		}
	}
	return -1
}
