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
