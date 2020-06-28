package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
)

var mux sync.Mutex
var wg sync.WaitGroup
var finishedMining bool
var nonceSolution []byte

const (
	// letterBytes - This is the random selection of bytes GenRandBytes picks from
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// initialBlocks - Initial number of blocks allocated by MakeBlockchain
	initialBlocks = 4096

	// Base - This is the base to format uints and ints
	// NOTE: ALL FLOATS ARE FORMATTED TO DECIMALS
	Base = 10

	// DefaultNonceLen - The default length of a nonce (in bytes)
	DefaultNonceLen = 32
)

// Block - This struct contains all necessary fields for
// a singular block. The blockchain is essentially an array
// of these blocks
type Block struct {
	/*Block headers*/
	Index      uint64 `json:"Index"`
	Hash       []byte `json:"Hash"`
	PrevHash   []byte `json:"PrevHash"`
	Timestamp  uint64 `json:"Timestamp"`
	Difficulty uint32 `json:"Difficulty"`
	Nonce      []byte `json:"Nonce"`

	/*Transaction data*/
	TXs []Transaction `json:"TXs"`
}

// Blockchain - user defined type:
// it's the type "alias" to a blockchain
type Blockchain []Block

/************************************
 * Block initialization
 * and utility functions
************************************/

// MakeBlockchain - Call this function to initialize the blockchain struct
func MakeBlockchain() Blockchain {
	return make([]Block, 0, initialBlocks)
}

// SeedRand - This seeds the insecure random number generator
// with a secure random number
func SeedRand() {
	c := 8
	b := make([]byte, c)
	_, err := crand.Read(b)
	if err != nil {
		os.Exit(-1)
	}
	var seed int64
	buf := bytes.NewBuffer(b)
	binary.Read(buf, binary.BigEndian, &seed)
	rand.Seed(seed)
}

// GenRandBytes - This generates a random selection of bytes of
// n length
func GenRandBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rand.Intn(127))
	}
	return b
}

// Unique - Takes a slice of ints and returns
// another slice of ints with all the
// duplicates removed
func Unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// CalcAccountBalanceOnBC - This returns the total number of
// coins associated with a public key on the blockchain.
// The index parameter specifies how far up the blockchain
// we want to go. If we want to go all the way up,
// pass -1 as the index parameter
// NOTE: THIS ASSUMES THAT ALL BLOCKS IN THE BLOCKCHAIN
// AND TRANSACTIONS IN THE TRANSACTION POOL ARE VALID!
func (bc *Blockchain) CalcAccountBalanceOnBC(pubKey *ecdsa.PublicKey, index int64) float64 {
	var totalBalance float64 = 0
	var count int64 = 0
	for _, block := range *bc {
		for _, tx := range block.TXs {
			if strings.Compare(pubKey.X.String(), tx.XInput.String()) == 0 {
				if strings.Compare(pubKey.Y.String(), tx.YInput.String()) == 0 {
					totalBalance -= tx.Amount
				}
			}
			if strings.Compare(pubKey.X.String(), tx.XOutput.String()) == 0 {
				if strings.Compare(pubKey.Y.String(), tx.YOutput.String()) == 0 {
					totalBalance += tx.Amount
				}
			}
		}

		if index > 0 {
			count++
			if count >= index {
				break
			}
		}
	}

	return totalBalance
}

// CalcAccountBalanceOnTXPool - This returns the total number of
// coins associated with a public key on the blockchain.
// NOTE: THIS ASSUMES THAT ALL BLOCKS IN THE BLOCKCHAIN
// ARE VALID!
func CalcAccountBalanceOnTXPool(pubKey *ecdsa.PublicKey, txpool []Transaction) float64 {
	var totalBalance float64 = 0
	// Next, get the balance of the person within the current transaction
	// pool and add it to the totalBalance
	for _, tx := range txpool {
		if strings.Compare(pubKey.X.String(), tx.XInput.String()) == 0 {
			if strings.Compare(pubKey.Y.String(), tx.YInput.String()) == 0 {
				totalBalance -= tx.Amount
			}
		}
		if strings.Compare(pubKey.X.String(), tx.XOutput.String()) == 0 {
			if strings.Compare(pubKey.Y.String(), tx.YOutput.String()) == 0 {
				totalBalance += tx.Amount
			}
		}
	}

	return totalBalance
}

/************************************
 * Blockchain and block/tx addition
 * functions
************************************/

// AddTransaction - Add a transaction to the
// last block in the blockchain
func (bc *Blockchain) AddTransaction(t Transaction) {
	(*bc)[len(*bc)-1].TXs = append((*bc)[len(*bc)-1].TXs, t)
}

// AddBlock - This takes a block and adds it to the blockchain if it
// proves to be valid. Returns true if block was added. Returns
// false if block wasn't added.
// NOTE: This function should only be called on the second block
// of the blockchain and on
func (bc *Blockchain) AddBlock(b *Block) bool {
	if b.BlockHashIsValid() {
		var txpool []Transaction
		if bc.BlockIsValid(b, txpool) {
			b.PrevHash = (*bc)[len(*bc)-1].Hash
			(*bc) = append(*bc, *b)
			return true
		}
		return false
	}
	return false
}

/**********************************
 * Block hashing functions
**********************************/

