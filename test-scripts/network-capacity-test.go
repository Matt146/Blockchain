package main

import (
	"Blockchain/network"
	"fmt"
	mrand "math/rand"
	"sync"
)

var wg sync.WaitGroup

const (
	MaxPeerCount = 1024
)

func tortureNetwork(peers []*network.Network) {
	for i := 0; i < 10000; i++ {
		peers[i].Ping(peers[mrand.Intn(MaxPeerCount)].MyID)
	}
}

func main() {
	// Initialize multiple peers
	var peers []*network.Network
	for i := 0; i < MaxPeerCount; i++ {
		peers = append(peers, network.MakeNetwork())
		err := peers[i].Join("127.0.0.1"+network.Port, 10000+i)
		if err != nil {
			fmt.Printf("Error occured making peer: %d\n", i)
		}
	}

	tortureNetwork(peers)

}
