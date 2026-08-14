package play

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
)

// 戰鬥畫面（docs/spec/16、docs/re/40）。
//
// 原版沒有專屬的戰鬥畫面：它把地圖視窗換成隊伍名單，訊息與選單都印在原本的
// 訊息視窗裡。這一層就照這個做——**畫面模式**一個旗標，加上名單與訊息兩份輸出。

// ScreenMode 是原版的 ds:46B9h（docs/re/25 §2.5）。
type ScreenMode int

const (
	ModeMap    ScreenMode = 0
	ModeRoster ScreenMode = 1
)

// 名單一行的欄座標（docs/re/15 §4）。名字從欄 1 起。
const (
	colName    = 1
	colAC      = 0x11
	colAmmo    = 0x15
	colMaxCON  = 0x18
	colCON     = 0x1C
	colWeapon  = 0x20
	rosterCols = 38
)

// RosterHeader 是原版字串表 0xB270 的第 136 條（`exe:2:136`），
// 欄位的**順序**由它釘死。
//
// ⚠ 表頭的標籤位置與資料欄座標**不完全對齊**（例如 AC 的標籤在字串第 15 個字元，
// 值卻寫在欄 0x11）——那是原版就有的，不要為了「對齊好看」去動欄座標。
// 原字串尾端還有一個 `` 控制碼，這裡是顯示用的部分，不含控制碼。
const RosterHeader = "   NAME        AC AMM MAX CON WEAPON "

// RosterRow 是名單的一行。欄座標分開存，中文化重排時只改這裡。
type RosterRow struct {
	Name   string
	AC     string
	Ammo   string
	MaxCON string
	CON    string // CON ≤ 0 時是狀態字而不是數字
	Weapon string
}

