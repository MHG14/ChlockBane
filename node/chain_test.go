package node

import (
	"encoding/hex"
	"testing"

	"github.com/mhg14/ChlockBane/crypto"
	"github.com/mhg14/ChlockBane/proto"
	"github.com/mhg14/ChlockBane/types"
	"github.com/mhg14/ChlockBane/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	for i := 0; i < 100; i++ {
		b := randomBlock(t, chain)
		require.Nil(t, chain.AddBlock(b))
		require.Equal(t, chain.Height(), i+1)
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	for i := 0; i < 100; i++ {
		block := randomBlock(t, chain)
		blockHash := types.HashBlock(block)

		require.Nil(t, chain.AddBlock(block))

		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i + 1)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlockByHeight)
	}
}

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	assert.Equal(t, 0, chain.Height())
	_, err := chain.GetBlockByHeight(0)
	assert.Nil(t, err)
}

func randomBlock(t *testing.T, chain *Chain) *proto.Block {
	b := util.RandomBlock()
	prevBlock, err := chain.GetBlockByHeight(chain.Height())
	require.Nil(t, err)
	b.Header.PrevHash = types.HashBlock(prevBlock)

	privKey := crypto.GeneratePrivateKey()
	types.SignBlock(privKey, b)
	return b
}

func TestAddBlockWithTx(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromSeedString(godSeed)
		recipient = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)

	fetchedTransaction, err := chain.txStore.Get("8f26b010c9db9857962c5faaaf1aa629506cc1646a129ce92f525c1776bb8b78")
	assert.Nil(t, err)

	inputs := []*proto.TxInput{
		{
			PrevTxHash:   types.HashTransaction(fetchedTransaction),
			PrevOutIndex: 0,
			PublicKey:    privKey.Public().Bytes(),
		},
	}
	outputs := []*proto.TxOutput{
		{Amount: 100, Address: recipient},
		{Amount: 900, Address: privKey.Public().Address().Bytes()},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}
	txHash := hex.EncodeToString(types.HashTransaction(tx))

	block.Transactions = append(block.Transactions, tx)
	require.Nil(t, chain.AddBlock(block))

	fetchedTx, err := chain.txStore.Get(txHash)
	assert.Nil(t, err)
	assert.Equal(t, tx, fetchedTx)
}
