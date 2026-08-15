package game

import "github.com/wicanr2/wasteland_cht/internal/game/rng"

// 遭遇的生成與距離（docs/spec/13、docs/re/37）。
//
// 規格 12 收的是「已經建好的敵人」；這一層補上前一步——敵人從哪裡來、血量怎麼定。

// 敵方記錄的版面（docs/re/37 §1）。94 = 4 + 3 × 30。
const (
	EnemyRecordSize = 0x5E
	EnemyHeaderSize = 4
	EnemyGroupSize  = 0x1E
)

// 地圖記錄裡三組敵人的欄位位移（ds:A5B1h ＝ 型別、ds:A5AEh ＝ 數量）。
var (
	enemyTypeAt  = [EnemyGroups]int{0x03, 0x05, 0x07}
	enemyCountAt = [EnemyGroups]int{0x04, 0x06, 0x08}
)

// EncounterHeader 是敵方記錄的 4 bytes 標頭（docs/re/37 §2.1）。
type EncounterHeader struct {
	X, Y     byte // +0x00/+0x01，這場遭遇在地圖上的位置
	Resource byte // +0x02，資源（地圖）編號
	Distance byte // +0x03，與隊伍的距離（10 × 歐氏）
}

// Present 回報這一格是不是真的有遭遇。原版全檔十幾處都先測 +0x00 再往下做。
func (h EncounterHeader) Present() bool { return h.X != 0 }

// ParseEncounterHeader 拆敵方記錄的標頭。
func ParseEncounterHeader(b []byte) EncounterHeader {
	return EncounterHeader{X: b[0], Y: b[1], Resource: b[2], Distance: b[3]}
}

// SpawnGroup 是一組敵人的來源：型別與數量都取自地圖記錄。
type SpawnGroup struct {
	Type    byte
	Count   int
	Clamped bool // 數量超過 EnemiesPerGroup，被夾住了
}

// ReadSpawnGroups 從地圖記錄讀出三組的型別與數量（docs/spec/13 §3）。
//
// ⚠ 原版沒有上限檢查——數量寫超過 10 就會寫進下一組的格子。
// 這裡夾在 10 並用 Clamped 回報，**不靜靜溢出**；
// 原版資料是不是保證 ≤ 10 沒有驗過，所以不能假設這個旗標永遠是 false。
func ReadSpawnGroups(record []byte) [EnemyGroups]SpawnGroup {
	var out [EnemyGroups]SpawnGroup
	for g := 0; g < EnemyGroups; g++ {
		if enemyCountAt[g] >= len(record) || enemyTypeAt[g] >= len(record) {
			continue
		}
		n := int(record[enemyCountAt[g]])
		if n > EnemiesPerGroup {
			n, out[g].Clamped = EnemiesPerGroup, true
		}
		out[g] = SpawnGroup{Type: record[enemyTypeAt[g]], Count: n, Clamped: out[g].Clamped}
	}
	return out
}

// EnemyTable 是一張地圖的敵人資料表（8 bytes 一筆）。
type EnemyTable []EnemyData

// ParseEnemyTable 把一段連續的 bytes 拆成敵人資料表。
// 尾巴不足 8 bytes 的殘料丟掉——表長由呼叫者決定，這裡不猜。
func ParseEnemyTable(b []byte) EnemyTable {
	t := make(EnemyTable, 0, len(b)/8)
	for at := 0; at+8 <= len(b); at += 8 {
		t = append(t, ParseEnemyData(b[at:at+8]))
	}
	return t
}

// Lookup 取第 n 筆。越界或第 0 筆（＝ 沒有敵人）回 false。
func (t EnemyTable) Lookup(n byte) (EnemyData, bool) {
	if int(n) >= len(t) {
		return EnemyData{}, false
	}
	d := t[n]
	if d.Empty() {
		return EnemyData{}, false
	}
	return d, true
}

