package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()

	assert.Equal(t, privateKeyLen, len(privKey.Bytes()))

	msg := []byte("A secret message")

	sig := privKey.Sign(msg)

	assert.True(t, sig.Verify(pubKey, msg))

	// Test with invalid message
	assert.False(t, sig.Verify(pubKey, []byte("another message")))

	// Test with invalid publicKey
	anotherPublicKey := GeneratePrivateKey().Public()
	assert.False(t, sig.Verify(anotherPublicKey, msg))
}

func PublicKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	address := pubKey.Address()

	assert.Equal(t, addressLen, len(address.Bytes()))
	fmt.Println(address)
}

func TestNewPrivateKeyFromString(t *testing.T) {
	var (
		seed       = "58618bc746c4c3b20211878cdae7a47041f7f031f4eadefdc399d25700939b95"
		privKey    = NewPrivateKeyFromString(seed)
		addressStr = "a808d31c6c3f810be84d70da3b55d6d7c4505aa1"
	)

	assert.Equal(t, privateKeyLen, len(privKey.Bytes()))
	address := privKey.Public().Address()
	assert.Equal(t, addressStr, address.String())
}
