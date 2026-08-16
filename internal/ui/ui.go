// Package ui 是 Ebiten 那一層（docs/spec/03 §2.8）。
//
// 職責只有三件事，**不做任何遊戲判斷、也不合成畫面**：
//
//  1. 把 render.Frame 的索引畫面上色送上螢幕
//  2. 把 Ebiten 的按鍵翻成 internal/input 的按鍵
//  3. 視窗縮放（內部永遠 320 × 200）
//
// ⚠ **這個套件在沒有 DISPLAY 的環境會在 package init 就 panic**（Ebiten 的行為）。
// 所以無頭用得到的東西一律不要放這裡——按鍵對應在 `internal/input`、
// 畫面合成在 `internal/render`，兩者都測得到。
package ui

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	wlaudio "github.com/wicanr2/wasteland_cht/internal/audio"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// keyOf 是唯一與 Ebiten 綁在一起的對照。
func keyOf(k ebiten.Key) input.Key {
	switch k {
	case ebiten.KeyArrowUp:
		return input.KeyUp
	case ebiten.KeyArrowDown:
		return input.KeyDown
	case ebiten.KeyArrowLeft:
		return input.KeyLeft
	case ebiten.KeyArrowRight:
		return input.KeyRight
	case ebiten.KeyW:
		return input.KeyW
	case ebiten.KeyA:
		return input.KeyA
	case ebiten.KeyS:
		return input.KeyS
	case ebiten.KeyD:
		return input.KeyD
	case ebiten.KeyEscape:
		return input.KeyEscape
	case ebiten.KeyF1:
		return input.KeyF1
	case ebiten.KeyF2:
		return input.KeyF2
	case ebiten.KeyF5:
		return input.KeyF5
	case ebiten.KeyF9:
		return input.KeyF9
	case ebiten.KeyF10:
		return input.KeyF10
	case ebiten.KeyEnter, ebiten.KeyNumpadEnter:
		return input.KeyEnter
	case ebiten.KeySpace:
		return input.KeySpace
	}
	return input.KeyNone
}

// Scene 是呈現層向上層要的東西：一幀畫面，外加收到的輸入。
// 用介面隔開，internal/ui 就不必認識任何遊戲或檢視器的型別。
type Scene interface {
	// Update 收到這一幀的輸入；回傳 false 表示要離開。
	Update(in input.Input) (bool, error)
	// Frame 回傳要顯示的畫面。
	Frame() *render.Frame
}

// Poller 是有亂數產生器要跟著輪詢推進的場景（規格 02 §1.1）。
type Poller interface{ PollRNG() }

// HiScene 是會輸出 640 × 400 高解析畫面的場景（中文用，`docs/spec/10`）。
//
// 中文點陣字是 16 × 15，塞不進原版 8 × 8 的字元格——所以 CJK 那條路
// 把畫面放大兩倍再疊字。場景實作了這個介面，整個視窗就走 640 × 400。
type HiScene interface{ HiFrame() *render.HiFrame }

// Sounder 是會發出 PC 喇叭音效的場景。
//
// TakeSound 回這一幀要播的音效編號（0–8，`docs/re/44` §6），−1 表示沒有；
// **取走就清掉**，同一個觸發不會播兩次。沒實作也照跑。
type Sounder interface{ TakeSound() int }

// Animator 是會隨時間變化的場景（設施圖的局部動畫，規格 26）。
// 沒實作也照跑——檢視器場景就沒有。
type Animator interface{ TickAnim() bool }



// animTicksPerFrame 是幾幀推一拍動畫。60 TPS ÷ 3 ＝ 20 Hz，
// 是整數分頻裡最接近原版 18.2 Hz 的一檔。
const animTicksPerFrame = 3

// Game 是 ebiten.Game 的實作。
type Game struct {
	scene Scene
	img   *ebiten.Image
	synth *wlaudio.Synth
	// hi 非 nil 就走 640 × 400 的中文畫面。
	hi HiScene
	keys  []ebiten.Key
	runes []rune
	buf   []input.Key
	// animTick 是動畫的分頻計數（見 animTicksPerFrame）。
	animTick int
	// music 是背景音樂（重製版自己加的，nil ＝ 沒有）。
	music *Music
}

// SetMusic 掛上背景音樂播放器。nil 就沒有音樂，遊戲照跑。
func (g *Game) SetMusic(m *Music) { g.music = m }

// New 建立一個 Game。**不開音訊**——測試與檢視器用這一支。
func New(scene Scene) *Game {
	g := &Game{scene: scene}
	if h, ok := scene.(HiScene); ok {
		g.hi = h
		g.img = ebiten.NewImage(render.HiScreenWidth, render.HiScreenHeight)
	} else {
		g.img = ebiten.NewImage(render.ScreenWidth, render.ScreenHeight)
	}
	return g
}

