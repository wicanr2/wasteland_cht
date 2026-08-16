# 第六章　COMMAND SUMMARY（指令總覽）

原文章節：COMMAND SUMMARY 全節（Game Play 起至 Macro Functions）。

英文照原樣保留（含原文的錯字與標點），中文譯文接在每一段之後或另立一欄。譯名依
[`translations/glossary.md`](../../translations/glossary.md)。

## Game Play（遊戲進行）

IMPORTANT: Wasteland is a dynamic game and it’s very important that you understand how it saves and keeps track of the fame. The game changes as you play and updates changes to the disk permanently. If you take an item, it won’t be resurrected just because you leave and return.

重要：《荒野遊俠》是一個會變的遊戲，你一定要弄懂它怎麼存檔、怎麼記錄進度。遊戲隨著你玩而改變，而且把改變**永久寫進磁片**。你拿走一件東西，它不會因為你離開再回來就又長出來。

The game takes place in many locations. As you explore, you’ll often be asked “Enter New Location (Y/N)?” If you answer “yes” the game will save any changes to that location, your party’s status, and become your new saved game locations. If you switch to another location to view a disbanded party, the status’s of all parties are saved. You should save the game before ending each session of play. Then when you go to play again you can pick u where you left off. However, if your computer will search for the last place it saved. This way, you’re unlikely to lose any important recent acquisitions. What can you do if a character dies? DO NOT ENTER A NEW LOCATION OR SAVE THE GAME! Turn off your computer and reboot, and your character will live again, but without anything they acquired since you last saved. If all the characters die in the midst of general carnage and mayhem, your computer will state the obvious: “Your life in Wasteland is over.” Don’t get depressed, just reboot and the game will return your characters to either the last time you saved or the last time the game map changed. (This assumes, of course, that three’s someplace to return to.)

遊戲發生在許多地點。探索途中常會被問「Enter New Location (Y/N)?」。答「yes」的話，遊戲會把那個地點的改動與隊伍狀態存下來，成為新的存檔位置。切到另一個地點去看分出去的隊伍時，所有隊伍的狀態也都會存下來。每次玩完之前應該存一次檔，下次才接得回原處。**角色死了怎麼辦？不要進入新地點，也不要存檔！** 關機重開，你的角色就會活過來，只是身上少了上次存檔之後拿到的所有東西。要是在一片殺戮與混亂之中全部角色都死光了，電腦會告訴你一件很明白的事：**「Your life in Wasteland is over.」**（你在荒野的人生結束了。）別沮喪，重開機就好，遊戲會把角色帶回上一次存檔、或上一次地圖發生變動的時候。（當然，前提是還有地方回得去。）

> 這一段講的機制在 DOS 版的執行檔裡讀出來了：全隊倒下但還救得回來時**自動切到下一支隊伍**；每個人都到底時進**死亡畫面**——地點名換成 `Grim Reaper`、換一張圖、印出 `Your life has ended in The Wasteland.`。手冊寫的是 Apple II 版的措辭，DOS 版的字串不同、機制相同（`docs/re/99`）。

## Time and Distance（時間與距離）

> 與第五章的 `TIME AND DISTANCE` 是同一主題的另一版敘述，原文兩處都有，照原樣各留一份。

Wasteland’s maps vary in scale. The desert map contains the city maps which in turn contain maps of buildings and underground locations. In combat, distances may seem a bit off for the map you’re on, but these are tactical distances valid for combat only.

《荒野遊俠》的地圖比例尺各不相同。沙漠地圖裡包著城鎮地圖，城鎮地圖裡又包著建築物與地下場所的地圖。戰鬥中的距離看起來可能與所在地圖對不太上，但那是只在戰鬥裡有效的戰術距離。

Because the maps differ in scale, time passes differently on them. A single keystroke will move you one space in both the desert and in a building, but the amount of time each move takes is different. Time passes more quickly during overland travel, which the game takes into account for healing and deterioration purposes. And remember that time passes for both the main party and disbanded characters. If you send a disbanded character off to find a doctor for an injured comrade, that comrade will keep on bleeding.

