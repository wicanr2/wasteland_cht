# 寶箱逐格內容（工具輸出，不含推論）

`已定` ＝ 出貨資料就寫死的物品編號；`類別` ＝ 第一次踩到才擲。
寶箱格 **43** 個、讀得準的記錄 **469** 筆。section 5 的指標與別的型別撞在一起而跳過的區塊：**6**（那些圖沒有寶箱）。

走出類別值域而**不報內容**的記錄：**0**。

沒有走歪的記錄，所以「某個物品沒出現」可以當成結論。
件數接在物品後面：`×N` 是出貨就寫死的，`×1dN` 是踩到才擲（第二個 byte 的 bit7）。
「格」空白 ＝ 出貨地圖沒有格子指到這一筆（要靠改寫才到得了）。

| 資源 | 檔案 | 記錄 | 格 | 內容 |
|---:|---|---:|---|---|
| 1 | game1 | 0 |  | 64 Broken toaster、〔擲骰特例〕 ×44、類別 1 ×1d127 |
| 2 | game1 | 0 |  | 〔擲骰特例〕 ×115 |
| 2 | game1 | 1 |  | 4 Knife、57 Snake squeezin、〔擲骰特例〕 ×12 |
| 2 | game1 | 2 |  | 13 M1911A1 45 pistol、30 45 clip |
| 2 | game1 | 3 |  | 81 Room key #18 |
| 2 | game1 | 4 |  | 4 Knife ×3、〔擲骰特例〕 ×75 |
| 2 | game1 | 5 |  | 13 M1911A1 45 pistol、38 Leather jacket、30 45 clip ×2 |
| 2 | game1 | 6 |  | 16 VP91Z 9mm pistol ×3、32 9mm clip ×1d6、〔擲骰特例〕 ×124、類別 1 ×1d127 |
| 3 | game1 | 0 |  | 4 Knife、〔擲骰特例〕 ×15 |
| 3 | game1 | 1 |  | 38 Leather jacket、〔擲骰特例〕 ×65 |
| 3 | game1 | 2 |  | 57 Snake squeezin |
| 3 | game1 | 3 |  | 75 Passkey |
| 3 | game1 | 4 |  | 32 9mm clip |
| 3 | game1 | 5 |  | 7 Plastic explosive、43 Book、〔擲骰特例〕 ×1d122 |
| 3 | game1 | 6 |  | 〔擲骰特例〕 ×20 |
| 3 | game1 | 7 |  | 30 45 clip ×2 |
| 3 | game1 | 8 |  | 4 Knife、92 Fruit ×2 |
| 3 | game1 | 9 |  | 9 Mangler ×2、32 9mm clip ×4、6 Grenade ×5、13 M1911A1 45 pistol ×2、〔擲骰特例〕 ×1d116、類別 1 ×1d127 |
| 3 | game1 | 10 |  | 88 Servo motor、13 M1911A1 45 pistol、〔擲骰特例〕 ×50 |
| 3 | game1 | 11 |  | 65 Chemical |
| 3 | game1 | 12 |  | 55 Shovel、53 Pick ax、1 Ax |
| 3 | game1 | 13 |  | 88 Servo motor |
| 3 | game1 | 14 |  | 〔擲骰特例〕 ×1d22 |
| 3 | game1 | 15 |  | 32 9mm clip ×2、〔擲骰特例〕 ×50 |
| 3 | game1 | 16 |  | 8 TNT、30 45 clip ×5、〔擲骰特例〕 ×44、類別 1 ×1d127 |
| 3 | game1 | 17 |  | 57 Snake squeezin、〔擲骰特例〕 ×15 |
| 3 | game1 | 18 |  | 50 Jug ×3、43 Book、45 Crowbar、〔擲骰特例〕 ×1d32 |
| 3 | game1 | 19 |  | 30 45 clip ×5、〔擲骰特例〕 ×1d72 |
| 3 | game1 | 20 |  | 30 45 clip、〔擲骰特例〕 ×10 |
| 3 | game1 | 21 |  | 32 9mm clip ×3、43 Book ×6 |
| 4 | game1 | 0 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 1 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 2 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 3 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 4 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 5 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 6 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 7 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 8 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 9 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 10 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 11 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 12 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 13 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 14 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 15 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 16 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 17 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 18 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 19 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 20 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 21 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 22 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 23 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 24 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 25 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 26 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 27 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 28 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 29 |  | 32 9mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 30 |  | 30 45 clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 31 |  | 31 7.62mm clip ×1d5、〔擲骰特例〕 ×25 |
| 4 | game1 | 32 | (14,1) | 55 Shovel |
| 4 | game1 | 33 |  | 91 Clay pot ×1d3、50 Jug ×1d3、92 Fruit ×1d5 |
| 4 | game1 | 34 |  | 91 Clay pot ×1d3、50 Jug ×1d3、92 Fruit ×1d5、〔擲骰特例〕 ×44、類別 1 ×1d127 |
| 4 | game1 | 35 |  | 65 Chemical ×3 |
| 4 | game1 | 36 |  | 22 Uzi SMG Mark 27、32 9mm clip ×1d10、36 Bullet proof shirt |
| 4 | game1 | 37 |  | 93 Jewelry ×1d4 |
| 4 | game1 | 38 | (30,30) | 22 Uzi SMG Mark 27、79 Quasar key |
| 4 | game1 | 39 | (30,28) | 32 9mm clip ×1d10 |
| 4 | game1 | 40 | (30,29) | 〔擲骰特例〕 ×1d80、類別 7 ×1d127 |
| 4 | game1 | 41 |  | 37 Kevlar vest、6 Grenade ×1d5、10 Sabot rocket ×1d2、19 M19 rifle、31 7.62mm clip ×1d10、〔擲骰特例〕 ×1d104、類別 3 ×1d127 |
| 4 | game1 | 42 |  | 7 Plastic explosive |
| 6 | game1 | 0 |  | 〔擲骰特例〕 ×25 |
| 6 | game1 | 1 |  | 〔擲骰特例〕 ×75 |
| 6 | game1 | 2 |  | 31 7.62mm clip ×2 |
| 6 | game1 | 3 |  | 57 Snake squeezin、〔擲骰特例〕 ×25 |
| 6 | game1 | 4 |  | 93 Jewelry |
| 6 | game1 | 5 |  | 93 Jewelry、31 7.62mm clip |
| 6 | game1 | 6 |  | 93 Jewelry、36 Bullet proof shirt |
| 6 | game1 | 7 |  | 93 Jewelry ×2、36 Bullet proof shirt |
| 6 | game1 | 8 |  | 30 45 clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 9 |  | 31 7.62mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 10 |  | 32 9mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 11 |  | 30 45 clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 12 |  | 31 7.62mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 13 |  | 32 9mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 14 |  | 30 45 clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 15 |  | 31 7.62mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 16 |  | 32 9mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 17 |  | 30 45 clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 18 |  | 31 7.62mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 19 |  | 32 9mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 20 |  | 30 45 clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 21 |  | 31 7.62mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 22 |  | 89 Sonic key、〔擲骰特例〕 ×100 |
| 6 | game1 | 23 |  | 32 9mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 24 |  | 30 45 clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 25 |  | 31 7.62mm clip ×1d4、〔擲骰特例〕 ×100 |
| 6 | game1 | 26 |  | 57 Snake squeezin、50 Jug ×1d3 |
| 6 | game1 | 27 |  | 〔擲骰特例〕 ×50 |
| 6 | game1 | 28 |  | 〔擲骰特例〕 ×100 |
| 6 | game1 | 29 |  | 94 ×1d104、類別 3 ×1d127 |
| 7 | game2 | 0 |  | 31 7.62mm clip ×1d10 |
| 7 | game2 | 1 |  | 83 Secpass 1 |
| 7 | game2 | 2 |  | 23 AK 97 assault rifle ×2 |
| 7 | game2 | 3 |  | 84 Secpass 3 |
| 7 | game2 | 4 |  | 25 Laser pistol、34 Power pack ×1d5 |
| 7 | game2 | 5 |  | 34 Power pack ×1d5 |
| 7 | game2 | 6 |  | 45 Crowbar |
| 7 | game2 | 7 |  | 28 Laser rifle、34 Power pack ×1d5、12 RPG-7 ×1d5 |
| 7 | game2 | 8 |  | 43 Book ×3 |
| 7 | game2 | 9 |  | 31 7.62mm clip ×5 |
| 8 | game1 | 1 |  | 45 Crowbar |
| 8 | game1 | 2 |  | 45 Crowbar |
| 8 | game1 | 3 |  | 45 Crowbar |
| 8 | game1 | 4 |  | 6 Grenade ×1d5、16 VP91Z 9mm pistol ×1d4、93 Jewelry ×1d4、32 9mm clip ×1d6、30 45 clip ×1d6、〔擲骰特例〕 ×1d72 |
| 8 | game1 | 5 |  | 45 Crowbar ×2、1 Ax ×2、4 Knife ×2、56 Sledge hammer、〔擲骰特例〕 ×22 |
| 8 | game1 | 6 |  | 4 Knife、93 Jewelry ×2 |
| 8 | game1 | 7 |  | 4 Knife、42 Robe、55 Shovel |
| 8 | game1 | 8 |  | 4 Knife ×2、57 Snake squeezin、64 Broken toaster |
| 8 | game1 | 9 |  | 〔擲骰特例〕 ×0 |
| 8 | game1 | 10 |  | 67 Visa card |
| 8 | game1 | 11 |  | 16 VP91Z 9mm pistol |
| 8 | game1 | 12 |  | 55 Shovel |
| 9 | game1 | 0 | (25,29) | 45 Crowbar、92 Fruit ×10 |
| 9 | game1 | 1 |  | 〔擲骰特例〕 ×20 |
| 9 | game1 | 4 | (22,3) | 92 Fruit ×5 |
| 9 | game1 | 6 | (2,13) | 2 Club |
| 10 | game1 | 0 |  | 50 Jug |
| 10 | game1 | 1 |  | 49 Hand mirror |
| 10 | game1 | 2 |  | 32 9mm clip |
| 10 | game1 | 3 |  | 91 Clay pot |
| 10 | game1 | 4 |  | 4 Knife |
| 10 | game1 | 5 |  | 43 Book |
| 10 | game1 | 6 |  | 30 45 clip、93 Jewelry |
| 10 | game1 | 7 |  | 31 7.62mm clip |
| 10 | game1 | 8 |  | 93 Jewelry |
| 10 | game1 | 9 |  | 13 M1911A1 45 pistol、93 Jewelry |
| 10 | game1 | 10 |  | 32 9mm clip ×1d3、93 Jewelry、〔擲骰特例〕 ×20 |
| 10 | game1 | 12 |  | 93 Jewelry ×2 |
| 10 | game1 | 13 |  | 9 Mangler、〔擲骰特例〕 ×50 |
| 10 | game1 | 14 |  | 〔擲骰特例〕 ×45 |
| 10 | game1 | 15 |  | 93 Jewelry、〔擲骰特例〕 ×5 |
| 10 | game1 | 16 |  | 16 VP91Z 9mm pistol、32 9mm clip ×1d5、38 Leather jacket、〔擲骰特例〕 ×100 |
| 10 | game1 | 17 |  | 38 Leather jacket ×3、〔擲骰特例〕 ×100 |
| 10 | game1 | 18 |  | 9 Mangler、38 Leather jacket ×4、32 9mm clip ×8、93 Jewelry ×4、〔擲骰特例〕 ×1d116、類別 1 ×1d127 |
| 12 | game2 | 0 |  | 7 Plastic explosive ×1d5、〔擲骰特例〕 ×1d116、類別 1 ×1d127 |
| 12 | game2 | 1 |  | 2 Club、4 Knife |
| 12 | game2 | 2 |  | 89 Sonic key |
| 12 | game2 | 3 |  | 12 RPG-7 ×1d5 |
| 12 | game2 | 4 |  | 6 Grenade ×1d5、〔擲骰特例〕 ×1d116、類別 1 ×1d127 |
| 12 | game2 | 5 |  | 23 AK 97 assault rifle ×1d2、89 Sonic key、31 7.62mm clip ×1d10 |
| 12 | game2 | 6 |  | 23 AK 97 assault rifle ×1d2、89 Sonic key、31 7.62mm clip ×1d6、39 Kevlar suit、〔擲骰特例〕 ×1d116、類別 1 ×1d127 |
| 12 | game2 | 7 |  | 11 LAW rocket ×1d10、7 Plastic explosive ×1d5 |
| 12 | game2 | 8 |  | 9 Mangler ×1d10、7 Plastic explosive ×1d5 |
| 12 | game2 | 9 |  | 12 RPG-7 ×1d5、8 TNT ×1d5 |
| 12 | game2 | 10 |  | 11 LAW rocket ×1d10、8 TNT ×1d5 |
| 12 | game2 | 11 |  | 12 RPG-7 ×1d5、6 Grenade ×1d5 |
| 13 | game2 | 0 |  | 50 Jug |
| 13 | game2 | 1 |  | 66 Clone fluid |
| 13 | game2 | 2 |  | 50 Jug |
| 13 | game2 | 3 |  | 50 Jug |
| 13 | game2 | 4 |  | 50 Jug |
| 13 | game2 | 5 | (21,11) | 40 Pseudo-chitin armor、34 Power pack ×25 |
| 13 | game2 | 6 |  | 86 Secpass A |
| 13 | game2 | 7 |  | 86 Secpass A |
| 13 | game2 | 8 |  | 41 Rad suit ×1d12 |
| 13 | game2 | 9 |  | 50 Jug ×1d3 |
| 13 | game2 | 10 | (30,11) | 13 M1911A1 45 pistol ×1d10、16 VP91Z 9mm pistol ×1d10、19 M19 rifle ×1d10、23 AK 97 assault rifle ×1d10、24 M1989A1 Nato assault rifle ×1d10、21 Mac 17 SMG ×1d10、22 Uzi SMG Mark 27 ×1d10、30 45 clip ×1d100、31 7.62mm clip ×1d100、32 9mm clip ×1d100、34 Power pack ×1d10 |
| 13 | game2 | 11 | (23,11) | 40 Pseudo-chitin armor、29 Meson cannon |
| 13 | game2 | 12 | (25,11) | 40 Pseudo-chitin armor、11 LAW rocket ×1d20 |
| 13 | game2 | 13 | (25,14) | 40 Pseudo-chitin armor、12 RPG-7 ×1d20 |
| 13 | game2 | 14 | (25,16) | 40 Pseudo-chitin armor、10 Sabot rocket ×1d20 |
| 13 | game2 | 15 | (23,16) | 40 Pseudo-chitin armor |
| 13 | game2 | 16 | (21,16) | 40 Pseudo-chitin armor |
| 13 | game2 | 17 |  | 66 Clone fluid |
| 13 | game2 | 18 |  | 50 Jug |
| 13 | game2 | 19 |  | 50 Jug ×1d3 |
| 13 | game2 | 20 |  | 50 Jug ×1d3 |
| 15 | game2 | 0 |  | 68 Fusion cell |
| 15 | game2 | 1 | (20,26) | 76 Plasma coupler |
| 15 | game2 | 2 |  | 85 Secpass 7、77 Power converter |
| 15 | game2 | 3 |  | 34 Power pack |
| 16 | game2 | 0 | (16,13) | 34 Power pack ×1d5、68 Fusion cell ×3、11 LAW rocket ×1d5 |
| 16 | game2 | 1 |  | 80 Rom board、12 RPG-7 ×1d5 |
| 16 | game2 | 2 |  | 77 Power converter、34 Power pack ×1d5 |
| 16 | game2 | 3 |  | 11 LAW rocket ×2 |
| 16 | game2 | 4 |  | 68 Fusion cell ×2、34 Power pack ×1d5 |
| 16 | game2 | 5 |  | 34 Power pack ×1d5 |
| 16 | game2 | 6 |  | 88 Servo motor |
| 16 | game2 | 7 |  | 34 Power pack ×1d5、76 Plasma coupler |
| 16 | game2 | 8 |  | 34 Power pack ×1d7 |
| 16 | game2 | 9 |  | 77 Power converter |
| 17 | game2 | 0 |  | 34 Power pack ×1d10、12 RPG-7 ×1d5 |
| 17 | game2 | 1 |  | 34 Power pack ×1d10 |
| 19 | game2 | 0 | (25,1) | 35 Power armor ×2、11 LAW rocket ×10、12 RPG-7 ×10、34 Power pack ×20 |
| 21 | game2 | 0 |  | 22 Uzi SMG Mark 27、32 9mm clip ×1d10、〔擲骰特例〕 ×100 |
| 21 | game2 | 1 |  | 22 Uzi SMG Mark 27、32 9mm clip ×1d10、〔擲骰特例〕 ×100 |
| 21 | game2 | 2 |  | 59 Antitoxin |
| 21 | game2 | 3 |  | 59 Antitoxin |
| 21 | game2 | 4 |  | 59 Antitoxin |
| 22 | game2 | 1 |  | 60 Finster's head、〔擲骰特例〕 ×1d104、類別 3 ×1d127 |
| 22 | game2 | 2 |  | 27 Laser carbine ×2、34 Power pack ×6、68 Fusion cell、40 Pseudo-chitin armor |
| 23 | game2 | 0 |  | 38 Leather jacket、42 Robe |
| 23 | game2 | 1 |  | 50 Jug |
| 23 | game2 | 2 |  | 58 Android head |
| 23 | game2 | 3 |  | 87 Secpass B |
| 24 | game2 | 0 |  | 68 Fusion cell |
| 24 | game2 | 1 |  | 77 Power converter |
| 24 | game2 | 2 |  | 類別 14 ×1d10 |
| 24 | game2 | 3 |  | 93 Jewelry、88 Servo motor |
| 24 | game2 | 4 |  | 88 Servo motor |
| 24 | game2 | 5 |  | 80 Rom board |
| 24 | game2 | 6 |  | 30 45 clip ×3 |
| 24 | game2 | 7 |  | 34 Power pack ×3 |
| 24 | game2 | 8 |  | 25 Laser pistol |
| 24 | game2 | 9 |  | 31 7.62mm clip ×4 |
| 24 | game2 | 10 |  | 49 Hand mirror |
| 24 | game2 | 11 |  | 31 7.62mm clip ×3、93 Jewelry、54 Rope |
| 24 | game2 | 12 |  | 27 Laser carbine、34 Power pack ×1d10 |
| 24 | game2 | 13 |  | 25 Laser pistol |
| 25 | game2 | 0 |  | 10 Sabot rocket ×1d5 |
| 25 | game2 | 1 |  | 34 Power pack ×1d5 |
| 25 | game2 | 2 |  | 3 Chainsaw |
| 25 | game2 | 3 |  | 11 LAW rocket ×1d5 |
| 25 | game2 | 4 |  | 88 Servo motor |
| 25 | game2 | 5 |  | 80 Rom board |
| 25 | game2 | 6 |  | 34 Power pack ×3 |
| 25 | game2 | 7 |  | 80 Rom board、39 Kevlar suit |
| 25 | game2 | 8 |  | 25 Laser pistol、34 Power pack ×1d3 |
| 25 | game2 | 9 |  | 39 Kevlar suit |
| 25 | game2 | 10 |  | 64 Broken toaster |
| 25 | game2 | 11 |  | 88 Servo motor |
| 25 | game2 | 12 |  | 12 RPG-7 ×1d3 |
| 25 | game2 | 13 |  | 58 Android head |
| 25 | game2 | 14 |  | 64 Broken toaster |
| 27 | game1 | 0 |  | 18 M17 carbine ×2、31 7.62mm clip ×1d7、〔擲骰特例〕 ×50 |
| 27 | game1 | 1 |  | 50 Jug ×1d5、15 Throwing knife ×1d5 |
| 27 | game1 | 2 |  | 4 Knife ×2、63 Bloodstaff |
| 27 | game1 | 3 |  | 37 Kevlar vest |
| 27 | game1 | 4 |  | 36 Bullet proof shirt ×2 |
| 27 | game1 | 5 |  | 21 Mac 17 SMG、30 45 clip ×1d4 |
| 28 | game1 | 0 |  | 62 Bloodstaff |
| 28 | game1 | 1 |  | 19 M19 rifle ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 2 |  | 11 LAW rocket ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 3 |  | 11 LAW rocket ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 4 |  | 11 LAW rocket ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 5 |  | 19 M19 rifle ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 6 |  | 11 LAW rocket ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 7 |  | 11 LAW rocket ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 8 |  | 19 M19 rifle ×1d2、類別 14 ×1d7、63 Bloodstaff、〔擲骰特例〕 ×50 |
| 28 | game1 | 9 |  | 19 M19 rifle ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 10 |  | 19 M19 rifle ×1d2、類別 14 ×1d7、〔擲骰特例〕 ×50 |
| 28 | game1 | 11 |  | 19 M19 rifle、31 7.62mm clip ×1d5 |
| 28 | game1 | 12 |  | 22 Uzi SMG Mark 27、32 9mm clip ×1d5 |
| 28 | game1 | 13 |  | 21 Mac 17 SMG、30 45 clip ×1d5、63 Bloodstaff |
| 28 | game1 | 14 |  | 24 M1989A1 Nato assault rifle、31 7.62mm clip ×1d5 |
| 29 | game1 | 0 |  | 6 Grenade |
| 29 | game1 | 1 |  | 16 VP91Z 9mm pistol ×2 |
| 29 | game1 | 2 |  | 7 Plastic explosive |
| 29 | game1 | 3 |  | 32 9mm clip |
| 29 | game1 | 4 |  | 92 Fruit |
| 29 | game1 | 5 |  | 6 Grenade |
| 29 | game1 | 6 |  | 92 Fruit |
| 29 | game1 | 7 |  | 92 Fruit |
| 29 | game1 | 8 |  | 92 Fruit |
| 31 | game1 | 0 |  | 41 Rad suit ×2、36 Bullet proof shirt |
| 31 | game1 | 1 |  | 24 M1989A1 Nato assault rifle ×2、13 M1911A1 45 pistol ×2、30 45 clip ×5、31 7.62mm clip ×7 |
| 31 | game1 | 2 |  | 6 Grenade ×1d15、7 Plastic explosive ×1d3、8 TNT |
| 31 | game1 | 3 |  | 30 45 clip ×1d25、31 7.62mm clip ×1d10、32 9mm clip ×1d15、34 Power pack |
| 31 | game1 | 4 |  | 33 Howitzer shell ×1d3 |
| 31 | game1 | 5 |  | 82 Ruby ring、63 Bloodstaff |
| 32 | game1 | 0 |  | 4 Knife ×1d2 |
| 32 | game1 | 1 |  | 4 Knife ×1d2、50 Jug |
| 32 | game1 | 2 |  | 4 Knife ×1d2、91 Clay pot、54 Rope |
| 32 | game1 | 3 |  | 4 Knife ×1d2、54 Rope |
| 32 | game1 | 4 |  | 4 Knife ×1d2 |
| 32 | game1 | 5 |  | 4 Knife ×1d2、52 Match、56 Sledge hammer |
| 32 | game1 | 6 |  | 16 VP91Z 9mm pistol ×2、13 M1911A1 45 pistol、32 9mm clip ×5、30 45 clip ×3、48 Geiger counter |
| 32 | game1 | 7 |  | 23 AK 97 assault rifle、31 7.62mm clip ×5、63 Bloodstaff |
| 32 | game1 | 8 |  | 49 Hand mirror |
| 32 | game1 | 9 |  | 4 Knife、52 Match、94 ×110 |
| 32 | game1 | 10 |  | 50 Jug |
| 32 | game1 | 15 |  | 〔擲骰特例〕 ×16 |
| 32 | game1 | 16 |  | 4 Knife ×2、94 ×4 |
| 32 | game1 | 17 |  | 4 Knife、92 Fruit |
| 32 | game1 | 18 |  | 32 9mm clip ×2 |
| 32 | game1 | 19 | (27,29) | 57 Snake squeezin |
| 33 | game1 | 0 | (1,1) | 31 7.62mm clip ×1d3、〔擲骰特例〕 ×1d72 |
| 33 | game1 | 1 | (2,1) | 42 Robe ×1d5 |
| 33 | game1 | 2 |  | 13 M1911A1 45 pistol、30 45 clip ×2、63 Bloodstaff |
| 33 | game1 | 4 |  | 91 Clay pot |
| 33 | game1 | 5 |  | 57 Snake squeezin |
| 33 | game1 | 6 |  | 57 Snake squeezin |
| 33 | game1 | 7 |  | 62 Bloodstaff |
| 33 | game1 | 8 |  | 62 Bloodstaff |
| 34 | game1 | 0 | (27,10) | 22 Uzi SMG Mark 27、32 9mm clip ×3 |
| 34 | game1 | 1 | (30,10) | 21 Mac 17 SMG |
| 34 | game1 | 2 | (28,12) | 32 9mm clip ×5 |
| 34 | game1 | 3 | (29,12) | 31 7.62mm clip ×5 |
| 34 | game1 | 4 |  | 23 AK 97 assault rifle ×1d2、24 M1989A1 Nato assault rifle ×1d2、32 9mm clip ×1d20、31 7.62mm clip ×1d20、39 Kevlar suit、8 TNT ×1d10、46 Engine |
| 34 | game1 | 5 |  | 62 Bloodstaff |
| 34 | game1 | 6 |  | 63 Bloodstaff |
| 35 | game2 | 0 |  | 34 Power pack ×3 |
| 35 | game2 | 1 |  | 78 Pulsar key、1 Ax |
| 35 | game2 | 2 |  | 42 Robe ×3、34 Power pack、43 Book ×4 |
| 35 | game2 | 3 |  | 27 Laser carbine、34 Power pack ×2 |
| 35 | game2 | 4 | (1,9) | 23 AK 97 assault rifle ×12、31 7.62mm clip ×50 |
| 35 | game2 | 5 | (2,10) | 28 Laser rifle ×2、29 Meson cannon、34 Power pack ×12 |
| 35 | game2 | 6 |  | 7 Plastic explosive ×2 |
| 35 | game2 | 7 |  | 26 Ion beamer、34 Power pack ×1d6 |
| 35 | game2 | 8 |  | 4 Knife ×2 |
| 35 | game2 | 9 |  | 69 Grazer bat fetish |
| 35 | game2 | 10 |  | 47 Gas mask ×2 |
| 35 | game2 | 11 |  | 79 Quasar key |
| 35 | game2 | 12 |  | 91 Clay pot ×4 |
| 35 | game2 | 13 |  | 64 Broken toaster |
| 36 | game2 | 0 |  | 34 Power pack ×1d3、25 Laser pistol |
| 36 | game2 | 1 |  | 5 Proton ax |
| 36 | game2 | 2 | (30,16) | 42 Robe ×2、57 Snake squeezin、43 Book ×4 |
| 36 | game2 | 3 |  | 3 Chainsaw、43 Book ×2 |
| 36 | game2 | 4 | (7,14) | 42 Robe ×1d2、38 Leather jacket ×1d3 |
| 36 | game2 | 5 | (9,18) | 57 Snake squeezin ×1d30、67 Visa card ×1d10、92 Fruit ×1d15 |
| 36 | game2 | 6 | (9,22) | 31 7.62mm clip ×1d5、34 Power pack ×8、6 Grenade ×1d4、11 LAW rocket ×1d3 |
| 36 | game2 | 7 |  | 27 Laser carbine、34 Power pack ×3 |
| 36 | game2 | 8 |  | 25 Laser pistol ×2、34 Power pack ×4 |
| 36 | game2 | 9 |  | 74 Onyx ring |
| 36 | game2 | 10 |  | 74 Onyx ring、7 Plastic explosive ×1d7 |
| 36 | game2 | 11 |  | 73 Nova key、34 Power pack ×1d9、〔擲骰特例〕 ×1d116、類別 1 ×1d127 |
| 36 | game2 | 12 |  | 73 Nova key、34 Power pack ×1d5、〔擲骰特例〕 ×44、類別 1 ×1d127 |
| 36 | game2 | 13 |  | 73 Nova key、34 Power pack ×1d3、〔擲骰特例〕 ×100 |
| 36 | game2 | 14 |  | 67 Visa card ×2、43 Book ×2、47 Gas mask、〔擲骰特例〕 ×80 |
| 36 | game2 | 15 |  | 61 Blackstar key |
| 36 | game2 | 16 |  | 27 Laser carbine、34 Power pack ×1d5、〔擲骰特例〕 ×44、類別 1 ×1d127 |
| 36 | game2 | 17 |  | 57 Snake squeezin ×1d4、34 Power pack ×1d4、〔擲骰特例〕 ×1d72 |
| 36 | game2 | 18 |  | 34 Power pack ×1d10、〔擲骰特例〕 ×20 |
| 36 | game2 | 19 |  | 34 Power pack ×1d10、〔擲骰特例〕 ×50 |
| 36 | game2 | 20 |  | 34 Power pack ×1d10、〔擲骰特例〕 ×50 |
| 36 | game2 | 21 |  | 34 Power pack ×1d10、〔擲骰特例〕 ×50 |
| 36 | game2 | 22 |  | 42 Robe ×2、57 Snake squeezin、43 Book ×4 |
| 36 | game2 | 23 |  | 73 Nova key |
| 38 | game2 | 0 |  | 41 Rad suit ×1d3、42 Robe ×1d5 |
| 38 | game2 | 1 |  | 6 Grenade ×3、22 Uzi SMG Mark 27、32 9mm clip ×1d5 |
| 38 | game2 | 2 |  | 22 Uzi SMG Mark 27、32 9mm clip ×2、〔擲骰特例〕 ×10 |
| 38 | game2 | 3 |  | 4 Knife ×3、1 Ax |
| 38 | game2 | 4 |  | 7 Plastic explosive ×1d5 |
| 38 | game2 | 5 |  | 24 M1989A1 Nato assault rifle、31 7.62mm clip ×1d10 |
| 38 | game2 | 6 |  | 30 45 clip ×5、31 7.62mm clip ×1d10 |
| 38 | game2 | 7 | (24,6) | 92 Fruit ×1d3、43 Book、50 Jug |
| 38 | game2 | 8 | (23,6) | 2 Club ×1d4、13 M1911A1 45 pistol、31 7.62mm clip ×1d5 |
| 38 | game2 | 9 |  | 25 Laser pistol、34 Power pack ×1d2、〔擲骰特例〕 ×100 |
| 39 | game2 | 0 |  | 74 Onyx ring |
| 39 | game2 | 1 |  | 11 LAW rocket ×1d10、10 Sabot rocket ×1d5、12 RPG-7 ×1d5 |
| 39 | game2 | 2 |  | 24 M1989A1 Nato assault rifle ×1d5、31 7.62mm clip ×1d20 |
| 39 | game2 | 3 |  | 〔擲骰特例〕 ×1d116、類別 1 ×1d127 |
| 40 | game2 | 5 |  | 4 Knife、〔擲骰特例〕 ×100 |
| 40 | game2 | 6 |  | 4 Knife、〔擲骰特例〕 ×100 |
| 40 | game2 | 7 |  | 4 Knife、〔擲骰特例〕 ×100 |
| 40 | game2 | 8 |  | 4 Knife、〔擲骰特例〕 ×100 |
| 40 | game2 | 9 |  | 4 Knife、〔擲骰特例〕 ×100 |
| 40 | game2 | 10 |  | 30 45 clip ×1d5、13 M1911A1 45 pistol、21 Mac 17 SMG ×2、〔擲骰特例〕 ×1d116、類別 1 ×1d127 |
| 40 | game2 | 11 |  | 94 ×1d104、類別 3 ×1d127 |
| 40 | game2 | 12 |  | 16 VP91Z 9mm pistol ×1d5、32 9mm clip ×1d5、30 45 clip ×1d5、13 M1911A1 45 pistol ×1d5、〔擲骰特例〕 ×1d72 |
| 40 | game2 | 13 |  | 16 VP91Z 9mm pistol ×1d5、32 9mm clip ×1d5、30 45 clip ×1d5、13 M1911A1 45 pistol ×1d5、〔擲骰特例〕 ×1d72 |
| 40 | game2 | 14 |  | 43 Book |
| 40 | game2 | 16 |  | 16 VP91Z 9mm pistol ×1d5、32 9mm clip ×1d5、30 45 clip ×1d5、13 M1911A1 45 pistol ×1d5、6 Grenade ×1d10、93 Jewelry、〔擲骰特例〕 ×100 |
| 40 | game2 | 18 |  | 67 Visa card |
| 40 | game2 | 20 |  | 4 Knife、94 ×20 |
| 40 | game2 | 21 |  | 19 M19 rifle ×1d3、31 7.62mm clip ×1d6、〔擲骰特例〕 ×20 |
| 40 | game2 | 22 |  | 13 M1911A1 45 pistol、45 Crowbar、38 Leather jacket、94 ×20 |
| 40 | game2 | 23 |  | 5 Proton ax |
| 40 | game2 | 24 |  | 44 Canteen |
| 40 | game2 | 25 |  | 〔擲骰特例〕 ×20 |
| 40 | game2 | 26 | (1,1) | 16 VP91Z 9mm pistol ×1d5、32 9mm clip ×1d5、30 45 clip ×1d5、13 M1911A1 45 pistol ×1d5、6 Grenade ×1d10、93 Jewelry、〔擲骰特例〕 ×100 |
| 41 | game2 | 0 |  | 〔擲骰特例〕 ×50 |
| 41 | game2 | 1 |  | 〔擲骰特例〕 ×25 |
| 41 | game2 | 2 |  | 〔擲骰特例〕 ×5 |
| 41 | game2 | 3 |  | 〔擲骰特例〕 ×10 |
| 41 | game2 | 4 |  | 〔擲骰特例〕 ×7 |
| 41 | game2 | 5 |  | 〔擲骰特例〕 ×3 |
| 41 | game2 | 6 |  | 〔擲骰特例〕 ×5 |
| 41 | game2 | 7 |  | 〔擲骰特例〕 ×5 |
| 41 | game2 | 8 |  | 〔擲骰特例〕 ×5 |
| 41 | game2 | 9 |  | 94 ×20 |
| 41 | game2 | 10 |  | 94 ×15 |
| 41 | game2 | 11 |  | 〔擲骰特例〕 ×25 |
| 41 | game2 | 12 |  | 15 Throwing knife ×2、〔擲骰特例〕 ×100 |
| 41 | game2 | 13 |  | 50 Jug、〔擲骰特例〕 ×5 |
| 41 | game2 | 14 |  | 93 Jewelry、〔擲骰特例〕 ×50 |
| 41 | game2 | 15 |  | 〔擲骰特例〕 ×20 |
| 41 | game2 | 17 |  | 25 Laser pistol、52 Match ×1d20、〔擲骰特例〕 |
| 41 | game2 | 18 |  | 〔擲骰特例〕 ×75 |
| 41 | game2 | 19 |  | 〔擲骰特例〕 ×7 |
| 41 | game2 | 20 |  | 16 VP91Z 9mm pistol、32 9mm clip ×3、38 Leather jacket、54 Rope、51 Map、〔擲骰特例〕 ×100 |
| 41 | game2 | 21 |  | 67 Visa card ×1d2、59 Antitoxin、92 Fruit ×1d3 |
| 41 | game2 | 22 |  | 〔擲骰特例〕 ×10 |
| 41 | game2 | 23 |  | 13 M1911A1 45 pistol、6 Grenade |
| 41 | game2 | 24 |  | 21 Mac 17 SMG、30 45 clip ×3、6 Grenade、〔擲骰特例〕 ×50 |
| 41 | game2 | 25 |  | 22 Uzi SMG Mark 27、32 9mm clip ×1d2 |
| 41 | game2 | 26 |  | 24 M1989A1 Nato assault rifle、31 7.62mm clip ×1d4、36 Bullet proof shirt |
| 41 | game2 | 27 |  | 25 Laser pistol、34 Power pack ×1d2、39 Kevlar suit、〔擲骰特例〕 ×50 |
| 41 | game2 | 28 |  | 13 M1911A1 45 pistol、30 45 clip ×1d5、〔擲骰特例〕 ×75 |
| 41 | game2 | 29 |  | 〔擲骰特例〕 ×100 |
| 41 | game2 | 30 |  | 〔擲骰特例〕 ×50 |
| 41 | game2 | 31 |  | 〔擲骰特例〕 ×25 |
| 41 | game2 | 32 |  | 〔擲骰特例〕 ×5 |
| 41 | game2 | 33 |  | 〔擲骰特例〕 ×10 |
| 41 | game2 | 34 |  | 〔擲骰特例〕 ×7 |
| 41 | game2 | 35 |  | 〔擲骰特例〕 ×3 |
| 41 | game2 | 36 |  | 〔擲骰特例〕 ×5 |
| 41 | game2 | 37 |  | 〔擲骰特例〕 ×5 |
| 41 | game2 | 38 |  | 〔擲骰特例〕 ×5 |
| 41 | game2 | 39 |  | 94 ×20 |
| 41 | game2 | 40 |  | 94 ×15 |
| 41 | game2 | 41 |  | 13 M1911A1 45 pistol、30 45 clip ×3、19 M19 rifle、31 7.62mm clip ×2、38 Leather jacket、54 Rope、52 Match ×5、55 Shovel、51 Map、44 Canteen、〔擲骰特例〕 ×100 |
| 41 | game2 | 42 |  | 57 Snake squeezin ×1d5、50 Jug |
| 41 | game2 | 43 |  | 21 Mac 17 SMG、30 45 clip ×3、4 Knife |
| 41 | game2 | 44 |  | 51 Map、52 Match ×1d2、30 45 clip ×1d5、〔擲骰特例〕 ×50 |
| 41 | game2 | 45 |  | 13 M1911A1 45 pistol、30 45 clip、〔擲骰特例〕 ×1d6 |
| 41 | game2 | 46 |  | 21 Mac 17 SMG ×1d3、38 Leather jacket ×1d2、4 Knife ×1d5、6 Grenade ×1d3、34 Power pack ×1d3、〔擲骰特例〕 ×57 |
| 42 | game2 | 0 | (29,1) | 35 Power armor、34 Power pack ×1d7 |
| 42 | game2 | 1 | (21,1) | 35 Power armor、34 Power pack ×1d7 |
| 42 | game2 | 2 | (23,1) | 35 Power armor、34 Power pack ×1d7 |
| 42 | game2 | 3 | (25,1) | 35 Power armor、34 Power pack ×1d7 |
| 42 | game2 | 4 | (27,1) | 35 Power armor、34 Power pack ×1d7 |
| 42 | game2 | 5 |  | 76 Plasma coupler、92 Fruit ×2、90 Toaster、67 Visa card ×2 |
| 42 | game2 | 6 |  | 73 Nova key、90 Toaster、34 Power pack ×1d4 |
| 42 | game2 | 7 |  | 78 Pulsar key、90 Toaster、56 Sledge hammer、34 Power pack ×4 |
| 42 | game2 | 8 |  | 90 Toaster、7 Plastic explosive ×1d6、46 Engine、57 Snake squeezin |
| 42 | game2 | 9 | (1,11) | 42 Robe ×1d5 |
| 42 | game2 | 10 |  | 54 Rope |
| 42 | game2 | 11 |  | 87 Secpass B |
| 43 | game1 | 0 |  | 44 Canteen ×1d3、49 Hand mirror ×1d2、55 Shovel ×1d2、53 Pick ax |
| 43 | game1 | 1 |  | 54 Rope、〔擲骰特例〕 ×50 |
| 43 | game1 | 2 |  | 54 Rope、47 Gas mask ×4 |
| 43 | game1 | 3 |  | 91 Clay pot ×1d2、49 Hand mirror、〔擲骰特例〕 ×10 |
| 43 | game1 | 4 |  | 91 Clay pot、8 TNT、〔擲骰特例〕 ×10 |
| 43 | game1 | 5 |  | 91 Clay pot ×1d2、49 Hand mirror、〔擲骰特例〕 ×10 |
| 43 | game1 | 6 |  | 16 VP91Z 9mm pistol、30 45 clip ×1d2、〔擲骰特例〕 ×10 |
| 43 | game1 | 7 |  | 13 M1911A1 45 pistol、32 9mm clip ×1d2、〔擲骰特例〕 ×10 |
| 43 | game1 | 8 |  | 15 Throwing knife ×1d3、6 Grenade、〔擲骰特例〕 ×10 |
| 43 | game1 | 9 |  | 〔擲骰特例〕 ×10 |
| 43 | game1 | 10 |  | 〔擲骰特例〕 ×10 |
| 43 | game1 | 13 |  | 44 Canteen |
| 43 | game1 | 14 |  | 15 Throwing knife、36 Bullet proof shirt、〔擲骰特例〕 ×10 |
| 43 | game1 | 15 |  | 4 Knife ×1d2 |
| 43 | game1 | 16 |  | 1 Ax ×1d2 |
| 43 | game1 | 17 |  | 8 TNT、30 45 clip ×1d2、54 Rope |
| 43 | game1 | 18 |  | 〔擲骰特例〕 ×10 |
| 43 | game1 | 19 |  | 〔擲骰特例〕 ×10 |
| 49 | game1 | 0 |  | 13 M1911A1 45 pistol、30 45 clip ×1d5 |
| 49 | game1 | 1 |  | 24 M1989A1 Nato assault rifle ×1d2、6 Grenade ×1d3、〔擲骰特例〕 ×10 |
| 49 | game1 | 2 |  | 14 Spear ×1d10 |
| 49 | game1 | 3 |  | 1 Ax ×1d10、15 Throwing knife ×1d15 |
| 49 | game1 | 4 |  | 13 M1911A1 45 pistol、24 M1989A1 Nato assault rifle ×1d2、30 45 clip ×1d7、〔擲骰特例〕 ×10 |
| 49 | game1 | 5 |  | 1 Ax ×1d10、2 Club ×1d9、4 Knife ×1d3 |
| 49 | game1 | 6 |  | 4 Knife ×1d10 |
| 49 | game1 | 7 |  | 22 Uzi SMG Mark 27、32 9mm clip ×1d2、〔擲骰特例〕 ×50 |
| 49 | game1 | 8 |  | 24 M1989A1 Nato assault rifle ×1d2、31 7.62mm clip ×1d3 |
| 49 | game1 | 9 |  | 14 Spear ×1d5、4 Knife ×1d3、22 Uzi SMG Mark 27 |
| 49 | game1 | 12 | (3,11) | 32 9mm clip ×1d10、31 7.62mm clip ×1d10、30 45 clip ×1d10 |
| 49 | game1 | 14 | (6,4) | 4 Knife ×1d3、14 Spear ×1d3、1 Ax ×1d3 |
| 49 | game1 | 15 | (6,2) | 13 M1911A1 45 pistol ×1d3、24 M1989A1 Nato assault rifle ×1d2、11 LAW rocket ×1d3 |
| 49 | game1 | 16 |  | 21 Mac 17 SMG、〔擲骰特例〕 ×100 |
| 49 | game1 | 17 |  | 〔擲骰特例〕 ×10 |
| 49 | game1 | 18 |  | 〔擲骰特例〕 ×10 |
| 49 | game1 | 19 |  | 〔擲骰特例〕 ×10 |
| 49 | game1 | 20 |  | 〔擲骰特例〕 ×10 |
