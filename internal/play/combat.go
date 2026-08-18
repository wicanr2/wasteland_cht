package play

import (
	"unicode/utf8"
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/render"
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

// 名單一行的欄座標（docs/re/15 §4）。行首是序號 ＋ `>`，名字從欄 2 起。
const (
	colIndex   = 0 // 序號 ＋ `>`（`0x1709A`／`0x170A0`）
	colName    = 2
	colAC      = 0x11
	colAmmo    = 0x15
	colMaxCON  = 0x18
	colCON     = 0x1C
	colWeapon  = 0x20
	// rosterCols ＝ **39** 不是 38：武器欄從 `0x20` 起，而原版畫面上
	// `Crowbar`（7 個字）整個看得到（`docs/re/103` §4 的實機截圖）。
	// 少一格會把它切成 `Crowba`——而那看起來只是「名字比較長」。
	rosterCols = 39
)

// 中文名單的欄座標。**與英文那一套不同**，因為兩邊要塞的東西不一樣長。
//
// 數值欄各留三格就夠：`AMM` 是 `slot.Value & 0x3F`（0–63，兩位數封頂），
// `AC` 與 `MAX` 實務上兩到三位數，而中文表頭全部是兩個字（`ui:combat.hdr*`）。
// 英文那一套各留四格，省下來的四格連同尾巴那一格全部給武器欄。
//
// 名字欄 12 格：原版的名字欄位只有 13 bytes（`internal/input.MaxName` 的說明），
// 出廠隊伍最長的是 `Snake Vargas`（12）。
//
// 武器欄 14 格是**量出來的**：`translations/zh-Hant/exe-skills-items.tsv`
// 裡最長的單數名字是「M1989A1 北約突擊步槍」14 格。給 13 格的話它會變成
// 「M1989A1 北約突擊步」——而畫面上看起來只是「這把武器叫這個名字」。
const (
	cjkColName   = 2
	cjkColAC     = 14
	cjkColAmmo   = 17
	cjkColMaxCON = 20
	cjkColCON    = 23
	cjkColWeapon = 26
	// cjkRosterCols ＝ 40 ＝ 整個畫面的字元欄數（`render.ScreenWidth / CharWidth`）。
	// 英文那一版停在 39，最後一格原版沒有用到；中文這一版用掉它，
	// 武器欄才有 14 格。名單那幾列上沒有別的東西會被蓋到。
	cjkRosterCols = render.ScreenWidth / render.CharWidth
)

// rosterLayout 是一套欄座標。英文與中文各一套，繪製、表頭與反白範圍共用
// ——三處各自寫一份的話，反白會慢慢漂到隔壁欄，而畫面上只是「反白的位置有點怪」。
type rosterLayout struct{ name, ac, ammo, maxCON, con, weapon, cols int }

var (
	enRoster = rosterLayout{colName, colAC, colAmmo, colMaxCON, colCON,
		colWeapon, rosterCols}
	cjkRoster = rosterLayout{cjkColName, cjkColAC, cjkColAmmo,
		cjkColMaxCON, cjkColCON, cjkColWeapon, cjkRosterCols}
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
	// Index 是名單上的序號（1 起算）。原版印在行首，後面接一個 `>`
	// （`0x1709A` 印數字、`0x170A0` 印 `0x3E`）。
	Index  int
	Name   string
	AC     string
	Ammo   string
	MaxCON string
	CON    string // CON ≤ 0 時是狀態字而不是數字
	Weapon string

	// CONInverse／WeaponInverse 是原版的「這一格有問題」標記
	// （`sub_19E2A` → `ds:4678h`，`docs/re/111` §1）：
	// 角色記錄 `+0x28` 的狀態位元非 0 → **`MAX` 那一欄**反白；
	// 裝備武器附屬 byte 的 bit7（卡彈）→ 武器名反白。
	//
	// ⚠ 名字留 `CONInverse` 是因為條件來自體力那一族（狀態位元），
	// **但反白的是 `MAX` 欄**——`sub_1708B` 在印 MAXCON 之前開、之後就關。
	CONInverse    bool
	WeaponInverse bool
}

// InverseAt 回答「這一行的第 col 欄要不要反白」。
//
// `lay` 是這一行用哪一套欄座標，`maxCON`／`weapon` 是**這一次真的要畫的那兩段字**
// ——中文與英文長度不同，拿錯就會反白到隔壁欄或反白不足。
//
// ⚠ **狀態位元反白的是 `MAX` 欄，不是 `CON` 欄。** `sub_1708B` 的順序是
// `0x17102` 開 → `0x1711C` 印 MAXCON → `0x1711F` 關，CON 那一欄在關掉之後才印
// （`docs/re/111` §1）。
//
// ⚠ 只反白欄位本身的字，不含後面的補白——**這是從程式碼讀出來的，不是取保守**：
// 欄與欄之間是把游標欄寫進 `ds:4672h`（`0x17105`、`0x17122`、`0x1715D`），
// 中間那幾格根本沒有被印過，所以反白旗標碰不到它們。
func (r RosterRow) InverseAt(lay rosterLayout, maxCON, weapon string) func(col int) bool {
	type span struct{ lo, hi int }
	var spans []span
	if r.CONInverse {
		spans = append(spans, span{lay.maxCON, lay.maxCON + utf8.RuneCountInString(maxCON)})
	}
	if r.WeaponInverse {
		spans = append(spans, span{lay.weapon, lay.weapon + utf8.RuneCountInString(weapon)})
	}
	if len(spans) == 0 {
		return nil
	}
	return func(col int) bool {
		for _, sp := range spans {
			if col >= sp.lo && col < sp.hi {
				return true
			}
		}
		return false
	}
}

// Text 把一行排成 39 欄的字串（欄座標照 docs/re/15 §4）。
func (r RosterRow) Text() string {
	line := []byte(spaces(rosterCols))
	put := func(col int, s string) {
		for i := 0; i < len(s) && col+i < rosterCols; i++ {
			line[col+i] = s[i]
		}
	}
	if r.Index > 0 {
		put(colIndex, fmt.Sprintf("%d>", r.Index))
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
		slot, hasWeapon := equippedSlot(m)
		rows = append(rows, RosterRow{
			Index:  len(rows) + 1,
			Name:   m.Name,
			AC:     fmt.Sprintf("%d", m.AC),
			Ammo:   ammoColumn(m, items),
			MaxCON: fmt.Sprintf("%d", m.MaxCON),
			CON:    con,
			Weapon: weaponColumn(m, name),
			// 反白的兩個條件（`docs/re/111` §1）。
			CONInverse:    m.Status != 0,
			WeaponInverse: hasWeapon && game.Jammed(slot),
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
// ⚠ **原版不是塞進訊息視窗的**：戰鬥的選單畫在欄 15–38、列 1–13 那一塊
// （`docs/re/105` §2），八行綽綽有餘。這一支只回報放不放得下，
// 不自己決定捲動或分頁——那是呈現層的事。
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

	// pick 是「換武器」清單開著時的狀態（`docs/re/107` §2）。
	pick weaponPick
	// hire 是「哪一組？」開著時的狀態（`docs/re/110`）。
	hire hirePick
	// EncX／EncY 是這場遭遇**在地圖上的那一格**（0 ＝ 沒有格子，
	// 空地 `ENC` 的那種回合就是）。敵人在地圖上移動時會改它
	// （`docs/re/116` §5：移動是「把那一格搬過去」，不是加一個新的）。
	EncX, EncY int
	// EncRecord 是這場遭遇那一格的記錄——雇用要讀它的 `+0x09`
	// （`docs/re/110` §2）。**空的話 `H` 一律失敗**，不猜一個 NPC 出來。
	EncRecord []byte
	// useID 是每個成員這一回合 `USE` 選中的技能／物品／屬性編號。
	//
	// ⚠ 原版存在 `ds:A9FDh`，索引是**角色編號**不是隊伍槽（`docs/re/108` §1）——
	// 那是為了讓同一個角色換到別的槽也認得。這裡一回合就用完，用槽號等價。
	useID [combatPartyMax]byte

	// World 是規則層——戰鬥版 `USE` 要拿它取目標格的記錄並改寫
	// （`docs/re/108` §2）。沒接上時 `USE` 什麼都不做。
	World *game.World

	// Items 是物品資料表（存檔區那一份，docs/re/45 §2）。
	// 沒接上時是 nil，武器一律當成零值——傷害 0 而不是崩掉。
	Items game.ItemTable
	// Names 是六個敵人種類的名稱，訊息要用（`docs/re/85`）。
	Names EnemyNames
	// CJKNames 是同六個種類的中文名（Big5），與 Names 同一組編號。
	CJKNames [6]string
	// BlockNames 是這張地圖的明文敵人名字表（`docs/re/114` §6），索引 1 起算。
	// 遭遇記錄 `+0x09` 的 bit0 設起來時用它——那才是原版畫面上的名字。
	BlockNames []string
	// MonsterCJK 查一條明文名字的中文（目錄 key ＝ `monster:<原文>`）。
	// nil 或查不到就走英文那一份。
	MonsterCJK func(raw string) string

	// CJK 查執行檔字串表 1 第 n 條的譯文（Big5，控制碼已解）。
	// 由 `Scene` 開場時接上；nil ＝ 沒有中文，畫面走英文那一份。
	CJK func(n int, opt textlayout.Options) string
	// UI 查重製版自己的介面文字（`translations/*/ui.tsv`）。
	// 原版沒有對應字串、或組法是重製版自己決定的句子走這一支。
	UI func(name string) string

	// LastCJK 是 `Log` 最後一句的中文（Big5）。指令階段那幾句
	//（卡彈、逃跑）不走 `ResolveRound`，所以另外留一格。
	LastCJK string
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
func (s *CombatScene) zhStr(n int, opt textlayout.Options) string {
	if s.CJK == nil {
		return ""
	}
	return s.CJK(n, opt)
}

// zhJoin 把幾段接起來；**任何一段是空的就整句放棄**——
// 半句中文半句英文比整句英文更難讀，也讓「哪裡沒翻」看不出來。
func zhJoin(parts ...string) string {
	var out string
	for _, p := range parts {
		if p == "" {
			return ""
		}
		out += p
	}
	return out
}

// zhEnemy 是敵人的中文名（UTF-8）。查不到回空字串，整句就跟著放棄。
func (s *CombatScene) zhEnemy(e *game.Enemy) string {
	if raw := s.properName(e); raw != "" && s.MonsterCJK != nil {
		if zh := s.MonsterCJK(raw); zh != "" {
			return singular(zh)
		}
	}
	k := int(e.Data.Kind)
	if k < 0 || k >= len(s.CJKNames) || len(s.CJKNames[k]) == 0 {
		return ""
	}
	return s.CJKNames[k]
}

// properName 是這一隻在**這張地圖的明文名字表**裡的名字（含 `\n` 分段）。
//
// 兩個條件都要成立：遭遇記錄 `+0x09` 的 bit0 設著（`0x129E9` 的 `shr`），
// 而且名字表裡有那個編號。任一個不成立就回空字串，呼叫端走種類名。
func (s *CombatScene) properName(e *game.Enemy) string {
	if e == nil || !game.UseProperName(s.EncRecord) {
		return ""
	}
	i := int(e.Type)
	if i <= 0 || i >= len(s.BlockNames) {
		return ""
	}
	return s.BlockNames[i]
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
	if cmd == game.CmdUse {
		// `USE` 走與地圖同一套三層選單，只是不用挑人、選完要再問方向
		// （`docs/re/108` §1）。開選單這一步在 `Scene` 那一層。
		return false // 由 Scene.updateCombat 接手，這個人還沒下完令
	}
	if cmd == game.CmdHire {
		// 開「哪一組？」（`docs/re/110`）。候選是**還有敵人的組**，
		// 能不能雇用要到結算才知道——原版也是這個順序。
		return s.beginHirePick()
	}
	if cmd == game.CmdWeapon {
		// ⚠ **換武器要先選一件**：參數留 0 的話結算階段什麼都不會做
		// （`slotOf(0)` ＝ 沒有這一格），而畫面上看不出任何異狀。
		return s.beginWeaponPick()
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
func rosterHeaderCJK(ui func(string) string) string {
	if ui == nil {
		return ""
	}
	var line string
	// ⚠ 欄座標的單位是**格**不是 byte：一個中文字兩個 byte、只佔一格
	// （`docs/spec/10` §3）。拿 len 去對欄位會越補越偏。
	pad := func(to int) {
		for cjkCells(line) < to {
			line += " "
		}
	}
	for _, f := range []struct {
		col  int
		name string
	}{
		{cjkRoster.name, "combat.hdrname"},
		{cjkRoster.ac, "combat.hdrac"},
		{cjkRoster.ammo, "combat.hdrammo"},
		{cjkRoster.maxCON, "combat.hdrmax"},
		{cjkRoster.con, "combat.hdrcon"},
		{cjkRoster.weapon, "combat.hdrweapon"},
	} {
		t := ui(f.name)
		if t == "" {
			return "" // 少一條就整條退回英文，不要拼出半中半英的表頭
		}
		pad(f.col)
		line += t
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
func rosterRowCJK(r RosterRow, ui func(string) string, item func(byte) string,
	weaponID byte, hasWeapon bool) string {
	if ui == nil {
		return ""
	}
	con := r.CON
	if key, ok := woundKeys[r.CON]; ok {
		t := ui(key)
		if len(t) == 0 {
			return ""
		}
		con = t
	}
	weapon := r.Weapon
	if hasWeapon && item != nil {
		if t := item(weaponID); len(t) > 0 {
			weapon = t
		}
	}

	var line string
	pad := func(to int) {
		for cjkCells(line) < to {
			line += " "
		}
	}
	var idx string
	if r.Index > 0 {
		idx = fmt.Sprintf("%d>", r.Index)
	}
	for _, f := range []struct {
		col  int
		text string
	}{
		{colIndex, idx},
		{cjkRoster.name, r.Name},
		{cjkRoster.ac, r.AC},
		{cjkRoster.ammo, r.Ammo},
		{cjkRoster.maxCON, r.MaxCON},
		{cjkRoster.con, con},
		{cjkRoster.weapon, weapon},
	} {
		if f.text == "" {
			continue
		}
		pad(f.col)
		line += f.text
	}
	return clipCells(line, cjkRoster.cols)
}

// clipCells 把一串 UTF-8 截到 n 格（**一個 rune 一格**）。
//
// ⚠ **不能用 len 截**：一個中文字兩個 byte、只佔一格（`docs/spec/10` §3）。
// 拿 byte 數去截會把長名字切在中文字中間，畫面上出現半個字。
func clipCells(s string, n int) string {
	cells := 0
	for i := range s {
		if cells == n {
			return s[:i]
		}
		cells++
	}
	return s
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
func cjkCells(s string) int {
	// UTF-8：一個 rune 一格。**不用管一個字幾個 byte**——
	// 以前走 Big5 得判「≥ 0x80 就吃兩個」，而那個規則對 UTF-8 是錯的
	// （一個漢字三個 byte），改型別的時候漏掉這一支就會整排偏掉。
	n := 0
	for range s {
		n++
	}
	return n
}

// —— 換武器的清單（`docs/re/107` §2）——————————————————————————
//
// ⚠ **原版的清單長什麼樣還沒逆向**：`sub_19394` 的回傳值編碼未解
// （`docs/re/41` §7）。所以這一層的**呈現**是重製版自己的介面
// （數字鍵選、`0` 翻頁，與 `USE` 那份同一套，`docs/spec/25`），
// 但**選完之後做什麼**照原版（`sub_1949E`：裝／卸切換）。

// weaponPickSize 是一頁幾項——九是因為選項用數字鍵 `1`–`9`。
const weaponPickSize = 9

// weaponPick 是「換武器」開著的清單狀態。
type weaponPick struct {
	open  bool
	slots []byte // 1-based 槽號，照背包順序
	page  int
}

func (w *weaponPick) pages() int {
	if len(w.slots) == 0 {
		return 1
	}
	return (len(w.slots) + weaponPickSize - 1) / weaponPickSize
}

func (w *weaponPick) pageSlice() []byte {
	lo := w.page * weaponPickSize
	if lo >= len(w.slots) {
		return nil
	}
	hi := lo + weaponPickSize
	if hi > len(w.slots) {
		hi = len(w.slots)
	}
	return w.slots[lo:hi]
}

// beginWeaponPick 開清單。空手就回 false 並留下原版那句話
// （字串 64「You don't have anything.」，`docs/re/41` §2）。
func (s *CombatScene) beginWeaponPick() bool {
	m := s.Battle.Party.Members[s.Turn]
	if m == nil {
		return false
	}
	var slots []byte
	for i := range m.Items {
		if m.Items[i].ID != 0 {
			slots = append(slots, byte(i+1)) // ⚠ 1-based
		}
	}
	if len(slots) == 0 {
		s.Log = append(s.Log, "You don't have anything.")
		s.LastCJK = s.zhStr(int(game.MsgNothingToUse), textlayout.Options{})
		return false
	}
	s.pick = weaponPick{open: true, slots: slots}
	return true
}

// WeaponPicking 回答清單開著沒有。
func (s *CombatScene) WeaponPicking() bool { return s.pick.open }

// WeaponPickLines 是清單的英文顯示。
func (s *CombatScene) WeaponPickLines(name func(byte) string) []string {
	m := s.Battle.Party.Members[s.Turn]
	out := []string{m.Name + ", which item?"}
	for i, slot := range s.pick.pageSlice() {
		label := ""
		if name != nil {
			label = name(m.Items[slot-1].ID)
		}
		mark := " "
		if slot == m.EquipIndex || slot == m.ArmorIndex {
			mark = "*" // 裝備中（原版清單會標，`docs/re/42` §3.1）
		}
		out = append(out, fmt.Sprintf("%d%s%s", i+1, mark, label))
	}
	if s.pick.pages() > 1 {
		out = append(out, fmt.Sprintf("0 more (%d/%d)", s.pick.page+1, s.pick.pages()))
	}
	return out
}

// PickWeapon 收清單上的一個按鍵。回傳這個按鍵有沒有被吃掉。
func (s *CombatScene) PickWeapon(key byte) bool {
	if !s.pick.open {
		return false
	}
	switch {
	case key == '0' && s.pick.pages() > 1:
		s.pick.page = (s.pick.page + 1) % s.pick.pages()
		return true
	case key >= '1' && key <= '9':
		// ⚠ 數字是**這一頁的第幾項**，不是整份清單的第幾項。
		// 混淆的話翻到第二頁按 1 會裝到第一項，而且完全不會報錯。
		page := s.pick.pageSlice()
		i := int(key - '1')
		if i >= len(page) {
			return true // 這一頁沒有那一項：吃掉按鍵，不要當成指令
		}
		s.pick = weaponPick{}
		s.Phase.Set(s.Turn, game.CmdWeapon, page[i])
		s.advance(s.Turn + 1)
		return true
	}
	return true // 清單開著就吃掉所有按鍵，避免誤觸別的指令
}

// CancelWeaponPick 關掉清單，回到指令選單（原版回傳 0xFF ＝ 取消，重問）。
func (s *CombatScene) CancelWeaponPick() { s.pick = weaponPick{} }

// combatPartyMax 是一支隊伍的人數上限（存檔槽表 `+0x00`–`+0x07`，`docs/re/30` §3）。
const combatPartyMax = 8

// SetUse 記下這個人的 `USE` 指令：參數 ＝ `(選項 << 4) | 方向`，
// 編號另存一格（`docs/re/108` §1）。
func (s *CombatScene) SetUse(member int, kind, id, dir byte) {
	if member < 0 || member >= len(s.useID) {
		return
	}
	s.useID[member] = id
	s.Phase.Set(member, game.CmdUse, kind<<4|dir&0x0F)
	s.advance(member + 1)
}

// UseParts 把指令參數拆回來（結算階段用）。
func (s *CombatScene) UseParts(member int) (kind game.UseKind, id, dir byte) {
	if member < 0 || member >= len(s.useID) || member >= len(s.Phase.Arg) {
		return 0, 0, 0
	}
	arg := s.Phase.Arg[member]
	return game.UseKind(arg >> 4), s.useID[member], arg & 0x0F
}

// wireCombat 把 `CombatScene` 需要的外部東西一次接齊。
//
// ⚠ **兩個建構點（`StartEncounter` 與 `beginEmptyRound`）必須共用這一支。**
// 各自手接的話遲早會漏掉一個欄位，而漏掉的症狀是「那個功能安靜地不做事」：
// 漏 `World` 的時候戰鬥版 `USE` 什麼都不會發生，畫面上完全正常。
func (s *Scene) wireCombat(c *CombatScene) *CombatScene {
	c.Items = s.items
	c.Names = s.enemyNames()
	c.CJKNames = s.enemyNamesCJK()
	if s.world != nil && s.world.Block != nil {
		c.BlockNames = s.world.Block.MonsterNames()
	}
	c.MonsterCJK = s.monsterCJK
	c.CJK = func(n int, opt textlayout.Options) string { return s.cjkExe(exeTable1, n, opt) }
	c.UI = s.uiText
	c.World = s.world
	return c
}
