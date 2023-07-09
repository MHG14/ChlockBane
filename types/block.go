package types

import (
	"bytes"
	"crypto/sha256"
	"log"

	"github.com/cbergoon/merkletree"
	"github.com/mhg14/ChlockBane/crypto"

	"github.com/mhg14/ChlockBane/proto"
	pb "google.golang.org/protobuf/proto"
)

type TxHash struct {
	hash []byte
}

func NewtTxHash(hash []byte) TxHash {
	return TxHash{
		hash: hash,
	}
}

func (h TxHash) CalculateHash() ([]byte, error) {
	return h.hash, nil
}

func (h TxHash) Equals(other merkletree.Content) (bool, error) {
	equals := bytes.Equal(h.hash, other.(TxHash).hash)
	return equals, nil
}

func SignBlock(privKey *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	if len(block.Transactions) > 0 {
		tree, err := GetMerkleTree(block)
		if err != nil {
			panic(err)
		}
		block.Header.RootHash = tree.MerkleRoot()
	}

	hash := HashBlock(block)
	signature := privKey.Sign(hash)
	block.PublicKey = privKey.Public().Bytes()
	block.Signature = signature.Bytes()

	return signature
}

// This function returns a SHA256 of only the block header
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:] // converting an array to a slice and returning the slice
}

func VerifyBlock(b *proto.Block) bool {
	if len(b.Transactions) > 0 {
		if !VerifyRootHash(b) {
			log.Println("INVALID ROOT HASH")
			return false
		}
	}

	if len(b.PublicKey) != crypto.PublicKeyLen {
		log.Println("INVALID PUBLIC KEY LENGTH")
		return false
	}

	if len(b.Signature) != crypto.SigLen {
		log.Println("INVALID SIGNATURE LENGTH")
		return false
	}
	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey)
	hash := HashBlock(b)
	return sig.Verify(pubKey, hash)
}

func VerifyRootHash(b *proto.Block) bool {
	merkleTree, err := GetMerkleTree(b)
	if err != nil {
		return false
	}

	valid, err := merkleTree.VerifyTree()
	if err != nil {
		return false
	}

	if !valid {
		return false
	}

	return bytes.Equal(b.Header.RootHash, merkleTree.MerkleRoot())
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	list := make([]merkletree.Content, len(b.Transactions))

	for i := 0; i < len(b.Transactions); i++ {
		list[i] = NewtTxHash(HashTransaction(b.Transactions[i]))
	}

	// create a new merkle tree from the list of content
	t, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}

	return t, nil
}
