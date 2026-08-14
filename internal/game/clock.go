package game

// 遊戲時鐘（docs/spec/04 §4、docs/re/27）。
//
// 24 小時制，走一步推進的量由該地圖的記錄區標頭決定：
//
//	Frac|Minute（16-bit）＋= 標頭 +0x34/+0x35   ＝ 每步分鐘數 × 256
//	Tick             ＋= 標頭 +0x36
//	Total            ＋= 標頭 +0x36            ← 32-bit 累計，換地圖不歸零
type Clock struct {
	Frac   uint8  // ds:4658h，分的小數
	Minute uint8  // ds:4659h
	Hour   uint8  // ds:465Ah
	Tick   uint8  // ds:4722h，週期任務用；換地圖歸零
	Total  uint32 // ds:4650h–4652h ＋ ds:722Eh，不歸零
}

// 晝夜門檻（docs/re/27 §5）。
const (
	DawnHour = 6
	DuskHour = 18
)

// Advance 推進一步。stepTime 是標頭 +0x34/+0x35，stepTick 是 +0x36。
// 回傳 true 表示這一步跨過了 16 刻的邊界，呼叫者要跑週期性的角色處理。
func (c *Clock) Advance(stepTime uint16, stepTick uint8) (periodic bool) {
	// 原版把「分的小數」與「分」當成一個 16-bit 加法，進位自然帶上去。
	frac := uint16(c.Frac) | uint16(c.Minute)<<8
	frac += stepTime
	c.Frac = uint8(frac)
	c.Minute = uint8(frac >> 8)

	// ⚠ 原版只減一次 60，不是取餘數。每步最多 4 分鐘，所以夠用；
	// 寫成 %= 60 會在餵進異常大的 stepTime 時與原版分歧。
	if c.Minute >= 60 {
		c.Minute -= 60
		c.Hour++
		if c.Hour >= 24 {
			c.Hour = 0
		}
	}

	c.Tick += stepTick
	c.Total += uint32(stepTick)
	return c.Tick&0x0F == 0
}

// Night 回傳這個時刻是不是夜間（門檻 6 時與 18 時）。
func (c *Clock) Night() bool { return c.Hour < DawnHour || c.Hour >= DuskHour }

// EnterMap 是換地圖時的處理：刻歸零，其餘不動（sub_18350）。
func (c *Clock) EnterMap() { c.Tick = 0 }
