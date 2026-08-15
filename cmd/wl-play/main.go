// 指令 wl-play 用一串腳本驅動遊戲，**不開視窗**——驗「正常玩家路徑」用。
//
//	go run ./cmd/wl-play -script "up,up,left,esc" -trace
//
// 測試全綠不等於玩得通（CLAUDE.md §4）。單元測試各自驗一塊規則，
// 這一支把它們串起來走一遍：走地圖、遇敵、進設施、按選單、離開，
// 中間任何一步 panic 或卡住都會停下來並印出走到哪裡。
//
// 腳本語法（逗號分隔，大小寫不拘）：
//
//	up／down／left／right     方向
//	esc／enter／space／quit   動作
//	B／S／P／…               單一字元 ＝ 按那個鍵
//	x30                      把前一個動作重複 30 次
//	@57:39                   把隊伍搬到 (57, 39)（用冒號——逗號是腳本的分隔符）
//	hour=2                   把時鐘的「時」設成 2
//	shot=/tmp/a.png          當下截一張圖
//
// 這支不依賴 Ebiten，所以在沒有 X／Wayland 的容器裡也跑得動。
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/play"
)

func main() {
	romDir := flag.String("rom", "workplace/orig/wastland", "原版資料目錄（玩家自備）")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	script := flag.String("script", "", "動作腳本，逗號分隔")
	trace := flag.Bool("trace", false, "每一步印出狀態")
	stopOnMsg := flag.String("stop-on", "", "訊息含這段字就停下來並回非 0")
	flag.Parse()

	rom, err := assets.Open(*romDir)
	if err != nil {
		fail("開啟原版資料：%v", err)
	}
	if err := rom.LoadImage(*imagePath); err != nil {
		fail("載入分析映像：%v", err)
	}
	scene, err := play.New(rom)
	if err != nil {
		fail("建立場景：%v", err)
	}

	steps, err := parse(*script)
	if err != nil {
		fail("腳本：%v", err)
	}
	r := &runner{scene: scene, trace: *trace, stopOn: *stopOnMsg}
	for i, s := range steps {
		if err := r.do(s); err != nil {
			fail("第 %d 步（%s）：%v", i+1, s.raw, err)
		}
	}
	fmt.Printf("走完 %d 步：%s\n", len(steps), r.state())
	if r.tripped != "" {
		fail("訊息命中 -stop-on %q：%s", *stopOnMsg, r.tripped)
	}
}

type step struct {
	raw string
	in  input.Input
	// 非輸入的指令
	moveTo *[2]uint8
	hour   int
	shot   string
}

type runner struct {
	scene   *play.Scene
	trace   bool
	stopOn  string
	tripped string
	n       int
}

// state 是一行狀態摘要，出事時看它就知道走到哪裡。
func (r *runner) state() string {
	w := r.scene.World()
	mode := "地圖"
	switch {
	case r.scene.InFacility():
		mode = "設施"
	case r.scene.InCombat():
		mode = "戰鬥"
	}
	alive := 0
	for _, c := range w.Party.Members {
		if c.CON > 0 {
			alive++
		}
	}
	return fmt.Sprintf("(%d, %d) %02d:%02d %s 活著 %d/%d｜%s",
		w.Party.X, w.Party.Y, w.Clock.Hour, w.Clock.Minute, mode,
		alive, len(w.Party.Members), r.scene.Message())
}

func (r *runner) do(s step) error {
	switch {
	case s.moveTo != nil:
		w := r.scene.World()
		if int(s.moveTo[0]) >= w.Block.Dim || int(s.moveTo[1]) >= w.Block.Dim {
			return fmt.Errorf("座標超出這張地圖的 %d × %d", w.Block.Dim, w.Block.Dim)
		}
		w.Teleport(s.moveTo[0], s.moveTo[1])
		r.scene.Invalidate()
	case s.hour >= 0:
		w := r.scene.World()
		w.Clock.Hour = uint8(s.hour)
		r.scene.Invalidate()
	case s.shot != "":
		return writePNG(r.scene, s.shot)
	default:
		if _, err := r.scene.Update(s.in); err != nil {
			return err
		}
	}
	r.n++
	if r.trace {
		fmt.Printf("%4d %-8s %s\n", r.n, s.raw, r.state())
	}
	if r.stopOn != "" && r.tripped == "" &&
		strings.Contains(r.scene.Message(), r.stopOn) {
		r.tripped = r.state()
	}
	return nil
}

func writePNG(scene *play.Scene, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, scene.Frame().ToImage())
}

// parse 把腳本字串拆成一連串動作，`xN` 就地展開。
func parse(s string) ([]step, error) {
	var out []step
	for _, tok := range strings.Split(s, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if n, ok := repeat(tok); ok {
			if len(out) == 0 {
				return nil, fmt.Errorf("%q 前面沒有可重複的動作", tok)
			}
			last := out[len(out)-1]
			for i := 1; i < n; i++ {
				out = append(out, last)
			}
			continue
		}
		st, err := one(tok)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, nil
}

func repeat(tok string) (int, bool) {
	if len(tok) < 2 || (tok[0] != 'x' && tok[0] != 'X') {
		return 0, false
	}
	n, err := strconv.Atoi(tok[1:])
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

func one(tok string) (step, error) {
	st := step{raw: tok, hour: -1}
	switch strings.ToLower(tok) {
	case "up":
		st.in.Dir = input.DirUp
	case "down":
		st.in.Dir = input.DirDown
	case "left":
		st.in.Dir = input.DirLeft
	case "right":
		st.in.Dir = input.DirRight
	case "esc":
		st.in.Action = input.ActionCancel
	case "enter", "space":
		st.in.Action = input.ActionConfirm
	case "quit":
		st.in.Action = input.ActionQuit
	default:
		switch {
		case strings.HasPrefix(tok, "@"):
			var x, y int
			if _, err := fmt.Sscanf(tok[1:], "%d:%d", &x, &y); err != nil {
				return st, fmt.Errorf("%q 不是 @x:y（逗號是腳本的分隔符，座標用冒號）", tok)
			}
			st.moveTo = &[2]uint8{uint8(x), uint8(y)}
		case strings.HasPrefix(strings.ToLower(tok), "hour="):
			h, err := strconv.Atoi(tok[5:])
			if err != nil || h < 0 || h > 23 {
				return st, fmt.Errorf("%q 的時不在 0–23", tok)
			}
			st.hour = h
		case strings.HasPrefix(strings.ToLower(tok), "shot="):
			st.shot = tok[5:]
		case len(tok) == 1:
			st.in.Char = tok[0]
		default:
			return st, fmt.Errorf("看不懂 %q", tok)
		}
	}
	return st, nil
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "wl-play: "+format+"\n", args...)
	os.Exit(1)
}
