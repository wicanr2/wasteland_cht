package audio

// 位元組碼直譯器（`sub_1CD52`，`docs/re/44` §4）。
//
// 一個 byte 的 opcode，四種形狀。**解到「設 +0x00」為止算一批**——
// 時值數完才會再進來解下一批，這就是「時值」與「指令」共用一個欄位的原因。

const (
	opSet    = 0xFF // FF <欄位> <u16>：設欄位
	opLoop   = 0xFE // FE <欄位> <u16 位址>：計數器迴圈
	opRebase = 0xFD // FD <u16 位址>：換寫入基底
)

// step 解一批位元組碼。對應 `sub_1CD52`。
//
// 「寫入基底」預設就是這個聲部自己；`0xFD` 與音符指令可以把它換成別的聲部——
// **一個序列可以指揮其他聲部**，那是原版做和弦與多軌的方式。
func (p *Player) step(self *Voice) {
	di := self.Pointer
	if di == 0 {
		return
	}
	target := self // 欄位設定寫到哪個聲部
	for guard := 0; guard < 4096; guard++ {
		op := p.Data.u8(int(di))
		di++
		switch {
		case op == opSet:
			f := p.Data.u8(int(di))
			val := p.Data.u16(int(di) + 1)
			di += 3
			target.Set(f, val)
			if f == fDuration {
				// 設時值 ＝ 這一批結束。時值 0 代表整個序列走完。
				if self.Duration == 0 {
					di = 0
				}
				self.Pointer = di
				return
			}
		case op == opLoop:
			f := p.Data.u8(int(di))
			dst := p.Data.u16(int(di) + 1)
			di += 3
			if c := target.Get(f); c == 0 {
				di = dst // 計數器為 0 ＝ 無條件跳
			} else {
				target.Set(f, c-1)
				if c-1 != 0 {
					di = dst
				}
			}
		case op == opRebase:
			dst := p.Data.u16(int(di))
			di += 2
			// 基底是聲部陣列裡的位移；不對齊聲部就當成「寫不到任何聲部」，
			// 用一個丟棄用的暫存體接住（原版會寫進 seg005 的別處）。
			if int(dst)%voiceFields == 0 && int(dst)/voiceFields < VoiceCount {
				target = &p.Voices[int(dst)/voiceFields]
			} else {
				target = &Voice{}
			}
		default: // 音符
			for {
				vi := int(op >> 5)
				idx := int(op & 0x1F)
				pitchByte := p.Data.u8(int(di))
				di++

				dst := &Voice{} // 聲部編號超出範圍就丟掉，不要 panic
				if vi < VoiceCount {
					dst = &p.Voices[vi]
				}
				scale := uint16(dst.LenScale & 0xFF)
				if scale == 0 {
					scale = 1
				}
				base, ok := p.Data.Length(idx)
				if !ok {
					p.OutOfRange++
				}
				dur := uint16(base) * scale
				dst.Duration = dur
				self.Duration = dur

				pitch := int(pitchByte&0x7F) + int(dst.Transpose)
				div := p.Data.Divisor(pitch)
				dst.Divisor = div
				dst.Output = div
				dst.EnvOffset = 0
				dst.EnvCount = 1
				if p.OnNote != nil {
					p.OnNote(vi, pitch, dur)
				}

				if pitchByte&0x80 == 0 {
					break
				}
				op = p.Data.u8(int(di))
				di++
			}
			if self.Duration == 0 {
				di = 0
			}
			self.Pointer = di
			return
		}
	}
	// 解了四千個指令還沒收——資料壞了或我們讀錯了，停掉這個聲部而不是卡死。
	self.Pointer = 0
	self.Duration = 0
}
