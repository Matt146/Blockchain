package wallet

import (
	"Blockchain/blockchain"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"math/big"
)

// Wallet - wallet data
type Wallet struct {
	KeyPair *ecdsa.PrivateKey
}

// MakeWallet - Generate a wallet and return it
func MakeWallet() (Wallet, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
	w := Wallet{KeyPair: privKey}
	if err != nil {
		return w, err
	}

	return w, nil
}

// SignTransaction - Signs the transaction using a private key,
// sets the signature of the transaction to the one computed
// in the function, and returns the signature of the transaction
// NOTE: ALWAYS CHECK FOR ERRORS ON THIS FUNCTION. OTHERWISE,
// USING THE VALUES IT LEAVES WILL LEAD TO A SEGFAULT
func (w *Wallet) SignTransaction(t *blockchain.Transaction) (*big.Int, *big.Int, error) {
	// Create a signature
	r, s, err := ecdsa.Sign(crand.Reader, w.KeyPair, t.HashTransaction())
	if err != nil {
		return nil, nil, err
	}

	return r, s, nil
}
