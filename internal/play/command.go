package play

// 地圖指令列（`docs/re/91`、`docs/re/72` §4）。
//
// 原版底部固定印著 `USE ENC ORDER DISBAND VIEW SAVE RADIO`，七項各一個
// 處理程式（`ds:AB1Ch` 的七個 word）。這裡實作已經 RE 確認的那些；
// 其餘的**不猜行為**，按下去只印一句「還沒接上」。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/textlayout"
)

// 這一檔用到的原版字串編號（字串表 1）。
const (
	strNoMoreDisband = 21 // `No more can disband.`
	strWhoDisbands   = 22 // `Who wants to disband?`
)

// Command 是指令列的一項，順序就是原版字串的順序（`docs/re/91` §1）。
type Command int

const (
	CmdUse Command = iota
	CmdEnc
	CmdOrder
	CmdDisband
	CmdView
	CmdSave
	CmdRadio
	CommandCount
)

// CommandNames 是原版指令列印出來的字（`ds:A9CCh`）。
//
// **熱鍵不跟著翻譯走**（`docs/re/40` §4 的硬規則），所以這裡是原文；
// 中文化要顯示的字走翻譯目錄，比對永遠用這一份的首字母。
var CommandNames = [CommandCount]string{
	"Use", "Enc", "Order", "Disband", "View", "Save", "Radio",
}

// commandKey 是每一項的熱鍵（字串的首字母，大寫化之後比對）。
func commandKey(c Command) byte {
	if c < 0 || c >= CommandCount {
		return 0
	}
	return CommandNames[c][0] &^ 0x20 // 轉大寫
}

// CommandFor 把一個按鍵翻成指令；不是指令就回 −1。
func CommandFor(ch byte) Command {
	up := input.Upper(ch)
	for i := Command(0); i < CommandCount; i++ {
		if commandKey(i) == up {
			return i
		}
	}
	return -1
}

// runCommand 執行指令列的一項。回傳 false 表示要離開遊戲。
func (s *Scene) runCommand(c Command) (bool, error) {
	switch c {
	case CmdSave:
		return s.cmdSave()
	case CmdRadio:
		return s.cmdRadio()
	case CmdUse:
		s.beginUse()
		return true, nil
	case CmdView:
		return s.cmdView()
	case CmdOrder:
		s.beginOrder()
		return true, nil
	case CmdDisband:
		return s.cmdDisband()
	case CmdEnc:
		return s.cmdEnc()
	default:
		// **不猜**：入口已經在 `docs/re/91` §1 定位，行為還沒讀。
		s.message = fmt.Sprintf("%s: not wired yet", CommandNames[c])
		s.dirty = true
		return true, nil
	}
}

// cmdSave 是 `Save`（`0x1A290`）：問一句 Y／N，答 Y 才把狀態寫回存檔。
//
// 確認流程與 `Radio` 共用同一套（`sub_19B4F`，`docs/re/91` §2）——
// 原版只有訊息編號不同。
//
// ⚠ **一定要真的寫到檔案。** 只更新記憶體那份明文的話，畫面上會出現
// 「Game saved.」而下一次開遊戲什麼都沒保留——這種謊比當掉還難查。
// 沒有可寫的目錄時就照實說（`SetSaveDir` 沒設 ＝ 無頭工具與測試）。
func (s *Scene) cmdSave() (bool, error) {
	if s.save == nil {
		s.sayEN("No save loaded.", "save.none")
		return true, nil
	}
	return s.askConfirm(confirmSaveString, s.doSave)
}

// doSave 是答 Y 之後真的寫檔那一段。
func (s *Scene) doSave() (bool, error) {
	if err := s.StoreTo(s.save); err != nil {
		s.message = "ERROR: " + err.Error()
		s.dirty = true
		return true, nil
	}
	switch {
	case s.saveDir == "":
		s.sayEN("Game state updated (not written to disk).", "save.memoryonly")
	case s.rom == nil:
		s.sayEN("Game state updated (no data files loaded).", "save.nodata")
	default:
		// 物品表（店家庫存）與存檔是同一個檔案裡的兩個資源，
		// 先把它蓋回記憶體，再由 WriteSave 一次寫出去。
		if len(s.itemsRaw) > 0 {
			if err := s.rom.SetItemTable(s.save.File, 0, s.itemsRaw); err != nil {
				s.message = "SAVE FAILED: " + err.Error()
				s.cjk = nil
				s.dirty = true
				return true, nil
			}
		}
		if err := s.rom.WriteSave(s.save, s.saveDir); err != nil {
			s.message = "SAVE FAILED: " + err.Error()
			s.cjk = nil
		} else {
			s.sayEN("Game saved.", "save.done")
		}
	}
	s.dirty = true
	return true, nil
}

// cmdRadio 是 `Radio`（`0x15260` → `loc_1B8AD`）：呼叫總部升級。
//
// 兩輪都做（`docs/re/91` §2.1、`docs/re/96` §5）：
//
//	第一輪 —— 參與過摧毀 Base Cochise（`+0x4B` bit0）而還沒被表揚過
//	          （`+0x4C` bit0）的人，總部念一次賀詞，然後把 `+0x4C` 設起來
//	第二輪 —— 經驗值夠就升級（`docs/re/31` §1、§2），可以連升
//
// 倒下的人兩輪都跳過（`loc_172BE` ＝ CON ≤ 0，`docs/re/89` §2）。
func (s *Scene) cmdRadio() (bool, error) {
	// 與 `Save` 共用同一套確認（`sub_19B4F`）——原版只有訊息編號不同。
	return s.askConfirm(confirmRadioString, s.doRadio)
}