因為比例尺不同，各張地圖上時間流逝的速度也不同。在沙漠與在建築物裡，按一次鍵都是走一格，但這一步花掉的時間不一樣。在野外趕路時時間過得比較快，而遊戲在計算療癒與惡化時會把這一點算進去。還要記得：時間對主隊與分出去的角色**同時**在走。你分一個人出去替受傷的同伴找醫生，那個同伴會一直流血下去。

If you want time to pass without moving your party, press <ESC> or place the mouse icon directly on your party and press the mouse button. If you wish time to pass more quickly, hold down the <ESC> key or keep the mouse button depressed.

想讓時間過去但隊伍不動，按 <ESC>，或把滑鼠游標直接放在隊伍上按下滑鼠鍵。想讓時間過得快一點，就按住 <ESC> 或按住滑鼠鍵不放。

## Reviewing Messages（回看訊息）

Wasteland involves a great deal of text. This text includes descriptions of your surroundings, descriptions of non-player characters, clues and references to the Wasteland paragraph section included in this manual. If you wish to refer back to or review a previous message, press the Pg Up key and keep it depressed until the desired message appears. To return to the most recent message, press the Pg Down key and keep it depressed until that message reappears.

《荒野遊俠》有大量文字：周遭環境的描述、非玩家角色的描述、線索，以及指向本說明書劇本段落的編號。想回頭看前面的訊息，按住 Pg Up 直到要找的那一則出現；要回到最新的一則，按住 Pg Dn 直到它重新出現。

## Selecting Options（選項怎麼選）

Whenever you need to select an option, press the first letter in that option (unEquip if the exception; in this case press E) or click on it with your mouse.

要選一個選項時，按下該選項的第一個字母（unEquip 是例外，這一個按 E），或用滑鼠點它。

Whenever you need to select an item, skill or attribute from a list, press its number or click on it with your mouse. To scroll through a list use the up and down arrows, the right or left arrow, the I key to scroll up or the K key to scroll down, or use the mouse to click on the next option or click on the up or down arrows on the right side of the option window.

要從清單裡選物品、技能或屬性時，按它的編號或用滑鼠點它。捲動清單可以用上下方向鍵、左右方向鍵，或用 I 往上、K 往下，也可以用滑鼠點下一個選項、或點選項視窗右側的上下箭頭。

## Movement Commands（移動指令）

There are three ways to move your party: Use the cursor keys, the mouse, or type I to move up, J to move left, K to move down or L to more right. When you use a mouse, a directional arrow will appear on the screen pointing forward, left, right or backward. Move the mouse in the direction you want to go until the directional arrow points in that direction. Then hold down the mouse button to more in that direction. The Spacebar toggles the view of the party roster on and off.

移動隊伍有三種辦法：方向鍵、滑鼠，或者按 I 上、J 左、K 下、L 右。用滑鼠時螢幕上會出現一個指向前、左、右、後的方向箭頭；把滑鼠往你要去的方向移動，直到箭頭指向那邊，再按住滑鼠鍵往那個方向走。空白鍵切換隊伍名單的顯示與隱藏。

## Ranger Center（遊俠中心）

The following options appear at the bottom of the screen when you’re at Ranger Center.

在遊俠中心時，畫面底部會出現以下選項。

| 選項（原文） | 說明（原文） | 說明（繁中） |
|---|---|---|
| `Create` | Creates a character | 建立一個角色 |
| `Delete` | Deletes a character | 刪掉一個角色 |
| `Play` | Begins Play outside Ranger Center | 離開遊俠中心，開始遊戲 |

## Non-Combat Commands（非戰鬥指令）

Except during combat, you can use the following commands by pressing the first letter of the command or clicking on it with your mouse.

除了戰鬥中以外，按下指令的第一個字母或用滑鼠點它，就可以使用下列指令。

