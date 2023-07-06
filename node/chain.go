package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/mhg14/ChlockBane/crypto"
	"github.com/mhg14/ChlockBane/proto"
	"github.com/mhg14/ChlockBane/types"
)

const godSeed = "6bc49ae98a0f9a9df49427788eb7c73f30299165035c040ab8b4ef56c97b2480"

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	txStore    TXStorer
	blockStore BlockStorer
	headers    *HeaderList
	utxoStore  UTXOStorer
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

func NewChain(bs BlockStorer, txStore TXStorer) *Chain {
	chain := &Chain{
		blockStore: bs,
		txStore:    txStore,
		utxoStore:  NewMemoryUTXOStore(),
		headers:    NewHeaderList(),
	}

	chain.addBlock(createGenesisBlock())
	return chain
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}

	return c.addBlock(b)

}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) addBlock(b *proto.Block) error {
	c.headers.Add(b.Header)

	for _, tx := range b.Transactions {
		if err := c.txStore.Put(tx); err != nil {
			return err
		}

		hash := hex.EncodeToString(types.HashTransaction(tx))

		for it, output := range tx.Outputs {
			utxo := &UTXO{
				Hash:     hex.EncodeToString(types.HashTransaction(tx)),
				Amount:   output.Amount,
				OutIndex: it,
				Spent:    false,
			}

			address := crypto.AddressFromBytes(output.Address)
			key := fmt.Sprintf("%s_%s", address, hash)

			if err := c.utxoStore.Put(key, utxo); err != nil {
				return err
			}
		}

	}

	return c.blockStore.Put(b)
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

func (c *Chain) ValidateBlock(b *proto.Block) error {
	// Validate the signature of the block
	if !types.VerifyBlock(b) {
		return fmt.Errorf("invalid block signature")
	}

	// Validate if the prevHash is the actual hash of the current block
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}

	hash := types.HashBlock(currentBlock)

	if !bytes.Equal(hash, b.Header.PrevHash) {
		return fmt.Errorf("invalid previous block hash")
	}

	for _, tx := range b.Transactions {
		if !types.VerifyTransaction(tx) {
			return fmt.Errorf("invalid tx signature")
		}
	}
	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.NewPrivateKeyFromSeedString(godSeed)
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{},
		Outputs: []*proto.TxOutput{{
			Amount:  1000,
			Address: privKey.Public().Address().Bytes(),
		}},
	}

	block.Transactions = append(block.Transactions, tx)

	types.SignBlock(privKey, block)

	return block
}
