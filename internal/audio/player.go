package audio

import (
	"encoding/binary"
	"fmt"
)

// seg005 的表（`docs/re/44` §5）。
const (
	pitchTable = 0x0174 // 12 × u16，最低八度的除數
	sfxTable   = 0x0192 // 9 × 4 × u16
	lenTable   = 0x01DB // 音長表
	lenUsable  = 0x01EF - lenTable // 表後面就是全域變數，只有這麼多筆在表內
	SFXCount   = 9
)

// Data 是原版 seg005 的內容（864 bytes）。玩家自備原版執行檔，
// 我們不散布它（`CLAUDE.md` §7）。
type Data []byte

func (d Data) u16(at int) uint16 {
	if at < 0 || at+2 > len(d) {
		return 0
	}
	return binary.LittleEndian.Uint16(d[at:])
}

func (d Data) u8(at int) uint8 {
	if at < 0 || at >= len(d) {
		return 0
	}
	return d[at]
}

// Divisor 回某個音高的 8253 除數：音高 ÷ 12 是八度、餘數查半音表、右移八度數
// （`docs/re/44` §4.1）。
func (d Data) Divisor(pitch int) uint16 {
	if pitch < 0 {
		pitch = 0
	}
	oct, semi := pitch/12, pitch%12
	base := d.u16(pitchTable + semi*2)
	if oct > 15 {
		return 0
	}
	return base >> uint(oct)
}

// Length 回某個音長索引的基礎時值。
//
// ⚠ 索引可以到 31，但只有 0–19 落在表內（`docs/spec/08` §5.1）。
// 超出的回 0（該聲部靜音）並回報 false——**不夾住成 19**，
// 夾住會讓一首本來會停的音效變成一直響。
func (d Data) Length(idx int) (uint8, bool) {
	if idx < 0 || idx >= lenUsable {
		return 0, false
	}
	return d.u8(lenTable + idx), true
}

// Player 是整台音效引擎：四個聲部 ＋ 一份資料。
type Player struct {
	Data   Data
	Voices [VoiceCount]Voice

	// OutOfRange 記下用到表外音長索引的次數（`docs/spec/08` §5.1）。
	OutOfRange int

	// OnNote 在位元組碼發出一個音符時呼叫（可為 nil）。
	//
	// 這是唯一看得到「同一個音連續彈四次」的地方——只看每個 tick 的輸出，
	// 四個一樣的音會塌成一個，重試與延長分不出來。
	OnNote func(voice, pitch int, duration uint16)
}

// New 建一台引擎。
func New(d Data) (*Player, error) {
	if len(d) < lenTable+lenUsable {
		return nil, fmt.Errorf("seg005 太短：%d bytes", len(d))
	}
	return &Player{Data: d}, nil
}

// SFXStart 回某個音效四個聲部的位元組碼起點；0 代表那個聲部不參與。
func (p *Player) SFXStart(n int) ([VoiceCount]uint16, bool) {
	var out [VoiceCount]uint16
	if n < 0 || n >= SFXCount {
		return out, false
	}
	for i := range out {
		out[i] = p.Data.u16(sfxTable + n*8 + i*2)
	}
	return out, true
}

// Play 觸發一個音效，照 `sub_1CBD3`：起點非 0 的聲部**整個清掉**，
// 寫進指標並把時值設成 1（下一個 tick 就開始解）。起點為 0 的聲部不動。
func (p *Player) Play(n int) bool {
	starts, ok := p.SFXStart(n)
	if !ok {
		return false
	}
	for i, at := range starts {
		if at == 0 {
			continue
		}
		p.Voices[i].reset()
		p.Voices[i].Pointer = at
		p.Voices[i].Duration = 1
	}
	return true
}

// Busy 回報還有沒有聲部在跑。
func (p *Player) Busy() bool {
	for i := range p.Voices {
		if p.Voices[i].Duration != 0 {
			return true
		}
	}
	return false
}

// Tick 推進一個 tick，回傳這一 tick 要送到喇叭的東西。
//
// 順序照 `sub_1CC76`：四個聲部各推進一步，**再**掃一次挑出贏家。
// 仲裁是「第一個時值非 0 且閘控非 0 的聲部」——編號小的優先，
// 四個都不符合就用最後一個（原版 `sub si, 2Eh` 那條路）。
func (p *Player) Tick() Output {
	for i := range p.Voices {
		p.advance(&p.Voices[i])
	}
	pick := VoiceCount - 1
	for i := range p.Voices {
		if p.Voices[i].Duration != 0 && p.Voices[i].Gate != 0 {
			pick = i
			break
		}
	}
	v := &p.Voices[pick]
	return Output{Divisor: v.Output, Gate: v.Gate}
}

// advance 是 `sub_1CCC9`：一個聲部的一個 tick。步驟順序不能換。
func (p *Player) advance(v *Voice) {
	if v.Duration == 0 {
		return
	}
	v.Gate += v.GateStep
	v.Divisor += v.Slide

	var vib uint16
	if phase := v.VibPhase + v.VibStep; phase != 0 {
		if v.VibWrap != 0 && phase >= v.VibWrap {
			phase -= v.VibWrap
		}
		v.VibPhase = phase
		// 波表的 byte 當成高位元組（原版 `mov ah,[di]; mov al,0`）。
		sample := uint16(p.Data.u8(int(int16(phase)>>4)+int(v.VibBase))) << 8
		vib = sample * v.VibDepth
	}
	v.Output = vib + v.Divisor

	if v.Duration == v.EnvSwitch {
		v.EnvOffset = 0x10
		v.EnvCount = 1
	}

	v.Duration--
	if v.Duration == 0 {
		p.step(v)
	}

	if v.EnvCount == 0 {
		return
	}
	v.EnvCount--
	if v.EnvCount != 0 {
		return
	}
	// 封套：每段四個 byte。後兩個 ＝ 0xFFFF 代表「直接設閘控」，繼續讀下一段。
	for {
		at := int(v.EnvBase) + int(v.EnvOffset)
		lo, hi := p.Data.u16(at), p.Data.u16(at+2)
		v.EnvOffset += 4
		if hi != 0xFFFF {
			v.GateStep = lo
			v.EnvCount = hi
			return
		}
		v.Gate = lo
		if int(v.EnvOffset) > len(p.Data) {
			return
		}
	}
}
