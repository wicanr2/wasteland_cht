package assets

import "fmt"

// MSQ 區塊：資源定址、解密與三層地圖（docs/re/06、08、16、18、24）。

// 執行檔內的定址表（ds: 位移，docs/re/00-master-index.md §5.1）。
const (
	tblDirectory = 0xBEC9 // 資源目錄，逐 byte 到 0xFF；高 2 bits ＝ 哪個檔案
	tblTotalLen  = 0xBD86 // 區塊總長度
	tblReadLen   = 0xBD22 // 讀取量（＝ 交給 XOR 解密的長度）
	tblMapSize   = 0xBF1C // 地圖大小選擇：0x40 → 0x1800，其餘 → 0x600
)

// 記錄區標頭的欄位（自 P 起算，docs/spec/01 §2.4）。
const (
	hdrSize      = 0x5C
	hdrEncLen    = 0x00 // 加密長度 L，同時是字串表基址
	hdrNames     = 0x02 // 明文敵人名表位移
	hdrDim       = 0x2C // 地圖邊長 D
	hdrEncounter = 0x2F // 遭遇機率分母
	hdrTileset   = 0x30 // 圖磚組編號
	hdrOutside   = 0x33 // 地圖範圍外要畫的圖形編號
	hdrStepTime  = 0x34 // 走一步推進的時間（分鐘 × 256）
	hdrStepTick  = 0x36 // 走一步推進的刻
)

// Resource 是資源目錄裡的一筆。
type Resource struct {
	ID       int
	File     string // "game1" 或 "game2"
	Offset   int    // 在該檔案裡的位移
	TotalLen int
	ReadLen  int
	MapSize  int // 0x600 或 0x1800
}

// Resources 讀出資源目錄，回傳有 MSQ 區塊的那些。
//
// 同一個檔案裡的區塊首尾相接，位移是「前面所有區塊總長度的和」。
func (r *Rom) Resources() ([]Resource, error) {
	dir, err := r.dsOffset(tblDirectory)
	if err != nil {
		return nil, err
	}
	var kinds []byte
	for i := dir; i < len(r.image) && r.image[i] != 0xFF; i++ {
		kinds = append(kinds, r.image[i])
	}
	if len(kinds) == 0 {
		return nil, fmt.Errorf("資源目錄是空的（%#x）", tblDirectory)
	}

	total, err := r.wordTable(tblTotalLen, len(kinds))
	if err != nil {
		return nil, err
	}
	read, err := r.wordTable(tblReadLen, len(kinds))
	if err != nil {
		return nil, err
	}
	sizeSel, err := r.dsOffset(tblMapSize)
	if err != nil {
		return nil, err
	}

	cursor := map[string]int{"game1": 0, "game2": 0}
	var out []Resource
	for id, kind := range kinds {
		var file string
		switch kind & 0xC0 {
		case 0x80:
			file = "game1"
		case 0x40:
			file = "game2"
		default:
			continue // 目錄裡有沒有 MSQ 區塊的欄位
		}
		if total[id] == 0 {
			continue
		}
		mapSize := 0x600
		if r.image[sizeSel+id] == 0x40 {
			mapSize = 0x1800
		}
		out = append(out, Resource{
			ID:       id,
			File:     file,
			Offset:   cursor[file],
			TotalLen: int(total[id]),
			ReadLen:  int(read[id]),
			MapSize:  mapSize,
		})
		cursor[file] += int(total[id])
	}
	return out, nil
}

func (r *Rom) wordTable(dsOff, n int) ([]uint16, error) {
	at, err := r.dsOffset(dsOff)
	if err != nil {
		return nil, err
	}
	if at+n*2 > len(r.image) {
		return nil, fmt.Errorf("表 ds:%04Xh 的 %d 項超出映像", dsOff, n)
	}
	out := make([]uint16, n)
	for i := range out {
		out[i] = le16(r.image, at+i*2)
	}
	return out, nil
}

// decryptMSQ 是原版的 XOR 串流（docs/re/08）。
//
// ⚠ 只解到「記錄區標頭第一個 word」那麼長，不是整個區塊（docs/re/18 §2）。
// 多解的部分會被破壞成高熵資料，而症狀跟「這一段是壓縮的」一模一樣。
func decryptMSQ(raw []byte, checksum uint16, headerAt int) ([]byte, int, error) {
	out := make([]byte, len(raw))
	copy(out, raw)
	key := byte(checksum&0xFF) ^ byte(checksum>>8)

	first := headerAt + 2
	if first > len(out) {
		first = len(out)
	}
	for i := 0; i < first; i++ {
		out[i] = raw[i] ^ key
		key += 0x1F
	}
	if headerAt+2 > len(out) {
		return nil, 0, fmt.Errorf("區塊只有 %d bytes，放不下記錄區標頭（P＝%#x）", len(out), headerAt)
	}
	length := int(le16(out, headerAt))
	if length > len(out) {
		length = len(out)
	}
	for i := first; i < length; i++ {
		out[i] = raw[i] ^ key
		key += 0x1F
	}
	return out, length, nil
}