| 按鍵／指令（原文） | 說明（原文） | 說明（繁中） |
|---|---|---|
| `Use` | Use a skill, item or attribute. | 使用技能、物品或屬性。 |
| `Enc` | Simulate an Encounter. This calls up combat commands, which you can use to initiate combat or use the Hire command to hire a non-player character into your party. | 模擬一場遭遇。它叫出戰鬥指令，你可以拿來主動開打，或用 Hire 把非玩家角色雇進隊伍。 |
| `Order` | Establish a new party marching Order. | 重新排定隊伍的行進順序。 |
| `Disband` | Disband the party into two or more groups. This command can also be used to permanently dismiss a Non-Player Character from your party. | 把隊伍拆成兩支以上。這個指令也可以用來把非玩家角色永久遣散。 |
| `View` | Alternate the View between two or more groups. | 在兩支以上的隊伍之間切換檢視。 |
| `Save` | Save the game. When you use the Save command, the computer will ask “Save Game(Y/N)?” If you answer “yes” the computer will save the game at that point and ask “Quite Game (Y/N)?” If you answer “yes” the computer will return you to the DOS screen and if you answer “no” the computer will continue the game. If you answer “no” to “Save Game(Y/N)?” the computer will still ask “Quit Game(Y/N?)” If you answer “no” the computer will continue the game. If you answer “yes” the computer will return you to the DOS screen, and the next time you reboot the game, it will start at the last point you saved. | 存檔。用 Save 時電腦會問「Save Game(Y/N)?」。答「yes」就在該處存檔，接著問「Quit Game (Y/N)?」；答「yes」回到 DOS 畫面，答「no」則繼續遊戲。「Save Game(Y/N)?」答「no」的話，電腦一樣會問「Quit Game(Y/N?)」：答「no」繼續遊戲，答「yes」回到 DOS 畫面，下次重開遊戲時會從你最後一次存檔的地方開始。 |
| `Radio` | Radio Ranger Center to see if any party members have earned promotion. | 用無線電呼叫遊俠中心，看看隊上有沒有人夠格升級。 |
| `Print` | Prints party information when the roster is displayed. | 名單顯示中時，把隊伍資料列印出來。 |
| `<SHIFT>-#` | Call up the Use command for a specific character. | 直接對某個角色叫出 Use 指令。 |
| `<CONTROL>R` | Reorder items and skills for a selected character when those menus are displayed. | 那兩個選單顯示中時，重排選定角色的物品與技能順序。 |
| `PgUp & Pg Dn` | Scrolls through the messages at the bottom of the screen. | 捲動畫面底部的訊息。 |

## Combat Commands（戰鬥指令）

Note: Some weapons have a limited range in combat situations. Contact weapons, such as knives, axes, fists, etc., are ineffective against opponents more than 14 feet away. Attacking opponents more than 14 feet away requires projectile weapons, such as throwing knives, pistols, rifles, etc.

注意：有些武器在戰鬥中射程有限。接觸類武器（刀、斧、拳頭等）對十四呎以外的對手無效；要打十四呎以外的對手，得用投射類武器（飛刀、手槍、步槍等）。

When you engage in battle, choose from the following options by pressing the command’s first letter or clicking on the command with your mouse.

開打之後，按指令的第一個字母或用滑鼠點它，從下列選項中挑一個。

| 按鍵／指令（原文） | 說明（原文） | 說明（繁中） |
|---|---|---|
| `Run` | Move party or individual character one space. | 讓全隊或單一角色移動一格。 |
| `Use` | Use a skill, item or attribute. | 使用技能、物品或屬性。 |
| `Hire` | Hire a Non-Player Character to join your party. | 雇用一個非玩家角色加入隊伍。 |
| `Evade` | Evade an enemy. | 迴避敵人。 |
| `Attack` | Attack an enemy. | 攻擊敵人。 |
| `Weapon` | Change Weapons. | 換武器。 |
| `Load/Unjam` | Load and/or Unjam a weapon. | 裝填武器，或排除卡彈。 |
| `<SPACEBAR>` | Show map of immediate area during combat. | 戰鬥中顯示周遭地圖。 |
| `<CONTROL>A` | Show list of enemy groups and their distance from the party. This will only work with player characters, not hired NPC’s, and only when your foes are within range of your weapons. | 顯示敵方各組與隊伍的距離。只對玩家角色有效，雇來的 NPC 不行；而且只有敵人落在你武器射程內時才有作用。 |
| `<ESC>` | Cancels commands. | 取消指令。 |

To speed the combat scrolling rate, press the up arrow key on the keyboard or click on the “fast” command on the screen with the mouse. To make it slower, press the down arrow key on the keyboard or click on the “slow” command on the screen with the mouse.

要加快戰鬥訊息的捲動，按鍵盤上的上方向鍵，或用滑鼠點畫面上的「fast」；要放慢就按下方向鍵，或點「slow」。

## Viewing Characters（檢視角色）

