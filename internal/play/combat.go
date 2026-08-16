package play

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
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

// Roster 把隊伍排成名單（`sub_1708B`，docs/re/103）。
//
// ⚠ **CON ≤ 0 印狀態字不印數字**（docs/re/15 §4）——負的 CON 直接顯示會讓
// 玩家以為那是血量，而原版在那個欄位放的是 UNC／SER／CRT／MRT／COM。
//
// items 與 name 是 `AMM` 與 `WEAPON` 兩欄要用的（裝備武器的類別與名字）。
// 兩個都可以是 nil／空——那時候那兩欄留白，與接上去之前一樣。
func Roster(p *game.Party, items game.ItemTable, name func(byte) string) []RosterRow {
	rows := make([]RosterRow, 0, len(p.Members))
	for _, m := range p.Members {
		if m == nil {
			continue
		}
		con := fmt.Sprintf("%d", m.CON)
		if m.CON <= 0 {
			// 傷勢等級直接當索引（`docs/re/17` §4.4）：0 是 `UNC`、
			// 5（CON ＝ 0）是骷髏字模。**這裡不做二次對照**——
			// 原版的表本來就是六格對六個等級。
			con = game.WoundNames[m.WoundLevel()]
		}
		rows = append(rows, RosterRow{
			Name:   m.Name,
			AC:     fmt.Sprintf("%d", m.AC),
			Ammo:   ammoColumn(m, items),
			MaxCON: fmt.Sprintf("%d", m.MaxCON),
			CON:    con,
			Weapon: weaponColumn(m, name),
		})
	}
	return rows
}

// equippedSlot 取出裝備武器那一格（`sub_196C9` ＋ `sub_19AC8`）。
//
// ⚠ **`+0x1F` 是 1 起算的**：原版算的是 `0xBB ＋ 2n`，而物品陣列從 `+0xBD` 起，
// 所以 n ＝ 1 指的是第 0 格。0 ＝ 沒有裝備。
func equippedSlot(c *game.Character) (game.Slot, bool) {
	n := int(c.EquipIndex)
	if n <= 0 || n > len(c.Items) {
		return game.Slot{}, false
	}
	return c.Items[n-1], true
}

// ammoColumn 是 `AMM` 那一欄（`0x170C4`–`0x170F3`）。
//
// 原版的三道閘，少一道就會在不該有數字的地方印數字：
//
//	沒有裝備（`+0x1F` ＝ 0）                → 0
//	那一格的附屬 byte bit7 設起來            → 0
//	武器類別不在 `ds:CD00h` 那張表裡（近戰）  → 0
//
// 過了三道才印**低 6 位**——高 2 位不是次數（`docs/re/45`）。
func ammoColumn(c *game.Character, items game.ItemTable) string {
	slot, ok := equippedSlot(c)
	if !ok || slot.Value&0x80 != 0 {
		return "0"
	}
	if int(slot.ID) >= len(items) || !items[slot.ID].Class.Ranged() {
		return "0"
	}
	return fmt.Sprintf("%d", slot.Value&0x3F)
}

