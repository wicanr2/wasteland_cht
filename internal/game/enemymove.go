package game

// 敵人在地圖上移動（`docs/re/87`、`docs/re/101` §4–§6、`docs/re/116`）。
//
// 原版每個戰鬥回合算一次計畫（`sub_14BF0`）再執行（`sub_15036`）：
// 一筆遭遇要嘛換位置、要嘛逃跑、要嘛朝隊伍衝，要嘛留在原地。
// 這個檔只做**計畫**——執行要動地圖與遭遇佇列，那是呼叫端的事。

// 步向索引與位移（`ds:AAB1h` 的九個近指標，`docs/re/116` §1）。
//
// ⚠ **0–3 與隊伍走路的方向編號是同一套**（上／下／左／右，`docs/re/26` §2），
// 斜向排在 5–8，中央 4 是不動。順序不是「照方向繞一圈」——
// 憑直覺重排會讓敵人往反方向走，而畫面上看起來只是「它亂走」。
var stepDeltas = [9][2]int{
	{0, -1},  // 0 上
	{0, 1},   // 1 下
	{-1, 0},  // 2 左
	{1, 0},   // 3 右
	{0, 0},   // 4 不動
	{-1, -1}, // 5 左上
	{1, -1},  // 6 右上
	{-1, 1},  // 7 左下
	{1, 1},   // 8 右下
}

// StepStay 是「留在原地」那一格（`ds:AAB1h` 的第 4 格指向 `0x18016`）。
const StepStay = 4

// 計畫 byte 的高 2 位 ＝ 三句訊息的哪一句（`docs/re/101` §4）。
const (
	PlanReposition = 0x00 // " moves to a better position."（字串 72）
	PlanFlee       = 0x40 // " runs away."（字串 74）
	PlanCharge     = 0x80 // " runs at you."（字串 75）
)

// PlanStep 取計畫的步向（低 6 位）。
func PlanStep(plan int) int { return plan & 0x3F }

// PlanMessage 取計畫的訊息索引（高 2 位，0／1／2）。
func PlanMessage(plan int) int { return (plan >> 6) & 3 }

// StepDelta 回步向的位移。步向不在 0–8 就回 false。
func StepDelta(step int) (dx, dy int, ok bool) {
	if step < 0 || step >= len(stepDeltas) {
		return 0, 0, false
	}
	return stepDeltas[step][0], stepDeltas[step][1], true
}

// signStep 是 `ds:AA05h` 那張 3 × 3：索引 ＝ (sy+1)×3 + (sx+1)。
var signStep = [9]int{
	5, 0, 6, // sy ＝ −1：左上 上 右上
	2, 4, 3, // sy ＝  0：左 不動 右
	7, 1, 8, // sy ＝ +1：左下 下 右下
}

// StepToward 把兩軸的正負號換成步向（`0x14F03`–`0x14F1E`）。
func StepToward(sx, sy int) int { return signStep[(sign(sy)+1)*3+sign(sx)+1] }

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// 距離走 `game.Distance`（`ds:CD0Dh` 那張 10 × 歐氏距離的表，
// 已經在 `encounter.go` 照抄過一份）。**不要再抄第二份**——
// 同一張表兩個副本遲早會漂，而症狀是「命中率差一點」這種查不出來的偏差。

// RangeBand 是武器類別的射程帶（`sub_156DE`，`docs/re/116` §4.2）。
func RangeBand(c ItemClass) int {
	switch c {
	case 2, 5, 10, 13:
		return 0
	case 3, 6, 8, 11:
		return 1
	}
	return 2
}

// 每一帶的起始值與增量（`ds:A5B8h` ＝ `fb 08 13`、`ds:A5BBh` ＝ `05 02 01`）。
var (
	bandStart = [3]int{-5, 8, 19}
	bandRise  = [3]int{5, 2, 1}
)

// PositionScore 是「站在這個距離上有多不划算」（`sub_139CE`）。
//
//	分數 ＝ 起始 ＋ 增量 × (k ＋ 1)，k ＝ ⌈距離 ／ 3⌉
//
// 三條都隨距離遞增，所以**挑分數最小的那一格 ＝ 靠近隊伍**。
// 帶的差別不在挑哪一格，在**要不要挑**：同一個分數在 `PlanEnemyMove` 裡
// 當門檻用——分數高過「行動值 ＋ 種類常數」才動，所以遠程武器（帶 0，
// 分數隨距離長得快）在遠處反而先動，近戰（帶 2，幾乎是常數 20）要行動值
// 夠低才動。
func PositionScore(dist, band int) int {
	if band < 0 || band >= len(bandStart) {
		band = 2
	}
	k := 0
	for k*3 < dist {
		k++
	}
	return bandStart[band] + bandRise[band]*(k+1)
}

// MoveEnv 是算一次移動計畫要看的東西。
//
// **座標是地圖格**，`PartyX`／`PartyY` 是隊伍那一格。
type MoveEnv struct {
	X, Y           int // 這一筆遭遇現在在哪一格
	PartyX, PartyY int
	Data           EnemyData
	// Distance 是敵方記錄標頭 `+0x03`（與隊伍的距離），與 `game.Distance` 同單位。
	Distance int
	// GroupHP 是這一組剩下的血量總和（`sub_14FDE`）。
	GroupHP int
	// NoMove 是遭遇記錄 `+0x09` 的 bit2：這一筆遭遇不參與移動。
	NoMove bool
	// Passable 回答「這一格站得上去嗎」（地形 ＋ 沒有別的隊伍或遭遇）。
	Passable func(x, y int) bool
	// InView 回答「這一格還在畫面視窗裡嗎」（`sub_169CF`）。
	// **只有「換位置」那一支查它**——追出畫面的敵人在原版是走得出去的。
	InView func(x, y int) bool
}

