package assets

import "fmt"

// Huffman 解壓（docs/re/10、docs/re/11）。
//
// 容器：
//
//	+0  4 bytes  解壓後大小
//	+4  3 bytes  'm' 's' 'q'（尾段沒有，載入器走跳過驗證的那條路徑）
//	+7  1 byte   磁碟編號
//	+8  …        位元流：前序編碼的 Huffman 樹，接著是編碼資料
//
// 位元順序 MSB first。樹在原版是 5 bytes 一個節點、上限 768 個。

const (
	huffMaxNodes = 0x300
	huffHeader   = 8
)

// bitReader 逐位元讀，MSB 先。pos 永遠指向「下一個要讀進來的 byte」，
// 解完之後它就是這個子區塊消耗掉的長度——串接下一個子區塊要用。
type bitReader struct {
	data []byte
	pos  int
	cur  byte
	mask byte
}

func newBitReader(data []byte, at int) (*bitReader, error) {
	if at >= len(data) {
		return nil, fmt.Errorf("位元流起點 %d 超出資料長度 %d", at, len(data))
	}
	return &bitReader{data: data, pos: at + 1, cur: data[at], mask: 0x80}, nil
}

func (b *bitReader) bit() (int, error) {
	if b.mask == 0 {
		if b.pos >= len(b.data) {
			return 0, fmt.Errorf("位元流在第 %d byte 就用完了", b.pos)
		}
		b.cur = b.data[b.pos]
		b.pos++
		b.mask = 0x80
	}
	v := 0
	if b.cur&b.mask != 0 {
		v = 1
	}
	b.mask >>= 1
	return v, nil
}

func (b *bitReader) byte() (byte, error) {
	var v byte
	for i := 0; i < 8; i++ {
		bit, err := b.bit()
		if err != nil {
			return 0, err
		}
		v = v<<1 | byte(bit)
	}
	return v, nil
}

// huffNode 對應原版 5 bytes 的節點（左 2、右 2、值 1）。葉節點的 left 是 0。
type huffNode struct {
	left, right int
	value       byte
}

type huffTree struct {
	nodes []huffNode
}

func (t *huffTree) alloc() (int, error) {
	if len(t.nodes) >= huffMaxNodes {
		return 0, fmt.Errorf("節點數超過原版上限 %d", huffMaxNodes)
	}
	t.nodes = append(t.nodes, huffNode{})
	return len(t.nodes) - 1, nil
}

// build 讀前序編碼的樹。
//
// ⚠ 兩次遞迴之間會多讀一個 bit（原版 sub_11C28 的 0x11C44）——
// 少讀那一個 bit 的症狀是「樹建得起來但解出來是雜訊」，很難從輸出看出來。
func (t *huffTree) build(r *bitReader, at int) error {
	leaf, err := r.bit()
	if err != nil {
		return err
	}
	if leaf == 1 {
		v, err := r.byte()
		if err != nil {
			return err
		}
		t.nodes[at].value = v
		return nil
	}
	left, err := t.alloc()
	if err != nil {
		return err
	}
	right, err := t.alloc()
	if err != nil {
		return err
	}
	t.nodes[at].left, t.nodes[at].right = left, right
	if err := t.build(r, left); err != nil {
		return err
	}
	if _, err := r.bit(); err != nil { // 分隔位元
		return err
	}
	return t.build(r, right)
}

// HuffInfo 是一次解壓的統計，串接子區塊時要用 Consumed。
type HuffInfo struct {
	DeclaredSize int
	Disk         byte
	TreeNodes    int
	Consumed     int // 這個子區塊用掉幾個 byte
}

// Decompress 解一個 Huffman 子區塊。
//
// verifyMagic=false 對應原版 sub_11AE8 的 AL=0 路徑：MSQ 區塊的尾段一樣有
// 8 bytes 標頭，但第 5–7 bytes 不是 msq，原版直接跳過驗證。
// **沒有 magic 不代表格式不同**，這件事只能從程式碼看出來。
func Decompress(data []byte, verifyMagic bool) ([]byte, HuffInfo, error) {
	var info HuffInfo
	if len(data) < huffHeader {
		return nil, info, fmt.Errorf("資料只有 %d bytes，連標頭都不夠", len(data))
	}
	size := int(le32(data, 0))
	if verifyMagic && string(data[4:7]) != "msq" {
		return nil, info, fmt.Errorf("magic 不是 msq：%q", data[4:7])
	}
	info.DeclaredSize = size
	info.Disk = data[7]

	r, err := newBitReader(data, huffHeader)
	if err != nil {
		return nil, info, err
	}
	tree := &huffTree{nodes: make([]huffNode, 1, 64)} // [0] 是根
	if err := tree.build(r, 0); err != nil {
		return nil, info, fmt.Errorf("建樹失敗：%w", err)
	}
	info.TreeNodes = len(tree.nodes) - 1

	out := make([]byte, 0, size)
	for len(out) < size {
		node := 0
		for tree.nodes[node].left != 0 {
			bit, err := r.bit()
			if err != nil {
				return nil, info, fmt.Errorf("解到第 %d／%d byte 時位元流用完：%w", len(out), size, err)
			}
			if bit == 1 {
				node = tree.nodes[node].right
			} else {
				node = tree.nodes[node].left
			}
		}
		out = append(out, tree.nodes[node].value)
	}
	info.Consumed = r.pos
	return out, info, nil
}

// SubBlock 是一個檔案裡的一個 Huffman 子區塊。
type SubBlock struct {
	Index    int
	Offset   int
	Size     int // 解壓後長度
	Consumed int
}

// SplitAll 把一個整檔切成子區塊。
//
// 子區塊首尾相接，下一個的起點就是上一個位元流消耗完的位置——這是實測出來的
// （allhtds1 第一塊消耗到 5122，而檔案裡下一個 msq 出現在 5126，差的 4 bytes
// 是大小前綴）。**不要用搜尋 magic 的方式切**，那會被資料裡剛好出現的 msq 騙。
func SplitAll(data []byte) ([]SubBlock, error) {
	var out []SubBlock
	pos := 0
	for pos+huffHeader <= len(data) && string(data[pos+4:pos+7]) == "msq" {
		blob, info, err := Decompress(data[pos:], true)
		if err != nil {
			return out, fmt.Errorf("第 %d 個子區塊（位移 %#x）：%w", len(out), pos, err)
		}
		out = append(out, SubBlock{
			Index:    len(out),
			Offset:   pos,
			Size:     len(blob),
			Consumed: info.Consumed,
		})
		pos += info.Consumed
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("開頭不是 Huffman 子區塊")
	}
	return out, nil
}

// DecompressAt 解第 n 個子區塊的內容。
func DecompressAt(data []byte, blocks []SubBlock, n int) ([]byte, error) {
	if n < 0 || n >= len(blocks) {
		return nil, fmt.Errorf("子區塊編號 %d 超出範圍（共 %d 個）", n, len(blocks))
	}
	blob, _, err := Decompress(data[blocks[n].Offset:], true)
	return blob, err
}
