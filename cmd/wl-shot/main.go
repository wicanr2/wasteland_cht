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
	"github.com/wicanr2/wasteland_cht/internal/play"
	"github.com/wicanr2/wasteland_cht/internal/render"
	"github.com/wicanr2/wasteland_cht/internal/viewer"
)

func main() {
	romDir := flag.String("rom", "workplace/orig/wastland", "原版資料目錄（玩家自備）")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	mode := flag.String("mode", "map", "play｜map｜title｜pic")
	block := flag.Int("block", 0, "MSQ 區塊編號（0–41）")
	pic := flag.Int("pic", 0, "ALLPICS 圖片編號")
	out := flag.String("out", "shot.png", "輸出 PNG")
	// play 模式的兩個對拍用開關：把隊伍擺到指定座標、把時鐘撥到指定的時。
	// 原版沒有這種捷徑，**只用來重現某個畫面**，不要拿它當遊戲功能。
	at := flag.String("at", "", "play 模式：把隊伍移到 x,y")
	hour := flag.Int("hour", -1, "play 模式：把時鐘的「時」設成這個值（0–23）")
	flag.Parse()

	if err := run(*romDir, *imagePath, *mode, *block, *pic, *out, *at, *hour); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}

// place 套用 -at／-hour。兩個都沒給就什麼都不做。
func place(scene *play.Scene, at string, hour int) error {
	w := scene.World()
	if at != "" {
		var x, y int
		if _, err := fmt.Sscanf(at, "%d,%d", &x, &y); err != nil {
			return fmt.Errorf("-at 要寫成 x,y：%w", err)
		}
		if x < 0 || y < 0 || x >= w.Block.Dim || y >= w.Block.Dim {
			return fmt.Errorf("座標 (%d, %d) 超出這張地圖的 %d × %d", x, y, w.Block.Dim, w.Block.Dim)
		}
		w.Teleport(uint8(x), uint8(y))
	}
	if hour >= 0 {
		if hour > 23 {
			return fmt.Errorf("-hour 要在 0–23，得到 %d", hour)
		}
		w.Clock.Hour = uint8(hour)
	}
	if at != "" || hour >= 0 {
		scene.Invalidate()
	}
	return nil
}

func run(romDir, imagePath, mode string, blockID, picID int, outPath, at string, hour int) error {
	rom, err := assets.Open(romDir)
	if err != nil {
		return err
	}
	if err := rom.LoadImage(imagePath); err != nil {
		return err
	}
	var frame *render.Frame
	if mode == "play" {
		scene, err := play.New(rom)
		if err != nil {
			return err
		}
		if err := place(scene, at, hour); err != nil {
			return err
		}
		frame = scene.Frame()
	} else {
		scene, err := viewer.New(rom, viewer.Mode(mode), blockID, picID)
		if err != nil {
			return err
		}
		frame = scene.Frame()
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, frame.ToImage()); err != nil {
		return err
	}
	fmt.Printf("已寫出 %s（%d × %d）\n", outPath, render.ScreenWidth, render.ScreenHeight)
	return nil
}
