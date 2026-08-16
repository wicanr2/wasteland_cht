package ui

// 背景音樂（重製版自己加的）。
//
// **原版沒有 BGM**：DOS 版只有九首 PC 喇叭音效（`docs/re/44`）。所以這一層
// 與 `internal/audio` 那台方波引擎是兩件不同的東西——
// 音效走原版的位元組碼直譯器，音樂走這裡的 ogg。
//
// 曲子由 `tools/make_music.py` 譜、`tools/render_music.sh` 用 Roland MT-32
// 算成 ogg。**檔案不入版控**（波形裡有 MT-32 的 PCM 取樣），
// 載不到就整條關掉，遊戲照跑——與字型、翻譯目錄同一個原則。

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

// musicRate 是音樂的取樣率。**要與音效那條路一致**——Ebiten 一個行程只有
// 一個 audio.Context，兩邊取樣率不同的話後開的那個會失敗。
//
// 值取自 `internal/audio` 的 SampleRate，由 NewMusic 的呼叫端保證。
type musicTrack struct {
	name string
	data []byte
}

// Music 是背景音樂播放器。零值可用，代表「沒有音樂」。
type Music struct {
	ctx     *audio.Context
	tracks  map[string]musicTrack
	player  *audio.Player
	current string
	vol     float64
	on      bool
}

// LoadMusic 從目錄讀進所有 `*.ogg`。
//
// 目錄不存在或一首都沒有時回 nil——**不當成錯誤**：沒有音樂的遊戲仍然玩得動，
// 而把它變成啟動失敗會讓「沒跑 render_music.sh」的人連遊戲都開不起來。
func LoadMusic(ctx *audio.Context, dir string) (*Music, error) {
	if ctx == nil || dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	m := &Music{ctx: ctx, tracks: map[string]musicTrack{}, vol: 0.6, on: true}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".ogg") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("讀音樂 %s：%w", e.Name(), err)
		}
		name := strings.TrimSuffix(e.Name(), ".ogg")
		m.tracks[name] = musicTrack{name: name, data: raw}
	}
	if len(m.tracks) == 0 {
		return nil, nil
	}
	return m, nil
}

// Tracks 回報載進來的曲名（驗收工具用）。
func (m *Music) Tracks() []string {
	if m == nil {
		return nil
	}
	out := make([]string, 0, len(m.tracks))
	for k := range m.tracks {
		out = append(out, k)
	}
	return out
}

// Apply 讓音樂跟上場景這一幀的狀態：要播哪一首、開著沒、音量多少。
//
// 每幀叫一次。**只有變了才動**——每幀重開播放器會一直從頭播，
// 而症狀是「音樂一直卡在第一秒」。
func (m *Music) Apply(track string, on bool, vol int) {
	if m == nil {
		return
	}
	v := float64(vol) / 10
	if v < 0 {
		v = 0
	} else if v > 1 {
		v = 1
	}
	if v != m.vol {
		m.vol = v
		if m.player != nil {
			m.player.SetVolume(v)
		}
	}
	if on != m.on {
		m.on = on
		if !on {
			m.stop()
		}
	}
	if !m.on {
		return
	}
	if track == m.current && m.player != nil {
		return
	}
	m.play(track)
}

func (m *Music) stop() {
	if m.player != nil {
		_ = m.player.Close()
		m.player = nil
	}
	m.current = ""
}

func (m *Music) play(track string) {
	t, ok := m.tracks[track]
	if !ok {
		// 沒有這一首就維持現況（比突然靜音好）。曲目缺一首不該讓別的也停。
		return
	}
	m.stop()
	stream, err := vorbis.DecodeF32(bytes.NewReader(t.data))
	if err != nil {
		return
	}
	// 無縫循環：整首當成一個迴圈段（`audio.NewInfiniteLoop` 吃 byte 長度）。
	loop := audio.NewInfiniteLoopF32(stream, stream.Length())
	pl, err := m.ctx.NewPlayerF32(io.Reader(loop))
	if err != nil {
		return
	}
	pl.SetVolume(m.vol)
	pl.Play()
	m.player, m.current = pl, track
}

// Close 收掉播放器。
func (m *Music) Close() {
	if m == nil {
		return
	}
	m.stop()
}

// Musical 是「會告訴呈現層要播哪一首」的場景。沒實作就沒有背景音樂。
type Musical interface {
	// MusicTrack 是這一幀該播的曲名（空字串 ＝ 不播）。
	MusicTrack() string
	// MusicSetting 是玩家在設定裡選的開關與音量（0–10）。
	MusicSetting() (on bool, vol int)
}
