// 指令 wasteland 開視窗。
//
//	-mode play    從出廠存檔開始走地圖（規則層：能不能進、時鐘、遭遇、事件）
//	-mode map     資產檢視器：只搬視窗原點，沒有任何判定
//	-mode title   標題畫面
//	-mode pic     ALLPICS 的一張圖
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
	scale := flag.Int("scale", 3, "視窗放大倍率")
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
		err = ui.Run(scene, title, *scale, synth)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}
