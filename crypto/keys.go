package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
)

const (
	PrivateKeyLen = 64
	PublicKeyLen  = 32
	SeedLen       = 32
	AddressLen    = 20
	SigLen        = 64
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func NewPrivateKeyFromString(s string) *PrivateKey {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(b)
}

func NewPrivateKeyFromSeedString(seed string) *PrivateKey {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		panic(err)
	}

	return NewPrivateKeyFromSeed(seedBytes)
}

func NewPrivateKeyFromSeed(seed []byte) *PrivateKey {
	if len(seed) != SeedLen {
		panic("invalid seed length, must be 32")
	}

	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, SeedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func (p *PrivateKey) Bytes() []byte {
	return p.key
}

func (p *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		value: ed25519.Sign(p.key, msg),
	}
}

func (p *PrivateKey) Public() *PublicKey {
	b := make([]byte, PublicKeyLen)
	copy(b, p.key[32:])

	return &PublicKey{
		key: b,
	}
}

type PublicKey struct {
	key ed25519.PublicKey
}

func PublicKeyFromBytes(b []byte) *PublicKey {
	if len(b) != PublicKeyLen {
		panic("invalid public key bytes length. should be 32")
	}
	return &PublicKey{
		key: ed25519.PublicKey(b),
	}
}

func (p *PublicKey) Address() Address {
	return Address{
		value: p.key[len(p.key)-AddressLen:],
	}
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

type Signature struct {
	value []byte
}

func SignatureFromBytes(b []byte) *Signature {
	if len(b) != SigLen {
		panic("invalid signature bytes length. should be 64")
	}
	return &Signature{
		value: b,
	}
}

func (s *Signature) Bytes() []byte {
	return s.value
}

func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

type Address struct {
	value []byte
}

func (a Address) Bytes() []byte {
	return a.value
}

func (a Address) String() string {
	return hex.EncodeToString(a.value)
}
