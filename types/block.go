package types

import (
	"crypto/sha256"

	"github.com/mhg14/ChlockBane/crypto"

	"github.com/mhg14/ChlockBane/proto"
	pb "google.golang.org/protobuf/proto"
)

func SignBlock(privKey *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	return privKey.Sign(HashBlock(block))
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
	return hash[:] // converting an arrya to a slice and returning the slice
}
