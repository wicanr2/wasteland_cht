package audio

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 音效資料在原版執行檔裡，玩家自備。沒有就跳過——與其他吃原版素材的測試一致。
func load(t *testing.T) *Player {
	t.Helper()
	img := filepath.Join("..", "..", "workplace", "analysis", "unpacked", "wl.merged.exe")
	if _, err := os.Stat(img); err != nil {
		t.Skip("沒有分析映像，跳過")
	}
	var r assets.Rom
	if err := r.LoadImage(img); err != nil {
		t.Fatalf("載入映像：%v", err)
	}
	data, err := r.AudioData()
	if err != nil {
		t.Fatalf("取音效資料：%v", err)
	}
	p, err := New(data)
	if err != nil {
		t.Fatalf("建引擎：%v", err)
	}
	return p
}

// 驗收 2：音高表**照原樣用**。十一個半音準到 0.12 音分內，
// 但 E 差了 −2.91 音分——那是原始資料的錯字（`docs/re/44` §5），
// 這個測試同時擋住「把 E 修正掉」與「改用公式算頻率」。
func TestPitchTableIsUsedVerbatim(t *testing.T) {
	p := load(t)
	const c1 = 32.7032 // 音高 0
	for n := 0; n < 12; n++ {
		ideal := c1 * math.Pow(2, float64(n)/12)
		got := PITHz / float64(p.Data.Divisor(n))
		cents := 1200 * math.Log2(got/ideal)

		// 其餘十一個實測最大 +0.12 音分；門檻放 0.13 免得卡在等號上。
		want := 0.13
		if n == 4 {
			// E 是走音的。用區間而不是等號——這裡要擋的是「被修好」，
			// 不是把浮點誤差寫死。
			if cents > -2.8 || cents < -3.0 {
				t.Fatalf("E 的偏差是 %+.2f 音分，應該在 −2.91 附近"+
					"（被修正掉了？原版就是這個錯字，照抄）", cents)
			}
			continue
		}
		if math.Abs(cents) > want {
			t.Fatalf("半音 %d 偏差 %+.2f 音分，超過 %.2f", n, cents, want)
		}
	}
	// 順帶擋住「乾脆用公式」：真的用公式的話 E 會變準。
	if p.Data.Divisor(4) != 0x714F {
		t.Fatalf("E 的除數是 %#x，原版是 0x714F", p.Data.Divisor(4))
	}
}

// 驗收 1：八首自己會停，長度寫死當回歸護欄；音效 6 是唯一停不下來的那一首。
//
// ⚠ 靜態反組譯**回答不了這一題**——`0xFE` 走哪一條路要看執行期的計數器，
// `tools/summarize_sfx.py` 兩條路都印得出來，但決定不了實際走哪條。
// 「一首會不會停」只有實跑算數。
func TestSFXLengths(t *testing.T) {
	want := map[int]int{0: 1, 1: 5, 2: 3, 3: 71, 4: 225, 5: 121, 7: 109, 8: 161}
	for n := 0; n < SFXCount; n++ {
		if n == 6 {
			continue // 見 TestSFX6NeverEnds
		}
		p := load(t)
		if !p.Play(n) {
			t.Fatalf("音效 %d 觸發失敗", n)
		}
		const maxTicks = 100000
		ticks := 0
		for p.Busy() && ticks < maxTicks {
			p.Tick()
			ticks++
		}
		if p.Busy() {
			t.Fatalf("音效 %d 跑了 %d 個 tick 還沒停", n, ticks)
		}
		if ticks != want[n] {
			t.Fatalf("音效 %d 跑了 %d 個 tick，應該是 %d", n, ticks, want[n])
		}
	}
}

// 音效 6 停不下來，而且是**資料本身就這樣設計**，不是直譯器跑掉。
//
// `0x01FA` 起兩段各 180 次的滑音擺盪，跑完之後 `0x0232` 的計數器已經是 0，
// `0xFE` 於是變成無條件跳，跳回 `0x01FA` 重來。要停只能 `Play(0)`。
// 這個測試要抓到「跳回頭」這件事，光看「還在跑」分不出設計與失控。
func TestSFX6NeverEnds(t *testing.T) {
	p := load(t)
	if !p.Play(6) {
		t.Fatal("音效 6 觸發失敗")
	}
	// 第一段的批次停在 0x020E／0x0216，第二段停在 0x0226／0x022E。
	const phase2 = 0x0226
	var reachedPhase2, restarted bool
	for i := 0; i < 2000 && !restarted; i++ {
		p.Tick()
		switch ptr := p.Voices[0].Pointer; {
		case ptr >= phase2:
			reachedPhase2 = true
		case reachedPhase2 && ptr != 0 && ptr < phase2:
			restarted = true
		}
	}
	if !reachedPhase2 {
		t.Fatal("音效 6 沒走到第二段")
	}
	if !restarted {
		t.Fatal("音效 6 沒跳回開頭——這首應該是無限循環的")
	}
	if !p.Busy() {
		t.Fatal("音效 6 停了，它不該停")
	}
	// 靠 Play(0) 停，這是原版唯一的關法。
	p.Play(0)
	for i := 0; i < 10; i++ {
		p.Tick()
	}
	if p.Busy() {
		t.Fatal("Play(0) 之後音效 6 還在跑")
	}
}

