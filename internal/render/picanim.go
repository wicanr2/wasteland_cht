package render

import "github.com/wicanr2/wasteland_cht/internal/assets"

// 規格 26：圖片的局部動畫。
//
// 每一格是一串 XOR 上去的像素，**疊上去之後不還原**——畫面永遠是
// 「底圖 ⊕ 已播過的所有格」。這能成立是因為一輪播完所有格會互相抵消
// （`docs/re/23` §5.3，82 張圖全部驗過）。

// PicPlayer 依拍數推進一張圖的動畫。
type PicPlayer struct {
	anim  assets.PicAnim
	count []int // 每個通道還要等幾拍
	next  []int // 每個通道下一個要播的格在腳本裡的位置
}

// NewPicPlayer 起一個播放器，初始倒數照腳本第一個 byte。
func NewPicPlayer(a assets.PicAnim) *PicPlayer {
	p := &PicPlayer{
		anim:  a,
		count: make([]int, len(a.Channels)),
		next:  make([]int, len(a.Channels)),
	}
	for i, ch := range a.Channels {
		if len(ch.Delay) > 0 {
			p.count[i] = int(ch.Delay[0])
		}
	}
	return p
}

// Tick 推進一拍，回傳這一拍要疊的元素（沒有就是空的）。
func (p *PicPlayer) Tick() []assets.AnimElem {
	var out []assets.AnimElem
	for i, ch := range p.anim.Channels {
		if len(ch.Frame) == 0 {
			continue
		}
		if p.count[i] > 0 {
			p.count[i]--
			continue
		}
		fi := int(ch.Frame[p.next[i]])
		if fi < len(p.anim.Frames) {
			out = append(out, p.anim.Frames[fi]...)
		}
		// 播完換下一個延遲；腳本走到底就回開頭（原版的 0xFF）。
		if d := p.next[i] + 1; d < len(ch.Delay) {
			p.count[i] = int(ch.Delay[d])
		} else if len(ch.Delay) > 0 {
			p.count[i] = int(ch.Delay[0])
		}
		p.next[i] = (p.next[i] + 1) % len(ch.Frame)
	}
	return out
}

// ApplyAnim 把一批元素 XOR 進畫面，(originX, originY) 是圖片左上角。
//
// 元素的座標是**圖片內座標**——原版把 y ＋ 8 烘在列位址表裡，
// 重製版由這裡加一次就好（規格 26 §2）。
func (f *Frame) ApplyAnim(elems []assets.AnimElem, originX, originY int) {
	for _, e := range elems {
		y := originY + e.Y
		for i, v := range e.Pixels {
			x := originX + e.X + i
			if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
				continue
			}
			f.Pix[y*ScreenWidth+x] ^= v & 0x0F
		}
	}
}

// ApplyAnimMask 把一張累積的動畫遮罩 XOR 進畫面。
//
// 動畫是累積的（疊上去不還原），所以重畫一幀時不能只補「這一拍的元素」——
// 要把**至今播過的全部**再疊一次。呼叫端維護那張遮罩，這裡只負責套用。
func (f *Frame) ApplyAnimMask(mask []byte, width, originX, originY int) {
	if width <= 0 {
		return
	}
	for i, v := range mask {
		if v == 0 {
			continue
		}
		x, y := originX+i%width, originY+i/width
		if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
			continue
		}
		f.Pix[y*ScreenWidth+x] ^= v & 0x0F
	}
}

// XorInto 把一批元素 XOR 進一張 width 寬的遮罩（用來累積播過的格）。
func XorInto(mask []byte, width int, elems []assets.AnimElem) {
	for _, e := range elems {
		for i, v := range e.Pixels {
			x := e.X + i
			if x < 0 || x >= width {
				continue
			}
			if k := e.Y*width + x; k >= 0 && k < len(mask) {
				mask[k] ^= v & 0x0F
			}
		}
	}
}
