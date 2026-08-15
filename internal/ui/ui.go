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

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	keys  []ebiten.Key
	runes []rune
	buf   []input.Key
	// animTick 是動畫的分頻計數（見 animTicksPerFrame）。
	animTick int
}

// New 建立一個 Game。
func New(scene Scene) *Game {
	return &Game{
		scene: scene,
		img:   ebiten.NewImage(render.ScreenWidth, render.ScreenHeight),
	}
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
	// 設施圖的動畫一拍 ≈ 55 ms，Ebiten 是 60 TPS——三幀一拍 ＝ 20 Hz，
	// 是整數分頻裡最接近的（規格 26 §5，實際頻率還沒與實機錄影對過）。
	g.animTick++
	if g.animTick >= animTicksPerFrame {
		g.animTick = 0
		if a, ok := g.scene.(Animator); ok {
			a.TickAnim()
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
	frame := g.scene.Frame()
	if frame == nil {
		return
	}
	g.img.WritePixels(frame.RGBA())
	screen.DrawImage(g.img, nil)
}

// Layout 固定回傳 320 × 200——放大由 Ebiten 處理，內部座標永遠是原版的。
func (g *Game) Layout(int, int) (int, int) {
	return render.ScreenWidth, render.ScreenHeight
}

// Run 開視窗跑起來。**無頭環境不能呼叫。**
func Run(scene Scene, title string, scale int) error {
	if scale < 1 {
		return fmt.Errorf("縮放倍率要 ≥ 1，收到 %d", scale)
	}
	ebiten.SetWindowSize(render.ScreenWidth*scale, render.ScreenHeight*scale)
	ebiten.SetWindowTitle(title)
	return ebiten.RunGame(New(scene))
}
