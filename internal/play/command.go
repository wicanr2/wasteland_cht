package play

// 地圖指令列（`docs/re/91`、`docs/re/72` §4）。
//
// 原版底部固定印著 `USE ENC ORDER DISBAND VIEW SAVE RADIO`，七項各一個
// 處理程式（`ds:AB1Ch` 的七個 word）。這裡實作已經 RE 確認的那些；
// 其餘的**不猜行為**，按下去只印一句「還沒接上」。

import (
	"fmt"

	"github.com/wicanr2/wasteland_cht/internal/input"
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
	default:
		// **不猜**：入口已經在 `docs/re/91` §1 定位，行為還沒讀。
		s.message = fmt.Sprintf("%s: not wired yet", CommandNames[c])
		s.dirty = true
		return true, nil
	}
}

// cmdSave 是 `Save`（`0x1A290`）：把目前狀態寫回存檔。
//
// 原版先印訊息 0x49 問 Y／N 再寫；這一版直接寫，**確認流程還沒接**
// （與 Radio 共用的 `sub_19B4F` 是同一套，兩邊要一起做）。
func (s *Scene) cmdSave() (bool, error) {
	if s.save == nil {
		s.message = "No save loaded."
		s.dirty = true
		return true, nil
	}
	if err := s.StoreTo(s.save); err != nil {
		s.message = "ERROR: " + err.Error()
		s.dirty = true
		return true, nil
	}
	s.message = "Game saved."
	s.dirty = true
	return true, nil
}

// cmdRadio 是 `Radio`（`0x15260` → `loc_1B8AD`）：呼叫總部升級。
//
// 只做第二輪（經驗值 → 等級，`docs/re/31` §1、§2）。
// **第一輪沒做**：那一段看角色記錄 `+0x4B`／`+0x4C` 的 bit0，
// 兩個欄位的語意未解（`docs/re/91` §2.1），不猜。
//
// 倒下的人跳過（`loc_172BE` ＝ CON ≤ 0，`docs/re/89` §2）。
func (s *Scene) cmdRadio() (bool, error) {
	var lines []string
	total := 0
	for _, m := range s.world.Party.Members {
		if m == nil || m.Down() {
			continue
		}
		if n := m.LevelUp(); n > 0 {
			total += n
			lines = append(lines, fmt.Sprintf("%s is now level %d.", m.Name, m.Level))
		}
	}
	if total == 0 {
		s.message = "HQ: nothing to report."
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
