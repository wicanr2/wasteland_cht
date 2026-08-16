package game

// 自毀倒數 —— 結局的觸發點（`docs/re/100`）。
//
// 結局那一段程式（`0x1B4F0`）掛在設施跳表的第 4 格，而**沒有任何一筆地圖記錄
// 指到第 4 格**（`TestEndingHasNoTriggerInData` 守著這個負面事實）。
// 真正的入口是主迴圈裡的 `sub_1CB30`（`0x16C28` 呼叫，一輪一次）：
//
//	ds:722Ah ＝ 0 → 什麼都不做
//	24-bit 時鐘 ds:4650h–4652h 減掉 ds:722Bh–722Dh ＜ 0xF0 → 什麼都不做
//	否則 al ← 84h；jmp sub_12C80        ← **合成一個 kind 4 的設施分派**
//
// 也就是說「走進結局那一格」這件事在原版裡是**程式自己合成的**，
// 玩家踩的是啟動自毀的那一格，結局在 240 刻之後自己找上門。

// SelfDestructTicks 是啟動到爆炸之間的刻數（`0x1CB51` 的 `cmp dl, 0F0h`）。
const SelfDestructTicks = 0xF0

// selfDestructMask 是比較用的位元寬度：原版做的是 24-bit 減法
// （`ds:4650h`–`ds:4652h`），第四個 byte `ds:722Eh` 不參與。
const selfDestructMask = 0xFFFFFF

// SelfDestruct 是那三個全域的狀態（`ds:722Ah` 與 `ds:722Bh`–`ds:722Dh`）。
//
// **這一組不進存檔**：它們是執行檔資料段的全域，不在存檔的 MSQ 資源裡。
// 倒數期間存檔再讀回來，倒數就沒了——原版就是這樣。
type SelfDestruct struct {
	Armed bool   // ds:722Ah
	At    uint32 // ds:722Bh–722Dh：啟動當下的 Clock.Total
}

// ArmSelfDestruct 是腳本 opcode 35（`0x1AB0E`）做的事。
//
// 除了記下時間，它還把**這張地圖的遭遇分母歸零**（標頭 `+0x2F ← 0`）——
// 自毀啟動之後基地裡不再擲遭遇，玩家逃得出去。
func (w *World) ArmSelfDestruct() {
	w.SelfDestruct = SelfDestruct{Armed: true, At: w.Clock.Total}
	if len(w.Block.Header) > hdrEncounterDenom {
		w.Block.Header[hdrEncounterDenom] = 0
	}
}

// SelfDestructDue 回報倒數到了沒（`sub_1CB30` 的兩道判斷）。
func (w *World) SelfDestructDue() bool {
	if !w.SelfDestruct.Armed {
		return false
	}
	return (w.Clock.Total-w.SelfDestruct.At)&selfDestructMask >= SelfDestructTicks
}

// DisarmSelfDestruct 是結局播完之後的收尾（`0x1B6B0` 的 `ds:722Ah ← 0`）。
func (w *World) DisarmSelfDestruct() { w.SelfDestruct = SelfDestruct{} }
