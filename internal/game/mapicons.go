package game

// 地圖疊圖的選擇規則（`sub_18024`，docs/re/48 §2.1、docs/spec/03 §2.9）。
//
// 這一層只回「這一格該畫幾號」，畫的動作在 internal/render。
// 分成兩層是因為條件裡有時鐘與記錄，那些是規則層的事。

// IconCount 是疊圖與圖磚的分界：0–9 是 IC0_9.WLF 的十張，
// ≥ 10 是圖磚編號 `值 − 10`。原版寫成 `cmp al, 0Ah`。
const IconCount = 10

// 疊圖編號（與 render 的同名常數同一套；這裡重列是為了不讓規則層依賴呈現層）。
const (
	IconBlack      = 0
	IconRobot      = 1
	IconCyborg     = 2
	IconMutant     = 3
	IconHumanoid   = 4
	IconLoot       = 5
	IconAnimal     = 6
	IconParty      = 7
	IconRadiation  = 8
	IconOtherGroup = 9
)

// kindIcon 是 `ds:AA17h`：敵人種類 → 疊圖編號。
//
// 原始 bytes `00 06 03 04 02 01`（線性 0x27837，docs/re/48 §3）。
// 第 0 項是佔位——42 個區塊 397 筆敵人資料的種類全部落在 1–5。
var kindIcon = [6]byte{0, IconAnimal, IconMutant, IconHumanoid, IconCyborg, IconRobot}

// KindIcon 回某個敵人種類在地圖上畫成哪一張。
//
// ⚠ 種類超出 1–5 時回 IconHumanoid，這是**原版的預設**（`sub_14664` 的
// `mov bl, 3` 走在讀記錄之前），不是我們補的保險。
func KindIcon(k EnemyKind) byte {
	if int(k) < len(kindIcon) && k != 0 {
		return kindIcon[k]
	}
	return IconHumanoid
}

// RadiationVisible 回報這個時刻看不看得到輻射標誌。
//
// `sub_18024` 的 `cmp al, 6` / `cmp al, 12h` 與時鐘的晝夜門檻是同一組數字，
// 所以這裡直接沿用 DawnHour／DuskHour，不另立一份會漂移的常數。
// 白天（6 ≤ 時 < 18）那些格子畫成一般地形，**標誌完全不出現**——
// 實機在 05:56 有、06:00 沒有，同樣那幾格（docs/re/48 §4）。
func RadiationVisible(hour int) bool { return hour < DawnHour || hour >= DuskHour }

// VisibleIcon 是視窗裡某一格要疊的圖：格座標（視窗內的 col/row）＋ 編號。
type VisibleIcon struct {
	Col, Row int
	Icon     byte
}

// ViewIcons 掃整個地圖視窗，回所有要疊圖的格。
//
// 原版是 `sub_167CE` 每走一步重畫 nibble 為 4／5／9 的格子（docs/re/26 §5）；
// 這裡一次算完，畫的順序與原版一樣是**先地形後疊圖**。
// 隊伍與其他分隊不在這裡——那兩個不是由地圖格決定的。
func (w *World) ViewIcons() []VisibleIcon {
	var out []VisibleIcon
	for row := 0; row < ViewRows; row++ {
		for col := 0; col < ViewCols; col++ {
			x, y := w.ViewX+col, w.ViewY+row
			if x < 0 || y < 0 || x >= w.Block.Dim || y >= w.Block.Dim {
				continue
			}
			terrain, _, _, err := w.Block.At(x, y)
			if err != nil {
				continue
			}
			var rec1 byte
			if terrain == 4 {
				// nibble 4 才需要記錄；其他型別不去讀，省得無謂的錯誤。
				if rec, _, err := w.Block.CellRecord(x, y); err == nil && len(rec) > 1 {
					rec1 = rec[1]
				}
			}
			if icon, ok := CellIcon(terrain, rec1, int(w.Clock.Hour)); ok {
				out = append(out, VisibleIcon{Col: col, Row: row, Icon: icon})
			}
		}
	}
	return out
}

// CellIcon 回這一格要疊哪一張圖。
//
// terrain 是第 1 層的 nibble（section 型別），rec1 是該格記錄的 `+0x01`
// （只有 nibble 4 用得到，其餘傳什麼都不影響），hour 是遊戲時鐘的「時」。
//
// 回傳 ok 為 false 表示**這一格不疊任何東西**，照一般圖磚畫就好。
func CellIcon(terrain, rec1 byte, hour int) (icon byte, ok bool) {
	switch terrain {
	case 5: // 寶箱／掉落物
		return IconLoot, true
	case 9: // 輻射區
		if RadiationVisible(hour) {
			return IconRadiation, true
		}
		return 0, false
	case 4:
		// 記錄 +0x01 去掉 bit7；< 10 才是疊圖，≥ 10 是圖磚編號。
		if v := rec1 & 0x7F; v < IconCount {
			return v, true
		}
		return 0, false
	}
	return 0, false
}
