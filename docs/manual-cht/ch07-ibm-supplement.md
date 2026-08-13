# IBM 版補充說明書

掃描檔 wl27.jpg – wl30.jpg。這是與說明書分開的一本黃色小冊，另行編頁（第 1 頁至第 6 頁），內容為 IBM PC 版的安裝與操作，與主說明書描述的 Apple II 版流程不同。

---

## wl27（補充說明書 第 1 頁）

〔版面：黃色紙張，四周為斜線花邊框；右上角插圖為一名持雷射槍、戴護目鏡的短髮人物；右下角為 SOFT-WORLD 星形標誌〕

### 荒野遊俠 IBM 版補充說明書

#### 一、進入遊戲前的準備工作

　　在正式啟動遊戲前，你必須先利用遊戲磁片 MASTER DISK1 中的 SETUP.EXE 檔進行磁片之轉錄，將二片原始磁片轉錄成三片磁片，請先準備三片已經格式化（formatted）的 360K 空白磁片，並分別將此三片空白磁片標名為 “Program Disk”、“Scenario Disk1” 及 “Scenario Disk2”；如果你的電腦內裝有硬碟，也可利用這檔案將遊戲載入硬碟中。執行這些步驟，需要大約十五分鐘，請務必耐心操作。以下介紹遊戲轉錄之方法：

※開機出現 A＞後，將遊戲之 MASTER DISK1 置於 A 磁碟機中，關好機門；鍵入 SETUP，並按 ENTER 鍵。

※螢幕上會出現如下的選擇畫面：

```
Select 1-5：
 1. Complete Set-Up.（遊戲之載入）
 2. Create Program & Pictures only.（改變圖形模式）
   （to change graphics modes）
 3. Create new Maps only.（取消原有的人物）
   （destroys old characters）
```

---

## wl27（補充說明書 第 2 頁）

```
 4. Restart game with old characters.（利用原有人物重新進行遊戲）
 5. Quit.（退出遊戲，回到 DOS）
```

　　請鍵入 `1` 鍵，選擇執行 Complete Set-Up 項。隨後電腦會詢問你是否要在硬碟中進行遊戲。

```
Play Wasteland on a hard drive (y／n)?
```

### A：載入硬碟

1. 如果你要將遊戲載入硬碟，請鍵入 Y，螢幕出現下列訊息，要求你輸入硬式磁碟機之代號，鍵入 C 即可。

```
Please enter the hard drive letter:
```

2. 接著出現下列的顯示卡選擇項目：

```
Select the appropriate graphics／monitor combination：
❶EGA／RGB
❷Tandy 16 color／RGB or Composite
❸CGA／RGB
❹CGA／Composite
```

　　請依你的電腦裝備加以選擇；EGA 顯示卡請選 1，而 MGA 顯示卡或 CGA 顯示卡則選 3。

3. 選擇完畢後，螢幕會再出現下列訊息，說明你的硬碟內至少須有 850000 byte 的空間才能載入，而且，此步驟還會在硬碟中建立一個 WASTELAN 之次目錄。

---

## wl28（補充說明書 第 3 頁）

```
WARNINGS：
You need at least 850,000 bytes of space free on your
hard drive to transfer Wasteland.
This setup will create a sub-directory called WASTE-
LAN.
Are these options correct (y／n)?
```

　　如果合乎此項要求，請鍵入 `Y` 鍵，便可直接將遊戲 MASTER DISK 1 載入硬碟。

4. 當 MASTER DISK 1 載入完畢後，請將 MASTER DISK2 置於 A 磁碟機中，並按 ENTER 鍵，將遊戲完全載入硬碟。

5. 接下來是耗時最久的 “圖形畫面設定” 部分，由電腦自動執行，請耐心地依電腦指示抽換磁片即可。

### B：轉錄於磁片中

1. 當出現 `Play Wasteland on a hard drive(y／n)?` 訊息時，鍵入 N，表示不將遊戲載入硬碟中。

2. 隨後立即出現如下的訊息，詢問你將使用幾部磁碟機進行轉錄。

```
Please enter the number of floppy disk drives you wish
to use when playing the game (1 or 2):
```

---

## wl29（補充說明書 第 4 頁）

3. `Setup Wasteland on drive (A or B)`：如果你有兩部磁碟機，可以鍵入 B，設定空白磁片欲放入的磁碟機為 B，以便轉錄工作更易進行；若只有一部磁碟機則鍵入 A。

4. 此步驟與 A 載入硬碟之第 2 項相同。

5. WARNINGS：

```
You will need three formatted disk labeled
'Program Disk','Scenario Disk 1', and 'Scenario Disk 2'
You will need exactly 362,496 bytes free for 5¼ disks
or 730,112 bytes free for 3½ disks.
Are these options correct (y／n)?
```

