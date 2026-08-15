package audio

import (
	"encoding/binary"
	"sync"
)

// PC 喇叭的呈現層：把直譯器每個 tick 吐出的「8253 除數 ＋ 閘控」變成 PCM。
//
// 這一層是**重製決策**，原版沒有對應的程式碼——原版把除數寫進 8253，
// 喇叭的方波是硬體產生的。這裡照同一個除數自己算方波
// （`docs/re/44` §7：不要換算成浮點頻率再合成，除數的整數誤差是音色的一部分）。

// SampleRate 是輸出取樣率。
//
// 44100 除以 TickHz（≈126.6）不是整數，所以每個 tick 的樣本數用累加器分配，
// 不四捨五入——否則長音效會慢慢跑掉。
const SampleRate = 44100

// Amplitude 是方波的振幅（16-bit 全幅的四分之一）。
//
// **重製決策**：PC 喇叭是全幅方波，直接用 0x7FFF 會又吵又容易削波，
// 這裡壓到 1/4。原版沒有音量概念。
const Amplitude = 0x2000

// Synth 是一台可以餵給音訊裝置的無盡串流。
//
// Read 永遠不回 EOF：沒有音效在跑時輸出靜音。**這是刻意的**——
// 音訊裝置拿到 EOF 就會停掉整條串流，之後觸發的音效一個都不會響。
type Synth struct {
	mu      sync.Mutex
	p       *Player
	rate    int
	pending []int // 遊戲執行緒排進來的音效編號

	// 相位與 tick 的分配都是累加器，不四捨五入。
	phase     float64 // 0–1，方波的相位
	tickAcc   float64 // 還欠這個 tick 幾個樣本
	out       Output
	perTick   float64
	stopAfter bool
}

// NewSynth 建一台合成器。rate ≤ 0 時用 SampleRate。
func NewSynth(p *Player, rate int) *Synth {
	if rate <= 0 {
		rate = SampleRate
	}
	return &Synth{p: p, rate: rate, perTick: float64(rate) / TickHz}
}

// Trigger 從遊戲執行緒排一個音效。
//
// **不直接呼叫 `Player.Play`**：音訊裝置在另一個執行緒讀 Read，
// 而 Play 會改四個聲部的狀態。排進佇列、由 Read 那一側消化，
// 兩邊就只共用一把鎖。
func (s *Synth) Trigger(n int) {
	if n < 0 || n >= SFXCount {
		return
	}
	s.mu.Lock()
	s.pending = append(s.pending, n)
	s.mu.Unlock()
}

// Busy 回報還有沒有音效在響（含排隊中的）。
func (s *Synth) Busy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pending) > 0 || s.p.Busy()
}

// Read 產生 16-bit little-endian 立體聲樣本。
//
// b 的長度不是 4 的倍數時只填得下的部分——音訊裝置允許這樣，
// 下一次呼叫會從缺口接下去。
func (s *Synth) Read(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := 0
	for n+4 <= len(b) {
		if s.tickAcc < 1 {
			s.step()
			s.tickAcc += s.perTick
		}
		s.tickAcc--

		var v int16
		if s.out.Gate != 0 && s.out.Divisor != 0 {
			// 方波：相位前半正、後半負。
			step := PITHz / float64(s.out.Divisor) / float64(s.rate)
			s.phase += step
			for s.phase >= 1 {
				s.phase--
			}
			if s.phase < 0.5 {
				v = Amplitude
			} else {
				v = -Amplitude
			}
		} else {
			// 靜音時把相位歸零，下一個音才不會從隨機位置起跳（會有咔聲）。
			s.phase = 0
		}
		binary.LittleEndian.PutUint16(b[n:], uint16(v))
		binary.LittleEndian.PutUint16(b[n+2:], uint16(v))
		n += 4
	}
	return n, nil
}

// step 推進一個直譯器 tick，先消化排隊中的觸發。
func (s *Synth) step() {
	if len(s.pending) > 0 {
		// 一個 tick 只吃一個：原版的呼叫端之間至少隔一個中斷，
		// 同一 tick 連播兩首會讓前一首完全聽不到。
		s.p.Play(s.pending[0])
		s.pending = s.pending[1:]
	}
	s.out = s.p.Tick()
}
