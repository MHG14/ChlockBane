package node

import (
	"encoding/hex"
	"fmt"

	"github.com/mhg14/ChlockBane/proto"
	"github.com/mhg14/ChlockBane/types"
)

type Chain struct {
	blockStore BlockStorer
	headers    *HeaderList
}

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: []*proto.Header{},
	}
}

func (list *HeaderList) Add(h *proto.Header) {
	list.headers = append(list.headers, h)
}

func (list *HeaderList) Len() int {
	return len(list.headers)
}

func (list *HeaderList) Height() int {
	return list.Len() - 1
}

func (list *HeaderList) Get(index int) *proto.Header {
	if index > list.Height() {
		panic("index too high")
	}
	return list.headers[index]
}

func NewChain(bs BlockStorer) *Chain {
	return &Chain{
		blockStore: bs,
		headers:    NewHeaderList(),
	}
}

func (c *Chain) AddBlock(b *proto.Block) error {
	c.headers.Add(b.Header)
	return c.blockStore.Put(b)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if c.Height() < height {
		return nil, fmt.Errorf("given height %d too high, current chain height is %d", height, c.Height())
	}
	header := c.headers.Get(height)
	hash := types.HashHeader(header)
	return c.GetBlockByHash(hash)
}

func (c *Chain) Height() int {
	return c.headers.Height()
}