　　本畫面在於確定你是否已經準備好標名為 “Program Disk”、“Scenario Disk 1” 及 “Scenario Disk 2” 之三片已格式化的空白磁片。若已準備妥當，鍵入 `Y`，並依電腦指示，將所需之磁片置於指定之磁碟機中，即可進行轉錄。接著下來是耗時最久的部分：圖形畫面之設定，由電腦自動執行，請耐心地依電腦指示抽換磁片即可。

#### 二、啟動遊戲

　　若要使用滑鼠控制，請先將滑鼠接上。

##### 軟式磁碟機啟動

1. 以 DOS 開機，待 A＞出現後，取出 DOS 磁片，並將轉錄好的 “Program Disk” 置於 A 磁碟機，而將 “Scenario Disk 2” 置於 B 磁碟機

---

## wl29（補充說明書 第 5 頁）

中，關好機門。

2. 使用 MGA 顯示卡（單色螢幕），請鍵入 PLAY；若是使用 CGA 或 EGA 顯示卡（彩色螢幕），請鍵入 WL，並按 ENTER 鍵。

3. 隨後會出現本遊戲的片頭畫面，要求你將 “Scenario Disk 1” 置於 A 磁碟機中，並按 ENTER 鍵，以讀取所需之資料；依電腦所顯示的訊息執行，即可進入遊戲。

##### 硬式磁碟機啟動

1. 依一般之硬碟開機程序開機，待 C＞出現後，鍵入 `CD\WASTELAN`，並按 ENTER 鍵。

2. 使用 MGA 顯示卡（單色螢幕），請鍵入 PLAY；若是使用 CGA 或 EGA 顯示卡（彩色螢幕），請鍵入 WL，並按 ENTER 鍵。

3. 隨後即可出現遊戲的片頭畫面並進入遊戲。

##### 控制方法

A 鍵盤控制：（其它控制方法與手冊相同）

＊螢幕上出現選擇項時，直接按你所需項目的數字或第一個字母，即可執行該項之功能。

＊移動人物有下列兩種方法：

〔插圖：兩組方向鍵配置圖〕

(1)

```
        上
       ┌───┐
       │ I │
  ┌───┐├───┤┌───┐
左│ J ││ K ││ L │右
  └───┘└───┘└───┘
        下
```

(2)

```
        上
 ┌───┐┌───┐┌───┐
 │ 7 ││ 8↑││ 9 │
 ├───┤├───┤├───┤
左│4←││ 5 ││6→│右
 ├───┤├───┤├───┤
 │ 1 ││ 2↓││ 3 │
 └───┘└───┘└───┘
        下
```

---

## wl30（補充說明書 第 6 頁）

```
＊PgUp 鍵：訊息欄向上顯示。
  PgDn 鍵：訊息欄向下顯示。
  ESC  鍵：取消命令。
```

＊在作戰的狀況下

　空白鍵：立即顯示當時的作戰地圖。

　`↑` 鍵：增加作戰之速度。

　`↓` 鍵：減慢作戰之速度。

B 滑鼠控制：

＊所使用之滑鼠須為 Microsoft™ 或與其相容之滑鼠才能操作。

＊與一般滑鼠的操作方法相同，可利用滑鼠將畫面上的箭頭移到你所需的選擇項上，並按滑鼠上的控制鈕，即可執行該項之功能。

C 聚結功能的設定：

　　為了您操作上的方便，本遊戲增加了聚結功能（Macro function）之設定，可供你設定 10 個你常用的功能於 `F1`～`F10` 鍵中。你只要在有選擇項目出現的畫面下，同時按 `Ctrl` 鍵及 `F1`～`F10` 中欲設定的任一鍵，在畫面的左上角，會出現 REC.MAC 的字樣，此時，你可鍵入你所要設定的選擇項之第一個字母，再同時按 `Ctrl` 鍵及 `F1`～`F10` 中的任一鍵，即可完成設定此鍵的功能。假如，同時按 `Ctrl` 鍵及 `F5` 鍵後，螢幕左上角出現 REC.MAC05 的字樣，再按 `S` 鍵（表示 Save 儲存遊戲），接著同時按 `Ctrl` 鍵及 `F5` 鍵，即可將儲存遊戲的功能，設定在 `F5` 鍵上，以後，直接按 `F5` 即可進行遊戲之儲存步驟了。你可依以上介紹的設定方法，將常用的 10 個功能設定在 `F1`～`F10` 各鍵中。

> 註：`聚結功能` 為當年對 Macro function 的譯法，照原樣轉錄。
