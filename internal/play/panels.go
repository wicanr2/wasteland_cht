package play

// F1 說明、F2 設定、F10 離開確認 —— 三個覆蓋在遊戲上的面板。
//
// 這三個都是**重製版自己加的**，原版沒有。原版只有一個「主選單」而且只有
// `Start` 一項（`docs/re/95`），離開遊戲靠 DOS。
//
// 離開的語意照 `esc-cancel-f10-quit-autosave` 的鐵則：
//
//	ESC  只取消／退一層，**任何一層都不會結束遊戲**
//	F10  唯一的離開手勢 → 先跳 Y／N → 選 Y 才**先存檔再退出**
//
// 存檔失敗時**不退出**，把錯誤留在畫面上——默默吞掉存檔失敗比當掉更難查。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// helpLines 是 F1 說明的內容（英文後備；中文走 `ui:help.*`）。
//
// **熱鍵字母不跟著翻譯走**（`docs/re/40` §4）：這裡列的字母是 Go 這邊的常數，
// 譯文只換說明文字。
//
// ⚠ 兩個限制決定了這張表的長度與字母欄的內容：
//
//   - **訊息視窗只有六行**（`docs/re/25` §2）。標題佔一行，所以最多五條。
//     多寫的不會捲動也不會報錯，**會安靜地被切掉**。
//   - **字母欄只能放 ASCII**。這一欄是直接接進 Big5 位元組串的，
//     寫中文會被當成 Big5 解讀成亂碼（`方向鍵` 會變成三個不相干的字）。
//     中文一律走譯文那一欄。
var helpLines = []struct{ key, en, ui string }{
	{"IKJL", "Move (arrow keys too)", "help.move"},
	{"U E O D V S R  P", "Command bar / journal", "help.cmdbar"},
	{"F1  F2", "Help / settings", "help.panels"},
	{"F5  F9", "Quick save / quick load", "help.f5f9"},
	{"F10 ESC", "Quit (asks, saves) / cancel", "help.quit"},
}

// openHelp／openSettings 由功能鍵路由進來（見 Scene.Update）。
func (s *Scene) openHelp() {
	s.help = true
	s.showHelp()
}

func (s *Scene) showHelp() {
	var en strings.Builder
	var zh []byte
	title := s.uiText("help.title")
	if len(title) > 0 {
		zh = append(zh, title...)
		zh = append(zh, '\r')
	} else {
		en.WriteString("KEYS\r")
	}
	for _, l := range helpLines {
		en.WriteString(l.key + "  " + l.en + "\r")
		if t := s.uiText(l.ui); len(t) > 0 && zh != nil {
			zh = append(zh, l.key...)
			zh = append(zh, ' ')
			zh = append(zh, t...)
			zh = append(zh, '\r')
		}
	}
	s.message, s.cjk = en.String(), zh
	if zh != nil {
		s.message = ""
	}
	s.dirty = true
}

// updateHelp：任何鍵都關掉（F1 再按一次也是關）。
func (s *Scene) updateHelp(in input.Input) (bool, error) {
	if in.Dir == input.DirNone && in.Action == input.ActionNone &&
		in.Char == 0 && in.Fn == input.FnNone {
		return true, nil
	}
	s.help = false
	s.message, s.cjk = "", nil
	s.dirty = true
	return true, nil
}

// settingsState 是 F2 的設定。
//
// 這一版只有音樂那一組——**原版沒有背景音樂**（九首 PC 喇叭音效而已，
// `docs/re/44`），BGM 是重製版加的，所以「能關掉」是必要的禮貌。
type settingsState struct {
	MusicOn  bool
	MusicVol int // 0–10
	SFXOn    bool
}

func defaultSettings() settingsState {
	return settingsState{MusicOn: true, MusicVol: 6, SFXOn: true}
}

// Settings 讓呈現層讀設定（音樂音量、開關）。
func (s *Scene) Settings() settingsState { return s.settings }

func (s *Scene) openSettings() {
	s.settingsOpen = true
	s.showSettings()
}

