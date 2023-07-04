package types

import (
	"crypto/sha256"

	"github.com/mhg14/ChlockBane/crypto"

	"github.com/mhg14/ChlockBane/proto"
	pb "google.golang.org/protobuf/proto"
)

func SignBlock(privKey *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
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
	if len(b.PublicKey) != crypto.PublicKeyLen {
		return false
	}

	if len(b.Signature) != crypto.SigLen {
		return false
	}
	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey)
	hash := HashBlock(b)
	return sig.Verify(pubKey, hash)
}