// Block 是一個解開的 MSQ 區塊。
type Block struct {
	Resource Resource
	Header   []byte // 記錄區標頭 0x5C bytes
	Dim      int    // 地圖邊長 D（實測只有 32 與 64）
	Tileset  int    // 標頭 +0x30
	Terrain  []byte // 第 1 層，已展開成一格一個 byte（值 0–15）
	Record   []byte // 第 2 層，一格一個 byte
	Graphic  []byte // 第 3 層（Huffman 尾段），一格一個 byte
	Strings  []string

	// Raw 是解密後的區塊本體（不含 6 bytes 標頭），未解區域原樣保留。
	// 存檔是改寫不是重建，這份不能丟。
	Raw []byte

	// EncLen 是加密長度，同時是字串表在 Raw 裡的位移。
	EncLen int
}

// Block 解開第 n 個 MSQ 資源（n 是 Resources() 的索引，不是資源編號）。
func (r *Rom) Block(n int) (*Block, error) {
	res, err := r.Resources()
	if err != nil {
		return nil, err
	}
	if n < 0 || n >= len(res) {
		return nil, fmt.Errorf("區塊編號 %d 超出範圍（共 %d 個）", n, len(res))
	}
	return r.blockFrom(res[n])
}

func (r *Rom) blockFrom(res Resource) (*Block, error) {
	data, err := r.File(res.File)
	if err != nil {
		return nil, err
	}
	if res.Offset+res.TotalLen > len(data) {
		return nil, fmt.Errorf("資源 %d 超出 %s 的長度", res.ID, res.File)
	}
	span := data[res.Offset : res.Offset+res.TotalLen]
	if len(span) < 6 {
		return nil, fmt.Errorf("資源 %d 太短", res.ID)
	}
	if magic := string(span[0:3]); magic != "msq" {
		return nil, fmt.Errorf("資源 %d 的 magic 是 %q，不是 msq", res.ID, magic)
	}
	checksum := le16(span, 4)

	body, encLen, err := decryptMSQ(span[6:res.ReadLen], checksum, res.MapSize)
	if err != nil {
		return nil, fmt.Errorf("資源 %d：%w", res.ID, err)
	}

	header := body[res.MapSize : res.MapSize+hdrSize]
	dim := int(header[hdrDim])
	if dim == 0 || dim*dim*3/2 != res.MapSize {
		return nil, fmt.Errorf("資源 %d 的邊長 %d 與地圖區長度 %#x 對不上", res.ID, dim, res.MapSize)
	}

	// 第 1 層是 4 bits 一格，偶數行取高 4 位（docs/re/24 §2.1）。
	terrain := make([]byte, dim*dim)
	for i := range terrain {
		b := body[i/2]
		if i%2 == 0 {
			terrain[i] = b >> 4
		} else {
			terrain[i] = b & 0x0F
		}
	}
	record := make([]byte, dim*dim)
	copy(record, body[dim*dim/2:res.MapSize])

	// 第 3 層在 Huffman 尾段，起點是區塊裡的「讀取量」位移。
	tail := span[res.ReadLen:]
	graphic, _, err := Decompress(tail, false)
	if err != nil {
		return nil, fmt.Errorf("資源 %d 的尾段：%w", res.ID, err)
	}
	if len(graphic) != dim*dim {
		return nil, fmt.Errorf("資源 %d 的尾段解出 %d bytes，應該是 %d（D²）",
			res.ID, len(graphic), dim*dim)
	}

	strings, err := decodeTable(body, encLen)
	if err != nil {
		return nil, fmt.Errorf("資源 %d 的字串表：%w", res.ID, err)
	}

	return &Block{
		Resource: res,
		Header:   header,
		Dim:      dim,
		Tileset:  int(header[hdrTileset]),
		Terrain:  terrain,
		Record:   record,
		Graphic:  graphic,
		Strings:  strings,
		Raw:      body,
		EncLen:   encLen,
	}, nil
}

// At 回傳 (x, y) 那一格的三層值。
func (b *Block) At(x, y int) (terrain, record, graphic byte, err error) {
	if x < 0 || y < 0 || x >= b.Dim || y >= b.Dim {
		return 0, 0, 0, fmt.Errorf("座標 (%d, %d) 超出 %d × %d 的地圖", x, y, b.Dim, b.Dim)
	}
	i := y*b.Dim + x
	return b.Terrain[i], b.Record[i], b.Graphic[i], nil
}

// OutsideGraphic 是地圖範圍外要畫的圖形編號（標頭 +0x33）。
func (b *Block) OutsideGraphic() byte { return b.Header[hdrOutside] }

// StepMinutes 是在這張地圖走一步推進的遊戲分鐘數（標頭 +0x34/+0x35 ÷ 256）。
func (b *Block) StepMinutes() float64 {
	return float64(le16(b.Header, hdrStepTime)) / 256
}

// StepTick 是走一步推進的刻（標頭 +0x36），週期性角色處理用。
func (b *Block) StepTick() byte { return b.Header[hdrStepTick] }
