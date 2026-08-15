package render

import "fmt"

// `IC0_9.WLF` 的十張疊圖（docs/re/48、docs/spec/03 §2.9）。
//
// 編號空間與地圖圖形共用：0–9 是這十張，≥ 10 是圖磚編號 `值 − 10`。
// 分界寫在原版的 `cmp al, 0Ah` 裡，見 IconCount。
const (
	IconBlack      = 0 // 遮罩全 0、圖形全 0 → 把一格塗黑。**不是「不畫」**
	IconRobot      = 1 // 敵人種類 5
	IconCyborg     = 2 // 敵人種類 4
	IconMutant     = 3 // 敵人種類 2
	IconHumanoid   = 4 // 敵人種類 3，也是查不到種類時的預設
	IconLoot       = 5 // 第 1 層 nibble ＝ 5 的格（寶箱／掉落物）
	IconAnimal     = 6 // 敵人種類 1
	IconParty      = 7 // 隊伍自己（sub_16716 的 `mov al, 7`）
	IconRadiation  = 8 // 第 1 層 nibble ＝ 9 的格（輻射區），只在夜間
	IconOtherGroup = 9 // 其他分隊（DISBAND 拆出去的那幾組）

	// IconCount 是疊圖與圖磚的分界（原版 sub_18024 的 `cmp al, 0Ah`）。
	IconCount = 10
)

// DrawIcon 把一張疊圖合成到地圖視窗的第 (col, row) 格。
//
// 合成規則見 DrawOverlay：`螢幕 ← (背景 AND 遮罩) OR 疊圖`。
// 地圖畫完之後才呼叫這一支。
func (f *Frame) DrawIcon(g *Graphics, icon byte, col, row int) error {
	if icon >= IconCount {
		return fmt.Errorf("疊圖編號 %d 不是疊圖（≥ %d 是圖磚）", icon, IconCount)
	}
	im, err := g.Get(icon)
	if err != nil {
		return err
	}
	if int(icon) >= len(g.Masks) {
		return fmt.Errorf("疊圖 %d 沒有對應的遮罩（只有 %d 張）", icon, len(g.Masks))
	}
	x := ViewX - TileSize/2 + col*TileSize
	y := ViewY - TileSize/2 + row*TileSize
	f.DrawOverlay(im, g.Masks[icon], x, y, MapClip())
	return nil
}

// DrawParty 把隊伍圖示疊在固定的第 (9, 4) 格上。DrawMap 之後呼叫。
func (f *Frame) DrawParty(g *Graphics) error {
	return f.DrawIcon(g, IconParty, PartyCol, PartyRow)
}