// Text 把一行排成 38 欄的字串（欄座標照 docs/re/15 §4）。
func (r RosterRow) Text() string {
	line := []byte(spaces(rosterCols))
	put := func(col int, s string) {
		for i := 0; i < len(s) && col+i < rosterCols; i++ {
			line[col+i] = s[i]
		}
	}
	put(colName, r.Name)
	put(colAC, r.AC)
	put(colAmmo, r.Ammo)
	put(colMaxCON, r.MaxCON)
	put(colCON, r.CON)
	put(colWeapon, r.Weapon)
	return string(line)
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

// Roster 把隊伍排成名單。
//
// ⚠ **CON ≤ 0 印狀態字不印數字**（docs/re/15 §4）——負的 CON 直接顯示會讓
// 玩家以為那是血量，而原版在那個欄位放的是 UNC／SER／CRT／MRT／COM。
func Roster(p *game.Party) []RosterRow {
	rows := make([]RosterRow, 0, len(p.Members))
	for _, m := range p.Members {
		if m == nil {
			continue
		}
		con := fmt.Sprintf("%d", m.CON)
		if m.CON <= 0 {
			// CON 為負但還沒破 −11 時 WoundLevel 是 0；那一段與 CON ＝ 0
			// 一樣印 UNC（docs/re/20 §5 的五個帶）。
			w := m.WoundLevel()
			if w == 0 {
				w = len(game.WoundNames) - 1
			}
			con = game.WoundNames[w]
		}
		rows = append(rows, RosterRow{
			Name:   m.Name,
			AC:     fmt.Sprintf("%d", m.AC),
			Ammo:   "",
			MaxCON: fmt.Sprintf("%d", m.MaxCON),
			CON:    con,
			Weapon: "",
		})
	}
	return rows
}

// CommandOption 是指令選單的一個選項。
//
// ⚠ **Key 不跟著翻譯走。** 指令碼是熱鍵字母表的索引（docs/re/38 §2），
// 中文選項要把字母顯示出來，不能只留中文。
type CommandOption struct {
	Key   byte
	Cmd   game.Command
	Label string
}

// commandMenu 是原版第 55 條字串的顯示順序：Run Use Hire Evade Attack Weapon Load。
//
// ⚠ **顯示順序不是指令碼**——碼在 Cmd 欄位裡，兩者刻意不同。
var commandMenu = []CommandOption{
	{'R', game.CmdRun, "Run"},
	{'U', game.CmdUse, "Use"},
	{'H', game.CmdHire, "Hire"},
	{'E', game.CmdEvade, "Evade"},
	{'A', game.CmdAttack, "Attack"},
	{'W', game.CmdWeapon, "Weapon"},
	{'L', game.CmdLoad, "Load/unjam"},
}

// CommandMenu 回傳指令選單。label 可以換成翻譯過的文字，
// 熱鍵與指令碼原樣保留。
func CommandMenu(label func(game.Command) string) []CommandOption {
	out := make([]CommandOption, len(commandMenu))
	copy(out, commandMenu)
	if label == nil {
		return out
	}
	for i := range out {
		if s := label(out[i].Cmd); s != "" {
			out[i].Label = s
		}
	}
	return out
}

// MenuLines 把選單排成每行最多 width 欄的多行文字。
//
// 七個選項在 640 × 400 的中文畫布下一行放不下（docs/spec/16 §4），
// 所以這裡回傳多行而不是一整條。每個選項印成「<字母> <文字>」。
func MenuLines(opts []CommandOption, width int) []string {
	var lines []string
	cur := ""
	for _, o := range opts {
		item := fmt.Sprintf("%c %s", o.Key, o.Label)
		switch {
		case cur == "":
			cur = item
		case len(cur)+2+len(item) <= width:
			cur += "  " + item
		default:
			lines = append(lines, cur)
			cur = item
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// CombatScene 是一場戰鬥的畫面狀態。
//
// 它只保存「畫面上有什麼」，規則全部在 internal/game——
// 這一層一條規則都不做決定。
type CombatScene struct {
	Mode   ScreenMode
	Battle *game.Battle
	Phase  *game.CommandPhase

	// Turn 是目前輪到誰下令（隊伍成員索引）；-1 表示指令階段結束。
	Turn int
	Log  []string
}

// NewCombatScene 開一場戰鬥的畫面：模式切成名單，開一個新的指令階段。
func NewCombatScene(b *game.Battle) *CombatScene {
	s := &CombatScene{Mode: ModeRoster, Battle: b}
	s.BeginCommands()
	return s
}

// BeginCommands 開始一回合的指令階段，並把游標移到第一個能下令的人。
func (s *CombatScene) BeginCommands() {
	s.Phase = game.NewCommandPhase(len(s.Battle.Party.Members))
	s.Turn = -1
	s.advance(0)
}

// advance 從 from 起找下一個能下令的成員（CON ≤ 0 的跳過，docs/re/38 §4）。
func (s *CombatScene) advance(from int) {
	for i := from; i < len(s.Battle.Party.Members); i++ {
		if game.CanCommand(s.Battle.Party.Members[i]) {
			s.Turn = i
			return
		}
	}
	s.Turn = -1
}

// Prompt 是目前這個人的提示行（名字 ＋ 選單）。指令階段結束時回 nil。
func (s *CombatScene) Prompt(label func(game.Command) string, width int) []string {
	if s.Turn < 0 {
		return nil
	}
	m := s.Battle.Party.Members[s.Turn]
	out := []string{m.Name + ", choose:"}
	return append(out, MenuLines(CommandMenu(label), width)...)
}

// Choose 收一個按鍵。回傳這個按鍵有沒有被接受。
//
// 沒有裝備武器的人選攻擊會被要求重選（`0x120D3` 印字串 56 之後 stc），
// 所以 armed 由呼叫端提供——裝備欄還沒解到能判斷。
func (s *CombatScene) Choose(key byte, armed bool) bool {
	if s.Turn < 0 {
		return false
	}
	cmd, ok := game.CommandFromKey(key)
	if !ok || cmd == game.CmdNone {
		return false
	}
	if cmd == game.CmdAttack && !armed {
		s.Log = append(s.Log, "Your weapon is jammed.")
		return false // 重問這個人
	}
	s.Phase.Set(s.Turn, cmd, 0)
	s.advance(s.Turn + 1)
	return true
}

// PartyFlees 整隊逃跑：指令階段立刻結束，畫面切回地圖（docs/re/38 §3）。
func (s *CombatScene) PartyFlees(dir game.FleeDirection) {
	s.Phase.PartyFlees(dir)
	s.Turn = -1
	s.Mode = ModeMap
	s.Log = append(s.Log, "The party runs.")
}

// Done 回報指令階段是不是問完了。
func (s *CombatScene) Done() bool { return s.Turn < 0 }