// NewWithAudio 建立一個會發聲的 Game。
//
// synth 為 nil 就跟 New 一樣。**音訊裝置在另一個執行緒讀 synth**，
// 所以場景那一側只准透過 Trigger 排隊（`internal/audio/pcm.go`）。
func NewWithAudio(scene Scene, synth *wlaudio.Synth) (*Game, error) {
	g := New(scene)
	if synth == nil {
		return g, nil
	}
	pl, err := audioContext().NewPlayer(synth)
	if err != nil {
		return nil, fmt.Errorf("開音訊：%w", err)
	}
	// 緩衝區太大時觸發到出聲會有明顯延遲；50 ms 在 60 fps 下是三幀。
	pl.SetBufferSize(50 * time.Millisecond)
	pl.Play()
	g.synth = synth
	return g, nil
}

// Update 是 Ebiten 的每幀更新。
func (g *Game) Update() error {
	g.keys = inpututil.AppendJustPressedKeys(g.keys[:0])
	g.runes = ebiten.AppendInputChars(g.runes[:0])
	g.buf = g.buf[:0]
	for _, k := range g.keys {
		if mapped := keyOf(k); mapped != input.KeyNone {
			g.buf = append(g.buf, mapped)
		}
	}
	// 每幀推進一次亂數產生器 ＝ 原版的鍵盤輪詢（規格 02 §1.1）。
	// **不叫它每一局的遭遇序列都會一樣**——產生器沒有種子。
	if p, ok := g.scene.(Poller); ok {
		p.PollRNG()
	}

	// 設施圖的動畫一拍 ≈ 55 ms，Ebiten 是 60 TPS——三幀一拍 ＝ 20 Hz，
	// 是整數分頻裡最接近的（規格 26 §5，實際頻率還沒與實機錄影對過）。
	g.animTick++
	if g.animTick >= animTicksPerFrame {
		g.animTick = 0
		if a, ok := g.scene.(Animator); ok {
			a.TickAnim()
		}
	}
	// 背景音樂跟著場景走：換模式換曲、設定裡關掉就停。
	if m, ok := g.scene.(Musical); ok && g.music != nil {
		on, vol := m.MusicSetting()
		g.music.Apply(m.MusicTrack(), on, vol)
	}

	// 音效在輸入之後、Update 之前取：這一幀觸發的音效由下一幀送出，
	// 與原版「呼叫端設好狀態、計時器中斷再播」的時序一致。
	if s, ok := g.scene.(Sounder); ok && g.synth != nil {
		if n := s.TakeSound(); n >= 0 {
			g.synth.Trigger(n)
		}
	}
	keep, err := g.scene.Update(input.Read(g.buf, g.runes))
	if err != nil {
		return err
	}
	if !keep {
		return ebiten.Termination
	}
	return nil
}

// Draw 把索引畫面上色送上螢幕。
func (g *Game) Draw(screen *ebiten.Image) {
	if g.hi != nil {
		h := g.hi.HiFrame()
		if h == nil {
			return
		}
		g.img.WritePixels(h.ToImage().Pix)
		screen.DrawImage(g.img, nil)
		return
	}
	frame := g.scene.Frame()
	if frame == nil {
		return
	}
	g.img.WritePixels(frame.RGBA())
	screen.DrawImage(g.img, nil)
}

// Layout 回傳內部解析度：中文畫面 640 × 400，其餘 320 × 200。
// 放大由 Ebiten 處理，內部座標永遠是原版的（或原版的兩倍）。
func (g *Game) Layout(int, int) (int, int) {
	if g.hi != nil {
		return render.HiScreenWidth, render.HiScreenHeight
	}
	return render.ScreenWidth, render.ScreenHeight
}

// audioContext 回傳這個行程唯一的 audio.Context。
//
// ⚠ **一個行程只能有一個**，而且取樣率固定在 `wlaudio.SampleRate`——
// 音效與背景音樂共用它。第二次用不同取樣率建會失敗。
func audioContext() *audio.Context {
	if ctx := audio.CurrentContext(); ctx != nil {
		return ctx
	}
	return audio.NewContext(wlaudio.SampleRate)
}

// Run 開視窗跑起來。**無頭環境不能呼叫。**
//
// synth 非 nil 時一併開音訊；場景要實作 Sounder 才會有聲音。
// musicDir 非空時從那裡讀 `*.ogg` 當背景音樂，場景要實作 Musical；
// **讀不到就沒有音樂，不是錯誤**（見 LoadMusic）。
func Run(scene Scene, title string, scale int, synth *wlaudio.Synth, musicDir string) error {
	if scale < 1 {
		return fmt.Errorf("縮放倍率要 ≥ 1，收到 %d", scale)
	}
	g, err := NewWithAudio(scene, synth)
	if err != nil {
		return err
	}
	if musicDir != "" {
		m, merr := LoadMusic(audioContext(), musicDir)
		if merr != nil {
			return merr
		}
		g.SetMusic(m)
		defer m.Close()
	}
	w, h := g.Layout(0, 0)
	ebiten.SetWindowSize(w*scale, h*scale)
	ebiten.SetWindowTitle(title)
	return ebiten.RunGame(g)
}