// PlanEnemyMove 算這一回合的移動計畫（`sub_14BF0`，`docs/re/101` §6）。
//
// 回 `NoMovePlan` 表示這一回合不動（原版的 `0xFF`）。
func PlanEnemyMove(e MoveEnv) int {
	if e.NoMove {
		return NoMovePlan
	}
	// 第一個分岔：手上有射程武器的走「換位置」，只有近戰的往下走
	// （`ds:CD00h` ＝ 類別 2–13，少的正好是徒手與近戰）。
	if e.Data.Weapon.Ranged() {
		return planReposition(e)
	}
	// 種類 1（Animal）不做士氣檢查——動物不會逃。
	if e.Data.Kind != KindAnimal && !morale(e, 1) {
		return planFlee(e)
	}
	// 距離 < 16 → 已經夠近，不動。
	if e.Distance < 0x10 {
		return NoMovePlan
	}
	return planCharge(e)
}

// morale 是「整組剩下的血量 ≥ 一隻的基礎血量 >> shift」（`sub_14F7A`／§6.2）。
func morale(e MoveEnv, shift int) bool {
	return e.GroupHP >= int(e.Data.Base)>>uint(shift)
}

// planReposition 是「換位置」那一支（`loc_14CCA`–`0x14DFC`）。
func planReposition(e MoveEnv) int {
	// 種類決定要不要先做士氣檢查，以及那個加在行動值上的常數
	// （Robot 0x0F 不檢查／Cyborg 0x0B 檢查一半／其餘 7 檢查四分之一）。
	konst := 7
	switch e.Data.Kind {
	case KindRobot:
		konst = 0x0F
	case KindCyborg:
		if !morale(e, 1) {
			return planFlee(e)
		}
		konst = 0x0B
	default:
		if !morale(e, 2) {
			return planFlee(e)
		}
	}
	band := RangeBand(e.Data.Weapon)
	cur := PositionScore(e.Distance, band)
	// 那道閘（`0x14D2C`–`0x14D2F`）：**現在這一格有多不划算**要超過
	// 行動值（敵人資料 `+0x02` 的原值，不是 × 8 之後的行動值）加上種類常數，
	// 才值得動。忍得下去就留在原地。
	//
	// ⚠ 方向容易記反：原版是 `sub_19C69` ＋ `jnb`，
	// 也就是「行動值 ＋ 常數 **小於** 分數 → 換位置」。
	// 反過來寫的話遠處的遠程敵人會一直逼近，而畫面上看起來只是「它很積極」。
	if int(e.Data.Speed)+konst >= cur {
		return NoMovePlan
	}

	best, bestScore := -1, 1<<20
	for step := 0; step < len(stepDeltas); step++ {
		dx, dy, _ := StepDelta(step)
		nx, ny := e.X+dx, e.Y+dy
		// 座標 0 是邊界不是格子（`0x14D69`、`0x14D6F`）。
		if nx == 0 || ny == 0 {
			continue
		}
		// 原地那一格不查地形（它本來就站在那裡）。
		if step != StepStay && e.Passable != nil && !e.Passable(nx, ny) {
			continue
		}
		if e.InView != nil && !e.InView(nx, ny) {
			continue
		}
		d, ok := Distance(nx-e.PartyX, ny-e.PartyY)
		if !ok {
			continue
		}
		// ⚠ **相等也跳過**（`0x14DA8` 的 `jz`），所以先掃到的候選贏：
		// 正交（0–3）永遠比斜向（5–8）先掃到。
		if sc := PositionScore(d, band); sc < bestScore {
			best, bestScore = step, sc
		}
	}
	if best < 0 || best == StepStay || bestScore == cur {
		return NoMovePlan
	}
	return PlanReposition | best
}

// planFlee 是「逃跑」那一支（`0x14E03`）：往隊伍的反方向走一格。
func planFlee(e MoveEnv) int {
	step, ok := stepAway(e, true)
	if !ok {
		return NoMovePlan
	}
	return PlanFlee | step
}

// planCharge 是「朝隊伍衝」那一支（`0x14E0C`）。
//
// ⚠ 多一道檢查：走過去**真的要更近**（`0x14E2E` 比記錄 `+0x03`），
// 否則不動。少了它的話擋在障礙後面的敵人會左右橫跳。
func planCharge(e MoveEnv) int {
	step, ok := stepAway(e, false)
	if !ok {
		return NoMovePlan
	}
	dx, dy, _ := StepDelta(step)
	d, ok := Distance(e.X+dx-e.PartyX, e.Y+dy-e.PartyY)
	if !ok || d >= e.Distance {
		return NoMovePlan
	}
	return PlanCharge | step
}

// stepAway 算朝向（away ＝ false）或背向（away ＝ true）隊伍的那一步。
//
// 先試斜的，不行退成單軸，三個都不行就不動（`sub_14ECB`）。
func stepAway(e MoveEnv, away bool) (int, bool) {
	sx, sy := sign(e.PartyX-e.X), sign(e.PartyY-e.Y)
	if away {
		sx, sy = -sx, -sy
	}
	for _, try := range [][2]int{{sx, sy}, {sx, 0}, {0, sy}} {
		if try[0] == 0 && try[1] == 0 {
			continue
		}
		nx, ny := e.X+try[0], e.Y+try[1]
		if nx == 0 || ny == 0 {
			continue
		}
		if e.Passable != nil && !e.Passable(nx, ny) {
			continue
		}
		return StepToward(try[0], try[1]), true
	}
	return 0, false
}
