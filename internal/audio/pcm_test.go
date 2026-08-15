package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

// 合成層的門檻：**沉默不是成功**。沒有音效在跑時 Read 要回靜音而不是 EOF，
// 觸發之後要真的有波形，而且波形的頻率要對得上 8253 的除數。
func TestSynthSilentWhenIdle(t *testing.T) {
	p := load(t)
	s := NewSynth(p, SampleRate)

	b := make([]byte, 4*1000)
	n, err := s.Read(b)
	if err != nil {
		t.Fatalf("閒置時 Read 不該回錯：%v", err)
	}
	if n != len(b) {
		t.Fatalf("Read 應該填滿緩衝區，得到 %d／%d", n, len(b))
	}
	for i := 0; i+1 < n; i += 2 {
		if v := int16(binary.LittleEndian.Uint16(b[i:])); v != 0 {
			t.Fatalf("閒置時第 %d 個樣本應該是 0，得到 %d", i/2, v)
		}
	}
	if s.Busy() {
		t.Error("沒觸發任何音效卻說在忙")
	}
}

// 音效 4 是唯一帶旋律與封套的一首（docs/re/44 §6），拿它當波形的正對照。
func TestSynthProducesWave(t *testing.T) {
	p := load(t)
	s := NewSynth(p, SampleRate)
	s.Trigger(4)

	// 半秒足夠讓它出聲（音效 4 有 225 個 tick ≈ 1.8 秒）。
	b := make([]byte, 4*SampleRate/2)
	if _, err := s.Read(b); err != nil {
		t.Fatal(err)
	}
	var nonZero, peak int
	for i := 0; i+1 < len(b); i += 4 {
		v := int(int16(binary.LittleEndian.Uint16(b[i:])))
		if v != 0 {
			nonZero++
		}
		if a := abs(v); a > peak {
			peak = a
		}
	}
	t.Logf("半秒裡 %d／%d 個樣本非零，峰值 %d", nonZero, len(b)/4, peak)
	if nonZero == 0 {
		t.Fatal("觸發了音效 4 卻一個樣本都沒有——合成層沒接上")
	}
	if peak != Amplitude {
		t.Errorf("方波峰值應該是 %d，得到 %d", Amplitude, peak)
	}
	// 左右聲道必須一致（PC 喇叭是單聲道）。
	for i := 0; i+3 < len(b); i += 4 {
		l := binary.LittleEndian.Uint16(b[i:])
		r := binary.LittleEndian.Uint16(b[i+2:])
		if l != r {
			t.Fatalf("第 %d 個樣本左右不一致：%d ≠ %d", i/4, int16(l), int16(r))
		}
	}
}

// 頻率要照除數算出來，不是照「聽起來差不多」。
//
// 用固定除數直接驅動合成器，數過零點——這繞開位元組碼，量的是方波本身。
func TestSynthFrequencyMatchesDivisor(t *testing.T) {
	p := load(t)
	s := NewSynth(p, SampleRate)
	const divisor = 2000 // 1193182 ÷ 2000 ≈ 596.6 Hz
	s.out = Output{Divisor: divisor, Gate: 1}
	s.tickAcc = 1 << 30 // 不讓 step() 把 out 蓋掉

	b := make([]byte, 4*SampleRate) // 一秒
	if _, err := s.Read(b); err != nil {
		t.Fatal(err)
	}
	edges := 0
	prev := int16(binary.LittleEndian.Uint16(b))
	for i := 4; i+1 < len(b); i += 4 {
		v := int16(binary.LittleEndian.Uint16(b[i:]))
		if (prev < 0) != (v < 0) {
			edges++
		}
		prev = v
	}
	got := float64(edges) / 2 // 一個週期兩次過零
	want := PITHz / divisor
	t.Logf("除數 %d：量到 %.1f Hz，算出來是 %.1f Hz", divisor, got, want)
	if math.Abs(got-want) > want*0.01 {
		t.Errorf("頻率差超過 1%%：量到 %.1f，應該是 %.1f", got, want)
	}
}

// 音效放完要自己安靜下來——除了 6 號（資料本身就是無限迴圈，docs/re/44 §6）。
func TestSynthStopsWhenDone(t *testing.T) {
	p := load(t)
	s := NewSynth(p, SampleRate)
	s.Trigger(1) // 5 個 tick 的點擊

	b := make([]byte, 4*SampleRate/2) // 半秒遠超過 5 tick
	if _, err := s.Read(b); err != nil {
		t.Fatal(err)
	}
	if s.Busy() {
		t.Error("音效 1 只有 5 個 tick，半秒之後不該還在忙")
	}
	// 尾端必須已經靜音。
	tail := b[len(b)-4*100:]
	for i := 0; i+1 < len(tail); i += 2 {
		if v := int16(binary.LittleEndian.Uint16(tail[i:])); v != 0 {
			t.Fatalf("尾端第 %d 個樣本應該靜音，得到 %d", i/2, v)
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
