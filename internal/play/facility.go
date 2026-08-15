package play

// 設施場景：踩上去到走出來（docs/spec/23、docs/re/29 §5.4）。
//
// 規格 09 是規則（價格、治療、學習），規格 18 是買賣與治療的迴圈，
// 這個檔只做場景：載圖、印地點名、離開時切回地圖。

import (
	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

// facilityPicture 是每個設施進場要載的 ALLPICS 圖（docs/re/29 §5.4 那張表）。
//
// ⚠ 第五種（FacilityEnding）**沒有圖**——原版那一支（0x1B4F0）連
// `sub_190A6` 都沒叫，因為它根本不是設施畫面：走進去就播結局
// （`docs/re/96`）。用 -1 表示「沒有」，不要猜一個編號。
var facilityPicture = [game.FacilityCount]int{
	game.FacilityDoctor:  0,
	game.FacilityShop:    1,
	game.FacilityTrainer: 2,
	game.FacilityRoster:  3,
	game.FacilityEnding: -1,
}

// FacilityScene 是一個設施畫面的狀態。
type FacilityScene struct {
	Facility game.Facility
	Picture  int // ALLPICS 的圖片編號；-1 ＝ 這種設施沒有圖
	Lines    []string

	// state 是選單的狀態（docs/spec/25）。離開設施就整個丟掉——
	// 原版也沒有跨場次保留的東西。
	state *shopState
	note  string // 這一步要多印的一行（「背包滿了」那類）

	// Skills 是訓練師教得了的技能。**這家店教什麼由呼叫端給**——
	// 原版的清單走 `sub_1BDFF`，那一支與商店的清單框架同源、還沒逆向
	// （docs/re/52 §4），所以不在這裡假裝算得出來。
	Skills []TrainableSkill
}

// TrainableSkill 是訓練師教得了的一個技能。
type TrainableSkill struct {
	ID   byte
	Data game.SkillData
}

// EnterFacility 從地圖進設施。
//
// 回 nil 表示這一格不是設施（記錄 `+0x00` 的 bit7 沒設 → 那是腳本指令，
// 走規格 07 的腳本直譯器，不是這裡）。
func (s *Scene) EnterFacility(record []byte) *FacilityScene {
	f, ok := game.ParseFacility(record)
	if !ok {
		return nil
	}
	// 選單狀態一開始就建好：`Frame()` 與 `Key()` 都會讀它，
	// 只在 Key 裡懶初始化的話，先畫再按就會踩到 nil。
	// 第 4 種不是設施，是**結局**（`ds:A4E0h[4]` → `0x1B4F0`）。
	if f.Kind == game.FacilityEnding {
		s.BeginEnding()
		return nil
	}
	fs := &FacilityScene{
		Facility: f,
		Picture:  facilityPicture[f.Kind],
		state:    &shopState{Stock: map[byte]byte{}},
	}
	if f.Name != "" {
		fs.Lines = append(fs.Lines, f.Name)
	}
	if f.Kind == game.FacilityTrainer {
		fs.Skills = s.trainableSkills()
	}
	s.facility = fs
	// 動畫從頭起：遮罩清空、播放器重建（規格 26 §3）。
	s.animMask = make([]byte, render.FacilityPicWidth*render.FacilityPicHeight)
	s.player = nil
	if fs.Picture >= 0 && fs.Picture < len(s.anims) {
		s.player = render.NewPicPlayer(s.anims[fs.Picture])
	}
	s.dirty = true
	return fs
}

// trainableSkills 是訓練師列出來的技能。
//
// **不是每家店各有一份清單**——訓練師記錄裡只有 kind、下一步、招呼字串與
// 名稱（`docs/re/79` §2）。原版 `0x1BC30` 用清單框架列完整張技能資料表，
// 篩選（IQ 需求、費用、角色技能欄還有沒有空位）全都發生在**選完之後**：
// `0x1BC60` 選好才 `sub_1CA8D` 算費用、跟角色記錄 `+0x20` 的技能點數比。
//
// 推論等級：**強證據**（清單框架與選後檢查都直讀，「清單沒有篩選」是
// 從「篩選在選之後」推出來的）。
func (s *Scene) trainableSkills() []TrainableSkill {
	if s.rom == nil {
		return nil // 沒有映像的場景（單元測試用）就不列
	}
	raw, err := s.rom.SkillTableRaw()
	if err != nil {
		return nil
	}
	out := make([]TrainableSkill, 0, len(raw)/2)
	for i := 0; i+1 < len(raw); i += 2 {
		d := game.ParseSkillData(raw[i], raw[i+1])
		if d.BaseCost == 0 {
			continue // 費用 0 的槽不是可學的技能
		}
		out = append(out, TrainableSkill{ID: byte(i / 2), Data: d})
	}
	return out
}

// LeaveFacility 回地圖。
//
// 座標、時鐘、地圖都不用還原——踩到設施格之前隊伍就已經走完那一步了
// （規格 07 §6），設施只是接在後面跑。
// enterRosterIfNeeded 讓設施 3 走角色管理那條路（`docs/re/72` §3）。
func (s *Scene) enterRosterIfNeeded() {
	if s.facility != nil && s.facility.Facility.Kind == game.FacilityRoster {
		s.beginRoster()
	}
}

func (s *Scene) LeaveFacility() {
	s.facility = nil
	s.player, s.animMask = nil, nil
	s.dirty = true
}

// TickAnim 推進設施圖的局部動畫一拍（規格 26 §3）。
//
// 一拍 ≈ 55 ms（原版比對 BIOS 計時器 `0040:006C`）。**分頻是呈現層的事**——
// 這裡收到一次就推一拍，呼叫端自己決定多久叫一次。
// 回報畫面有沒有變，沒變就不必重畫。
func (s *Scene) TickAnim() bool {
	if s.facility == nil || s.player == nil {
		return false
	}
	elems := s.player.Tick()
	if len(elems) == 0 {
		return false
	}
	render.XorInto(s.animMask, render.FacilityPicWidth, elems)
	s.dirty = true
	return true
}

// InFacility 回報現在是不是在設施畫面。
func (s *Scene) InFacility() bool { return s.facility != nil }

// Facility 回傳目前的設施畫面（沒有就 nil）。
func (s *Scene) Facility() *FacilityScene { return s.facility }
