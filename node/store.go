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

type TXStorer interface {
	Put(*proto.Transaction) error
	Get(string) (*proto.Transaction, error)
}

type UTXOStorer interface {
	Put(string, *UTXO) error
	Get(string) (*UTXO, error)
}

type MemoryTXStore struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

type MemoryUTXOStore struct {
	lock sync.RWMutex
	data map[string]*UTXO
}

func NewMemoryUTXOStore() *MemoryUTXOStore {
	return &MemoryUTXOStore{
		data: make(map[string]*UTXO),
	}
}

func (s *MemoryUTXOStore) Get(hash string) (utxo *UTXO, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	utxo, ok := s.data[hash]
	if !ok {
		return nil, fmt.Errorf("could not fin UTXO with hash %s", hash)
	}

	return utxo, nil
}

func (s *MemoryUTXOStore) Put(key string, utxo *UTXO) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = utxo
	return nil
}

func NewMemoryTXStore() *MemoryTXStore {
	return &MemoryTXStore{
		txx: make(map[string]*proto.Transaction),
	}
}

func (s *MemoryTXStore) Get(hash string) (*proto.Transaction, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	tx, ok := s.txx[hash]
	if !ok {
		return nil, fmt.Errorf("couldn't find tx with hash %s", hash)
	}
	return tx, nil
}

func (s *MemoryTXStore) Put(tx *proto.Transaction) error {
	hash := hex.EncodeToString(types.HashTransaction(tx))
	s.lock.Lock()
	defer s.lock.Unlock()
	s.txx[hash] = tx
	return nil
}

type MemoryBlockStore struct {
	lock   sync.RWMutex
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
