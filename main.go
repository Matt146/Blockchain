package main

import (
	"Blockchain/blockchain"
	"time"
)

func main() {
	// Test mining
	b := blockchain.Block{
		Index:      1,
		Timestamp:  uint64(time.Now().Unix()), // if time machines are a thing, this code is broken
		Difficulty: 1,
	}
	b.MineBlock()
}
