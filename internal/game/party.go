package game

// 隊伍與角色的最小模型（docs/spec/04 §4.1、docs/re/15、35）。
//
// 這一輪只放「每 16 刻的體力處理」需要的欄位。完整的角色記錄（技能、物品、
// 經驗值）等規格 05 收攏之後再擴。欄位名沿用記錄位移的語意，不自創術語。

// 狀態位元（角色記錄 +0x28，docs/re/35 §1）。
const (
	StatusRadiation  = 1 << iota // Radiation poisoning
	StatusHerpes                 // Wasteland Herpes
	StatusBugByte                // Bug byte
	StatusSewerRot               // Sewer rot
	StatusDesertDust             // Desert dust
	StatusRabies                 // Rabies
	StatusD6                     // D6（原版就叫這個名字，看起來是佔位）
	StatusD7                     // D7
)

// StatusWorsening 是「會隨時間惡化」的那一組（sub_12440 的 and al, 0F0h）。
const StatusWorsening = StatusDesertDust | StatusRabies | StatusD6 | StatusD7

// 傷勢門檻與死亡線（docs/re/15 §3、docs/re/35 §2）。
const (
	woundedFloor = -10 // CON 低於這個值就不再自然回復，改成繼續惡化
	deathFloor   = -50 // 掉破這條直接歸零
)

// Character 是角色記錄裡規則層目前用得到的部分（docs/spec/05 §4）。
type Character struct {
	Name    string
	CON     int16 // +0x1D/+0x1E，可為負
	MaxCON  int16 // +0x1B/+0x1C
	PreHurt int16 // +0x26/+0x27，受傷前的備份
	Status  uint8 // +0x28，八個狀態位元

	Attributes [AttrCount]byte // +0x0E–+0x14
	Money      uint32          // +0x15–+0x17，24-bit
	Gender     byte            // +0x18
	Nation     byte            // +0x19
	AC         byte            // +0x1A，＝ 裝備護甲的 Dice（docs/re/45 §3.4）
	EquipIndex byte            // +0x1F，裝備武器的物品槽
	ArmorIndex byte            // +0x25，裝備護甲的物品槽
	SkillPts   byte            // +0x20
	XP         uint32          // +0x21–+0x23，24-bit
	Level      byte            // +0x24
	Rank       string          // +0x32…
	// Mission 是記錄 `+0x4B` 的 bit0：**參與過摧毀 Base Cochise**。
	// 結局那一段逐人設起來（`0x1B4FB`），Radio 的第一輪讀它。
	Mission bool
	// Praised 是記錄 `+0x4C` 的 bit0：**總部已經表揚過**。
	// Radio 第一輪設起來，所以那段賀詞一個人只聽得到一次（`docs/re/96` §5）。
	Praised bool
	Skills     []Slot          // +0x80，30 格
	Items      []Slot          // +0xBD，30 格

	// Source 是這個角色**整筆 256 bytes 的來源記錄**，只有雇用來的 NPC 有
	// （`docs/re/110` §2：原版是逐 byte 抄 256 bytes 進隊伍的記錄槽）。
	//
	// ⚠ 存檔時要**先把它原樣寫下去、再讓 `StoreTo` 蓋已解欄位**。
	// 少了這一步，新隊員在存檔裡就只有我們解過的那些欄位，
	// 未解區域會是上一個佔用那一格的人留下的殘骸——而**畫面上完全看不出來**。
	Source []byte
}

// Dead 照原版的判定：CON 兩個 byte 都是 0（`sub_172AE`）。**不看正負**。
// 它的 10 個呼叫端都在體力處理與角色管理那條路上（`docs/re/35` §2）。
func (c *Character) Dead() bool { return c.CON == 0 }

// Down 是「倒下」：CON ≤ 0（`sub_172BB` ＝ 高位為負 → CF；兩 byte 為 0 → CF）。
//
// ⚠ **戰鬥的每一個判斷都用這個，不是 `Dead`**：CON 可以是負的
// （負到 −40 以下有五級傷勢，`docs/re/19`），而重傷倒下的人不能行動、
// 不會被敵人挑中，也不算在「還有沒有人能打」裡（`sub_19D0E`）。
// 用 `Dead` 會讓 CON 負值的人被當成好手好腳——症狀是**戰鬥永遠打不完**：
// 全隊倒下但誰都下不了令，指令階段空轉（`docs/re/89` §3）。
func (c *Character) Down() bool { return c.CON <= 0 }