// weaponColumn 是 `WEAPON` 那一欄（`0x17165`–`0x17185`）：裝備武器的名字。
// 沒有裝備就留白。
func weaponColumn(c *game.Character, name func(byte) string) string {
	slot, ok := equippedSlot(c)
	if !ok || name == nil {
		return ""
	}
	return name(slot.ID)
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

// MenuLines 把選單排成**一行一個選項**，照原版的結構
// （每個選項一段 `\x10<文字>\x0D`，docs/re/40 §4）。
//
// 每一行的第一個字元就是熱鍵——`\x10` 捕捉的是下一個字元，
// 所以中文選項寫成「R 逃跑」時，捕捉到的仍然是 R。
func MenuLines(opts []CommandOption) []string {
	lines := make([]string, 0, len(opts))
	for _, o := range opts {
		lines = append(lines, fmt.Sprintf("%c %s", o.Key, o.Label))
	}
	return lines
}

// MenuFits 回報這份選單放不放得進 rows 行 × width 格。
//
// ⚠ **原版自己也放不下**：標題 ＋ 空行 ＋ 七個選項要 9 行，訊息視窗只有 6 行，
// 而原版怎麼容納還沒逆向（docs/re/40 §5）。所以這裡只回報，不自己決定
// 捲動或分頁——那是呈現層的事，而且要等 RE 補上才知道原版怎麼做。
func MenuFits(lines []string, width, rows int) bool {
	if len(lines) > rows {
		return false
	}
	for _, l := range lines {
		if len([]rune(l)) > width {
			return false
		}
	}
	return true
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

	// Items 是物品資料表（存檔區那一份，docs/re/45 §2）。
	// 沒接上時是 nil，武器一律當成零值——傷害 0 而不是崩掉。
	Items game.ItemTable
	// Names 是六個敵人種類的名稱，訊息要用（`docs/re/85`）。
	Names EnemyNames
	// CJKNames 是同六個種類的中文名（Big5），與 Names 同一組編號。
	CJKNames [6][]byte

	// CJK 查執行檔字串表 1 第 n 條的譯文（Big5，控制碼已解）。
	// 由 `Scene` 開場時接上；nil ＝ 沒有中文，畫面走英文那一份。
	CJK func(n int, opt textlayout.Options) []byte
	// UI 查重製版自己的介面文字（`translations/*/ui.tsv`）。
	// 原版沒有對應字串、或組法是重製版自己決定的句子走這一支。
	UI func(name string) []byte

	// LastCJK 是 `Log` 最後一句的中文（Big5）。指令階段那幾句
	//（卡彈、逃跑）不走 `ResolveRound`，所以另外留一格。
	LastCJK []byte
}

// firstName 是隊伍第一個人的名字（`\x0B` 的代入來源）。
func (s *CombatScene) firstName() string {
	for _, m := range s.Battle.Party.Members {
		if m != nil {
			return m.Name
		}
	}
	return ""
}

// 戰鬥訊息用到的原版字串編號（字串表 1，`docs/re/40` §3）。
//
// ⚠ 這些是**片段**不是整句：`\x0B` 是名字、`\x0F` 是數量、`\x0A`／`\x0C`
// 是單複數與性別分段。整句是原版字串的（35／31／48／39＋40／30／56／43）
// 就照用；由重製版自己拼出來的（命中那一句，`docs/re/86` §2）走 `ui:`。
const (
	strEncounterBegins = 30 // `Encounter begins...`
	strEnemyMisses     = 31 // ` miss\x0Aes\x0A\x0A.`
	strPartyMisses     = 35 // `\x0B misses.`
	strGainsXP         = 39 // `\x0B gains `
	strExperience      = 40 // ` experience.`
	strRuns            = 43 // `\x0B runs.`
	strDied            = 48 // ` died!`
	strChoose          = 55 // `, choose:` ＋ 七個 `\x10<文字>`
	strJammed          = 56 // `Your weapon is jammed.`
)

// zhStr 取字串表 1 第 n 條的譯文。沒接上或查不到回 nil。
func (s *CombatScene) zhStr(n int, opt textlayout.Options) []byte {
	if s.CJK == nil {
		return nil
	}
	return s.CJK(n, opt)
}

// zhJoin 把幾段接起來；**任何一段是 nil 就整句放棄**——
// 半句中文半句英文比整句英文更難讀，也讓「哪裡沒翻」看不出來。
func zhJoin(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		if p == nil {
			return nil
		}
		out = append(out, p...)
	}
	return out
}

// zhEnemy 是敵人的中文名（Big5）。查不到回 nil，整句就跟著放棄。
func (s *CombatScene) zhEnemy(e *game.Enemy) []byte {
	k := int(e.Data.Kind)
	if k < 0 || k >= len(s.CJKNames) || len(s.CJKNames[k]) == 0 {
		return nil
	}
	return s.CJKNames[k]
}

// NewCombatScene 開一場戰鬥的畫面：模式切成名單，開一個新的指令階段。
// EnemyNames 是六個敵人種類的單數名稱（執行檔字串表 1 的 `0x52 + Kind`，
// `docs/re/85`）。原文是 `Animal\n\ns\n` 這種單複數格式（控制碼 `0x0A`），
// 這裡只取單數那一段。
type EnemyNames [6]string

// Name 回某個種類的名稱；查不到回空字串（呼叫端自己決定要不要留空白）。
func (n EnemyNames) Name(k game.EnemyKind) string {
	if int(k) < len(n) {
		return n[k]
	}
	return ""
}

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
func (s *CombatScene) Prompt(label func(game.Command) string) []string {
	if s.Turn < 0 {
		return nil
	}
	m := s.Battle.Party.Members[s.Turn]
	out := []string{m.Name + ", choose:", ""}
	return append(out, MenuLines(CommandMenu(label))...)
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
		s.LastCJK = s.zhStr(strJammed, textlayout.Options{})
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
	// 字串 43（`\x0B runs.`）的 `\x0B` 是逃跑的人；整隊逃就用第一個人的名字。
	s.LastCJK = s.zhStr(strRuns, textlayout.Options{Name: s.firstName})
}

