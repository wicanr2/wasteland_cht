// 指令 wasteland 開視窗。
//
//	-mode play    從出廠存檔開始走地圖（規則層：能不能進、時鐘、遭遇、事件）
//	-mode map     資產檢視器：只搬視窗原點，沒有任何判定
//	-mode title   標題畫面
//	-mode pic     ALLPICS 的一張圖
//
// 遊戲中的功能鍵：F1 說明、F2 設定（音樂開關與音量）、F5／F9 快速存讀檔、
// F10 離開（先問再存再退）。ESC 一律只取消，任何一層都不會結束遊戲。
//
//	go run ./cmd/wasteland -rom workplace/orig/wastland \
//	    -image workplace/analysis/unpacked/wl.merged.exe -mode play
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/audio"
	"github.com/wicanr2/wasteland_cht/internal/play"
	"github.com/wicanr2/wasteland_cht/internal/ui"
	"github.com/wicanr2/wasteland_cht/internal/viewer"
)

func main() {
	romDir := flag.String("rom", "workplace/orig/wastland", "原版資料目錄（玩家自備）")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	mode := flag.String("mode", "play", "play｜map｜title｜pic")
	skipTitle := flag.Bool("skip-title", false, "跳過標題畫面直接進地圖（測試與截圖用）")
	langFile := flag.String("lang", "translations/zh-Hant.cat", "翻譯目錄（空字串 ＝ 英文）")
	fontDir := flag.String("font", "workplace/eten", "倚天點陣字目錄（玩家自備）")
	refsFile := flag.String("refs", "docs/re/generated/paragraph-refs.tsv", "段落引用表")
	paraFile := flag.String("paragraphs", "translations/paragraphs-zh-Hant.cat", "段落正文")
	block := flag.Int("block", 0, "MSQ 區塊編號（0–41）")
	pic := flag.Int("pic", 0, "ALLPICS 圖片編號")
	// ⚠ 中文畫面本身已經是原版的 3 倍（960 × 600），所以這裡預設 2 就好——
	// 再乘 3 會開出 2880 × 1800 的視窗。檢視器模式是 320 × 200，要 3 才夠看。
	scale := flag.Int("scale", 2, "視窗放大倍率（play 是 960×600 的畫布，檢視器是 320×200）")
	// 指令列的 `Save` 要寫回哪裡。原版是就地寫回自己的 `GAME1`／`GAME2`，
	// 但**預設不指到 `-rom`**：那份是驗過 SHA-256 的原版，寫過就開不起來
	// （`assets.Open`），而且專案規定不覆蓋原版資料。要真的存檔就先複製
	// 一份資料目錄出來、把 `-save-dir` 指到副本。空字串 ＝ 不寫檔，
	// `Save` 會照實說它沒寫出去。
	saveDir := flag.String("save-dir", "", "存檔寫回的**可寫**資料目錄副本（空 ＝ 不寫檔）")
	// F5／F9／F10 的快速存檔。**與 `-save-dir` 是兩條不同的路**：
	// 那條是原版的 `Save`（就地改寫 GAME1／GAME2），這條寫自己的容器格式，
	// 一個 byte 都不碰原版資料，所以預設就能用。
	quickSave := flag.String("quicksave", play.DefaultQuickSavePath(), "快速存檔檔案（空 ＝ 關掉 F5／F9）")
	// 背景音樂目錄。**原版沒有 BGM**，這是重製版加的；曲子要自己跑
	// `tools/render_music.sh` 算出來（ogg 不入版控）。載不到就靜靜沒有音樂。
	musicDir := flag.String("music", "workplace/music", "背景音樂目錄（*.ogg，空 ＝ 不播）")
	flag.Parse()

	rom, err := assets.Open(*romDir)
	if err == nil {
		err = rom.LoadImage(*imagePath)
	}
	var scene ui.Scene
	var synth *audio.Synth
	title := "Wasteland 資產檢視器"
	if err == nil {
		if *mode == "play" {
			var s *play.Scene
			s, err = play.New(rom)
			scene, title = s, "Wasteland（荒野遊俠）"
			if err == nil {
				s.SetSaveDir(*saveDir)
				s.SetQuickSavePath(*quickSave)
				if *saveDir == "" {
					fmt.Fprintln(os.Stderr,
						"提示：沒給 -save-dir，指令列的 Save 只會更新記憶體、不寫檔")
				}
			}
			// 原版開機是「標題畫面 → Start → 地圖」（`docs/re/95`）。
			// **沒有新遊戲／讀檔**：存檔就是 GAME1／GAME2 本身。
			if err == nil && !*skipTitle {
				s.BeginTitle()
			}
			// 中文化的三條路徑：翻譯目錄、倚天點陣字、段落手札。
			// **三個都載不到也照跑**，只是顯示英文、翻不開手札
			// （`docs/spec/11` §7：半成品的中文化要能玩）。
			if err == nil {
				if *langFile != "" {
					if lerr := s.LoadCatalogue(*langFile); lerr != nil {
						fmt.Fprintln(os.Stderr, "提示：翻譯目錄載不到，顯示英文 —", lerr)
					}
				}
				if ferr := s.LoadFont(*fontDir); ferr != nil {
					fmt.Fprintln(os.Stderr, "提示：倚天字型載不到，中文不顯示 —", ferr)
				}
				if jerr := s.LoadJournal(*refsFile, *paraFile); jerr != nil {
					fmt.Fprintln(os.Stderr, "提示：段落手札載不到 —", jerr)
				}
				// 滑鼠游標（原版的 `CURS`）。載不到就沒有游標圖，
				// 滑鼠照樣能點（`docs/spec/29` §6）。
				if cerr := s.LoadCursors(); cerr != nil {
					fmt.Fprintln(os.Stderr, "提示：滑鼠游標載不到 —", cerr)
				}
			}
			// 音效資料在執行檔的 seg005 裡（docs/re/44），拿不到就靜音跑，
			// **不要讓沒有聲音變成開不起來**。
			if err == nil {
				if data, aerr := rom.AudioData(); aerr == nil {
					if p, perr := audio.New(data); perr == nil {
						synth = audio.NewSynth(p, audio.SampleRate)
					}
				}
			}
		} else {
			scene, err = viewer.New(rom, viewer.Mode(*mode), *block, *pic)
		}
	}
	if err == nil {
		// 檢視器模式沒有 Musical，傳了也不會播；只有 play 需要。
		music := ""
		if *mode == "play" {
			music = *musicDir
		}
		err = ui.Run(scene, title, *scale, synth, music)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}
