package game

// 傳送與換地圖（第 1 層 nibble ＝ 10），照 docs/spec/07 §6.7。
//
// **沒有這一條玩家出不了起始地圖**：23 筆設施記錄裡只有 2 筆走得進去，
// 商店與醫生全在別的地圖（docs/re/60 §1）。

// TeleportBackMarker 是「回程」的標記：地圖記錄 +0x03 等於它時，
// 目的地是隊伍槽表存起來的那一組，而不是記錄裡的座標。
const TeleportBackMarker = 0xFF

// TeleportTarget 是一次傳送要去哪。
type TeleportTarget struct {
	X, Y  uint8
	MapID uint8
	Back  bool // 這一次是回程（記錄 +0x03 ＝ 0xFF）
}

// Return 是隊伍槽表 +0x0B–+0x0D 存的回程位置。
type Return struct {
	X, Y  uint8
	MapID uint8
}

// ResolveTeleport 依地圖記錄與目前位置算出目的地，並回傳**新的回程值**。
//
// 原版的順序是「先把現在在哪存進槽表，再決定要去哪」（docs/re/60 §2），
// 所以回程永遠指向踩上傳送格的那一格——**兩件事不能顛倒**，
// 顛倒的話回程會指到目的地本身，玩家會卡在原地來回。
func ResolveTeleport(rec []byte, here Return, back Return) (TeleportTarget, Return) {
	newBack := here // 踩上去就把現在的位置存成回程
	if len(rec) < 4 {
		return TeleportTarget{X: here.X, Y: here.Y, MapID: here.MapID}, newBack
	}
	if rec[3] == TeleportBackMarker {
		return TeleportTarget{X: back.X, Y: back.Y, MapID: back.MapID, Back: true}, newBack
	}
	// bit7 設起來時 +0x01／+0x02 是相對目前座標的位移，否則相對原點。
	var baseX, baseY uint8
	if rec[0]&0x80 != 0 {
		baseX, baseY = here.X, here.Y
	}
	return TeleportTarget{
		X:     baseX + rec[1],
		Y:     baseY + rec[2],
		MapID: rec[3],
	}, newBack
}
