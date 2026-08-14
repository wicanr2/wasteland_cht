package game

// 戰鬥的指令階段（docs/spec/14、docs/re/38）。
//
// 一回合的開頭是「每個隊員這回合要做什麼」；規格 12 的行動順序排在這之後。

// Command 是戰鬥指令碼。
//
// ⚠ **碼是熱鍵字母表 ds:A44Dh（" HAWRELU"）的索引，不是選單的顯示順序。**
// 選單印的是 Run / Use / Hire / Evade / Attack / Weapon / Load-unjam——
// 照選單順序編號會把每一條規則都對錯人（docs/re/38 §2）。
type Command byte

const (
	CmdNone   Command = 0 // ' '
	CmdHire   Command = 1 // 'H'
	CmdAttack Command = 2 // 'A'
	CmdWeapon Command = 3 // 'W'
	CmdRun    Command = 4 // 'R'
	CmdEvade  Command = 5 // 'E'
	CmdLoad   Command = 6 // 'L'
	CmdUse    Command = 7 // 'U'

	// CmdPartyRan 是整隊逃跑之後每個人被寫上的碼（loc_11F93）。
	// 選單選不到它。
	CmdPartyRan Command = 8
)

// CommandKeys 是 ds:A44Dh 那張熱鍵字母表，索引就是指令碼。
var CommandKeys = [8]byte{' ', 'H', 'A', 'W', 'R', 'E', 'L', 'U'}

// CommandFromKey 把按鍵（原版已經轉成大寫）轉成指令碼。
// 掃不到回 false——原版是回 0xFF，呼叫端當成「重問」。
func CommandFromKey(key byte) (Command, bool) {
	for i, k := range CommandKeys {
		if k == key {
			return Command(i), true
		}
	}
	return CmdNone, false
}

// 命中門檻的基礎值（docs/re/20 §1.1）：敵方要打中這個人得擲贏多少。
// 值越大越難被打中。
const (
	baseEvading   = 60
	baseAttacking = 50
	baseDefault   = 40
)

// DefenceBase 回傳這個指令碼對應的命中門檻基礎值。
//
// **迴避的處理程式是空的**——效果就只在這一行（docs/re/38 §2）。
func (c Command) DefenceBase() int {
	switch c {
	case CmdEvade:
		return baseEvading
	case CmdAttack:
		return baseAttacking
	default:
		return baseDefault
	}
}

// FleeDirection 是逃跑選的方向。原版的方向字母表有兩組按鍵
// （`I K J L ' '` 與四個方向鍵的掃描碼）對到同五個值。
type FleeDirection byte

const (
	FleeUp FleeDirection = iota
	FleeDown
	FleeLeft
	FleeRight
	FleeStay

	fleeDirections = 5
)

// FleeKeys 是 ds:A45Ch 那張方向字母表：**兩組按鍵對到同五個方向**。
// 前五個是字母，後四個是方向鍵的掃描碼（原版收到索引 > 4 就減 5）。
var FleeKeys = [9]byte{'I', 'K', 'J', 'L', ' ', 0xC8, 0xD0, 0xCB, 0xCD}

// FleeDirectionFromKey 把按鍵轉成方向。掃不到回 false。
func FleeDirectionFromKey(key byte) (FleeDirection, bool) {
	for i, k := range FleeKeys {
		if k != key {
			continue
		}
		if i >= fleeDirections {
			i -= fleeDirections
		}
		return FleeDirection(i), true
	}
	return FleeStay, false
}

// CommandPhase 是一回合的指令階段狀態。
//
// 兩個陣列對應原版的 ds:46D8h（指令碼）與 ds:46DAh（參數），
// 索引是隊伍成員編號。
type CommandPhase struct {
	Cmd   []Command
	Arg   []byte
	Ran   bool          // 整隊逃跑了，這一回合到此為止
	RanTo FleeDirection // Ran 為真時有效
}

// NewCommandPhase 開一個新的指令階段：所有人的指令碼歸零（0x11F80）。
func NewCommandPhase(members int) *CommandPhase {
	return &CommandPhase{Cmd: make([]Command, members), Arg: make([]byte, members)}
}

// CanCommand 回報這個成員能不能下令（sub_172BB）：CON ≤ 0 就不能。
func CanCommand(c *Character) bool {
	return c != nil && c.CON > 0
}

// Set 記下一個成員的指令與參數。
func (p *CommandPhase) Set(member int, cmd Command, arg byte) bool {
	if member < 0 || member >= len(p.Cmd) {
		return false
	}
	p.Cmd[member], p.Arg[member] = cmd, arg
	return true
}

// PartyFlees 是整隊逃跑（loc_11F93）：每個人的指令碼寫 8、參數寫同一個方向，
// 指令階段就此結束。
//
// ⚠ **逃跑沒有擲骰，也沒有失敗分支**——不要在這裡加成功率（docs/re/38 §3）。
func (p *CommandPhase) PartyFlees(dir FleeDirection) {
	for i := range p.Cmd {
		p.Cmd[i], p.Arg[i] = CmdPartyRan, byte(dir)
	}
	p.Ran, p.RanTo = true, dir
}

// Defence 回傳這個成員這回合的命中門檻基礎值。
// 沒下令（或編號越界）就是預設值。
func (p *CommandPhase) Defence(member int) int {
	if member < 0 || member >= len(p.Cmd) {
		return baseDefault
	}
	return p.Cmd[member].DefenceBase()
}
