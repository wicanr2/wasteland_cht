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
//	map=19:32:32             換到地圖 19 並站到 (32, 32)
//	fight                    直接開一場遭遇（不等擲骰），驗戰鬥流程用
//	path=29:43               自動尋路走到 (29, 43)，中途照常觸發事件
//	shot=/tmp/a.png          當下截一張圖
//
// 幾個會改變結論的旗標：
//
//	-poll N        每一步之前推進亂數 N 次。**預設 0 ＝ 序列退化**
//	               （產生器初值全零、熵來自鍵盤輪詢，`docs/re/13`），
//	               那時候「走 N 步沒遇到敵人」證明不了任何事。
//	-save-dir DIR  `S` 指令寫回哪個**可寫的**資料目錄；空的就不寫檔。
//	-modified      那個目錄已經被寫過，跳過 SHA-256 驗證（存檔重開要用）。
//	-lang／-font   中文那兩條路徑，預設接上；trace 會把中文印在〔〕裡。
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
	"github.com/wicanr2/wasteland_cht/internal/lang"
	"github.com/wicanr2/wasteland_cht/internal/play"
)

func main() {
	romDir := flag.String("rom", "workplace/orig/wastland", "原版資料目錄（玩家自備）")
	imagePath := flag.String("image", "workplace/analysis/unpacked/wl.merged.exe", "解包合成映像")
	script := flag.String("script", "", "動作腳本，逗號分隔")
	trace := flag.Bool("trace", false, "每一步印出狀態")
	stopOnMsg := flag.String("stop-on", "", "訊息含這段字就停下來並回非 0")
	emitKeys := flag.Bool("emit-keys", false, "把走過的方向印成 tools/dosbox.sh 的 timeline")
	// 中文那三條路徑與 wl-shot 同一組預設。載不到就跑英文，不是錯誤
	// （`docs/spec/11` §7）——`-lang ""` 可以明確關掉。
	langFile := flag.String("lang", "translations/zh-Hant.cat", "翻譯目錄")
	fontDir := flag.String("font", "workplace/eten", "倚天點陣字目錄（玩家自備）")
	refsFile := flag.String("refs", "docs/re/generated/paragraph-refs.tsv", "段落引用表")
	paraFile := flag.String("paragraphs", "translations/paragraphs-zh-Hant.cat", "段落正文")
	// ⚠ **熵**：原版的亂數靠鍵盤輪詢推進（`docs/re/13`），呈現層每幀叫一次
	// `PollRNG`。無頭預設不叫是為了可重現，代價是序列退化成
	// 「初值全零的前幾項」——那時候「走 49 步沒遇到敵人」證明不了任何事。
	// 要抽樣遭遇率就給 `-poll N`，模擬玩家按一次鍵之間的輪詢次數。
	poll := flag.Int("poll", 0, "每一步之前推進亂數 N 次（模擬玩家按鍵的輪詢）")
	// 存檔驗收：把原版資料複製一份出來，`-rom` 與 `-save-dir` 都指到副本，
	// 第二次開就要加 `-modified`（寫過之後 SHA-256 當然對不上）。
	saveDir := flag.String("save-dir", "", "`S` 指令寫回的**可寫**資料目錄（空 ＝ 不寫檔）")
	modified := flag.Bool("modified", false, "資料目錄已經被寫過，跳過 SHA-256 驗證")
	flag.Parse()

	open := assets.Open
	if *modified {
		open = assets.OpenModified
	}
	rom, err := open(*romDir)
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
	scene.SetSaveDir(*saveDir)
	if *langFile != "" {
		_ = scene.LoadCatalogue(*langFile)
	}
	hasFont := *fontDir != "" && scene.LoadFont(*fontDir) == nil
	if *refsFile != "" {
		_ = scene.LoadJournal(*refsFile, *paraFile)
	}

	steps, err := parse(*script)
	if err != nil {
		fail("腳本：%v", err)
	}
	r := &runner{scene: scene, trace: *trace, stopOn: *stopOnMsg, emit: *emitKeys,
		hi: hasFont, poll: *poll}
	for i, s := range steps {
		if err := r.do(s); err != nil {
			fail("第 %d 步（%s）：%v", i+1, s.raw, err)
		}
	}
	if *emitKeys {
		fmt.Printf("timeline: %s\n", strings.Join(r.keys, ";"))
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
	loadTo *[3]int // map=id:x:y
	fight  bool
	pathTo *[2]int // path=x:y
}

type runner struct {
	scene   *play.Scene
	trace   bool
	stopOn  string
	tripped string
	n       int
	emit    bool
	hi      bool     // 有字型：截圖走 640 × 400 的中文畫面
	poll    int      // 每一步前推進亂數的次數
	keys    []string // -emit-keys：走過的方向，送得進 tools/dosbox.sh
}

// tick 模擬玩家按下一個鍵之前，主迴圈輪詢了幾次鍵盤（熵的唯一來源）。
func (r *runner) tick() {
	for i := 0; i < r.poll; i++ {
		r.scene.PollRNG()
	}
}

// state 是一行狀態摘要，出事時看它就知道走到哪裡。
//
// 模式取自 `Scene.Mode()`——訊息列看不出「進了設施但選單沒開」，模式看得出來。
func (r *runner) state() string {
	w := r.scene.World()
	alive := 0
	for _, c := range w.Party.Members {
		if c.CON > 0 {
			alive++
		}
	}
	return fmt.Sprintf("地圖 %-3d (%2d,%2d) %02d:%02d %-16s 活 %d/%d｜%s%s%s",
		r.scene.MapID(), w.Party.X, w.Party.Y, w.Clock.Hour, w.Clock.Minute,
		r.scene.Mode(), alive, len(w.Party.Members), r.scene.Message(), cjkOf(r.scene),
		facilityLines(r.scene))
}

// facilityLines 是設施畫面印在地點名底下那幾行（選單與清單）。
// 它們**不走訊息列**，所以只看 Message() 會以為設施是一片空白。
func facilityLines(s *play.Scene) string {
	if !s.InFacility() {
		return ""
	}
	return " ⟨" + strings.Join(s.Facility().Lines, " ／ ") + "⟩"
}

// cjkOf 把這一步的中文訊息解回 UTF-8 印在終端機上。
// 解不出來就照 byte 印十六進位——**不要靜靜吞掉**，那正是要驗的東西。
func cjkOf(s *play.Scene) string {
	b := s.CJK()
	if len(b) == 0 {
		return ""
	}
	if txt, ok := lang.FromBig5(b); ok {
		return "〔" + txt + "〕"
	}
	return fmt.Sprintf("〔Big5 解不開：% x〕", b)
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
	case s.loadTo != nil:
		if err := r.scene.LoadMap(s.loadTo[0], uint8(s.loadTo[1]), uint8(s.loadTo[2])); err != nil {
			return err
		}
	case s.pathTo != nil:
		return r.walkTo(s.pathTo[0], s.pathTo[1])
	case s.fight:
		c, err := r.scene.StartEncounter()
		if err != nil {
			return err
		}
		if c == nil {
			return fmt.Errorf("這一格附近沒有打得起來的遭遇（視窗裡沒有敵人格，或距離過不了門檻）")
		}
	case s.shot != "":
		return writePNG(r.scene, s.shot, r.hi)
	default:
		r.tick()
		if _, err := r.scene.Update(s.in); err != nil {
			return err
		}
		r.record(s.in.Dir)
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

// writePNG 截一張圖。有字型就走 640 × 400 的中文畫面——
// 低解那張畫不出中文，拿它驗中文化會每次都「看起來沒翻」（`docs/spec/10`）。
func writePNG(scene *play.Scene, path string, hi bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if hi {
		return png.Encode(f, scene.HiFrame().ToImage())
	}
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
	// ⚠ Dir 一定要明確設 DirNone：零值是 DirUp（internal/input）。
	st := step{raw: tok, hour: -1, in: input.Input{Dir: input.DirNone}}
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
	case "fight":
		st.fight = true
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
		case strings.HasPrefix(strings.ToLower(tok), "path="):
			var x, y int
			if _, err := fmt.Sscanf(tok[5:], "%d:%d", &x, &y); err != nil {
				return st, fmt.Errorf("%q 不是 path=x:y", tok)
			}
			st.pathTo = &[2]int{x, y}
		case strings.HasPrefix(strings.ToLower(tok), "map="):
			var id, x, y int
			if _, err := fmt.Sscanf(tok[4:], "%d:%d:%d", &id, &x, &y); err != nil {
				return st, fmt.Errorf("%q 不是 map=id:x:y", tok)
			}
			st.loadTo = &[3]int{id, x, y}
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

// walkTo 用 BFS 找一條路走過去，每一步都走 Scene.Update（事件照常觸發）。
//
// ⚠ **尋路只看地形**：中途踩到傳送格會換地圖，那時就停下來——
// 這不是失敗，是遊戲本來就會發生的事。
func (r *runner) walkTo(tx, ty int) error {
	w := r.scene.World()
	startMap := r.scene.MapID()
	for step := 0; step < 4096; step++ {
		if int(w.Party.X) == tx && int(w.Party.Y) == ty {
			return nil
		}
		// 停在「進新地點？」上就答 Yes——尋路的目的就是走過去。
		if r.scene.Asking() {
			if _, err := r.scene.Update(input.Input{Dir: input.DirNone, Char: 'Y'}); err != nil {
				return err
			}
			r.keysAppend("y")
			r.n++
			if r.trace {
				fmt.Printf("%4d %-8s %s\n", r.n, "yes", r.state())
			}
			if r.scene.MapID() != startMap {
				return nil
			}
			continue
		}
		dir, ok := r.nextStep(tx, ty)
		if !ok {
			return fmt.Errorf("從 (%d, %d) 找不到往 (%d, %d) 的路", w.Party.X, w.Party.Y, tx, ty)
		}
		r.tick()
		if _, err := r.scene.Update(input.Input{Dir: dir}); err != nil {
			return err
		}
		r.record(dir)
		r.n++
		if r.trace {
			fmt.Printf("%4d %-8s %s\n", r.n, "path", r.state())
		}
		if r.scene.MapID() != startMap {
			return nil // 換地圖了，路徑到此為止
		}
	}
	return fmt.Errorf("走了 4096 步還沒到 (%d, %d)", tx, ty)
}

// nextStep 回報從目前位置往目標走的下一步（BFS 的第一步）。
func (r *runner) nextStep(tx, ty int) (input.Direction, bool) {
	w := r.scene.World()
	dim := w.Block.Dim
	type node struct{ x, y int }
	start := node{int(w.Party.X), int(w.Party.Y)}
	prev := map[node]node{start: start}
	queue := []node{start}
	dirs := []struct {
		d  input.Direction
		dx int
		dy int
	}{{input.DirUp, 0, -1}, {input.DirDown, 0, 1}, {input.DirLeft, -1, 0}, {input.DirRight, 1, 0}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.x == tx && cur.y == ty {
			// 從終點回溯到起點的第一步
			for prev[cur] != start {
				cur = prev[cur]
			}
			for _, d := range dirs {
				if start.x+d.dx == cur.x && start.y+d.dy == cur.y {
					return d.d, true
				}
			}
			return input.DirNone, false
		}
		for _, d := range dirs {
			nx, ny := cur.x+d.dx, cur.y+d.dy
			if nx < 0 || ny < 0 || nx >= dim || ny >= dim {
				continue
			}
			n := node{nx, ny}
			if _, seen := prev[n]; seen || !w.Passable(nx, ny) {
				continue
			}
			// **繞開沿途的傳送格**：踩上去會換地圖，路就斷了。
			// 終點自己是傳送格時當然要留著。
			if !(nx == tx && ny == ty) {
				if terr, _, _, err := w.Block.At(nx, ny); err == nil && terr == 10 {
					continue
				}
			}
			prev[n] = cur
			queue = append(queue, n)
		}
	}
	return input.DirNone, false
}

// keysAppend 記一個字元鍵（給 -emit-keys 用）。
func (r *runner) keysAppend(s string) {
	if r.emit {
		r.keys = append(r.keys, "type:"+s, "wait:1")
	}
}

// record 把一步的方向記成 tools/dosbox.sh 的 timeline 片段。
func (r *runner) record(d input.Direction) {
	if !r.emit {
		return
	}
	name := map[input.Direction]string{
		input.DirUp: "Up", input.DirDown: "Down",
		input.DirLeft: "Left", input.DirRight: "Right",
	}[d]
	if name == "" {
		return
	}
	r.keys = append(r.keys, "key:"+name, "wait:1")
}
