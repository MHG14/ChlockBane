package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/mhg14/ChlockBane/proto"
	"github.com/mhg14/ChlockBane/types"
)

type BlockStorer interface {
	Put(*proto.Block) error
	Get(string) (*proto.Block, error)
}

type MemoryBlockStore struct {
	lock sync.RWMutex
	blocks map[string]*proto.Block
}


func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{
		blocks: make(map[string]*proto.Block),
	}
}

func (s *MemoryBlockStore) Get(hash string) (*proto.Block, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	block, ok := s.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash [%s] does not exist", hash)
	}

	return block, nil
}


func (s *MemoryBlockStore) Put(block *proto.Block) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	hash := hex.EncodeToString(types.HashBlock(block))
	s.blocks[hash] = block
	return nil
}