// Spawn 依地圖記錄建出一場遭遇的敵人，直接填進 Battle 的格子（docs/spec/13）。
//
// 回傳建出來的敵人總數。型別 0（或查不到）與數量 0 的組都跳過，
// **不會補一隻上去湊數**。
func (b *Battle) Spawn(record []byte, table EnemyTable, r *rng.State) int {
	total := 0
	for g, sg := range ReadSpawnGroups(record) {
		if sg.Count == 0 {
			continue
		}
		data, ok := table.Lookup(sg.Type)
		if !ok {
			continue
		}
		for i := 0; i < sg.Count; i++ {
			// 每一隻各擲各的——同組同型別的血量不一樣（docs/re/37 §3）。
			b.AddEnemy(g, i, &Enemy{HP: data.RollHP(r), Data: data})
			total++
		}
	}
	return total
}

// 距離表的形狀（ds:CD0Dh，docs/re/37 §4）：5 列 × 10 行，正好是地圖視窗的半徑。
const (
	DistanceMaxDX = 9
	DistanceMaxDY = 4
)

// distanceTable 是 ds:CD0Dh 那 50 個 byte，逐格照抄。
//
// ⚠ **這是資料，不是公式。** 值大致是 10 × √(dx² + dy²)，但沒有任何一種取整
// （截斷／四捨五入／進位）能重現全部 50 格——最接近的四捨五入也差 5 格。
// 與精確值的偏差在 −1 … +2 之間，看起來是當年手填的。
// 用公式生會有十來格對不上，而且症狀是「命中率差一點」這種查不出來的偏差。
var distanceTable = [(DistanceMaxDY + 1) * (DistanceMaxDX + 1)]byte{
	2, 10, 20, 30, 40, 50, 60, 70, 80, 90,
	10, 14, 22, 31, 41, 51, 61, 71, 81, 92,
	20, 22, 28, 36, 44, 54, 63, 73, 82, 92,
	30, 31, 36, 42, 50, 58, 67, 76, 85, 95,
	40, 41, 45, 50, 57, 64, 72, 81, 89, 98,
}

// Distance 回傳原版那個「10 倍的歐氏距離」（sub_19D4D）。
//
// 超出視窗（|dx| > 9 或 |dy| > 4）時原版會讀到表後面的零區——
// 那不是「距離 0」，所以這裡回報 false，由呼叫端決定怎麼辦。
func Distance(dx, dy int) (int, bool) {
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx > DistanceMaxDX || dy > DistanceMaxDY {
		return 0, false
	}
	return int(distanceTable[dy*(DistanceMaxDX+1)+dx]), true
}

// ── 遭遇生成（`sub_16890` 的後半，docs/re/78）───────────────────────

// 記錄區標頭裡跟遭遇生成有關的三個欄位。
const (
	hdrSpawnDenom = 0x2F // 每步觸發的分母：擲 1..denom，等於 1 才生成
	hdrSpawnKinds = 0x31 // 敵人種類數
	hdrSpawnSlots = 0x32 // section 15 的槽數上限
)

// SpawnTables 是執行檔裡三張各 13 項的表（`ds:AA60h`／`AA6Dh`／`AA7Ah`）。
//
// 索引來自敵人資料表第 kind 筆的 `+0x05` 低 4 位。
type SpawnTables struct {
	Near [13]byte // → 記錄 +0x00，第一道距離門檻
	Far  [13]byte // → 記錄 +0x01，第二道距離門檻
	Dist [13]byte // 從隊伍沿一個方向走幾步放敵人
}

// spawnStep 是方向跳表 `ds:AAB1h` 的 9 個方向（`docs/re/78` §2）。
// 方向 4 是原地——原版擲到它就放棄這一次。
var spawnStep = [9][2]int{
	{0, -1}, // 0 上
	{0, +1}, // 1 下
	{-1, 0}, // 2 左
	{+1, 0}, // 3 右
	{0, 0},  // 4 原地（不生成）
	{-1, -1},
	{+1, -1},
	{-1, +1},
	{+1, +1},
}