// Done 回報指令階段是不是問完了。
func (s *CombatScene) Done() bool { return s.Turn < 0 }

// rosterHeaderCJK 組戰鬥名單的中文表頭。
//
// **原版的表頭是寫死的 ASCII**（`RosterHeader`），不在字串表裡，所以翻譯走
// `ui:combat.hdr*` 這條路，與指令列同一個做法。
//
// ⚠ 與原版的一個差別：**中文標籤對齊值的欄座標**。原版的標籤位置與資料欄
// 差了一兩格（`RosterHeader` 的註解），照抄那個偏移在中文下會看起來像沒對準。
// 值的欄座標一格都沒動，動的只有標籤。
func rosterHeaderCJK(ui func(string) []byte) []byte {
	if ui == nil {
		return nil
	}
	var line []byte
	// ⚠ 欄座標的單位是**格**不是 byte：一個中文字兩個 byte、只佔一格
	// （`docs/spec/10` §3）。拿 len 去對欄位會越補越偏。
	pad := func(to int) {
		for cjkCells(line) < to {
			line = append(line, ' ')
		}
	}
	for _, f := range []struct {
		col  int
		name string
	}{
		{colName, "combat.hdrname"},
		{colAC, "combat.hdrac"},
		{colAmmo, "combat.hdrammo"},
		{colMaxCON, "combat.hdrmax"},
		{colCON, "combat.hdrcon"},
		{colWeapon, "combat.hdrweapon"},
	} {
		t := ui(f.name)
		if len(t) == 0 {
			return nil // 少一條就整條退回英文，不要拼出半中半英的表頭
		}
		pad(f.col)
		line = append(line, t...)
	}
	return line
}

// rosterRowCJK 組戰鬥名單資料行的中文版。
//
// 只有兩欄會變成中文：**傷勢狀態字**（`UNC`／`SER`／`CRT`／`MRT`／`COM`）
// 與**武器名**。數字欄與名字欄本來就不必翻。
//
// 查不到翻譯就回 nil，呼叫端畫英文那一行——**不要拼出半中半英的名單**。
// 死亡那一格是字模不是文字（`game.WoundDead`），原樣送過去。
func rosterRowCJK(r RosterRow, ui func(string) []byte, item func(byte) []byte,
	weaponID byte, hasWeapon bool) []byte {
	if ui == nil {
		return nil
	}
	con := []byte(r.CON)
	if key, ok := woundKeys[r.CON]; ok {
		t := ui(key)
		if len(t) == 0 {
			return nil
		}
		con = t
	}
	weapon := []byte(r.Weapon)
	if hasWeapon && item != nil {
		if t := item(weaponID); len(t) > 0 {
			weapon = t
		}
	}

	var line []byte
	pad := func(to int) {
		for cjkCells(line) < to {
			line = append(line, ' ')
		}
	}
	for _, f := range []struct {
		col  int
		text []byte
	}{
		{colName, []byte(r.Name)},
		{colAC, []byte(r.AC)},
		{colAmmo, []byte(r.Ammo)},
		{colMaxCON, []byte(r.MaxCON)},
		{colCON, con},
		{colWeapon, weapon},
	} {
		if len(f.text) == 0 {
			continue
		}
		pad(f.col)
		line = append(line, f.text...)
	}
	return line
}

// woundKeys 把狀態字對到翻譯目錄的鍵。
//
// ⚠ **死亡（`game.WoundDead`）不在裡面**：那一格是字型第 `0x7F` 格的骷髏字模，
// 不是可以翻譯的文字（`docs/re/17` §4.4）。
var woundKeys = map[string]string{
	"UNC": "wound.unc", "SER": "wound.ser", "CRT": "wound.crt",
	"MRT": "wound.mrt", "COM": "wound.com",
}

// cjkCells 是這串 Big5 佔幾格。
func cjkCells(b []byte) int {
	n := 0
	for i := 0; i < len(b); i++ {
		if b[i] >= 0x80 {
			i++
		}
		n++
	}
	return n
}
