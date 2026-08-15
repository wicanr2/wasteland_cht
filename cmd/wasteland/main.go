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
