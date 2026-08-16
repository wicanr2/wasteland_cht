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
	"image"
	"image/png"
	"os"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/input"
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
	// 中文那三條路徑（`docs/spec/11` §7）。載不到就輸出英文畫面，不當成錯誤。
	langFile := flag.String("lang", "translations/zh-Hant.cat", "翻譯目錄")
	fontDir := flag.String("font", "workplace/eten", "倚天點陣字目錄（玩家自備）")
	refsFile := flag.String("refs", "docs/re/generated/paragraph-refs.tsv", "段落引用表")
	paraFile := flag.String("paragraphs", "translations/paragraphs-zh-Hant.cat", "段落正文")
	// keys 是要送進去的按鍵（每個字元一次 Update），用來截到選單／戰鬥／手札。
	keys := flag.String("keys", "", "play 模式：依序送這些按鍵")
	mapID := flag.Int("map", -1, "play 模式：先切到這張地圖（資源編號）")
	journal := flag.Int("journal", 0, "play 模式：打開手札停在這一頁（1 起算，0 ＝ 不開）")
	ending := flag.Bool("ending", false, "play 模式：直接進結局")
	endingTicks := flag.Int("ending-ticks", 130, "結局播到第幾個 tick 再截圖")
	// 功能鍵不是字元，`-keys` 送不出去（那一支收的是 ASCII）。
	fn := flag.String("fn", "", "play 模式：`-keys` 之後送的功能鍵，逗號分隔（help｜settings｜quit）")
	// ⚠ 原版的亂數靠鍵盤輪詢推進（`docs/re/13`），無頭預設不推是為了可重現。
	// 代價是序列退化成「初值全零的前幾項」——那時候走再多步也不會遇到敵人，
	// 而畫面看起來只是「運氣不好」。要截遭遇畫面就給 `-poll`（同 `wl-play`）。
	poll := flag.Int("poll", 0, "play 模式：每個按鍵之前推進亂數 N 次")
	titleScreen := flag.Bool("title", false, "play 模式：停在標題畫面（玩家開機看到的那一張）")
	cursor := flag.String("cursor", "", "play 模式：把滑鼠游標畫在 x,y（高解畫布的像素）")
	flag.Parse()

	opt := shotOptions{
		lang: *langFile, font: *fontDir, refs: *refsFile,
		paragraphs: *paraFile, keys: *keys, mapID: *mapID,
		journal: *journal, ending: *ending, endingTicks: *endingTicks,
		fn: *fn, poll: *poll, title: *titleScreen, cursor: *cursor,
	}
	if err := run(*romDir, *imagePath, *mode, *block, *pic, *out, *at, *hour, opt); err != nil {
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

// shotOptions 是中文路徑與按鍵序列。
type shotOptions struct {
	lang, font, refs, paragraphs string
	keys                         string
	mapID                        int
	journal                      int
	ending                       bool
	endingTicks                  int
	fn                           string
	poll                         int
	title                        bool
	cursor                       string
}

// sendFn 送一個面板類的功能鍵。
//
// 只收這三個：快速存讀檔會寫檔，截圖工具不該有副作用。
// ⚠ F10 走 `Action` 不走 `Fn`（`internal/input`）——離開是動作不是功能鍵。
func sendFn(scene *play.Scene, name string) error {
	in := input.Input{Dir: input.DirNone}
	switch name {
	case "help":
		in.Fn = input.FnHelp
	case "settings":
		in.Fn = input.FnSettings
	case "quit":
		in.Action = input.ActionQuit
	default:
		return fmt.Errorf("-fn 不認得 %q（help｜settings｜quit）", name)
	}
	_, err := scene.Update(in)
	return err
}

func run(romDir, imagePath, mode string, blockID, picID int,
	outPath, at string, hour int, opt shotOptions) error {
	rom, err := assets.Open(romDir)
	if err != nil {
		return err
	}
	if err := rom.LoadImage(imagePath); err != nil {
		return err
	}
	var img *image.RGBA
	var w, h int
	if mode == "play" {
		scene, err := play.New(rom)
		if err != nil {
			return err
		}
		// 中文三條路徑：載不到就是英文畫面，遊戲照跑（`docs/spec/11` §7）。
		if opt.lang != "" {
			_ = scene.LoadCatalogue(opt.lang)
		}
		hasFont := scene.LoadFont(opt.font) == nil
		_ = scene.LoadJournal(opt.refs, opt.paragraphs)
		if opt.mapID >= 0 {
			x, y := uint8(18), uint8(2)
			if at != "" {
				var ax, ay int
				if _, err := fmt.Sscanf(at, "%d,%d", &ax, &ay); err == nil {
					x, y = uint8(ax), uint8(ay)
				}
			}
			if err := scene.LoadMap(opt.mapID, x, y); err != nil {
				return err
			}
		} else if err := place(scene, at, hour); err != nil {
			return err
		}
		if opt.cursor != "" {
			var cx, cy int
			if _, err := fmt.Sscanf(opt.cursor, "%d,%d", &cx, &cy); err != nil {
				return fmt.Errorf("-cursor 要寫成 x,y：%w", err)
			}
			if cerr := scene.LoadCursors(); cerr != nil {
				return cerr
			}
			if _, err := scene.Update(input.Input{Dir: input.DirNone,
				Mouse: input.Mouse{X: cx, Y: cy}}); err != nil {
				return err
			}
		}
		if opt.title {
			scene.BeginTitle()
		}
		if opt.ending {
			scene.BeginEnding()
			for i := 0; i < opt.endingTicks; i++ {
				scene.TickEnding()
			}
		}
		if opt.journal > 0 {
			scene.OpenJournal(opt.journal)
		}
		for i := 0; i < len(opt.keys); i++ {
			for p := 0; p < opt.poll; p++ {
				scene.PollRNG()
			}
			// `IKJL` 是原版的方向鍵（`docs/re/72` §4），其餘當字元送。
			in := input.Input{Dir: input.DirNone, Char: opt.keys[i]}
			switch opt.keys[i] {
			case 'i', 'I':
				in = input.Input{Dir: input.DirUp}
			case 'k', 'K':
				in = input.Input{Dir: input.DirDown}
			case 'j', 'J':
				in = input.Input{Dir: input.DirLeft}
			case 'l', 'L':
				in = input.Input{Dir: input.DirRight}
			}
			if _, err := scene.Update(in); err != nil {
				return fmt.Errorf("送第 %d 個按鍵 %q：%w", i, opt.keys[i], err)
			}
		}
		for _, name := range strings.Split(opt.fn, ",") {
			if name == "" {
				continue
			}
			if err := sendFn(scene, name); err != nil {
				return err
			}
		}
		if hasFont {
			// 有字型就走 640 × 400 的中文畫面（`docs/spec/10`）。
			img = scene.HiFrame().ToImage()
			w, h = render.HiScreenWidth, render.HiScreenHeight
		} else {
			img = scene.Frame().ToImage()
			w, h = render.ScreenWidth, render.ScreenHeight
		}
	} else {
		scene, err := viewer.New(rom, viewer.Mode(mode), blockID, picID)
		if err != nil {
			return err
		}
		img = scene.Frame().ToImage()
		w, h = render.ScreenWidth, render.ScreenHeight
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return err
	}
	fmt.Printf("已寫出 %s（%d × %d）\n", outPath, w, h)
	return nil
}