func (s *Scene) showSettings() {
	onoff := func(b bool) (string, string) {
		if b {
			return "ON", "settings.on"
		}
		return "OFF", "settings.off"
	}
	mEN, mUI := onoff(s.settings.MusicOn)
	xEN, xUI := onoff(s.settings.SFXOn)
	s.message = fmt.Sprintf("SETTINGS\rM Music: %s\r- / + Volume: %d\rX Sound: %s\rESC close",
		mEN, s.settings.MusicVol, xEN)
	var zh []byte
	if t := s.uiText("settings.title"); len(t) > 0 {
		zh = append(zh, t...)
		zh = append(zh, '\r')
		zh = appendUILine(zh, s.uiText("settings.music"), "M", s.uiText(mUI))
		zh = appendUILine(zh, s.uiText("settings.volume"), "- +",
			[]byte(fmt.Sprintf("%d", s.settings.MusicVol)))
		zh = appendUILine(zh, s.uiText("settings.sfx"), "X", s.uiText(xUI))
		zh = append(zh, s.uiText("settings.close")...)
		s.message = ""
	}
	s.cjk = zh
	s.dirty = true
}

// appendUILine 組一行「〔熱鍵〕標籤：值」。熱鍵是空字串就不印。
//
// ⚠ 熱鍵**只能是 ASCII**：這一行會接進 Big5 位元組串（同 helpLines 的字母欄）。
func appendUILine(dst, label []byte, key string, value []byte) []byte {
	if key != "" {
		dst = append(dst, key...)
		dst = append(dst, ' ')
	}
	dst = append(dst, label...)
	dst = append(dst, value...)
	return append(dst, '\r')
}

// updateSettings 收設定畫面的按鍵。
//
// ⚠ **ESC 在這裡是關面板，不是離開遊戲**（鐵則 1）。
func (s *Scene) updateSettings(in input.Input) (bool, error) {
	if in.Action == input.ActionCancel || in.Fn == input.FnSettings {
		s.settingsOpen = false
		s.message, s.cjk = "", nil
		s.dirty = true
		return true, nil
	}
	switch input.Upper(in.Char) {
	case 'M':
		s.settings.MusicOn = !s.settings.MusicOn
	case 'X':
		s.settings.SFXOn = !s.settings.SFXOn
	case '+', '=':
		if s.settings.MusicVol < 10 {
			s.settings.MusicVol++
		}
	case '-', '_':
		if s.settings.MusicVol > 0 {
			s.settings.MusicVol--
		}
	default:
		return true, nil
	}
	s.showSettings()
	return true, nil
}

// openQuit 是 F10：**不直接離開**，先問一句（鐵則 3）。
func (s *Scene) openQuit() {
	s.quitAsk = true
	s.message = "Quit? It will save first.  Y / N"
	if t := s.uiText("quit.ask"); len(t) > 0 {
		s.cjk = t
		s.message = ""
	}
	s.dirty = true
}

// updateQuit 是離開確認的按鍵。
//
// 順序不可顛倒（鐵則 4）：**先存檔、存成功了才退出**。
// 存檔失敗就留在畫面上並取消離開——玩家選 Y 是因為相信我們不會吃他的存檔。
func (s *Scene) updateQuit(in input.Input) (bool, error) {
	switch {
	case input.Upper(in.Char) == 'Y' || in.Action == input.ActionConfirm:
		if err := s.quitSave(); err != nil {
			s.quitAsk = false
			s.message = "SAVE FAILED: " + err.Error()
			s.cjk = s.cjkFmtText("quit.savefailed", err.Error())
			s.dirty = true
			return true, nil // **不退出**
		}
		return false, nil
	case input.Upper(in.Char) == 'N' || in.Action == input.ActionCancel:
		s.quitAsk = false
		s.message, s.cjk = "", nil
		s.dirty = true
	}
	return true, nil
}

// quitSave 是離開前的自動存檔。
//
// 標題畫面還沒進遊戲就沒東西可存（鐵則 4 的例外）。
// 有快速存檔路徑就寫那一份；沒有就退回指令列 `Save` 的那條路（`-save-dir`）。
// 兩條都沒有時**不算失敗**——無頭工具與測試就是這樣跑的。
func (s *Scene) quitSave() error {
	if s.title || s.save == nil {
		return nil
	}
	if s.quickPath != "" {
		return s.QuickSave()
	}
	if s.saveDir == "" || s.rom == nil {
		return nil
	}
	if err := s.StoreTo(s.save); err != nil {
		return err
	}
	if len(s.itemsRaw) > 0 {
		if err := s.rom.SetItemTable(s.save.File, 0, s.itemsRaw); err != nil {
			return err
		}
	}
	return s.rom.WriteSave(s.save, s.saveDir)
}

// cjkFmtText 是 cjkFmt 的「回傳而不是寫進 s.cjk」版本。
func (s *Scene) cjkFmtText(name string, args ...any) []byte {
	f := s.uiText(name)
	if len(f) == 0 {
		return nil
	}
	return []byte(fmt.Sprintf(string(f), args...))
}
