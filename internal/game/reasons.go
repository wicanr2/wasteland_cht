package game

// 失敗原因是**代碼，不是給玩家看的文字**。
//
// 規則層不認識畫面，也不認識語言：呈現層拿代碼去查原版字串表或翻譯目錄
// （`internal/play` 的 `setNoteReason`）。
//
// ⚠ 這裡一度直接回中文字面（`"錢不夠"`）。規則層的字串是 UTF-8，
// 而設施畫面走的是倚天字型那一層，於是那三個字被當成 Big5 逐 byte 畫出來，
// 螢幕上是一行希臘字母。**編得過、測得過、畫出來是亂碼**——
// 規則層產生玩家看得到的字，這一類錯就會一直有。
const (
	ReasonNoMoney        = "no-money"
	ReasonNoHealNeeded   = "no-heal-needed"
	ReasonNoSuchDisease  = "no-such-disease"
	ReasonInventoryFull  = "inventory-full"
	ReasonLowIQ          = "low-iq"
	ReasonNoSkillPoints  = "no-skill-points"
	ReasonSkillSlotsFull = "skill-slots-full"
)
