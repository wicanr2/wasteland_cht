# 推廣片的設計 token —— 這一部叫「鏽與沙」。
#
# 色票不是憑喜好挑的，是從遊戲自己的畫面量出來的（`-colors 8` 直方圖）：
#
#	近黑 #020202   四張截圖的第一名，佔比都在四成以上（EGA 的黑底）
#	鏽色 #9A5331   標題畫面的地面、沙漠的岩層、結局的土坡
#	沙黃 #F0E046   荒漠地圖的主色
#	骨白 #B0AEAB   標題的 WASTELAND 字體
#
# 母題選「軍規檔案」：掃描線 ＋ 四角的框線記號。理由是這款遊戲的主角是
# 一支準軍事單位（Desert Rangers），而 1988 年的 EGA 畫面本來就帶 CRT 味。
#
# ⚠ **設計 token 不沿用上一部片。** 這一部與先前兩部的差異：
# 色票來源（從實機截圖量 vs 憑主題挑）、母題（檔案框線 vs 羊皮紙／魔法陣）、
# 字體（黑體 vs 襯線）、敘事骨架（保存敘事 vs 工程／世界觀）。

THEME_NAME="鏽與沙"

BG_DEEP='#0b0704'  # 背景漸層深端（近黑，帶一點暖）
BG_LITE='#3a2415'  # 背景漸層淺端（暗鏽）
RUST='#9a5331'     # 框線與英文小標
EMBER='#c8541e'    # 標題的陰影層
SAND='#e0c455'     # 標題主色
BONE='#e8dcc4'     # 內文
DIM='#8a7658'      # 出處、註記

FT=/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc     # 標題
FB=/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc  # 內文

W=1280; H=720; FPS=25

PACE_CARD=6   # 文字卡
PACE_SHOT=5   # 截圖
PACE_QUOTE=6  # 引用卡
