package main

import (
	"Blockchain/blockchain"
	"fmt"
	"time"
)

func main() {
	b := blockchain.Block{
		Index:      1,
		Timestamp:  uint64(time.Now().Unix()), // if time machines are a thing, this code is broken
		Difficulty: 2,
	}
	b.MineBlock()
	fmt.Printf("%v\n", b)
	fmt.Println("Hello, world!")
}