Enter a character’s number to view their statistics. The options you can use in this mode are:

輸入角色編號就能看他的數值。這個模式下可以用的選項是：

### From the first screen（第一頁）

This screen shows a character’s attributes:

這一頁顯示角色的屬性：

| 選項（原文） | 說明（原文） | 說明（繁中） |
|---|---|---|
| `Pool` | Pool all the party’s cash and give it to the character you are viewing. | 把全隊的錢集中起來交給你正在看的這個角色。 |
| `Div Cash` | Divide cash evenly among the party. | 把錢平分給全隊。 |
| `<ESC>` | Cancels commands. | 取消指令。 |

(Press <enter> to go to the next screen.)

（按 <enter> 到下一頁。）

### From the second screen（第二頁）

This screen shows what items the character has. Enter an item number and the following options will appear:

這一頁顯示角色身上有什麼。輸入物品編號之後會出現以下選項：

| 選項（原文） | 說明（原文） | 說明（繁中） |
|---|---|---|
| `Reload` | Reload weapon, (Only appears if you choose an ammo clip for the currently Equipped weapon.) | 裝填武器。（只有在你選了與目前裝備的武器相符的彈匣時才會出現。） |
| `Unjam` | Unjam weapon (Only appears if your currently equipped weapon is jammed.) | 排除卡彈。（只有在目前裝備的武器卡住時才會出現。） |
| `Drop` | Drop an item. | 丟掉一件物品。 |
| `Trade` | Trade an item. | 把一件物品交給隊友。 |
| `Equip` | Equip or unequip an item. | 裝備或卸下一件物品。 |
| `<CONTROL>-R` | Reorder items. | 重排物品順序。 |
| `<ESC>` | Cancels commands. | 取消指令。 |

When prompted Y/N, press Y or <enter> to accept the option.

出現 Y/N 的提示時，按 Y 或 <enter> 表示接受。

(Press <enter> to go to the next screen.)

（按 <enter> 到下一頁。）

### From the third screen（第三頁）

This screen shows the character’s skills.

這一頁顯示角色的技能。

| 選項（原文） | 說明（原文） | 說明（繁中） |
|---|---|---|
| `<CONTROL>-R` | Reorder skills. | 重排技能順序。 |
| `<ESC>` | Cancels commands. | 取消指令。 |

## Macro Functions（巨集功能）

Macro functions condense the several key strokes needed to give certain commands into one key stroke. To create a macro function, press <control> and any one of the function keys, F1 to F10, simultaneously. A message, REC.MAC. (with a number from 01 to 10 corresponding to the number of the function key you are pressing), will appear in the upper left corner of the screen; when it does, release the <control> and the function key again, the message in the upper left corner of the screen will vanish and the macro function will have been created. Pressing the appropriate function key thereafter will repeat the entire command or series of commands. (Example: If you want time to pass more quickly, press <control> and F1 and then release them when REC.MAC.01 appears in the upper left corner of the screen. Now press <ESC> several times, and then press <control> and F1. Every subsequent time you press F1, time will pass as if you had pressed <ESC> several times. A macro function can be erased by pressing and holding down <control> and pressing the appropriate function key twice.

巨集功能把某些指令原本要按的好幾個鍵壓縮成一個鍵。要錄一個巨集，同時按下 <control> 與 F1 到 F10 其中一個功能鍵。畫面左上角會出現 `REC.MAC.` 加上一個 01 到 10 的號碼（對應你按的那個功能鍵）；看到之後放開兩個鍵，開始按你要錄的那串鍵，再按一次 <control> 加同一個功能鍵，左上角的訊息就會消失，巨集錄好了。之後按那個功能鍵，就會重播整串指令。（例如：想讓時間過得快一點，先按 <control> 與 F1，等左上角出現 `REC.MAC.01` 之後放開；接著按幾次 <ESC>，再按 <control> 與 F1。往後每次按 F1，時間都會像你連按了好幾次 <ESC> 一樣往前走。）要刪掉一個巨集，按住 <control> 再按那個功能鍵兩次。

---

來源：`workplace/orig/wastland/manual.txt`（53,322 bytes，CP1252／CRLF，SHA-256 `4b222c7dc22229bffce83455989a1d164182f40396ffcf54a9445c09a2bc342e`）