// HashBlock - Generates a hash to a block in the blockchain,
// then returns it as a byte slice
func (b *Block) HashBlock() []byte {
	// here is the buffer that stores the data temporarily
	var buff string

	// write the block headers to the buffer
	buff += strconv.FormatUint(b.Timestamp, Base)
	buff += strconv.FormatUint(uint64(b.Difficulty), Base)
	buff += string(b.Nonce)

	// write all the transactions to the buffer
	for _, tx := range b.TXs {
		// write the version
		buff += strconv.FormatUint(uint64(tx.Version), Base)

		// write the input, output, and amount
		buff += tx.XInput.String()
		buff += tx.YInput.String()
		buff += tx.XOutput.String()
		buff += tx.YOutput.String()
		buff += strconv.FormatFloat(tx.Amount, 'f', -1, 64)

		// write the timestamp and extra data
		buff += strconv.FormatUint(tx.Timestamp, Base)
		buff += string(tx.Data)

		// write the signature
		buff += tx.RSignature.String()
		buff += tx.SSignature.String()
	}

	// now, hash that buffer, assign it to the block,
	// and return it
	hasher := sha256.New()
	hasher.Write([]byte(buff))
	return hasher.Sum(nil)
}

// MineBlock - This takes a block and hashes and updates
// the nonce, until it produces a hash that matches the difficulty
// rating of the block. Then, assigns the nonce
// that made it happen, and finally returns the block
// NOTE: YOU CAN'T RUN THIS CONCURRENTLY
func (b *Block) MineBlock() []byte {
	var nonce []byte
	var count int64 = 0
	SeedRand()
	for {
		// First, Generate a nonce
		nonce = GenRandBytes(32)
		b.Nonce = nonce

		// Second, hash the block
		b.Hash = b.HashBlock()

		// Next, check how many bytes that are equal to zero there are in a row
		var numZero uint32 = 0
		for _, v := range b.Hash {
			if v != 0 {
				break
			}
			numZero++
		}

		// Then, check to see if that number corresponds to the difficulty
		if numZero == b.Difficulty {
			count++
			fmt.Printf("[%d] Nonce: %s | Hash: %s\n", count, base64.URLEncoding.EncodeToString(nonce),
				base64.URLEncoding.EncodeToString(b.Hash))
			break
		}

		// Debug:
		count++
		//fmt.Printf("[%d] Nonce: %s | Hash: %s\n", count, base64.URLEncoding.EncodeToString(nonce),
		//	base64.URLEncoding.EncodeToString(b.Hash))
	}

	return nonce
}

/********************************
 * Block validation functions
********************************/

// BlockHashIsValid - Returns true if the hash of the block is valid and meets
// the difficulty
func (b *Block) BlockHashIsValid() bool {
	// Shallow copy the struct and deep
	// copy the slice
	var bCopy Block
	bCopy = *b
	copy(bCopy.Hash, b.Hash)

	bCopy.Hash = bCopy.HashBlock()
	if bytes.Compare(bCopy.Hash, b.Hash) == 0 {
		// Get the number of prefixing zeroes
		var numZero uint32 = 0
		for _, v := range bCopy.Hash {
			if v != 0 {
				break
			}
			numZero++
		}

		// If the number of prefixing zeroes matches
		// the number of zeroes needed, which
		// is specified by the difficulty,
		// then return true. Otherwise, return false
		if numZero == bCopy.Difficulty {
			return true
		}
		return false
	}

	return false
}

// RemoveTransaction - Removes a transaction from
// a transaction slice and returns that slice
func RemoveTransaction(TXs []Transaction, index int) []Transaction {
	return append(TXs[:index], TXs[index+1:]...)
}

// BlockIsValid - This checks to see if all the data in the block is
// valid other than its previous hash, which is supposed to be
// set when finally adding a block to the blockchain.
// If an individual transaction is invalid in the block,
// that transaction gets removed and the function still returns
// true.
// If you would like to not factor in the current transaction pool,
// please pass an empty slice
// (@TODO-OPTIMIZE)
func (bc *Blockchain) BlockIsValid(b *Block, txpool []Transaction) bool {
	var invalidTXIndicies []int

	// First, check the block hash
	if b.BlockHashIsValid() {
		// Validate the transaction signatures in the blockchain
		for i, tx := range b.TXs {
			if tx.TransactionSignatureIsValid() == false {
				invalidTXIndicies = append(invalidTXIndicies, i)
			}
		}

		// Check to see the cost of every single transaction and if
		// the person who paid for it has enough money to do so
		for i, tx := range b.TXs {
			if tx.TransactionCostIsValid(bc, txpool, -1) == false {
				invalidTXIndicies = append(invalidTXIndicies, i)
			}
		}

		// Remove duplicate invalid transaction indexes
		invalidTXIndicies = Unique(invalidTXIndicies)

		// Remove all invalid transactions:
		for _, invalid := range invalidTXIndicies {
			RemoveTransaction(b.TXs, invalid)
		}

		return true
	}

	return false
}

// BlockInBlockchainIsValid - Checks to see if a specific index
// of a block in the blockchain is valid
// (@TODO-OPTIMIZE)
func (bc *Blockchain) BlockInBlockchainIsValid(index int64) bool {
	for i, b := range *bc {
		// First check if the hash is valid
		if b.BlockHashIsValid() {
			// Next, check every single transaction signature in the block
			// and also check if the person who paid in the transaction had
			// enough money to do so (to prevent double spending)
			var txpool []Transaction
			for _, tx := range b.TXs {
				if tx.TransactionSignatureIsValid() == false {
					return false
				}
				if tx.TransactionCostIsValid(bc, txpool, index) == false {
					return false
				}
			}

			// Next, for every other block other than genesis,
			// check its previous hahs
			if i > 0 {
				if bytes.Compare(b.Hash, (*bc)[i-1].PrevHash) != 0 {
					return false
				}
			}
		}

		return false
	}

	return true
}
