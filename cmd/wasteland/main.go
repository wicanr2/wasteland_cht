// 指令 wasteland 開視窗跑資產檢視器。
//
// ⚠ 現在**還不是遊戲**：規則層的規格還沒 READY（docs/spec/00-index.md），
// 所以這支只把畫面畫出來、方向鍵搬視窗，沒有任何遊戲判定。
//
//	go run ./cmd/wasteland -rom workplace/orig/wastland \
//	    -image workplace/analysis/unpacked/wl.merged.exe -mode map -block 0
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/ui"
	"github.com/wicanr2/wasteland_cht/internal/viewer"
)

func main() {
	romDir := flag.String("rom", "workplace/orig/wastland", "原版資料目錄（玩家自備）")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	mode := flag.String("mode", "map", "map｜title｜pic")
	block := flag.Int("block", 0, "MSQ 區塊編號（0–41）")
	pic := flag.Int("pic", 0, "ALLPICS 圖片編號")
	scale := flag.Int("scale", 3, "視窗放大倍率")
	flag.Parse()

	rom, err := assets.Open(*romDir)
	if err == nil {
		err = rom.LoadImage(*imagePath)
	}
	var scene *viewer.Viewer
	if err == nil {
		scene, err = viewer.New(rom, viewer.Mode(*mode), *block, *pic)
	}
	if err == nil {
		err = ui.Run(scene, "Wasteland 資產檢視器（尚未是遊戲）", *scale)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}