// Party 是目前這一組隊伍。
type Party struct {
	Members  []*Character
	Selected int // 目前選中的角色索引；他的自然回復慢一半
	X, Y     uint8

	// PlayerStepped 對應原版的 ds:916Bh（docs/re/32 §7.1）：
	// 最後一次移動是不是玩家自己走的。檢定成功時只有這個是 true 才給經驗值——
	// 站著讓時間流逝所觸發的檢定不給。**它記的是最後一次移動，不是每次檢定**。
	PlayerStepped bool
}

// Tick16 是每 16 刻跑一次的體力處理（sub_12440，docs/re/35 §2）。
//
// 分支順序照原版：先看死亡、再看重症狀態、再看 CON 的正負。
func (p *Party) Tick16(tick uint8) {
	for i, c := range p.Members {
		if c == nil || c.Dead() {
			continue
		}
		// 目前選中的角色用 0x7F 當遮罩，也就是兩倍的間隔。
		gate := uint8(0x3F)
		if i == p.Selected {
			gate = 0x7F
		}

		switch {
		case c.Status&StatusWorsening != 0:
			// 惡化只在每 64 刻動一次，而且不用選中角色的慢速遮罩。
			if tick&0x3F != 0 {
				continue
			}
			if c.CON < 0 {
				// 已經是負的：走含死亡線的那一條（sub_1251E）。
				c.CON--
				if c.CON < deathFloor {
					c.CON = 0
				}
			} else {
				c.CON--
				if c.CON == 0 {
					c.CON = -1 // 歸零代表死亡，惡化要跨過去繼續往下掉
				}
			}

		case c.CON < 0:
			if c.CON >= woundedFloor {
				c.CON++
				// ⚠ 只有「加完剛好回到 0」時才做備份還原——原版在 sub_12537
				// 之後立刻測 CON 是否為零，非零就跳去下一個人（0x12477）。
				// 零代表死亡，所以要用受傷前的備份右移一位把人救回來。
				if c.CON == 0 {
					c.CON = c.PreHurt >> 1
					if c.CON == 0 {
						c.CON = 1
					}
				}
			} else {
				c.CON--
				if c.CON < deathFloor {
					c.CON = 0
				}
			}

		case c.CON < c.MaxCON:
			if tick&gate == 0 {
				c.CON++
			}
		}
	}
}

// WipeState 是「這一組倒了沒」的三種結果（`docs/spec/28` §2）。
type WipeState int

const (
	// WipeNone：還有人站著，什麼都不做。
	WipeNone WipeState = iota
	// WipeSwitch：全倒但至少一個人救得回來 —— 原版自動切到下一支隊伍。
	WipeSwitch
	// WipeDead：每個人都倒下而且都有傷勢或狀態 —— 死亡畫面。
	WipeDead
)

// Wipe 判斷這一組的處置（`0x16C2B`，`docs/re/99` §3）。
//
// ⚠ **「救不救得回來」不是「倒了幾個」。** CON ≤ 0 但傷勢等級 0、
// 狀態位元也是 0 的人是**會自己醒的昏迷**（`docs/re/35`：不省人事會自己醒，
// 重傷開始才會一路往下掉）。只要有這種人，這一組就還有救，原版走換隊。
//
// 原版的寫法是「找到第一個**不是**『倒下且有傷勢或狀態』的人就跳出迴圈」，
// 跳出去之後再問一次「還有沒有人站著」。這裡照它的語意寫成三個分支，
// 行為一樣而讀得懂。
func (p *Party) Wipe() WipeState {
	if p == nil || len(p.Members) == 0 {
		return WipeDead // 0 人也走死亡畫面（原版的第一個分支）
	}
	standing, salvageable := false, false
	for _, c := range p.Members {
		if c == nil {
			continue
		}
		if !c.Down() {
			standing = true
			continue
		}
		if c.WoundLevel() == 0 && c.Status == 0 {
			salvageable = true
		}
	}
	switch {
	case standing:
		return WipeNone
	case salvageable:
		return WipeSwitch
	}
	return WipeDead
}
