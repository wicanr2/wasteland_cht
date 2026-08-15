package play

// 主選單（`sub_1630C`，開機序列 `sub_161C0` 的 `0x16207`）。
//
// 它與地圖指令列（`sub_16C7C`）是**同一套雙表結構**，只是短得多：
//
// ```
// ds:468Dh ← 0A6FFh      ; [A6FFh] ＝ A6F4h → 標籤字串 "Start\0"
// ds:468Fh ← 0A701h      ; [A701h] ＝ A6FBh → 處理程式陣列
// ds:468Ch ← 18h
// ds:4689h ← 1           ; **最大索引 1**
// jmp sub_173D2          ; 等按鍵
// ```
//
// 處理程式只有兩支：
//
// | 索引 | 位址 | 做什麼 |
// |---:|---|---|
// | 0 | `0x16330` | `sub_173D2` → `sub_16356` → `ds:46C5h ← 0FFh` → `sub_185E6` → `sub_18744` → `ds:46E0h ← 0FFh` → `jmp loc_16B31`（**地圖迴圈**）|
// | 1 | `0x1632F` | `retn`——什麼都不做，回開機序列繼續播片頭 |
//
// ⚠ **原版沒有「新遊戲／讀檔」。** 標籤字串整個只有 `Start` 一個字，
// 第二支處理程式是空的。存檔就是 `GAME1`／`GAME2` 本身，
// 「開始」＝ 接著上次的狀態走——角色的增刪在 Ranger Center 做
// （`docs/re/72` §3），不在這裡。

import (
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// MainMenuLabel 是原版標籤字串的內容（`ds:A6F4h`）。
//
// **熱鍵比對用它的首字母**，不跟著翻譯走（`docs/re/40` §4）。
const MainMenuLabel = "Start"

// showTitle 為 true 時畫面停在標題畫面加那一個選項。
func (s *Scene) showTitle() bool { return s.title }

// BeginTitle 讓場景從標題畫面開始（`cmd/wasteland` 的正常開場）。
//
// 沒有標題圖也照跑——**開不起來比沒有標題圖嚴重**。
func (s *Scene) BeginTitle() {
	if im, err := s.rom.Title(); err == nil {
		s.titlePic = im
	}
	s.title = true
	s.message = ""
	s.dirty = true
}

// updateTitle 是標題畫面的按鍵。
//
// 原版走 `sub_173D2` 的熱區比對；這裡照它的兩支處理程式做：
// `S` ＝ 開始，其餘鍵什麼都不做（＝ 索引 1 的那支 `retn`）。
// ⚠ ESC **不是**離開遊戲——原版那一支也只是 `retn` 回去繼續播片頭。
func (s *Scene) updateTitle(in input.Input) (bool, error) {
	if input.Upper(in.Char) != MainMenuLabel[0] && in.Action != input.ActionConfirm {
		return true, nil
	}
	s.title = false
	s.titlePic = nil
	s.message = ""
	s.dirty = true
	return true, nil
}

// drawTitle 畫標題圖與那一個選項。
func (s *Scene) drawTitle(f *render.Frame) {
	if s.titlePic != nil {
		_ = f.DrawPicture(s.titlePic)
	}
	_ = f.DrawLineAt(s.font, MainMenuLabel, 0, render.CmdRow)
}