// 驗收 3：音效 4 是唯一一首旋律，音序要對得上，
// 而且與 docs/re/31 的「升級播音效 4」是獨立的兩個來源。
func TestSFX4IsTheLevelUpFanfare(t *testing.T) {
	p := load(t)
	var pitches []int
	p.OnNote = func(voice, pitch int, _ uint16) {
		if voice == 0 {
			pitches = append(pitches, pitch)
		}
	}
	if !p.Play(4) {
		t.Fatal("音效 4 觸發失敗")
	}
	for i := 0; i < 5000 && p.Busy(); i++ {
		p.Tick()
	}
	// F3 F3 F3 F3 C3 F3 C4（音高 0 ＝ C1，所以 41 ＝ 第 3 個八度的 F）。
	want := []int{41, 41, 41, 41, 36, 41, 48}
	if len(pitches) != len(want) {
		t.Fatalf("音序長度 %d（%v），應該是 %d（%v）", len(pitches), pitches, len(want), want)
	}
	for i := range want {
		if pitches[i] != want[i] {
			t.Fatalf("第 %d 個音是 %d，應該是 %d（整串：%v）", i, pitches[i], want[i], pitches)
		}
	}
}

// 驗收 4：兩個聲部同時發聲時，輸出的是編號小的那一個。
func TestArbitrationPrefersLowerVoice(t *testing.T) {
	p := &Player{Data: make(Data, lenTable+lenUsable)}
	// ⚠ 設 Output 沒有用——advance 每個 tick 都會用 Divisor 重算它。
	p.Voices[1] = Voice{Duration: 10, Gate: 3, Divisor: 0x1111}
	p.Voices[2] = Voice{Duration: 10, Gate: 3, Divisor: 0x2222}
	out := p.Tick()
	if out.Divisor != 0x1111 {
		t.Fatalf("挑到除數 %#x，應該是聲部 1 的 %#x", out.Divisor, 0x1111)
	}
	// 聲部 1 靜音之後換聲部 2 上場。
	p.Voices[1].Gate = 0
	if out := p.Tick(); out.Divisor != 0x2222 {
		t.Fatalf("聲部 1 靜音後挑到 %#x，應該是聲部 2 的 %#x", out.Divisor, 0x2222)
	}
}

// 驗收 5：音長索引 ≥ 20 回時值 0，**不夾住成 19**。
func TestLengthIndexOutOfTableIsNotClamped(t *testing.T) {
	p := load(t)
	// ⚠ 不能拿「值等於表尾」當夾住的判準——表尾（索引 19）本來就是 0。
	// 回報 ok == false 才是「沒被夾住」的證據：夾住的實作會回 true。
	if _, ok := p.Data.Length(lenUsable - 1); !ok {
		t.Fatalf("索引 %d 應該在表內", lenUsable-1)
	}
	for _, idx := range []int{lenUsable, 25, 31} {
		v, ok := p.Data.Length(idx)
		if ok {
			t.Fatalf("索引 %d 不該被當成表內（夾住的實作會回 true）", idx)
		}
		if v != 0 {
			t.Fatalf("索引 %d 回 %d，應該是 0", idx, v)
		}
	}
}

// 驗收 6：沒有原版執行檔時整個音效關掉，不得 panic。
func TestNilPlayerIsSafe(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("沒有資料時 New 應該回錯誤")
	}
	var p *Player
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil Player 不該 panic：%v", r)
		}
	}()
	if p != nil && p.Busy() {
		t.Fatal("不會走到這裡")
	}
}

// 音效 0 是「全部靜音」：四個聲部都指同一段，各自把閘控清成 0 並結束。
func TestSFX0SilencesEveryVoice(t *testing.T) {
	p := load(t)
	p.Play(6) // 先弄出聲音
	for i := 0; i < 50; i++ {
		p.Tick()
	}
	p.Play(0)
	for i := 0; i < 10; i++ {
		p.Tick()
	}
	if p.Busy() {
		t.Fatal("音效 0 之後不該還有聲部在跑")
	}
	if out := p.Tick(); out.Gate != 0 {
		t.Fatalf("音效 0 之後閘控是 %#x，應該是 0", out.Gate)
	}
}
