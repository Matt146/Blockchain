package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
	"strconv"
)

// Transaction - This struct contains the necessary fields for each transaction
// on the network
type Transaction struct {
	Version    uint32   `json:"Version"`
	XInput     *big.Int `json:"XInput"`
	YInput     *big.Int `json:"YInput"`
	XOutput    *big.Int `json:"XOutput"`
	YOutput    *big.Int `json:"YOutput"`
	Amount     float64  `json:"Amount"`
	Timestamp  uint64   `json:"Timestamp"`
	Data       []byte   `json:"Data"`
	RSignature *big.Int `json:"RSignature"`
	SSignature *big.Int `json:"SSignature"`
}

// ConvToBytes - DO NOT use this for serialization. This
// should only used when shoving a transaction into a hash function
func (t *Transaction) convToBytes() []byte {
	var buff string
	buff += strconv.FormatUint(uint64(t.Version), Base)
	buff += t.XInput.String()
	buff += t.YInput.String()
	buff += t.XOutput.String()
	buff += t.YOutput.String()
	buff += strconv.FormatFloat(float64(t.Amount), 'f', -1, 64)
	buff += strconv.FormatUint(t.Timestamp, Base)
	buff += string(t.Data)
	buff += t.RSignature.String()
	buff += t.SSignature.String()

	return []byte(buff)
}

// HashTransaction - Returns a SHA 256 hash for the transaction
func (t *Transaction) HashTransaction() []byte {
	bytes := t.convToBytes()
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// TransactionSignatureIsValid - Checks to see if the
// signature of the transaction is valid
func (t *Transaction) TransactionSignatureIsValid() bool {
	pubKey := &ecdsa.PublicKey{elliptic.P384(), t.XInput, t.YInput}
	return ecdsa.Verify(pubKey, t.HashTransaction(), t.RSignature, t.SSignature)
}

// TransactionCostIsValid - Checks to see if the person
// who paid for the transaction has enough money to do so.
// Takes in the current status of the blockchain and the
// current transaction pool. If you don't want to use
// the txpool, just pass in a slice with zero elements.
// The index parameter specifies until what index of the blockchain
// you would like to go up until. If that number is -1, that means
// you have to go up the entire blockchain and check everything
func (t *Transaction) TransactionCostIsValid(bc *Blockchain, txpool []Transaction, index int64) bool {
	pubKey := &ecdsa.PublicKey{elliptic.P384(), t.XInput, t.YInput}
	var curAccountBalance float64

	// Get the current
	curAccountBalance = bc.CalcAccountBalanceOnBC(pubKey, index)
	curAccountBalance += CalcAccountBalanceOnTXPool(pubKey, txpool)

	// Check to see if we have enough money to pay
	if t.Amount-curAccountBalance >= 0 {
		return true
	}

	return false
}
