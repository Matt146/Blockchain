package main

import (
	"Blockchain/blockchain"
	"Blockchain/network"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go/log"
)

var wg sync.WaitGroup

func main() {
	/*Test mining*/
	fmt.Println("[+] Testing blockchain...")
	b := blockchain.Block{
		Index:      1,
		Timestamp:  uint64(time.Now().Unix()), // if time machines are a thing, this code is broken
		Difficulty: 1,
	}
	b.MineBlock()

	/*Test networking*/
	fmt.Println("[+] Testing network...")

	// Start Server
	net := network.MakeNetwork()
	// Wont' work because it uses public IP and my router has NAT: net.BootstrapNetwork()
	net.MyIP = "127.0.0.1"
	net.MyID = blockchain.GenRandBytes(32)
	http.HandleFunc("/JOIN", net.JoinHandler)
	http.HandleFunc("/PING", net.PingHandler)
	wg.Add(1)
	defer wg.Done()
	go func() {
		log.Error(http.ListenAndServe(network.Port, nil))
	}()

	fmt.Println("Client sending JOIN request")

	// Start client
	err := net.Join("localhost")
	if err != nil {
		log.Error(err)
	}
	fmt.Printf("Client IP: %s\n", net.MyIP)
	for k := range net.Nodes {
		fmt.Printf("Pinging peer: %v\n", []byte(k))
		net.Ping([]byte(k))
	}

	fmt.Println("Hello, world!")

	// Wait now
	wg.Wait()
}