// doRadio 是答 Y 之後真的呼叫總部那一段。
func (s *Scene) doRadio() (bool, error) {
	var lines []string
	total := 0

	// 第一輪：表揚。原版的賀詞是階級表（`ds:D622h`）的第 0x3D 條，
	// **一個人只聽得到一次**——`ds:D436h` 讓它整場只印一次。
	praise := 0
	for _, m := range s.world.Party.Members {
		if m == nil || m.Down() || !m.Mission || m.Praised {
			continue
		}
		m.Praised = true
		praise++
	}
	if praise > 0 {
		lines = append(lines, s.rankString(RadioPraiseString))
		s.playSound(4)
	}
	for _, m := range s.world.Party.Members {
		if m == nil || m.Down() {
			continue
		}
		if n := m.LevelUp(); n > 0 {
			total += n
			lines = append(lines, fmt.Sprintf("%s is now level %d.", m.Name, m.Level))
		}
	}
	if total == 0 && praise == 0 {
		s.sayEN("HQ: nothing to report.", "radio.nothing")
	} else if len(lines) == 0 {
		s.sayEN("HQ: nothing to report.", "radio.nothing")
	} else {
		s.message = lines[0]
		if len(lines) > 1 {
			s.message = fmt.Sprintf("%s (+%d more)", lines[0], len(lines)-1)
		}
		// 升級的號角（音效 4，`sub_1B8A0` 在 `0x1B90E`）。
		s.playSound(4)
	}
	s.dirty = true
	return true, nil
}

// cmdView 是 `View`（`0x160A8`）：切到下一支隊伍。
//
// 原版從目前這一組往後找下一組能切的，**繞回起點就把畫面切回原本那組**
// ——也就是「什麼都沒發生」（`docs/re/93` §3）。remake 只有第 0 組
// （`docs/spec/21` §4），所以永遠走那條路；接上多隊伍之後這一支不用改。
func (s *Scene) cmdView() (bool, error) {
	n, ok := s.nextGroup()
	if !ok {
		// 繞回起點：原版就是把畫面切回原本那組，什麼都沒發生。
		s.sayEN("No other party.", "view.none")
		return true, nil
	}
	if err := s.SwitchGroup(n); err != nil {
		s.message = "ERROR: " + err.Error()
		s.dirty = true
		return true, nil
	}
	s.message = fmt.Sprintf("Party %d.", n+1)
	s.cjkFmt("view.switched", n+1)
	s.dirty = true
	return true, nil
}

// cmdDisband 是 `Disband`（`0x15E77`）：把一個人分出去自成一組。
//
// **一個人不能分**、已經有四組就不能再分（`docs/re/93` §2）。
func (s *Scene) cmdDisband() (bool, error) {
	alive := 0
	for _, m := range s.world.Party.Members {
		if m != nil {
			alive++
		}
	}
	if alive <= 1 {
		s.sayEN("Can't disband a single ranger.", "disband.single")
		return true, nil
	}
	if _, ok := s.freeGroup(); !ok {
		// 原版字串 21：`No more can disband.`
		s.say(strNoMoreDisband, textlayout.Options{})
		return true, nil
	}
	s.disband = true
	// 原版字串 22：`Who wants to disband?`，後面接隊員清單。
	s.message = "Who leaves? " + s.memberMenu()
	if zh := s.cjkExe(exeTable1, strWhoDisbands, textlayout.Options{}); zh != nil {
		s.cjk = append(append([]byte{}, zh...), []byte(" "+s.memberMenu())...)
		s.message = ""
	}
	s.dirty = true
	return true, nil
}

// updateDisband 是分隊進行中的按鍵（選一個人）。
func (s *Scene) updateDisband(in input.Input) (bool, error) {
	if in.Action == input.ActionCancel {
		s.disband = false
		s.message = ""
		s.dirty = true
		return true, nil
	}
	ch := input.Upper(in.Char)
	if ch < '1' || ch > '9' {
		return true, nil
	}
	i := int(ch - '1')
	if i >= len(s.world.Party.Members) || s.world.Party.Members[i] == nil {
		return true, nil
	}
	name := s.world.Party.Members[i].Name
	s.disband = false
	if err := s.disbandMember(i); err != nil {
		s.message = "ERROR: " + err.Error()
	} else {
		s.message = name + " leaves the party."
		s.cjkFmt("disband.left", name)
	}
	s.dirty = true
	return true, nil
}

// RadioPraiseString 是總部的賀詞在**階級表**（`ds:D622h`，`ExeStrings()`
// 第 5 張）裡的編號。`sub_1BB5D` 印之前會把 `ds:4692h` 換成那張表——
// ⚠ 拿第 1 張表去查 0x3D 會取到 `"That doesn't seem to work."`，
// 而那句在畫面上完全說得通（`docs/re/96` §5）。
const RadioPraiseString = 0x3D

// rankString 取階級表（`ds:D622h`）的第 n 條。
func (s *Scene) rankString(n int) string {
	tables, err := s.rom.ExeStrings()
	if err != nil || len(tables) < 6 || n < 0 || n >= len(tables[5]) {
		return ""
	}
	return tables[5][n]
}

// commandBar 是要畫在畫面底部的那一行。
func commandBar() string {
	out := ""
	for i := Command(0); i < CommandCount; i++ {
		if i > 0 {
			out += " "
		}
		out += CommandNames[i]
	}
	return out
}
