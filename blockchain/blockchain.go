package blockchain

import (
	"bytes"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"os"
	"strconv"
)

const (
	// letterBytes - This is the random selection of bytes GenRandBytes picks from
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// initialBlocks - Initial number of blocks allocated by MakeBlockchain
	initialBlocks = 4096

	// Base - This is the base to format uints and ints
	// NOTE: ALL FLOATS ARE FORMATTED TO DECIMALS
	Base = 16

	// DefaultNonceLen - The default length of a nonce (in bytes)
	DefaultNonceLen = 32
)

// Transaction - This struct contains the necessary fields for each transaction
// on the network
type Transaction struct {
	Version   uint32  `json:"Version"`
	Input     []byte  `json:"Input"`
	Output    []byte  `json:"Output"`
	Amount    float64 `json:"Amount"`
	Timestamp uint64  `json:"Timestamp"`
	Data      []byte  `json:"Data"`
	Signature []byte  `json:"Signature"`
}

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

// MakeBlockchain - Call this function to initialize the blockchain struct
func MakeBlockchain() Blockchain {
	return make([]Block, 0, initialBlocks)
}

// AddTransaction - Add a transaction to the
// last block in the blockchain
func (bc *Blockchain) AddTransaction(t Transaction) {
	(*bc)[len(*bc)-1].TXs = append((*bc)[len(*bc)-1].TXs, t)
}

// HashBlock - Generates a hash to a block in the blockchain,
// then assigns the hash to the block and returns it
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
		buff += string(tx.Input)
		buff += string(tx.Output)
		buff += strconv.FormatFloat(tx.Amount, 'f', -1, 64)

		// write the timestamp and extra data
		buff += strconv.FormatUint(tx.Timestamp, Base)
		buff += string(tx.Data)

		// write the signature
		buff += string(tx.Signature)
	}

	// now, hash that buffer, assign it to the block,
	// and return it
	hasher := sha256.New()
	hasher.Write([]byte(buff))
	sum := hasher.Sum(nil)
	b.Hash = sum
	return sum
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

// BlockHashIsValid - Returns true if the hash of the block is valid and meets
// the difficulty
func (b *Block) BlockHashIsValid() bool {
	// Shallow copy the struct and deep
	// copy the slice
	var bCopy Block
	bCopy = *b
	copy(bCopy.Hash, b.Hash)

	bCopy.HashBlock()
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

// MineBlock - This takes a block and hashes and updates
// the nonce, until it produces a hash that matches the difficulty
// rating of the block. Then, assigns the nonce
// that made it happen, and finally returns the block
func (b *Block) MineBlock() []byte {
	var nonce []byte
	SeedRand()
	for {
		// First, hash the block
		b.HashBlock()

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
			break
		}

		// If it isn't equal to the difficulty, then
		// change the nonce
		nonce = GenRandBytes(32)
	}

	return nonce
}

// CalcAccountBalance - This returns the total number of
// coins associated with a public key
func (bc *Blockchain) CalcAccountBalance(pubKey []byte) float64 {
	var totalBalance float64 = 0
	for _, block := range *bc {
		for _, tx := range block.TXs {
			if bytes.Compare(pubKey, tx.Input) == 0 {
				totalBalance -= tx.Amount
			}
			if bytes.Compare(pubKey, tx.Output) == 0 {
				totalBalance += tx.Amount
			}
		}
	}

	return totalBalance
}

// @TODO:
// AddBlock - This takes a block and adds it to the blockchain.
// If the block has not yet been mined, this function automatically
// mines the block. If it has already been mined and is a valid block,
// we add it to the blockchain and set its previous hash
func (bc *Blockchain) AddBlock(b *Block) {

}
