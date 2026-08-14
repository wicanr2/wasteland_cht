// 指令 wl-shot 把一幀畫面寫成 PNG，**不開視窗**——與 DOSBox 的原版截圖對拍用。
//
//	go run ./cmd/wl-shot -rom workplace/orig/wastland \
//	    -image workplace/analysis/unpacked/wl.merged.exe -mode title -out /tmp/title.png
//
// 這支不依賴 Ebiten，所以在沒有 X／Wayland 的容器裡也跑得動。
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/render"
	"github.com/wicanr2/wasteland_cht/internal/viewer"
)

func main() {
	romDir := flag.String("rom", "workplace/orig/wastland", "原版資料目錄（玩家自備）")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	mode := flag.String("mode", "map", "map｜title｜pic")
	block := flag.Int("block", 0, "MSQ 區塊編號（0–41）")
	pic := flag.Int("pic", 0, "ALLPICS 圖片編號")
	out := flag.String("out", "shot.png", "輸出 PNG")
	flag.Parse()

	if err := run(*romDir, *imagePath, *mode, *block, *pic, *out); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}

func run(romDir, imagePath, mode string, blockID, picID int, outPath string) error {
	rom, err := assets.Open(romDir)
	if err != nil {
		return err
	}
	if err := rom.LoadImage(imagePath); err != nil {
		return err
	}
	scene, err := viewer.New(rom, viewer.Mode(mode), blockID, picID)
	if err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, scene.Frame().ToImage()); err != nil {
		return err
	}
	fmt.Printf("已寫出 %s（%d × %d）\n", outPath, render.ScreenWidth, render.ScreenHeight)
	return nil
}