// SpawnResult 是生成的結果。Placed ＝ false 時什麼都沒改。
type SpawnResult struct {
	Placed bool
	Slot   int  // 用了 section 15 的第幾槽
	X, Y   int  // 敵人格的座標
	Kind   byte // 敵人種類
	Count  int  // 這一組幾隻
}

// SpawnEncounter 跑一次遭遇生成（原版每走一步跑一次）。
//
// 流程照 `sub_16890`：擲 1／分母 → 找 section 15 的空槽 → 擲種類 →
// 查三張表填記錄的兩道距離門檻 → 擲方向與距離、沿路每一格都要是空地 →
// 把那一格改成 nibble 15。
//
// ⚠ **會改寫區塊本體**（section 15 的記錄與地圖第 1／2 層），原版也是這樣。
func (w *World) SpawnEncounter(t SpawnTables) SpawnResult {
	hdr := w.Block.Header
	if len(hdr) <= hdrSpawnSlots {
		return SpawnResult{}
	}
	denom := int(hdr[hdrSpawnDenom])
	if denom == 0 || w.RNG.Roll(denom) != 1 {
		return SpawnResult{}
	}
	slot, rec := w.freeSpawnSlot(int(hdr[hdrSpawnSlots]))
	if rec == nil {
		return SpawnResult{}
	}
	kinds := int(hdr[hdrSpawnKinds])
	if kinds == 0 {
		return SpawnResult{}
	}
	kind := byte(w.RNG.Roll(kinds))

	raw, err := w.Block.EnemyData()
	if err != nil || int(kind)*8+8 > len(raw) {
		return SpawnResult{}
	}
	data := raw[int(kind)*8 : int(kind)*8+8]
	idx := int(data[0x05] & 0x0F)
	if idx >= len(t.Dist) {
		return SpawnResult{}
	}

	dir := w.RNG.Roll(9) - 1
	if dir < 0 || dir >= len(spawnStep) || dir == 4 {
		return SpawnResult{} // 原地：這一次不生成
	}
	x, y := int(w.Party.X), int(w.Party.Y)
	for i := 0; i < int(t.Dist[idx]); i++ {
		x += spawnStep[dir][0]
		y += spawnStep[dir][1]
		if x <= 0 || y <= 0 || x >= w.Block.Dim || y >= w.Block.Dim {
			return SpawnResult{}
		}
		terrain, _, _, err := w.Block.At(x, y)
		if err != nil || terrain != 0 {
			return SpawnResult{} // 沿路必須是空地
		}
	}

	count := w.RNG.Roll(int(data[0x04]>>4) + 1)
	rec[0x00], rec[0x01] = t.Near[idx], t.Far[idx]
	rec[0x03], rec[0x04] = kind, byte(count)
	if err := w.Block.SetCell(x, y, nibbleEnemy, byte(slot)); err != nil {
		return SpawnResult{}
	}
	return SpawnResult{Placed: true, Slot: slot, X: x, Y: y, Kind: kind, Count: count}
}

// nibbleEnemy 是敵人格的第 1 層值（`sub_16890` 的 `ds:46B3h ← 0x0F`）。
const nibbleEnemy = 15

// freeSpawnSlot 找 section 15 的第一個空槽。
//
// 「空」的判準是三組的**型別**都是 0（`sub_16890` 的
// `記錄[3] | 記錄[5] | 記錄[7]`）——數量欄位不看。
func (w *World) freeSpawnSlot(limit int) (int, []byte) {
	for i := 0; i < limit; i++ {
		rec, err := w.Block.SectionRecord(15, i)
		if err != nil || len(rec) < 9 {
			continue
		}
		if rec[0x03]|rec[0x05]|rec[0x07] == 0 {
			return i, rec
		}
	}
	return 0, nil
}
