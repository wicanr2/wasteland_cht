// Package audio 是原版的 PC 喇叭音效引擎（docs/spec/08、docs/re/44）。
//
// 原版是一台跑在計時器中斷上的位元組碼直譯器：四個聲部各自解一串位元組碼，
// 每個 tick 推進一步，再從四個裡面**挑一個**送進 8253 channel 2。
// PC 喇叭一次只發得出一個音，所以四個聲部是排隊不是混音。
//
// 這個套件只吐「8253 除數 ＋ 開關」的時間序列，**不做合成**——
// 方波怎麼變成取樣是呈現層的事。
package audio

// TickHz 是直譯器的更新率：1193182 ÷ 0x24E3（`docs/re/44` §2）。
const TickHz = 1193182.0 / 0x24E3

// PITHz 是 8253 的輸入頻率，把除數換算成頻率時用。
const PITHz = 1193182.0

// VoiceCount 是聲部數。
const VoiceCount = 4

// 聲部結構的欄位位移。位元組碼用**位移**定址，所以這張表就是它的介面
// （`docs/re/44` §3）。
const (
	fDuration   = 0x00 // 剩餘時值；0 ＝ 沒在發聲
	fPointer    = 0x02 // 位元組碼指標
	fDivisor    = 0x04 // 頻率除數
	fSlide      = 0x06 // 滑音增量
	fOutput     = 0x08 // 這一 tick 送進 8253 的除數
	fGate       = 0x0A // 喇叭閘控；0 ＝ 靜音
	fGateStep   = 0x0C // 閘控增量
	fLenScale   = 0x0E // 音長倍率
	fEnvSwitch  = 0x10 // 時值走到這裡就切封套第二段
	fTranspose  = 0x12 // 移調（半音）
	fEnvBase    = 0x14 // 封套資料基底
	fEnvOffset  = 0x16 // 封套目前位移
	fEnvCount   = 0x18 // 封套步進倒數
	fVibBase    = 0x1A // 顫音波表基底
	fVibPhase   = 0x1C // 顫音相位
	fVibStep    = 0x1E // 顫音相位增量
	fVibDepth   = 0x20 // 顫音深度
	fVibWrap    = 0x22 // 顫音相位環繞值
	fCounter    = 0x24 // 通用計數器，只有 0xFE 迴圈用
	voiceFields = 0x2E // 結構大小
)

// Voice 是一個聲部的狀態。用具名欄位而不是 [0x2E]byte，
// 位移的對照放在 get/set——**位移不在表裡的就寫進 other，不影響輸出**
// （原版 +0x26–+0x2D 沒有任何程式碼讀它）。
type Voice struct {
	Duration  uint16
	Pointer   uint16
	Divisor   uint16
	Slide     uint16
	Output    uint16
	Gate      uint16
	GateStep  uint16
	LenScale  uint16
	EnvSwitch uint16
	Transpose uint16
	EnvBase   uint16
	EnvOffset uint16
	EnvCount  uint16
	VibBase   uint16
	VibPhase  uint16
	VibStep   uint16
	VibDepth  uint16
	VibWrap   uint16
	Counter   uint16

	// other 收下那些沒有語意的位移，讓「設進去再讀回來」照樣成立。
	other map[uint8]uint16
}

func (v *Voice) ptr(off uint8) *uint16 {
	switch off {
	case fDuration:
		return &v.Duration
	case fPointer:
		return &v.Pointer
	case fDivisor:
		return &v.Divisor
	case fSlide:
		return &v.Slide
	case fOutput:
		return &v.Output
	case fGate:
		return &v.Gate
	case fGateStep:
		return &v.GateStep
	case fLenScale:
		return &v.LenScale
	case fEnvSwitch:
		return &v.EnvSwitch
	case fTranspose:
		return &v.Transpose
	case fEnvBase:
		return &v.EnvBase
	case fEnvOffset:
		return &v.EnvOffset
	case fEnvCount:
		return &v.EnvCount
	case fVibBase:
		return &v.VibBase
	case fVibPhase:
		return &v.VibPhase
	case fVibStep:
		return &v.VibStep
	case fVibDepth:
		return &v.VibDepth
	case fVibWrap:
		return &v.VibWrap
	case fCounter:
		return &v.Counter
	}
	return nil
}

// Get 讀某個位移。未解的位移回先前寫進去的值。
func (v *Voice) Get(off uint8) uint16 {
	if p := v.ptr(off); p != nil {
		return *p
	}
	return v.other[off]
}

// Set 寫某個位移。
func (v *Voice) Set(off uint8, val uint16) {
	if p := v.ptr(off); p != nil {
		*p = val
		return
	}
	if v.other == nil {
		v.other = map[uint8]uint16{}
	}
	v.other[off] = val
}

func (v *Voice) reset() { *v = Voice{} }

// Output 是一個 tick 的結果：要送進 8253 的除數與喇叭閘控。
// Gate 為 0 代表這個 tick 不出聲。
type Output struct {
	Divisor uint16
	Gate    uint16
}

// Hz 把除數換算成頻率。除數為 0 時回 0（原版會被 8253 當成 65536，
// 但那種狀態在九首資料裡不會出現）。
func (o Output) Hz() float64 {
	if o.Divisor == 0 {
		return 0
	}
	return PITHz / float64(o.Divisor)
}
